#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 前端 build 目錄結構 ==="
cd ~/Link/frontend/build
find . -maxdepth 2 -type f -name "*.html" | sort
echo ""
find . -maxdepth 2 -type d | sort

echo ""
echo "=== 檢查 fallback 設定 ==="
cat ~/Link/frontend/svelte.config.js | grep -A5 fallback
EOF