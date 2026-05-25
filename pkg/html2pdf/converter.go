package html2pdf

import (
	"time"

	"github.com/PiZhai/html2pdf-chrome/internal/app"
	"github.com/PiZhai/html2pdf-chrome/internal/pool"
)

// ConverterConfig configures the Chrome instance pool used by Converter.
type ConverterConfig struct {
	// ChromePath is the explicit path to Chrome/Chromium executable.
	// Empty means auto-detect.
	ChromePath string

	// MaxInstances is the maximum number of Chrome instances in the pool.
	// Default: runtime.NumCPU()
	MaxInstances int

	// MinInstances is the minimum number of idle instances to keep warm.
	// Instances are pre-warmed at creation time.
	// Default: 1
	MinInstances int

	// MaxTasksPerInstance is the maximum number of tasks a single Chrome
	// instance handles before being recycled. This prevents memory leak
	// accumulation in long-running Chrome processes.
	// Default: 100 (0 means unlimited)
	MaxTasksPerInstance int

	// IdleTimeout is how long an idle instance can sit unused before being
	// reaped. The pool always keeps at least MinInstances alive.
	// Default: 5 minutes (0 means never reap)
	IdleTimeout time.Duration

	// ChromeDebugLog enables Chrome process debug logging to stderr.
	ChromeDebugLog bool

	// NoSandbox disables Chrome's sandbox. Required when running as root
	// in Docker containers. Not recommended outside of containers.
	NoSandbox bool
}

// Converter is a reusable HTML-to-PDF converter backed by a Chrome instance
// pool. It is safe for concurrent use.
//
// Use NewConverter to create one, and call Close when done to release all
// Chrome processes.
type Converter struct {
	pool *pool.Pool
}

// NewConverter creates a Converter and pre-warms the configured minimum number
// of Chrome instances.
func NewConverter(cfg ConverterConfig) (*Converter, error) {
	p, err := pool.New(pool.Config{
		ChromePath:          cfg.ChromePath,
		MaxInstances:        cfg.MaxInstances,
		MinInstances:        cfg.MinInstances,
		MaxTasksPerInstance: cfg.MaxTasksPerInstance,
		IdleTimeout:         cfg.IdleTimeout,
		DebugLog:            cfg.ChromeDebugLog,
		NoSandbox:           cfg.NoSandbox,
	})
	if err != nil {
		return nil, err
	}

	return &Converter{pool: p}, nil
}

// Convert renders a PDF using a pooled Chrome instance. It is safe to call
// concurrently from multiple goroutines.
func (c *Converter) Convert(req Request) error {
	cfg, err := req.toConfig()
	if err != nil {
		return err
	}
	return app.RunWithPool(c.pool, cfg)
}

// Close gracefully shuts down the Converter. It waits for all in-progress
// conversions to complete, then closes all Chrome instances.
func (c *Converter) Close() error {
	return c.pool.Close()
}

// Stats returns current pool statistics (idle, active, total instance counts).
func (c *Converter) Stats() ConverterStats {
	s := c.pool.Stats()
	return ConverterStats{
		IdleInstances:   s.Idle,
		ActiveInstances: s.Active,
		TotalInstances:  s.Total,
	}
}

// ConverterStats holds pool statistics.
type ConverterStats struct {
	IdleInstances   int
	ActiveInstances int
	TotalInstances  int
}
