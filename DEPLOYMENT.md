# LINK 部署指南

## 安全注意事項 ⚠️

**重要：絕對不要將密碼或敏感資訊 commit 到 Git！**

## 部署流程

### 1. 環境準備

在 server 上設置環境變數：

```bash
cd ~/Link/backend
cp .env.production.example .env
nano .env  # 編輯並填入實際密碼
```

### 2. 部署更新

從 GitHub 拉取最新代碼並重新部署：

```bash
./deploy-from-github.sh
```

### 3. 密碼管理

- **Admin 密碼**：存儲在 `backend/.env` 的 `ADMIN_PASSWORD`
- **資料庫密碼**：存儲在 `DATABASE_URL` 中
- **JWT Secret**：至少 32 字元，存儲在 `JWT_SECRET`

這些密碼都只存在於 server 的 `.env` 檔案中，不會被 commit。

### 4. 訪問 Admin Panel

1. 打開 https://link.mcphub.tw/admin
2. 輸入 Admin 密碼
3. 生成卡片對
4. 複製 URL 來燒錄到 NFC 卡片

### 5. 更新流程

正確的更新流程：

1. **本地開發**
   ```bash
   cd ~/Link
   # 修改代碼
   git add .
   git commit -m "描述"
   git push origin main
   ```

2. **部署到 Server**
   ```bash
   ./deploy-from-github.sh
   ```

### 6. 檔案結構

```
Link/
├── backend/
│   ├── .env                 # ⚠️ 生產環境變數 (不要 commit)
│   └── .env.production.example  # 範例檔案
├── deploy-from-github.sh    # 部署腳本
└── fix-server-conflicts.sh  # 修復衝突腳本
```

### 7. 故障排除

如果 server 上有檔案衝突：
```bash
./fix-server-conflicts.sh
```

查看服務狀態：
```bash
ssh rocketmantw5516@34.136.217.56
sudo systemctl status link-backend
sudo journalctl -u link-backend -f
```

## 安全檢查清單

- [ ] `.env` 檔案沒有被 commit
- [ ] Admin 密碼足夠複雜
- [ ] JWT Secret 至少 32 字元
- [ ] HTTPS 已啟用
- [ ] 防火牆規則已設定