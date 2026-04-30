# FayHub 后端Docker镜像构建文件
# 基于Alpine Linux的轻量级Go镜像

# 构建阶段
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fayhub ./cmd

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -S fayhub && adduser -S fayhub -G fayhub

WORKDIR /app

COPY --from=builder /app/fayhub .

COPY config_prod.yaml ./config.yaml

RUN mkdir -p logs uploads

RUN chown -R fayhub:fayhub /app

USER fayhub

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

CMD ["./fayhub"]
