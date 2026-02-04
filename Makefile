# MCP Server Makefile

BINARY_NAME=mcp-server
VERSION=1.0.0
BUILD_DIR=bin

# 默认目标
.PHONY: all
all: build

# 构建应用
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mcp-server

# 清理构建文件
.PHONY: clean
clean:
	@echo "Cleaning build files..."
	@rm -rf $(BUILD_DIR)

# 运行应用
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# 测试
.PHONY: test
test:
	@echo "Running tests..."
	@go test ./... -v

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# 依赖检查
.PHONY: deps
deps:
	@echo "Checking dependencies..."
	@go mod tidy
	@go mod verify

# 构建 Docker 镜像
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) .

# 运行 Docker 容器
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8888:8888 -v $(PWD)/.env:/app/.env -v $(PWD)/tool-config.json:/app/tool-config.json $(BINARY_NAME):$(VERSION)

# 安装依赖
.PHONY: install
install:
	@echo "Installing dependencies..."
	@go mod download

# 显示帮助
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build       - Build the application"
	@echo "  run         - Build and run the application"
	@echo "  test        - Run tests"
	@echo "  fmt         - Format code"
	@echo "  deps        - Check and tidy dependencies"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  clean        - Clean build files"
	@echo "  help         - Show this help message"