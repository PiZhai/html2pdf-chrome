package config

import (
	"fmt"
	neturl "net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	URL            string        `json:"url,omitempty"`
	HTMLFile       string        `json:"htmlFile,omitempty"`
	OutputFile     string        `json:"outputFile,omitempty"`
	ChromePath     string        `json:"chromePath,omitempty"`
	ChromeDebugLog *bool         `json:"chromeDebugLog,omitempty"`
	Timeout        time.Duration `json:"timeout,omitempty"`
	WaitSelector   string        `json:"waitSelector,omitempty"`

	// Landscape 设置横向（横版）打印。
	// 默认值: false（纵向打印）
	Landscape *bool `json:"landscape,omitempty"`

	// DisplayHeaderFooter 是否显示页眉和页脚。
	// 默认值: false
	DisplayHeaderFooter *bool `json:"displayHeaderFooter,omitempty"`

	// PrintBackground 是否打印 CSS 背景色和背景图片。
	// 默认值: false
	PrintBackground *bool `json:"printBackground,omitempty"`

	// Scale 页面渲染的缩放比例。
	// 取值范围: 0.1 ~ 2
	// 默认值: 1.0
	Scale *float64 `json:"scale,omitempty"`

	// PaperWidth 纸张宽度，单位：英寸。
	// 默认值: 8.5
	PaperWidth *float64 `json:"paperWidth,omitempty"`

	// PaperHeight 纸张高度，单位：英寸。
	// 默认值: 11.0
	PaperHeight *float64 `json:"paperHeight,omitempty"`

	// MarginTop 上边距，单位：英寸。
	// 默认值: 1cm ≈ 0.3937007874 英寸
	MarginTop *float64 `json:"marginTop,omitempty"`

	// MarginBottom 下边距，单位：英寸。
	// 默认值: 1cm ≈ 0.3937007874 英寸
	MarginBottom *float64 `json:"marginBottom,omitempty"`

	// MarginLeft 左边距，单位：英寸。
	// 默认值: 1cm ≈ 0.3937007874 英寸
	MarginLeft *float64 `json:"marginLeft,omitempty"`

	// MarginRight 右边距，单位：英寸。
	// 默认值: 1cm ≈ 0.3937007874 英寸
	MarginRight *float64 `json:"marginRight,omitempty"`

	// PageRanges 要打印的页码范围，格式如 "1-5, 8, 11-13"。
	// 空字符串表示打印所有页面。
	// 默认值: ""（全部页面）
	PageRanges *string `json:"pageRanges,omitempty"`

	// HeaderTemplate 页眉的 HTML 模板。
	// 模板内可使用以下 CSS 类名，Chrome 会自动替换为实际值：
	//   .date        - 格式化的打印日期
	//   .title       - 文档标题（<title> 标签内容）
	//   .url         - 页面 URL
	//   .pageNumber  - 当前页码
	//   .totalPages  - 总页数
	// 示例: `<span class="title"></span> — <span class="pageNumber"></span>`
	// 注意: 必须同时设置 DisplayHeaderFooter = true 才生效。
	// 默认值: ""（使用 Chrome 默认页眉）
	HeaderTemplate *string `json:"headerTemplate,omitempty"`

	// FooterTemplate 页脚的 HTML 模板。
	// 规则同 HeaderTemplate。
	// 默认值: ""（使用 Chrome 默认页脚）
	FooterTemplate *string `json:"footerTemplate,omitempty"`

	// PreferCSSPageSize 是否优先使用 CSS @page 规则中定义的页面尺寸，
	// 而非 PaperWidth / PaperHeight 参数。
	// 默认值: false
	PreferCSSPageSize *bool `json:"preferCSSPageSize,omitempty"`

	// TransferMode 指定 PDF 返回方式。
	// 默认值: ReturnAsBase64
	// 可选值:
	//   TransferModeReturnAsBase64 - 以 base64 编码字符串返回
	//   TransferModeReturnAsStream - 以流（IOStreamHandle）方式返回（适合大文件）
	TransferMode *string `json:"transferMode,omitempty"`

	// GenerateTaggedPDF 是否生成带标签（无障碍/辅助功能友好）的 PDF。
	// 默认值: false
	GenerateTaggedPDF *bool `json:"generateTaggedPDF,omitempty"`

	// GenerateDocumentOutline 是否在 PDF 中生成文档大纲/书签。
	// 默认值: false
	GenerateDocumentOutline *bool `json:"generateDocumentOutline,omitempty"`
}

const (
	// Letter: 8.5 x 11 英寸（北美标准）
	LetterWidth  = 8.5
	LetterHeight = 11.0

	// Legal: 8.5 x 14 英寸
	LegalWidth  = 8.5
	LegalHeight = 14.0

	// Tabloid: 11 x 17 英寸
	TabloidWidth  = 11.0
	TabloidHeight = 17.0

	// A4: 210 x 297 mm → 8.2677 x 11.6929 英寸（国际标准）
	A4Width  = 8.2677165354
	A4Height = 11.6929133858

	// A3: 297 x 420 mm → 11.6929 x 16.5354 英寸
	A3Width  = 11.6929133858
	A3Height = 16.5354330709

	// A5: 148 x 210 mm → 5.8268 x 8.2677 英寸
	A5Width  = 5.8267716535
	A5Height = 8.2677165354
)

func (c *Config) ParsePaperPreset(name string) error {

	name = strings.TrimSpace(strings.ToLower(name))

	if name == "" {
		name = "a4"
	}

	switch name {

	case "letter":
		c.PaperWidth = float64Ptr(LetterWidth)
		c.PaperHeight = float64Ptr(LetterHeight)
		return nil

	case "legal":
		c.PaperWidth = float64Ptr(LegalWidth)
		c.PaperHeight = float64Ptr(LegalHeight)
		return nil

	case "tabloid":
		c.PaperWidth = float64Ptr(TabloidWidth)
		c.PaperHeight = float64Ptr(TabloidHeight)
		return nil

	case "a4":
		c.PaperWidth = float64Ptr(A4Width)
		c.PaperHeight = float64Ptr(A4Height)
		return nil

	case "a3":
		c.PaperWidth = float64Ptr(A3Width)
		c.PaperHeight = float64Ptr(A3Height)
		return nil

	case "a5":
		c.PaperWidth = float64Ptr(A5Width)
		c.PaperHeight = float64Ptr(A5Height)
		return nil

	default:
		return fmt.Errorf("unknown paper preset: %s", name)
	}

}

func (c *Config) Validate() error {
	rawURL := strings.TrimSpace(c.URL)
	htmlFile := strings.TrimSpace(c.HTMLFile)

	switch {
	case rawURL == "" && htmlFile == "":
		return fmt.Errorf("must specify either URL or HTMLFile")
	case rawURL != "" && htmlFile != "":
		return fmt.Errorf("cannot specify both URL and HTMLFile")
	}

	if rawURL != "" {
		parsed, err := neturl.ParseRequestURI(rawURL)
		if err != nil {
			return fmt.Errorf("invalid URL %q: %w", rawURL, err)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return fmt.Errorf("URL scheme must be http or https, got %q", parsed.Scheme)
		}
	}

	if htmlFile != "" {
		info, err := os.Stat(htmlFile)
		if err != nil {
			return fmt.Errorf("invalid HTMLFile %q: %w", htmlFile, err)
		}
		if info.IsDir() {
			return fmt.Errorf("HTMLFile %q is a directory", htmlFile)
		}
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	if c.Scale != nil && (*c.Scale < 0.1 || *c.Scale > 2.0) {
		return fmt.Errorf("scale must be between 0.1 and 2.0")
	}

	if (c.PaperWidth == nil) != (c.PaperHeight == nil) {
		return fmt.Errorf("paperWidth and paperHeight must be set together")
	}
	if c.PaperWidth == nil || c.PaperHeight == nil {
		return fmt.Errorf("paperWidth and paperHeight must both be set")
	}
	if *c.PaperWidth <= 0 {
		return fmt.Errorf("paperWidth must be greater than 0")
	}
	if *c.PaperHeight <= 0 {
		return fmt.Errorf("paperHeight must be greater than 0")
	}

	if c.MarginTop != nil && *c.MarginTop < 0 {
		return fmt.Errorf("marginTop must be greater than or equal to 0")
	}
	if c.MarginBottom != nil && *c.MarginBottom < 0 {
		return fmt.Errorf("marginBottom must be greater than or equal to 0")
	}
	if c.MarginLeft != nil && *c.MarginLeft < 0 {
		return fmt.Errorf("marginLeft must be greater than or equal to 0")
	}
	if c.MarginRight != nil && *c.MarginRight < 0 {
		return fmt.Errorf("marginRight must be greater than or equal to 0")
	}

	hasHeader := c.HeaderTemplate != nil && strings.TrimSpace(*c.HeaderTemplate) != ""
	hasFooter := c.FooterTemplate != nil && strings.TrimSpace(*c.FooterTemplate) != ""
	if (hasHeader || hasFooter) && (c.DisplayHeaderFooter == nil || !*c.DisplayHeaderFooter) {
		return fmt.Errorf("displayHeaderFooter must be true when headerTemplate or footerTemplate is set")
	}

	if c.TransferMode != nil {
		switch *c.TransferMode {
		case "ReturnAsBase64", "ReturnAsStream":
		default:
			return fmt.Errorf("invalid transferMode %q", *c.TransferMode)
		}
	}

	return nil
}

func (c *Config) PrepareOutputPath() (string, error) {
	outputFile := c.OutputFile
	if outputFile == "" {
		outputFile = "./output.pdf"
	}
	// 创建目录
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return "", err
	}
	// 获取绝对路径
	return filepath.Abs(outputFile)
}

func (c *Config) InputTarget() (string, error) {
	rawURL := strings.TrimSpace(c.URL)
	htmlFile := strings.TrimSpace(c.HTMLFile)

	switch {
	case rawURL != "":
		return rawURL, nil
	case htmlFile == "":
		return "", fmt.Errorf("must specify either URL or HTMLFile")
	}

	absPath, err := filepath.Abs(htmlFile)
	if err != nil {
		return "", fmt.Errorf("resolve HTMLFile %q: %w", htmlFile, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("stat HTMLFile %q: %w", absPath, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("HTMLFile %q is a directory", absPath)
	}

	filePath := filepath.ToSlash(absPath)
	if runtime.GOOS == "windows" && !strings.HasPrefix(filePath, "/") {
		filePath = "/" + filePath
	}

	u := neturl.URL{
		Scheme: "file",
		Path:   filePath,
	}

	return u.String(), nil
}

func float64Ptr(v float64) *float64 { return &v }
