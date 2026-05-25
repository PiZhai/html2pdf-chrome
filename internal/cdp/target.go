package cdp

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// PageOptions configures how OpenPage navigates and waits for readiness.
type PageOptions struct {
	TargetURL       string
	WaitSelector    string
	WaitNetworkIdle bool
	NetworkIdleTime time.Duration
	WaitExpression  string
	Timeout         time.Duration
}

// OpenPage navigates to the target URL and executes the wait strategy chain:
//  1. body ready
//  2. document.readyState === "complete"
//  3. document.fonts.status === "loaded"
//  4. Network idle (optional)
//  5. CSS selector visible (optional)
//  6. Custom JS expression truthy (optional)
func OpenPage(ctx context.Context, opts PageOptions) error {
	actions := []chromedp.Action{
		chromedp.Navigate(opts.TargetURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		WaitDocumentReady(15 * time.Second),
		WaitFontsReady(10 * time.Second),
	}

	// Network idle detection — placed after fonts because font fetches are
	// network requests themselves.
	if opts.WaitNetworkIdle {
		idleTime := opts.NetworkIdleTime
		if idleTime == 0 {
			idleTime = 500 * time.Millisecond
		}
		timeout := opts.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		actions = append(actions, WaitNetworkIdle(idleTime, timeout))
	}

	// CSS selector visibility.
	if opts.WaitSelector != "" {
		actions = append(actions, chromedp.WaitVisible(opts.WaitSelector, chromedp.ByQuery))
	}

	// Custom JS expression — last gate before export.
	if opts.WaitExpression != "" {
		timeout := opts.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		actions = append(actions, WaitForExpression(opts.WaitExpression, timeout))
	}

	return chromedp.Run(ctx, actions...)
}
