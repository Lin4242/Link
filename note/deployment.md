# LINK 部署指南

## 系統需求

| 項目 | 最低需求 | 建議 |
|------|----------|------|
| OS | Linux / macOS | Ubuntu 22.04 LTS |
| CPU | 2 cores | 4 cores |
| RAM | 2 GB | 4 GB |
| Disk | 10 GB | 20 GB |

---

## 1. 安裝必要工具

### Ubuntu / Debian

```bash
# 更新套件
sudo apt update && sudo apt upgrade -y

# 安裝基本工具
sudo apt install -y curl git build-essential

# 安裝 Go 1.21+
wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc
go version

# 安裝 Node.js 20+ (使用 nvm)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
source ~/.bashrc
nvm install 22
node -v

# 安裝 pnpm
npm install -g pnpm

# 安裝 PostgreSQL 15
sudo apt install -y postgresql-15 postgresql-contrib-15
sudo systemctl enable postgresql
sudo systemctl start postgresql
```

### macOS (Homebrew)

```bash
# 安裝 Homebrew (如果沒有)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 安裝工具
brew install go node pnpm postgresql@15 mkcert

# 啟動 PostgreSQL
brew services start postgresql@15

# PATH 設定 (~/.zshrc)
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zshrc
echo 'export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

---

## 2. Clone 專案

```bash
git clone https://github.com/Lin4242/Link.git
cd Link
```

---

## 3. 資料庫設定

### 建立資料庫和使用者

```bash
# 切換到 postgres 使用者
sudo -u postgres psql

# 在 psql 中執行
CREATE USER link WITH PASSWORD 'your_secure_password';
CREATE DATABASE link OWNER link;
GRANT ALL PRIVILEGES ON DATABASE link TO link;
\q
```

### 執行 Migration

```bash
cd backend

# 設定資料庫連線 (或用環境變數)
export DATABASE_URL="postgres://link:your_secure_password@localhost:5432/link?sslmode=disable"

# 執行 migration
go run cmd/migrate/main.go up
```

**Migration 檔案位置**: `backend/migrations/`

---

## 4. TLS 憑證

### 開發環境 (mkcert)

```bash
# 安裝本地 CA
mkcert -install

# 產生憑證
mkdir -p certs
cd certs
mkcert localhost 127.0.0.1 ::1 192.168.1.99
# 產生 localhost+3.pem 和 localhost+3-key.pem
```

### 生產環境 (Let's Encrypt)

```bash
# 使用 certbot
sudo apt install certbot
sudo certbot certonly --standalone -d your-domain.com

# 憑證位置
# /etc/letsencrypt/live/your-domain.com/fullchain.pem
# /etc/letsencrypt/live/your-domain.com/privkey.pem
```

---

## 5. 後端部署

### 環境變數

建立 `backend/.env` 或設定系統環境變數：

```bash
# 資料庫
DATABASE_URL=postgres://link:password@localhost:5432/link?sslmode=disable

# JWT 密鑰 (請用隨機字串)
JWT_SECRET=your-super-secret-jwt-key-at-least-32-chars

# TLS 憑證路徑
TLS_CERT=/path/to/cert.pem
TLS_KEY=/path/to/key.pem

# 伺服器設定
SERVER_ADDR=:9443
ADMIN_KEY=your-admin-api-key

# 可選
LOG_LEVEL=info
```

### 編譯和執行

```bash
cd backend

# 編譯
go build -o bin/server ./cmd/server

# 執行
./bin/server

# 或使用 systemd (見下方)
```

### Systemd Service (生產環境)

建立 `/etc/systemd/system/link-backend.service`:

```ini
[Unit]
Description=LINK Backend Server
After=network.target postgresql.service

[Service]
Type=simple
User=link
WorkingDirectory=/home/link/Link/backend
ExecStart=/home/link/Link/backend/bin/server
Restart=always
RestartSec=5
Environment=DATABASE_URL=postgres://link:password@localhost:5432/link?sslmode=disable
Environment=JWT_SECRET=your-jwt-secret
Environment=TLS_CERT=/home/link/certs/cert.pem
Environment=TLS_KEY=/home/link/certs/key.pem
Environment=ADMIN_KEY=your-admin-key

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable link-backend
sudo systemctl start link-backend
sudo systemctl status link-backend
```

---

## 6. 前端部署

### 環境變數

建立 `frontend/.env`:

```bash
# API 伺服器位置
VITE_API_URL=https://your-domain.com:9443

# WebSocket 位置
VITE_WS_URL=wss://your-domain.com:9443/ws

# WebTransport (可選，留空則使用 WebSocket)
VITE_WT_URL=
```

### 建置

```bash
cd frontend

# 安裝依賴
pnpm install

# 建置
pnpm build

# 產出在 build/ 目錄
```

### 靜態檔案伺服器

**選項 1: Nginx**

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    root /home/link/Link/frontend/build;
    index index.html;

    # SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API proxy (可選)
    location /api/ {
        proxy_pass https://localhost:9443;
        proxy_ssl_verify off;
    }

    # WebSocket proxy (可選)
    location /ws {
        proxy_pass https://localhost:9443;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_ssl_verify off;
    }
}
```

**選項 2: 直接用後端 serve (開發/小規模)**

後端已經可以 serve 靜態檔案，只需將 `frontend/build` 複製到指定位置。

---

## 7. 防火牆設定

```bash
# Ubuntu UFW
sudo ufw allow 443/tcp   # HTTPS
sudo ufw allow 9443/tcp  # API/WebSocket
sudo ufw enable
```

---

## 8. 驗證部署

```bash
# 檢查後端健康
curl -k https://localhost:9443/health
# 應該回傳: OK

# 檢查資料庫連線
curl -k https://localhost:9443/api/v1/auth/check-card/test
# 應該回傳 JSON

# 檢查前端
curl -k https://localhost:443
# 應該回傳 HTML
```

---

## 9. 管理指令

### 產生 NFC 卡片配對

```bash
# 使用 Admin API
curl -k -X POST https://localhost:9443/api/v1/admin/cards/generate \
  -H "X-Admin-Key: your-admin-key" \
  -H "Content-Type: application/json"

# 回傳 primary_token 和 backup_token
```

### 查看日誌

```bash
# Systemd
journalctl -u link-backend -f

# 或直接看檔案
tail -f /var/log/link/server.log
```

### 資料庫備份

```bash
pg_dump -U link link > backup_$(date +%Y%m%d).sql
```

---

## 10. 常見問題

### Q: 憑證錯誤 (self-signed)
開發環境使用 mkcert 產生的憑證，瀏覽器需要先安裝本地 CA：
```bash
mkcert -install
```

### Q: WebSocket 連不上
檢查：
1. 防火牆是否開放 9443 port
2. Nginx proxy 設定是否正確 (需要 Upgrade header)
3. 憑證是否有效

### Q: 資料庫連線失敗
檢查：
1. PostgreSQL 是否運行: `systemctl status postgresql`
2. 使用者權限: `psql -U link -d link`
3. pg_hba.conf 是否允許連線

### Q: 前端 API 呼叫失敗
檢查：
1. VITE_API_URL 是否正確
2. CORS 設定 (後端已開啟)
3. 瀏覽器 Console 錯誤訊息

---

## 快速部署腳本

```bash
#!/bin/bash
set -e

echo "=== LINK 部署腳本 ==="

# 1. Clone
git clone https://github.com/Lin4242/Link.git
cd Link

# 2. 後端
cd backend
go build -o bin/server ./cmd/server
echo "後端編譯完成"

# 3. 前端
cd ../frontend
pnpm install
pnpm build
echo "前端建置完成"

# 4. 提示
echo ""
echo "=== 部署完成 ==="
echo "請設定環境變數後執行:"
echo "  cd backend && ./bin/server"
echo ""
echo "前端靜態檔案位於: frontend/build/"
```

---

## 目錄結構

```
Link/
├── backend/
│   ├── cmd/server/main.go      # 入口
│   ├── bin/server              # 編譯後執行檔
│   ├── migrations/             # 資料庫 migration
│   └── .env                    # 環境變數
├── frontend/
│   ├── build/                  # 建置產出
│   └── .env                    # 環境變數
├── certs/                      # TLS 憑證
└── note/                       # 開發筆記
```
