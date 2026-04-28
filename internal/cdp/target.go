package cdp

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

func OpenPage(ctx context.Context, targetURL string, waitSelector string) error {
	actions := []chromedp.Action{
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		WaitDocumentReady(15 * time.Second),
		WaitFontsReady(10 * time.Second),
	}

	if waitSelector != "" {
		actions = append(actions, chromedp.WaitVisible(waitSelector, chromedp.ByQuery))
	}

	return chromedp.Run(ctx, actions...)
}
