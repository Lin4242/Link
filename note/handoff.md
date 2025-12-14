# Claude 交接文件

> 每次重啟 Claude 後閱讀此文件快速恢復上下文
> **敏感資訊 (密碼/token)**: 見 `local_note.md` (gitignored)

---

## 專案狀態 (2024-12-15)

### 當前狀態
- **線上環境**: 已部署，前端有更新
- **Seed 資料**: 小安 + 阿詠 + Demo 卡片
- **自動加好友**: 新用戶註冊後自動與小安成為好友
- **Playwright MCP**: 可用於自動化測試

### 已完成
- [x] Seed script (`backend/cmd/seed/main.go`)
- [x] 小安服務帳號 (可登入)
- [x] 阿詠測試帳號 (可登入)
- [x] Demo 卡片 (只有 card_pairs，未註冊，可走完整 demo)
- [x] Auto-friend 功能 (SERVICE_USER_ID)
- [x] NFC 跨 tab cookie 修復 (iPhone Safari)
- [x] Admin 面板 (產生卡片 token)
- [x] 好友列表 UI (沒對話時顯示好友，可點擊開始聊天)
- [x] 解鎖邏輯簡化 (placeholder 自動更新，不匹配視為密碼錯誤)

### 待修 Bug (優先順序)
1. [ ] **即時推送失效** - 收訊方需 refresh 才看到新訊息
2. [ ] **公鑰 placeholder 問題** - seed 應推導真正公鑰
3. [ ] **環境檔分離** - 需要 .env.development / .env.production

### 待測試
- [ ] 完整 NFC 註冊流程 (刷主卡 → 刷副卡 → 註冊)
- [x] 小安電腦端登入 ✓
- [ ] 聊天功能 (E2EE 加解密) - 發送成功，接收有問題
- [x] 好友系統 ✓

---

## 測試流程

### 1. 小安登入測試 (電腦端)
```
1. 開啟 小安 Login URL (見 local_note.md)
2. 輸入密碼登入
3. 確認進入聊天頁面
```

### 2. Demo 註冊流程測試 (NFC)
```
1. 刷主卡 → 應進入註冊頁面
2. 刷副卡 → 配對成功
3. 填寫暱稱和密碼
4. 完成註冊後確認好友列表有小安
```

### 3. 聊天測試
```
1. 小安和新用戶各開一個瀏覽器
2. 互傳訊息
3. 確認雙方都能看到且能解密
```

---

## 重要路徑

| 用途 | 路徑 |
|------|------|
| 後端入口 | `backend/cmd/server/main.go` |
| Seed Script | `backend/cmd/seed/main.go` |
| Auth Service | `backend/internal/service/auth.go` |
| 前端聊天頁 | `frontend/src/routes/chat/+page.svelte` |
| 註冊頁 | `frontend/src/routes/register/+page.svelte` |
| 金鑰推導 | `frontend/src/lib/crypto/keys.ts` |
| WebSocket | `backend/internal/transport/` |
| 設定檔 | `backend/internal/config/config.go` |

---

## API 端點

### Public
| Method | Path | 用途 |
|--------|------|------|
| GET | `/health` | 健康檢查 |
| GET | `/w/:token` | NFC 卡片入口 |
| GET | `/api/v1/auth/check-card/:token` | 檢查卡片狀態 |
| POST | `/api/v1/auth/register` | 註冊 |
| POST | `/api/v1/auth/login` | 登入 (主卡) |
| POST | `/api/v1/auth/login/backup` | 登入 (副卡撤銷) |

### Auth Required
| Method | Path | 用途 |
|--------|------|------|
| GET | `/api/v1/users/me` | 取得自己資料 |
| PATCH | `/api/v1/users/me` | 更新資料 |
| GET | `/api/v1/friends` | 好友列表 |
| POST | `/api/v1/friends/request` | 發送好友請求 |
| GET | `/api/v1/conversations` | 對話列表 |
| GET | `/ws` | WebSocket 連線 |

### Admin
| Method | Path | 用途 |
|--------|------|------|
| POST | `/api/v1/admin/cards/generate` | 產生卡片 token |
| GET | `/api/v1/admin/cards` | 列出所有卡片 |
| DELETE | `/api/v1/admin/cards/:id` | 刪除卡片 |

---

## 伺服器資訊

| 項目 | 值 |
|------|---|
| SSH | `ssh jimmy@link.mcphub.tw` |
| App 路徑 | `/home/rocketmantw5516/Link/` |
| Run as | `rocketmantw5516` |
| 後端服務 | `sudo systemctl restart link-backend` |
| 資料庫 | PostgreSQL, user `postgres`, db `link` |
| 網址 | https://link.mcphub.tw |

---

## 常用指令

```bash
# 重啟後端
ssh jimmy@link.mcphub.tw 'sudo systemctl restart link-backend'

# 查看 logs
ssh jimmy@link.mcphub.tw 'sudo journalctl -u link-backend -f'

# Pull + Restart
ssh jimmy@link.mcphub.tw 'cd /home/rocketmantw5516/Link && sudo -u rocketmantw5516 git pull && cd backend && sudo -u rocketmantw5516 /usr/local/go/bin/go build -o bin/server ./cmd/server && sudo systemctl restart link-backend'

# 執行 seed
ssh jimmy@link.mcphub.tw 'cd /home/rocketmantw5516/Link/backend && sudo -u rocketmantw5516 bash -c "set -a && source .env.seed && set +a && /usr/local/go/bin/go run ./cmd/seed"'

# 資料庫查詢
ssh jimmy@link.mcphub.tw 'sudo -u postgres psql -d link -c "SELECT id, nickname FROM users;"'
```

---

## 已知地雷

| 問題 | 原因 | 解法 |
|------|------|------|
| SSH 權限 | jimmy 不是 app owner | `sudo -u rocketmantw5516` |
| Go 找不到 | PATH 問題 | `/usr/local/go/bin/go` |
| 環境變數失效 | sudo 不帶環境 | `bash -c "set -a && source .env && ..."` |
| iPhone NFC 跨 tab | Safari localStorage 隔離 | 已改用 cookie |
| 無痕模式失敗 | cookie 也被隔離 | 預期行為，告知用戶 |

---

## 架構筆記

### Auto-Friend 功能
- `SERVICE_USER_ID` 環境變數設定小安的 user ID
- `backend/internal/service/auth.go` Register() 自動建立友誼
- 新用戶註冊完成後立即與小安成為好友

### NFC 卡片流程
```
掃主卡 → /w/{token} → check-card API
  → 未註冊: redirect /register?token=xxx
  → 已註冊: redirect /login?token=xxx

註冊頁:
  掃第一張 → 存入 cookie
  掃第二張 → 驗證配對 → 完成表單 → 註冊
```

### E2EE 金鑰推導
```
password + userId → PBKDF2 → nacl.box.keyPair
同一組密碼在任何裝置都會產生相同 keypair
```

---

## Seed Script 說明

Seed 建立兩種資料:

1. **小安 (Service User)** - 完整建立
   - user + cards 都有
   - 可以用主卡 URL 登入
   - 作為所有新用戶的自動好友

2. **Demo Card Pair** - 只建立 card_pairs
   - 不建立 user 和 cards
   - 可以走完整的 NFC 註冊 demo 流程

---

## 瀏覽器測試

重啟後可用 WebFetch 或請用戶用瀏覽器測試:
- https://link.mcphub.tw/health - 健康檢查
- https://link.mcphub.tw/admin - Admin 面板
- 小安登入 URL - 見 local_note.md
