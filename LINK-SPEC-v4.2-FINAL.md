# LINK 即時通訊系統 - 完整開發規格

**版本**: 4.2 (Final Audit)  
**日期**: 2025-12-12  
**用途**: Claude CLI 開發指引與技術規格

---

## 目錄

1. [專案概述與開發原則](#1-專案概述與開發原則)
2. [技術棧與依賴](#2-技術棧與依賴)
3. [專案結構](#3-專案結構)
4. [端對端加密設計](#4-端對端加密設計)
5. [資料庫設計](#5-資料庫設計)
6. [後端實作](#6-後端實作)
7. [前端實作](#7-前端實作)
8. [WebTransport + WebSocket 協議](#8-webtransport--websocket-協議)
9. [測試策略](#9-測試策略)
10. [Agent 任務分配](#10-agent-任務分配)
11. [執行步驟](#11-執行步驟)
12. [附錄](#12-附錄)

---

## ⚠️ 已知地雷與預防措施

| 風險 | 等級 | 預防措施 |
|------|------|----------|
| webtransport-go 停止維護 | 🔴 | WebSocket Fallback 雙軌制 |
| Svelte 5 SSR 狀態洩漏 | 🔴 | 禁用 SSR (SPA 模式) |
| JWT none 算法攻擊 | 🔴 | 算法白名單 + 嚴格驗證 |
| pgxpool 連線池死鎖 | 🔴 | 超時配置 + Circuit Breaker |
| tweetnacl 無 padding | 🟠 | 隨機 padding 到固定區塊 |
| IndexedDB 私鑰安全 | 🟠 | 用戶警告提示 |
| 缺少 Rate Limiting | 🟡 | 登入/註冊限速 |

---

## 1. 專案概述與開發原則

### 1.1 系統簡介

**LINK** - NFC 卡片認證端對端加密即時通訊系統

### 1.2 Phase 1 範圍

| 功能 | 優先級 | 說明 |
|------|--------|------|
| NFC 雙卡認證 | P0 | 主卡日常用 + 附卡緊急撤銷 |
| 端對端加密聊天 | P0 | X25519 + XSalsa20-Poly1305 + Padding |
| 好友系統 | P0 | 發送/接受/拒絕好友請求 |
| 1-on-1 即時聊天 | P0 | WebTransport + WebSocket Fallback |
| 在線狀態 | P1 | 顯示好友是否在線 |
| 打字中提示 | P1 | 對方輸入時顯示 |
| 訊息歷史 | P1 | 本地儲存（伺服器只存密文） |

**Phase 1 不做**: Forward Secrecy、群組聊天、檔案傳輸、多裝置同步

### 1.2.1 雙卡機制

```
┌─────────────────────────────────────────────────────────────────┐
│                        雙卡安全機制                              │
├─────────────────────────────────────────────────────────────────┤
│  註冊時                                                         │
│  ├── 主卡 (Primary) → 日常使用，放錢包                          │
│  └── 附卡 (Backup)  → 緊急備援，放保險箱                        │
│                                                                 │
│  主卡遺失時                                                     │
│  └── 刷附卡 → 主卡立即失效 → 附卡升級為主卡                     │
│              → 強制登出所有 session                             │
│              → 帳號進入「單卡狀態」(無法再撤銷)                  │
│                                                                 │
│  ⚠️  附卡只能用一次，用完需重新註冊新帳號配新雙卡                │
└─────────────────────────────────────────────────────────────────┘
```

**使用場景**：
1. 正常登入：刷主卡 + 輸入密碼 → 正常進入
2. 緊急撤銷：刷附卡 + 輸入密碼 → 警告確認 → 主卡作廢 → 進入系統
3. 單卡狀態：只能用（原附卡現主卡）登入，無法再撤銷

### 1.3 安全架構

```
┌─────────────────────────────────────────────────────────────────┐
│                      Phase 1 安全層級                            │
├─────────────────────────────────────────────────────────────────┤
│  ✅ 傳輸層: WebTransport/WebSocket = TLS 1.3                    │
│  ✅ 應用層: E2EE (X25519 + XSalsa20-Poly1305 + Padding)        │
│  ✅ 密碼儲存: Argon2id (OWASP 參數)                             │
│  ✅ 認證: JWT (HS256 白名單) + NFC 卡片 token                   │
│  ✅ 連線穩定: Circuit Breaker + 連線池優化                      │
│  ✅ 防暴力: Rate Limiting                                       │
│  ⚠️  無 Forward Secrecy (Phase 2 加 Double Ratchet)             │
│  ⚠️  私鑰存本地 (遺失 = 歷史訊息無法解密)                        │
└─────────────────────────────────────────────────────────────────┘

伺服器可見: metadata (誰發給誰、時間)
伺服器不可見: 訊息內容 ✅
```

### 1.4 NFC 認證流程

```
用戶掃描 NFC 卡片 → 開啟 https://domain.com/w/{card_token}
    ↓
伺服器查詢 card_token
    ├─ 未註冊 → 檢查是否為配對卡
    │           ├─ 無配對 → 提示「請同時準備主卡和附卡」
    │           └─ 有配對 → 註冊頁（設密碼、暱稱、生成 keypair）
    │
    └─ 已註冊 → 檢查卡片類型
                ├─ 主卡 (active) → 登入頁（輸入密碼）
                ├─ 附卡 (active) → ⚠️ 警告頁「使用附卡將撤銷主卡」
                │                  → 確認 + 密碼 → 撤銷主卡 → 登入
                └─ 已撤銷 → 錯誤「此卡已失效」
    ↓
驗證成功 → JWT → WebTransport/WebSocket 連線
```

### 1.4.1 雙卡配對註冊流程

```
Step 1: 掃描主卡 → 記錄 primary_token → 提示「請掃描附卡」
Step 2: 掃描附卡 → 記錄 backup_token → 顯示註冊表單
Step 3: 填寫密碼、暱稱 → 生成 keypair
Step 4: POST /auth/register { primary_token, backup_token, password, ... }
Step 5: 兩張卡同時綁定到帳號
```

### 1.5 開發原則 (CLAUDE.md)

**直接複製此內容到 `link/CLAUDE.md`：**

```markdown
# LINK 專案開發指引

## 核心原則
1. **Zero Trust Server** - 伺服器不信任，只傳密文
2. **依賴反轉** - Service 依賴 Repository 介面
3. **錯誤優先** - 先處理 error path
4. **統一格式** - API 錯誤走 AppError，回應走 handler.OK/Error
5. **雙軌傳輸** - WebTransport 優先，WebSocket Fallback

## 代碼風格
### Go
- gofmt + golangci-lint
- Error 放最後 return，Context 第一個參數
- Repository 必須有介面，Service 依賴介面
- 不用 panic（除 init 和密鑰驗證）

### TypeScript
- biome
- 禁止 var 和 any
- Svelte 5 Runes ($state, $derived, $effect)
- 加解密邏輯集中在 lib/crypto/

## 檔案命名
- Go: snake_case.go
- TypeScript: kebab-case.ts
- Svelte: PascalCase.svelte
- 測試: *_test.go / *.test.ts

## 禁止
- 伺服器解密或記錄訊息內容
- 私鑰離開客戶端
- console.log 生產代碼（用結構化 log）
- 硬編碼 secrets
- 忽略 error
- SELECT *
- 超過 200 行的函數
- JWT none 算法
- 弱密鑰 (< 32 字元)

## 常用指令
cd backend && make dev      # 後端開發
cd backend && make test     # 後端測試
cd frontend && pnpm dev     # 前端開發
cd frontend && pnpm test    # 前端測試
```

---

## 2. 技術棧與依賴

### 2.1 後端

| 組件 | 選擇 | 版本 |
|------|------|------|
| 語言 | Go | 1.22+ |
| HTTP | Fiber | v2.52+ |
| WebTransport | quic-go/webtransport-go | 0.8.0 |
| WebSocket | gofiber/contrib/websocket | latest |
| 資料庫 | PostgreSQL | 15+ |
| 驅動 | pgx/v5 | 5.5.0 |
| 密碼 | Argon2id | golang.org/x/crypto |
| JWT | golang-jwt/jwt/v5 | 5.2.0 |

**go.mod**:
```go
module link

go 1.22

require (
    github.com/gofiber/fiber/v2 v2.52.0
    github.com/gofiber/contrib/websocket v1.3.0
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/jackc/pgx/v5 v5.5.0
    github.com/quic-go/webtransport-go v0.8.0
    golang.org/x/crypto v0.21.0
    github.com/joho/godotenv v1.5.1
    github.com/google/uuid v1.6.0
)
```

### 2.2 前端

| 組件 | 選擇 | 版本 | 說明 |
|------|------|------|------|
| 框架 | Svelte | 5.x | Runes |
| 元框架 | SvelteKit | 2.x | SPA 模式 |
| 加密 | tweetnacl | 1.0.3 | libsodium 相容 |
| 加密輔助 | tweetnacl-util | 0.15.1 | base64 編碼 |
| 樣式 | Tailwind CSS | 4.x | |

**package.json dependencies**:
```json
{
  "dependencies": {
    "tweetnacl": "^1.0.3",
    "tweetnacl-util": "^0.15.1"
  }
}
```

### 2.3 環境變數

**backend/.env**:
```env
SERVER_ADDR=:8443
SERVER_ENV=development
DATABASE_URL=postgres://app:secret@localhost:5432/link?sslmode=disable
JWT_SECRET=change-this-to-64-chars-minimum-use-openssl-rand-hex-32
JWT_EXPIRY=24h
TLS_CERT_FILE=./certs/localhost+2.pem
TLS_KEY_FILE=./certs/localhost+2-key.pem
CORS_ORIGINS=https://localhost:5173
LOG_LEVEL=debug
```

**frontend/.env**:
```env
VITE_API_URL=https://localhost:8443
VITE_WT_URL=https://localhost:8443/wt
VITE_WS_URL=wss://localhost:8443/ws
```

---

## 3. 專案結構

```
link/
├── CLAUDE.md                    # ⭐ 必須創建，內容見 1.5
├── docker-compose.yml
├── scripts/setup-certs.sh
├── certs/
│
├── backend/
│   ├── cmd/server/main.go       # 入口 + DI
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── domain/
│   │   │   ├── errors.go
│   │   │   ├── user.go
│   │   │   ├── card.go          # ⭐ 雙卡機制
│   │   │   ├── session.go       # ⭐ Session 管理
│   │   │   ├── friendship.go
│   │   │   ├── conversation.go
│   │   │   └── message.go
│   │   ├── repository/
│   │   │   ├── postgres/
│   │   │   │   ├── user.go
│   │   │   │   ├── card.go      # ⭐ 卡片 Repository
│   │   │   │   ├── session.go   # ⭐ Session Repository
│   │   │   │   └── ...
│   │   │   └── mock/
│   │   ├── service/
│   │   │   ├── auth.go          # ⭐ 含雙卡登入邏輯
│   │   │   ├── card.go          # ⭐ 卡片服務
│   │   │   ├── user.go
│   │   │   ├── friendship.go
│   │   │   └── message.go
│   │   ├── handler/
│   │   │   ├── auth.go          # ⭐ 含雙卡 API
│   │   │   ├── user.go
│   │   │   ├── friendship.go
│   │   │   ├── conversation.go
│   │   │   ├── response.go
│   │   │   └── routes.go
│   │   ├── dto/
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── ratelimit.go
│   │   │   ├── security.go
│   │   │   └── logger.go
│   │   ├── transport/
│   │   │   ├── protocol.go
│   │   │   ├── hub.go
│   │   │   ├── client.go        # WebTransport client
│   │   │   ├── ws_client.go     # WebSocket client
│   │   │   ├── handler.go
│   │   │   └── server.go
│   │   └── pkg/
│   │       ├── password/argon2.go
│   │       ├── token/jwt.go
│   │       └── circuitbreaker/
│   │           └── breaker.go
│   ├── migrations/
│   ├── go.mod
│   ├── Makefile
│   └── .air.toml
│
└── frontend/
    ├── src/
    │   ├── lib/
    │   │   ├── crypto/
    │   │   │   ├── keys.ts
    │   │   │   ├── encrypt.ts
    │   │   │   ├── decrypt.ts
    │   │   │   └── index.ts
    │   │   ├── stores/
    │   │   │   ├── auth.svelte.ts
    │   │   │   ├── keys.svelte.ts
    │   │   │   ├── messages.svelte.ts
    │   │   │   ├── conversations.svelte.ts
    │   │   │   ├── friends.svelte.ts
    │   │   │   ├── transport.svelte.ts
    │   │   │   └── index.ts
    │   │   ├── api/
    │   │   │   ├── client.ts
    │   │   │   ├── auth.ts      # ⭐ 含雙卡 API
    │   │   │   ├── users.ts
    │   │   │   ├── friends.ts
    │   │   │   └── conversations.ts
    │   │   ├── transport/
    │   │   │   ├── webtransport.ts
    │   │   │   ├── websocket.ts
    │   │   │   └── index.ts
    │   │   ├── components/
    │   │   │   ├── SecurityWarning.svelte
    │   │   │   └── BackupCardWarning.svelte  # ⭐ 附卡警告
    │   │   └── types.ts
    │   ├── routes/
    │   │   ├── +layout.svelte
    │   │   ├── +page.svelte
    │   │   ├── register/
    │   │   │   ├── start/+page.svelte    # ⭐ 掃主卡
    │   │   │   └── pair/+page.svelte     # ⭐ 掃附卡
    │   │   ├── login/
    │   │   │   ├── +page.svelte          # 主卡登入
    │   │   │   └── backup/+page.svelte   # ⭐ 附卡登入
    │   │   └── chat/
    │   │       └── +page.svelte
    │   └── app.css
    ├── svelte.config.js
    ├── vitest.config.ts
    ├── biome.json
    └── package.json
```

---

## 4. 端對端加密設計

### 4.1 加密演算法

```
金鑰交換: X25519 (Curve25519 ECDH)
加密: XSalsa20-Poly1305 (AEAD)
Padding: 隨機填充到 64 bytes 倍數，最小 256 bytes
實作: tweetnacl
```

### 4.2 金鑰管理生命週期

```
1. 註冊時
   ├── nacl.box.keyPair() 生成 keypair
   ├── 公鑰 → POST /auth/register → 存 DB
   └── 私鑰 → PBKDF2(密碼) 加密 → IndexedDB

2. 登入時
   ├── 從 IndexedDB 載入加密的私鑰
   ├── PBKDF2(密碼) 解密私鑰
   └── 若無私鑰 → 警告（無法解密歷史）

3. 加好友/開對話時
   ├── GET /users/:id/public-key
   └── 快取到 keys store

4. 發訊息時
   ├── 取對方公鑰（從快取）
   ├── padMessage() 填充訊息
   ├── nacl.box(paddedMsg, nonce, theirPubKey, mySecKey)
   └── 發送 { nonce, ciphertext }

5. 收訊息時
   ├── nacl.box.open(ciphertext, nonce, theirPubKey, mySecKey)
   ├── unpadMessage() 移除填充
   └── 顯示明文
```

### 4.3 公鑰獲取策略

```typescript
// 公鑰獲取時機（自動、透明）
// 1. 登入後載入好友列表時，好友資料包含公鑰
// 2. 接受好友請求時，回應包含對方公鑰
// 3. 開啟對話時，若無快取則 fetch

async function ensurePublicKey(userId: string): Promise<string> {
    let pk = publicKeyCache[userId];
    if (!pk) {
        const res = await usersApi.getPublicKey(userId);
        pk = res.public_key;
        publicKeyCache[userId] = pk;
    }
    return pk;
}
```

### 4.4 訊息格式

```typescript
// 傳輸格式（伺服器儲存）
{
  encrypted_content: "{\"nonce\":\"base64...\",\"ciphertext\":\"base64...\"}"
}

// 解密後（本地使用）
{
  content: "Hello!"
}
```

---

## 5. 資料庫設計

### 5.1 Migration

**migrations/001_init.up.sql**:
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    password_hash   VARCHAR(256) NOT NULL,
    nickname        VARCHAR(50) NOT NULL,
    public_key      VARCHAR(64) NOT NULL,
    avatar_url      VARCHAR(512),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at    TIMESTAMPTZ
);

-- 雙卡機制：每個用戶最多 2 張卡（主卡 + 附卡）
CREATE TABLE cards (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_token      VARCHAR(32) UNIQUE NOT NULL,
    card_type       VARCHAR(10) NOT NULL CHECK (card_type IN ('primary', 'backup')),
    status          VARCHAR(10) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    activated_at    TIMESTAMPTZ,  -- 附卡升級為主卡的時間
    revoked_at      TIMESTAMPTZ
);
CREATE INDEX idx_cards_user ON cards(user_id);
CREATE INDEX idx_cards_token ON cards(card_token);
-- 每種類型只能有一張 active 卡
CREATE UNIQUE INDEX idx_cards_user_type_active ON cards(user_id, card_type) WHERE status = 'active';

-- 配對暫存表（註冊流程用）
CREATE TABLE card_pairs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    primary_token   VARCHAR(32) UNIQUE NOT NULL,
    backup_token    VARCHAR(32) UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '30 minutes'
);
CREATE INDEX idx_card_pairs_expires ON card_pairs(expires_at);

CREATE TABLE friendships (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    requester_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    addressee_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (requester_id, addressee_id),
    CHECK (requester_id != addressee_id)
);
CREATE INDEX idx_friendships_requester ON friendships(requester_id, status);
CREATE INDEX idx_friendships_addressee ON friendships(addressee_id, status);

CREATE TABLE conversations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    participant_1   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    participant_2   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_message_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (participant_1, participant_2),
    CHECK (participant_1 < participant_2)
);

CREATE TABLE messages (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id   UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    encrypted_content TEXT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at      TIMESTAMPTZ,
    read_at           TIMESTAMPTZ
);
CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at DESC);

-- 登入 Session 追蹤（用於強制登出）
CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(64) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ
);
CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token_hash);

-- Triggers
CREATE OR REPLACE FUNCTION update_conversation_last_message()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE conversations SET last_message_at = NEW.created_at WHERE id = NEW.conversation_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_message_insert AFTER INSERT ON messages
FOR EACH ROW EXECUTE FUNCTION update_conversation_last_message();

CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER trg_friendships_updated BEFORE UPDATE ON friendships FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- 清理過期的配對暫存
CREATE OR REPLACE FUNCTION cleanup_expired_pairs()
RETURNS void AS $$
BEGIN
    DELETE FROM card_pairs WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;
```

**migrations/001_init.down.sql**:
```sql
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS friendships;
DROP TABLE IF EXISTS card_pairs;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS users;
DROP FUNCTION IF EXISTS update_conversation_last_message();
DROP FUNCTION IF EXISTS update_timestamp();
DROP FUNCTION IF EXISTS cleanup_expired_pairs();
```

---

## 6. 後端實作

### 6.1 Config

**internal/config/config.go**:
```go
package config

import (
    "os"
    "time"
)

type Config struct {
    ServerAddr  string
    ServerEnv   string
    DatabaseURL string
    JWTSecret   string
    JWTExpiry   time.Duration
    TLSCert     string
    TLSKey      string
    CORSOrigins string
    LogLevel    string
}

func Load() *Config {
    secret := getEnv("JWT_SECRET", "")
    if len(secret) < 32 {
        panic("JWT_SECRET must be at least 32 characters")
    }
    
    expiry, _ := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
    return &Config{
        ServerAddr:  getEnv("SERVER_ADDR", ":8443"),
        ServerEnv:   getEnv("SERVER_ENV", "development"),
        DatabaseURL: getEnv("DATABASE_URL", ""),
        JWTSecret:   secret,
        JWTExpiry:   expiry,
        TLSCert:     getEnv("TLS_CERT_FILE", "./certs/localhost+2.pem"),
        TLSKey:      getEnv("TLS_KEY_FILE", "./certs/localhost+2-key.pem"),
        CORSOrigins: getEnv("CORS_ORIGINS", "https://localhost:5173"),
        LogLevel:    getEnv("LOG_LEVEL", "info"),
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

### 6.2 Domain

**internal/domain/errors.go**:
```go
package domain

import "errors"

const (
    ErrCodeValidation   = "VALIDATION_ERROR"
    ErrCodeNotFound     = "NOT_FOUND"
    ErrCodeUnauthorized = "UNAUTHORIZED"
    ErrCodeConflict     = "CONFLICT"
    ErrCodeInternal     = "INTERNAL_ERROR"
    ErrCodeRateLimited  = "RATE_LIMITED"
)

type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"-"`
}

func (e *AppError) Error() string { return e.Message }

func ErrValidation(msg string) *AppError   { return &AppError{ErrCodeValidation, msg, 400} }
func ErrNotFound(msg string) *AppError     { return &AppError{ErrCodeNotFound, msg, 404} }
func ErrUnauthorized(msg string) *AppError { return &AppError{ErrCodeUnauthorized, msg, 401} }
func ErrConflict(msg string) *AppError     { return &AppError{ErrCodeConflict, msg, 409} }
func ErrInternal() *AppError               { return &AppError{ErrCodeInternal, "系統錯誤", 500} }
func ErrRateLimited() *AppError            { return &AppError{ErrCodeRateLimited, "請求過於頻繁", 429} }

var (
    ErrUserNotFound         = ErrNotFound("用戶不存在")
    ErrInvalidPassword      = ErrUnauthorized("密碼錯誤")
    ErrInvalidToken         = ErrUnauthorized("無效的 token")
    ErrCardAlreadyUsed      = ErrConflict("卡片已被註冊")
    ErrAlreadyFriends       = ErrConflict("已經是好友")
    ErrSelfFriendRequest    = ErrValidation("不能加自己為好友")
    ErrConversationNotFound = ErrNotFound("對話不存在")
)

func IsAppError(err error) (*AppError, bool) {
    var appErr *AppError
    if errors.As(err, &appErr) {
        return appErr, true
    }
    return nil, false
}
```

**internal/domain/user.go**:
```go
package domain

import (
    "context"
    "time"
)

type User struct {
    ID           string
    PasswordHash string
    Nickname     string
    PublicKey    string
    AvatarURL    *string
    CreatedAt    time.Time
    UpdatedAt    time.Time
    LastSeenAt   *time.Time
}

type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id string) (*User, error)
    GetPublicKey(ctx context.Context, id string) (string, error)
    Update(ctx context.Context, user *User) error
    UpdateLastSeen(ctx context.Context, id string) error
    Search(ctx context.Context, query string, limit int) ([]*User, error)
}
```

**internal/domain/card.go**:
```go
package domain

import (
    "context"
    "time"
)

type CardType string
type CardStatus string

const (
    CardTypePrimary CardType = "primary"
    CardTypeBackup  CardType = "backup"

    CardStatusActive  CardStatus = "active"
    CardStatusRevoked CardStatus = "revoked"
)

type Card struct {
    ID          string
    UserID      string
    CardToken   string
    CardType    CardType
    Status      CardStatus
    CreatedAt   time.Time
    ActivatedAt *time.Time  // 附卡升級時間
    RevokedAt   *time.Time
}

type CardPair struct {
    ID           string
    PrimaryToken string
    BackupToken  *string
    CreatedAt    time.Time
    ExpiresAt    time.Time
}

type CardRepository interface {
    // 卡片查詢
    FindByToken(ctx context.Context, token string) (*Card, error)
    FindByUserID(ctx context.Context, userID string) ([]*Card, error)
    FindActiveByUserAndType(ctx context.Context, userID string, cardType CardType) (*Card, error)
    
    // 卡片操作
    Create(ctx context.Context, card *Card) error
    Revoke(ctx context.Context, cardID string) error
    PromoteBackupToPrimary(ctx context.Context, cardID string) error
    
    // 配對暫存
    CreatePair(ctx context.Context, primaryToken string) (*CardPair, error)
    FindPairByPrimaryToken(ctx context.Context, token string) (*CardPair, error)
    FindPairByBackupToken(ctx context.Context, token string) (*CardPair, error)
    UpdatePairBackupToken(ctx context.Context, pairID, backupToken string) error
    DeletePair(ctx context.Context, pairID string) error
    CleanupExpiredPairs(ctx context.Context) error
}

// 撤銷主卡並升級附卡的事務操作
type CardService interface {
    // 檢查卡片狀態
    CheckCard(ctx context.Context, token string) (*CardCheckResult, error)
    
    // 使用附卡撤銷主卡（原子操作）
    RevokeWithBackupCard(ctx context.Context, backupCardID, userID string) error
}

type CardCheckResult struct {
    Status     string  // "not_found", "pair_started", "pair_waiting", "primary", "backup", "revoked"
    UserID     *string
    Nickname   *string
    CardType   *CardType
    PairID     *string
    Warning    *string // 附卡警告訊息
}
```

**internal/domain/friendship.go**:
```go
package domain

import (
    "context"
    "time"
)

type FriendshipStatus string

const (
    FriendshipPending  FriendshipStatus = "pending"
    FriendshipAccepted FriendshipStatus = "accepted"
)

type Friendship struct {
    ID          string
    RequesterID string
    AddresseeID string
    Status      FriendshipStatus
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type FriendWithUser struct {
    Friendship
    Friend *User
}

type FriendshipRepository interface {
    Create(ctx context.Context, f *Friendship) error
    FindByUsers(ctx context.Context, userA, userB string) (*Friendship, error)
    FindFriends(ctx context.Context, userID string) ([]*FriendWithUser, error)
    FindPendingRequests(ctx context.Context, userID string) ([]*FriendWithUser, error)
    UpdateStatus(ctx context.Context, id string, status FriendshipStatus) error
    Delete(ctx context.Context, id string) error
}
```

**internal/domain/conversation.go**:
```go
package domain

import (
    "context"
    "time"
)

type Conversation struct {
    ID            string
    Participant1  string
    Participant2  string
    LastMessageAt *time.Time
    CreatedAt     time.Time
}

type ConversationWithPeer struct {
    Conversation
    Peer        *User
    UnreadCount int
}

type ConversationRepository interface {
    Create(ctx context.Context, c *Conversation) error
    FindByID(ctx context.Context, id string) (*Conversation, error)
    FindByParticipants(ctx context.Context, userA, userB string) (*Conversation, error)
    FindByUser(ctx context.Context, userID string) ([]*ConversationWithPeer, error)
    GetOrCreate(ctx context.Context, userA, userB string) (*Conversation, error)
}
```

**internal/domain/message.go**:
```go
package domain

import (
    "context"
    "time"
)

type Message struct {
    ID               string
    ConversationID   string
    SenderID         string
    EncryptedContent string
    CreatedAt        time.Time
    DeliveredAt      *time.Time
    ReadAt           *time.Time
}

type MessageRepository interface {
    Create(ctx context.Context, msg *Message) error
    FindByConversation(ctx context.Context, convID string, limit int, before *time.Time) ([]*Message, error)
    MarkDelivered(ctx context.Context, id string) error
    MarkRead(ctx context.Context, id string) error
}
```

**internal/domain/session.go**:
```go
package domain

import (
    "context"
    "time"
)

type Session struct {
    ID        string
    UserID    string
    TokenHash string
    CreatedAt time.Time
    ExpiresAt time.Time
    RevokedAt *time.Time
}

type SessionRepository interface {
    Create(ctx context.Context, session *Session) error
    FindByTokenHash(ctx context.Context, hash string) (*Session, error)
    RevokeAllByUser(ctx context.Context, userID string) error
    Revoke(ctx context.Context, id string) error
    CleanupExpired(ctx context.Context) error
}
```

### 6.2.1 Service Layer

**internal/service/card.go**:
```go
package service

import (
    "context"
    "link/internal/domain"
)

type CardService struct {
    cardRepo    domain.CardRepository
    sessionRepo domain.SessionRepository
}

func NewCardService(cardRepo domain.CardRepository, sessionRepo domain.SessionRepository) *CardService {
    return &CardService{cardRepo: cardRepo, sessionRepo: sessionRepo}
}

func (s *CardService) CheckCard(ctx context.Context, token string) (*domain.CardCheckResult, error) {
    // 1. 檢查是否已註冊
    card, err := s.cardRepo.FindByToken(ctx, token)
    if err == nil && card != nil {
        if card.Status == domain.CardStatusRevoked {
            return &domain.CardCheckResult{Status: "revoked"}, nil
        }
        warning := ""
        if card.CardType == domain.CardTypeBackup {
            warning = "此為備援卡，使用後主卡將失效"
        }
        return &domain.CardCheckResult{
            Status:   string(card.CardType),
            UserID:   &card.UserID,
            CardType: &card.CardType,
            Warning:  &warning,
        }, nil
    }
    
    // 2. 檢查是否在配對流程中
    pair, _ := s.cardRepo.FindPairByPrimaryToken(ctx, token)
    if pair != nil {
        if pair.BackupToken != nil {
            return &domain.CardCheckResult{Status: "pair_waiting", PairID: &pair.ID}, nil
        }
        return &domain.CardCheckResult{Status: "pair_started", PairID: &pair.ID}, nil
    }
    
    pair, _ = s.cardRepo.FindPairByBackupToken(ctx, token)
    if pair != nil {
        return &domain.CardCheckResult{Status: "pair_waiting", PairID: &pair.ID}, nil
    }
    
    return &domain.CardCheckResult{Status: "not_found"}, nil
}

func (s *CardService) StartPair(ctx context.Context, primaryToken string) (*domain.CardPair, error) {
    // 檢查 token 是否已被使用
    existing, _ := s.cardRepo.FindByToken(ctx, primaryToken)
    if existing != nil {
        return nil, domain.ErrConflict("此卡片已被註冊")
    }
    
    return s.cardRepo.CreatePair(ctx, primaryToken)
}

func (s *CardService) CompletePair(ctx context.Context, primaryToken, backupToken string) error {
    if primaryToken == backupToken {
        return domain.ErrValidation("主卡和附卡不能是同一張")
    }
    
    pair, err := s.cardRepo.FindPairByPrimaryToken(ctx, primaryToken)
    if err != nil || pair == nil {
        return domain.ErrNotFound("配對不存在，請重新掃描主卡")
    }
    
    // 檢查附卡是否已被使用
    existing, _ := s.cardRepo.FindByToken(ctx, backupToken)
    if existing != nil {
        return domain.ErrConflict("附卡已被其他帳號使用")
    }
    
    return s.cardRepo.UpdatePairBackupToken(ctx, pair.ID, backupToken)
}

func (s *CardService) RevokeWithBackupCard(ctx context.Context, backupCardID, userID string) error {
    // 1. 找到主卡並撤銷
    primaryCard, err := s.cardRepo.FindActiveByUserAndType(ctx, userID, domain.CardTypePrimary)
    if err == nil && primaryCard != nil {
        if err := s.cardRepo.Revoke(ctx, primaryCard.ID); err != nil {
            return err
        }
    }
    
    // 2. 升級附卡為主卡
    if err := s.cardRepo.PromoteBackupToPrimary(ctx, backupCardID); err != nil {
        return err
    }
    
    // 3. 撤銷所有現有 session
    return s.sessionRepo.RevokeAllByUser(ctx, userID)
}
```

**internal/service/auth.go**:
```go
package service

import (
    "context"
    "link/internal/domain"
    "link/internal/pkg/password"
    "link/internal/pkg/token"
)

type AuthService struct {
    userRepo    domain.UserRepository
    cardRepo    domain.CardRepository
    sessionRepo domain.SessionRepository
    tokenMgr    *token.Manager
}

type RegisterInput struct {
    PrimaryToken string
    BackupToken  string
    Password     string
    Nickname     string
    PublicKey    string
}

type AuthResponse struct {
    User  *domain.User `json:"user"`
    Token string       `json:"token"`
}

func NewAuthService(
    userRepo domain.UserRepository,
    cardRepo domain.CardRepository,
    sessionRepo domain.SessionRepository,
    tokenMgr *token.Manager,
) *AuthService {
    return &AuthService{
        userRepo:    userRepo,
        cardRepo:    cardRepo,
        sessionRepo: sessionRepo,
        tokenMgr:    tokenMgr,
    }
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResponse, error) {
    // 驗證配對
    pair, err := s.cardRepo.FindPairByPrimaryToken(ctx, input.PrimaryToken)
    if err != nil || pair == nil || pair.BackupToken == nil || *pair.BackupToken != input.BackupToken {
        return nil, domain.ErrValidation("卡片配對無效")
    }
    
    // Hash 密碼
    hash, err := password.Hash(input.Password)
    if err != nil {
        return nil, domain.ErrInternal()
    }
    
    // 創建用戶
    user := &domain.User{
        PasswordHash: hash,
        Nickname:     input.Nickname,
        PublicKey:    input.PublicKey,
    }
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // 創建雙卡
    primaryCard := &domain.Card{
        UserID:    user.ID,
        CardToken: input.PrimaryToken,
        CardType:  domain.CardTypePrimary,
        Status:    domain.CardStatusActive,
    }
    backupCard := &domain.Card{
        UserID:    user.ID,
        CardToken: input.BackupToken,
        CardType:  domain.CardTypeBackup,
        Status:    domain.CardStatusActive,
    }
    if err := s.cardRepo.Create(ctx, primaryCard); err != nil {
        return nil, err
    }
    if err := s.cardRepo.Create(ctx, backupCard); err != nil {
        return nil, err
    }
    
    // 刪除配對暫存
    s.cardRepo.DeletePair(ctx, pair.ID)
    
    // 生成 token
    tokenStr, err := s.tokenMgr.Generate(user.ID)
    if err != nil {
        return nil, domain.ErrInternal()
    }
    
    return &AuthResponse{User: user, Token: tokenStr}, nil
}

func (s *AuthService) Login(ctx context.Context, cardToken, pwd string) (*AuthResponse, error) {
    card, err := s.cardRepo.FindByToken(ctx, cardToken)
    if err != nil || card == nil {
        return nil, domain.ErrUserNotFound
    }
    
    if card.Status == domain.CardStatusRevoked {
        return nil, domain.ErrUnauthorized("此卡片已失效")
    }
    
    if card.CardType == domain.CardTypeBackup {
        return nil, domain.ErrValidation("請使用主卡登入，或使用附卡撤銷流程")
    }
    
    user, err := s.userRepo.FindByID(ctx, card.UserID)
    if err != nil {
        return nil, domain.ErrUserNotFound
    }
    
    ok, err := password.Verify(pwd, user.PasswordHash)
    if err != nil || !ok {
        return nil, domain.ErrInvalidPassword
    }
    
    tokenStr, _ := s.tokenMgr.Generate(user.ID)
    return &AuthResponse{User: user, Token: tokenStr}, nil
}

func (s *AuthService) LoginWithBackupCard(ctx context.Context, cardToken, pwd string) (*AuthResponse, error) {
    card, err := s.cardRepo.FindByToken(ctx, cardToken)
    if err != nil || card == nil {
        return nil, domain.ErrUserNotFound
    }
    
    if card.Status == domain.CardStatusRevoked {
        return nil, domain.ErrUnauthorized("此卡片已失效")
    }
    
    if card.CardType != domain.CardTypeBackup {
        return nil, domain.ErrValidation("此為主卡，請使用一般登入")
    }
    
    user, err := s.userRepo.FindByID(ctx, card.UserID)
    if err != nil {
        return nil, domain.ErrUserNotFound
    }
    
    ok, err := password.Verify(pwd, user.PasswordHash)
    if err != nil || !ok {
        return nil, domain.ErrInvalidPassword
    }
    
    // 撤銷主卡並升級附卡
    cardSvc := &CardService{cardRepo: s.cardRepo, sessionRepo: s.sessionRepo}
    if err := cardSvc.RevokeWithBackupCard(ctx, card.ID, user.ID); err != nil {
        return nil, err
    }
    
    tokenStr, _ := s.tokenMgr.Generate(user.ID)
    return &AuthResponse{User: user, Token: tokenStr}, nil
}
```

### 6.3 Circuit Breaker

**internal/pkg/circuitbreaker/breaker.go**:
```go
package circuitbreaker

import (
    "errors"
    "sync"
    "time"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

type CircuitBreaker struct {
    mu              sync.RWMutex
    state           State
    failures        int
    threshold       int
    timeout         time.Duration
    lastFailureTime time.Time
}

func New(threshold int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        state:     StateClosed,
        threshold: threshold,
        timeout:   timeout,
    }
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mu.RLock()
    state := cb.state
    cb.mu.RUnlock()
    
    if state == StateOpen {
        cb.mu.Lock()
        if time.Since(cb.lastFailureTime) > cb.timeout {
            cb.state = StateHalfOpen
            cb.mu.Unlock()
        } else {
            cb.mu.Unlock()
            return ErrCircuitOpen
        }
    }
    
    err := fn()
    
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failures++
        cb.lastFailureTime = time.Now()
        if cb.failures >= cb.threshold {
            cb.state = StateOpen
        }
        return err
    }
    
    cb.failures = 0
    cb.state = StateClosed
    return nil
}
```

### 6.4 Password (Argon2id - OWASP 參數)

**internal/pkg/password/argon2.go**:
```go
package password

import (
    "crypto/rand"
    "crypto/subtle"
    "encoding/base64"
    "errors"
    "fmt"
    "runtime"
    "strings"
    
    "golang.org/x/crypto/argon2"
)

type Params struct {
    Memory      uint32
    Iterations  uint32
    Parallelism uint8
    SaltLength  uint32
    KeyLength   uint32
}

var ErrInvalidHash = errors.New("invalid hash format")

// DefaultParams - OWASP 建議參數
func DefaultParams() *Params {
    return &Params{
        Memory:      64 * 1024, // 64 MB
        Iterations:  3,
        Parallelism: uint8(runtime.NumCPU()),
        SaltLength:  16,
        KeyLength:   32,
    }
}

var params = DefaultParams()

func Hash(password string) (string, error) {
    salt := make([]byte, params.SaltLength)
    if _, err := rand.Read(salt); err != nil {
        return "", err
    }
    
    hash := argon2.IDKey(
        []byte(password), 
        salt, 
        params.Iterations, 
        params.Memory, 
        params.Parallelism, 
        params.KeyLength,
    )
    
    return fmt.Sprintf(
        "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
        argon2.Version,
        params.Memory,
        params.Iterations,
        params.Parallelism,
        base64.RawStdEncoding.EncodeToString(salt),
        base64.RawStdEncoding.EncodeToString(hash),
    ), nil
}

func Verify(password, encoded string) (bool, error) {
    parts := strings.Split(encoded, "$")
    if len(parts) != 6 || parts[1] != "argon2id" {
        return false, ErrInvalidHash
    }
    
    var memory, iterations uint32
    var parallelism uint8
    fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
    
    salt, _ := base64.RawStdEncoding.DecodeString(parts[4])
    expectedHash, _ := base64.RawStdEncoding.DecodeString(parts[5])
    
    hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expectedHash)))
    
    return subtle.ConstantTimeCompare(hash, expectedHash) == 1, nil
}
```

### 6.5 JWT Token (安全加固)

**internal/pkg/token/jwt.go**:
```go
package token

import (
    "errors"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

var (
    ErrInvalidToken     = errors.New("invalid token")
    ErrExpiredToken     = errors.New("token expired")
    ErrInvalidSignature = errors.New("invalid signature")
    ErrInvalidAlgorithm = errors.New("invalid algorithm")
)

type Manager struct {
    secret []byte
    expiry time.Duration
}

type Claims struct {
    UserID string `json:"uid"`
    jwt.RegisteredClaims
}

func NewManager(secret string, expiry time.Duration) *Manager {
    if len(secret) < 32 {
        panic("JWT secret must be at least 32 characters")
    }
    return &Manager{secret: []byte(secret), expiry: expiry}
}

func (m *Manager) Generate(userID string) (string, error) {
    claims := &Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.secret)
}

func (m *Manager) Verify(tokenStr string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
        // 🔴 關鍵：嚴格驗證算法，防止 none 攻擊
        if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
            return nil, ErrInvalidAlgorithm
        }
        return m.secret, nil
    })
    
    // 🔴 關鍵：正確處理錯誤
    if err != nil {
        if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
            return nil, ErrInvalidSignature
        }
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, ErrExpiredToken
        }
        return nil, ErrInvalidToken
    }
    
    if !token.Valid {
        return nil, ErrInvalidToken
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok {
        return nil, ErrInvalidToken
    }
    
    return claims, nil
}
```

**internal/pkg/token/jwt_test.go**:
```go
package token

import (
    "testing"
    "time"
)

func TestNoneAlgorithmAttack(t *testing.T) {
    mgr := NewManager("this-is-a-very-secure-secret-key-32", time.Hour)
    
    // 嘗試偽造 none 算法的 token
    noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1aWQiOiJhZG1pbiJ9."
    
    _, err := mgr.Verify(noneToken)
    if err == nil {
        t.Fatal("none algorithm token should be rejected")
    }
}

func TestWeakSecretPanic(t *testing.T) {
    defer func() {
        if r := recover(); r == nil {
            t.Fatal("should panic for weak secret")
        }
    }()
    NewManager("weak", time.Hour)
}

func TestValidToken(t *testing.T) {
    mgr := NewManager("this-is-a-very-secure-secret-key-32", time.Hour)
    
    tokenStr, err := mgr.Generate("user-123")
    if err != nil {
        t.Fatalf("failed to generate token: %v", err)
    }
    
    claims, err := mgr.Verify(tokenStr)
    if err != nil {
        t.Fatalf("failed to verify token: %v", err)
    }
    
    if claims.UserID != "user-123" {
        t.Errorf("expected user-123, got %s", claims.UserID)
    }
}
```

### 6.6 Middleware

**internal/middleware/auth.go**:
```go
package middleware

import (
    "strings"
    
    "link/internal/domain"
    "link/internal/pkg/token"
    
    "github.com/gofiber/fiber/v2"
)

func Auth(tm *token.Manager) fiber.Handler {
    return func(c *fiber.Ctx) error {
        auth := c.Get("Authorization")
        if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
            return c.Status(401).JSON(fiber.Map{
                "error": fiber.Map{"code": domain.ErrCodeUnauthorized, "message": "missing token"},
            })
        }
        
        tokenStr := strings.TrimPrefix(auth, "Bearer ")
        claims, err := tm.Verify(tokenStr)
        if err != nil {
            return c.Status(401).JSON(fiber.Map{
                "error": fiber.Map{"code": domain.ErrCodeUnauthorized, "message": err.Error()},
            })
        }
        
        c.Locals("userID", claims.UserID)
        return c.Next()
    }
}
```

**internal/middleware/ratelimit.go**:
```go
package middleware

import (
    "sync"
    "time"
    
    "link/internal/domain"
    
    "github.com/gofiber/fiber/v2"
)

type RateLimiter struct {
    visitors map[string]*visitor
    mu       sync.RWMutex
    rate     int
    window   time.Duration
}

type visitor struct {
    count    int
    lastSeen time.Time
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
    rl := &RateLimiter{
        visitors: make(map[string]*visitor),
        rate:     rate,
        window:   window,
    }
    go rl.cleanup()
    return rl
}

func (rl *RateLimiter) cleanup() {
    for {
        time.Sleep(rl.window)
        rl.mu.Lock()
        for ip, v := range rl.visitors {
            if time.Since(v.lastSeen) > rl.window {
                delete(rl.visitors, ip)
            }
        }
        rl.mu.Unlock()
    }
}

func (rl *RateLimiter) Middleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        ip := c.IP()
        
        rl.mu.Lock()
        v, exists := rl.visitors[ip]
        if !exists || time.Since(v.lastSeen) > rl.window {
            rl.visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
            rl.mu.Unlock()
            return c.Next()
        }
        
        v.count++
        v.lastSeen = time.Now()
        
        if v.count > rl.rate {
            rl.mu.Unlock()
            appErr := domain.ErrRateLimited()
            return c.Status(appErr.Status).JSON(fiber.Map{
                "error": fiber.Map{"code": appErr.Code, "message": appErr.Message},
            })
        }
        rl.mu.Unlock()
        
        return c.Next()
    }
}
```

**internal/middleware/security.go**:
```go
package middleware

import "github.com/gofiber/fiber/v2"

func SecurityHeaders() fiber.Handler {
    return func(c *fiber.Ctx) error {
        c.Set("X-Content-Type-Options", "nosniff")
        c.Set("X-Frame-Options", "DENY")
        c.Set("X-XSS-Protection", "1; mode=block")
        c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Set("Content-Security-Policy", "default-src 'self'; connect-src 'self' wss: https:;")
        return c.Next()
    }
}
```

**internal/middleware/logger.go**:
```go
package middleware

import (
    "log/slog"
    "time"
    
    "github.com/gofiber/fiber/v2"
)

func Logger() fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()
        err := c.Next()
        slog.Info("request",
            "method", c.Method(),
            "path", c.Path(),
            "status", c.Response().StatusCode(),
            "duration", time.Since(start),
        )
        return err
    }
}
```

### 6.7 Handler

**internal/handler/response.go**:
```go
package handler

import (
    "link/internal/domain"
    "log/slog"
    
    "github.com/gofiber/fiber/v2"
)

func OK(c *fiber.Ctx, data interface{}) error {
    return c.JSON(fiber.Map{"data": data})
}

func Error(c *fiber.Ctx, err error) error {
    if appErr, ok := domain.IsAppError(err); ok {
        return c.Status(appErr.Status).JSON(fiber.Map{
            "error": fiber.Map{"code": appErr.Code, "message": appErr.Message},
        })
    }
    slog.Error("unhandled error", "err", err)
    return c.Status(500).JSON(fiber.Map{
        "error": fiber.Map{"code": domain.ErrCodeInternal, "message": "系統錯誤"},
    })
}
```

**internal/handler/routes.go**:
```go
package handler

import (
    "time"
    
    "link/internal/middleware"
    
    "github.com/gofiber/fiber/v2"
)

type Handlers struct {
    Auth   *AuthHandler
    User   *UserHandler
    Friend *FriendHandler
    Conv   *ConversationHandler
}

func Setup(app *fiber.App, h *Handlers, authMw fiber.Handler) {
    app.Get("/health", func(c *fiber.Ctx) error { return c.SendString("OK") })
    
    api := app.Group("/api/v1")
    
    // Rate limiters
    loginLimiter := middleware.NewRateLimiter(10, time.Minute)
    registerLimiter := middleware.NewRateLimiter(5, time.Hour)

    // 公開 - 雙卡認證
    api.Get("/auth/check-card/:token", h.Auth.CheckCard)      // 檢查卡片狀態
    api.Post("/auth/pair/start", h.Auth.StartPair)            // 開始配對（掃主卡）
    api.Post("/auth/pair/complete", h.Auth.CompletePair)      // 完成配對（掃附卡）
    api.Post("/auth/register", registerLimiter.Middleware(), h.Auth.Register)
    api.Post("/auth/login", loginLimiter.Middleware(), h.Auth.Login)
    api.Post("/auth/login/backup", loginLimiter.Middleware(), h.Auth.LoginWithBackup) // 附卡登入（撤銷主卡）
    app.Get("/w/:token", h.Auth.CardEntry)

    // 需認證
    auth := api.Group("", authMw)
    auth.Get("/users/me", h.User.GetMe)
    auth.Get("/users/me/cards", h.User.GetMyCards)            // 查看我的卡片狀態
    auth.Patch("/users/me", h.User.UpdateMe)
    auth.Get("/users/search", h.User.Search)
    auth.Get("/users/:id/public-key", h.User.GetPublicKey)

    auth.Get("/friends", h.Friend.List)
    auth.Get("/friends/requests", h.Friend.Requests)
    auth.Post("/friends/request", h.Friend.SendRequest)
    auth.Post("/friends/:id/accept", h.Friend.Accept)
    auth.Post("/friends/:id/reject", h.Friend.Reject)
    auth.Delete("/friends/:id", h.Friend.Remove)

    auth.Get("/conversations", h.Conv.List)
    auth.Get("/conversations/:id/messages", h.Conv.Messages)
    
    auth.Post("/auth/logout", h.Auth.Logout)                  // 登出當前 session
}
```

### 6.8.1 Auth Handler (雙卡版)

**internal/handler/auth.go**:
```go
package handler

import (
    "link/internal/domain"
    "link/internal/service"
    
    "github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
    authSvc *service.AuthService
    cardSvc *service.CardService
}

func NewAuthHandler(authSvc *service.AuthService, cardSvc *service.CardService) *AuthHandler {
    return &AuthHandler{authSvc: authSvc, cardSvc: cardSvc}
}

// GET /auth/check-card/:token
func (h *AuthHandler) CheckCard(c *fiber.Ctx) error {
    token := c.Params("token")
    result, err := h.cardSvc.CheckCard(c.Context(), token)
    if err != nil {
        return Error(c, err)
    }
    return OK(c, result)
}

// POST /auth/pair/start - 掃描主卡開始配對
func (h *AuthHandler) StartPair(c *fiber.Ctx) error {
    var req struct {
        PrimaryToken string `json:"primary_token"`
    }
    if err := c.BodyParser(&req); err != nil {
        return Error(c, domain.ErrValidation("invalid request"))
    }
    
    pair, err := h.cardSvc.StartPair(c.Context(), req.PrimaryToken)
    if err != nil {
        return Error(c, err)
    }
    return OK(c, fiber.Map{"pair_id": pair.ID, "message": "請掃描附卡完成配對"})
}

// POST /auth/pair/complete - 掃描附卡完成配對
func (h *AuthHandler) CompletePair(c *fiber.Ctx) error {
    var req struct {
        PrimaryToken string `json:"primary_token"`
        BackupToken  string `json:"backup_token"`
    }
    if err := c.BodyParser(&req); err != nil {
        return Error(c, domain.ErrValidation("invalid request"))
    }
    
    err := h.cardSvc.CompletePair(c.Context(), req.PrimaryToken, req.BackupToken)
    if err != nil {
        return Error(c, err)
    }
    return OK(c, fiber.Map{"message": "配對完成，可以註冊"})
}

// POST /auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
    var req struct {
        PrimaryToken string `json:"primary_token"`
        BackupToken  string `json:"backup_token"`
        Password     string `json:"password"`
        Nickname     string `json:"nickname"`
        PublicKey    string `json:"public_key"`
    }
    if err := c.BodyParser(&req); err != nil {
        return Error(c, domain.ErrValidation("invalid request"))
    }
    
    res, err := h.authSvc.Register(c.Context(), service.RegisterInput{
        PrimaryToken: req.PrimaryToken,
        BackupToken:  req.BackupToken,
        Password:     req.Password,
        Nickname:     req.Nickname,
        PublicKey:    req.PublicKey,
    })
    if err != nil {
        return Error(c, err)
    }
    return OK(c, res)
}

// POST /auth/login - 主卡登入
func (h *AuthHandler) Login(c *fiber.Ctx) error {
    var req struct {
        CardToken string `json:"card_token"`
        Password  string `json:"password"`
    }
    if err := c.BodyParser(&req); err != nil {
        return Error(c, domain.ErrValidation("invalid request"))
    }
    
    res, err := h.authSvc.Login(c.Context(), req.CardToken, req.Password)
    if err != nil {
        return Error(c, err)
    }
    return OK(c, res)
}

// POST /auth/login/backup - 附卡登入（撤銷主卡）
func (h *AuthHandler) LoginWithBackup(c *fiber.Ctx) error {
    var req struct {
        CardToken string `json:"card_token"`
        Password  string `json:"password"`
        Confirm   bool   `json:"confirm"` // 必須確認撤銷
    }
    if err := c.BodyParser(&req); err != nil {
        return Error(c, domain.ErrValidation("invalid request"))
    }
    
    if !req.Confirm {
        return Error(c, domain.ErrValidation("必須確認撤銷主卡"))
    }
    
    res, err := h.authSvc.LoginWithBackupCard(c.Context(), req.CardToken, req.Password)
    if err != nil {
        return Error(c, err)
    }
    return OK(c, res)
}

// GET /w/:token - NFC 卡片入口
func (h *AuthHandler) CardEntry(c *fiber.Ctx) error {
    token := c.Params("token")
    result, _ := h.cardSvc.CheckCard(c.Context(), token)
    
    // 根據狀態重定向到前端對應頁面
    frontendURL := "https://localhost:5173"
    switch result.Status {
    case "not_found":
        return c.Redirect(frontendURL + "/register/start?token=" + token)
    case "pair_started", "pair_waiting":
        return c.Redirect(frontendURL + "/register/pair?token=" + token)
    case "primary":
        return c.Redirect(frontendURL + "/login?token=" + token)
    case "backup":
        return c.Redirect(frontendURL + "/login/backup?token=" + token)
    case "revoked":
        return c.Redirect(frontendURL + "/error?reason=card_revoked")
    default:
        return c.Redirect(frontendURL + "/error")
    }
}

// POST /auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    // TODO: 從 header 取得 token 並撤銷該 session
    _ = userID
    return OK(c, fiber.Map{"message": "已登出"})
}
```

### 6.8 WebTransport Protocol

**internal/transport/protocol.go**:
```go
package transport

const (
    TypeMessage   = "msg"
    TypeDelivered = "delivered"
    TypeTyping    = "typing"
    TypeRead      = "read"
    TypeOnline    = "online"
    TypeOffline   = "offline"
    TypeError     = "error"
)

type Message struct {
    Type    string      `json:"t"`
    Payload interface{} `json:"p,omitempty"`
}
```

**internal/transport/hub.go**:
```go
package transport

import (
    "log/slog"
    "sync"
)

// Client 介面（支援 WebTransport 和 WebSocket）
type Client interface {
    GetUserID() string
    SendStream(msg *Message) bool
    SendDatagram(msg *Message) bool
    Close()
}

type Hub struct {
    clients    map[string]Client
    mu         sync.RWMutex
    register   chan Client
    unregister chan Client
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[string]Client),
        register:   make(chan Client, 256),
        unregister: make(chan Client, 256),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case c := <-h.register:
            h.mu.Lock()
            if old, ok := h.clients[c.GetUserID()]; ok {
                old.Close()
            }
            h.clients[c.GetUserID()] = c
            h.mu.Unlock()
            slog.Info("client connected", "user_id", c.GetUserID())
            
        case c := <-h.unregister:
            h.mu.Lock()
            if curr, ok := h.clients[c.GetUserID()]; ok && curr == c {
                delete(h.clients, c.GetUserID())
            }
            h.mu.Unlock()
            slog.Info("client disconnected", "user_id", c.GetUserID())
        }
    }
}

func (h *Hub) Send(userID string, msg *Message) bool {
    h.mu.RLock()
    c, ok := h.clients[userID]
    h.mu.RUnlock()
    if !ok {
        return false
    }
    return c.SendStream(msg)
}

func (h *Hub) SendDatagram(userID string, msg *Message) bool {
    h.mu.RLock()
    c, ok := h.clients[userID]
    h.mu.RUnlock()
    if !ok {
        return false
    }
    return c.SendDatagram(msg)
}

func (h *Hub) IsOnline(userID string) bool {
    h.mu.RLock()
    defer h.mu.RUnlock()
    _, ok := h.clients[userID]
    return ok
}

func (h *Hub) Register(c Client)   { h.register <- c }
func (h *Hub) Unregister(c Client) { h.unregister <- c }
```

### 6.9 WebTransport Client (Go)

**internal/transport/client.go**:
```go
package transport

import (
    "context"
    "encoding/json"
    "io"
    "github.com/quic-go/webtransport-go"
)

type WTClient struct {
    userID  string
    session *webtransport.Session
    hub     *Hub
    handler *Handler
}

func NewWTClient(userID string, session *webtransport.Session, hub *Hub, handler *Handler) *WTClient {
    return &WTClient{userID: userID, session: session, hub: hub, handler: handler}
}

func (c *WTClient) GetUserID() string { return c.userID }

func (c *WTClient) Run(ctx context.Context) {
    go c.readDatagrams(ctx)
    go c.readStreams(ctx)
    <-ctx.Done()
    c.hub.Unregister(c)
}

func (c *WTClient) readDatagrams(ctx context.Context) {
    for {
        data, err := c.session.ReceiveDatagram(ctx)
        if err != nil {
            return
        }
        var msg struct {
            Type    string          `json:"t"`
            Payload json.RawMessage `json:"p"`
        }
        if json.Unmarshal(data, &msg) != nil {
            continue
        }
        switch msg.Type {
        case TypeTyping:
            var p struct {
                To             string `json:"to"`
                ConversationID string `json:"conversation_id"`
            }
            if json.Unmarshal(msg.Payload, &p) == nil {
                c.hub.SendDatagram(p.To, &Message{
                    Type:    TypeTyping,
                    Payload: map[string]string{"from": c.userID, "conversation_id": p.ConversationID},
                })
            }
        }
    }
}

func (c *WTClient) readStreams(ctx context.Context) {
    for {
        stream, err := c.session.AcceptStream(ctx)
        if err != nil {
            return
        }
        go c.handleStream(ctx, stream)
    }
}

func (c *WTClient) handleStream(ctx context.Context, stream webtransport.Stream) {
    defer stream.Close()
    data, err := io.ReadAll(stream)
    if err != nil {
        return
    }
    var msg struct {
        Type    string          `json:"t"`
        Payload json.RawMessage `json:"p"`
    }
    if json.Unmarshal(data, &msg) != nil {
        return
    }
    switch msg.Type {
    case TypeMessage:
        c.handler.HandleMessage(ctx, c.userID, msg.Payload)
    case TypeRead:
        c.handler.HandleRead(ctx, c.userID, msg.Payload)
    }
}

func (c *WTClient) SendStream(msg *Message) bool {
    stream, err := c.session.OpenStreamSync(context.Background())
    if err != nil {
        return false
    }
    defer stream.Close()
    data, _ := json.Marshal(msg)
    _, err = stream.Write(append(data, '\n'))
    return err == nil
}

func (c *WTClient) SendDatagram(msg *Message) bool {
    data, _ := json.Marshal(msg)
    return c.session.SendDatagram(data) == nil
}

func (c *WTClient) Close() {
    c.session.CloseWithError(0, "replaced")
}
```

### 6.10 WebTransport Server

**internal/transport/server.go**:
```go
package transport

import (
    "context"
    "log/slog"
    "net/http"
    
    "link/internal/pkg/token"
    "github.com/quic-go/quic-go/http3"
    "github.com/quic-go/webtransport-go"
)

type Server struct {
    wtServer *webtransport.Server
    hub      *Hub
    handler  *Handler
    tokenMgr *token.Manager
    certFile string
    keyFile  string
}

func NewServer(certFile, keyFile string, hub *Hub, handler *Handler, tm *token.Manager) *Server {
    return &Server{
        certFile: certFile,
        keyFile:  keyFile,
        hub:      hub,
        handler:  handler,
        tokenMgr: tm,
    }
}

func (s *Server) ListenAndServe(addr string) error {
    s.wtServer = &webtransport.Server{
        H3: http3.Server{Addr: addr},
    }
    http.HandleFunc("/wt", s.handleWT)
    slog.Info("WebTransport server starting", "addr", addr)
    return s.wtServer.ListenAndServeTLS(s.certFile, s.keyFile)
}

func (s *Server) handleWT(w http.ResponseWriter, r *http.Request) {
    tokenStr := r.URL.Query().Get("token")
    claims, err := s.tokenMgr.Verify(tokenStr)
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    session, err := s.wtServer.Upgrade(w, r)
    if err != nil {
        slog.Error("WebTransport upgrade failed", "err", err)
        return
    }
    client := NewWTClient(claims.UserID, session, s.hub, s.handler)
    s.hub.Register(client)
    ctx, cancel := context.WithCancel(r.Context())
    defer cancel()
    client.Run(ctx)
}

func (s *Server) Shutdown(ctx context.Context) error {
    if s.wtServer != nil {
        return s.wtServer.Close()
    }
    return nil
}
```

### 6.11 Transport Handler

**internal/transport/handler.go**:
```go
package transport

import (
    "context"
    "encoding/json"
    "log/slog"
    
    "link/internal/service"
)

type Handler struct {
    hub    *Hub
    msgSvc *service.MessageService
}

func NewHandler(hub *Hub, msgSvc *service.MessageService) *Handler {
    return &Handler{hub: hub, msgSvc: msgSvc}
}

func (h *Handler) HandleMessage(ctx context.Context, senderID string, payload json.RawMessage) {
    var p struct {
        To               string `json:"to"`
        EncryptedContent string `json:"encrypted_content"`
        TempID           string `json:"temp_id"`
    }
    if err := json.Unmarshal(payload, &p); err != nil {
        return
    }

    msg, err := h.msgSvc.Send(ctx, senderID, p.To, p.EncryptedContent)
    if err != nil {
        slog.Error("failed to save message", "err", err)
        h.hub.Send(senderID, &Message{Type: TypeError, Payload: map[string]string{"message": "發送失敗"}})
        return
    }

    h.hub.Send(p.To, &Message{Type: TypeMessage, Payload: msg})
    h.hub.Send(senderID, &Message{Type: TypeDelivered, Payload: map[string]interface{}{
        "temp_id": p.TempID,
        "message": msg,
    }})
}

func (h *Handler) HandleTyping(ctx context.Context, senderID string, payload json.RawMessage) {
    var p struct {
        To             string `json:"to"`
        ConversationID string `json:"conversation_id"`
    }
    if err := json.Unmarshal(payload, &p); err != nil {
        return
    }
    
    h.hub.SendDatagram(p.To, &Message{
        Type:    TypeTyping,
        Payload: map[string]string{"from": senderID, "conversation_id": p.ConversationID},
    })
}

func (h *Handler) HandleRead(ctx context.Context, userID string, payload json.RawMessage) {
    var p struct {
        ConversationID string `json:"conversation_id"`
        MessageID      string `json:"message_id"`
    }
    if err := json.Unmarshal(payload, &p); err != nil {
        return
    }
    h.msgSvc.MarkAsRead(ctx, p.MessageID)
}
```

### 6.9 WebSocket Client (Fallback)

**internal/transport/ws_client.go**:
```go
package transport

import (
    "context"
    "encoding/json"
    
    "github.com/gofiber/contrib/websocket"
)

type WSClient struct {
    userID  string
    conn    *websocket.Conn
    hub     *Hub
    handler *Handler
}

func NewWSClient(userID string, conn *websocket.Conn, hub *Hub, handler *Handler) *WSClient {
    return &WSClient{userID: userID, conn: conn, hub: hub, handler: handler}
}

func (c *WSClient) GetUserID() string { return c.userID }

func (c *WSClient) Run() {
    defer func() {
        c.hub.Unregister(c)
        c.conn.Close()
    }()
    
    for {
        _, data, err := c.conn.ReadMessage()
        if err != nil {
            return
        }
        
        var msg struct {
            Type    string          `json:"t"`
            Payload json.RawMessage `json:"p"`
        }
        if json.Unmarshal(data, &msg) != nil {
            continue
        }
        
        ctx := context.Background()
        switch msg.Type {
        case TypeMessage:
            c.handler.HandleMessage(ctx, c.userID, msg.Payload)
        case TypeTyping:
            c.handler.HandleTyping(ctx, c.userID, msg.Payload)
        case TypeRead:
            c.handler.HandleRead(ctx, c.userID, msg.Payload)
        }
    }
}

func (c *WSClient) SendStream(msg *Message) bool {
    data, _ := json.Marshal(msg)
    return c.conn.WriteMessage(websocket.TextMessage, data) == nil
}

func (c *WSClient) SendDatagram(msg *Message) bool {
    return c.SendStream(msg) // WS 無 datagram
}

func (c *WSClient) Close() {
    c.conn.Close()
}
```

### 6.10 Main (完整 DI + 連線池優化)

**cmd/server/main.go**:
```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/contrib/websocket"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/joho/godotenv"

    "link/internal/config"
    "link/internal/handler"
    "link/internal/middleware"
    "link/internal/repository/postgres"
    "link/internal/service"
    "link/internal/transport"
    "link/internal/pkg/token"
)

func main() {
    godotenv.Load()
    cfg := config.Load()

    slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

    // Database with optimized pool
    pool, err := createPool(context.Background(), cfg.DatabaseURL)
    if err != nil {
        slog.Error("failed to connect db", "err", err)
        os.Exit(1)
    }
    defer pool.Close()

    // Repositories
    userRepo := postgres.NewUserRepository(pool)
    cardRepo := postgres.NewCardRepository(pool)
    sessionRepo := postgres.NewSessionRepository(pool)
    friendRepo := postgres.NewFriendshipRepository(pool)
    convRepo := postgres.NewConversationRepository(pool)
    msgRepo := postgres.NewMessageRepository(pool)

    // Services
    tokenMgr := token.NewManager(cfg.JWTSecret, cfg.JWTExpiry)
    authSvc := service.NewAuthService(userRepo, cardRepo, sessionRepo, tokenMgr)
    cardSvc := service.NewCardService(cardRepo, sessionRepo)
    userSvc := service.NewUserService(userRepo)
    friendSvc := service.NewFriendshipService(friendRepo, userRepo)
    msgSvc := service.NewMessageService(msgRepo, convRepo)

    // WebTransport Hub
    hub := transport.NewHub()
    go hub.Run()
    wtHandler := transport.NewHandler(hub, msgSvc)

    // HTTP Handlers
    handlers := &handler.Handlers{
        Auth:   handler.NewAuthHandler(authSvc, cardSvc),
        User:   handler.NewUserHandler(userSvc, userRepo, cardRepo),
        Friend: handler.NewFriendHandler(friendSvc),
        Conv:   handler.NewConversationHandler(convRepo, msgRepo),
    }

    // Fiber App
    app := fiber.New(fiber.Config{
        ErrorHandler: func(c *fiber.Ctx, err error) error {
            return handler.Error(c, err)
        },
    })

    // Middleware
    app.Use(middleware.Logger())
    app.Use(middleware.SecurityHeaders())
    app.Use(cors.New(cors.Config{
        AllowOrigins:     cfg.CORSOrigins,
        AllowMethods:     "GET,POST,PATCH,DELETE",
        AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
        AllowCredentials: true,
        MaxAge:           86400,
    }))

    authMw := middleware.Auth(tokenMgr)
    handler.Setup(app, handlers, authMw)

    // WebSocket Fallback
    app.Use("/ws", func(c *fiber.Ctx) error {
        if websocket.IsWebSocketUpgrade(c) {
            return c.Next()
        }
        return fiber.ErrUpgradeRequired
    })
    app.Get("/ws", websocket.New(func(c *websocket.Conn) {
        tokenStr := c.Query("token")
        claims, err := tokenMgr.Verify(tokenStr)
        if err != nil {
            c.Close()
            return
        }
        client := transport.NewWSClient(claims.UserID, c, hub, wtHandler)
        hub.Register(client)
        client.Run()
    }))

    // WebTransport Server
    wtServer := transport.NewServer(cfg.TLSCert, cfg.TLSKey, hub, wtHandler, tokenMgr)
    go wtServer.ListenAndServe(cfg.ServerAddr)

    // Graceful Shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        slog.Info("server starting", "addr", cfg.ServerAddr)
        if err := app.ListenTLS(cfg.ServerAddr, cfg.TLSCert, cfg.TLSKey); err != nil {
            slog.Error("server error", "err", err)
        }
    }()

    <-quit
    slog.Info("shutting down...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    wtServer.Shutdown(ctx)
    app.ShutdownWithContext(ctx)

    slog.Info("server stopped")
}

func createPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    if err != nil {
        return nil, err
    }
    
    // 🔴 連線池優化配置
    config.MaxConns = 20
    config.MinConns = 5
    config.MaxConnLifetime = 30 * time.Minute
    config.MaxConnIdleTime = 5 * time.Minute
    config.MaxConnLifetimeJitter = 5 * time.Minute
    config.HealthCheckPeriod = 30 * time.Second
    config.ConnConfig.ConnectTimeout = 5 * time.Second
    
    return pgxpool.NewWithConfig(ctx, config)
}
```

### 6.11 Makefile

**backend/Makefile**:
```makefile
.PHONY: dev run build test lint migrate-up migrate-down

dev:
	air

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test -race ./internal/...

test-coverage:
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

migrate-up:
	psql "$(DATABASE_URL)" -f migrations/001_init.up.sql

migrate-down:
	psql "$(DATABASE_URL)" -f migrations/001_init.down.sql
```

### 6.12 Air Config

**backend/.air.toml**:
```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/server"
bin = "tmp/main"
include_ext = ["go"]
exclude_dir = ["tmp", "vendor"]
delay = 1000

[log]
time = false

[color]
main = "magenta"
watcher = "cyan"
build = "yellow"
runner = "green"
```

---

## 7. 前端實作

### 7.1 SvelteKit 配置 (SPA 模式)

**svelte.config.js**:
```javascript
import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

export default {
    preprocess: vitePreprocess(),
    kit: {
        adapter: adapter({
            fallback: 'index.html'  // SPA 模式
        }),
        prerender: {
            entries: []  // 不預渲染
        }
    }
};
```

### 7.2 Types

**src/lib/types.ts**:
```typescript
export interface User {
    id: string;
    nickname: string;
    public_key: string;
    avatar_url?: string;
    online?: boolean;
}

export interface AuthResponse {
    user: User;
    token: string;
}

export interface Friend {
    id: string;
    friend: User;
    status: 'pending' | 'accepted';
    created_at: string;
}

export interface Conversation {
    id: string;
    peer: User;
    last_message_at?: string;
    unread_count: number;
}

export interface EncryptedMessage {
    id: string;
    conversation_id: string;
    sender_id: string;
    encrypted_content: string;
    created_at: string;
    delivered_at?: string;
    read_at?: string;
}

export interface DecryptedMessage {
    id: string;
    conversation_id: string;
    sender_id: string;
    content: string;
    created_at: string;
    pending?: boolean;
}

export interface EncryptedData {
    nonce: string;
    ciphertext: string;
}

export interface WTMessage {
    t: string;
    p?: unknown;
}

export interface ApiError {
    code: string;
    message: string;
}

export class ApiException extends Error {
    constructor(public error: ApiError, public status: number) {
        super(error.message);
    }
}

export interface ITransport {
    connect(): Promise<void>;
    disconnect(): void;
    sendMessage(to: string, encryptedContent: string, tempId: string): Promise<void>;
    sendTyping(to: string, conversationId: string): Promise<void>;
    onMessage: ((msg: EncryptedMessage) => void) | null;
    onTyping: ((convId: string, userId: string) => void) | null;
    onOnline: ((userId: string) => void) | null;
    onOffline: ((userId: string) => void) | null;
    onDelivered: ((tempId: string, msg: EncryptedMessage) => void) | null;
    onConnected: ((connected: boolean) => void) | null;
}
```

### 7.3 API Client

**src/lib/api/client.ts**:
```typescript
import { auth } from '$lib/stores/auth.svelte';
import { ApiException, type ApiError } from '$lib/types';

const BASE = import.meta.env.VITE_API_URL;

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: HeadersInit = { 'Content-Type': 'application/json', ...options.headers };
    const token = auth.token;
    if (token) headers['Authorization'] = `Bearer ${token}`;

    const res = await fetch(`${BASE}${path}`, { ...options, headers });
    const body = await res.json();

    if (!res.ok) {
        if (res.status === 401) auth.logout();
        throw new ApiException(body.error as ApiError, res.status);
    }
    return body.data as T;
}

export const api = {
    get: <T>(path: string) => request<T>(path),
    post: <T>(path: string, data?: unknown) => request<T>(path, { method: 'POST', body: data ? JSON.stringify(data) : undefined }),
    patch: <T>(path: string, data: unknown) => request<T>(path, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
};
```

**src/lib/api/auth.ts**:
```typescript
import { api } from './client';
import type { AuthResponse } from '$lib/types';

export interface CardCheckResult {
    status: 'not_found' | 'pair_started' | 'pair_waiting' | 'primary' | 'backup' | 'revoked';
    user_id?: string;
    nickname?: string;
    card_type?: 'primary' | 'backup';
    pair_id?: string;
    warning?: string;
}

export const authApi = {
    // 檢查卡片狀態
    checkCard: (token: string) => 
        api.get<CardCheckResult>(`/api/v1/auth/check-card/${token}`),
    
    // 開始配對（掃主卡）
    startPair: (primaryToken: string) => 
        api.post<{ pair_id: string; message: string }>('/api/v1/auth/pair/start', { primary_token: primaryToken }),
    
    // 完成配對（掃附卡）
    completePair: (primaryToken: string, backupToken: string) => 
        api.post<{ message: string }>('/api/v1/auth/pair/complete', { 
            primary_token: primaryToken, 
            backup_token: backupToken 
        }),
    
    // 註冊（需要雙卡）
    register: (data: { 
        primary_token: string; 
        backup_token: string; 
        password: string; 
        nickname: string; 
        public_key: string 
    }) => api.post<AuthResponse>('/api/v1/auth/register', data),
    
    // 主卡登入
    login: (data: { card_token: string; password: string }) =>
        api.post<AuthResponse>('/api/v1/auth/login', data),
    
    // 附卡登入（撤銷主卡）
    loginWithBackup: (data: { card_token: string; password: string; confirm: boolean }) =>
        api.post<AuthResponse>('/api/v1/auth/login/backup', data),
    
    // 登出
    logout: () => api.post<{ message: string }>('/api/v1/auth/logout'),
};
```

**src/lib/api/users.ts**:
```typescript
import { api } from './client';
import type { User } from '$lib/types';

export const usersApi = {
    getMe: () => api.get<User>('/api/v1/users/me'),
    updateMe: (data: { nickname?: string; avatar_url?: string }) => api.patch<User>('/api/v1/users/me', data),
    search: (q: string) => api.get<User[]>(`/api/v1/users/search?q=${encodeURIComponent(q)}`),
    getPublicKey: (id: string) => api.get<{ public_key: string }>(`/api/v1/users/${id}/public-key`),
};
```

**src/lib/api/friends.ts**:
```typescript
import { api } from './client';
import type { Friend } from '$lib/types';

export const friendsApi = {
    list: () => api.get<Friend[]>('/api/v1/friends'),
    requests: () => api.get<Friend[]>('/api/v1/friends/requests'),
    sendRequest: (userId: string) => api.post<Friend>('/api/v1/friends/request', { user_id: userId }),
    accept: (id: string) => api.post<Friend>(`/api/v1/friends/${id}/accept`),
    reject: (id: string) => api.post<void>(`/api/v1/friends/${id}/reject`),
    remove: (id: string) => api.delete<void>(`/api/v1/friends/${id}`),
};
```

**src/lib/api/conversations.ts**:
```typescript
import { api } from './client';
import type { Conversation, EncryptedMessage } from '$lib/types';

export const conversationsApi = {
    list: () => api.get<Conversation[]>('/api/v1/conversations'),
    messages: (id: string, limit = 50, before?: string) => {
        let url = `/api/v1/conversations/${id}/messages?limit=${limit}`;
        if (before) url += `&before=${before}`;
        return api.get<EncryptedMessage[]>(url);
    },
};
```

### 7.4 Crypto (含 Padding)

**src/lib/crypto/keys.ts**:
```typescript
import nacl from 'tweetnacl';
import { encodeBase64, decodeBase64 } from 'tweetnacl-util';

const DB_NAME = 'link-keys';
const STORE_NAME = 'keypair';

export function generateKeyPair(): { publicKey: string; secretKey: Uint8Array } {
    const kp = nacl.box.keyPair();
    return { publicKey: encodeBase64(kp.publicKey), secretKey: kp.secretKey };
}

export async function saveSecretKey(secretKey: Uint8Array, password: string): Promise<void> {
    const salt = nacl.randomBytes(16);
    const key = await deriveKey(password, salt);
    const nonce = nacl.randomBytes(24);
    const encrypted = nacl.secretbox(secretKey, nonce, key);
    const data = { salt: encodeBase64(salt), nonce: encodeBase64(nonce), encrypted: encodeBase64(encrypted) };
    const db = await openDB();
    await putToDB(db, 'secretKey', JSON.stringify(data));
}

export async function loadSecretKey(password: string): Promise<Uint8Array | null> {
    const db = await openDB();
    const stored = await getFromDB(db, 'secretKey');
    if (!stored) return null;
    const data = JSON.parse(stored);
    const key = await deriveKey(password, decodeBase64(data.salt));
    const decrypted = nacl.secretbox.open(decodeBase64(data.encrypted), decodeBase64(data.nonce), key);
    return decrypted || null;
}

export async function hasSecretKey(): Promise<boolean> {
    const db = await openDB();
    return (await getFromDB(db, 'secretKey')) !== null;
}

export async function clearSecretKey(): Promise<void> {
    const db = await openDB();
    await deleteFromDB(db, 'secretKey');
}

async function deriveKey(password: string, salt: Uint8Array): Promise<Uint8Array> {
    const keyMaterial = await crypto.subtle.importKey('raw', new TextEncoder().encode(password), 'PBKDF2', false, ['deriveBits']);
    const bits = await crypto.subtle.deriveBits({ name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' }, keyMaterial, 256);
    return new Uint8Array(bits);
}

function openDB(): Promise<IDBDatabase> {
    return new Promise((resolve, reject) => {
        const req = indexedDB.open(DB_NAME, 1);
        req.onerror = () => reject(req.error);
        req.onsuccess = () => resolve(req.result);
        req.onupgradeneeded = () => req.result.createObjectStore(STORE_NAME);
    });
}

function putToDB(db: IDBDatabase, key: string, value: string): Promise<void> {
    return new Promise((resolve, reject) => {
        const tx = db.transaction(STORE_NAME, 'readwrite');
        tx.objectStore(STORE_NAME).put(value, key);
        tx.oncomplete = () => resolve();
        tx.onerror = () => reject(tx.error);
    });
}

function getFromDB(db: IDBDatabase, key: string): Promise<string | null> {
    return new Promise((resolve, reject) => {
        const tx = db.transaction(STORE_NAME, 'readonly');
        const req = tx.objectStore(STORE_NAME).get(key);
        req.onsuccess = () => resolve(req.result || null);
        req.onerror = () => reject(req.error);
    });
}

function deleteFromDB(db: IDBDatabase, key: string): Promise<void> {
    return new Promise((resolve, reject) => {
        const tx = db.transaction(STORE_NAME, 'readwrite');
        tx.objectStore(STORE_NAME).delete(key);
        tx.oncomplete = () => resolve();
        tx.onerror = () => reject(tx.error);
    });
}
```

**src/lib/crypto/encrypt.ts**:
```typescript
import nacl from 'tweetnacl';
import { encodeBase64, decodeBase64 } from 'tweetnacl-util';
import type { EncryptedData } from '$lib/types';

// 🔴 Padding 配置（防止長度分析攻擊）
const MIN_PADDED_LENGTH = 256;
const PADDING_BLOCK_SIZE = 64;

function padMessage(message: string): Uint8Array {
    const msgBytes = new TextEncoder().encode(message);
    const msgLen = msgBytes.length;
    
    // 計算填充後長度
    let paddedLen = Math.max(MIN_PADDED_LENGTH, msgLen + 4);
    paddedLen = Math.ceil(paddedLen / PADDING_BLOCK_SIZE) * PADDING_BLOCK_SIZE;
    
    const padded = new Uint8Array(paddedLen);
    
    // 前 4 bytes 存原始長度（big-endian）
    const view = new DataView(padded.buffer);
    view.setUint32(0, msgLen, false);
    
    // 複製原始訊息
    padded.set(msgBytes, 4);
    
    // 剩餘部分填充隨機數據
    const randomPadding = nacl.randomBytes(paddedLen - 4 - msgLen);
    padded.set(randomPadding, 4 + msgLen);
    
    return padded;
}

export function encryptMessage(
    message: string, 
    theirPublicKey: string, 
    mySecretKey: Uint8Array
): EncryptedData {
    const nonce = nacl.randomBytes(24);
    const paddedMsg = padMessage(message);
    const ciphertext = nacl.box(paddedMsg, nonce, decodeBase64(theirPublicKey), mySecretKey);
    
    return {
        nonce: encodeBase64(nonce),
        ciphertext: encodeBase64(ciphertext)
    };
}

export function encryptToString(
    message: string, 
    theirPublicKey: string, 
    mySecretKey: Uint8Array
): string {
    return JSON.stringify(encryptMessage(message, theirPublicKey, mySecretKey));
}
```

**src/lib/crypto/decrypt.ts**:
```typescript
import nacl from 'tweetnacl';
import { decodeBase64 } from 'tweetnacl-util';
import type { EncryptedData } from '$lib/types';

function unpadMessage(padded: Uint8Array): string {
    const view = new DataView(padded.buffer, padded.byteOffset, padded.byteLength);
    const msgLen = view.getUint32(0, false);
    
    if (msgLen > padded.length - 4) {
        throw new Error('Invalid padded message');
    }
    
    const msgBytes = padded.slice(4, 4 + msgLen);
    return new TextDecoder().decode(msgBytes);
}

export function decryptMessage(
    encrypted: EncryptedData, 
    theirPublicKey: string, 
    mySecretKey: Uint8Array
): string | null {
    try {
        const decrypted = nacl.box.open(
            decodeBase64(encrypted.ciphertext),
            decodeBase64(encrypted.nonce),
            decodeBase64(theirPublicKey),
            mySecretKey
        );
        
        if (!decrypted) return null;
        
        return unpadMessage(decrypted);
    } catch {
        return null;
    }
}

export function decryptFromString(
    encryptedContent: string, 
    theirPublicKey: string, 
    mySecretKey: Uint8Array
): string | null {
    try {
        return decryptMessage(JSON.parse(encryptedContent), theirPublicKey, mySecretKey);
    } catch {
        return null;
    }
}
```

**src/lib/crypto/index.ts**:
```typescript
export * from './keys';
export * from './encrypt';
export * from './decrypt';
```

### 7.5 Transport (自動降級)

**src/lib/transport/webtransport.ts**:
```typescript
import type { WTMessage, EncryptedMessage, ITransport } from '$lib/types';

export class WebTransportClient implements ITransport {
    private wt: WebTransport | null = null;
    private reconnectAttempts = 0;
    private maxReconnects = 5;

    public onMessage: ((msg: EncryptedMessage) => void) | null = null;
    public onTyping: ((convId: string, userId: string) => void) | null = null;
    public onOnline: ((userId: string) => void) | null = null;
    public onOffline: ((userId: string) => void) | null = null;
    public onDelivered: ((tempId: string, msg: EncryptedMessage) => void) | null = null;
    public onConnected: ((connected: boolean) => void) | null = null;

    constructor(private url: string, private token: string) {}

    async connect(): Promise<void> {
        this.wt = new WebTransport(`${this.url}?token=${this.token}`);
        await this.wt.ready;
        this.reconnectAttempts = 0;
        this.onConnected?.(true);
        this.receiveDatagrams();
        this.receiveStreams();
        this.wt.closed.then(() => {
            this.onConnected?.(false);
            this.scheduleReconnect();
        });
    }

    private async receiveDatagrams(): Promise<void> {
        if (!this.wt) return;
        const reader = this.wt.datagrams.readable.getReader();
        try {
            while (true) {
                const { value, done } = await reader.read();
                if (done) break;
                const msg: WTMessage = JSON.parse(new TextDecoder().decode(value));
                if (msg.t === 'typing') {
                    const p = msg.p as { from: string; conversation_id: string };
                    this.onTyping?.(p.conversation_id, p.from);
                } else if (msg.t === 'online') {
                    this.onOnline?.((msg.p as { user_id: string }).user_id);
                } else if (msg.t === 'offline') {
                    this.onOffline?.((msg.p as { user_id: string }).user_id);
                }
            }
        } catch {}
    }

    private async receiveStreams(): Promise<void> {
        if (!this.wt) return;
        const reader = this.wt.incomingBidirectionalStreams.getReader();
        try {
            while (true) {
                const { value: stream, done } = await reader.read();
                if (done) break;
                this.handleStream(stream);
            }
        } catch {}
    }

    private async handleStream(stream: WebTransportBidirectionalStream): Promise<void> {
        const reader = stream.readable.getReader();
        let buffer = '';
        try {
            while (true) {
                const { value, done } = await reader.read();
                if (done) break;
                buffer += new TextDecoder().decode(value);
                const lines = buffer.split('\n');
                buffer = lines.pop() || '';
                for (const line of lines) {
                    if (!line) continue;
                    const msg: WTMessage = JSON.parse(line);
                    if (msg.t === 'msg') this.onMessage?.(msg.p as EncryptedMessage);
                    else if (msg.t === 'delivered') {
                        const p = msg.p as { temp_id: string; message: EncryptedMessage };
                        this.onDelivered?.(p.temp_id, p.message);
                    }
                }
            }
        } catch {}
    }

    async sendMessage(to: string, encryptedContent: string, tempId: string): Promise<void> {
        if (!this.wt) throw new Error('Not connected');
        const stream = await this.wt.createBidirectionalStream();
        const writer = stream.writable.getWriter();
        const msg: WTMessage = { t: 'msg', p: { to, encrypted_content: encryptedContent, temp_id: tempId } };
        await writer.write(new TextEncoder().encode(JSON.stringify(msg) + '\n'));
        writer.releaseLock();
    }

    async sendTyping(to: string, conversationId: string): Promise<void> {
        if (!this.wt) return;
        const msg: WTMessage = { t: 'typing', p: { to, conversation_id: conversationId } };
        const writer = this.wt.datagrams.writable.getWriter();
        await writer.write(new TextEncoder().encode(JSON.stringify(msg)));
        writer.releaseLock();
    }

    private scheduleReconnect(): void {
        if (this.reconnectAttempts >= this.maxReconnects) return;
        const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
        this.reconnectAttempts++;
        setTimeout(() => this.connect(), delay);
    }

    disconnect(): void {
        this.reconnectAttempts = this.maxReconnects;
        this.wt?.close();
        this.wt = null;
    }
}
```

**src/lib/transport/websocket.ts**:
```typescript
import type { WTMessage, EncryptedMessage, ITransport } from '$lib/types';

export class WebSocketClient implements ITransport {
    private ws: WebSocket | null = null;
    private reconnectAttempts = 0;
    private maxReconnects = 5;

    public onMessage: ((msg: EncryptedMessage) => void) | null = null;
    public onTyping: ((convId: string, userId: string) => void) | null = null;
    public onOnline: ((userId: string) => void) | null = null;
    public onOffline: ((userId: string) => void) | null = null;
    public onDelivered: ((tempId: string, msg: EncryptedMessage) => void) | null = null;
    public onConnected: ((connected: boolean) => void) | null = null;

    constructor(private url: string, private token: string) {}

    async connect(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.ws = new WebSocket(`${this.url}?token=${this.token}`);
            
            this.ws.onopen = () => {
                this.reconnectAttempts = 0;
                this.onConnected?.(true);
                resolve();
            };
            
            this.ws.onerror = (e) => reject(e);
            
            this.ws.onclose = () => {
                this.onConnected?.(false);
                this.scheduleReconnect();
            };
            
            this.ws.onmessage = (e) => {
                const msg: WTMessage = JSON.parse(e.data);
                switch (msg.t) {
                    case 'msg':
                        this.onMessage?.(msg.p as EncryptedMessage);
                        break;
                    case 'typing':
                        const tp = msg.p as { from: string; conversation_id: string };
                        this.onTyping?.(tp.conversation_id, tp.from);
                        break;
                    case 'online':
                        this.onOnline?.((msg.p as { user_id: string }).user_id);
                        break;
                    case 'offline':
                        this.onOffline?.((msg.p as { user_id: string }).user_id);
                        break;
                    case 'delivered':
                        const dp = msg.p as { temp_id: string; message: EncryptedMessage };
                        this.onDelivered?.(dp.temp_id, dp.message);
                        break;
                }
            };
        });
    }

    async sendMessage(to: string, encryptedContent: string, tempId: string): Promise<void> {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) throw new Error('Not connected');
        this.ws.send(JSON.stringify({ t: 'msg', p: { to, encrypted_content: encryptedContent, temp_id: tempId } }));
    }

    async sendTyping(to: string, conversationId: string): Promise<void> {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
        this.ws.send(JSON.stringify({ t: 'typing', p: { to, conversation_id: conversationId } }));
    }

    private scheduleReconnect(): void {
        if (this.reconnectAttempts >= this.maxReconnects) return;
        const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
        this.reconnectAttempts++;
        setTimeout(() => this.connect(), delay);
    }

    disconnect(): void {
        this.reconnectAttempts = this.maxReconnects;
        this.ws?.close();
        this.ws = null;
    }
}
```

**src/lib/transport/index.ts**:
```typescript
import type { ITransport } from '$lib/types';
import { WebTransportClient } from './webtransport';
import { WebSocketClient } from './websocket';

export async function createTransport(wtUrl: string, wsUrl: string, token: string): Promise<ITransport> {
    // 先嘗試 WebTransport
    if ('WebTransport' in window) {
        try {
            const wt = new WebTransportClient(wtUrl, token);
            await wt.connect();
            console.log('Connected via WebTransport');
            return wt;
        } catch (e) {
            console.warn('WebTransport failed, falling back to WebSocket:', e);
        }
    }
    
    // Fallback to WebSocket
    const ws = new WebSocketClient(wsUrl, token);
    await ws.connect();
    console.log('Connected via WebSocket');
    return ws;
}

export { WebTransportClient } from './webtransport';
export { WebSocketClient } from './websocket';
```

### 7.6 Stores

**src/lib/stores/auth.svelte.ts**:
```typescript
import { browser } from '$app/environment';
import type { User, AuthResponse } from '$lib/types';

function createAuthStore() {
    let user = $state<User | null>(null);
    let token = $state<string | null>(null);
    let loading = $state(true);

    if (browser) {
        token = localStorage.getItem('token');
        const saved = localStorage.getItem('user');
        if (saved) user = JSON.parse(saved);
        loading = false;
    }

    return {
        get user() { return user; },
        get token() { return token; },
        get loading() { return loading; },
        get isAuthenticated() { return !!token; },

        login(res: AuthResponse) {
            user = res.user;
            token = res.token;
            if (browser) {
                localStorage.setItem('token', res.token);
                localStorage.setItem('user', JSON.stringify(res.user));
            }
        },

        logout() {
            user = null;
            token = null;
            if (browser) {
                localStorage.removeItem('token');
                localStorage.removeItem('user');
            }
        }
    };
}

export const auth = createAuthStore();
```

**src/lib/stores/keys.svelte.ts**:
```typescript
import { browser } from '$app/environment';
import { loadSecretKey, hasSecretKey } from '$lib/crypto';
import { usersApi } from '$lib/api/users';

function createKeysStore() {
    let secretKey = $state<Uint8Array | null>(null);
    let publicKeyCache = $state<Record<string, string>>({});
    let ready = $state(false);

    return {
        get secretKey() { return secretKey; },
        get ready() { return ready; },

        getPublicKey(userId: string): string | undefined {
            return publicKeyCache[userId];
        },

        setPublicKey(userId: string, publicKey: string) {
            publicKeyCache[userId] = publicKey;
        },

        async ensurePublicKey(userId: string): Promise<string> {
            let pk = publicKeyCache[userId];
            if (!pk) {
                const res = await usersApi.getPublicKey(userId);
                pk = res.public_key;
                publicKeyCache[userId] = pk;
            }
            return pk;
        },

        async load(password: string): Promise<boolean> {
            if (!browser) return false;
            secretKey = await loadSecretKey(password);
            ready = true;
            return secretKey !== null;
        },

        setSecretKey(key: Uint8Array) {
            secretKey = key;
            ready = true;
        },

        async hasKey(): Promise<boolean> {
            if (!browser) return false;
            return hasSecretKey();
        },

        clear() {
            secretKey = null;
            publicKeyCache = {};
            ready = false;
        }
    };
}

export const keys = createKeysStore();
```

**src/lib/stores/messages.svelte.ts**:
```typescript
import type { EncryptedMessage, DecryptedMessage } from '$lib/types';
import { decryptFromString } from '$lib/crypto';
import { keys } from './keys.svelte';

function createMessagesStore() {
    let byConv = $state<Record<string, DecryptedMessage[]>>({});
    let activeId = $state<string | null>(null);

    return {
        get active() { return activeId ? (byConv[activeId] || []) : []; },
        get activeConversationId() { return activeId; },

        setActive(id: string | null) { activeId = id; },

        async addEncrypted(msg: EncryptedMessage, senderPublicKey: string) {
            if (!keys.secretKey) return;
            const content = decryptFromString(msg.encrypted_content, senderPublicKey, keys.secretKey);
            if (!content) return;
            this.add({
                id: msg.id,
                conversation_id: msg.conversation_id,
                sender_id: msg.sender_id,
                content,
                created_at: msg.created_at
            });
        },

        add(msg: DecryptedMessage) {
            const cid = msg.conversation_id;
            if (!byConv[cid]) byConv[cid] = [];
            if (!byConv[cid].some(m => m.id === msg.id)) {
                byConv[cid] = [...byConv[cid], msg].sort(
                    (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
                );
            }
        },

        confirmPending(tempId: string, realMsg: DecryptedMessage) {
            const cid = realMsg.conversation_id;
            if (byConv[cid]) {
                byConv[cid] = byConv[cid].map(m => m.id === tempId ? realMsg : m);
            }
        },

        addPending(convId: string, content: string, senderId: string): string {
            const tempId = `temp-${Date.now()}`;
            this.add({ id: tempId, conversation_id: convId, sender_id: senderId, content, created_at: new Date().toISOString(), pending: true });
            return tempId;
        },

        setList(convId: string, msgs: DecryptedMessage[]) { byConv[convId] = msgs; },
        clear() { byConv = {}; activeId = null; }
    };
}

export const messages = createMessagesStore();
```

**src/lib/stores/conversations.svelte.ts**:
```typescript
import type { Conversation, EncryptedMessage } from '$lib/types';

function createConversationsStore() {
    let list = $state<Conversation[]>([]);
    let typingMap = $state<Record<string, string | null>>({});

    return {
        get list() { return list; },

        getTyping(convId: string): string | null {
            return typingMap[convId] || null;
        },

        setList(convs: Conversation[]) { list = convs; },

        setTyping(convId: string, userId: string | null) {
            typingMap[convId] = userId;
            if (userId) setTimeout(() => { typingMap[convId] = null; }, 3000);
        },

        updateLastMessage(msg: EncryptedMessage) {
            list = list.map(c => 
                c.id === msg.conversation_id ? { ...c, last_message_at: msg.created_at } : c
            ).sort((a, b) => 
                new Date(b.last_message_at || 0).getTime() - new Date(a.last_message_at || 0).getTime()
            );
        },

        clear() { list = []; typingMap = {}; }
    };
}

export const conversations = createConversationsStore();
```

**src/lib/stores/friends.svelte.ts**:
```typescript
import type { Friend } from '$lib/types';

function createFriendsStore() {
    let list = $state<Friend[]>([]);
    let requests = $state<Friend[]>([]);

    return {
        get list() { return list; },
        get requests() { return requests; },

        setList(friends: Friend[]) { list = friends; },
        setRequests(reqs: Friend[]) { requests = reqs; },

        setOnline(userId: string, online: boolean) {
            list = list.map(f => f.friend.id === userId ? { ...f, friend: { ...f.friend, online } } : f);
        },

        addFriend(friend: Friend) {
            if (!list.some(f => f.id === friend.id)) list = [...list, friend];
        },

        removeRequest(id: string) { requests = requests.filter(r => r.id !== id); },

        clear() { list = []; requests = []; }
    };
}

export const friends = createFriendsStore();
```

**src/lib/stores/transport.svelte.ts**:
```typescript
import type { EncryptedMessage, ITransport } from '$lib/types';
import { createTransport } from '$lib/transport';

function createTransportStore() {
    let client = $state<ITransport | null>(null);
    let connected = $state(false);

    return {
        get connected() { return connected; },

        onMessage: null as ((msg: EncryptedMessage) => void) | null,
        onTyping: null as ((convId: string, userId: string) => void) | null,
        onOnline: null as ((userId: string) => void) | null,
        onOffline: null as ((userId: string) => void) | null,
        onDelivered: null as ((tempId: string, msg: EncryptedMessage) => void) | null,

        async connect(wtUrl: string, wsUrl: string, token: string) {
            client = await createTransport(wtUrl, wsUrl, token);
            client.onConnected = (c) => { connected = c; };
            client.onMessage = (msg) => this.onMessage?.(msg);
            client.onTyping = (convId, userId) => this.onTyping?.(convId, userId);
            client.onOnline = (userId) => this.onOnline?.(userId);
            client.onOffline = (userId) => this.onOffline?.(userId);
            client.onDelivered = (tempId, msg) => this.onDelivered?.(tempId, msg);
        },

        async sendMessage(to: string, encryptedContent: string, tempId: string) {
            await client?.sendMessage(to, encryptedContent, tempId);
        },

        async sendTyping(to: string, convId: string) {
            await client?.sendTyping(to, convId);
        },

        disconnect() {
            client?.disconnect();
            client = null;
            connected = false;
        }
    };
}

export const transport = createTransportStore();
```

**src/lib/stores/index.ts**:
```typescript
import { auth } from './auth.svelte';
import { keys } from './keys.svelte';
import { messages } from './messages.svelte';
import { conversations } from './conversations.svelte';
import { friends } from './friends.svelte';
import { transport } from './transport.svelte';
import { decryptFromString } from '$lib/crypto';
import type { EncryptedMessage } from '$lib/types';

export function initStores() {
    transport.onMessage = async (msg: EncryptedMessage) => {
        const senderPubKey = await keys.ensurePublicKey(msg.sender_id);
        await messages.addEncrypted(msg, senderPubKey);
        conversations.updateLastMessage(msg);
    };

    transport.onTyping = (convId, userId) => conversations.setTyping(convId, userId);
    transport.onOnline = (userId) => friends.setOnline(userId, true);
    transport.onOffline = (userId) => friends.setOnline(userId, false);

    transport.onDelivered = async (tempId, msg) => {
        const myPubKey = auth.user?.public_key;
        if (myPubKey && keys.secretKey) {
            const content = decryptFromString(msg.encrypted_content, myPubKey, keys.secretKey);
            if (content) {
                messages.confirmPending(tempId, {
                    id: msg.id, conversation_id: msg.conversation_id,
                    sender_id: msg.sender_id, content, created_at: msg.created_at
                });
            }
        }
    };
}

export function clearAllStores() {
    messages.clear(); conversations.clear(); friends.clear();
    keys.clear(); transport.disconnect();
}

export { auth, keys, messages, conversations, friends, transport };
```

### 7.7 Security Warning Component

**src/lib/components/SecurityWarning.svelte**:
```svelte
<script lang="ts">
    let dismissed = $state(false);
    
    function dismiss() {
        dismissed = true;
        localStorage.setItem('security-warning-dismissed', 'true');
    }
    
    $effect(() => {
        dismissed = localStorage.getItem('security-warning-dismissed') === 'true';
    });
</script>

{#if !dismissed}
<div class="fixed bottom-4 right-4 max-w-sm bg-yellow-50 border border-yellow-200 rounded-lg p-4 shadow-lg z-50">
    <div class="flex items-start gap-3">
        <svg class="w-5 h-5 text-yellow-600 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
        </svg>
        <div class="flex-1">
            <h4 class="font-medium text-yellow-800">安全提醒</h4>
            <p class="text-sm text-yellow-700 mt-1">
                您的加密私鑰儲存在此瀏覽器中。請勿在公用電腦上使用，並確保瀏覽器擴展來自可信來源。
            </p>
            <button 
                onclick={dismiss}
                class="mt-2 text-sm text-yellow-800 underline hover:no-underline"
            >
                我了解風險
            </button>
        </div>
    </div>
</div>
{/if}
```

---

## 8. WebTransport + WebSocket 協議

### 8.1 傳輸模式

| 資料類型 | WebTransport | WebSocket |
|----------|--------------|-----------|
| 聊天訊息 | Stream | Message |
| 已讀回執 | Stream | Message |
| 打字中 | Datagram | Message |
| 在線狀態 | Datagram | Message |

### 8.2 訊息類型

**Client → Server**: msg, typing, read
**Server → Client**: msg, delivered, typing, online, offline, error

---

## 9. 測試策略

### 9.1 加密測試

**src/lib/crypto/crypto.test.ts**:
```typescript
import { describe, it, expect } from 'vitest';
import nacl from 'tweetnacl';
import { encodeBase64 } from 'tweetnacl-util';
import { encryptMessage, decryptMessage } from './index';

describe('E2EE with Padding', () => {
    it('encrypts and decrypts', () => {
        const alice = nacl.box.keyPair();
        const bob = nacl.box.keyPair();
        const encrypted = encryptMessage('Hello', encodeBase64(bob.publicKey), alice.secretKey);
        const decrypted = decryptMessage(encrypted, encodeBase64(alice.publicKey), bob.secretKey);
        expect(decrypted).toBe('Hello');
    });

    it('fails with wrong key', () => {
        const alice = nacl.box.keyPair();
        const bob = nacl.box.keyPair();
        const eve = nacl.box.keyPair();
        const encrypted = encryptMessage('Secret', encodeBase64(bob.publicKey), alice.secretKey);
        const decrypted = decryptMessage(encrypted, encodeBase64(alice.publicKey), eve.secretKey);
        expect(decrypted).toBeNull();
    });

    it('produces fixed-size ciphertext (padding)', () => {
        const alice = nacl.box.keyPair();
        const bob = nacl.box.keyPair();
        
        const short = encryptMessage('Hi', encodeBase64(bob.publicKey), alice.secretKey);
        const long = encryptMessage('Hello World!', encodeBase64(bob.publicKey), alice.secretKey);
        
        // 解碼後長度應該相同（都是 256 + overhead）
        const shortLen = atob(short.ciphertext).length;
        const longLen = atob(long.ciphertext).length;
        expect(shortLen).toBe(longLen);
    });
});
```

### 9.2 測試配置

**vitest.config.ts**:
```typescript
import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
    plugins: [svelte({ hot: !process.env.VITEST })],
    test: { include: ['src/**/*.{test,spec}.ts'], globals: true, environment: 'jsdom' }
});
```

---

## 10. Agent 任務分配

### Agent-1: 基礎建設 (15min)
```
□ 創建目錄結構
□ docker-compose.yml
□ scripts/setup-certs.sh
□ pnpm create svelte@latest frontend (選 SPA)
□ pnpm add tweetnacl tweetnacl-util
□ 創建 CLAUDE.md（複製 1.5 節）
□ svelte.config.js (SPA 模式)
□ docker compose up -d
✓ 驗收: postgres running
```

### Agent-2: 後端核心 + 雙卡 Domain (40min)
```
□ internal/config/config.go (含密鑰驗證)
□ internal/domain/errors.go
□ internal/domain/user.go
□ internal/domain/card.go          # 雙卡機制
□ internal/domain/session.go       # Session 管理
□ internal/domain/friendship.go
□ internal/domain/conversation.go
□ internal/domain/message.go
□ internal/pkg/password/argon2.go (OWASP 參數)
□ internal/pkg/token/jwt.go (算法白名單)
□ internal/pkg/circuitbreaker/breaker.go
□ migrations/*.sql (含 cards, sessions 表)
✓ 驗收: go test ./internal/domain/... 全過
```

### Agent-3: 後端 Repository (35min)
```
□ internal/repository/postgres/user.go
□ internal/repository/postgres/card.go      # 雙卡 Repo
□ internal/repository/postgres/session.go   # Session Repo
□ internal/repository/postgres/friendship.go
□ internal/repository/postgres/conversation.go
□ internal/repository/postgres/message.go
✓ 驗收: make migrate-up 成功，Repository 測試全過
```

### Agent-4: 後端 Service + Handler (40min)
```
□ internal/service/auth.go         # 含雙卡登入邏輯
□ internal/service/card.go         # 卡片配對、撤銷
□ internal/service/user.go
□ internal/service/friendship.go
□ internal/service/message.go
□ internal/handler/auth.go         # 含雙卡 API
□ internal/handler/user.go
□ internal/handler/friendship.go
□ internal/handler/conversation.go
□ internal/handler/response.go
□ internal/handler/routes.go
□ internal/middleware/*.go
□ Makefile + .air.toml
✓ 驗收: curl 測試雙卡配對、註冊、主卡登入、附卡撤銷
```

### Agent-5: Transport 雙軌制 (35min)
```
□ internal/transport/protocol.go
□ internal/transport/hub.go (介面化)
□ internal/transport/handler.go
□ internal/transport/client.go (WebTransport)
□ internal/transport/ws_client.go (WebSocket)
□ internal/transport/server.go
□ cmd/server/main.go (含連線池優化)
✓ 驗收: WebTransport 失敗時自動切換 WebSocket
```

### Agent-6: 前端加密 + 雙卡 UI (45min)
```
□ src/lib/types.ts
□ src/lib/crypto/*.ts (含 padding)
□ src/lib/api/*.ts (含雙卡 API)
□ src/lib/stores/*.svelte.ts
□ src/lib/components/SecurityWarning.svelte
□ src/lib/components/BackupCardWarning.svelte
□ src/routes/register/start/+page.svelte    # 掃主卡
□ src/routes/register/pair/+page.svelte     # 掃附卡 + 註冊
□ src/routes/login/+page.svelte             # 主卡登入
□ src/routes/login/backup/+page.svelte      # 附卡登入（含警告）
✓ 驗收: 完整雙卡註冊、登入流程可用
```

### Agent-7: 前端 Transport + 聊天 (40min)
```
□ src/lib/transport/webtransport.ts
□ src/lib/transport/websocket.ts
□ src/lib/transport/index.ts (自動降級)
□ src/routes/chat/+page.svelte
□ src/routes/+layout.svelte
□ biome.json, vitest.config.ts
□ 加密測試 + padding 測試
✓ 驗收: 完整 E2E 加密聊天，含 Transport fallback
```

---

## 11. 執行步驟

```bash
# 1. 環境檢查
go version      # >= 1.22
node --version  # >= 20
docker --version

# 2. 初始化
mkdir link && cd link
./scripts/setup-certs.sh
docker compose up -d

# 3. 後端
cd backend
go mod download
make migrate-up
make dev

# 4. 前端（另一個 terminal）
cd frontend
pnpm install
pnpm dev

# 5. 開啟 https://localhost:5173
```

---

## 12. 附錄

### 12.1 docker-compose.yml
```yaml
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: app
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: link
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

### 12.2 scripts/setup-certs.sh
```bash
#!/bin/bash
set -e
mkdir -p certs && cd certs
mkcert -install
mkcert localhost 127.0.0.1 ::1
echo "Done! Certificates created."
```

### 12.3 .gitignore
```
backend/bin/
backend/tmp/
node_modules/
.svelte-kit/
build/
.env
*.local
certs/
coverage*
```

### 12.4 biome.json
```json
{
  "$schema": "https://biomejs.dev/schemas/1.9.0/schema.json",
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true,
      "suspicious": { "noExplicitAny": "error" },
      "style": { "noVar": "error", "useConst": "error" }
    }
  }
}
```

### 12.5 安全聲明 (README.md)
```markdown
## Security

LINK uses **end-to-end encryption**. The server cannot read message content.

### Cryptography
- Key Exchange: X25519
- Encryption: XSalsa20-Poly1305 (AEAD) + Random Padding
- Library: tweetnacl (libsodium compatible)
- Password: Argon2id (OWASP parameters)

### Transport
- Primary: WebTransport (QUIC/HTTP3)
- Fallback: WebSocket (TLS 1.3)

### Authentication
- Dual NFC Card System (Primary + Backup)
- JWT with HS256 whitelist (none algorithm rejected)
- Rate limiting on login/register

### Dual Card Mechanism
```
┌─────────────────────────────────────────────┐
│  Primary Card: Daily use                    │
│  Backup Card: Emergency revocation          │
│                                             │
│  If primary is lost:                        │
│  Scan backup → Primary revoked → Login OK   │
│  ⚠️ Account enters single-card state        │
└─────────────────────────────────────────────┘
```

**Server sees**: metadata (who, when)  
**Server cannot see**: message content ✅

### Known Limitations
- No Forward Secrecy (Phase 2: Double Ratchet)
- Private key stored in browser IndexedDB
- Trust server for public keys (Phase 2: Safety Number)
```

---

**文件結束 - v4.0 Production Ready**
