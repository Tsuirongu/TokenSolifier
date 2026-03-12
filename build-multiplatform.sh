#!/bin/bash

set -e

echo "🚀 Building loji App for multiple platforms (macOS + Windows)..."

# 检查是否安装了mingw-w64 (用于Windows交叉编译)
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

# 获取当前系统的架构
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    if [[ $(uname -m) == 'arm64' ]]; then
        CURRENT_PLATFORM="darwin/arm64"
    else
        CURRENT_PLATFORM="darwin/amd64"
    fi
else
    echo "❌ This script is designed for macOS. For other platforms, please use platform-specific builds."
    exit 1
fi

# 构建多平台版本
echo "🏗️  Building applications for multiple platforms..."
echo "   Platforms: $CURRENT_PLATFORM, windows/amd64"

wails build -platform "$CURRENT_PLATFORM,windows/amd64" -clean

# 重命名Windows exe文件为更简洁的名字
if [ -f "./build/bin/loji-app-amd64.exe" ]; then
    mv "./build/bin/loji-app-amd64.exe" "./build/bin/loji-app.exe"
    echo "✅ Renamed Windows executable to loji-app.exe"
fi

echo "✅ Multi-platform applications built successfully"

echo ""
echo "🎉 Build complete!"
echo "📍 Output locations:"
echo "   macOS app: ./build/bin/loji-app.app"
echo "   Windows exe: ./build/bin/loji-app.exe"
echo ""
echo "📋 Distribution instructions:"
echo "   macOS: Copy loji-app.app to target Mac, or create .dmg installer"
echo "   Windows: Copy loji-app.exe to Windows machine (requires WebView2 runtime)"
echo ""
echo "💡 WebView2 for Windows: https://developer.microsoft.com/microsoft-edge/webview2/"
