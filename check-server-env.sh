#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 檢查前端環境變數檔案 ==="
cd ~/Link/frontend
ls -la .env* 2>/dev/null || echo "No .env files"
echo ""
if [ -f .env ]; then
    echo "內容 of .env:"
    cat .env
fi
if [ -f .env.production ]; then
    echo "內容 of .env.production:"
    cat .env.production
fi

echo ""
echo "=== 檢查 Git 狀態 ==="
cd ~/Link
git status --short

echo ""
echo "=== 檢查本地修改 ==="
git diff frontend/
EOF