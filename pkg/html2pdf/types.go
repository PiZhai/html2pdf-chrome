package html2pdf

import "time"

const DefaultTimeout = 45 * time.Second

type PaperPreset string

const (
	PaperLetter  PaperPreset = "letter"
	PaperLegal   PaperPreset = "legal"
	PaperTabloid PaperPreset = "tabloid"
	PaperA3      PaperPreset = "a3"
	PaperA4      PaperPreset = "a4"
	PaperA5      PaperPreset = "a5"
)

type TransferMode string

const (
	TransferModeBase64 TransferMode = "base64"
	TransferModeStream TransferMode = "stream"
)

type Request struct {
	URL        string
	HTMLFile   string
	OutputPath string
	Options    Options
}

type Options struct {
	ChromePath   string
	Timeout      time.Duration
	WaitSelector string
	Paper        PaperPreset

	Landscape               bool
	DisplayHeaderFooter     bool
	PrintBackground         bool
	Scale                   *float64
	MarginTop               *float64
	MarginBottom            *float64
	MarginLeft              *float64
	MarginRight             *float64
	PageRanges              string
	HeaderTemplate          string
	FooterTemplate          string
	PreferCSSPageSize       bool
	GenerateTaggedPDF       bool
	GenerateDocumentOutline bool
	TransferMode            TransferMode
	ChromeDebugLog          bool
}

func (r Request) Validate() error {
	_, err := r.toConfig()
	return err
}

func ConvertURL(url, outputPath string, options Options) error {
	return Convert(Request{
		URL:        url,
		OutputPath: outputPath,
		Options:    options,
	})
}

func ConvertFile(htmlFile, outputPath string, options Options) error {
	return Convert(Request{
		HTMLFile:   htmlFile,
		OutputPath: outputPath,
		Options:    options,
	})
}

func Float64(v float64) *float64 { return &v }
