#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}部署公鑰更新修復...${NC}"

# Server details
SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

# Step 1: Copy files
echo -e "${YELLOW}複製文件到伺服器...${NC}"

# Copy update-public-key page
scp -r frontend/src/routes/update-public-key ${SERVER_USER}@${SERVER_IP}:/tmp/

# Copy updated backend file
scp backend/internal/handler/update_public_key.go ${SERVER_USER}@${SERVER_IP}:/tmp/

# Step 2: Apply on server
ssh ${SERVER_USER}@${SERVER_IP} << 'ENDSSH'
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}應用更新...${NC}"

# Copy frontend files
cd ~/Link/frontend
cp -r /tmp/update-public-key src/routes/

# Copy backend files  
cd ~/Link/backend
cp /tmp/update_public_key.go internal/handler/

# Rebuild frontend
echo -e "${YELLOW}重建前端...${NC}"
cd ~/Link/frontend
export PATH=$PATH:$HOME/.nvm/versions/node/v22.12.0/bin
pnpm build

# Rebuild backend
echo -e "${YELLOW}重建後端...${NC}"
cd ~/Link/backend
go build -o link-backend cmd/server/main.go

# Restart backend
echo -e "${YELLOW}重啟後端...${NC}"
sudo systemctl restart link-backend

echo -e "${GREEN}✅ 部署完成！${NC}"

# Clean up
rm -rf /tmp/update-public-key /tmp/update_public_key.go

ENDSSH

echo -e "${GREEN}✅ 公鑰更新修復已部署！${NC}"
echo ""
echo -e "${YELLOW}使用方法：${NC}"
echo "1. F 和 N 用戶訪問: https://link.mcphub.tw/update-public-key"
echo "2. 輸入登入密碼"
echo "3. 系統會自動更新公鑰到資料庫"
echo "4. 更新後就可以正常收發加密訊息了"