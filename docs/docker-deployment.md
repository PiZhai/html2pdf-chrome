# Docker 部署指南（无桌面 Linux 服务器）

本文档覆盖在无桌面环境的 Linux 服务器上通过 Docker 部署 html2pdf-chrome 的完整流程，包括中文字体、数学公式渲染、系统依赖等。

## Dockerfile

```dockerfile
# ---- 构建阶段 ----
FROM golang:1.26-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /html2pdf-chrome ./cmd/html2pdf-chrome

# ---- 运行阶段 ----
FROM debian:bookworm-slim

# 1. 基础工具
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget gnupg ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# 2. 安装 Google Chrome
RUN wget -q -O - https://dl.google.com/linux/linux_signing_key.pub \
      | gpg --dearmor -o /usr/share/keyrings/google-chrome.gpg \
    && echo "deb [arch=amd64 signed-by=/usr/share/keyrings/google-chrome.gpg] \
       http://dl.google.com/linux/chrome/deb/ stable main" \
       > /etc/apt/sources.list.d/google-chrome.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends google-chrome-stable \
    && rm -rf /var/lib/apt/lists/*

# 3. 字体：中日韩 + 数学符号 + 西文
RUN apt-get update && apt-get install -y --no-install-recommends \
    fonts-noto-cjk \
    fonts-noto-cjk-extra \
    fonts-noto-color-emoji \
    fonts-liberation \
    fonts-dejavu-core \
    fonts-stix \
    fonts-lmodern \
    && rm -rf /var/lib/apt/lists/*

# 4. Chrome 运行所需的系统库（headless 模式仍需要这些）
RUN apt-get update && apt-get install -y --no-install-recommends \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libcups2 \
    libdbus-1-3 \
    libdrm2 \
    libgbm1 \
    libgtk-3-0 \
    libnspr4 \
    libnss3 \
    libx11-xcb1 \
    libxcomposite1 \
    libxdamage1 \
    libxfixes3 \
    libxrandr2 \
    libxshmfence1 \
    xdg-utils \
    && rm -rf /var/lib/apt/lists/*

# 5. 刷新字体缓存
RUN fc-cache -fv

# 6. 创建非 root 用户（避免使用 --no-sandbox）
RUN useradd -m -s /bin/bash chrome \
    && mkdir -p /app/output \
    && chown -R chrome:chrome /app

COPY --from=builder /html2pdf-chrome /usr/local/bin/html2pdf-chrome

USER chrome
WORKDIR /app

ENTRYPOINT ["html2pdf-chrome"]
```

## 构建镜像

```bash
docker build -t html2pdf-chrome:latest .
```

## 运行

### 基本用法

```bash
# 转换在线页面
docker run --rm -v $(pwd)/output:/app/output html2pdf-chrome:latest \
  -url https://example.com \
  -out /app/output/example.pdf

# 转换本地 HTML 文件
docker run --rm \
  -v $(pwd)/output:/app/output \
  -v $(pwd)/input:/app/input:ro \
  html2pdf-chrome:latest \
  -html-file /app/input/report.html \
  -out /app/output/report.pdf
```

### 带等待策略

```bash
# 等待网络空闲（适合有异步加载的页面）
docker run --rm -v $(pwd)/output:/app/output html2pdf-chrome:latest \
  -url https://example.com \
  -wait-network-idle \
  -out /app/output/example.pdf

# 等待自定义条件（适合 SPA 或动态渲染页面）
docker run --rm -v $(pwd)/output:/app/output html2pdf-chrome:latest \
  -url https://example.com \
  -wait-expression "window.__RENDER_DONE === true" \
  -out /app/output/example.pdf
```

### 以 root 运行（不推荐，但某些环境需要）

如果你的环境必须以 root 运行容器，需要加 `--no-sandbox`：

```bash
docker run --rm -u root -v $(pwd)/output:/app/output html2pdf-chrome:latest \
  -url https://example.com \
  -no-sandbox \
  -out /app/output/example.pdf
```

## 字体说明

Dockerfile 中安装的字体包覆盖以下场景：

| 字体包 | 覆盖范围 |
|--------|----------|
| `fonts-noto-cjk` | 中文、日文、韩文基础字符 |
| `fonts-noto-cjk-extra` | CJK 扩展区字符（生僻字） |
| `fonts-noto-color-emoji` | Emoji 表情 |
| `fonts-liberation` | Times New Roman / Arial / Courier 等西文等宽替代 |
| `fonts-dejavu-core` | DejaVu 系列，覆盖拉丁、希腊、西里尔字母 |
| `fonts-stix` | STIX 数学字体，覆盖数学符号和公式 |
| `fonts-lmodern` | Latin Modern，LaTeX 默认字体，数学排版常用 |

### 数学公式渲染

如果你的 HTML 使用 MathJax 或 KaTeX 渲染数学公式：

- **MathJax**：自带 Web 字体，不依赖系统字体。确保页面加载完成即可（用 `-wait-network-idle` 或 `-wait-expression "MathJax.isReady"`）。
- **KaTeX**：同样自带字体。等待条件示例：`-wait-expression "document.querySelector('.katex') !== null"`。
- **原生 MathML**：依赖系统数学字体，`fonts-stix` 已覆盖。

### 添加自定义字体

如果需要使用特定商业字体（如思源宋体、苹方等），将字体文件挂载进容器：

```bash
docker run --rm \
  -v $(pwd)/fonts:/usr/share/fonts/custom:ro \
  -v $(pwd)/output:/app/output \
  html2pdf-chrome:latest \
  -url https://example.com \
  -out /app/output/example.pdf
```

或在 Dockerfile 中直接 COPY：

```dockerfile
COPY ./fonts/ /usr/share/fonts/custom/
RUN fc-cache -fv
```

## 作为 HTTP 服务部署

实际生产中通常不会每次 `docker run` 一个容器，而是在容器内运行一个 HTTP 服务，调用 Go 库的池化模式。

示例服务代码：

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/PiZhai/html2pdf-chrome/pkg/html2pdf"
)

func main() {
    converter, err := html2pdf.NewConverter(html2pdf.ConverterConfig{
        MaxInstances: 4,
        MinInstances: 2,
        NoSandbox:    os.Getenv("NO_SANDBOX") == "true",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer converter.Close()

    http.HandleFunc("/convert", func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            URL string `json:"url"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        tmpFile := filepath.Join(os.TempDir(), "pdf-"+time.Now().Format("20060102150405")+".pdf")
        defer os.Remove(tmpFile)

        err := converter.Convert(html2pdf.Request{
            URL:        req.URL,
            OutputPath: tmpFile,
            Options: html2pdf.Options{
                Timeout:         30 * time.Second,
                WaitNetworkIdle: true,
                PrintBackground: true,
                NoSandbox:       os.Getenv("NO_SANDBOX") == "true",
            },
        })
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/pdf")
        http.ServeFile(w, r, tmpFile)
    })

    log.Println("listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

对应的 docker-compose.yml：

```yaml
version: "3.8"
services:
  html2pdf:
    build: .
    ports:
      - "8080:8080"
    environment:
      - NO_SANDBOX=false
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: "2"
    # 如果必须以 root 运行，取消下面的注释并设置 NO_SANDBOX=true
    # user: root
    # environment:
    #   - NO_SANDBOX=true
```

## 资源限制建议

| 配置项 | 建议值 | 说明 |
|--------|--------|------|
| 内存 | 每个 Chrome 实例 ~300-500MB | 4 实例建议至少 2GB |
| CPU | 每个实例 0.5 核 | 4 实例建议 2 核 |
| `/dev/shm` | 至少 256MB | Chrome 使用共享内存，Docker 默认 64MB 不够 |

如果遇到 Chrome 崩溃或页面渲染异常，先检查 `/dev/shm` 大小：

```bash
# 方法 1：增大 /dev/shm
docker run --shm-size=512m ...

# 方法 2：使用 /tmp 替代（Chrome 启动参数已包含 --disable-dev-shm-usage 时）
# 当前代码未加此参数，建议用方法 1
```

## 常见问题

### Chrome 启动失败：`Running as root without --no-sandbox is not supported`

以非 root 用户运行容器（Dockerfile 中已配置 `USER chrome`），或者加 `-no-sandbox` 参数。

### PDF 中文显示为方块

字体未安装或字体缓存未刷新。确认 `fonts-noto-cjk` 已安装且执行了 `fc-cache -fv`。

### 数学公式显示不完整

- MathJax/KaTeX：加 `-wait-network-idle` 确保字体文件加载完成。
- 原生 MathML：确认 `fonts-stix` 已安装。

### 页面内容不完整（异步加载）

使用 `-wait-network-idle` 或 `-wait-expression` 等待页面渲染完成。

### 容器内存不足导致 Chrome 被 OOM Kill

增加容器内存限制，或减少 `MaxInstances`。

### ARM64 服务器（如 AWS Graviton）

Google Chrome 官方不提供 ARM64 版本。使用 Chromium 替代：

```dockerfile
RUN apt-get update && apt-get install -y --no-install-recommends chromium
ENV CHROME_PATH=/usr/bin/chromium
```
