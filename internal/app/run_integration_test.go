package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"html2pdf-chrome/internal/browser"
	"html2pdf-chrome/internal/config"
)

func TestRunGeneratesPDF(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if _, err := browser.FindChrome(""); err != nil {
		t.Skipf("skipping integration test because Chrome/Chromium is unavailable: %v", err)
	}

	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "sample.html")
	outputFile := filepath.Join(tmpDir, "output.pdf")

	html := `<!doctype html><html><head><meta charset="utf-8"><title>Test</title></head><body><h1>Hello PDF</h1></body></html>`
	if err := os.WriteFile(htmlFile, []byte(html), 0o644); err != nil {
		t.Fatalf("write html file: %v", err)
	}

	cfg := &config.Config{
		HTMLFile:   htmlFile,
		OutputFile: outputFile,
		Timeout:    30 * time.Second,
	}
	if err := cfg.ParsePaperPreset("a4"); err != nil {
		t.Fatalf("ParsePaperPreset returned error: %v", err)
	}

	if err := Run(cfg); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("expected output PDF to exist: %v", err)
	}

	if info.Size() == 0 {
		t.Fatal("expected output PDF to be non-empty")
	}
}
