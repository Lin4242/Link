#!/bin/bash

echo "=== 測試新的卡片生成流程 ==="

# 1. 生成新卡片對
echo "1. 生成新的卡片對..."
response=$(curl -s -X POST https://link.mcphub.tw/api/v1/admin/cards/generate \
  -H "X-Admin-Password: link-admin-2024-secure" \
  -H "Content-Type: application/json")

echo "Response: $response"

# 提取 tokens
primary_token=$(echo "$response" | grep -o '"first_token":"[^"]*' | cut -d'"' -f4)
backup_token=$(echo "$response" | grep -o '"second_token":"[^"]*' | cut -d'"' -f4)

echo ""
echo "Primary Token: $primary_token"
echo "Backup Token: $backup_token"

# 2. 檢查卡片是否存入資料庫
echo ""
echo "2. 檢查資料庫中的卡片對..."
ssh rocketmantw5516@34.136.217.56 << EOF
sudo -u postgres psql -d link -c "SELECT * FROM card_pairs WHERE primary_token = '$primary_token';"
EOF

# 3. 檢查卡片狀態
echo ""
echo "3. 檢查卡片狀態..."
curl -s https://link.mcphub.tw/api/v1/auth/check-card/$primary_token | python3 -m json.tool

echo ""
echo "4. 檢查配對卡片狀態..."
curl -s https://link.mcphub.tw/api/v1/auth/check-card/$backup_token | python3 -m json.tool