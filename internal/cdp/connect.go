package cdp

import (
	"context"
	"fmt"

	"github.com/chromedp/chromedp"
)

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
