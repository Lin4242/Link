#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

echo "手動修復 F 和 N 的公鑰..."
echo "================================"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd ~/Link

echo "1. 生成新的密鑰對給 F 和 N..."
cat > /tmp/generate-keys.js << 'JSEOF'
const nacl = require('tweetnacl');
const { encodeBase64 } = require('tweetnacl-util');

// 生成新的密鑰對
const keypairF = nacl.box.keyPair();
const keypairN = nacl.box.keyPair();

console.log('F 的公鑰:', encodeBase64(keypairF.publicKey));
console.log('F 的私鑰:', encodeBase64(keypairF.secretKey));
console.log('');
console.log('N 的公鑰:', encodeBase64(keypairN.publicKey));
console.log('N 的私鑰:', encodeBase64(keypairN.secretKey));
JSEOF

cd frontend
npm list tweetnacl 2>/dev/null || npm install tweetnacl tweetnacl-util
node /tmp/generate-keys.js > /tmp/keys.txt

echo ""
echo "2. 生成的密鑰："
cat /tmp/keys.txt

echo ""
echo "3. 提取公鑰："
F_PUBLIC_KEY=$(grep "F 的公鑰:" /tmp/keys.txt | cut -d' ' -f3)
N_PUBLIC_KEY=$(grep "N 的公鑰:" /tmp/keys.txt | cut -d' ' -f3)

echo "F 公鑰: $F_PUBLIC_KEY"
echo "N 公鑰: $N_PUBLIC_KEY"

echo ""
echo "4. 更新資料庫中的公鑰："
sudo -u postgres psql -d link << SQL
-- 更新 F 的公鑰
UPDATE users 
SET public_key = '$F_PUBLIC_KEY'
WHERE nickname = 'F';

-- 更新 N 的公鑰
UPDATE users 
SET public_key = '$N_PUBLIC_KEY'
WHERE nickname = 'N';

-- 檢查更新結果
SELECT nickname, 
       CASE WHEN public_key IS NULL THEN '❌ 空' 
            WHEN public_key = '' THEN '❌ 空字串'
            ELSE '✅ 已更新 (' || LENGTH(public_key) || ' 字元)' 
       END as key_status
FROM users 
WHERE nickname IN ('F', 'N', '小安');
SQL

echo ""
echo "5. 保存密鑰資訊："
cat /tmp/keys.txt > ~/link-keys-backup-$(date +%Y%m%d-%H%M%S).txt
echo "密鑰已備份到 ~/link-keys-backup-*.txt"

echo ""
echo "================================"
echo "✅ 公鑰已更新到資料庫！"
echo ""
echo "⚠️ 重要：請保存以下私鑰資訊"
echo "================================"
cat /tmp/keys.txt

EOF