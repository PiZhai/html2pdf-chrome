# Architecture

## 技术选型

- 渲染引擎：Chrome/Chromium Headless
- 驱动方式：Chrome DevTools Protocol (WebSocket)
- 实现语言：Go
- CDP 客户端：chromedp

## 两条执行路径

### 单次模式（`Convert`）

每次调用启动新 Chrome 实例，完成后关闭。适合 CLI 和低频场景。

```text
Convert(Request)
  -> toConfig
  -> app.Run
     -> config.Validate / InputTarget / PrepareOutputPath
     -> browser.FindChrome
     -> browser.Launch          ← 启动新进程
     -> cdp.Connect
     -> cdp.OpenPage
     -> render.PrintToFile
     -> instance.Close()        ← 关闭进程
```

### 池化模式（`Converter`）

复用 Chrome 实例，每个任务使用独立 Tab。适合服务端并发场景。

```text
NewConverter(ConverterConfig)
  -> pool.New
     -> browser.FindChrome
     -> 预热 MinInstances 个 Chrome 进程

Converter.Convert(Request)
  -> toConfig
  -> app.RunWithPool
     -> pool.Acquire            ← 从池中获取实例
     -> cdp.Connect             ← 创建独立 Tab
     -> cdp.OpenPage
     -> render.PrintToFile
     -> cdpCancel()             ← 关闭 Tab
     -> pool.Release            ← 归还实例

Converter.Close()
  -> pool.Close
     -> 等待所有活跃任务完成
     -> 关闭所有 Chrome 进程
```

## 模块职责

### `cmd/html2pdf-chrome`

CLI 入口。解析 flag，组装 Request，调用 `Convert`。

### `cmd/html2pdf-server`

HTTP 服务入口。启动实例池，暴露 `POST /convert` 和 `GET /health` 接口。
适合 Docker 长期运行部署。

### `pkg/html2pdf`

公开 API 层。

- `Convert` / `ConvertURL` / `ConvertFile` — 单次模式
- `NewConverter` / `Converter` — 池化模式
- `Request` / `Options` / `ConverterConfig` — 公开类型

### `internal/config`

配置与校验。参数结构定义、纸张预设解析、输入目标规范化、输出路径创建。

### `internal/app`

编排层。

- `Run` — 单次模式编排（启动 → 连接 → 渲染 → 关闭）
- `RunWithPool` — 池化模式编排（获取 → 连接 → 渲染 → 归还）

### `internal/browser`

Chrome 进程管理。查找可执行文件（多平台候选路径 + PATH + 环境变量）、启动 headless 实例、获取 WebSocket 地址、关闭并清理临时目录。

### `internal/cdp`

CDP 协议层。

- `Connect` — 通过 WebSocket 连接浏览器
- `OpenPage` — 导航 + 等待策略编排
- `WaitDocumentReady` / `WaitFontsReady` — 基础等待
- `WaitNetworkIdle` — 网络空闲检测（CDP Network 事件）
- `WaitForExpression` — 自定义 JS 条件轮询

### `internal/pool`

Chrome 实例池。

- `Pool.Acquire` — 获取健康实例，池满时阻塞
- `Pool.Release` — 归还实例，超限时回收
- 后台 reaper 回收超时空闲实例
- 健康检查通过 HTTP `/json/version` 探测

### `internal/render`

PDF 导出。构建 `Page.printToPDF` 参数，支持 base64 和 stream 两种传输模式。

## 等待策略

```text
Navigate
  → body ready
  → document.readyState === "complete"
  → document.fonts.status === "loaded"
  → [可选] 网络空闲（CDP Network 事件，静默期可配）
  → [可选] CSS selector 可见
  → [可选] 自定义 JS 表达式为 truthy
  → printToPDF
```

网络空闲检测原理：监听 `Network.requestWillBeSent` / `LoadingFinished` / `LoadingFailed`，维护 inflight 计数器，归零后保持静默期即认为空闲。忽略 WebSocket、data:、blob: 请求。

## 实例池设计

```text
┌─────────────────────────────────────┐
│              Pool                    │
│                                     │
│  idle: []*PooledInstance            │
│  activeCount: int                   │
│  totalCount: int                    │
│  mu: sync.Mutex                     │
│  cond: *sync.Cond                   │
│                                     │
│  Acquire(ctx) → 获取或创建实例       │
│  Release(inst) → 归还或回收          │
│  reaper() → 定期清理超时空闲实例     │
└─────────────────────────────────────┘
```

关键决策：

- 创建实例前先预留 slot（totalCount++），防止并发超限
- 健康检查在 Acquire 时执行，不健康则丢弃重建
- MaxTasksPerInstance 防止 Chrome 内存泄漏累积
- 每个任务使用独立 Tab（chromedp NewContext），Tab 关闭时 Chrome 回收其资源

## 设计特点

- CLI 和 Go 库共用同一条内部渲染链路
- 池化是 opt-in，不影响原有 `Convert` 行为
- 每次单次调用使用独立 user-data-dir，避免状态污染
- 默认静默运行，`-chrome-debug-log` 开启调试输出
- `-paper` 与 `-prefer-css-page-size` 分离，避免静默覆盖

## 待做

1. Windows 实机验证
2. 错误分类与可观测性（启动失败 vs 渲染超时 vs 写入失败）
3. 发布工程（多平台构建、版本注入、校验和）
