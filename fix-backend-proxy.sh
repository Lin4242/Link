#!/bin/bash
set -e

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

echo "修復後端代理問題..."

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
# 檢查後端監聽的地址
echo "=== 檢查後端監聽地址 ==="
ss -tlnp | grep 8443 || echo "Port 8443 not listening"

# 檢查後端是否使用 HTTPS
echo ""
echo "=== 測試後端連線 ==="
curl -k https://127.0.0.1:8443/health || echo "HTTPS failed"
curl http://127.0.0.1:8443/health || echo "HTTP failed"

# 修改 Nginx 配置使用正確的協議
sudo tee /etc/nginx/sites-available/link > /dev/null << 'NGINX_CONFIG'
server {
    server_name link.mcphub.tw;
    
    # 前端靜態檔案
    root /home/rocketmantw5516/Link/frontend/build;
    index index.html;
    
    # SPA 路由
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    # API 代理 - 使用 HTTP 而非 HTTPS
    location /api/ {
        proxy_pass http://127.0.0.1:8443;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # WebSocket 代理
    location /ws {
        proxy_pass http://127.0.0.1:8443;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # Health check endpoint
    location /health {
        proxy_pass http://127.0.0.1:8443;
        proxy_http_version 1.1;
    }
    
    # 安全標頭
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    listen 443 ssl; # managed by Certbot
    ssl_certificate /etc/letsencrypt/live/link.mcphub.tw/fullchain.pem; # managed by Certbot
    ssl_certificate_key /etc/letsencrypt/live/link.mcphub.tw/privkey.pem; # managed by Certbot
    include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # managed by Certbot
}

server {
    if ($host = link.mcphub.tw) {
        return 301 https://$host$request_uri;
    } # managed by Certbot

    listen 80;
    server_name link.mcphub.tw;
    return 404; # managed by Certbot
}
NGINX_CONFIG

# 測試並重新載入
sudo nginx -t
sudo systemctl reload nginx

echo "Nginx 已更新為使用 HTTP 代理"
EOF