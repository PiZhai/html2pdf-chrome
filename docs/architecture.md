# Architecture

## 技术定位

- 渲染引擎：`Chrome/Chromium Headless`
- 驱动方式：`Chrome DevTools Protocol (WebSocket)`
- 实现语言：`Go`

## 主流程

```text
CLI flags
  -> config.Config
  -> Validate / InputTarget / PrepareOutputPath
  -> browser.FindChrome
  -> browser.Launch
  -> cdp.Connect
  -> cdp.OpenPage
  -> render.PrintToFile
  -> output.pdf
```

## 模块职责

### `cmd/html2pdf-chrome`

CLI 入口。

负责：

- 解析命令行参数
- 组装 `config.Config`
- 调用 `app.Run`

不负责：

- 浏览器启动
- CDP 控制
- PDF 导出

### `internal/config`

配置层。

负责：

- 参数结构定义
- 基础校验
- 输入目标规范化
- 输出路径规范化

### `internal/app`

应用编排层。

负责决定调用顺序：

- 校验配置
- 解析输入/输出
- 查找并启动浏览器
- 建立 CDP 连接
- 打开页面
- 导出 PDF

### `internal/browser`

浏览器进程层。

负责：

- 查找 Chrome / Chromium
- 启动 Headless Chrome
- 获取 WebSocket 调试地址
- 关闭浏览器并清理临时目录

### `internal/cdp`

浏览器协议层。

负责：

- 连接已有浏览器实例
- 导航到目标页面
- 等待页面 ready
- 等待字体加载完成

### `internal/render`

渲染层。

负责：

- 调用 `Page.printToPDF`
- 按选项导出 PDF 文件

## 当前设计特点

- 每次任务使用独立 `user-data-dir`
- 默认静默运行，必要时可开启 `-chrome-debug-log`
- `-paper` 与 `-prefer-css-page-size` 已拆分，避免 CLI 纸张设置被静默覆盖
- `Page.printToPDF` 常用参数已从 CLI 直连到底层渲染调用
- 页面等待策略目前包含：
  - `body` ready
  - `document.readyState === "complete"`
  - `document.fonts.status === "loaded"`

## 下一阶段建议

1. 做 Linux / Windows / macOS 三端实机验证
2. 增加更稳定的等待策略，例如网络空闲和自定义 JS 条件
3. 补充错误分类与可观测性，方便定位 Chrome 启动和页面渲染失败原因
4. 补发布工程，例如多平台构建脚本、校验和与版本信息注入
