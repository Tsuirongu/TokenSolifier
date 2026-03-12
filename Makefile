.PHONY: all dev build clean install help

# 默认目标
all: build

# 构建Windows版本
build-windows:
	@echo "Building Windows application..."
	@./build-windows.sh

# 构建多平台版本 (macOS + Windows)
build-multiplatform:
	@echo "Building applications for multiple platforms..."
	@./build-multiplatform.sh

# 开发模式
dev:
	@echo "Starting development mode..."
	@./dev.sh

# 构建应用
build:
	@echo "Building application..."
	@mkdir -p ./build/
	@cp ./appicon.png ./build/appicon.png
	@./build.sh

# 清理构建产物
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf build/
	@rm -rf frontend/dist/
	@rm -rf frontend/node_modules/
	@rm -f loji-app loji-app.exe
	@echo "Clean complete!"

# 安装依赖
install:
	@echo "Installing dependencies..."
	@cd frontend && npm install
	@go mod download
	@echo "Dependencies installed!"

# 初始化开发环境
init:
	@echo "Initializing development environment..."
	@if ! command -v wails &> /dev/null; then \
		echo "Installing Wails CLI..."; \
		go install github.com/wailsapp/wails/v2/cmd/wails@latest; \
	fi
	@chmod +x build.sh dev.sh
	@make install
	@echo "Development environment ready!"

# 运行应用
run: build
	@echo "Running application..."
	@if [ -f "./build/bin/loji-app" ]; then \
		./build/bin/loji-app; \
	elif [ -f "./loji-app" ]; then \
		./loji-app; \
	else \
		echo "Application not found. Please run 'make build' first."; \
	fi

# 帮助信息
help:
	@echo "loji App - Makefile Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  dev                - Start development mode with hot reload"
	@echo "  build              - Build the application for production (current platform)"
	@echo "  build-windows      - Build the application for Windows (cross-platform)"
	@echo "  build-multiplatform- Build applications for macOS + Windows simultaneously"
	@echo "  run                - Build and run the application"
	@echo "  clean              - Remove build artifacts"
	@echo "  install            - Install dependencies"
	@echo "  init               - Initialize development environment"
	@echo "  help               - Show this help message"
