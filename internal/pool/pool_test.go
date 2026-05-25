package pool

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/PiZhai/html2pdf-chrome/internal/browser"
)

func skipIfNoChrome(t *testing.T) {
	t.Helper()
	if _, err := browser.FindChrome(""); err != nil {
		t.Skipf("skipping: Chrome/Chromium not available: %v", err)
	}
}

func TestNewPoolPreWarmsInstances(t *testing.T) {
	skipIfNoChrome(t)

	p, err := New(Config{
		MinInstances:        2,
		MaxInstances:        4,
		MaxTasksPerInstance: 10,
		IdleTimeout:         1 * time.Minute,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer p.Close()

	stats := p.Stats()
	if stats.Idle != 2 {
		t.Fatalf("expected 2 idle instances, got %d", stats.Idle)
	}
	if stats.Total != 2 {
		t.Fatalf("expected 2 total instances, got %d", stats.Total)
	}
}

func TestAcquireAndRelease(t *testing.T) {
	skipIfNoChrome(t)

	p, err := New(Config{
		MinInstances:        1,
		MaxInstances:        2,
		MaxTasksPerInstance: 100,
		IdleTimeout:         1 * time.Minute,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer p.Close()

	ctx := context.Background()

	inst, err := p.Acquire(ctx)
	if err != nil {
		t.Fatalf("Acquire returned error: %v", err)
	}

	stats := p.Stats()
	if stats.Active != 1 {
		t.Fatalf("expected 1 active, got %d", stats.Active)
	}
	if stats.Idle != 0 {
		t.Fatalf("expected 0 idle, got %d", stats.Idle)
	}

	wsURL := inst.WebSocketURL()
	if wsURL == "" {
		t.Fatal("expected non-empty WebSocket URL")
	}

	p.Release(inst)

	stats = p.Stats()
	if stats.Active != 0 {
		t.Fatalf("expected 0 active after release, got %d", stats.Active)
	}
	if stats.Idle != 1 {
		t.Fatalf("expected 1 idle after release, got %d", stats.Idle)
	}
}

func TestAcquireBlocksWhenPoolFull(t *testing.T) {
	skipIfNoChrome(t)

	p, err := New(Config{
		MinInstances:        1,
		MaxInstances:        1,
		MaxTasksPerInstance: 100,
		IdleTimeout:         1 * time.Minute,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer p.Close()

	ctx := context.Background()

	// Acquire the only instance.
	inst, err := p.Acquire(ctx)
	if err != nil {
		t.Fatalf("Acquire returned error: %v", err)
	}

	// Try to acquire with a short timeout — should fail.
	shortCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	_, err = p.Acquire(shortCtx)
	if err == nil {
		t.Fatal("expected Acquire to fail when pool is full and context times out")
	}

	// Release and try again — should succeed.
	p.Release(inst)

	inst2, err := p.Acquire(ctx)
	if err != nil {
		t.Fatalf("Acquire after release returned error: %v", err)
	}
	p.Release(inst2)
}

func TestConcurrentAcquireRelease(t *testing.T) {
	skipIfNoChrome(t)

	p, err := New(Config{
		MinInstances:        1,
		MaxInstances:        3,
		MaxTasksPerInstance: 100,
		IdleTimeout:         1 * time.Minute,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer p.Close()

	const goroutines = 6
	const tasksPerGoroutine = 3

	var wg sync.WaitGroup
	errors := make(chan error, goroutines*tasksPerGoroutine)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < tasksPerGoroutine; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				inst, err := p.Acquire(ctx)
				cancel()
				if err != nil {
					errors <- err
					return
				}

				// Simulate some work.
				time.Sleep(10 * time.Millisecond)

				p.Release(inst)
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Fatalf("concurrent acquire/release error: %v", err)
	}
}

func TestMaxTasksPerInstanceRecycles(t *testing.T) {
	skipIfNoChrome(t)

	p, err := New(Config{
		MinInstances:        1,
		MaxInstances:        2,
		MaxTasksPerInstance: 2,
		IdleTimeout:         1 * time.Minute,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer p.Close()

	ctx := context.Background()

	// Use the instance twice (reaching MaxTasksPerInstance).
	for i := 0; i < 2; i++ {
		inst, err := p.Acquire(ctx)
		if err != nil {
			t.Fatalf("Acquire %d returned error: %v", i, err)
		}
		p.Release(inst)
	}

	// After 2 releases, the instance should have been recycled.
	// The pool should still be able to serve (by creating a new instance).
	inst, err := p.Acquire(ctx)
	if err != nil {
		t.Fatalf("Acquire after recycle returned error: %v", err)
	}

	if inst.taskCount != 0 {
		t.Fatalf("expected fresh instance with taskCount=0, got %d", inst.taskCount)
	}

	p.Release(inst)
}

func TestCloseWaitsForActive(t *testing.T) {
	skipIfNoChrome(t)

	p, err := New(Config{
		MinInstances:        1,
		MaxInstances:        2,
		MaxTasksPerInstance: 100,
		IdleTimeout:         1 * time.Minute,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	ctx := context.Background()
	inst, err := p.Acquire(ctx)
	if err != nil {
		t.Fatalf("Acquire returned error: %v", err)
	}

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- p.Close()
	}()

	// Close should block because we have an active instance.
	select {
	case <-closeDone:
		t.Fatal("Close returned before active instance was released")
	case <-time.After(100 * time.Millisecond):
		// Expected — Close is blocking.
	}

	// Release the instance — Close should now complete.
	p.Release(inst)

	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Close did not return after releasing instance")
	}
}

func TestAcquireOnClosedPoolFails(t *testing.T) {
	skipIfNoChrome(t)

	p, err := New(Config{
		MinInstances:        1,
		MaxInstances:        2,
		MaxTasksPerInstance: 100,
		IdleTimeout:         1 * time.Minute,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if err := p.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	_, err = p.Acquire(context.Background())
	if err == nil {
		t.Fatal("expected Acquire on closed pool to fail")
	}
}
