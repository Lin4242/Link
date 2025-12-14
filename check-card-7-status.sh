#!/bin/bash

echo "=== 檢查第 7 組卡片狀態 ==="

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

# 先查看最新的卡片對
echo "1. 查詢最新的卡片對（應該是第 7 組）..."
ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
sudo -u postgres psql -d link -c "SELECT * FROM card_pairs ORDER BY created_at DESC LIMIT 1;"

echo ""
echo "=== 檢查這組卡片是否已被註冊 ==="
# 獲取最新卡片的 token
latest_token=$(sudo -u postgres psql -d link -t -c "SELECT primary_token FROM card_pairs ORDER BY created_at DESC LIMIT 1;" | tr -d ' ')
echo "Latest token: $latest_token"

echo ""
echo "=== 檢查 cards 表中是否有這個 token ==="
sudo -u postgres psql -d link -c "SELECT * FROM cards WHERE card_token LIKE '%$latest_token%';"

echo ""
echo "=== 檢查所有已註冊的卡片 ==="
sudo -u postgres psql -d link -c "SELECT card_token, card_type, status, created_at FROM cards ORDER BY created_at DESC;"
EOF

echo ""
echo "2. 測試卡片 API 狀態..."
# 你需要提供第 7 組卡片的 token
read -p "請輸入第 7 組卡片的 primary token (或按 Enter 跳過): " card_token
if [ ! -z "$card_token" ]; then
  echo "檢查卡片狀態: $card_token"
  curl -s https://link.mcphub.tw/api/v1/auth/check-card/$card_token | python3 -m json.tool
fi