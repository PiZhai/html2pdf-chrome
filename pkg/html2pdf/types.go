// Package html2pdf provides a Go API for converting HTML pages to PDF using
// Chrome/Chromium in headless mode via the Chrome DevTools Protocol.
//
// Two usage patterns are supported:
//   - Single-shot: Convert / ConvertURL / ConvertFile — each call launches a
//     new Chrome instance and closes it after the conversion.
//   - Pooled: NewConverter returns a Converter backed by a Chrome instance pool,
//     suitable for concurrent, high-throughput server workloads.
package html2pdf

import "time"

// DefaultTimeout is the default overall timeout for a single conversion.
const DefaultTimeout = 45 * time.Second

// PaperPreset represents a named paper size.
type PaperPreset string

const (
	PaperLetter  PaperPreset = "letter"  // 8.5 × 11 inches (US)
	PaperLegal   PaperPreset = "legal"   // 8.5 × 14 inches
	PaperTabloid PaperPreset = "tabloid" // 11 × 17 inches
	PaperA3      PaperPreset = "a3"      // 297 × 420 mm
	PaperA4      PaperPreset = "a4"      // 210 × 297 mm (default)
	PaperA5      PaperPreset = "a5"      // 148 × 210 mm
)

// TransferMode controls how Chrome returns the PDF data internally.
type TransferMode string

const (
	TransferModeBase64 TransferMode = "base64" // Return PDF as base64 string (default)
	TransferModeStream TransferMode = "stream" // Return PDF via IO stream (better for large files)
)

// Request describes a single HTML-to-PDF conversion task.
type Request struct {
	// URL is the HTTP/HTTPS page to render. Mutually exclusive with HTMLFile.
	URL string

	// HTMLFile is the path to a local HTML file to render. Mutually exclusive with URL.
	HTMLFile string

	// OutputPath is the file path where the generated PDF will be written.
	OutputPath string

	// Options controls rendering behavior.
	Options Options
}

// Options configures how the PDF is rendered.
type Options struct {
	// ChromePath explicitly specifies the Chrome/Chromium executable path.
	// If empty, the tool auto-detects using platform-specific candidate paths,
	// the CHROME_PATH environment variable, and PATH lookup.
	ChromePath string

	// Timeout is the overall timeout for the entire conversion (navigation +
	// waiting + rendering). Default: 45s.
	Timeout time.Duration

	// WaitSelector is a CSS selector to wait for (visibility) before exporting.
	// Useful for waiting on a specific DOM element to appear.
	WaitSelector string

	// Paper is the paper size preset. Default: "a4".
	Paper PaperPreset

	// WaitNetworkIdle enables waiting for network idle before exporting PDF.
	// Network idle means zero inflight requests for a configurable quiet period.
	// Useful for pages with async resource loading (images, fonts, XHR).
	WaitNetworkIdle bool

	// NetworkIdleTime is the quiet period required to consider the network idle.
	// Only effective when WaitNetworkIdle is true. Default: 500ms.
	NetworkIdleTime time.Duration

	// WaitExpression is a custom JavaScript expression that is polled until it
	// returns a truthy value. Use for application-specific readiness signals.
	// Example: "window.__RENDER_DONE === true"
	WaitExpression string

	// Landscape enables landscape (horizontal) page orientation.
	Landscape bool

	// DisplayHeaderFooter enables header and footer rendering in the PDF.
	// Must be true for HeaderTemplate/FooterTemplate to take effect.
	DisplayHeaderFooter bool

	// PrintBackground enables printing of CSS background colors and images.
	PrintBackground bool

	// Scale is the rendering scale factor. Range: 0.1 to 2.0. Default: 1.0.
	// Nil means use Chrome's default (1.0).
	Scale *float64

	// MarginTop is the top page margin in inches. Default: ~0.394 (1cm).
	MarginTop *float64

	// MarginBottom is the bottom page margin in inches. Default: ~0.394 (1cm).
	MarginBottom *float64

	// MarginLeft is the left page margin in inches. Default: ~0.394 (1cm).
	MarginLeft *float64

	// MarginRight is the right page margin in inches. Default: ~0.394 (1cm).
	MarginRight *float64

	// PageRanges specifies which pages to print, e.g. "1-3, 5".
	// Empty string means all pages.
	PageRanges string

	// HeaderTemplate is the HTML template for the page header.
	// Supported CSS classes: .date, .title, .url, .pageNumber, .totalPages.
	// Requires DisplayHeaderFooter = true.
	HeaderTemplate string

	// FooterTemplate is the HTML template for the page footer.
	// Same CSS classes as HeaderTemplate. Requires DisplayHeaderFooter = true.
	FooterTemplate string

	// PreferCSSPageSize allows CSS @page { size: ... } to override the Paper setting.
	PreferCSSPageSize bool

	// GenerateTaggedPDF produces an accessible (tagged) PDF with structural information.
	GenerateTaggedPDF bool

	// GenerateDocumentOutline embeds a document outline (bookmarks) into the PDF.
	GenerateDocumentOutline bool

	// TransferMode controls how Chrome returns PDF data internally.
	// "stream" is better for very large PDFs. Default: "base64".
	TransferMode TransferMode

	// ChromeDebugLog enables Chrome process stderr output for debugging.
	ChromeDebugLog bool

	// NoSandbox disables Chrome's sandbox. Required when running as root in
	// Docker containers where user namespaces are restricted.
	NoSandbox bool
}

// Validate checks whether the Request is well-formed without performing the conversion.
func (r Request) Validate() error {
	_, err := r.toConfig()
	return err
}

// ConvertURL is a convenience function that converts a URL to PDF.
func ConvertURL(url, outputPath string, options Options) error {
	return Convert(Request{
		URL:        url,
		OutputPath: outputPath,
		Options:    options,
	})
}

// ConvertFile is a convenience function that converts a local HTML file to PDF.
func ConvertFile(htmlFile, outputPath string, options Options) error {
	return Convert(Request{
		HTMLFile:   htmlFile,
		OutputPath: outputPath,
		Options:    options,
	})
}

// Float64 is a helper that returns a pointer to the given float64 value.
// Useful for setting optional numeric fields like Scale, MarginTop, etc.
func Float64(v float64) *float64 { return &v }
