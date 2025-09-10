FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /twitter-image-api main.go


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /twitter-image-api .
# COPY .env .

CMD ["./twitter-image-api"]
#EXPOSE 8080
