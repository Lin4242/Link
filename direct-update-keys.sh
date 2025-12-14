#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

echo "直接更新 F 和 N 的公鑰..."
echo "================================"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd ~/Link/backend

echo "1. 創建 Go 程式來生成密鑰對..."
cat > /tmp/genkeys.go << 'GOEOF'
package main

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "golang.org/x/crypto/nacl/box"
)

func main() {
    // 生成 F 的密鑰對
    pubF, privF, _ := box.GenerateKey(rand.Reader)
    fmt.Printf("F_PUBLIC_KEY=%s\n", base64.StdEncoding.EncodeToString(pubF[:]))
    fmt.Printf("F_PRIVATE_KEY=%s\n", base64.StdEncoding.EncodeToString(privF[:]))
    
    // 生成 N 的密鑰對
    pubN, privN, _ := box.GenerateKey(rand.Reader)
    fmt.Printf("N_PUBLIC_KEY=%s\n", base64.StdEncoding.EncodeToString(pubN[:]))
    fmt.Printf("N_PRIVATE_KEY=%s\n", base64.StdEncoding.EncodeToString(privN[:]))
}
GOEOF

echo "2. 執行生成密鑰..."
go run /tmp/genkeys.go > /tmp/keys.env

echo "3. 顯示生成的密鑰："
cat /tmp/keys.env

echo ""
echo "4. 載入密鑰變數："
source /tmp/keys.env

echo "5. 更新資料庫："
source .env

PGPASSWORD=$DB_PASSWORD psql -h localhost -U $DB_USER -d $DB_NAME << SQL
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
       CASE WHEN public_key IS NULL THEN '❌ NULL'
            WHEN public_key = '' THEN '❌ 空字串' 
            ELSE '✅ 已更新 (' || LENGTH(public_key) || ' 字元)'
       END as key_status,
       LEFT(public_key, 20) || '...' as key_preview
FROM users
WHERE nickname IN ('F', 'N', '小安')
ORDER BY nickname;
SQL

echo ""
echo "6. 備份密鑰資訊："
cp /tmp/keys.env ~/keys-backup-$(date +%Y%m%d-%H%M%S).env
echo "密鑰已備份到 ~/keys-backup-*.env"

echo ""
echo "================================"
echo "✅ 完成！F 和 N 的公鑰已更新"
echo ""
echo "⚠️ 重要提醒："
echo "1. F 和 N 需要在瀏覽器中清除所有 LINK 相關的資料"
echo "2. 重新登入後系統會使用新的密鑰"
echo "3. 之前的訊息將無法解密（因為使用了不同的密鑰）"
echo "================================"

EOF