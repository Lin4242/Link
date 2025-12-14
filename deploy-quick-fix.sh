#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}快速部署 update-public-key 修復...${NC}"

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

# Copy and rebuild
ssh ${SERVER_USER}@${SERVER_IP} << 'ENDSSH'
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

cd ~/Link

# Pull latest changes
echo -e "${YELLOW}拉取最新代碼...${NC}"
git pull

# Rebuild frontend
echo -e "${YELLOW}重建前端...${NC}"
cd frontend
export PATH=$PATH:$HOME/.nvm/versions/node/v22.12.0/bin
pnpm build

echo -e "${GREEN}✅ 部署完成！${NC}"
ENDSSH

echo -e "${GREEN}✅ 修復已部署！${NC}"
echo ""
echo -e "${YELLOW}請 F 和 N 訪問:${NC}"
echo "https://link.mcphub.tw/update-public-key"
echo "輸入密碼更新公鑰"