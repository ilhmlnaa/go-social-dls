FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates wget
COPY --from=builder /app/server .

RUN adduser -D appuser && chown -R appuser /app
USER appuser

# Port default aplikasi (config.Load() -> PORT, default 3005)
ENV PORT=3005
EXPOSE 3005

# Healthcheck: hit endpoint /health via port dari env PORT.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider "http://127.0.0.1:${PORT}/health" || exit 1

CMD ["./server"]
