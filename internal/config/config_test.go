package config

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParsePaperPresetSetsA4Size(t *testing.T) {
	cfg := &Config{}

	if err := cfg.ParsePaperPreset("A4"); err != nil {
		t.Fatalf("ParsePaperPreset returned error: %v", err)
	}

	if cfg.PaperWidth == nil || cfg.PaperHeight == nil {
		t.Fatal("paper size pointers were not set")
	}

	if *cfg.PaperWidth != A4Width {
		t.Fatalf("unexpected paper width: got %v want %v", *cfg.PaperWidth, A4Width)
	}

	if *cfg.PaperHeight != A4Height {
		t.Fatalf("unexpected paper height: got %v want %v", *cfg.PaperHeight, A4Height)
	}
}

func TestValidateRequiresSingleInputSource(t *testing.T) {
	cfg := &Config{
		Timeout: 5 * time.Second,
	}
	cfg.PaperWidth = float64Ptr(A4Width)
	cfg.PaperHeight = float64Ptr(A4Height)

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate should reject empty input source")
	}

	cfg.URL = "https://example.com"
	cfg.HTMLFile = "sample.html"
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate should reject both URL and HTMLFile being set")
	}
}

func TestValidateRequiresPaperSize(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "sample.html")
	if err := os.WriteFile(htmlFile, []byte("<html><body>ok</body></html>"), 0o644); err != nil {
		t.Fatalf("write html file: %v", err)
	}

	cfg := &Config{
		HTMLFile: htmlFile,
		Timeout:  5 * time.Second,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate should reject missing paper size")
	}
}

func TestValidateAllowsBlankPageRanges(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "sample.html")
	if err := os.WriteFile(htmlFile, []byte("<html><body>ok</body></html>"), 0o644); err != nil {
		t.Fatalf("write html file: %v", err)
	}

	pageRanges := "   "
	cfg := &Config{
		HTMLFile:   htmlFile,
		Timeout:    5 * time.Second,
		PageRanges: &pageRanges,
	}
	cfg.PaperWidth = float64Ptr(A4Width)
	cfg.PaperHeight = float64Ptr(A4Height)

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate should allow blank page ranges: %v", err)
	}
}

func TestInputTargetBuildsFileURL(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "sample.html")
	if err := os.WriteFile(htmlFile, []byte("<html><body>ok</body></html>"), 0o644); err != nil {
		t.Fatalf("write html file: %v", err)
	}

	cfg := &Config{HTMLFile: htmlFile}

	target, err := cfg.InputTarget()
	if err != nil {
		t.Fatalf("InputTarget returned error: %v", err)
	}

	parsed, err := url.Parse(target)
	if err != nil {
		t.Fatalf("parse target URL: %v", err)
	}

	if parsed.Scheme != "file" {
		t.Fatalf("unexpected scheme: got %q want %q", parsed.Scheme, "file")
	}

	if !filepath.IsAbs(parsed.Path) && len(parsed.Path) > 2 {
		t.Fatalf("expected absolute file path, got %q", parsed.Path)
	}
}

func TestPrepareOutputPathCreatesParentDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		OutputFile: filepath.Join(tmpDir, "nested", "result.pdf"),
	}

	outputPath, err := cfg.PrepareOutputPath()
	if err != nil {
		t.Fatalf("PrepareOutputPath returned error: %v", err)
	}

	if !filepath.IsAbs(outputPath) {
		t.Fatalf("expected absolute path, got %q", outputPath)
	}

	parentDir := filepath.Dir(outputPath)
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Fatalf("stat parent dir: %v", err)
	}

	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", parentDir)
	}
}
