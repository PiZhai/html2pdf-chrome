package render

type Options struct {
	OutputFile string

	Landscape               *bool
	DisplayHeaderFooter     *bool
	PrintBackground         *bool
	Scale                   *float64
	PaperWidth              float64
	PaperHeight             float64
	MarginTop               *float64
	MarginBottom            *float64
	MarginLeft              *float64
	MarginRight             *float64
	PageRanges              *string
	HeaderTemplate          *string
	FooterTemplate          *string
	PreferCSSPageSize       *bool
	TransferMode            *string
	GenerateTaggedPDF       *bool
	GenerateDocumentOutline *bool
}
