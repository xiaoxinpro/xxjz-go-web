#!/usr/bin/env bash
# 构建脚本 (Linux/macOS)：输出到项目根目录 build/，可在该目录直接运行
set -e
cd "$(dirname "$0")/.."
BUILD_DIR="build"

# 若存在则删除 build，再创建
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

echo "Building frontend..."
cd frontend
npm ci 2>/dev/null || npm install
npm run build
cd ..

echo "Building backend..."
cd backend
go build -o "../$BUILD_DIR/server" ./cmd/server
cd ..

# 复制运行所需文件到 build
cp config.yaml "$BUILD_DIR/"
mkdir -p "$BUILD_DIR/static"
cp -r frontend/dist/. "$BUILD_DIR/static/"
cp -r backend/migrations "$BUILD_DIR/"

echo "Building Docker image..."
docker build -t xxjz-go-web:latest .

echo "Done. Run: cd $BUILD_DIR && ./server"
