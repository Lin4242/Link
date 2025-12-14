#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

echo "檢查用戶公鑰和訊息狀態..."
echo "================================"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd ~/Link

echo "1. 檢查所有用戶的公鑰狀態："
sudo -u postgres psql -d link -c "
SELECT id, nickname, 
       CASE WHEN public_key IS NULL THEN '❌ 沒有公鑰' 
            WHEN public_key = '' THEN '❌ 公鑰為空' 
            ELSE '✅ 有公鑰 (' || LENGTH(public_key) || ' 字元)' END as key_status,
       created_at::date
FROM users 
WHERE nickname IN ('N', 'F', '小安')
ORDER BY created_at DESC;"

echo ""
echo "2. 檢查最近的訊息（最新 10 筆）："
sudo -u postgres psql -d link -c "
SELECT 
    m.id,
    u1.nickname as sender,
    u2.nickname as receiver,
    LENGTH(m.encrypted_content) as content_size,
    m.created_at::timestamp(0) as sent_at,
    m.delivered_at IS NOT NULL as delivered,
    m.read_at IS NOT NULL as read
FROM messages m
JOIN conversations c ON m.conversation_id = c.id
JOIN users u1 ON m.sender_id = u1.id
JOIN users u2 ON 
    CASE 
        WHEN m.sender_id = c.user1_id THEN c.user2_id
        ELSE c.user1_id
    END = u2.id
WHERE u1.nickname IN ('N', 'F', '小安') OR u2.nickname IN ('N', 'F', '小安')
ORDER BY m.created_at DESC
LIMIT 10;"

echo ""
echo "3. 檢查朋友關係："
sudo -u postgres psql -d link -c "
SELECT 
    u1.nickname as user1,
    u2.nickname as user2,
    f.status,
    f.created_at::date
FROM friendships f
JOIN users u1 ON f.user_id = u1.id
JOIN users u2 ON f.friend_id = u2.id
WHERE (u1.nickname IN ('N', 'F', '小安') OR u2.nickname IN ('N', 'F', '小安'))
  AND f.status = 'accepted'
ORDER BY f.created_at DESC;"

echo ""
echo "4. 檢查 WebTransport 連線狀態："
ps aux | grep node | grep -v grep | head -2 || echo "Node 進程狀態"
netstat -tlnp 2>/dev/null | grep 4433 || echo "Port 4433 狀態"

EOF