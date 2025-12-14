#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 檢查 cards 表結構 ==="
sudo -u postgres psql -d link -c "\d cards"

echo ""
echo "=== 檢查所有卡片狀態 ==="
sudo -u postgres psql -d link -c "SELECT * FROM cards ORDER BY created_at DESC LIMIT 10;"

echo ""
echo "=== 檢查 card_pairs 表 ==="
sudo -u postgres psql -d link -c "SELECT * FROM card_pairs ORDER BY created_at DESC LIMIT 5;"

echo ""
echo "=== 檢查具體的卡片對 ==="
sudo -u postgres psql -d link -c "SELECT * FROM cards WHERE pair_uuid = 'd8544616490d074f';"

echo ""
echo "=== 檢查後端驗證邏輯 ==="
cd ~/Link/backend
grep -n "無效的卡片" internal/handler/auth.go
EOF