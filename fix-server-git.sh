#!/bin/bash

SERVER_IP="34.136.217.56"  
SERVER_USER="rocketmantw5516"

echo "修復伺服器 Git 狀態..."

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd ~/Link

echo "1. 備份當前更改："
cp -r frontend/src/routes /tmp/routes-backup-$(date +%s)

echo "2. 重置到乾淨狀態："
git stash
git reset --hard origin/main

echo "3. 拉取最新代碼："
git pull origin main

echo "4. 檢查 update-public-key 文件："
ls -la frontend/src/routes/update-public-key/+page.svelte

echo "5. 檢查關鍵修復是否存在："
grep -n "instanceof Uint8Array" frontend/src/routes/update-public-key/+page.svelte | head -3

echo "6. 重建前端："
cd frontend
export PATH=$PATH:$HOME/.nvm/versions/node/v22.12.0/bin
pnpm build

echo "✅ 完成！"
EOF