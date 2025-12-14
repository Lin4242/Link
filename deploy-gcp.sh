#!/bin/bash
set -e

# ====================================
# LINK GCP 部署腳本
# ====================================

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 預設值
PROJECT_ID=${1:-""}
REGION=${2:-"asia-east1"}
DB_PASSWORD=""

echo -e "${GREEN}=====================================${NC}"
echo -e "${GREEN}     LINK GCP 部署腳本${NC}"
echo -e "${GREEN}=====================================${NC}"
echo ""

# 檢查 gcloud CLI
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}錯誤: 未安裝 gcloud CLI${NC}"
    echo "請先安裝: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# 設定專案
if [ -z "$PROJECT_ID" ]; then
    echo -e "${YELLOW}請輸入 GCP Project ID:${NC}"
    read -r PROJECT_ID
fi

echo -e "${GREEN}使用專案: $PROJECT_ID${NC}"
gcloud config set project $PROJECT_ID

# 啟用必要的 API
echo -e "${GREEN}啟用 GCP APIs...${NC}"
gcloud services enable \
    cloudbuild.googleapis.com \
    run.googleapis.com \
    sqladmin.googleapis.com \
    secretmanager.googleapis.com \
    firebasehosting.googleapis.com \
    containerregistry.googleapis.com

# ====================================
# 1. 設定 Cloud SQL
# ====================================
echo -e "${GREEN}=====================================${NC}"
echo -e "${GREEN}1. 設定 Cloud SQL PostgreSQL${NC}"
echo -e "${GREEN}=====================================${NC}"

DB_INSTANCE="link-db"
DB_EXISTS=$(gcloud sql instances list --filter="name:$DB_INSTANCE" --format="value(name)" 2>/dev/null || echo "")

if [ -z "$DB_EXISTS" ]; then
    echo "建立 Cloud SQL 實例..."
    gcloud sql instances create $DB_INSTANCE \
        --database-version=POSTGRES_15 \
        --tier=db-f1-micro \
        --region=$REGION \
        --network=default \
        --database-flags=max_connections=100
    
    # 設定密碼
    echo -e "${YELLOW}請設定資料庫密碼:${NC}"
    read -s DB_PASSWORD
    echo ""
    
    gcloud sql users set-password postgres \
        --instance=$DB_INSTANCE \
        --password="$DB_PASSWORD"
    
    # 建立資料庫
    gcloud sql databases create link \
        --instance=$DB_INSTANCE
else
    echo "Cloud SQL 實例已存在"
    echo -e "${YELLOW}請輸入現有的資料庫密碼:${NC}"
    read -s DB_PASSWORD
    echo ""
fi

# 取得連線資訊
CONNECTION_NAME=$(gcloud sql instances describe $DB_INSTANCE --format="value(connectionName)")
echo "資料庫連線名稱: $CONNECTION_NAME"

# ====================================
# 2. 設定 Secret Manager
# ====================================
echo -e "${GREEN}=====================================${NC}"
echo -e "${GREEN}2. 設定 Secret Manager${NC}"
echo -e "${GREEN}=====================================${NC}"

# 產生 JWT Secret
JWT_SECRET=$(openssl rand -hex 32)

# 建立 Secrets
echo "建立 Secrets..."

# Database URL
DATABASE_URL="postgres://postgres:$DB_PASSWORD@localhost/link?host=/cloudsql/$CONNECTION_NAME"
echo -n "$DATABASE_URL" | gcloud secrets create database-url --data-file=- 2>/dev/null || \
    echo -n "$DATABASE_URL" | gcloud secrets versions add database-url --data-file=-

# JWT Secret
echo -n "$JWT_SECRET" | gcloud secrets create jwt-secret --data-file=- 2>/dev/null || \
    echo -n "$JWT_SECRET" | gcloud secrets versions add jwt-secret --data-file=-

# ====================================
# 3. 執行資料庫 Migration
# ====================================
echo -e "${GREEN}=====================================${NC}"
echo -e "${GREEN}3. 執行資料庫 Migration${NC}"
echo -e "${GREEN}=====================================${NC}"

# 使用 Cloud SQL Proxy 連線
echo "安裝 Cloud SQL Proxy..."
if [ ! -f "./cloud_sql_proxy" ]; then
    curl -o cloud_sql_proxy https://dl.google.com/cloudsql/cloud_sql_proxy.darwin.amd64
    chmod +x cloud_sql_proxy
fi

# 啟動 proxy
./cloud_sql_proxy -instances=$CONNECTION_NAME=tcp:5433 &
PROXY_PID=$!
sleep 3

# 執行 migration
echo "執行 Migration..."
PGPASSWORD=$DB_PASSWORD psql -h localhost -p 5433 -U postgres -d link -f backend/migrations/001_init.up.sql

# 停止 proxy
kill $PROXY_PID

# ====================================
# 4. 部署後端到 Cloud Run
# ====================================
echo -e "${GREEN}=====================================${NC}"
echo -e "${GREEN}4. 部署後端到 Cloud Run${NC}"
echo -e "${GREEN}=====================================${NC}"

# Build Docker image
echo "建立 Docker image..."
docker build -t gcr.io/$PROJECT_ID/link-backend:latest -f Dockerfile.backend .

# Push to GCR
echo "推送到 Container Registry..."
docker push gcr.io/$PROJECT_ID/link-backend:latest

# 部署到 Cloud Run
echo "部署到 Cloud Run..."
gcloud run deploy link-backend \
    --image gcr.io/$PROJECT_ID/link-backend:latest \
    --region $REGION \
    --platform managed \
    --allow-unauthenticated \
    --port 8443 \
    --min-instances 1 \
    --max-instances 10 \
    --memory 512Mi \
    --cpu 1 \
    --set-env-vars="SERVER_ENV=production,SERVER_ADDR=:8443" \
    --set-secrets="DATABASE_URL=database-url:latest,JWT_SECRET=jwt-secret:latest" \
    --add-cloudsql-instances=$CONNECTION_NAME

# 取得後端 URL
BACKEND_URL=$(gcloud run services describe link-backend --region=$REGION --format='value(status.url)')
echo -e "${GREEN}後端 URL: $BACKEND_URL${NC}"

# ====================================
# 5. 部署前端到 Firebase Hosting
# ====================================
echo -e "${GREEN}=====================================${NC}"
echo -e "${GREEN}5. 部署前端到 Firebase Hosting${NC}"
echo -e "${GREEN}=====================================${NC}"

# 初始化 Firebase (如果需要)
if [ ! -f ".firebaserc" ]; then
    echo "初始化 Firebase..."
    cat > .firebaserc <<EOF
{
  "projects": {
    "default": "$PROJECT_ID"
  }
}
EOF
fi

# Build 前端
echo "建立前端..."
cd frontend
npm install -g pnpm
pnpm install
VITE_API_URL=$BACKEND_URL VITE_WS_URL=${BACKEND_URL/https/wss}/ws pnpm build
mv build ../frontend-dist
cd ..

# 部署到 Firebase
echo "部署到 Firebase Hosting..."
npm install -g firebase-tools
firebase deploy --only hosting --project $PROJECT_ID

# ====================================
# 6. 設定 Cloud Build 觸發器 (可選)
# ====================================
echo -e "${GREEN}=====================================${NC}"
echo -e "${GREEN}6. 設定 CI/CD (可選)${NC}"
echo -e "${GREEN}=====================================${NC}"

echo -e "${YELLOW}要設定自動部署嗎? (y/n)${NC}"
read -r SETUP_CICD

if [ "$SETUP_CICD" = "y" ]; then
    echo "請連結 GitHub repository..."
    echo "1. 前往: https://console.cloud.google.com/cloud-build/triggers"
    echo "2. 點擊 'Connect Repository'"
    echo "3. 選擇 GitHub 並授權"
    echo "4. 選擇 repository: Lin4242/Link"
    echo ""
    echo "建立觸發器..."
    gcloud builds triggers create github \
        --repo-name=Link \
        --repo-owner=Lin4242 \
        --branch-pattern="^main$" \
        --build-config=cloudbuild.yaml \
        --substitutions="_PROJECT_ID=$PROJECT_ID,_REGION=$REGION"
fi

# ====================================
# 完成
# ====================================
echo ""
echo -e "${GREEN}=====================================${NC}"
echo -e "${GREEN}     🎉 部署完成！${NC}"
echo -e "${GREEN}=====================================${NC}"
echo ""
echo -e "${GREEN}後端 API: $BACKEND_URL${NC}"
echo -e "${GREEN}前端網址: https://$PROJECT_ID.web.app${NC}"
echo ""
echo "測試連線:"
echo "  curl $BACKEND_URL/health"
echo ""
echo -e "${YELLOW}注意事項:${NC}"
echo "1. 請確認 Cloud Run 服務已正常運行"
echo "2. 請確認 Firebase Hosting 已部署成功"
echo "3. 第一次載入可能較慢 (Cold Start)"
echo ""