# ---- 构建阶段 ----
FROM golang:1.26-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /html2pdf-chrome ./cmd/html2pdf-chrome

# ---- 运行阶段 ----
FROM debian:bookworm-slim

# 基础工具
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget gnupg ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# 安装 Google Chrome
RUN wget -q -O - https://dl.google.com/linux/linux_signing_key.pub \
      | gpg --dearmor -o /usr/share/keyrings/google-chrome.gpg \
    && echo "deb [arch=amd64 signed-by=/usr/share/keyrings/google-chrome.gpg] \
       http://dl.google.com/linux/chrome/deb/ stable main" \
       > /etc/apt/sources.list.d/google-chrome.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends google-chrome-stable \
    && rm -rf /var/lib/apt/lists/*

# 字体：中日韩 + 数学符号 + 西文 + Emoji
RUN apt-get update && apt-get install -y --no-install-recommends \
    fonts-noto-cjk \
    fonts-noto-cjk-extra \
    fonts-noto-color-emoji \
    fonts-liberation \
    fonts-dejavu-core \
    fonts-stix \
    fonts-lmodern \
    && rm -rf /var/lib/apt/lists/*

# Chrome 运行所需的系统库
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

# 刷新字体缓存
RUN fc-cache -fv

# 创建非 root 用户
RUN useradd -m -s /bin/bash chrome \
    && mkdir -p /app/output \
    && chown -R chrome:chrome /app

COPY --from=builder /html2pdf-chrome /usr/local/bin/html2pdf-chrome

USER chrome
WORKDIR /app

ENTRYPOINT ["html2pdf-chrome"]
