package app

import (
	"context"
	"github.com/PiZhai/html2pdf-chrome/internal/browser"
	"github.com/PiZhai/html2pdf-chrome/internal/cdp"
	"github.com/PiZhai/html2pdf-chrome/internal/config"
	"github.com/PiZhai/html2pdf-chrome/internal/render"
)

func Run(cfg *config.Config) error {
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

	chromePath, err := browser.FindChrome(cfg.ChromePath)
	if err != nil {
		return err
	}

	instance, err := browser.Launch(chromePath, browser.LaunchOptions{
		DebugLog: cfg.ChromeDebugLog != nil && *cfg.ChromeDebugLog,
	})
	if err != nil {
		return err
	}
	defer instance.Close()

	cdpCtx, cdpCancel, err := cdp.Connect(instance.WebSocketURL)
	if err != nil {
		return err
	}
	defer cdpCancel()

	renderCtx, renderCancel := context.WithTimeout(cdpCtx, cfg.Timeout)
	defer renderCancel()

	if err := cdp.OpenPage(renderCtx, target, cfg.WaitSelector); err != nil {
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
