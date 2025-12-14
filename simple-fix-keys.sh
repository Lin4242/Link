#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

echo "簡單修復：為 F 和 N 設置測試公鑰..."
echo "================================"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd ~/Link

echo "1. 使用小安的公鑰格式生成測試密鑰..."

# 這些是測試用的 Base64 編碼公鑰（44字元）
# 實際使用時應該由客戶端生成
F_TEST_KEY="FtestPublicKey1234567890abcdefghijklmnopqrs"
N_TEST_KEY="NtestPublicKey1234567890abcdefghijklmnopqrs"

echo "F 測試公鑰: $F_TEST_KEY"
echo "N 測試公鑰: $N_TEST_KEY"

echo ""
echo "2. 更新資料庫..."
sudo -u postgres psql -d link << SQL
-- 先查看當前狀態
SELECT nickname, 
       CASE WHEN public_key IS NULL THEN 'NULL'
            WHEN public_key = '' THEN '空字串'
            ELSE '有值 (' || LENGTH(public_key) || ' 字元)'
       END as before_status
FROM users 
WHERE nickname IN ('F', 'N', '小安');

-- 更新 F 的公鑰
UPDATE users 
SET public_key = '$F_TEST_KEY'
WHERE nickname = 'F';

-- 更新 N 的公鑰
UPDATE users 
SET public_key = '$N_TEST_KEY'
WHERE nickname = 'N';

-- 檢查更新後狀態
SELECT nickname,
       CASE WHEN public_key IS NULL THEN '❌ NULL'
            WHEN public_key = '' THEN '❌ 空字串'
            ELSE '✅ 有值 (' || LENGTH(public_key) || ' 字元)'
       END as after_status,
       public_key
FROM users
WHERE nickname IN ('F', 'N', '小安')
ORDER BY nickname;
SQL

echo ""
echo "================================"
echo "✅ 測試公鑰已設置！"
echo ""
echo "注意：這只是讓 F 和 N 可以被其他人發送訊息"
echo "F 和 N 仍需要："
echo "1. 清除瀏覽器資料（localStorage, IndexedDB）"
echo "2. 重新登入"
echo "3. 系統會自動生成新的密鑰對"
echo "================================"

EOF