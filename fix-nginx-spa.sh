#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 修復 Nginx SPA 路由 ==="

# 備份當前配置
sudo cp /etc/nginx/sites-available/link /etc/nginx/sites-available/link.backup-spa

# 更新配置 - 移除 $uri/ 避免目錄列表問題
sudo tee /etc/nginx/sites-available/link > /dev/null << 'NGINX_CONFIG'
server {
    server_name link.mcphub.tw;
    
    # 前端靜態檔案
    root /home/rocketmantw5516/Link/frontend/build;
    index index.html;
    
    # Card entry route - MUST be before the catch-all
    location /w/ {
        proxy_pass https://127.0.0.1:8443;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # API 代理
    location /api/ {
        proxy_pass https://127.0.0.1:8443;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # WebSocket 代理
    location /ws {
        proxy_pass https://127.0.0.1:8443;
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
        proxy_pass https://127.0.0.1:8443;
        proxy_http_version 1.1;
    }
    
    # Static assets
    location /_app/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
    
    # SPA routes - try file first, then fallback to index.html
    location / {
        try_files $uri $uri.html /index.html;
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

echo ""
echo "=== 測試配置 ==="
sudo nginx -t

echo ""
echo "=== 重載 Nginx ==="
sudo systemctl reload nginx

echo ""
echo "=== 測試 register 頁面 ==="
sleep 1
curl -s -I https://link.mcphub.tw/register?token=test | grep HTTP
EOF