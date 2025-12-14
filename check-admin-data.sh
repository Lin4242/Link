#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 檢查資料庫中的卡片對 ==="
sudo -u postgres psql -d link -c "SELECT * FROM card_pairs ORDER BY created_at DESC;"

echo ""
echo "=== 檢查 Admin Handler 的記憶體狀態 ==="
echo "（後端重啟後記憶體中的卡片對會清空）"

echo ""
echo "=== 檢查後端最近重啟時間 ==="
sudo systemctl status link-backend --no-pager | grep -E "Active:|Main PID:"

echo ""
echo "=== 檢查所有已註冊的卡片 ==="
sudo -u postgres psql -d link -c "SELECT card_token, card_type, status, created_at FROM cards ORDER BY created_at DESC;"

echo ""
echo "=== 檢查 Admin API 回應 ==="
curl -s https://link.mcphub.tw/api/v1/admin/cards \
  -H "X-Admin-Password: link-admin-2024-secure" | python3 -m json.tool
EOF