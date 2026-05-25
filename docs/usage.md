# 使用教程

html2pdf-chrome 提供三种使用方式：CLI 命令行、HTTP 服务、Go 库集成。

---

## 1. CLI 命令行

适合本地使用、脚本调用、一次性转换。

### 基本用法

```bash
# 在线页面转 PDF
html2pdf-chrome -url https://example.com -out output.pdf

# 本地 HTML 文件转 PDF
html2pdf-chrome -html-file ./report.html -out report.pdf
```

### 纸张和方向

```bash
# A3 横向
html2pdf-chrome -url https://example.com -paper a3 -landscape -out output.pdf

# Letter 纸张
html2pdf-chrome -url https://example.com -paper letter -out output.pdf
```

支持的纸张预设：`a3`、`a4`（默认）、`a5`、`letter`、`legal`、`tabloid`

### 边距

单位为英寸，默认约 1cm（0.394 英寸）。

```bash
# 窄边距
html2pdf-chrome -url https://example.com \
  -margin-top 0.2 -margin-bottom 0.2 \
  -margin-left 0.2 -margin-right 0.2 \
  -out output.pdf

# 无边距（全出血）
html2pdf-chrome -url https://example.com \
  -margin-top 0 -margin-bottom 0 \
  -margin-left 0 -margin-right 0 \
  -out output.pdf
```

### 背景和缩放

```bash
# 打印 CSS 背景色和背景图
html2pdf-chrome -url https://example.com -print-background -out output.pdf

# 缩小到 80%
html2pdf-chrome -url https://example.com -scale 0.8 -out output.pdf
```

### 页眉页脚

需要同时开启 `-display-header-footer`。模板中可用的 CSS 类：

- `.date` — 打印日期
- `.title` — 文档标题
- `.url` — 页面 URL
- `.pageNumber` — 当前页码
- `.totalPages` — 总页数

```bash
html2pdf-chrome -url https://example.com \
  -display-header-footer \
  -header-template '<div style="font-size:10px;width:100%;text-align:center;"><span class="title"></span></div>' \
  -footer-template '<div style="font-size:9px;width:100%;text-align:center;"><span class="pageNumber"></span> / <span class="totalPages"></span></div>' \
  -out output.pdf
```

### 页码范围

```bash
# 只打印第 1-3 页和第 5 页
html2pdf-chrome -url https://example.com -page-ranges "1-3, 5" -out output.pdf
```

### 等待策略

```bash
# 等待网络空闲（适合有异步加载的页面）
html2pdf-chrome -url https://example.com -wait-network-idle -out output.pdf

# 自定义静默期（默认 500ms）
html2pdf-chrome -url https://example.com -wait-network-idle -network-idle-time 1s -out output.pdf

# 等待某个元素出现
html2pdf-chrome -url https://example.com -wait-selector "#content-loaded" -out output.pdf

# 等待自定义 JS 条件
html2pdf-chrome -url https://example.com -wait-expression "window.__RENDER_DONE === true" -out output.pdf

# 组合使用
html2pdf-chrome -url https://example.com \
  -wait-network-idle \
  -wait-selector ".chart-container" \
  -wait-expression "window.chartsReady" \
  -out output.pdf
```

### 其他选项

```bash
# 生成无障碍 PDF（带标签）
html2pdf-chrome -url https://example.com -generate-tagged-pdf -out output.pdf

# 嵌入文档大纲
html2pdf-chrome -url https://example.com -generate-document-outline -out output.pdf

# 让 CSS @page size 覆盖 CLI 纸张设置
html2pdf-chrome -url https://example.com -prefer-css-page-size -out output.pdf

# 设置超时（默认 45s）
html2pdf-chrome -url https://example.com -timeout 60s -out output.pdf

# 查看 Chrome 调试日志
html2pdf-chrome -url https://example.com -chrome-debug-log -out output.pdf

# 指定 Chrome 路径
html2pdf-chrome -url https://example.com -chrome-path /usr/bin/chromium -out output.pdf

# 禁用 sandbox（Docker 容器中需要）
html2pdf-chrome -url https://example.com -no-sandbox -out output.pdf
```

### 完整参数列表

```bash
html2pdf-chrome -help
```

---

## 2. HTTP 服务

适合长期运行、高并发、服务端集成。内部维护 Chrome 实例池，避免重复启动浏览器。

### 启动服务

```bash
# 直接运行
html2pdf-server -addr :8080 -max-instances 4 -min-instances 2 -no-sandbox

# 指定 Chrome 路径
html2pdf-server -chrome-path /usr/bin/chromium -no-sandbox -addr :8080

# 也可以用环境变量
CHROME_PATH=/usr/bin/chromium html2pdf-server -no-sandbox -addr :8080

# Docker 运行
docker run -d \
  --name html2pdf \
  --restart always \
  --shm-size=512m \
  -p 8080:8080 \
  html2pdf-chrome
```

服务启动参数：

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-addr` | `:8080` | 监听地址 |
| `-chrome-path` | 自动查找 | Chrome/Chromium 可执行文件路径 |
| `-max-instances` | `4` | 最大 Chrome 实例数 |
| `-min-instances` | `2` | 最小空闲实例数 |
| `-no-sandbox` | `false` | 禁用 Chrome sandbox |
| `-output-dir` | `/tmp/html2pdf` | 临时文件目录 |

### 接口：POST /convert

请求 Content-Type 为 `application/json`，返回 PDF 二进制文件。

#### 请求参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `url` | string | — | 页面 URL（与 html 二选一） |
| `html` | string | — | HTML 内容（与 url 二选一） |
| `paper` | string | `"a4"` | 纸张：a3 / a4 / a5 / letter / legal / tabloid |
| `landscape` | bool | false | 横向 |
| `printBackground` | bool | false | 打印 CSS 背景 |
| `displayHeaderFooter` | bool | false | 显示页眉页脚 |
| `headerTemplate` | string | — | 页眉 HTML 模板 |
| `footerTemplate` | string | — | 页脚 HTML 模板 |
| `scale` | number | 1.0 | 缩放（0.1~2.0） |
| `marginTop` | number | 0.394 | 上边距（英寸） |
| `marginBottom` | number | 0.394 | 下边距（英寸） |
| `marginLeft` | number | 0.394 | 左边距（英寸） |
| `marginRight` | number | 0.394 | 右边距（英寸） |
| `pageRanges` | string | — | 页码范围，如 "1-3, 5" |
| `preferCSSPageSize` | bool | false | CSS @page size 优先 |
| `generateTaggedPDF` | bool | false | 无障碍 PDF |
| `generateDocumentOutline` | bool | false | 文档大纲 |
| `waitNetworkIdle` | bool | false | 等待网络空闲 |
| `waitExpression` | string | — | JS 等待条件 |
| `waitSelector` | string | — | CSS 选择器等待 |
| `timeout` | string | `"45s"` | 超时（如 "30s"、"1m"） |

#### 调用示例

```bash
# 最简单：URL 转 PDF
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}' \
  -o output.pdf

# HTML 内容转 PDF
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{"html": "<h1>标题</h1><p>正文</p>"}' \
  -o output.pdf

# A3 横向 + 打印背景
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "paper": "a3", "landscape": true, "printBackground": true}' \
  -o output.pdf

# 带页眉页脚
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "displayHeaderFooter": true,
    "headerTemplate": "<div style=\"font-size:10px;width:100%;text-align:center;\"><span class=\"title\"></span></div>",
    "footerTemplate": "<div style=\"font-size:9px;width:100%;text-align:center;\">第 <span class=\"pageNumber\"></span> 页 / 共 <span class=\"totalPages\"></span> 页</div>"
  }' \
  -o output.pdf

# 自定义边距 + 缩放
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "scale": 0.8,
    "marginTop": 0.5,
    "marginBottom": 0.5,
    "marginLeft": 0.3,
    "marginRight": 0.3
  }' \
  -o output.pdf

# 等待异步内容加载完成
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "waitNetworkIdle": true,
    "waitExpression": "document.querySelector(\".content\") !== null",
    "timeout": "60s"
  }' \
  -o output.pdf

# 只打印前 3 页
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "pageRanges": "1-3"}' \
  -o output.pdf
```

### 接口：GET /health

返回服务状态和实例池信息。

```bash
curl http://localhost:8080/health
```

响应：

```json
{
  "status": "ok",
  "pool": {
    "idle": 2,
    "active": 1,
    "total": 3
  }
}
```

---

## 3. Go 库集成

适合在自己的 Go 项目中直接调用。

### 安装

```bash
go get github.com/PiZhai/html2pdf-chrome
```

### 单次调用（每次新建 Chrome 实例）

```go
package main

import (
    "log"
    "github.com/PiZhai/html2pdf-chrome/pkg/html2pdf"
)

func main() {
    err := html2pdf.Convert(html2pdf.Request{
        URL:        "https://example.com",
        OutputPath: "./output.pdf",
        Options: html2pdf.Options{
            Paper:           html2pdf.PaperA4,
            PrintBackground: true,
            WaitNetworkIdle: true,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

### 池化调用（复用 Chrome 实例，适合服务端）

```go
package main

import (
    "log"
    "sync"
    "time"
    "github.com/PiZhai/html2pdf-chrome/pkg/html2pdf"
)

func main() {
    converter, err := html2pdf.NewConverter(html2pdf.ConverterConfig{
        MaxInstances: 4,
        MinInstances: 2,
        NoSandbox:    true, // Docker 环境
    })
    if err != nil {
        log.Fatal(err)
    }
    defer converter.Close()

    // 并发安全，可直接在多个 goroutine 中调用
    var wg sync.WaitGroup
    urls := []string{
        "https://example.com",
        "https://go.dev",
        "https://github.com",
    }

    for _, u := range urls {
        wg.Add(1)
        go func(url string) {
            defer wg.Done()
            err := converter.Convert(html2pdf.Request{
                URL:        url,
                OutputPath: "./output-" + sanitize(url) + ".pdf",
                Options: html2pdf.Options{
                    Timeout:         30 * time.Second,
                    PrintBackground: true,
                    WaitNetworkIdle: true,
                },
            })
            if err != nil {
                log.Printf("convert %s: %v", url, err)
            }
        }(u)
    }
    wg.Wait()
}

func sanitize(url string) string {
    // 简化示例，实际使用请做完整的文件名清理
    r := []byte(url)
    for i, b := range r {
        if b == '/' || b == ':' || b == '.' {
            r[i] = '_'
        }
    }
    return string(r)
}
```

### 便捷函数

```go
// 转换 URL
html2pdf.ConvertURL("https://example.com", "./output.pdf", html2pdf.Options{
    Paper: html2pdf.PaperA4,
})

// 转换本地文件
html2pdf.ConvertFile("./report.html", "./report.pdf", html2pdf.Options{
    PrintBackground: true,
})
```

### Options 完整字段

```go
type Options struct {
    ChromePath              string        // Chrome 路径（空=自动查找）
    Timeout                 time.Duration // 超时（默认 45s）
    WaitSelector            string        // 等待 CSS 选择器可见
    Paper                   PaperPreset   // 纸张预设
    WaitNetworkIdle         bool          // 等待网络空闲
    NetworkIdleTime         time.Duration // 网络空闲静默期（默认 500ms）
    WaitExpression          string        // 自定义 JS 等待条件
    Landscape               bool          // 横向
    DisplayHeaderFooter     bool          // 页眉页脚
    PrintBackground         bool          // CSS 背景
    Scale                   *float64      // 缩放（0.1~2.0）
    MarginTop               *float64      // 上边距（英寸）
    MarginBottom            *float64      // 下边距（英寸）
    MarginLeft              *float64      // 左边距（英寸）
    MarginRight             *float64      // 右边距（英寸）
    PageRanges              string        // 页码范围
    HeaderTemplate          string        // 页眉模板
    FooterTemplate          string        // 页脚模板
    PreferCSSPageSize       bool          // CSS @page size 优先
    GenerateTaggedPDF       bool          // 无障碍 PDF
    GenerateDocumentOutline bool          // 文档大纲
    TransferMode            TransferMode  // 传输模式（base64/stream）
    ChromeDebugLog          bool          // Chrome 调试日志
    NoSandbox               bool          // 禁用 sandbox
}
```

---

## 4. 等待策略选择指南

| 场景 | 推荐策略 |
|------|----------|
| 静态页面 | 不需要额外等待，默认即可 |
| 有图片/CSS/字体异步加载 | `-wait-network-idle` |
| SPA 应用（React/Vue 等） | `-wait-expression "window.__READY"` 或 `-wait-selector "#app"` |
| 图表库（ECharts/Chart.js） | `-wait-network-idle` + `-wait-expression "window.chartsReady"` |
| MathJax 数学公式 | `-wait-network-idle`（MathJax 会加载字体文件） |
| 有长轮询/SSE 的页面 | 不要用 network idle，改用 `-wait-expression` |
| 需要等待特定 DOM 元素 | `-wait-selector ".my-element"` |

等待顺序（固定）：

```
DOM ready → 字体加载 → 网络空闲 → CSS 选择器 → JS 表达式 → 导出 PDF
```

---

## 5. 常见问题

**Q: PDF 是空白的**
A: 页面可能还没渲染完就导出了。加 `-wait-network-idle` 或 `-wait-expression`。

**Q: 中文显示为方块**
A: 系统缺少中文字体。安装 `fonts-noto-cjk`（Linux）或确认系统有中文字体。

**Q: 数学公式不完整**
A: MathJax/KaTeX 需要加载字体文件，加 `-wait-network-idle` 等待加载完成。

**Q: 超时报错**
A: 页面太复杂或网络慢。增大 `-timeout`（如 `-timeout 120s`）。

**Q: Docker 中 Chrome 启动失败**
A: 加 `-no-sandbox` 参数，并确保 `--shm-size=512m`。

**Q: 页眉页脚不显示**
A: 必须同时设置 `-display-header-footer` 和模板参数。

**Q: 想用 CSS 里定义的页面大小**
A: 加 `-prefer-css-page-size`，此时 `-paper` 设置会被 CSS `@page { size: ... }` 覆盖。
