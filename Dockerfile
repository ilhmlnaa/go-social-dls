FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/


FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN ARCH=$(dpkg --print-architecture) && \
    echo "Building for architecture: $ARCH" && \
    apt-get update && apt-get install -y ca-certificates wget gnupg --no-install-recommends && \
    if [ "$ARCH" = "amd64" ]; then \
        echo "Installing Google Chrome for AMD64..."; \
        wget -q https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && \
        apt-get install -y ./google-chrome-stable_current_amd64.deb --no-install-recommends && \
        rm -f google-chrome-stable_current_amd64.deb && \
        ln -sf /usr/bin/google-chrome /usr/bin/chrome; \
    else \
        echo "Installing Chromium for non-AMD64..."; \
        apt-get install -y chromium chromium-common --no-install-recommends && \
        if [ -x /usr/bin/chromium ]; then ln -sf /usr/bin/chromium /usr/bin/chrome; \
        elif [ -x /usr/bin/chromium-browser ]; then ln -sf /usr/bin/chromium-browser /usr/bin/chrome; \
        else echo "Chromium binary not found"; exit 1; fi; \
    fi && \
    apt-get install -y \
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


ENV CHROME_BIN=/usr/bin/chrome \
    CHROME_PATH=/usr/bin/chrome

WORKDIR /app
COPY --from=builder /app/server .

RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

CMD ["./server"]

