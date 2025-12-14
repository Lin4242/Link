#!/bin/bash

echo "=== 創建小安測試帳號 ==="

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

# 1. 先在 Admin Panel 生成新的卡片對
echo "1. 生成小安的卡片對..."
response=$(curl -s -X POST https://link.mcphub.tw/api/v1/admin/cards/generate \
  -H "X-Admin-Password: link-admin-2024-secure" \
  -H "Content-Type: application/json")

primary_token=$(echo "$response" | grep -o '"first_token":"[^"]*' | cut -d'"' -f4)
backup_token=$(echo "$response" | grep -o '"second_token":"[^"]*' | cut -d'"' -f4)

echo "小安的主卡 Token: $primary_token"
echo "小安的副卡 Token: $backup_token"
echo ""

# 2. 註冊小安帳號
echo "2. 註冊小安帳號..."
register_response=$(curl -s -X POST https://link.mcphub.tw/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{
    \"primary_token\": \"$primary_token\",
    \"backup_token\": \"$backup_token\",
    \"nickname\": \"小安\",
    \"password\": \"link-admin-2024-secure\"
  }")

if echo "$register_response" | grep -q "error"; then
  echo "❌ 註冊失敗："
  echo "$register_response"
  exit 1
fi

echo "✅ 小安帳號創建成功！"
user_id=$(echo "$register_response" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "User ID: $user_id"
echo ""

# 3. 驗證帳號
echo "3. 驗證小安可以登入..."
login_response=$(curl -s -X POST https://link.mcphub.tw/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{
    \"card_token\": \"$primary_token\",
    \"password\": \"link-admin-2024-secure\"
  }")

if echo "$login_response" | grep -q "token"; then
  echo "✅ 小安登入成功"
  token=$(echo "$login_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
else
  echo "❌ 小安登入失敗"
  echo "$login_response"
  exit 1
fi

# 4. 確認資料庫中的資料
echo ""
echo "4. 確認資料庫資料..."
ssh ${SERVER_USER}@${SERVER_IP} << EOF
echo "=== 小安的用戶資料 ==="
sudo -u postgres psql -d link -c "SELECT id, nickname, created_at FROM users WHERE nickname = '小安';"

echo ""
echo "=== 小安的卡片資料 ==="
sudo -u postgres psql -d link -c "SELECT card_token, card_type, status FROM cards WHERE user_id = '$user_id';"
EOF

echo ""
echo "=========================================="
echo "小安帳號資訊總結："
echo "=========================================="
echo "暱稱: 小安"
echo "密碼: link-admin-2024-secure"
echo "主卡 Token: $primary_token"
echo "副卡 Token: $backup_token"
echo "主卡 URL: https://link.mcphub.tw/w/$primary_token"
echo "副卡 URL: https://link.mcphub.tw/w/$backup_token"
echo "=========================================="
echo ""
echo "✅ 小安帳號已經準備就緒，可以用來測試訊息功能！"