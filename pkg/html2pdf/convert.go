package html2pdf

import (
	"fmt"
	"strings"

	"github.com/PiZhai/html2pdf-chrome/internal/app"
	"github.com/PiZhai/html2pdf-chrome/internal/config"
)

func Convert(req Request) error {
	cfg, err := req.toConfig()
	if err != nil {
		return err
	}

	return app.Run(cfg)
}

func (r Request) toConfig() (*config.Config, error) {
	cfg := &config.Config{
		URL:                     strings.TrimSpace(r.URL),
		HTMLFile:                strings.TrimSpace(r.HTMLFile),
		OutputFile:              strings.TrimSpace(r.OutputPath),
		ChromePath:              strings.TrimSpace(r.Options.ChromePath),
		Timeout:                 r.Options.Timeout,
		WaitSelector:            strings.TrimSpace(r.Options.WaitSelector),
		WaitNetworkIdle:         boolPtr(r.Options.WaitNetworkIdle),
		NetworkIdleTime:         r.Options.NetworkIdleTime,
		WaitExpression:          strings.TrimSpace(r.Options.WaitExpression),
		ChromeDebugLog:          boolPtr(r.Options.ChromeDebugLog),
		Landscape:               boolPtr(r.Options.Landscape),
		DisplayHeaderFooter:     boolPtr(r.Options.DisplayHeaderFooter),
		PrintBackground:         boolPtr(r.Options.PrintBackground),
		MarginTop:               cloneFloat64Ptr(r.Options.MarginTop),
		MarginBottom:            cloneFloat64Ptr(r.Options.MarginBottom),
		MarginLeft:              cloneFloat64Ptr(r.Options.MarginLeft),
		MarginRight:             cloneFloat64Ptr(r.Options.MarginRight),
		PreferCSSPageSize:       boolPtr(r.Options.PreferCSSPageSize),
		GenerateTaggedPDF:       boolPtr(r.Options.GenerateTaggedPDF),
		GenerateDocumentOutline: boolPtr(r.Options.GenerateDocumentOutline),
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultTimeout
	}

	if r.Options.Scale != nil {
		cfg.Scale = cloneFloat64Ptr(r.Options.Scale)
	}

	if pageRanges := strings.TrimSpace(r.Options.PageRanges); pageRanges != "" {
		cfg.PageRanges = stringPtr(pageRanges)
	}
	if headerTemplate := strings.TrimSpace(r.Options.HeaderTemplate); headerTemplate != "" {
		cfg.HeaderTemplate = stringPtr(headerTemplate)
	}
	if footerTemplate := strings.TrimSpace(r.Options.FooterTemplate); footerTemplate != "" {
		cfg.FooterTemplate = stringPtr(footerTemplate)
	}

	switch r.Options.TransferMode {
	case "":
	case TransferModeBase64:
		cfg.TransferMode = stringPtr("ReturnAsBase64")
	case TransferModeStream:
		cfg.TransferMode = stringPtr("ReturnAsStream")
	default:
		return nil, fmt.Errorf("unsupported transfer mode %q", r.Options.TransferMode)
	}

	paper := string(r.Options.Paper)
	if paper == "" {
		paper = string(PaperA4)
	}

	if err := cfg.ParsePaperPreset(paper); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func boolPtr(v bool) *bool { return &v }

func cloneFloat64Ptr(v *float64) *float64 {
	if v == nil {
		return nil
	}

	cloned := *v
	return &cloned
}

func stringPtr(v string) *string { return &v }
