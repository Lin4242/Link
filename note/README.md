# LINK 開發筆記精華

> 本文件濃縮開發過程中的重要經驗與教訓

---

## 已知地雷與解決方案

| 風險 | 等級 | 問題描述 | 解決方案 |
|------|------|----------|----------|
| **密碼 Hash 格式** | 🔴 | Argon2id vs Bcrypt 格式不同導致登入失敗 | 後端統一用 `$argon2id$v=19$m=65536,t=1,p=4$...` |
| **IndexedDB 跨裝置** | 🔴 | 每個瀏覽器有獨立 IndexedDB，金鑰無法同步 | 改用密碼推導金鑰 (PBKDF2) |
| **JWT none 攻擊** | 🔴 | 演算法混淆攻擊 | 白名單只接受 HS256 |
| **webtransport-go** | 🟠 | 停止維護 | WebSocket Fallback 雙軌制 |
| **Svelte 5 SSR** | 🟠 | 狀態洩漏風險 | 禁用 SSR，純 SPA 模式 |
| **tweetnacl 無 padding** | 🟠 | 訊息長度洩漏 | 隨機 padding 到 64-byte 邊界，最小 256 bytes |

---

## 重要經驗教訓

### 1. E2EE 金鑰管理

**問題**: 每個裝置的 IndexedDB 獨立，導致：
- 換裝置 → 金鑰不存在 → 重新生成 → 公鑰不匹配 → 解密失敗

**解決方案**: 密碼推導確定性金鑰
```typescript
// 同樣的 password + userId = 同樣的 keypair
const salt = new TextEncoder().encode(`link-e2e-${userId}`);
const bits = await crypto.subtle.deriveBits(
  { name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' },
  keyMaterial, 256
);
const secretKey = new Uint8Array(bits);
const keyPair = nacl.box.keyPair.fromSecretKey(secretKey);
```

### 2. NFC 認證階段

| 階段 | 晶片 | 安全性 | 說明 |
|------|------|--------|------|
| Demo | NTAG215 | 基本 | 固定 UID，開源可驗證 |
| 正式 | NTAG424 DNA | 高 | SUN 零知識驗證，防克隆 |

### 3. 開源透明模型

密碼傳輸的安全性依賴：
- 程式碼公開可審計
- 使用者可自行部署
- 後端只做 Argon2id 比對，不記錄明文

---

## 技術架構摘要

```
Frontend (Svelte 5)          Backend (Go 1.22+)         Database
┌─────────────────┐         ┌─────────────────┐       ┌──────────┐
│ TweetNaCl E2EE  │◀──────▶ │ Fiber + WS/WT   │◀─────▶│ PG 15    │
│ PBKDF2 金鑰推導  │  HTTPS  │ Argon2id        │  pgx  │          │
│ IndexedDB 快取   │  WSS    │ JWT HS256       │       │          │
└─────────────────┘         └─────────────────┘       └──────────┘
```

### 加密流程
1. 發送方: `nacl.box(plaintext, nonce, recipientPubKey, senderSecretKey)`
2. 伺服器: 只傳遞密文
3. 接收方: `nacl.box.open(ciphertext, nonce, senderPubKey, recipientSecretKey)`

---

## 開發指令速查

```bash
# 後端
cd backend && make dev          # 開發模式
cd backend && go test ./... -v  # 測試

# 前端
cd frontend && pnpm dev         # 開發模式
cd frontend && pnpm build       # 建置

# 資料庫
brew services start postgresql@15
cd backend && make migrate-up
```

---

## Seed Script (Demo 快速建立)

當需要重頭 demo 或開發測試時，可用 seed script 快速建立資料：

```bash
cd backend

# 1. 複製設定範例
cp .env.seed.example .env.seed

# 2. 編輯 .env.seed，填入 NFC 卡片 token
#    SEED_DEMO_PRIMARY_TOKEN=你的主卡token
#    SEED_DEMO_BACKUP_TOKEN=你的附卡token

# 3. 執行 seed (會清空現有資料)
source .env.seed && go run ./cmd/seed
```

**注意**:
- `.env.seed` 包含敏感資料，已加入 `.gitignore`
- NFC 卡片 token 需與燒錄到實體卡片的一致
- Seed 會清空所有資料後重建 demo 用戶

---

## 檔案結構重點

```
backend/
├── cmd/
│   ├── server/main.go              # 應用程式入口
│   └── seed/main.go                # Seed Script
├── internal/
│   ├── pkg/password/argon2.go      # 密碼雜湊
│   ├── pkg/token/jwt.go            # JWT 管理
│   └── transport/                  # WebSocket/WebTransport
└── migrations/                     # 資料庫遷移

frontend/
├── src/lib/crypto/
│   ├── keys.ts                     # 金鑰管理 + 推導
│   ├── encrypt.ts                  # 加密 + padding
│   └── decrypt.ts                  # 解密
├── src/lib/transport/              # 傳輸層雙軌制
└── src/routes/chat/+page.svelte    # 聊天主頁
```

---

## 待辦事項

- [ ] NTAG424 DNA 零知識驗證（等硬體到貨）
- [ ] WebTransport 後端完整實作
- [ ] 訊息已讀狀態顯示
