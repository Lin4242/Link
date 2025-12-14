# Claude 交接文件

> 每次重啟 Claude 後閱讀此文件快速恢復上下文

---

## 專案狀態 (2024-12-15)

### 當前狀態
- **線上資料庫**: 已清空，無用戶資料
- **需要**: 用戶提供 NFC 卡片 token 才能用 seed script 建立 demo 資料
- **SPEC 文件**: `LINK-SPEC-v4.3.md` (完整規格)

### 已完成
- [x] Seed script 建立 (`backend/cmd/seed/main.go`)
- [x] 移除硬編碼密碼，強制環境變數
- [x] GCP 伺服器部署完成
- [x] Playwright MCP 已安裝 (瀏覽器自動化)
- [x] 副卡登入頁面更新為深色主題

### 待完成
- [ ] 取得 NFC 卡片 token 填入 seed script
- [ ] 完整 demo 流程測試
- [ ] NTAG424 DNA 零知識驗證（等硬體到貨）

---

## 重要路徑

| 用途 | 路徑 |
|------|------|
| 後端入口 | `backend/cmd/server/main.go` |
| Seed Script | `backend/cmd/seed/main.go` |
| 前端聊天頁 | `frontend/src/routes/chat/+page.svelte` |
| 金鑰推導 | `frontend/src/lib/crypto/keys.ts` |
| WebSocket | `backend/internal/transport/` |
| 設定檔 | `backend/internal/config/config.go` |

---

## 伺服器資訊

| 項目 | 值 |
|------|---|
| SSH | `ssh jimmy@link.mcphub.tw` |
| App 路徑 | `/home/rocketmantw5516/Link/` |
| 後端服務 | `sudo systemctl restart link-backend` |
| 資料庫 | `sudo -u postgres psql -d link` |
| 網址 | https://link.mcphub.tw |

---

## Seed Script 使用

```bash
cd backend

# 設定 NFC 卡片 token (需從實體卡片取得)
export DATABASE_URL=postgres://...
export SEED_DEMO_PRIMARY_TOKEN=主卡token
export SEED_DEMO_BACKUP_TOKEN=附卡token
export SEED_DEMO_NICKNAME=小安
export SEED_DEMO_PASSWORD=demo1234

go run ./cmd/seed
```

---

## Playwright MCP 使用

重啟 Claude 後可用瀏覽器自動化：

```
> "打開 https://link.mcphub.tw 並截圖"
> "填寫登入表單並提交"
> "驗證聊天功能是否正常"
```

---

## 已知地雷

1. **SSH 權限**: 用 `sudo -u rocketmantw5516` 執行 git/go 指令
2. **Go 路徑**: 伺服器上用 `/usr/local/go/bin/go`
3. **環境變數**: 必須設定 `JWT_SECRET`, `CARD_TOKEN_SECRET`, `ADMIN_PASSWORD`
4. **密碼 Hash**: Argon2id 格式 `$argon2id$v=19$m=65536,t=3,p=N$...`

---

## 開發指令

```bash
# 本機後端
cd backend && make dev

# 本機前端
cd frontend && pnpm dev

# 測試
cd backend && go test ./... -v
cd frontend && pnpm test
```

---

## Admin 面板

| 項目 | 值 |
|------|---|
| 網址 | https://link.mcphub.tw/admin |
| 密碼 | 環境變數 `ADMIN_PASSWORD` (不在 git 裡) |
| 用途 | 產生 NFC 卡片 token pair |

**流程**: Admin 產生 card pair → 燒錄到 NFC 卡片 → 用戶掃卡註冊/登入

---

## NFC 卡片流程

```
1. Admin 產生 card pair (primary_token + backup_token)
2. 燒錄 token 到實體 NFC 卡片 (寫入 NDEF URL)
3. 用戶掃主卡 → 導向 /w/{token} → 檢查狀態
   - 未註冊 → /register?token=xxx
   - 已註冊 → /login?token=xxx
4. 掃附卡 → /login/backup → 撤銷主卡警告
```

**卡片 URL 格式**: `https://link.mcphub.tw/w/{token}`

---

## 本地開發設定

```bash
# TLS 憑證位置 (mkcert 產生)
certs/localhost+2.pem
certs/localhost+2-key.pem

# 後端 .env 範例
JWT_SECRET=至少32字元的隨機字串
CARD_TOKEN_SECRET=隨機字串
ADMIN_PASSWORD=你的admin密碼
DATABASE_URL=postgres://postgres:postgres@localhost:5432/link?sslmode=disable
```

---

## 部署流程

```bash
# SSH 進入伺服器
ssh jimmy@link.mcphub.tw

# 後端部署
cd /home/rocketmantw5516/Link
sudo -u rocketmantw5516 git pull
cd backend
sudo -u rocketmantw5516 /usr/local/go/bin/go build -o bin/server ./cmd/server
sudo systemctl restart link-backend

# 前端部署
cd /home/rocketmantw5516/Link/frontend
sudo -u rocketmantw5516 pnpm install
sudo -u rocketmantw5516 pnpm build
# 靜態檔案由 nginx 提供
```

---

## 清空資料庫 (重置)

```bash
ssh jimmy@link.mcphub.tw
sudo -u postgres psql -d link -c "
TRUNCATE TABLE messages, conversations, sessions, cards, card_pairs, friendships, users CASCADE;
"
```
