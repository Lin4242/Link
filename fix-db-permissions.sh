#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
echo "=== 修復資料庫權限問題 ==="

# 切換到 postgres 用戶並執行 SQL
sudo -u postgres psql << 'SQL'
-- 連接到 link 資料庫
\c link

-- 授予 link 用戶所有權限
GRANT ALL PRIVILEGES ON DATABASE link TO link;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO link;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO link;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO link;

-- 設置默認權限給未來的表
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO link;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO link;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO link;

-- 確保 schema 權限
GRANT ALL ON SCHEMA public TO link;

-- 列出當前權限
\dt
\dp
SQL

echo ""
echo "=== 重啟後端服務 ==="
sudo systemctl restart link-backend

echo ""
echo "=== 檢查服務狀態 ==="
sudo systemctl status link-backend --no-pager

echo ""
echo "=== 測試資料庫連線 ==="
cd ~/Link/backend
export DATABASE_URL="postgres://link:LinkSecurePassword2024@localhost:5432/link?sslmode=disable"
echo "SELECT current_user, current_database();" | psql $DATABASE_URL
EOF