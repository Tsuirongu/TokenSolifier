#!/bin/bash

set -e

echo "🚀 Building loji App for Windows..."

# 检查是否安装了mingw-w64
if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    echo "❌ mingw-w64 not found. Please install it first:"
    echo "   macOS: brew install mingw-w64"
    echo "   Ubuntu/Debian: sudo apt-get install gcc-mingw-w64"
    echo "   CentOS/RHEL: sudo yum install mingw64-gcc"
    exit 1
fi

# 安装前端依赖
echo "📦 Installing frontend dependencies..."
cd frontend
npm install
echo "✅ Frontend dependencies installed"

# 构建前端
echo "🏗️  Building frontend..."
npm run build
echo "✅ Frontend built successfully"

cd ..

# 下载Go依赖
echo "📦 Downloading Go dependencies..."
go mod download
echo "✅ Go dependencies downloaded"

# 检查Wails CLI是否安装
if ! command -v wails &> /dev/null; then
    echo "❌ Wails CLI not found. Installing..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
fi

# 构建Windows版本
echo "🏗️  Building Windows application..."
wails build -platform windows/amd64 -clean
echo "✅ Windows application built successfully"

echo "🎉 Windows build complete!"
echo "📍 Windows executable location: ./build/bin/loji-app.exe"
echo ""
echo "📋 To run on Windows:"
echo "   1. Copy loji-app.exe to Windows machine"
echo "   2. Double-click to run (requires WebView2 runtime)"
echo "   3. Or run from command line: loji-app.exe"
