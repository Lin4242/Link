#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 檢查 admin.html 內容 ==="
cat ~/Link/frontend/build/admin.html

echo ""
echo "=== 檢查是否有 admin 相關的 JS 檔案 ==="
find ~/Link/frontend/build/_app -name "*.js" -exec grep -l "X-Admin-Password\|admin/cards" {} \; 2>/dev/null | head -5

echo ""
echo "=== 檢查 Node 3 (admin page) 內容片段 ==="
grep -o "API_BASE\|admin/cards\|X-Admin-Password" ~/Link/frontend/build/_app/immutable/nodes/3.*.js 2>/dev/null | head -10

echo ""
echo "=== 測試手機瀏覽器 User-Agent ==="
echo "User-Agent: Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X)"
echo ""
echo "測試從手機 User-Agent 訪問..."
curl -s https://link.mcphub.tw/admin \
  -H "User-Agent: Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X)" | grep -E "<script|<link" | head -10
EOF