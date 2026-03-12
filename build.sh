#!/bin/bash

set -e

echo "🚀 Building loji App..."

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

# 构建应用
echo "🏗️  Building application..."
if command -v wails &> /dev/null; then
    wails build
    echo "✅ Application built with Wails"
else
    echo "⚠️  Wails CLI not found, building with go build..."
    go build -o loji-app .
    echo "✅ Application built successfully"
fi

echo "🎉 Build complete!"
echo "📍 Binary location: ./build/bin/ or ./loji-app"

