package html2pdf

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/PiZhai/html2pdf-chrome/internal/browser"
)

func skipIfNoChromeConverter(t *testing.T) {
	t.Helper()
	if _, err := browser.FindChrome(""); err != nil {
		t.Skipf("skipping: Chrome/Chromium not available: %v", err)
	}
}

func TestConverterSingleConversion(t *testing.T) {
	skipIfNoChromeConverter(t)

	converter, err := NewConverter(ConverterConfig{
		MaxInstances: 2,
		MinInstances: 1,
	})
	if err != nil {
		t.Fatalf("NewConverter returned error: %v", err)
	}
	defer converter.Close()

	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	outputFile := filepath.Join(tmpDir, "output.pdf")

	html := `<!doctype html><html><head><meta charset="utf-8"></head><body><h1>Converter Test</h1></body></html>`
	if err := os.WriteFile(htmlFile, []byte(html), 0o644); err != nil {
		t.Fatalf("write html: %v", err)
	}

	err = converter.Convert(Request{
		HTMLFile:   htmlFile,
		OutputPath: outputFile,
		Options: Options{
			Timeout: 30 * time.Second,
		},
	})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}

	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("output PDF not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("output PDF is empty")
	}
}

func TestConverterConcurrentConversions(t *testing.T) {
	skipIfNoChromeConverter(t)

	converter, err := NewConverter(ConverterConfig{
		MaxInstances: 3,
		MinInstances: 1,
	})
	if err != nil {
		t.Fatalf("NewConverter returned error: %v", err)
	}
	defer converter.Close()

	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")

	html := `<!doctype html><html><head><meta charset="utf-8"></head><body><h1>Concurrent Test</h1></body></html>`
	if err := os.WriteFile(htmlFile, []byte(html), 0o644); err != nil {
		t.Fatalf("write html: %v", err)
	}

	const concurrency = 5
	var wg sync.WaitGroup
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			outputFile := filepath.Join(tmpDir, filepath.Base(t.Name())+"-"+string(rune('a'+idx))+".pdf")
			err := converter.Convert(Request{
				HTMLFile:   htmlFile,
				OutputPath: outputFile,
				Options: Options{
					Timeout: 30 * time.Second,
				},
			})
			if err != nil {
				errors <- err
				return
			}

			info, err := os.Stat(outputFile)
			if err != nil {
				errors <- err
				return
			}
			if info.Size() == 0 {
				errors <- os.ErrInvalid
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Fatalf("concurrent conversion error: %v", err)
	}

	stats := converter.Stats()
	t.Logf("Pool stats after concurrent test: idle=%d active=%d total=%d",
		stats.IdleInstances, stats.ActiveInstances, stats.TotalInstances)
}

func TestConverterReusesInstances(t *testing.T) {
	skipIfNoChromeConverter(t)

	converter, err := NewConverter(ConverterConfig{
		MaxInstances: 1,
		MinInstances: 1,
	})
	if err != nil {
		t.Fatalf("NewConverter returned error: %v", err)
	}
	defer converter.Close()

	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")

	html := `<!doctype html><html><head><meta charset="utf-8"></head><body><h1>Reuse Test</h1></body></html>`
	if err := os.WriteFile(htmlFile, []byte(html), 0o644); err != nil {
		t.Fatalf("write html: %v", err)
	}

	// Run 3 sequential conversions with MaxInstances=1.
	// They should all reuse the same instance.
	for i := 0; i < 3; i++ {
		outputFile := filepath.Join(tmpDir, "output-"+string(rune('0'+i))+".pdf")
		err := converter.Convert(Request{
			HTMLFile:   htmlFile,
			OutputPath: outputFile,
			Options: Options{
				Timeout: 30 * time.Second,
			},
		})
		if err != nil {
			t.Fatalf("Convert %d returned error: %v", i, err)
		}
	}

	stats := converter.Stats()
	if stats.TotalInstances != 1 {
		t.Fatalf("expected 1 total instance (reused), got %d", stats.TotalInstances)
	}
}
