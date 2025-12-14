#!/bin/bash

echo "=== 測試完整註冊流程 ==="

# 使用剛剛生成的卡片
PRIMARY_TOKEN="5cab2973e20d0f53-1-a5fdc2e3"
BACKUP_TOKEN="5cab2973e20d0f53-2-296ef5c3"

echo "Primary Token: $PRIMARY_TOKEN"
echo "Backup Token: $BACKUP_TOKEN"
echo ""

# 測試註冊 API
echo "執行註冊請求..."
response=$(curl -s -X POST https://link.mcphub.tw/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{
    \"primary_token\": \"$PRIMARY_TOKEN\",
    \"backup_token\": \"$BACKUP_TOKEN\",
    \"nickname\": \"TestUser2\",
    \"password\": \"TestPassword123\"
  }")

echo "Response: $response"
echo ""

# 檢查是否成功
if echo "$response" | grep -q "error"; then
  echo "❌ 註冊失敗"
  echo ""
  echo "檢查後端日誌..."
  ssh rocketmantw5516@34.136.217.56 << 'EOF'
sudo journalctl -u link-backend --since "2 minutes ago" --no-pager | tail -20
EOF
else
  echo "✅ 註冊成功！"
  echo ""
  echo "檢查資料庫中的用戶..."
  ssh rocketmantw5516@34.136.217.56 << 'EOF'
sudo -u postgres psql -d link -c "SELECT id, nickname, created_at FROM users ORDER BY created_at DESC LIMIT 1;"
sudo -u postgres psql -d link -c "SELECT card_token, card_type, status FROM cards ORDER BY created_at DESC LIMIT 2;"
EOF
fi