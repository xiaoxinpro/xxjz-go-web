#!/usr/bin/env bash
set -e
cd "$(dirname "$0")/.."

# Build frontend
echo "Building frontend..."
cd frontend
npm ci 2>/dev/null || npm install
npm run build
cd ..

# Build backend (optional, for local binary)
echo "Building backend..."
cd backend
go build -o ../server ./cmd/server
cd ..

# Docker build
echo "Building Docker image..."
docker build -t xxjz-go-web:latest .

echo "Done. Image: xxjz-go-web:latest"
