# Developer Guide

> 写给想要继续开发的贡献者——项目概览、架构详解、开发环境搭建、已知限制与改进方向。

## 最新提交

`de40946` — **docs: update code comments across all packages**

为所有源文件添加了完整的包级文档和函数级注释，遵循 Go 文档约定。共更新 11 个文件，新增 203 行，删除 65 行，影响以下包：

| 包 | 更新内容 |
|---|---|
| `pkg/html2pdf` | `Request`、`Options`、`ConverterConfig`、`Convert`、`NewConverter`、所有便捷函数和字段 |
| `internal/app` | `Run`、`RunWithPool` |
| `internal/browser` | `FindChrome`、`Launch`、`Instance`、`Close`、`pickFreePort` |
| `internal/cdp` | `Connect`、`OpenPage`、`WaitDocumentReady`、`WaitFontsReady`、`WaitForExpression` |
| `internal/render` | `PrintToFile`、`buildPrintToPDFParams`、所有 `Options` 字段 |
| `internal/pool` | 之前文档已经完善，无需变更 |

代码库现在具备良好的注释基础，可以直接开始开发。

---

## 1. 项目概述

**html2pdf-chrome** 是一个基于 Chrome/Chromium Headless + CDP（WebSocket）的 HTML 转 PDF 工具，完全使用 Go 语言编写。它通过 Chrome DevTools Protocol 驱动真实浏览器进程完成渲染，因而能够完美支持现代 CSS（如 Grid、Flexbox、Web Fonts）、JavaScript 动态渲染、Canvas、SVG 等 Web 平台能力。

### 核心优势

- **像素级还原**：由真实 Chrome/Chromium 引擎渲染，效果与浏览器预览完全一致
- **三种使用模式**：CLI 实参、HTTP 服务（JSON 接口）、Go 内嵌库
- **高并发**：内置连接池复用 Chrome 实例，服务端编程无需反复启动进程
- **灵活等待机制**：支持网络空闲、CSS 选择器可见、自定义 JS 表达式等多种页面就绪判断
- **完整 PDF 参数**：所有 `Page.printToPDF` 参数均已暴露（纸张、边距、缩放、页眉模板、无障碍标签……）
- **多平台支持**：macOS / Linux / Windows 自动路径探测，Docker 一键部署

### 技术栈

| 组件 | 技术 |
|---|---|
| 渲染引擎 | Chrome/Chromium Headless（`--headless=new`） |
| 驱动协议 | Chrome DevTools Protocol over WebSocket |
| CDP 客户端 | `github.com/chromedp/chromedp` + `github.com/chromedp/cdproto` |
| 并发原语 | `sync.Mutex`、`sync.Cond`、`runtime.NumCPU` |
| 语言 | Go 1.26 |
| 构建产物 | 静态二进制（`CGO_ENABLED=0`） |

---

## 2. 项目结构详解

```
html2pdf-chrome/
├── cmd/
│   ├── html2pdf-chrome/    # CLI 入口 — 解析 flag，调用 html2pdf.Convert
│   └── html2pdf-server/    # HTTP 服务入口 — 池化模式暴露 /convert、/health
├── pkg/html2pdf/           # 公开 API — Convert、Converter、类型定义
│   ├── types.go            # Request、Options、ConverterConfig、常量
│   ├── convert.go          # Convert、ConvertURL、ConvertFile、toConfig 转换逻辑
│   └── converter.go        # NewConverter、Converter（池化模式）
├── internal/
│   ├── app/                # 编排层 — 串联浏览器生命周期、CDP、渲染
│   │   ├── run.go          # 单次模式流程
│   │   └── run_pooled.go   # 池化模式流程
│   ├── config/             # 配置校验 — Config 结构体、Validate、InputTarget、ParsePaperPreset
│   ├── browser/            # Chrome 进程管理
│   │   ├── find.go         # 多平台路径查找
│   │   ├── launch.go       # 启动 Chrome 进程、等待 WS 端点
│   │   └── instance.go     # Instance 结构体
│   ├── cdp/                # CDP 协议交互
│   │   ├── connect.go      # 连接 Chrome、创建隔离 Tab
│   │   ├── target.go       # OpenPage 导航 + 等待链编排
│   │   ├── wait.go         # WaitDocumentReady、WaitFontsReady、WaitForExpression
│   │   └── wait_network.go # WaitNetworkIdle（Network 事件计数）
│   ├── pool/               # Chrome 实例池
│   │   ├── pool.go         # Pool 结构体、Acquire、Release、reaper、健康检查
│   │   └── instance.go     # PooledInstance 封装
│   └── render/             # PDF 导出
│       ├── options.go      # render.Options 参数结构
│       └── pdf.go          # PrintToFile、Page.printToPDF 调用、流式读取
├── docs/
│   ├── architecture.md     # 架构设计文档
│   ├── usage.md            # 使用教程（CLI / HTTP / Go 集成）
│   ├── docker-deployment.md# Docker 部署指南
│   └── developer-guide.md  # 本文档
├── testdata/               # 测试用 HTML 页面
├── Dockerfile              # 国际版（Google Chrome）
├── Dockerfile.cn           # 国内版（Chromium + 阿里云镜像）
├── go.mod / go.sum
└── README.md
```

### 模块依赖关系（由外到内）

```
cmd/*  ──调用──▶  pkg/html2pdf  ──调用──▶  internal/app
                                               │
                  internal/config ◀─────────────┘
                                               │
                  internal/browser ◀────────────┤
                  internal/cdp     ◀────────────┤
                  internal/pool    ◀────────────┤
                  internal/render  ◀────────────┘
```

**设计原则**：所有内部实现放在 `internal/` 下，外部依赖只能引用 `pkg/html2pdf/`。

---

## 3. 关键代码路径

### 3.1 单次转换（`html2pdf.Convert`）

每次调用启动一个新的 Chrome 进程，用完即销毁：

```
html2pdf.Convert(Request)
  → req.toConfig()                          // 公共类型 → 内部 Config，校验
  → app.Run(cfg)
       → cfg.Validate()                     // 二次校验
       → cfg.InputTarget()                  // URL 规范化 或 本地文件 → file://
       → cfg.PrepareOutputPath()            // 创建输出目录
       → browser.FindChrome(cfg.ChromePath) // 多平台查找 Chrome
       → browser.Launch(execPath, opts)     // 启动进程，等待 WS 端点
       → cdp.Connect(wsURL)                 // chromedp: 远程分配器 + 独立上下文
       → cdp.OpenPage(renderCtx, opts)      // Navigate + 等待链
       → render.PrintToFile(renderCtx, opts)// Page.printToPDF → 写入文件
       → instance.Close()                   // Kill 进程 + 删除临时目录
```

### 3.2 池化转换（`Converter.Convert`）

复用 Chrome 实例，每个任务用独立 Tab 隔离：

```
Converter.Convert(Request)
  → req.toConfig()
  → app.RunWithPool(pool, cfg)
       → pool.Acquire(ctx)                   // 获取健康实例（阻塞至可用）
       → cdp.Connect(inst.WebSocketURL())    // chromedp 创建独立 Tab
       → cdp.OpenPage(...)                   // 导航 + 等待（在独立的 Tab 中）
       → render.PrintToFile(...)             // PDF 渲染
       → cdpCancel()                         // 关闭 Tab（释放 Chrome 内存）
       → pool.Release(inst)                  // 归还实例或回收
```

### 3.3 等待策略链

```
Navigate
  → chromedp.WaitReady("body")              // body 元素存在
  → WaitDocumentReady(15s)                  // readyState === "complete"
  → WaitFontsReady(10s)                     // fonts.status === "loaded"
  → [可选] WaitNetworkIdle(idleTime, timeout)// 网络请求归零 + 静默期
  → [可选] chromedp.WaitVisible(selector)    // CSS 选择器可见
  → [可选] WaitForExpression(expr, timeout)  // 自定义 JS 为 truthy
  → printToPDF
```

### 3.4 实例池设计

```
┌──────────────────────────────────────────┐
│                Pool                       │
│                                           │
│  idle: []*PooledInstance     ← 空闲队列   │
│  activeCount: int            ← 使用中数量 │
│  totalCount: int             ← 存活总数   │
│  mu: sync.Mutex              ← 互斥锁     │
│  cond: *sync.Cond            ← 条件变量   │
│                                           │
│  Acquire(ctx):                              │
│    1. 从 idle 队列弹出健康实例               │
│    2. 若空闲为空且 total < max → 新建实例    │
│    3. 若已满 → cond.Wait() 阻塞             │
│                                           │
│  Release(inst):                             │
│    1. 若 taskCount >= MaxTasksPerInstance  → 销毁│
│    2. 若不健康 → 销毁                       │
│    3. 否则 → 放回 idle 队列，cond.Broadcast()│
│                                           │
│  reaper(): 后台 goroutine，定期回收超时空闲实例│
└──────────────────────────────────────────┘
```

---

## 4. 开发环境搭建

### 前置条件

- Go 1.26+
- Google Chrome 或 Chromium（本机安装）
- Git

### 克隆并构建

```bash
git clone https://github.com/PiZhai/html2pdf-chrome.git
cd html2pdf-chrome
go build ./...
```

### 运行测试

```bash
go test ./...
```

测试使用本机安装的 Chrome，不依赖外部服务。

### IDE 设置

- `.idea/` 已在 `.gitignore` 中排除
- 推荐使用 GoLand 或 VS Code + Go 插件，打开项目根目录即可
- Go modules 自动解析 `github.com/PiZhai/html2pdf-chrome`

---

## 5. Docker 开发

### 国际版（Google Chrome）

```bash
docker build -t html2pdf-chrome .
docker run --rm html2pdf-chrome -url https://example.com -out /app/output/out.pdf
```

### 国内版（Chromium + 阿里云镜像）

```bash
docker build -f Dockerfile.cn -t html2pdf-chrome:cn .
docker run --rm -p 8080:8080 html2pdf-chrome:cn
```

国内版默认启动 HTTP 服务模式，监听 `:8080`。

---

## 6. 公开 API 速查

### 核心类型

```go
// 单次转换请求
type Request struct {
    URL        string   // 在线页面（与 HTMLFile 二选一）
    HTMLFile   string   // 本地 HTML 文件
    OutputPath string   // 输出 PDF 路径
    Options    Options  // 渲染选项
}

// 转换器配置（池化模式）
type ConverterConfig struct {
    ChromePath          string
    MaxInstances        int           // 默认: runtime.NumCPU()
    MinInstances        int           // 默认: 1
    MaxTasksPerInstance int           // 默认: 100，0 表示无限
    IdleTimeout         time.Duration // 默认: 5min，0 表示不回收
    ChromeDebugLog      bool
    NoSandbox           bool
}
```

### 核心函数

```go
// 单次模式
html2pdf.Convert(req Request) error
html2pdf.ConvertURL(url, outputPath string, opts Options) error
html2pdf.ConvertFile(htmlFile, outputPath string, opts Options) error

// 池化模式
converter, err := html2pdf.NewConverter(cfg ConverterConfig)
converter.Convert(req Request) error
converter.Stats() ConverterStats
converter.Close() error
```

---

## 7. 已知限制与改进方向

### 当前限制

| 限制项 | 说明 |
|---|---|
| **测试覆盖** | 浏览器路径查找仅 macOS 实测；Linux/Windows 虽兼容但未经实机验证 |
| **错误处理** | 错误信息为扁平 wrap，未结构化分类（启动失败 vs 渲染超时 vs 写入失败） |
| **可观测性** | 缺少指标（metrics）、结构化日志、请求追踪 |
| **发布工程** | 无多平台交叉编译脚本、版本注入、校验和生成 |
| **内存管理** | 池中实例无显式内存限制，`MaxTasksPerInstance` 是间接保护 |
| **PDF 后处理** | 不支持 PDF 合并、加密、压缩等后处理 |

### 建议的开发优先级

1. **错误结构化分类**：定义 `ErrorCode` 枚举（`ErrBrowserNotFound`、`ErrLaunchTimeout`、`ErrRenderTimeout` 等），便于调用方判断错误类型
2. **metrics/可观测性**：添加 Prometheus 指标（请求数、延迟、实例数、错误率）、结构化 slog 日志
3. **Linux/Windows 验证**：在 Docker Linux 和 Windows 环境完成完整功能测试
4. **发布工程**：添加 `Makefile` / GoReleaser 配置，生成多平台二进制
5. **性能优化**：PDF 流读取支持并发写入、Chrome 进程启动时长优化
6. **增强服务端**：限流、请求队列、优先级调度、异步回调
7. **增加测试**：`internal/cdp` 包的单元测试（目前仅 network 测试）、集成测试覆盖

---

## 8. 开发调试技巧

### 查看 Chrome 调试日志

```bash
# CLI 模式
go run ./cmd/html2pdf-chrome -url https://example.com -chrome-debug-log

# Go 库模式
opts := html2pdf.Options{ChromeDebugLog: true}
```

### 使用 chromedp 调试 URL

在浏览器中直接访问 Chrome 调试端口（`http://localhost:<port>` 或 `http://localhost:<port>/json/version`），查看当前打开的 Tab 和页面状态。

### 临时目录清理

Chrome 的 `--user-data-dir` 设为临时目录，位于：
- macOS: `/var/folders/.../T/html2pdf-chrome-profile-*`
- Linux: `/tmp/html2pdf-chrome-profile-*`

Instance 的 `Close()` 方法会自动清理。若异常中断，可手动删除。

### 网络空闲调试

若页面有长轮询（SSE、WebSocket），`WaitNetworkIdle` 可能永不触发（WebSocket 已被忽略，但长轮询 XHR 不会被忽略）。此时应使用 `WaitExpression`：

```bash
-wait-expression "document.querySelector('.content-loaded') !== null"
```

---

## 9. 贡献指南

1. Fork 本仓库，从 `main` 分支创建特性分支
2. 遵循 Go 代码风格（`gofmt`、`go vet`）
3. 为新功能添加测试
4. 更新受影响的文档（`docs/` 和 `README.md`）
5. 确保 `go test ./...` 全部通过
6. PR 标题使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式（如 `feat:`、`fix:`、`docs:`、`test:`）

---

如需对任何模块进行深入分析或有具体的开发问题，欢迎查看其他文档：

- [架构设计](./architecture.md)
- [使用教程](./usage.md)
- [Docker 部署指南](./docker-deployment.md)
