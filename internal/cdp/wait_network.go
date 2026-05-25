package cdp

import (
	"context"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// WaitNetworkIdle waits until there are no inflight network requests for the
// specified idle duration. It enables Network domain events, tracks requests
// via RequestWillBeSent / LoadingFinished / LoadingFailed, and considers the
// network idle once the inflight count stays at zero for idleTime.
//
// The overall wait is bounded by timeout.
func WaitNetworkIdle(idleTime, timeout time.Duration) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		if err := network.Enable().Do(ctx); err != nil {
			return err
		}

		var mu sync.Mutex
		inflight := make(map[network.RequestID]struct{})
		idleTimer := time.NewTimer(idleTime)
		idleTimer.Stop()

		resetIdle := func() {
			idleTimer.Stop()
			// Drain channel if needed.
			select {
			case <-idleTimer.C:
			default:
			}
			if len(inflight) == 0 {
				idleTimer.Reset(idleTime)
			}
		}

		chromedp.ListenTarget(ctx, func(ev interface{}) {
			switch e := ev.(type) {
			case *network.EventRequestWillBeSent:
				if shouldIgnoreRequest(e) {
					return
				}
				mu.Lock()
				inflight[e.RequestID] = struct{}{}
				idleTimer.Stop()
				mu.Unlock()

			case *network.EventLoadingFinished:
				mu.Lock()
				delete(inflight, e.RequestID)
				if len(inflight) == 0 {
					idleTimer.Reset(idleTime)
				}
				mu.Unlock()

			case *network.EventLoadingFailed:
				mu.Lock()
				delete(inflight, e.RequestID)
				if len(inflight) == 0 {
					idleTimer.Reset(idleTime)
				}
				mu.Unlock()
			}
		})

		// Kick off the idle timer in case there are no pending requests at all.
		mu.Lock()
		resetIdle()
		mu.Unlock()

		deadline := time.NewTimer(timeout)
		defer deadline.Stop()

		select {
		case <-idleTimer.C:
			return nil
		case <-deadline.C:
			return context.DeadlineExceeded
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// shouldIgnoreRequest filters out requests that should not block idle detection.
func shouldIgnoreRequest(e *network.EventRequestWillBeSent) bool {
	if e.Request == nil {
		return true
	}

	url := e.Request.URL

	// Ignore data: and blob: URLs — they are local and resolve instantly.
	if len(url) > 5 && (url[:5] == "data:" || url[:5] == "blob:") {
		return true
	}

	// Ignore WebSocket upgrades.
	if e.Type == network.ResourceTypeWebSocket {
		return true
	}

	return false
}
