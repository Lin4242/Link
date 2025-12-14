#!/bin/bash
set -e

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

echo "檢查前端 admin panel..."

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 檢查 admin panel 程式碼 ==="
grep -A2 "API_BASE" ~/Link/frontend/src/routes/admin/+page.svelte | head -5

echo ""
echo "=== 檢查建置後的檔案 ==="
ls -la ~/Link/frontend/build/ | head -10

echo ""
echo "=== 檢查 Nginx 錯誤 ==="
sudo tail -5 /var/log/nginx/error.log

echo ""
echo "=== 檢查後端日誌 ==="
sudo journalctl -u link-backend -n 10 --no-pager | grep -E "admin|Admin|password" || echo "No admin logs"
EOF