#!/bin/bash
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}部署自動修復工具...${NC}"

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
set -e

cd ~/Link

echo "拉取最新代碼..."
git pull origin main

echo "重建前端..."
cd frontend
export PATH=$PATH:$HOME/.nvm/versions/node/v22.12.0/bin
pnpm build

echo "✅ 部署完成！"
EOF

echo -e "${GREEN}✅ 自動修復工具已部署！${NC}"
echo ""
echo -e "${YELLOW}=== 解決方案總結 ===${NC}"
echo ""
echo "1. F 和 N 的公鑰已設置臨時值（可以接收訊息）"
echo ""
echo "2. 請 F 和 N 執行以下步驟完全修復："
echo "   a) 清除瀏覽器所有資料（Cmd+Shift+Delete）"
echo "   b) 訪問: https://link.mcphub.tw/auto-fix-keys"
echo "   c) 等待自動生成新密鑰"
echo "   d) 成功後會自動跳轉到聊天頁面"
echo ""
echo "3. 如果還是不行，請手動操作："
echo "   a) 登出"
echo "   b) 清除所有瀏覽器資料"
echo "   c) 重新登入"
echo "   d) 系統會自動生成新密鑰"