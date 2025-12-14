# LINK 部署成功總結

## 部署狀態：✅ 完全成功

### 線上環境
- **URL**: https://link.mcphub.tw
- **Server**: GCP VM (34.136.217.56)
- **SSL**: Let's Encrypt (自動更新)
- **域名**: link.mcphub.tw (A record 指向 VM)

## 已修復問題

### 1. 卡片儲存問題 ✅
- **問題**: Admin 新增的卡片只存在記憶體，重啟後消失
- **解決**: 修改 AdminHandler 將卡片存入 PostgreSQL
- **效期**: 從 30 分鐘延長至 30 天

### 2. CSS 兼容性 ✅
- **問題**: 舊版 iPhone SE 顯示白色背景
- **原因**: Tailwind opacity 修飾符不兼容
- **解決**: 改用 inline styles 與 rgba()

### 3. Admin 面板滾動 ✅
- **問題**: 無法上下滑動查看卡片
- **解決**: 移除全域 overflow:hidden

### 4. 註冊快取問題 ✅
- **問題**: localStorage 快取阻擋新卡註冊
- **解決**: 建立 /debug 頁面清理快取

### 5. 硬編碼 IP ✅
- **問題**: 192.168.1.99 硬編碼在前端
- **解決**: 全部改為相對路徑

## 功能驗證

### 核心功能
- ✅ NFC 雙卡註冊流程
- ✅ E2E 加密訊息傳送
- ✅ 訊息刪除（點擊自己的訊息）
- ✅ WebTransport/WebSocket 雙軌連線
- ✅ 即時打字指示器
- ✅ 隱私保護螢幕（切換分頁自動啟用）
- ✅ 線上/離線狀態顯示

### Admin 功能
- ✅ 密碼保護（環境變數配置）
- ✅ 生成 NFC 卡片配對
- ✅ 列出所有卡片（30 天內有效）
- ✅ 手機端正常滾動

### 安全性
- ✅ HTTPS (Let's Encrypt)
- ✅ E2E 加密（伺服器無法讀取訊息）
- ✅ 隱私螢幕保護
- ✅ XSS/CSRF 防護

## 部署流程

```bash
# 本地提交
git add .
git commit -m "commit message"
git push

# 伺服器部署
ssh rocketmantw5516@34.136.217.56
cd ~/Link
git pull
cd backend && /usr/local/go/bin/go build -o link-backend ./cmd/server
sudo systemctl restart link-backend
cd ../frontend && pnpm build
sudo systemctl restart link-frontend
```

## 系統架構

```
Nginx (443/80)
├── /api → Backend (8443)
├── /ws → WebSocket (8443)
└── /* → Frontend (3000)
```

## 測試帳號
- 主要: F (已註冊)
- 測試: 小安 (已建立)
- 密碼: 統一使用設定值

## 監控指令

```bash
# 查看服務狀態
sudo systemctl status link-backend
sudo systemctl status link-frontend
sudo systemctl status nginx

# 查看日誌
sudo journalctl -u link-backend -f
sudo journalctl -u link-frontend -f

# 資料庫連線
sudo -u postgres psql link_db
```

## 成功部署里程碑

1. **2024-12-14 05:50**: 初始部署到 GCP
2. **2024-12-14 06:00**: 設定 SSL 與域名
3. **2024-12-14 06:30**: 修復硬編碼 IP
4. **2024-12-14 07:00**: 修復卡片儲存問題
5. **2024-12-14 08:00**: 修復 UI/UX 問題
6. **2024-12-14 09:00**: 修復卡片過期問題
7. **2024-12-14 09:18**: 完成所有測試

## 注意事項

- 卡片效期 30 天，需定期清理過期資料
- Admin 密碼存於環境變數 ADMIN_PASSWORD
- 備份策略：每日備份 PostgreSQL
- SSL 憑證自動更新（Let's Encrypt）