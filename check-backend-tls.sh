#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 檢查後端 TLS 設定 ==="
cd ~/Link/backend
grep -E "TLS|tls|cert" .env || echo "No TLS settings in .env"

echo ""
echo "=== 檢查憑證檔案 ==="
ls -la certs/ 2>/dev/null || echo "No certs directory"

echo ""
echo "=== 測試 HTTPS 連線 ==="
curl -k https://127.0.0.1:8443/health 2>&1 | head -5

echo ""
echo "=== 檢查後端程式碼中的 TLS 邏輯 ==="
grep -A5 "ListenTLS\|Listen" bin/server 2>/dev/null | head -10 || echo "Binary check"

echo ""
echo "=== 最近的後端日誌 ==="
sudo journalctl -u link-backend -n 20 --no-pager | grep -E "HTTPS|HTTP|TLS|starting" || echo "No relevant logs"
EOF