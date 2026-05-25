package app

import (
	"context"

	"github.com/PiZhai/html2pdf-chrome/internal/cdp"
	"github.com/PiZhai/html2pdf-chrome/internal/config"
	"github.com/PiZhai/html2pdf-chrome/internal/pool"
	"github.com/PiZhai/html2pdf-chrome/internal/render"
)

// RunWithPool executes a conversion using a pooled Chrome instance.
// The instance is acquired from the pool, used to render the PDF in an
// isolated browser tab, and then returned to the pool.
func RunWithPool(p *pool.Pool, cfg *config.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	target, err := cfg.InputTarget()
	if err != nil {
		return err
	}

	outputPath, err := cfg.PrepareOutputPath()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	inst, err := p.Acquire(ctx)
	if err != nil {
		return err
	}
	defer p.Release(inst)

	cdpCtx, cdpCancel, err := cdp.Connect(inst.WebSocketURL())
	if err != nil {
		return err
	}
	defer cdpCancel()

	renderCtx, renderCancel := context.WithTimeout(cdpCtx, cfg.Timeout)
	defer renderCancel()

	pageOpts := cdp.PageOptions{
		TargetURL:      target,
		WaitSelector:   cfg.WaitSelector,
		WaitExpression: cfg.WaitExpression,
		Timeout:        cfg.Timeout,
	}
	if cfg.WaitNetworkIdle != nil && *cfg.WaitNetworkIdle {
		pageOpts.WaitNetworkIdle = true
		pageOpts.NetworkIdleTime = cfg.NetworkIdleTime
	}

	if err := cdp.OpenPage(renderCtx, pageOpts); err != nil {
		return err
	}

	renderOptions := render.Options{
		OutputFile:              outputPath,
		Landscape:               cfg.Landscape,
		DisplayHeaderFooter:     cfg.DisplayHeaderFooter,
		PrintBackground:         cfg.PrintBackground,
		Scale:                   cfg.Scale,
		PaperWidth:              *cfg.PaperWidth,
		PaperHeight:             *cfg.PaperHeight,
		MarginTop:               cfg.MarginTop,
		MarginBottom:            cfg.MarginBottom,
		MarginLeft:              cfg.MarginLeft,
		MarginRight:             cfg.MarginRight,
		PageRanges:              cfg.PageRanges,
		HeaderTemplate:          cfg.HeaderTemplate,
		FooterTemplate:          cfg.FooterTemplate,
		PreferCSSPageSize:       cfg.PreferCSSPageSize,
		TransferMode:            cfg.TransferMode,
		GenerateTaggedPDF:       cfg.GenerateTaggedPDF,
		GenerateDocumentOutline: cfg.GenerateDocumentOutline,
	}

	if err := render.PrintToFile(renderCtx, renderOptions); err != nil {
		return err
	}

	return nil
}
