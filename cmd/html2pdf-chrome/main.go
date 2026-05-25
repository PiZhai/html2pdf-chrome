package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PiZhai/html2pdf-chrome/pkg/html2pdf"
)

func main() {
	log.SetFlags(0)

	req, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	if err := html2pdf.Convert(req); err != nil {
		log.Fatal(err)
	}
}

func parseFlags() (html2pdf.Request, error) {
	req := html2pdf.Request{}
	var paper string
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
	var waitNetworkIdle bool
	var networkIdleTime time.Duration
	var waitExpression string

	flag.BoolVar(&landscape, "landscape", false, "Render PDF in landscape orientation")
	flag.BoolVar(&displayHeaderFooter, "display-header-footer", false, "Display header and footer")
	flag.BoolVar(&printBackground, "print-background", false, "Print CSS backgrounds")
	flag.BoolVar(&req.Options.PreferCSSPageSize, "prefer-css-page-size", false, "Prefer CSS @page size over configured paper size")
	flag.BoolVar(&generateTaggedPDF, "generate-tagged-pdf", false, "Generate tagged (accessible) PDF")
	flag.BoolVar(&generateDocumentOutline, "generate-document-outline", false, "Embed document outline into the PDF")
	flag.BoolVar(&req.Options.ChromeDebugLog, "chrome-debug-log", false, "Emit Chrome process logs to stderr for debugging")
	flag.BoolVar(&req.Options.NoSandbox, "no-sandbox", false, "Disable Chrome sandbox (required when running as root in containers)")
	flag.BoolVar(&waitNetworkIdle, "wait-network-idle", false, "Wait for network idle before printing")
	flag.DurationVar(&networkIdleTime, "network-idle-time", 500*time.Millisecond, "Network idle quiet period duration")
	flag.StringVar(&waitExpression, "wait-expression", "", "Custom JS expression to poll until truthy before printing")
	flag.StringVar(&req.URL, "url", "", "HTTP/HTTPS URL to render")
	flag.StringVar(&req.HTMLFile, "html-file", "", "Local HTML file to render")
	flag.StringVar(&req.OutputPath, "out", "output.pdf", "Output PDF file path")
	flag.StringVar(&req.Options.ChromePath, "chrome-path", "", "Chrome/Chromium executable path")
	flag.DurationVar(&req.Options.Timeout, "timeout", 45*time.Second, "Overall render timeout")
	flag.StringVar(&req.Options.WaitSelector, "wait-selector", "", "Optional CSS selector to wait for before printing")
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

	req.Options.Paper = html2pdf.PaperPreset(strings.ToLower(strings.TrimSpace(paper)))
	req.Options.Landscape = landscape
	req.Options.DisplayHeaderFooter = displayHeaderFooter
	req.Options.PrintBackground = printBackground
	req.Options.Scale = html2pdf.Float64(scale)
	req.Options.MarginTop = html2pdf.Float64(marginTop)
	req.Options.MarginBottom = html2pdf.Float64(marginBottom)
	req.Options.MarginLeft = html2pdf.Float64(marginLeft)
	req.Options.MarginRight = html2pdf.Float64(marginRight)
	req.Options.PageRanges = pageRanges
	req.Options.HeaderTemplate = headerTemplate
	req.Options.FooterTemplate = footerTemplate
	req.Options.GenerateTaggedPDF = generateTaggedPDF
	req.Options.GenerateDocumentOutline = generateDocumentOutline

	switch strings.ToLower(strings.TrimSpace(transferMode)) {
	case "":
		req.Options.TransferMode = ""
	case "base64":
		req.Options.TransferMode = html2pdf.TransferModeBase64
	case "stream":
		req.Options.TransferMode = html2pdf.TransferModeStream
	default:
		return html2pdf.Request{}, fmt.Errorf("unsupported transfer-mode %q; use base64 or stream", transferMode)
	}

	req.Options.WaitNetworkIdle = waitNetworkIdle
	req.Options.NetworkIdleTime = networkIdleTime
	req.Options.WaitExpression = waitExpression

	return req, nil
}
