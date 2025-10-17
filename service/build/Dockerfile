ARG BUILD_ENV=production
ARG GO_VERSION=1.24.6

FROM registry.sciol.ac.cn/library/golang:${GO_VERSION}-alpine AS base

FROM base AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

# 使用国内 GOPROXY，加速依赖下载
ENV GOPROXY=https://goproxy.cn,direct

# Install swag
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
COPY . .

RUN go mod download

# Generate swagger documentation before building
# RUN /go/bin/swag init -g main.go

ARG BUILD_ENV
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o backend \
    main.go

FROM base AS production

RUN apk --no-cache add ca-certificates tzdata make
WORKDIR /app

# Copy only the final compiled binary. Swag is no longer needed.
COPY --from=builder /app/backend .

EXPOSE 48197

CMD ["./backend", "apiserver"]
