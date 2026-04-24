# ============================================================
# UC-Server Makefile
# 项目构建、开发、测试、部署的统一入口
# ============================================================

# ----------------------------------------------------------
# 应用信息（构建时注入到二进制中）
# ----------------------------------------------------------

# 应用名称
APP_NAME := uc-service

# 版本号，支持通过环境变量覆盖：VERSION=1.1.0 make build
VERSION ?= 1.0.0

# 构建时间（UTC，24小时制）
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Git 短提交哈希，用于追踪构建来源
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 链接标志：strip 符号表和调试信息，注入版本元数据到 main 包
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"


# ----------------------------------------------------------
# 目录与文件路径
# ----------------------------------------------------------

# 编译输出目录
BIN_DIR := bin

# Docker 部署文件目录
DEPLOY_DIR := deploy

# 主程序入口目录
CMD_DIR := cmd/server

# 测试覆盖率输出文件
COVERAGE_FILE := coverage.out


# ----------------------------------------------------------
# Go 命令别名
# ----------------------------------------------------------

GOCMD  := go
GOBUILD := $(GOCMD) build
GORUN  := $(GOCMD) run
GOTEST := $(GOCMD) test
GOMOD  := $(GOCMD) mod
GOFMT  := $(GOCMD) fmt
GOVET  := $(GOCMD) vet
RUN_ARGS ?=


# ----------------------------------------------------------
# 外部工具
# ----------------------------------------------------------

GOLANGCI_LINT := golangci-lint
AIR := air
SWAG := swag


# ----------------------------------------------------------
# 默认目标：显示帮助信息
# ----------------------------------------------------------
.DEFAULT_GOAL := help


# ----------------------------------------------------------
# 声明伪目标（不对应实际文件，避免与同名文件冲突）
# ----------------------------------------------------------
.PHONY: all build run dev test test-coverage test-coverage-html clean lint fmt vet tidy deps docker docker-compose-up docker-compose-down swag install check help


# ============================================================
# 核心构建目标
# ============================================================

## all: 完整 CI 流程（清理 → 代码检查 → 测试 → 构建）
all: clean lint test build


## build: 编译应用二进制文件
build:
	@echo "==> Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) ./$(CMD_DIR)
	@echo "==> Build complete: $(BIN_DIR)/$(APP_NAME)"


## run: 直接运行应用（不生成二进制，适合快速验证）
run:
	@echo "==> Running $(APP_NAME)..."
	$(GORUN) $(LDFLAGS) ./$(CMD_DIR) $(RUN_ARGS)


## dev: 启动热重载开发服务器（需提前安装 air）
dev:
	@echo "==> Starting development server with hot reload..."
	@which $(AIR) > /dev/null || (echo "Error: air not found. Install with: go install github.com/air-verse/air@latest" && exit 1)
	$(AIR)


# ============================================================
# 测试相关目标
# ============================================================

## test: 运行全部单元测试（开启 race 检测，超时 60 秒）
test:
	@echo "==> Running tests..."
	$(GOTEST) -v -race -timeout 60s ./...


## test-coverage: 运行测试并生成覆盖率报告
test-coverage:
	@echo "==> Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "==> Coverage report generated: $(COVERAGE_FILE)"
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print "==> Total coverage: " $$3}'


## test-coverage-html: 生成可视化的 HTML 覆盖率报告
test-coverage-html: test-coverage
	@echo "==> Generating HTML coverage report..."
	@$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "==> HTML coverage report: coverage.html"


# ============================================================
# 代码质量与依赖管理
# ============================================================

## fmt: 格式化所有 Go 源代码（调用 go fmt）
fmt:
	@echo "==> Formatting code..."
	@$(GOFMT) ./...
	@echo "==> Format complete"


## vet: 运行 go vet 静态分析，检测常见代码问题
vet:
	@echo "==> Running go vet..."
	@$(GOVET) ./...
	@echo "==> Vet complete"


## lint: 运行 golangci-lint 综合代码检查（需提前安装）
lint:
	@echo "==> Running linter..."
	@which $(GOLANGCI_LINT) > /dev/null || (echo "Error: golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	@$(GOLANGCI_LINT) run ./...
	@echo "==> Lint complete"


## check: 一键运行全部代码质量检查（fmt + vet + lint + test）
check: fmt vet lint test
	@echo "==> All checks passed!"


## tidy: 整理并校验 Go module 依赖
tidy:
	@echo "==> Tidying Go modules..."
	@$(GOMOD) tidy
	@$(GOMOD) verify
	@echo "==> Tidy complete"


## deps: 下载项目依赖（首次克隆或 go.mod 变更后执行）
deps:
	@echo "==> Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) verify
	@echo "==> Dependencies downloaded"


# ============================================================
# 清理与安装
# ============================================================

## clean: 清理构建产物和缓存（保留 module 缓存以加速后续构建）
clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -f $(COVERAGE_FILE) coverage.html
	@$(GOCMD) clean -cache -testcache
	@echo "==> Clean complete"


## install: 将编译好的二进制安装到 GOPATH/bin（需先执行 build）
install: build
	@echo "==> Installing $(APP_NAME) to $(shell go env GOPATH)/bin..."
	@cp $(BIN_DIR)/$(APP_NAME) $(shell go env GOPATH)/bin/
	@echo "==> Install complete"


# ============================================================
# Docker 相关
# ============================================================

## docker: 构建 Docker 镜像（打版本标签和 latest 标签）
docker:
	@echo "==> Building Docker image..."
	@test -f $(DEPLOY_DIR)/Dockerfile || (echo "Error: Dockerfile not found at $(DEPLOY_DIR)/Dockerfile" && exit 1)
	docker build -f $(DEPLOY_DIR)/Dockerfile -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .
	@echo "==> Docker image built: $(APP_NAME):$(VERSION)"


## docker-compose-up: 使用 Docker Compose 启动依赖服务（如 MySQL、Redis）
docker-compose-up:
	@echo "==> Starting services with docker compose..."
	@test -f docker-compose.yml || (echo "Warning: docker-compose.yml not found in root directory" && exit 1)
	docker compose up -d


## docker-compose-down: 停止 Docker Compose 运行的服务
docker-compose-down:
	@echo "==> Stopping services with docker compose..."
	@test -f docker-compose.yml || (echo "Warning: docker-compose.yml not found in root directory" && exit 1)
	docker compose down


# ============================================================
# 文档生成
# ============================================================

## swag: 根据代码注释生成 Swagger API 文档（需提前安装 swag）
swag:
	@echo "==> Generating Swagger docs..."
	@which $(SWAG) > /dev/null || (echo "Error: swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	$(SWAG) init -g $(CMD_DIR)/main.go -o api/docs
	@echo "==> Swagger docs generated"


# ============================================================
# 帮助信息
# ============================================================

## help: 显示所有可用的 Make 目标及其说明
help:
	@echo "$(APP_NAME) Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/  /'
