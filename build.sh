#!/bin/bash
set -e

echo "=== 构建前端 ==="
cd "$(dirname "$0")/frontend"
npm install --silent
npx vite build --outDir=../backend/frontend/dist --emptyOutDir

echo ""
echo "=== 构建后端 ==="
cd ../backend
GOPROXY=https://goproxy.cn,direct go build -o ../cangye.bin .
GOPROXY=https://goproxy.cn,direct go vet ./...

echo ""
echo "=== 构建完成 ==="
cd ..
echo "产物: $(pwd)/cangye.bin"
echo "运行: ./cangye.bin"
