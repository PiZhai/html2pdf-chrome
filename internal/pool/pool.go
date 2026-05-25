package pool

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/PiZhai/html2pdf-chrome/internal/browser"
)

// Config controls pool behavior.
type Config struct {
	// ChromePath is the explicit path to Chrome/Chromium. Empty means auto-detect.
	ChromePath string

	// MaxInstances is the maximum number of Chrome instances in the pool.
	// Default: runtime.NumCPU()
	MaxInstances int

	// MinInstances is the minimum number of idle instances to keep warm.
	// Default: 1
	MinInstances int

	// MaxTasksPerInstance is the maximum number of tasks a single instance
	// handles before being recycled. 0 means unlimited.
	// Default: 100
	MaxTasksPerInstance int

	// IdleTimeout is how long an idle instance can sit unused before being
	// reaped. 0 means never reap.
	// Default: 5 minutes
	IdleTimeout time.Duration

	// DebugLog enables Chrome process debug logging.
	DebugLog bool
}

func (c *Config) withDefaults() Config {
	out := *c
	if out.MaxInstances <= 0 {
		out.MaxInstances = runtime.NumCPU()
	}
	if out.MinInstances < 0 {
		out.MinInstances = 0
	}
	if out.MinInstances > out.MaxInstances {
		out.MinInstances = out.MaxInstances
	}
	if out.MinInstances == 0 {
		out.MinInstances = 1
	}
	if out.MaxTasksPerInstance == 0 {
		out.MaxTasksPerInstance = 100
	}
	if out.IdleTimeout == 0 {
		out.IdleTimeout = 5 * time.Minute
	}
	return out
}

// Pool manages a set of reusable Chrome instances.
type Pool struct {
	mu          sync.Mutex
	cond        *sync.Cond
	idle        []*PooledInstance
	activeCount int
	totalCount  int
	closed      bool
	config      Config
	chromePath  string
	stopReaper  chan struct{}
}

// New creates a new Pool and pre-warms MinInstances Chrome processes.
func New(cfg Config) (*Pool, error) {
	cfg = cfg.withDefaults()

	chromePath, err := browser.FindChrome(cfg.ChromePath)
	if err != nil {
		return nil, fmt.Errorf("pool: %w", err)
	}

	p := &Pool{
		config:     cfg,
		chromePath: chromePath,
		idle:       make([]*PooledInstance, 0, cfg.MaxInstances),
		stopReaper: make(chan struct{}),
	}
	p.cond = sync.NewCond(&p.mu)

	// Pre-warm minimum instances.
	for i := 0; i < cfg.MinInstances; i++ {
		inst, err := p.launchInstance()
		if err != nil {
			// Clean up already-created instances.
			for _, existing := range p.idle {
				_ = existing.Close()
			}
			return nil, fmt.Errorf("pool: pre-warm instance %d: %w", i, err)
		}
		p.idle = append(p.idle, inst)
		p.totalCount++
	}

	// Start the idle reaper goroutine.
	if cfg.IdleTimeout > 0 {
		go p.reaper()
	}

	return p, nil
}

// Acquire obtains a Chrome instance from the pool. If no idle instance is
// available and the pool has not reached MaxInstances, a new instance is
// created. If the pool is full, Acquire blocks until an instance is released
// or ctx is cancelled.
func (p *Pool) Acquire(ctx context.Context) (*PooledInstance, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		if p.closed {
			return nil, fmt.Errorf("pool: closed")
		}

		// Try to get a healthy idle instance.
		for len(p.idle) > 0 {
			inst := p.idle[len(p.idle)-1]
			p.idle = p.idle[:len(p.idle)-1]

			if inst.IsHealthy() {
				inst.lastUsedAt = time.Now()
				p.activeCount++
				return inst, nil
			}

			// Unhealthy — discard.
			_ = inst.Close()
			p.totalCount--
		}

		// No idle instances. Can we create a new one?
		if p.totalCount < p.config.MaxInstances {
			// Reserve the slot before releasing the lock for launch.
			p.totalCount++
			p.activeCount++

			inst, err := p.createInstance()
			if err != nil {
				// Unreserve on failure.
				p.totalCount--
				p.activeCount--
				return nil, fmt.Errorf("pool: create instance: %w", err)
			}
			inst.lastUsedAt = time.Now()
			return inst, nil
		}

		// Pool is full — wait for a release or context cancellation.
		// We need to release the lock while waiting.
		waitDone := make(chan struct{})
		go func() {
			select {
			case <-ctx.Done():
				p.cond.Broadcast()
			case <-waitDone:
			}
		}()

		p.cond.Wait()
		close(waitDone)

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
}

// Release returns an instance to the pool. If the instance has exceeded
// MaxTasksPerInstance or is unhealthy, it is closed and discarded.
func (p *Pool) Release(inst *PooledInstance) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.activeCount--

	if p.closed {
		_ = inst.Close()
		p.totalCount--
		p.cond.Broadcast()
		return
	}

	inst.taskCount++
	inst.lastUsedAt = time.Now()

	// Recycle if task limit reached or unhealthy.
	if (p.config.MaxTasksPerInstance > 0 && inst.taskCount >= p.config.MaxTasksPerInstance) || !inst.IsHealthy() {
		_ = inst.Close()
		p.totalCount--
		p.cond.Broadcast()
		return
	}

	p.idle = append(p.idle, inst)
	p.cond.Signal()
}

// Close gracefully shuts down the pool. It waits for all active instances to
// be released, then closes all Chrome processes.
func (p *Pool) Close() error {
	p.mu.Lock()

	if p.closed {
		p.mu.Unlock()
		return nil
	}

	p.closed = true
	close(p.stopReaper)

	// Wait for all active instances to be returned.
	for p.activeCount > 0 {
		p.cond.Wait()
	}

	// Close all idle instances.
	var firstErr error
	for _, inst := range p.idle {
		if err := inst.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	p.idle = nil
	p.totalCount = 0

	p.mu.Unlock()
	return firstErr
}

// Stats returns current pool statistics.
func (p *Pool) Stats() PoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return PoolStats{
		Idle:   len(p.idle),
		Active: p.activeCount,
		Total:  p.totalCount,
	}
}

// PoolStats holds pool statistics.
type PoolStats struct {
	Idle   int
	Active int
	Total  int
}

// createInstance launches a new Chrome process. Must be called with mu held.
// It releases the lock during the potentially slow launch and re-acquires it.
func (p *Pool) createInstance() (*PooledInstance, error) {
	// Release lock during the potentially slow launch.
	p.mu.Unlock()
	inst, err := browser.Launch(p.chromePath, browser.LaunchOptions{
		DebugLog: p.config.DebugLog,
	})
	p.mu.Lock()

	if err != nil {
		return nil, err
	}

	return &PooledInstance{
		instance:   inst,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
	}, nil
}

// launchInstance launches a new Chrome process without any mutex interaction.
// Used during pool initialization when the lock is not held.
func (p *Pool) launchInstance() (*PooledInstance, error) {
	inst, err := browser.Launch(p.chromePath, browser.LaunchOptions{
		DebugLog: p.config.DebugLog,
	})
	if err != nil {
		return nil, err
	}

	return &PooledInstance{
		instance:   inst,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
	}, nil
}

// reaper periodically scans idle instances and closes those that have been
// idle longer than IdleTimeout, while keeping at least MinInstances alive.
func (p *Pool) reaper() {
	ticker := time.NewTicker(p.config.IdleTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopReaper:
			return
		case <-ticker.C:
			p.reapIdle()
		}
	}
}

func (p *Pool) reapIdle() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	now := time.Now()
	kept := make([]*PooledInstance, 0, len(p.idle))

	for _, inst := range p.idle {
		// Keep at least MinInstances total (idle + active).
		if p.totalCount <= p.config.MinInstances {
			kept = append(kept, inst)
			continue
		}

		if now.Sub(inst.lastUsedAt) > p.config.IdleTimeout {
			_ = inst.Close()
			p.totalCount--
		} else {
			kept = append(kept, inst)
		}
	}

	p.idle = kept
}
