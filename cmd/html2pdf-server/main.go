package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/PiZhai/html2pdf-chrome/pkg/html2pdf"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var (
		addr         string
		maxInstances int
		minInstances int
		noSandbox    bool
		outputDir    string
	)

	flag.StringVar(&addr, "addr", ":8080", "HTTP listen address")
	flag.IntVar(&maxInstances, "max-instances", 4, "Max Chrome instances in pool")
	flag.IntVar(&minInstances, "min-instances", 2, "Min idle Chrome instances")
	flag.BoolVar(&noSandbox, "no-sandbox", false, "Disable Chrome sandbox (required in containers)")
	flag.StringVar(&outputDir, "output-dir", "/tmp/html2pdf", "Temporary output directory")
	flag.Parse()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("create output dir: %v", err)
	}

	converter, err := html2pdf.NewConverter(html2pdf.ConverterConfig{
		MaxInstances: maxInstances,
		MinInstances: minInstances,
		NoSandbox:    noSandbox,
	})
	if err != nil {
		log.Fatalf("create converter: %v", err)
	}
	defer converter.Close()

	var requestID atomic.Int64

	http.HandleFunc("/convert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ConvertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.URL == "" && req.HTML == "" {
			http.Error(w, `must provide "url" or "html"`, http.StatusBadRequest)
			return
		}

		id := requestID.Add(1)
		start := time.Now()

		// 准备临时文件
		tmpFile := filepath.Join(outputDir, fmt.Sprintf("pdf-%d-%d.pdf", id, start.UnixMilli()))

		opts := html2pdf.Options{
			Timeout:         parseDurationOrDefault(req.Timeout, 45*time.Second),
			PrintBackground: req.PrintBackground,
			WaitNetworkIdle: req.WaitNetworkIdle,
			WaitExpression:  req.WaitExpression,
			WaitSelector:    req.WaitSelector,
			NoSandbox:       noSandbox,
		}

		if req.Paper != "" {
			opts.Paper = html2pdf.PaperPreset(req.Paper)
		}
		if req.Landscape {
			opts.Landscape = true
		}

		var convertReq html2pdf.Request

		if req.URL != "" {
			convertReq = html2pdf.Request{
				URL:        req.URL,
				OutputPath: tmpFile,
				Options:    opts,
			}
		} else {
			// HTML 内容写入临时文件
			htmlFile := tmpFile + ".html"
			if err := os.WriteFile(htmlFile, []byte(req.HTML), 0644); err != nil {
				http.Error(w, "write html: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer os.Remove(htmlFile)

			convertReq = html2pdf.Request{
				HTMLFile:   htmlFile,
				OutputPath: tmpFile,
				Options:    opts,
			}
		}

		if err := converter.Convert(convertReq); err != nil {
			log.Printf("[req=%d] convert error: %v (took %v)", id, err, time.Since(start))
			http.Error(w, "convert failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(tmpFile)

		log.Printf("[req=%d] converted in %v", id, time.Since(start))

		// 返回 PDF
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", `attachment; filename="output.pdf"`)

		f, err := os.Open(tmpFile)
		if err != nil {
			http.Error(w, "read pdf: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		info, _ := f.Stat()
		if info != nil {
			w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		stats := converter.Stats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"pool": map[string]int{
				"idle":   stats.IdleInstances,
				"active": stats.ActiveInstances,
				"total":  stats.TotalInstances,
			},
		})
	})

	log.Printf("html2pdf-server listening on %s (pool: min=%d, max=%d)", addr, minInstances, maxInstances)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

// ConvertRequest is the JSON body for POST /convert.
type ConvertRequest struct {
	// URL to render. Mutually exclusive with HTML.
	URL string `json:"url,omitempty"`

	// Raw HTML content to render. Mutually exclusive with URL.
	HTML string `json:"html,omitempty"`

	// Paper preset: a4, letter, legal, etc.
	Paper string `json:"paper,omitempty"`

	// Landscape orientation.
	Landscape bool `json:"landscape,omitempty"`

	// Print CSS backgrounds.
	PrintBackground bool `json:"printBackground,omitempty"`

	// Wait for network idle before export.
	WaitNetworkIdle bool `json:"waitNetworkIdle,omitempty"`

	// Custom JS expression to wait for.
	WaitExpression string `json:"waitExpression,omitempty"`

	// CSS selector to wait for visibility.
	WaitSelector string `json:"waitSelector,omitempty"`

	// Timeout as a duration string (e.g. "30s", "1m"). Default: 45s.
	Timeout string `json:"timeout,omitempty"`
}

func parseDurationOrDefault(s string, def time.Duration) time.Duration {
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return def
	}
	return d
}
