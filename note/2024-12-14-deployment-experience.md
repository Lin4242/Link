# 2024-12-14 部署經驗記錄

## 部署環境
- **平台**: GCP VM (Compute Engine)
- **IP**: 34.136.217.56
- **域名**: link.mcphub.tw
- **使用者**: rocketmantw5516
- **OS**: Debian GNU/Linux

## 部署流程總結

### Phase 1: 初始部署 (使用 deploy-vm.sh)
1. 安裝系統依賴 (Go 1.23, Node.js 22, PostgreSQL, Nginx)
2. 設定 PostgreSQL 資料庫
3. Clone GitHub repository
4. 執行 database migration
5. 建置後端和前端
6. 設定 systemd service
7. 配置 Nginx 反向代理

### Phase 2: 問題處理
**問題 1: TLS 證書缺失**
- 錯誤: `cannot load TLS key pair: open ../certs/localhost+2.pem`
- 解決: 生成自簽證書
```bash
cd ~/Link/backend
mkdir -p certs
openssl req -x509 -newkey rsa:4096 -keyout certs/server-key.pem -out certs/server.pem -days 365 -nodes
```

**問題 2: 前端建置失敗**
- 錯誤: pnpm 權限問題
- 解決: 使用 sudo 安裝全域套件

**問題 3: 網站無法訪問**
- 原因: GCP 防火牆未開放 80/443 port
- 解決: 使用者在 GCP console 設定防火牆規則

### Phase 3: HTTPS 設定
使用 Let's Encrypt 取得 SSL 證書：
```bash
sudo certbot --nginx -d link.mcphub.tw
```

### Phase 4: Admin Panel 設定
**初始問題**：
- Admin 密碼硬編碼為 "42424242"
- Base URL 硬編碼為 "https://192.168.1.99:9443"

**解決方案**：
1. 修改後端支援環境變數：
   - 新增 `ADMIN_PASSWORD` 和 `BASE_URL` 到 config
   - 更新 admin handler 使用 config 值

2. 修改前端 admin panel：
   - 移除硬編碼 API URL
   - 使用環境變數或 window.location.origin

### Phase 5: 建立正確的部署流程
**錯誤做法**：
- 直接用 scp 複製檔案到 server
- 在 server 上直接修改程式碼
- 造成 Git 狀態不同步

**正確流程**：
1. 本地開發 → Git commit → Push to GitHub
2. Server 執行 `deploy-from-github.sh`
3. 自動 pull、build、restart

## 關鍵檔案和設定

### 環境變數 (backend/.env)
```bash
SERVER_ADDR=:8443
SERVER_ENV=production
DATABASE_URL=postgres://link:LinkSecurePassword2024@localhost:5432/link?sslmode=disable
JWT_SECRET=[32+ chars]
JWT_EXPIRY=24h
CORS_ORIGINS=https://link.mcphub.tw,http://link.mcphub.tw
LOG_LEVEL=info
ADMIN_PASSWORD=link-admin-2024-secure
BASE_URL=https://link.mcphub.tw
TLS_CERT_FILE=../certs/server.pem
TLS_KEY_FILE=../certs/server-key.pem
```

### Nginx 設定 (/etc/nginx/sites-available/link)
- 前端靜態檔案: `/home/rocketmantw5516/Link/frontend/build`
- API 代理: `proxy_pass http://localhost:8443`
- WebSocket: 升級連線支援
- Let's Encrypt SSL 自動配置

### Systemd Service (/etc/systemd/system/link-backend.service)
- 自動重啟
- 環境變數從 .env 載入
- 依賴 PostgreSQL

## 技術決策

1. **使用 VM 而非 Cloud Run**
   - 更簡單的部署流程
   - 完整控制環境
   - 適合 WebSocket/WebTransport

2. **Nginx 作為反向代理**
   - 處理 SSL termination
   - 靜態檔案服務
   - WebSocket 代理

3. **Git-based 部署**
   - 保持程式碼同步
   - 可追蹤的變更歷史
   - 避免手動修改產生的問題

## 學到的教訓

1. **永遠不要直接在 server 修改程式碼**
   - 使用 Git 作為單一事實來源
   - 所有變更都要經過 commit

2. **密碼管理**
   - 絕不 commit 密碼到 Git
   - 使用 .env 檔案
   - 提供 .env.example

3. **部署腳本化**
   - 減少人為錯誤
   - 確保一致性
   - 容易回滾

4. **測試 Admin 功能**
   - 先用 curl 測試 API
   - 確認環境變數正確載入
   - 驗證 URL 格式

## 部署檢查清單

- [x] 資料庫已建立並執行 migration
- [x] 後端服務正常運行
- [x] 前端已建置並部署
- [x] Nginx 正確配置
- [x] SSL 證書已安裝
- [x] 防火牆規則已設定
- [x] Admin panel 可正常訪問
- [x] WebSocket 連線正常
- [x] 環境變數已正確設定
- [x] Git repository 同步

## 常用指令

```bash
# 查看服務狀態
sudo systemctl status link-backend

# 查看日誌
sudo journalctl -u link-backend -f

# 重啟服務
sudo systemctl restart link-backend

# 部署更新
./deploy-from-github.sh

# 修復 Git 衝突
./fix-server-conflicts.sh

# 測試 API
curl https://link.mcphub.tw/api/v1/health
```