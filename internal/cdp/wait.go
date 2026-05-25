package cdp

import (
	"time"

	"github.com/chromedp/chromedp"
)

// WaitDocumentReady polls until document.readyState === "complete", indicating
// that the page and all sub-resources (images, stylesheets) have finished loading.
func WaitDocumentReady(timeout time.Duration) chromedp.Action {
	var ready bool

	return chromedp.Poll(
		`document.readyState === "complete"`,
		&ready,
		chromedp.WithPollingInterval(100*time.Millisecond),
		chromedp.WithPollingTimeout(timeout),
	)
}

// WaitFontsReady polls until all web fonts have finished loading.
// Falls back to true if the document.fonts API is not available.
func WaitFontsReady(timeout time.Duration) chromedp.Action {
	var ready bool

	return chromedp.Poll(
		`document.fonts ? document.fonts.status === "loaded" : true`,
		&ready,
		chromedp.WithPollingInterval(100*time.Millisecond),
		chromedp.WithPollingTimeout(timeout),
	)
}

// WaitForExpression polls a user-provided JS expression until it returns a
// truthy value. This allows callers to gate PDF export on application-specific
// readiness signals (e.g. "window.__RENDER_DONE === true").
func WaitForExpression(expr string, timeout time.Duration) chromedp.Action {
	var ready bool

	return chromedp.Poll(
		expr,
		&ready,
		chromedp.WithPollingInterval(200*time.Millisecond),
		chromedp.WithPollingTimeout(timeout),
	)
}
