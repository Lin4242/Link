#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 最近的 403 錯誤 ==="
sudo tail -30 /var/log/nginx/access.log | grep -E "403|/w/" | tail -10

echo ""
echo "=== Nginx 錯誤日誌 ==="
sudo tail -20 /var/log/nginx/error.log

echo ""
echo "=== 檢查 /w/ 路由在 Nginx ==="
grep -B2 -A5 "location /w/" /etc/nginx/sites-available/link

echo ""
echo "=== 後端日誌中的 /w/ 請求 ==="
sudo journalctl -u link-backend -n 30 --no-pager | grep -E "/w/|403|card|token"

echo ""
echo "=== 測試後端直接連線 ==="
curl -v https://127.0.0.1:8443/w/test-token -k 2>&1 | grep -E "< HTTP|< Location|403"
EOF