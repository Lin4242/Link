#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 最近的註冊相關錯誤 ==="
sudo journalctl -u link-backend -n 50 --no-pager | grep -E "register|Register|ERROR|error|failed"

echo ""
echo "=== 最近的 API 請求 ==="
sudo tail -20 /var/log/nginx/access.log | grep -E "register|POST"

echo ""
echo "=== 檢查資料庫連線 ==="
cd ~/Link/backend
export DATABASE_URL="postgres://link:LinkSecurePassword2024@localhost:5432/link?sslmode=disable"
psql -U link -d link -c "SELECT 1;" 2>&1 || echo "Database connection failed"

echo ""
echo "=== 檢查後端完整日誌 ==="
sudo journalctl -u link-backend --since "5 minutes ago" --no-pager | tail -30
EOF