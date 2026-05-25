// Package render handles PDF generation via Chrome's Page.printToPDF command.
package render

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/chromedp/cdproto/cdp"
	cdpio "github.com/chromedp/cdproto/io"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// defaultMarginInches is approximately 1cm, used when no margin is specified.
const defaultMarginInches = 0.3937007874

// PrintToFile calls Page.printToPDF with the given options and writes the
// resulting PDF to the specified output file.
func PrintToFile(ctx context.Context, options Options) error {
	var pdfBytes []byte

	action := chromedp.ActionFunc(func(ctx context.Context) error {
		params, err := buildPrintToPDFParams(options)
		if err != nil {
			return err
		}

		data, stream, err := params.Do(ctx)
		if err != nil {
			return err
		}

		if stream != "" {
			pdfBytes, err = readPDFStream(ctx, stream)
			if err != nil {
				return err
			}
			return nil
		}

		pdfBytes = data
		return nil
	})

	if err := chromedp.Run(ctx, action); err != nil {
		return fmt.Errorf("print to PDF: %w", err)
	}

	if err := os.WriteFile(options.OutputFile, pdfBytes, 0o644); err != nil {
		return fmt.Errorf("write PDF file: %w", err)
	}

	return nil
}

// buildPrintToPDFParams constructs the Chrome Page.printToPDF parameters from
// the given options, applying defaults for unset margins.
func buildPrintToPDFParams(options Options) (*page.PrintToPDFParams, error) {
	params := page.PrintToPDF().
		WithPaperWidth(options.PaperWidth).
		WithPaperHeight(options.PaperHeight).
		WithMarginTop(valueOrDefault(options.MarginTop, defaultMarginInches)).
		WithMarginBottom(valueOrDefault(options.MarginBottom, defaultMarginInches)).
		WithMarginLeft(valueOrDefault(options.MarginLeft, defaultMarginInches)).
		WithMarginRight(valueOrDefault(options.MarginRight, defaultMarginInches))

	if options.Landscape != nil {
		params = params.WithLandscape(*options.Landscape)
	}
	if options.DisplayHeaderFooter != nil {
		params = params.WithDisplayHeaderFooter(*options.DisplayHeaderFooter)
	}
	if options.PrintBackground != nil {
		params = params.WithPrintBackground(*options.PrintBackground)
	}
	if options.Scale != nil {
		params = params.WithScale(*options.Scale)
	}
	if options.PageRanges != nil && strings.TrimSpace(*options.PageRanges) != "" {
		params = params.WithPageRanges(strings.TrimSpace(*options.PageRanges))
	}
	if options.HeaderTemplate != nil && strings.TrimSpace(*options.HeaderTemplate) != "" {
		params = params.WithHeaderTemplate(*options.HeaderTemplate)
	}
	if options.FooterTemplate != nil && strings.TrimSpace(*options.FooterTemplate) != "" {
		params = params.WithFooterTemplate(*options.FooterTemplate)
	}
	if options.PreferCSSPageSize != nil {
		params = params.WithPreferCSSPageSize(*options.PreferCSSPageSize)
	}
	if options.TransferMode != nil {
		switch *options.TransferMode {
		case "ReturnAsBase64":
			params = params.WithTransferMode(page.PrintToPDFTransferModeReturnAsBase64)
		case "ReturnAsStream":
			params = params.WithTransferMode(page.PrintToPDFTransferModeReturnAsStream)
		default:
			return nil, fmt.Errorf("unsupported transfer mode %q", *options.TransferMode)
		}
	}
	if options.GenerateTaggedPDF != nil {
		params = params.WithGenerateTaggedPDF(*options.GenerateTaggedPDF)
	}
	if options.GenerateDocumentOutline != nil {
		params = params.WithGenerateDocumentOutline(*options.GenerateDocumentOutline)
	}

	return params, nil
}

// readPDFStream reads a PDF from a Chrome IO stream handle, reassembling
// chunks and decoding base64 if needed.
func readPDFStream(ctx context.Context, handle cdpio.StreamHandle) ([]byte, error) {
	defer func() {
		_ = cdpio.Close(handle).Do(ctx)
	}()

	var buf bytes.Buffer

	for {
		readParams := cdpio.Read(handle).WithSize(1 << 20)
		var res cdpio.ReadReturns
		if err := cdp.Execute(ctx, cdpio.CommandRead, readParams, &res); err != nil {
			return nil, fmt.Errorf("read PDF stream: %w", err)
		}

		chunk := []byte(res.Data)
		if res.Base64encoded {
			decoded, err := base64.StdEncoding.DecodeString(res.Data)
			if err != nil {
				return nil, fmt.Errorf("decode PDF stream chunk: %w", err)
			}
			chunk = decoded
		}

		if _, err := buf.Write(chunk); err != nil {
			return nil, fmt.Errorf("buffer PDF stream: %w", err)
		}

		if res.EOF {
			break
		}
	}

	return buf.Bytes(), nil
}

func valueOrDefault(v *float64, def float64) float64 {
	if v != nil {
		return *v
	}

	return def
}
