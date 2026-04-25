# FayHub 后端Docker镜像构建文件
# 基于Alpine Linux的轻量级Go镜像

# 构建阶段
FROM golang:1.22-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fayhub ./cmd

# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 创建应用用户
RUN addgroup -S fayhub && adduser -S fayhub -G fayhub

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/fayhub .

# 复制配置文件模板
COPY config.yaml.production ./config.yaml
COPY .env.production ./.env

# 创建必要的目录
RUN mkdir -p logs uploads

# 设置文件权限
RUN chown -R fayhub:fayhub /app

# 切换到非root用户
USER fayhub

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# 启动命令
CMD ["./fayhub"]