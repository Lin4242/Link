# LINK 專案開發指引

> **重啟後先讀**: `note/handoff.md` 有最新狀態與待辦事項

## 核心原則
1. **Zero Trust Server** - 伺服器不信任，只傳密文
2. **開源透明** - 程式碼公開可審計，建立信任
3. **多裝置支援** - 密碼推導金鑰，任何裝置同一 keypair
4. **雙軌傳輸** - WebTransport 優先，WebSocket Fallback

## 重要：密碼與金鑰

### 密碼 Hash 格式
後端使用 **Argon2id**，格式必須是：
```
$argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
```
**不是** bcrypt (`$2a$...`)，搞混會導致登入失敗。

### E2EE 金鑰推導 (v4.3)
```typescript
// 同一組 password + userId = 同一組 keypair（任何裝置）
const salt = new TextEncoder().encode(`link-e2e-${userId}`);
const bits = await crypto.subtle.deriveBits(
  { name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' },
  keyMaterial, 256
);
const keyPair = nacl.box.keyPair.fromSecretKey(new Uint8Array(bits));
```

### 金鑰流程
1. **註冊**: 臨時 keypair → 取得 userId → 推導確定性 keypair → 更新伺服器公鑰
2. **登入**: 嘗試 IndexedDB → 若無則推導 → 比對伺服器公鑰
3. **換裝置**: 輸入密碼 → 推導相同 keypair → 無縫解密

## 關鍵檔案

| 檔案 | 用途 |
|------|------|
| `frontend/src/lib/crypto/keys.ts` | 金鑰推導、IndexedDB 存取 |
| `frontend/src/lib/crypto/encrypt.ts` | 加密 + padding |
| `frontend/src/lib/crypto/decrypt.ts` | 解密 |
| `frontend/src/routes/chat/+page.svelte` | 聊天主頁、金鑰解鎖 |
| `backend/internal/pkg/password/argon2.go` | Argon2id 實作 |
| `backend/internal/transport/hub.go` | WebSocket 連線管理 |

## 已知地雷

| 問題 | 原因 | 解法 |
|------|------|------|
| 登入失敗 | 密碼 hash 格式錯誤 | 確認是 argon2id 不是 bcrypt |
| 解密失敗 | 公鑰不匹配 | 檢查 DB 的 public_key 是否正確 |
| 換裝置無法解密 | IndexedDB 跨裝置不同步 | 已改用密碼推導金鑰 |
| ASCII 圖表錯位 | 中文字寬度不一致 | 用英文繪製框線圖 |

## 代碼風格

### Go
- gofmt + golangci-lint
- Error 放最後 return，Context 第一個參數
- Repository 必須有介面，Service 依賴介面

### TypeScript
- biome 格式化
- Svelte 5 Runes (`$state`, `$derived`, `$effect`)
- 加解密邏輯集中在 `lib/crypto/`

## 禁止事項
- 伺服器解密或記錄訊息內容
- 私鑰離開客戶端
- console.log 生產代碼
- 硬編碼 secrets
- JWT none 算法
- SELECT *

## 常用指令
```bash
# 開發
cd backend && make dev
cd frontend && pnpm dev

# 測試
cd backend && go test ./... -v
cd frontend && pnpm test

# 建置部署
cd frontend && pnpm build
scp -r frontend/build/* server:/path/to/build/

# 資料庫
psql -h localhost -U link -d link
```

## 資料庫快速查詢
```sql
-- 查看用戶
SELECT id, nickname, public_key FROM users;

-- 更新公鑰
UPDATE users SET public_key = 'base64...' WHERE nickname = 'F';

-- 查看卡片狀態
SELECT * FROM cards WHERE user_id = 'uuid...';
```

## 部署資訊
- **前端**: SvelteKit SPA，build 後為靜態檔案
- **後端**: Go binary，需要 TLS 憑證
- **資料庫**: PostgreSQL 15+
- **詳細部署指南**: `note/deployment.md`
