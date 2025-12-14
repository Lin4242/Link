#!/bin/bash
set -e

# ====================================
# LINK VM 部署腳本
# 在 GCP VM 上執行此腳本
# ====================================

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}=====================================${NC}"
echo -e "${GREEN}     LINK VM 部署腳本${NC}"
echo -e "${GREEN}=====================================${NC}"
echo ""

# ====================================
# 1. 更新系統並安裝依賴
# ====================================
echo -e "${GREEN}1. 更新系統並安裝依賴...${NC}"
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl git build-essential nginx certbot python3-certbot-nginx postgresql postgresql-contrib

# ====================================
# 2. 安裝 Go 1.23
# ====================================
echo -e "${GREEN}2. 安裝 Go 1.23...${NC}"
if ! command -v go &> /dev/null; then
    wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
    rm go1.23.4.linux-amd64.tar.gz
fi
go version

# ====================================
# 3. 安裝 Node.js 22 和 pnpm
# ====================================
echo -e "${GREEN}3. 安裝 Node.js 22...${NC}"
if ! command -v node &> /dev/null; then
    curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
    sudo apt install -y nodejs
fi
node -v

echo -e "${GREEN}安裝 pnpm...${NC}"
sudo npm install -g pnpm

# ====================================
# 4. 設定 PostgreSQL
# ====================================
echo -e "${GREEN}4. 設定 PostgreSQL...${NC}"
sudo systemctl start postgresql
sudo systemctl enable postgresql

# 建立資料庫和使用者
sudo -u postgres psql << EOF
CREATE USER link WITH PASSWORD 'LinkSecurePassword2024';
CREATE DATABASE link OWNER link;
GRANT ALL PRIVILEGES ON DATABASE link TO link;
\q
EOF

echo -e "${GREEN}PostgreSQL 設定完成${NC}"

# ====================================
# 5. Clone 專案
# ====================================
echo -e "${GREEN}5. Clone 專案...${NC}"
cd ~
if [ ! -d "Link" ]; then
    git clone https://github.com/Lin4242/Link.git
fi
cd Link

# ====================================
# 6. 執行資料庫 Migration
# ====================================
echo -e "${GREEN}6. 執行資料庫 Migration...${NC}"
export DATABASE_URL="postgres://link:LinkSecurePassword2024@localhost:5432/link?sslmode=disable"
sudo -u postgres psql -d link -f backend/migrations/001_init.up.sql

# ====================================
# 7. 設定後端
# ====================================
echo -e "${GREEN}7. 設定後端...${NC}"
cd ~/Link/backend

# 建立 .env 檔案
cat > .env << EOF
SERVER_ADDR=:8443
SERVER_ENV=production
DATABASE_URL=postgres://link:LinkSecurePassword2024@localhost:5432/link?sslmode=disable
JWT_SECRET=$(openssl rand -hex 32)
JWT_EXPIRY=24h
CORS_ORIGINS=https://34.136.217.56,http://34.136.217.56
LOG_LEVEL=info
EOF

# 編譯後端
go mod download
go build -o bin/server ./cmd/server

# ====================================
# 8. 建立 systemd 服務
# ====================================
echo -e "${GREEN}8. 建立 systemd 服務...${NC}"
sudo tee /etc/systemd/system/link-backend.service > /dev/null << EOF
[Unit]
Description=LINK Backend Server
After=network.target postgresql.service

[Service]
Type=simple
User=$USER
WorkingDirectory=/home/$USER/Link/backend
ExecStart=/home/$USER/Link/backend/bin/server
Restart=always
RestartSec=5
EnvironmentFile=/home/$USER/Link/backend/.env

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable link-backend
sudo systemctl start link-backend

# ====================================
# 9. 建置前端
# ====================================
echo -e "${GREEN}9. 建置前端...${NC}"
cd ~/Link/frontend

# 建立環境變數
cat > .env << EOF
VITE_API_URL=http://34.136.217.56:8443
VITE_WS_URL=ws://34.136.217.56:8443/ws
EOF

# 安裝依賴並建置
pnpm install
pnpm build

# ====================================
# 10. 設定 Nginx
# ====================================
echo -e "${GREEN}10. 設定 Nginx...${NC}"
sudo tee /etc/nginx/sites-available/link > /dev/null << 'EOF'
server {
    listen 80;
    server_name 34.136.217.56;
    
    # 前端靜態檔案
    root /home/rocketmantw5516/Link/frontend/build;
    index index.html;
    
    # SPA 路由
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    # API 代理
    location /api/ {
        proxy_pass http://localhost:8443;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # WebSocket 代理
    location /ws {
        proxy_pass http://localhost:8443;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # 安全標頭
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
}
EOF

# 啟用網站
sudo ln -sf /etc/nginx/sites-available/link /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl reload nginx

# ====================================
# 11. 設定防火牆
# ====================================
echo -e "${GREEN}11. 設定防火牆...${NC}"
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 8443/tcp
sudo ufw --force enable

# ====================================
# 12. 顯示狀態
# ====================================
echo ""
echo -e "${GREEN}=====================================${NC}"
echo -e "${GREEN}     🎉 部署完成！${NC}"
echo -e "${GREEN}=====================================${NC}"
echo ""
echo -e "${GREEN}服務狀態：${NC}"
sudo systemctl status link-backend --no-pager | head -10
echo ""
echo -e "${GREEN}Nginx 狀態：${NC}"
sudo systemctl status nginx --no-pager | head -10
echo ""
echo -e "${GREEN}訪問網址：${NC}"
echo "  http://34.136.217.56"
echo ""
echo -e "${GREEN}測試 API：${NC}"
echo "  curl http://34.136.217.56:8443/health"
echo ""
echo -e "${YELLOW}後續步驟：${NC}"
echo "1. 設定域名 DNS 指向 34.136.217.56"
echo "2. 使用 certbot 申請 SSL 證書："
echo "   sudo certbot --nginx -d your-domain.com"
echo ""
echo -e "${GREEN}查看日誌：${NC}"
echo "  sudo journalctl -u link-backend -f"
echo "  sudo tail -f /var/log/nginx/error.log"