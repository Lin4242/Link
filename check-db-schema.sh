#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== card_pairs 表結構 ==="
sudo -u postgres psql -d link -c "\d card_pairs"

echo ""
echo "=== 查看 backend 資料庫 migration ==="
find ~/Link/backend -name "*.sql" -o -name "*migration*" 2>/dev/null | head -10

echo ""
echo "=== 查看 repository 介面 ==="
grep -n "CreateCardPair\|SaveCardPair" ~/Link/backend/internal/repository/*.go 2>/dev/null || echo "No CardPair methods found"
EOF