#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 測試註冊 API ==="

# 模擬註冊請求
curl -X POST https://link.mcphub.tw/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "primaryToken": "d8544616490d074f-1-7bca2eca",
    "backupToken": "d8544616490d074f-2-6c659571",
    "nickname": "TestUser",
    "password": "TestPassword123"
  }' \
  -v 2>&1 | grep -E "< HTTP|{" | tail -5

echo ""
echo "=== 檢查後端日誌 ==="
sudo journalctl -u link-backend --since "1 minute ago" --no-pager | tail -20

echo ""
echo "=== 檢查資料庫中的用戶 ==="
sudo -u postgres psql -d link -c "SELECT id, nickname, created_at FROM users ORDER BY created_at DESC LIMIT 5;"

echo ""
echo "=== 檢查卡片狀態 ==="
sudo -u postgres psql -d link -c "SELECT card_number, card_type, status FROM cards WHERE pair_uuid = 'd8544616490d074f';"
EOF