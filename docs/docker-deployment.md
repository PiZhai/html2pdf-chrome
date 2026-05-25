# Docker 部署指南（无桌面 Linux 服务器）

本文档覆盖在无桌面环境的 Linux 服务器上通过 Docker 部署 html2pdf-chrome 的完整流程，包括中文字体、数学公式渲染、系统依赖等。

## Dockerfile

推荐使用项目根目录的 `Dockerfile.cn`（国内镜像加速版）或 `Dockerfile`（国际版）。

以下是 `Dockerfile.cn` 的核心结构说明（完整内容见项目根目录文件）：

- 构建阶段：编译 `html2pdf-chrome`（CLI）和 `html2pdf-server`（HTTP 服务）两个二进制
- 运行阶段：基于 `debian:bookworm-slim`，安装 Chromium + 字体 + 系统库
- 默认以 root 运行，ENTRYPOINT 自带 `-no-sandbox`（避免挂载卷权限问题和 sandbox 报错）
- 默认启动 HTTP 服务模式（`html2pdf-server`）

如果只需要 CLI 模式，可以覆盖 ENTRYPOINT：

```bash
docker run --rm --shm-size=512m \
  --entrypoint html2pdf-chrome \
  -v $(pwd)/output:/app/output \
  html2pdf-chrome \
  -no-sandbox \
  -url https://example.com \
  -out /app/output/example.pdf
```

## 构建镜像

```bash
# 国内环境（使用阿里云镜像 + Chromium）
docker build -f Dockerfile.cn -t html2pdf-chrome .

# 国际环境（使用 Google Chrome）
docker build -t html2pdf-chrome .
```

## 运行

### HTTP 服务模式（推荐，长期运行）

```bash
docker run -d \
  --name html2pdf \
  --restart always \
  --shm-size=512m \
  -p 8080:8080 \
  html2pdf-chrome

# 调用
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "printBackground": true}' \
  -o output.pdf

# 健康检查
curl http://localhost:8080/health
```

### CLI 模式（一次性转换）

```bash
# 转换在线页面
docker run --rm --shm-size=512m \
  --entrypoint html2pdf-chrome \
  -v $(pwd)/output:/app/output \
  html2pdf-chrome \
  -no-sandbox \
  -url https://example.com \
  -out /app/output/example.pdf

# 转换本地 HTML 文件
docker run --rm --shm-size=512m \
  --entrypoint html2pdf-chrome \
  -v $(pwd)/output:/app/output \
  -v $(pwd)/input:/app/input:ro \
  html2pdf-chrome \
  -no-sandbox \
  -html-file /app/input/report.html \
  -out /app/output/report.pdf
```

### 带等待策略

```bash
# 等待网络空闲（适合有异步加载的页面）
docker run --rm --shm-size=512m \
  --entrypoint html2pdf-chrome \
  -v $(pwd)/output:/app/output \
  html2pdf-chrome \
  -no-sandbox \
  -url https://example.com \
  -wait-network-idle \
  -out /app/output/example.pdf
```

HTTP 服务模式下通过 JSON 参数指定：

```bash
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "waitNetworkIdle": true}' \
  -o output.pdf
```

### 关于 --no-sandbox

当前 Dockerfile 默认以 root 运行并在 ENTRYPOINT 中带上 `-no-sandbox`，因此不需要额外处理。
如果你自定义了 Dockerfile 并使用非 root 用户，在某些内核配置下仍可能需要 `-no-sandbox`
（取决于 unprivileged user namespaces 是否可用）。

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

## 作为 HTTP 服务部署（推荐）

`Dockerfile.cn` 默认启动 `html2pdf-server`，无需额外代码。直接运行：

```bash
docker run -d \
  --name html2pdf \
  --restart always \
  --shm-size=512m \
  -p 8080:8080 \
  html2pdf-chrome
```

调整实例池大小：

```bash
docker run -d \
  --name html2pdf \
  --restart always \
  --shm-size=1g \
  -p 8080:8080 \
  html2pdf-chrome \
  -max-instances 8 \
  -min-instances 4
```

服务参数：

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-addr` | `:8080` | 监听地址 |
| `-max-instances` | `4` | 最大 Chrome 实例数 |
| `-min-instances` | `2` | 最小空闲实例数 |
| `-no-sandbox` | ENTRYPOINT 已带 | 禁用 sandbox |

对应的 docker-compose.yml：

```yaml
version: "3.8"
services:
  html2pdf:
    build:
      context: .
      dockerfile: Dockerfile.cn
    ports:
      - "8080:8080"
    shm_size: "512m"
    restart: always
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: "2"
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

### Chrome 启动失败：`No usable sandbox` 或 `Running as root without --no-sandbox`

当前 Dockerfile 的 ENTRYPOINT 已默认带 `-no-sandbox`。如果你自定义了启动命令，确保加上 `-no-sandbox` 参数。

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
