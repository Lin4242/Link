#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

echo "更新 F 和 N 的真實公鑰..."
echo "================================"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd ~/Link

echo "1. 更新資料庫中的公鑰..."
sudo -u postgres psql -d link << SQL
-- 更新 F 的真實公鑰
UPDATE users 
SET public_key = '491IZL9EeBgQER+zM8q1DjZxisq+1F4ONmvucU4Xxmc='
WHERE nickname = 'F';

-- 更新 N 的真實公鑰
UPDATE users 
SET public_key = 'pNEEV/aCJ4K0XUS5FGFX9hXV5+eh/IJeAEL76h/sqzs='
WHERE nickname = 'N';

-- 檢查結果
SELECT nickname,
       CASE WHEN public_key IS NULL THEN '❌ NULL'
            WHEN public_key = '' THEN '❌ 空'
            ELSE '✅ 已更新 (' || LENGTH(public_key) || ' 字元)'
       END as status,
       LEFT(public_key, 20) || '...' as key_preview
FROM users
WHERE nickname IN ('F', 'N', '小安')
ORDER BY nickname;
SQL

echo ""
echo "2. 保存密鑰資訊供 F 和 N 使用..."
cat > ~/f-n-keys.txt << 'KEYS'
================================
F 的帳號資訊:
暱稱: F
密碼: 000000
公鑰: 491IZL9EeBgQER+zM8q1DjZxisq+1F4ONmvucU4Xxmc=
私鑰: VpRfgl9QTuNwxtHJ++EjyeOZTqODz0I2pekKLfJQ1tg=

================================
N 的帳號資訊:
暱稱: N
密碼: 999999
公鑰: pNEEV/aCJ4K0XUS5FGFX9hXV5+eh/IJeAEL76h/sqzs=
私鑰: 2ar9400RoPfWMGXDlACx3x25hl2JCkIo44Adsuo5YgI=
================================
KEYS

echo "密鑰已保存到 ~/f-n-keys.txt"
echo ""
echo "✅ 公鑰已更新完成！"
echo ""
echo "現在 F 和 N 需要："
echo "1. 清除瀏覽器所有資料"
echo "2. 重新登入"
echo "3. 系統會使用新的公鑰"

EOF