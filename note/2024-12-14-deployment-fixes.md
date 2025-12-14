# 2024-12-14 部署修復記錄

## 重要問題與解決方案

### 1. ❌ 硬編碼 IP 問題

**症狀**：
- Admin Panel 連線失敗
- NFC 卡片重定向到錯誤的 IP (192.168.1.99)
- WebSocket 連線失敗

**發現的硬編碼位置**：
```
frontend/src/lib/stores/transport.svelte.ts:4: wss://192.168.1.99:9443/ws
frontend/src/lib/api/client.ts:3: https://192.168.1.99:9443
frontend/src/routes/admin/+page.svelte:19: http://34.136.217.56:8443/api/v1
backend/internal/handler/auth.go:94: https://192.168.1.99:5173
backend/internal/handler/admin.go:62-63: https://192.168.1.99:9443/w/
```

**解決方案**：
1. 前端使用動態 URL：
   - API: `window.location.origin`
   - WebSocket: `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`
2. 後端使用環境變數：
   - 新增 `BASE_URL` 到 config
   - AuthHandler 和 AdminHandler 使用 config.BaseURL

### 2. ❌ Nginx 代理配置錯誤

**症狀**：
- API 返回 502 Bad Gateway
- Admin Panel 無法連線

**問題**：
- Nginx 使用 `proxy_pass http://` 但後端實際運行 HTTPS

**解決方案**：
```nginx
# 錯誤
proxy_pass http://127.0.0.1:8443;

# 正確
proxy_pass https://127.0.0.1:8443;
```

### 3. ❌ NFC 卡片路由未設定

**症狀**：
- 掃描 NFC 卡片返回 404
- `/w/` 路由被當成靜態檔案

**問題**：
- Nginx 沒有設定 `/w/` 路由代理

**解決方案**：
```nginx
# Card entry route - MUST be before the catch-all
location /w/ {
    proxy_pass https://127.0.0.1:8443;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

### 4. ❌ SPA 路由 403 Forbidden

**症狀**：
- 訪問 `/register?token=xxx` 返回 403 Forbidden
- Nginx 錯誤：directory index of "/register/" is forbidden

**問題**：
- `try_files $uri $uri/` 導致 Nginx 嘗試列出目錄

**解決方案**：
```nginx
# 錯誤
location / {
    try_files $uri $uri/ /index.html;
}

# 正確
location / {
    try_files $uri $uri.html /index.html;
}
```

### 5. ❌ 環境變數混亂

**症狀**：
- 前端建置時使用錯誤的 API URL
- 部署後還是連到舊 IP

**問題**：
- Server 上有舊的 .env 檔案
- 建置時沒有清空環境變數

**解決方案**：
在 `deploy-from-github.sh` 中加入：
```bash
# 確保使用正確的環境變數（留空以使用 window.location.origin）
rm -f .env .env.production
echo "VITE_API_URL=" > .env.production
echo "VITE_WS_URL=" >> .env.production
```

## 完整的 Nginx 配置

```nginx
server {
    server_name link.mcphub.tw;
    
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
    
    # Health check
    location /health {
        proxy_pass https://127.0.0.1:8443;
        proxy_http_version 1.1;
    }
    
    # Static assets with cache
    location /_app/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
    
    # SPA routes
    location / {
        try_files $uri $uri.html /index.html;
    }
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # SSL configuration (managed by Certbot)
    listen 443 ssl;
    ssl_certificate /etc/letsencrypt/live/link.mcphub.tw/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/link.mcphub.tw/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
}
```

## 部署檢查清單

部署前務必檢查：

- [ ] 沒有硬編碼的 IP 地址
- [ ] 環境變數正確設定
- [ ] Nginx 配置包含所有必要的路由
- [ ] 後端使用 HTTPS，Nginx proxy_pass 也用 https://
- [ ] SPA 路由配置正確（避免 403 錯誤）
- [ ] `/w/` 路由在 Nginx 中正確代理
- [ ] WebSocket 路由設定正確

## 快速修復指令

```bash
# 搜尋硬編碼 IP
grep -r "192.168\|localhost:[0-9]" --include="*.go" --include="*.ts" --include="*.svelte" .

# 檢查 Nginx 配置
sudo nginx -t
sudo nginx -s reload

# 檢查服務狀態
sudo systemctl status link-backend
sudo journalctl -u link-backend -f

# 測試 NFC 卡片路由
curl -v https://link.mcphub.tw/w/test-token

# 從 GitHub 部署
./deploy-from-github.sh
```