#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 最近的 404 錯誤 ==="
sudo tail -30 /var/log/nginx/access.log | grep -E "404|/w/"

echo ""
echo "=== 檢查 /w/ 路由設定 ==="
grep -A5 "/w/" ~/Link/backend/internal/handler/routes.go

echo ""
echo "=== 檢查 Nginx 對 /w/ 的處理 ==="
grep -E "/w/|location /" /etc/nginx/sites-available/link

echo ""
echo "=== 最近的後端日誌 ==="
sudo journalctl -u link-backend -n 20 --no-pager | grep -E "/w/|404|card"

echo ""
echo "=== 檢查資料庫中的卡片 ==="
cd ~/Link/backend
export DATABASE_URL="postgres://link:LinkSecurePassword2024@localhost:5432/link?sslmode=disable"
psql -d link -U link -c "SELECT card_token, status FROM cards LIMIT 10;" 2>/dev/null || echo "Database check failed"
EOF