#!/bin/bash

set -e

echo "🔧 Starting loji App in development mode..."

# 检查Wails CLI
if ! command -v wails &> /dev/null; then
    echo "❌ Wails CLI not found!"
    echo "Please install it with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    exit 1
fi

# 安装前端依赖（如果需要）
if [ ! -d "frontend/node_modules" ]; then
    echo "📦 Installing frontend dependencies..."
    cd frontend
    npm install
    cd ..
fi

# 启动开发模式
echo "🚀 Starting Wails dev server..."
wails dev

