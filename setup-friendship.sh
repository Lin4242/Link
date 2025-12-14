#!/bin/bash

echo "=== 建立 F 和小安的好友關係 ==="

# F 的 ID 和 token
F_USER_ID="b7036fc7-b49e-43d3-941a-ab4fa572a39d"
F_TOKEN="5cab2973e20d0f53-1-a5fdc2e3"
F_PASSWORD="TestPassword123"

# 小安的 ID
XIAOAN_USER_ID="5c86c416-070b-476e-953a-5ea3ac54163b"
XIAOAN_TOKEN="2ea2040ad0d9de47-1-4546b88c"
XIAOAN_PASSWORD="link-admin-2024-secure"

# 1. F 登入取得 token
echo "1. F 登入..."
f_login_response=$(curl -s -X POST https://link.mcphub.tw/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{
    \"card_token\": \"$F_TOKEN\",
    \"password\": \"$F_PASSWORD\"
  }")

f_auth_token=$(echo "$f_login_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo "F's auth token: ${f_auth_token:0:20}..."

# 2. F 發送好友請求給小安
echo ""
echo "2. F 發送好友請求給小安..."
friend_request_response=$(curl -s -X POST https://link.mcphub.tw/api/v1/friends/request \
  -H "Authorization: Bearer $f_auth_token" \
  -H "Content-Type: application/json" \
  -d "{
    \"addressee_id\": \"$XIAOAN_USER_ID\"
  }")

echo "Response: $friend_request_response"
friendship_id=$(echo "$friend_request_response" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

# 3. 小安登入
echo ""
echo "3. 小安登入..."
xiaoan_login_response=$(curl -s -X POST https://link.mcphub.tw/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{
    \"card_token\": \"$XIAOAN_TOKEN\",
    \"password\": \"$XIAOAN_PASSWORD\"
  }")

xiaoan_auth_token=$(echo "$xiaoan_login_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo "小安's auth token: ${xiaoan_auth_token:0:20}..."

# 4. 小安接受好友請求
echo ""
echo "4. 小安接受好友請求..."
accept_response=$(curl -s -X POST "https://link.mcphub.tw/api/v1/friends/accept/$friendship_id" \
  -H "Authorization: Bearer $xiaoan_auth_token")

echo "Response: $accept_response"

# 5. 創建對話
echo ""
echo "5. 創建對話..."
conversation_response=$(curl -s -X POST https://link.mcphub.tw/api/v1/conversations \
  -H "Authorization: Bearer $f_auth_token" \
  -H "Content-Type: application/json" \
  -d "{
    \"participant_id\": \"$XIAOAN_USER_ID\"
  }")

echo "Response: $conversation_response"

# 6. 檢查結果
echo ""
echo "6. 檢查資料庫狀態..."
ssh rocketmantw5516@34.136.217.56 << 'EOF'
echo "=== 好友關係 ==="
sudo -u postgres psql -d link -c "SELECT * FROM friendships;"

echo ""
echo "=== 對話 ==="
sudo -u postgres psql -d link -c "SELECT * FROM conversations;"
EOF

echo ""
echo "✅ 好友關係建立完成！現在可以開始聊天了"