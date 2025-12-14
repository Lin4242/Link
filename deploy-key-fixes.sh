#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Deploying secret key fixes to LINK production...${NC}"

# Server details
SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

# Step 1: Copy updated files to server
echo -e "${YELLOW}Step 1: Copying updated files...${NC}"

# Create a temporary directory for the files we need to copy
mkdir -p /tmp/link-deploy

# Copy the updated files to temp directory
cp -r frontend/src/routes/fix-keys /tmp/link-deploy/
cp frontend/src/routes/login/+page.svelte /tmp/link-deploy/login.svelte
cp frontend/src/routes/register/+page.svelte /tmp/link-deploy/register.svelte
cp frontend/src/lib/stores/messages.svelte.ts /tmp/link-deploy/messages.svelte.ts
cp frontend/src/routes/test-messages/+page.svelte /tmp/link-deploy/test-messages.svelte

# Transfer files to server
echo -e "${YELLOW}Transferring files to server...${NC}"
scp -r /tmp/link-deploy/* ${SERVER_USER}@${SERVER_IP}:/tmp/link-deploy/

# Step 2: Apply updates on server
echo -e "${YELLOW}Step 2: Applying updates on server...${NC}"
ssh ${SERVER_USER}@${SERVER_IP} << 'ENDSSH'
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Backing up current files...${NC}"
cd ~/Link/frontend
mkdir -p backup-$(date +%Y%m%d-%H%M%S)
cp -r src/routes/login backup-$(date +%Y%m%d-%H%M%S)/
cp -r src/routes/register backup-$(date +%Y%m%d-%H%M%S)/

echo -e "${YELLOW}Applying updates...${NC}"
# Copy fix-keys page (new)
cp -r /tmp/link-deploy/fix-keys src/routes/
# Update login page
cp /tmp/link-deploy/login.svelte src/routes/login/+page.svelte
# Update register page
cp /tmp/link-deploy/register.svelte src/routes/register/+page.svelte
# Update messages store
cp /tmp/link-deploy/messages.svelte.ts src/lib/stores/messages.svelte.ts
# Update test-messages page
cp /tmp/link-deploy/test-messages.svelte src/routes/test-messages/+page.svelte

echo -e "${YELLOW}Rebuilding frontend...${NC}"
export PATH=$PATH:$HOME/.nvm/versions/node/v22.12.0/bin
pnpm build

echo -e "${GREEN}✅ Frontend rebuilt successfully!${NC}"

# Clean up temp files
rm -rf /tmp/link-deploy

echo -e "${GREEN}Deployment complete on server!${NC}"
ENDSSH

# Clean up local temp files
rm -rf /tmp/link-deploy

echo -e "${GREEN}✅ Deployment complete!${NC}"
echo ""
echo -e "${YELLOW}Important next steps:${NC}"
echo "1. Visit https://link.mcphub.tw/fix-keys with F and N accounts to repair their keys"
echo "2. Enter the login password when prompted"
echo "3. Once keys are repaired, messages should become visible"
echo "4. Test messaging between F and N users"
echo ""
echo -e "${YELLOW}Debug pages available:${NC}"
echo "- https://link.mcphub.tw/test-messages - Check message system status"
echo "- https://link.mcphub.tw/fix-keys - Repair missing encryption keys"