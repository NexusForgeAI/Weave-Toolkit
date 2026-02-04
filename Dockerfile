# 多阶段构建 Dockerfile
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o mcp-server ./cmd/mcp-server

# 运行阶段
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates

# 创建非 root 用户
RUN adduser -D -s /bin/sh appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/mcp-server .

# 复制配置文件
COPY .env ./
COPY tool-config.json ./

# 更改文件所有权
RUN chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8888

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8888/health || exit 1

# 启动命令
CMD ["./mcp-server"]