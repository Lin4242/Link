#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 檢查當前的 SPA 路由配置 ==="
grep -A2 "location /" /etc/nginx/sites-available/link | tail -3

echo ""
echo "=== 檢查是否有 register 目錄 ==="
ls -la ~/Link/frontend/build/ | grep register

echo ""
echo "=== 測試直接訪問 register 頁面 ==="
curl -s https://link.mcphub.tw/register | head -5
EOF