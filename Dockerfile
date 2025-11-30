FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/


# ============= RUNTIME (MULTI-ARCH) =============
FROM debian:bookworm-slim

# Install common deps
RUN apt-get update && apt-get install -y \
    ca-certificates \
    wget \
    gnupg \
    fonts-liberation \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libatspi2.0-0 \
    libcups2 \
    libdbus-1-3 \
    libdrm2 \
    libgbm1 \
    libgtk-3-0 \
    libnspr4 \
    libnss3 \
    libxcomposite1 \
    libxdamage1 \
    libxfixes3 \
    libxkbcommon0 \
    libxrandr2 \
    xdg-utils \
    libu2f-udev \
    libvulkan1 \
    libxshmfence1 \
    && rm -rf /var/lib/apt/lists/*

RUN ARCH=$(dpkg --print-architecture) && \
    if [ "$ARCH" = "amd64" ]; then \
        echo ">>> Installing Google Chrome for AMD64"; \
        wget -q https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && \
        apt-get update && apt-get install -y ./google-chrome-stable_current_amd64.deb && \
        rm google-chrome-stable_current_amd64.deb; \
    else \
        echo ">>> Installing Chromium for ARM64"; \
        apt-get update && apt-get install -y chromium chromium-common chromium-driver; \
    fi && \
    rm -rf /var/lib/apt/lists/*

RUN echo "ARCH: $(dpkg --print-architecture)" && \
    (which google-chrome || which chromium || true)

WORKDIR /app
COPY --from=builder /app/server .

RUN mkdir -p /app/cookies && chmod -R 755 /app

ENV CHROME_BIN="/usr/bin/google-chrome" \
    CHROME_PATH="/usr/bin/google-chrome" \
    ENV=production

RUN if [ -f "/usr/bin/chromium" ]; then \
      sed -i 's/google-chrome/chromium/g' /etc/environment || true; \
      echo 'CHROME_BIN=/usr/bin/chromium' >> /etc/environment; \
      echo 'CHROME_PATH=/usr/bin/chromium' >> /etc/environment; \
    fi

RUN useradd -m -u 1000 appuser && \
    chown -R appuser:appuser /app

USER appuser

CMD ["./server"]
