#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 檢查 admin JS 中的 API_BASE 設定 ==="
# 尋找包含 API_BASE 的部分
grep -o "VITE_API_URL[^,]*\|window.location.origin[^}]*" ~/Link/frontend/build/_app/immutable/nodes/3.*.js | head -5

echo ""
echo "=== 檢查完整的 API_BASE 邏輯 ==="
# 用更寬的 context 查看
grep -A2 -B2 "api/v1" ~/Link/frontend/build/_app/immutable/nodes/3.*.js | head -20

echo ""
echo "=== 檢查環境變數是否正確傳遞 ==="
grep -o "VITE_[^\"]*" ~/Link/frontend/build/_app/immutable/nodes/3.*.js | sort -u | head -10

echo ""
echo "=== 測試 CORS preflight ==="
curl -X OPTIONS https://link.mcphub.tw/api/v1/admin/cards \
  -H "Origin: https://link.mcphub.tw" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: X-Admin-Password,Content-Type" \
  -v 2>&1 | grep -E "< HTTP|< Access-Control"
EOF