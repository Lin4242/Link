#!/bin/bash
set -e

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

echo "重新建置前端（使用正確的環境變數）..."

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
export PATH=$PATH:$HOME/.nvm/versions/node/v22.12.0/bin
cd ~/Link/frontend

echo "=== 檢查當前源碼的 API_BASE ==="
grep -A2 "API_BASE" src/routes/admin/+page.svelte

echo ""
echo "=== 設定正確的環境變數 ==="
cat > .env.production << 'ENV_FILE'
VITE_API_URL=
VITE_WS_URL=
ENV_FILE

echo "環境變數設定為空（將使用 window.location.origin）"

echo ""
echo "=== 重新建置前端 ==="
pnpm build

echo ""
echo "=== 驗證建置後的檔案 ==="
grep -o "window.location.origin" build/_app/immutable/nodes/3.*.js | head -2 || echo "檢查 API_BASE 邏輯"

echo ""
echo "建置完成！"
EOF

echo "前端已重新建置"