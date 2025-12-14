#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Deploying LINK from GitHub...${NC}"

# Server details
SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

# SSH to server and pull from GitHub
ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin:$HOME/.nvm/versions/node/v22.12.0/bin

echo "Pulling latest changes from GitHub..."
cd ~/Link
git pull origin main

echo "Rebuilding backend..."
cd ~/Link/backend
go build -o bin/server ./cmd/server

echo "Restarting backend service..."
sudo systemctl restart link-backend

echo "Rebuilding frontend..."
cd ~/Link/frontend
pnpm install
pnpm build

echo "Deployment complete!"
systemctl status link-backend --no-pager | head -5
EOF

echo -e "${GREEN}✅ Deployment from GitHub completed!${NC}"