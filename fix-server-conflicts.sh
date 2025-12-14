#!/bin/bash
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}修正 server 上的 Git 衝突...${NC}"

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd ~/Link

# 備份 .env (包含密碼)
cp backend/.env ~/link-env-backup.txt

# 放棄本地修改，強制同步 GitHub
git reset --hard HEAD
git pull origin main

# 恢復 .env
cp ~/link-env-backup.txt backend/.env

echo "衝突已解決，檔案已同步"
EOF

echo -e "${GREEN}✅ Server 已同步 GitHub${NC}"