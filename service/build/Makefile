# ========================
# Studio Service Makefile
# ========================

# 项目配置
PROJECT_NAME := studio-service
MODULE_NAME := studio/service
BINARY_NAME := api
BINARY_DIR := bin
BUILD_DIR := build
CMD_DIR := ./

# Go 配置
GO := go
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
CGO_ENABLED := 0

# 版本信息
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS := -ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 默认目标
.DEFAULT_GOAL := help

# ===== 帮助信息 =====
.PHONY: help
help: ## 显示帮助信息
	@echo "Studio Service Makefile"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ===== 开发相关 =====
.PHONY: init
init: ## 初始化项目依赖
	@echo "🔧 初始化项目..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "✅ 项目初始化完成"

.PHONY: dev
dev: ## 开发模式启动 (热重载)
	@echo "🚀 启动开发服务器..."
	@if ! command -v air > /dev/null 2>&1; then \
		echo "📦 安装 air (热重载工具)..."; \
		$(GO) install github.com/air-verse/air@v1.62.0; \
		echo "✅ air 安装完成"; \
	else \
		echo "✅ air 已存在，跳过安装"; \
	fi
	air -c .air.web.toml

.PHONY: dev-schedule
dev-schedule: ## 开发模式启动 schedule (热重载)
	@echo "🚀 启动 schedule 开发服务器..."
	@if ! command -v air > /dev/null 2>&1; then \
		echo "📦 安装 air (热重载工具)..."; \
		$(GO) install github.com/air-verse/air@v1.62.0; \
		echo "✅ air 安装完成"; \
	else \
		echo "✅ air 已存在，跳过安装"; \
	fi
	air -c .air.schedule.toml

.PHONY: run
run: build ## 构建并运行应用
	@echo "🚀 启动应用..."
	./$(BINARY_DIR)/$(BINARY_NAME)

.PHONY: start
start: ## 快速启动 (开发模式)
	@echo "⚡ 快速启动..."
	$(GO) run ./$(CMD_DIR)

# ===== 构建相关 =====
.PHONY: build
build: clean ## 构建应用
	@echo "🔨 构建应用..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
	$(GO) build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "✅ 构建完成: $(BINARY_DIR)/$(BINARY_NAME)"

.PHONY: build-linux
build-linux: clean ## 构建 Linux 版本
	@echo "🔨 构建 Linux 版本..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 \
	$(GO) build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux ./$(CMD_DIR)
	@echo "✅ Linux 构建完成: $(BINARY_DIR)/$(BINARY_NAME)-linux"

.PHONY: build-all
build-all: clean ## 构建所有平台版本
	@echo "🔨 构建所有平台版本..."
	@mkdir -p $(BINARY_DIR)
	# Linux
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 \
	$(GO) build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	# macOS
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=amd64 \
	$(GO) build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=arm64 \
	$(GO) build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	# Windows
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 \
	$(GO) build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	@echo "✅ 所有平台构建完成"

# ===== 测试相关 =====
.PHONY: test
test: ## 运行测试
	@echo "🧪 运行测试..."
	$(GO) test -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "🧪 运行测试覆盖率..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "📊 覆盖率报告: coverage.html"

.PHONY: bench
bench: ## 运行基准测试
	@echo "⚡ 运行基准测试..."
	$(GO) test -bench=. -benchmem ./...

# ===== 代码质量 =====
.PHONY: fmt
fmt: ## 格式化代码
	@echo "🎨 格式化代码..."
	$(GO) fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	fi

.PHONY: lint
lint: ## 代码检查
	@echo "🔍 代码检查..."
	@if ! command -v golangci-lint > /dev/null; then \
		echo "📦 安装 golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
	fi
	golangci-lint run

.PHONY: vet
vet: ## Go vet 检查
	@echo "🔍 Go vet 检查..."
	$(GO) vet ./...

.PHONY: mod
mod: ## 整理依赖
	@echo "📦 整理依赖..."
	$(GO) mod tidy
	$(GO) mod verify

# ===== Docker 相关 =====
.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	@echo "🐳 构建 Docker 镜像..."
	docker build -f $(BUILD_DIR)/Dockerfile -t $(PROJECT_NAME):$(VERSION) .
	docker tag $(PROJECT_NAME):$(VERSION) $(PROJECT_NAME):latest

.PHONY: docker-run
docker-run: ## 运行 Docker 容器
	@echo "🐳 运行 Docker 容器..."
	docker run -p 8080:8080 --name $(PROJECT_NAME) $(PROJECT_NAME):latest

.PHONY: docker-stop
docker-stop: ## 停止 Docker 容器
	@echo "🛑 停止 Docker 容器..."
	docker stop $(PROJECT_NAME) || true
	docker rm $(PROJECT_NAME) || true

# ===== 清理相关 =====
.PHONY: clean
clean: ## 清理构建文件
	@echo "🧹 清理构建文件..."
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ 清理完成"

.PHONY: clean-all
clean-all: clean ## 深度清理 (包括 Docker)
	@echo "🧹 深度清理..."
	@docker rmi $(PROJECT_NAME):$(VERSION) $(PROJECT_NAME):latest 2>/dev/null || true
	$(GO) clean -cache -modcache -testcache
	@echo "✅ 深度清理完成"

# ===== 部署相关 =====
.PHONY: install
install: build ## 安装到系统
	@echo "📦 安装应用..."
	@install -d $(DESTDIR)/usr/local/bin
	@install -m 755 $(BINARY_DIR)/$(BINARY_NAME) $(DESTDIR)/usr/local/bin/
	@echo "✅ 安装完成"

# ===== 信息查看 =====
.PHONY: info
info: ## 显示项目信息
	@echo "项目信息:"
	@echo "  项目名称: $(PROJECT_NAME)"
	@echo "  模块名称: $(MODULE_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Git 提交: $(GIT_COMMIT)"
	@echo "  Go 版本: $$($(GO) version)"
	@echo "  系统: $(GOOS)/$(GOARCH)"

# 确保这些目标不被误认为是文件
.PHONY: all build clean test generate apiserver migrate lint fmt-install fmt build-web stringer-install


stringer-install:
	@if ! command -v stringer > /dev/null; then \
		echo "📦 安装 stringer..."; \
		go install golang.org/x/tools/cmd/stringer@latest; \
	fi
# 生成错误码
generate: stringer-install
	go generate ./pkg/...

# 运行 migrate
migrate:
	go run main.go migrate

# 运行服务
apiserver: generate
	go run main.go apiserver

# 运行调度
schedule: generate
	go run main.go schedule


fmt-install:
	go install mvdan.cc/gofumpt@v0.8.0

fmt-go: fmt-install
	gofumpt -l -w ./pkg ./cmd ./internal
	go fmt ./pkg/... ./cmd/... ./internal/...
	goimports -w ./pkg ./cmd ./internal


lint-install-go:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.3.0

lint-go: fmt-go lint-install-go
	golangci-lint run -v --max-same-issues=0 --timeout=5m --tests=false -c .golangci.yaml  ./...


build-web: fmt
	go build -o ./bin/web main.go
