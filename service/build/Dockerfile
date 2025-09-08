# 多阶段构建 Dockerfile
ARG BUILD_ENV=production
ARG GO_VERSION=1.24.6

# 构建阶段
FROM registry.sciol.ac.cn/library/golang:${GO_VERSION}-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

ENV GOPROXY=https://goproxy.cn,direct

# 复制 go mod 文件
COPY go.mod go.sum ./

# 复制源码
COPY . .

# 下载依赖
RUN go mod download

# 构建应用
ARG BUILD_ENV
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o backend \
    main.go

# 生产环境镜像
FROM registry.sciol.ac.cn/library/golang:${GO_VERSION}-alpine AS production

RUN apk --no-cache add ca-certificates tzdata make
WORKDIR /app

COPY --from=builder /app/backend .

EXPOSE 48197

CMD ["./backend", "apiserver"]
