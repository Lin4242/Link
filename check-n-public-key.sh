#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

echo "檢查 N 用戶的公鑰狀態..."
echo "================================"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd ~/Link

echo "1. 檢查 N 用戶的資料："
sudo -u postgres psql link_db -c "
SELECT id, nickname, public_key, 
       CASE WHEN public_key IS NULL THEN '❌ 沒有公鑰' 
            WHEN public_key = '' THEN '❌ 公鑰為空' 
            ELSE '✅ 有公鑰' END as key_status,
       LEFT(public_key, 20) || '...' as key_preview
FROM users 
WHERE nickname IN ('N', 'F', '小安')
ORDER BY nickname;"

echo ""
echo "2. 檢查 N 的朋友關係："
sudo -u postgres psql link_db -c "
SELECT 
    u1.nickname as user1,
    u2.nickname as user2,
    f.status,
    f.created_at
FROM friendships f
JOIN users u1 ON f.user_id = u1.id
JOIN users u2 ON f.friend_id = u2.id
WHERE u1.nickname = 'N' OR u2.nickname = 'N'
ORDER BY f.created_at DESC;"

echo ""
echo "3. 檢查對話狀態："
sudo -u postgres psql link_db -c "
SELECT 
    c.id as conv_id,
    u1.nickname as user1,
    u2.nickname as user2,
    c.created_at,
    COUNT(m.id) as message_count
FROM conversations c
JOIN users u1 ON c.user1_id = u1.id
JOIN users u2 ON c.user2_id = u2.id
LEFT JOIN messages m ON m.conversation_id = c.id
WHERE u1.nickname IN ('N', 'F') OR u2.nickname IN ('N', 'F')
GROUP BY c.id, u1.nickname, u2.nickname, c.created_at
ORDER BY c.created_at DESC;"

echo ""
echo "4. 檢查最近的訊息錯誤（如果有）："
tail -20 backend/logs/app.log | grep -i "public_key\|peer.*no\|encrypt" || echo "沒有相關錯誤"

EOF