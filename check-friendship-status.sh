#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 檢查所有用戶 ==="
sudo -u postgres psql -d link -c "SELECT id, nickname, created_at FROM users ORDER BY created_at DESC;"

echo ""
echo "=== 檢查好友關係 ==="
sudo -u postgres psql -d link -c "SELECT * FROM friendships;"

echo ""
echo "=== 檢查對話 ==="
sudo -u postgres psql -d link -c "SELECT * FROM conversations;"

echo ""
echo "=== 檢查好友系統的 API ==="
grep -n "AddFriend\|GetFriends" ~/Link/backend/internal/handler/*.go

echo ""
echo "=== 檢查前端是否有新增好友功能 ==="
grep -r "add.*friend\|新增好友\|加好友" ~/Link/frontend/src --include="*.svelte" --include="*.ts" | head -5
EOF