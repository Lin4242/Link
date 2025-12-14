#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

echo "驗證 update-public-key 部署..."
echo "================================"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd ~/Link

echo "1. 檢查文件是否包含修復："
echo "   檢查 instanceof Uint8Array 修復："
grep -c "instanceof Uint8Array" frontend/src/routes/update-public-key/+page.svelte
echo ""

echo "2. 檢查編譯後的文件："
ls -la frontend/build/update-public-key.html
echo ""

echo "3. 檢查最後修改時間："
stat -c "最後修改: %y" frontend/build/update-public-key.html
echo ""

echo "4. 測試頁面載入："
curl -s https://link.mcphub.tw/update-public-key | grep -o "<title>.*</title>" || echo "無法取得標題"
echo ""

echo "5. 檢查前端服務狀態："
ps aux | grep "pnpm\|node" | grep -v grep | head -2
echo ""

echo "6. 查看最新的前端日誌："
ls -la frontend/build/_app/immutable/nodes/ | tail -5

EOF