# html2pdf-chrome

基于 Chrome/Chromium Headless + CDP (WebSocket) 的 HTML 转 PDF 工具，Go 实现。

提供 CLI 和 Go 库两种使用方式。支持实例池化，适用于高并发服务场景。

## 运行要求

- Go 1.26+
- 本机已安装 Google Chrome 或 Chromium

## 快速开始

构建：

```bash
go build ./...
```

运行测试：

```bash
go test ./...
```

### CLI 使用

```bash
# 本地 HTML 转 PDF
go run ./cmd/html2pdf-chrome -html-file ./testdata/sample.html -out ./output.pdf

# 在线页面转 PDF
go run ./cmd/html2pdf-chrome -url https://example.com -out ./output.pdf

# 等待网络空闲后再导出（适合有异步加载的页面）
go run ./cmd/html2pdf-chrome -url https://example.com -wait-network-idle -out ./output.pdf

# 等待自定义 JS 条件满足后再导出
go run ./cmd/html2pdf-chrome -url https://example.com -wait-expression "window.__RENDER_DONE === true" -out ./output.pdf
```

### Go 库使用（单次调用）

```go
import "github.com/PiZhai/html2pdf-chrome/pkg/html2pdf"

err := html2pdf.Convert(html2pdf.Request{
    URL:        "https://example.com",
    OutputPath: "./output.pdf",
    Options: html2pdf.Options{
        Paper:           html2pdf.PaperA4,
        PrintBackground: true,
    },
})
```

每次调用启动一个新 Chrome 实例，完成后关闭。适合低频调用或 CLI 场景。

### Go 库使用（实例池，适合服务端）

```go
import (
    "sync"
    "github.com/PiZhai/html2pdf-chrome/pkg/html2pdf"
)

// 创建转换器，内部维护 Chrome 实例池
converter, err := html2pdf.NewConverter(html2pdf.ConverterConfig{
    MaxInstances: 4,   // 最多 4 个 Chrome 进程
    MinInstances: 2,   // 保持 2 个热实例
})
if err != nil {
    log.Fatal(err)
}
defer converter.Close()

// 并发安全，可直接在 HTTP handler 中调用
err = converter.Convert(html2pdf.Request{
    URL:        "https://example.com",
    OutputPath: "./output.pdf",
})
```

池化模式下每个任务使用独立浏览器 Tab，任务结束后关闭 Tab，不会有状态泄漏。

## CLI 参数

输入输出：

- `-url` — HTTP/HTTPS 页面地址
- `-html-file` — 本地 HTML 文件路径
- `-out` — 输出 PDF 路径（默认 `output.pdf`）
- `-chrome-path` — 显式指定 Chrome 可执行文件路径

等待策略：

- `-wait-selector` — 等待指定 CSS 选择器可见后再打印
- `-wait-network-idle` — 等待网络空闲后再打印
- `-network-idle-time` — 网络空闲静默期（默认 500ms）
- `-wait-expression` — 自定义 JS 表达式，轮询直到返回 truthy

打印参数：

- `-paper` — 纸张预设：`letter`、`legal`、`tabloid`、`a3`、`a4`、`a5`
- `-landscape` — 横向打印
- `-print-background` — 打印 CSS 背景
- `-scale` — 渲染缩放（0.1 ~ 2.0）
- `-margin-top` / `-margin-bottom` / `-margin-left` / `-margin-right` — 页边距（英寸）
- `-page-ranges` — 页码范围，如 `1-3, 5`
- `-display-header-footer` — 显示页眉页脚
- `-header-template` / `-footer-template` — 页眉页脚 HTML 模板
- `-prefer-css-page-size` — 允许 CSS `@page size` 覆盖纸张设置
- `-generate-tagged-pdf` — 生成无障碍 PDF
- `-generate-document-outline` — 嵌入文档大纲
- `-transfer-mode` — PDF 传输模式：`base64` 或 `stream`

调试：

- `-chrome-debug-log` — 输出 Chrome 进程日志到 stderr
- `-timeout` — 整体超时（默认 45s）

## 等待策略

页面导出前的等待顺序：

1. `body` ready
2. `document.readyState === "complete"`
3. `document.fonts.status === "loaded"`
4. 网络空闲（可选，`-wait-network-idle`）
5. CSS 选择器可见（可选，`-wait-selector`）
6. 自定义 JS 表达式为 truthy（可选，`-wait-expression`）

网络空闲通过 CDP Network 事件跟踪 inflight 请求数，归零后保持静默期即认为空闲。会忽略 WebSocket、`data:`、`blob:` 请求。

如果页面有长轮询或 SSE，网络空闲可能永远不会触发，此时由 `-timeout` 兜底。这种场景建议用 `-wait-expression` 替代。

## 实例池配置

`ConverterConfig` 字段：

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `ChromePath` | 自动查找 | Chrome 可执行文件路径 |
| `MaxInstances` | `runtime.NumCPU()` | 池中最大 Chrome 实例数 |
| `MinInstances` | 1 | 最小空闲实例数（启动时预热） |
| `MaxTasksPerInstance` | 100 | 单实例最大任务数，达到后回收 |
| `IdleTimeout` | 5 分钟 | 空闲实例超时回收时间 |
| `ChromeDebugLog` | false | Chrome 调试日志 |
| `NoSandbox` | false | 禁用 sandbox（Docker 容器中需要） |

池满时 `Acquire` 会阻塞等待，受调用方 context 超时控制。

## HTTP 服务模式

项目提供 `html2pdf-server` 命令，以 HTTP 服务方式长期运行，内部使用实例池：

```bash
# 直接运行
html2pdf-server -addr :8080 -max-instances 4 -min-instances 2 -no-sandbox

# Docker 运行（Dockerfile.cn 默认启动服务模式）
docker run -d --name html2pdf --restart always --shm-size=512m -p 8080:8080 html2pdf-chrome
```

接口：

- `POST /convert` — 接收 JSON，返回 PDF 文件
- `GET /health` — 返回池状态

详见 [使用教程](./docs/usage.md)。

## 目录结构

```text
cmd/html2pdf-chrome/   CLI 入口
cmd/html2pdf-server/   HTTP 服务入口（长期运行，实例池模式）
pkg/html2pdf/          对外 Go API（Convert、Converter）
internal/app/          主流程编排
internal/config/       配置、校验、路径规范化
internal/browser/      Chrome 查找、启动、关闭
internal/cdp/          CDP 连接、导航、等待策略
internal/pool/         Chrome 实例池
internal/render/       PDF 导出
testdata/              测试用 HTML
docs/                  文档（架构、使用教程、Docker 部署、开发者指南）
```

## 已知限制

- 浏览器路径发现兼容 macOS / Linux / Windows，已在 macOS 和 Linux (Docker) 上验证
- 复杂页面稳定性依赖等待策略的正确选择
- 错误信息目前是扁平的 wrap，没有结构化分类
- 没有发布工程（多平台构建、版本注入）

## 贡献 / 开发

详见 [开发者指南](./docs/developer-guide.md)。
