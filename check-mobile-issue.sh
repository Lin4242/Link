#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 最近的 Nginx 訪問日誌 ==="
sudo tail -20 /var/log/nginx/access.log | grep -E "admin|Admin"

echo ""
echo "=== 最近的 Nginx 錯誤日誌 ==="
sudo tail -20 /var/log/nginx/error.log

echo ""
echo "=== 最近的後端日誌 ==="
sudo journalctl -u link-backend -n 30 --no-pager | grep -E "admin|error|ERROR|failed"

echo ""
echo "=== 檢查 CORS 設定 ==="
grep CORS ~/Link/backend/.env

echo ""
echo "=== 檢查前端建置的 admin 檔案 ==="
ls -la ~/Link/frontend/build/admin* 2>/dev/null || echo "No admin files"
ls -la ~/Link/frontend/build/_app/immutable/nodes/ | grep -E "3|admin"

echo ""
echo "=== 檢查 Browser Console 錯誤 ==="
echo "請在手機瀏覽器查看 Console 錯誤訊息"
EOF