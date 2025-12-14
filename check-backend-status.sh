#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 檢查後端服務狀態 ==="
sudo systemctl status link-backend --no-pager | head -10

echo ""
echo "=== 檢查後端錯誤日誌 ==="
sudo journalctl -u link-backend -p err -n 20 --no-pager

echo ""
echo "=== 檢查後端環境變數 ==="
grep -E "SERVER_ADDR|TLS_" ~/Link/backend/.env

echo ""
echo "=== 直接測試後端 ==="
curl -v http://127.0.0.1:8443/health 2>&1 | head -20
EOF