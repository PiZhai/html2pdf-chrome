// Package cdp provides Chrome DevTools Protocol operations including
// connecting to a browser instance, navigating pages, and implementing
// various wait strategies for page readiness detection.
package cdp

import (
	"context"
	"fmt"

	"github.com/chromedp/chromedp"
)

// Connect establishes a CDP connection to a running Chrome instance via its
// WebSocket debugging URL. It returns a chromedp context (which represents an
// isolated browser tab) and a cancel function that closes the tab and
// disconnects.
func Connect(wsURL string) (context.Context, context.CancelFunc, error) {
	if wsURL == "" {
		return nil, nil, fmt.Errorf("empty WebSocket URL")
	}

	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(), wsURL)

	ctx, cancel := chromedp.NewContext(allocCtx)

	stop := func() {
		cancel()
		allocCancel()
	}

	return ctx, stop, nil
}
