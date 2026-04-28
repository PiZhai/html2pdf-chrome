package cdp

import (
	"time"

	"github.com/chromedp/chromedp"
)

func WaitDocumentReady(timeout time.Duration) chromedp.Action {
	var ready bool

	return chromedp.Poll(
		`document.readyState === "complete"`,
		&ready,
		chromedp.WithPollingInterval(100*time.Millisecond),
		chromedp.WithPollingTimeout(timeout),
	)
}

func WaitFontsReady(timeout time.Duration) chromedp.Action {
	var ready bool

	return chromedp.Poll(
		`document.fonts ? document.fonts.status === "loaded" : true`,
		&ready,
		chromedp.WithPollingInterval(100*time.Millisecond),
		chromedp.WithPollingTimeout(timeout),
	)
}
