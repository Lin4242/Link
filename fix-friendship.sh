#!/bin/bash

echo "=== 修正好友關係 ==="

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

# 先清理舊的錯誤請求，然後直接在資料庫建立好友關係
ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "1. 清理現有的好友請求..."
sudo -u postgres psql -d link << SQL
-- 刪除錯誤的好友請求
DELETE FROM friendships WHERE status = 'pending';

-- 直接建立 F 和小安的好友關係
INSERT INTO friendships (requester_id, addressee_id, status)
VALUES (
  'b7036fc7-b49e-43d3-941a-ab4fa572a39d', -- F 的 ID
  '5c86c416-070b-476e-953a-5ea3ac54163b', -- 小安的 ID  
  'accepted'
);

-- 創建對話
INSERT INTO conversations (participant_1, participant_2)
VALUES (
  'b7036fc7-b49e-43d3-941a-ab4fa572a39d', -- F 的 ID
  '5c86c416-070b-476e-953a-5ea3ac54163b'  -- 小安的 ID
);
SQL

echo ""
echo "2. 檢查結果..."
sudo -u postgres psql -d link -c "SELECT f.*, u1.nickname as requester, u2.nickname as addressee FROM friendships f JOIN users u1 ON f.requester_id = u1.id JOIN users u2 ON f.addressee_id = u2.id;"

echo ""
sudo -u postgres psql -d link -c "SELECT c.*, u1.nickname as user1, u2.nickname as user2 FROM conversations c JOIN users u1 ON c.participant_1 = u1.id JOIN users u2 ON c.participant_2 = u2.id;"
EOF

echo ""
echo "✅ 好友關係和對話已建立！請重新整理聊天頁面"