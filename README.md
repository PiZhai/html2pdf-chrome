# html2pdf-chrome

一个基于 `Chrome/Chromium Headless + CDP(WebSocket)` 的 HTML 转 PDF 工具原型，使用 Go 实现。

当前版本已经打通最小主链路：

- 启动本机 Chrome/Chromium
- 通过 CDP 连接浏览器
- 打开 URL 或本地 HTML 文件
- 等待页面和字体加载完成
- 导出 PDF 到本地文件
- 支持常用 `Page.printToPDF` 打印参数
- 提供最小自动化测试覆盖

## 当前状态

这是一个可运行的 `v0.1.0` 初始版本。

已完成：

- CLI 入口
- 配置校验与输入规范化
- Chrome 查找与启动
- CDP 连接
- 页面导航与等待
- PDF 导出
- 常用打印参数下沉
- Chrome 调试日志开关
- 最小自动化测试

尚未完成：

- Linux / Windows 实机验证
- 发布产物打包与三端分发
- 更稳定的等待策略，例如网络空闲和自定义 JS 条件

## 运行要求

- Go `1.26+`
- 本机已安装 `Google Chrome` 或 `Chromium`

## 快速开始

构建：

```bash
go build ./...
```

运行测试：

```bash
go test ./...
```

把本地 HTML 转成 PDF：

```bash
go run ./cmd/html2pdf-chrome -html-file ./testdata/sample.html -out ./output.pdf
```

把在线页面转成 PDF：

```bash
go run ./cmd/html2pdf-chrome -url https://example.com -out ./output.pdf
```

如果需要查看 Chrome 启动日志：

```bash
go run ./cmd/html2pdf-chrome -html-file ./testdata/sample.html -chrome-debug-log
```

## 常用参数

- `-url`：要打印的 HTTP/HTTPS 页面
- `-html-file`：要打印的本地 HTML 文件
- `-out`：输出 PDF 路径
- `-chrome-path`：显式指定 Chrome/Chromium 可执行文件
- `-timeout`：整体渲染超时
- `-wait-selector`：等待某个 CSS 选择器出现后再打印
- `-paper`：纸张预设，支持 `letter`、`legal`、`tabloid`、`a3`、`a4`、`a5`
- `-landscape`：横向打印
- `-display-header-footer`：显示页眉页脚
- `-print-background`：打印 CSS 背景
- `-scale`：页面渲染缩放比例
- `-margin-top` / `-margin-bottom` / `-margin-left` / `-margin-right`：页边距，单位英寸
- `-page-ranges`：打印页码范围，例如 `1-3, 5`
- `-header-template` / `-footer-template`：PDF 页眉页脚 HTML 模板
- `-generate-tagged-pdf`：生成带标签的无障碍 PDF
- `-generate-document-outline`：在 PDF 中嵌入文档大纲
- `-transfer-mode`：PDF 传输模式，支持 `base64` 或 `stream`
- `-prefer-css-page-size`：允许页面里的 `@page size` 覆盖 CLI 纸张设置
- `-chrome-debug-log`：输出 Chrome 进程调试日志到 `stderr`

## 目录结构

```text
cmd/html2pdf-chrome/   CLI 入口
internal/app/          主流程编排
internal/config/       配置、校验、路径规范化
internal/browser/      Chrome 查找、启动、关闭
internal/cdp/          CDP 连接、导航、等待
internal/render/       PDF 导出
testdata/              本地测试 HTML
docs/                  架构说明
```

## 已知限制

- 目前每次运行都会启动一个新的 Chrome 实例
- 目前只验证了最小单任务流程
- 浏览器路径发现已经兼容多平台路径规则，但尚未完成三端实测
- 复杂页面的稳定性仍依赖后续增强，例如网络空闲判断、自定义等待条件和更完整的错误分类
