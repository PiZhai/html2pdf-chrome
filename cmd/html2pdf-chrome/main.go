package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"html2pdf-chrome/internal/app"
	"html2pdf-chrome/internal/config"
)

func main() {
	log.SetFlags(0)

	cfg, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}

func parseFlags() (*config.Config, error) {
	cfg := &config.Config{}
	var paper string
	var preferCSSPageSize bool
	var chromeDebugLog bool
	var landscape bool
	var displayHeaderFooter bool
	var printBackground bool
	var scale float64
	var marginTop float64
	var marginBottom float64
	var marginLeft float64
	var marginRight float64
	var pageRanges string
	var headerTemplate string
	var footerTemplate string
	var generateTaggedPDF bool
	var generateDocumentOutline bool
	var transferMode string

	flag.BoolVar(&chromeDebugLog, "chrome-debug-log", false, "Emit Chrome process logs to stderr for debugging")
	flag.BoolVar(&landscape, "landscape", false, "Render PDF in landscape orientation")
	flag.BoolVar(&displayHeaderFooter, "display-header-footer", false, "Display header and footer")
	flag.BoolVar(&printBackground, "print-background", false, "Print CSS backgrounds")
	flag.BoolVar(&preferCSSPageSize, "prefer-css-page-size", false, "Prefer CSS @page size over configured paper size")
	flag.BoolVar(&generateTaggedPDF, "generate-tagged-pdf", false, "Generate tagged (accessible) PDF")
	flag.BoolVar(&generateDocumentOutline, "generate-document-outline", false, "Embed document outline into the PDF")
	flag.StringVar(&cfg.URL, "url", "", "HTTP/HTTPS URL to render")
	flag.StringVar(&cfg.HTMLFile, "html-file", "", "Local HTML file to render")
	flag.StringVar(&cfg.OutputFile, "out", "output.pdf", "Output PDF file path")
	flag.StringVar(&cfg.ChromePath, "chrome-path", "", "Chrome/Chromium executable path")
	flag.DurationVar(&cfg.Timeout, "timeout", 45*time.Second, "Overall render timeout")
	flag.StringVar(&cfg.WaitSelector, "wait-selector", "", "Optional CSS selector to wait for before printing")
	flag.StringVar(&paper, "paper", "a4", "Paper preset: letter, legal, tabloid, a3, a4, a5")
	flag.StringVar(&headerTemplate, "header-template", "", "HTML template for the PDF header")
	flag.StringVar(&footerTemplate, "footer-template", "", "HTML template for the PDF footer")
	flag.StringVar(&transferMode, "transfer-mode", "", "PDF transfer mode: base64 or stream")
	flag.Float64Var(&scale, "scale", 1.0, "Scale of webpage rendering (0.1 to 2.0)")
	flag.Float64Var(&marginTop, "margin-top", 0.3937007874, "Top margin in inches")
	flag.Float64Var(&marginBottom, "margin-bottom", 0.3937007874, "Bottom margin in inches")
	flag.Float64Var(&marginLeft, "margin-left", 0.3937007874, "Left margin in inches")
	flag.Float64Var(&marginRight, "margin-right", 0.3937007874, "Right margin in inches")
	flag.StringVar(&pageRanges, "page-ranges", "", "Page ranges to print, e.g. '1-3, 5'")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n")
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "  %s -url https://example.com -out output.pdf\n", os.Args[0])
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "  %s -html-file ./testdata/sample.html -out output.pdf\n\n", os.Args[0])
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	cfg.Landscape = boolPtr(landscape)
	cfg.DisplayHeaderFooter = boolPtr(displayHeaderFooter)
	cfg.PrintBackground = boolPtr(printBackground)
	cfg.ChromeDebugLog = &chromeDebugLog
	cfg.PreferCSSPageSize = &preferCSSPageSize
	cfg.Scale = float64Ptr(scale)
	cfg.MarginTop = float64Ptr(marginTop)
	cfg.MarginBottom = float64Ptr(marginBottom)
	cfg.MarginLeft = float64Ptr(marginLeft)
	cfg.MarginRight = float64Ptr(marginRight)
	cfg.PageRanges = stringPtr(pageRanges)
	cfg.HeaderTemplate = stringPtr(headerTemplate)
	cfg.FooterTemplate = stringPtr(footerTemplate)
	cfg.GenerateTaggedPDF = boolPtr(generateTaggedPDF)
	cfg.GenerateDocumentOutline = boolPtr(generateDocumentOutline)

	switch strings.ToLower(strings.TrimSpace(transferMode)) {
	case "":
		cfg.TransferMode = nil
	case "base64":
		cfg.TransferMode = stringPtr("ReturnAsBase64")
	case "stream":
		cfg.TransferMode = stringPtr("ReturnAsStream")
	default:
		return nil, fmt.Errorf("unsupported transfer-mode %q; use base64 or stream", transferMode)
	}

	if err := cfg.ParsePaperPreset(paper); err != nil {
		return nil, err
	}

	return cfg, nil
}

func boolPtr(v bool) *bool { return &v }

func float64Ptr(v float64) *float64 { return &v }

func stringPtr(v string) *string { return &v }
