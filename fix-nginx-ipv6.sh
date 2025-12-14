#!/bin/bash
set -e

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

echo "修復 Nginx IPv6 問題..."

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
# 備份現有配置
sudo cp /etc/nginx/sites-available/link /etc/nginx/sites-available/link.backup

# 修改配置，強制使用 IPv4 127.0.0.1 而非 localhost
sudo sed -i 's|proxy_pass http://localhost:8443|proxy_pass http://127.0.0.1:8443|g' /etc/nginx/sites-available/link

# 顯示修改後的配置
echo "=== 修改後的 proxy_pass 設定 ==="
grep proxy_pass /etc/nginx/sites-available/link

# 測試配置
sudo nginx -t

# 重新載入 Nginx
sudo systemctl reload nginx

echo "Nginx 已重新載入"
EOF