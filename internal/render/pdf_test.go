package render

import (
	"testing"

	"github.com/chromedp/cdproto/page"
)

func TestBuildPrintToPDFParamsAppliesConfiguredFields(t *testing.T) {
	landscape := true
	displayHeaderFooter := true
	printBackground := true
	scale := 1.25
	marginTop := 0.5
	marginBottom := 0.6
	marginLeft := 0.7
	marginRight := 0.8
	pageRanges := "1-2"
	headerTemplate := "<span class='title'></span>"
	footerTemplate := "<span class='pageNumber'></span>"
	preferCSSPageSize := true
	transferMode := "ReturnAsStream"
	generateTaggedPDF := true
	generateDocumentOutline := true

	params, err := buildPrintToPDFParams(Options{
		OutputFile:              "out.pdf",
		Landscape:               &landscape,
		DisplayHeaderFooter:     &displayHeaderFooter,
		PrintBackground:         &printBackground,
		Scale:                   &scale,
		PaperWidth:              8.5,
		PaperHeight:             11.0,
		MarginTop:               &marginTop,
		MarginBottom:            &marginBottom,
		MarginLeft:              &marginLeft,
		MarginRight:             &marginRight,
		PageRanges:              &pageRanges,
		HeaderTemplate:          &headerTemplate,
		FooterTemplate:          &footerTemplate,
		PreferCSSPageSize:       &preferCSSPageSize,
		TransferMode:            &transferMode,
		GenerateTaggedPDF:       &generateTaggedPDF,
		GenerateDocumentOutline: &generateDocumentOutline,
	})
	if err != nil {
		t.Fatalf("buildPrintToPDFParams returned error: %v", err)
	}

	if !params.Landscape {
		t.Fatal("expected landscape to be true")
	}
	if !params.DisplayHeaderFooter {
		t.Fatal("expected displayHeaderFooter to be true")
	}
	if !params.PrintBackground {
		t.Fatal("expected printBackground to be true")
	}
	if params.Scale != scale {
		t.Fatalf("unexpected scale: got %v want %v", params.Scale, scale)
	}
	if params.PaperWidth != 8.5 || params.PaperHeight != 11.0 {
		t.Fatalf("unexpected paper size: got %vx%v", params.PaperWidth, params.PaperHeight)
	}
	if params.MarginTop != marginTop || params.MarginBottom != marginBottom || params.MarginLeft != marginLeft || params.MarginRight != marginRight {
		t.Fatalf("unexpected margins: %+v", params)
	}
	if params.PageRanges != pageRanges {
		t.Fatalf("unexpected page ranges: got %q want %q", params.PageRanges, pageRanges)
	}
	if params.HeaderTemplate != headerTemplate {
		t.Fatalf("unexpected header template: got %q want %q", params.HeaderTemplate, headerTemplate)
	}
	if params.FooterTemplate != footerTemplate {
		t.Fatalf("unexpected footer template: got %q want %q", params.FooterTemplate, footerTemplate)
	}
	if !params.PreferCSSPageSize {
		t.Fatal("expected preferCSSPageSize to be true")
	}
	if params.TransferMode != page.PrintToPDFTransferModeReturnAsStream {
		t.Fatalf("unexpected transfer mode: got %q", params.TransferMode)
	}
	if !params.GenerateTaggedPDF {
		t.Fatal("expected generateTaggedPDF to be true")
	}
	if !params.GenerateDocumentOutline {
		t.Fatal("expected generateDocumentOutline to be true")
	}
}

func TestBuildPrintToPDFParamsUsesDefaultMargins(t *testing.T) {
	params, err := buildPrintToPDFParams(Options{
		OutputFile:  "out.pdf",
		PaperWidth:  8.5,
		PaperHeight: 11.0,
	})
	if err != nil {
		t.Fatalf("buildPrintToPDFParams returned error: %v", err)
	}

	if params.MarginTop != defaultMarginInches ||
		params.MarginBottom != defaultMarginInches ||
		params.MarginLeft != defaultMarginInches ||
		params.MarginRight != defaultMarginInches {
		t.Fatalf("expected default margins of %v, got top=%v bottom=%v left=%v right=%v",
			defaultMarginInches, params.MarginTop, params.MarginBottom, params.MarginLeft, params.MarginRight)
	}
}

func TestBuildPrintToPDFParamsRejectsUnknownTransferMode(t *testing.T) {
	transferMode := "NotSupported"

	_, err := buildPrintToPDFParams(Options{
		OutputFile:   "out.pdf",
		PaperWidth:   8.5,
		PaperHeight:  11.0,
		TransferMode: &transferMode,
	})
	if err == nil {
		t.Fatal("expected buildPrintToPDFParams to reject unknown transfer mode")
	}
}
