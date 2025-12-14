#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 檢查用戶的公鑰 ==="
sudo -u postgres psql -d link -c "SELECT id, nickname, public_key, LENGTH(public_key) as key_length FROM users;"

echo ""
echo "=== 檢查前端公鑰生成邏輯 ==="
grep -n "generateKeyPair\|publicKey\|public_key" ~/Link/frontend/src/lib/crypto/*.ts | head -10

echo ""
echo "=== 檢查註冊時是否傳送公鑰 ==="
grep -B5 -A5 "public_key" ~/Link/frontend/src/routes/register/+page.svelte | head -20
EOF