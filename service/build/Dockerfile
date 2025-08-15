# 多阶段构建 Dockerfile
ARG BUILD_ENV=production
ARG GO_VERSION=1.24.2

# 构建阶段
FROM golang:${GO_VERSION}-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 设置 Go proxy
ENV GOPROXY=https://goproxy.cn,direct

# 复制 go mod 文件并下载依赖
# 这样做可以利用 Docker 的层缓存机制，只有当 go.mod 或 go.sum 文件发生变化时，才会重新下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 构建应用
ARG BUILD_ENV
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o backend \
    cmd/api/main.go

# 生产环境镜像
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/backend .

EXPOSE 48197

CMD ["./backend apiserver"]
