package html2pdf

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidateAcceptsURLWithDefaults(t *testing.T) {
	req := Request{
		URL: "https://example.com",
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidateRejectsUnsupportedTransferMode(t *testing.T) {
	req := Request{
		URL: "https://example.com",
		Options: Options{
			TransferMode: TransferMode("pipe"),
		},
	}

	if err := req.Validate(); err == nil {
		t.Fatal("Validate should reject unsupported transfer mode")
	}
}

func TestToConfigAppliesDefaults(t *testing.T) {
	req := Request{
		URL: "https://example.com",
	}

	cfg, err := req.toConfig()
	if err != nil {
		t.Fatalf("toConfig returned error: %v", err)
	}

	if cfg.Timeout != DefaultTimeout {
		t.Fatalf("unexpected timeout: got %v want %v", cfg.Timeout, DefaultTimeout)
	}

	if cfg.PaperWidth == nil || cfg.PaperHeight == nil {
		t.Fatal("paper dimensions were not set")
	}

	if *cfg.PaperWidth != 8.2677165354 || *cfg.PaperHeight != 11.6929133858 {
		t.Fatalf("unexpected default paper size: got %v x %v", *cfg.PaperWidth, *cfg.PaperHeight)
	}
}

func TestToConfigMapsOptions(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "sample.html")
	if err := os.WriteFile(htmlFile, []byte("<html><body>ok</body></html>"), 0o644); err != nil {
		t.Fatalf("write html file: %v", err)
	}

	scale := 1.25
	marginTop := 0.0
	marginBottom := 0.5

	req := Request{
		HTMLFile:   htmlFile,
		OutputPath: filepath.Join(tmpDir, "out.pdf"),
		Options: Options{
			Timeout:                 10 * time.Second,
			Paper:                   PaperLetter,
			PrintBackground:         true,
			DisplayHeaderFooter:     true,
			Scale:                   Float64(scale),
			MarginTop:               Float64(marginTop),
			MarginBottom:            Float64(marginBottom),
			PageRanges:              "1-2",
			HeaderTemplate:          `<span class="title"></span>`,
			PreferCSSPageSize:       true,
			TransferMode:            TransferModeStream,
			GenerateTaggedPDF:       true,
			GenerateDocumentOutline: true,
			ChromeDebugLog:          true,
		},
	}

	cfg, err := req.toConfig()
	if err != nil {
		t.Fatalf("toConfig returned error: %v", err)
	}

	if cfg.Timeout != 10*time.Second {
		t.Fatalf("unexpected timeout: got %v want %v", cfg.Timeout, 10*time.Second)
	}
	if cfg.Scale == nil || *cfg.Scale != scale {
		t.Fatalf("unexpected scale: got %#v want %v", cfg.Scale, scale)
	}
	if cfg.MarginTop == nil || *cfg.MarginTop != marginTop {
		t.Fatalf("unexpected margin top: got %#v want %v", cfg.MarginTop, marginTop)
	}
	if cfg.MarginBottom == nil || *cfg.MarginBottom != marginBottom {
		t.Fatalf("unexpected margin bottom: got %#v want %v", cfg.MarginBottom, marginBottom)
	}
	if cfg.TransferMode == nil || *cfg.TransferMode != "ReturnAsStream" {
		t.Fatalf("unexpected transfer mode: got %#v want %q", cfg.TransferMode, "ReturnAsStream")
	}
	if cfg.HeaderTemplate == nil || *cfg.HeaderTemplate == "" {
		t.Fatal("expected header template to be set")
	}
	if cfg.PageRanges == nil || *cfg.PageRanges != "1-2" {
		t.Fatalf("unexpected page ranges: got %#v", cfg.PageRanges)
	}
	if cfg.PaperWidth == nil || *cfg.PaperWidth != 8.5 {
		t.Fatalf("unexpected paper width: got %#v want %v", cfg.PaperWidth, 8.5)
	}
	if cfg.PaperHeight == nil || *cfg.PaperHeight != 11.0 {
		t.Fatalf("unexpected paper height: got %#v want %v", cfg.PaperHeight, 11.0)
	}
}
