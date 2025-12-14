#!/bin/bash

echo "=== 測試登入流程 ==="

PRIMARY_TOKEN="5cab2973e20d0f53-1-a5fdc2e3"
BACKUP_TOKEN="5cab2973e20d0f53-2-296ef5c3"
PASSWORD="TestPassword123"

echo "1. 測試主卡登入..."
response=$(curl -s -X POST https://link.mcphub.tw/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{
    \"card_token\": \"$PRIMARY_TOKEN\",
    \"password\": \"$PASSWORD\"
  }")

if echo "$response" | grep -q "token"; then
  echo "✅ 主卡登入成功"
  token=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
  echo "Token: ${token:0:20}..."
else
  echo "❌ 主卡登入失敗"
  echo "$response"
fi

echo ""
echo "2. 測試副卡登入..."
response=$(curl -s -X POST https://link.mcphub.tw/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{
    \"card_token\": \"$BACKUP_TOKEN\",
    \"password\": \"$PASSWORD\"
  }")

if echo "$response" | grep -q "token"; then
  echo "✅ 副卡登入成功"
  token=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
  echo "Token: ${token:0:20}..."
else
  echo "❌ 副卡登入失敗"
  echo "$response"
fi

echo ""
echo "3. 測試錯誤密碼..."
response=$(curl -s -X POST https://link.mcphub.tw/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{
    \"card_token\": \"$PRIMARY_TOKEN\",
    \"password\": \"WrongPassword\"
  }")

if echo "$response" | grep -q "error"; then
  echo "✅ 正確拒絕錯誤密碼"
  echo "$response"
else
  echo "❌ 應該要拒絕錯誤密碼"
fi