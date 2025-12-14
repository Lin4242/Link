#!/bin/bash

echo "=== 為小安生成公私鑰對 ==="

# 使用 Node.js 和 tweetnacl 生成密鑰
cat > /tmp/generate-keys.js << 'EOF'
const nacl = require('tweetnacl');
const { encode: encodeBase64 } = require('@stablelib/base64');

const keyPair = nacl.box.keyPair();
console.log(JSON.stringify({
  publicKey: encodeBase64(keyPair.publicKey),
  secretKey: encodeBase64(keyPair.secretKey)
}));
EOF

# 在本地生成密鑰（使用前端的代碼）
echo "1. 生成密鑰對..."
cd frontend
npm install tweetnacl @stablelib/base64 2>/dev/null

# 創建簡單的密鑰生成腳本
cat > generate-keys.mjs << 'EOF'
import nacl from 'tweetnacl';
import { encode as encodeBase64 } from '@stablelib/base64';

const keyPair = nacl.box.keyPair();
console.log(JSON.stringify({
  publicKey: encodeBase64(keyPair.publicKey),
  secretKey: encodeBase64(keyPair.secretKey)
}));
EOF

keys=$(node generate-keys.mjs)
public_key=$(echo "$keys" | grep -o '"publicKey":"[^"]*' | cut -d'"' -f4)
secret_key=$(echo "$keys" | grep -o '"secretKey":"[^"]*' | cut -d'"' -f4)

echo "Public Key: $public_key"
echo "Secret Key: ${secret_key:0:20}..."

# 清理臨時文件
rm generate-keys.mjs

# 2. 更新資料庫中小安的公鑰
echo ""
echo "2. 更新小安的公鑰..."
ssh rocketmantw5516@34.136.217.56 << EOF
sudo -u postgres psql -d link << SQL
-- 更新小安的公鑰
UPDATE users 
SET public_key = '$public_key'
WHERE id = '5c86c416-070b-476e-953a-5ea3ac54163b';

-- 檢查結果
SELECT id, nickname, public_key, LENGTH(public_key) as key_length 
FROM users 
WHERE nickname IN ('F', '小安');
SQL
EOF

echo ""
echo "3. 儲存小安的密鑰資訊..."
cat > xiaoan-keys.json << EOF
{
  "nickname": "小安",
  "userId": "5c86c416-070b-476e-953a-5ea3ac54163b",
  "publicKey": "$public_key",
  "secretKey": "$secret_key",
  "primaryToken": "2ea2040ad0d9de47-1-4546b88c",
  "backupToken": "2ea2040ad0d9de47-2-04dec010",
  "password": "link-admin-2024-secure"
}
EOF

echo "✅ 小安的密鑰已生成並更新到資料庫"
echo "密鑰資訊已儲存到 xiaoan-keys.json（包含私鑰，請妥善保管）"