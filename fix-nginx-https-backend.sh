#!/bin/bash
set -e

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

echo "修復 Nginx 使用 HTTPS 後端..."

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
# 更新 Nginx 配置使用 HTTPS
sudo sed -i 's|proxy_pass http://127.0.0.1:8443|proxy_pass https://127.0.0.1:8443|g' /etc/nginx/sites-available/link

# 顯示修改後的配置
echo "=== 修改後的 proxy_pass 設定 ==="
grep proxy_pass /etc/nginx/sites-available/link

# 測試配置
sudo nginx -t

# 重新載入 Nginx
sudo systemctl reload nginx

echo ""
echo "=== 測試 API ==="
sleep 1
curl -s https://link.mcphub.tw/health
echo ""
curl -s https://link.mcphub.tw/api/v1/admin/cards -H "X-Admin-Password: link-admin-2024-secure" | head -20
EOF