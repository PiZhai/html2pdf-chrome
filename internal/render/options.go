package render

// Options holds the parameters passed to Chrome's Page.printToPDF command.
// Pointer fields are optional — nil means use Chrome's default.
type Options struct {
	OutputFile string // Destination file path for the generated PDF

	Landscape               *bool    // Landscape orientation
	DisplayHeaderFooter     *bool    // Show header/footer
	PrintBackground         *bool    // Print CSS backgrounds
	Scale                   *float64 // Rendering scale (0.1–2.0)
	PaperWidth              float64  // Paper width in inches
	PaperHeight             float64  // Paper height in inches
	MarginTop               *float64 // Top margin in inches
	MarginBottom            *float64 // Bottom margin in inches
	MarginLeft              *float64 // Left margin in inches
	MarginRight             *float64 // Right margin in inches
	PageRanges              *string  // Page ranges, e.g. "1-3, 5"
	HeaderTemplate          *string  // Header HTML template
	FooterTemplate          *string  // Footer HTML template
	PreferCSSPageSize       *bool    // Prefer CSS @page size
	TransferMode            *string  // "ReturnAsBase64" or "ReturnAsStream"
	GenerateTaggedPDF       *bool    // Accessible tagged PDF
	GenerateDocumentOutline *bool    // Embed document outline/bookmarks
}
