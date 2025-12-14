#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Updating LINK on server...${NC}"

# Server details
SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

# Copy updated files to server
echo -e "${GREEN}Copying backend files...${NC}"
scp -r backend/internal/config/config.go ${SERVER_USER}@${SERVER_IP}:~/Link/backend/internal/config/
scp backend/cmd/server/main.go ${SERVER_USER}@${SERVER_IP}:~/Link/backend/cmd/server/
scp backend/internal/handler/admin.go ${SERVER_USER}@${SERVER_IP}:~/Link/backend/internal/handler/

# SSH to server and rebuild
echo -e "${GREEN}Rebuilding backend on server...${NC}"
ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
cd ~/Link/backend

# Update environment variables if not exist
if ! grep -q "ADMIN_PASSWORD" .env; then
    echo "ADMIN_PASSWORD=link-admin-2024-secure" >> .env
fi
if ! grep -q "BASE_URL" .env; then
    echo "BASE_URL=https://link.mcphub.tw" >> .env
fi

# Rebuild backend
go build -o bin/server ./cmd/server

# Restart service
sudo systemctl restart link-backend

echo "Backend updated and restarted"
systemctl status link-backend --no-pager | head -5
EOF

echo -e "${GREEN}✅ Server updated successfully!${NC}"