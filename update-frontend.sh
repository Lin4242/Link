#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}Updating LINK frontend...${NC}"

# Server details
SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

# Copy updated frontend files
echo -e "${GREEN}Copying frontend files...${NC}"
scp frontend/src/routes/admin/+page.svelte ${SERVER_USER}@${SERVER_IP}:~/Link/frontend/src/routes/admin/

# SSH to server and rebuild
echo -e "${GREEN}Rebuilding frontend on server...${NC}"
ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
export PATH=$PATH:$HOME/.nvm/versions/node/v22.12.0/bin
cd ~/Link/frontend

# Rebuild frontend
pnpm build

echo "Frontend rebuilt successfully"
EOF

echo -e "${GREEN}✅ Frontend updated successfully!${NC}"
echo "Visit https://link.mcphub.tw/admin to access the admin panel"
echo "Password: link-admin-2024-secure"