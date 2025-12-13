# 2024-12-13 開發記錄 #01 - Phase 8 Transport + Chat

## 本次完成工作

### Phase 8 - 前端 Transport + 聊天 UI

#### 1. Transport 模組 (`frontend/src/lib/transport/`)

建立了雙軌制傳輸層，支援 WebTransport 優先、WebSocket fallback：

```
src/lib/transport/
├── websocket.ts     # WebSocket 客戶端實作
├── webtransport.ts  # WebTransport 客戶端實作
└── index.ts         # 統一介面 + 自動 fallback
```

**設計邏輯**:
- `createTransport()` 會先嘗試 WebTransport（如果瀏覽器支援且有 URL）
- 失敗則自動降級到 WebSocket
- 兩者實作相同的 `ITransport` 介面，上層無感知

**ITransport 介面**:
```typescript
interface ITransport {
  connect(): Promise<void>;
  disconnect(): void;
  sendMessage(to: string, encryptedContent: string, tempId: string): Promise<void>;
  sendTyping(to: string, conversationId: string): void;
  sendRead(conversationId: string, messageId: string): void;
  onMessage: ((msg: EncryptedMessage) => void) | null;
  onTyping: ((convId: string, userId: string) => void) | null;
  onOnline: ((userId: string) => void) | null;
  onOffline: ((userId: string) => void) | null;
  onDelivered: ((tempId: string, msg: EncryptedMessage) => void) | null;
  onConnected: ((connected: boolean) => void) | null;
}
```

**訊息協定** (JSON over WS/WT):
```json
{ "t": "msg", "p": { "to": "userId", "encrypted_content": "...", "temp_id": "..." } }
{ "t": "typing", "p": { "to": "userId", "conversation_id": "..." } }
{ "t": "read", "p": { "conversation_id": "...", "message_id": "..." } }
```

#### 2. Transport Store 更新

更新 `frontend/src/lib/stores/transport.svelte.ts`：
- 使用新的 transport 模組
- 暴露 `transportType` 讓 UI 知道目前用什麼協定
- 加入 `connecting` 狀態

#### 3. Chat 頁面 (`frontend/src/routes/chat/+page.svelte`)

完整的聊天 UI：
- 左側對話列表 (conversations)
- 右側訊息區域 (messages)
- 即時訊息收發
- 輸入中提示 (typing indicator)
- 在線狀態顯示
- 端對端加密提示
- 連線狀態指示器

**狀態管理**:
- 使用 Svelte 5 的 `$state` 和 `$derived`
- 整合 authStore, keysStore, conversationsStore, messagesStore, friendsStore, transportStore

#### 4. Layout 更新

`frontend/src/routes/+layout.svelte`:
- 加入 PWA meta tags
- mobile web app 支援

---

## Phase 9 - 測試

### 前端測試 (29 tests)

**Crypto 測試** (`src/lib/crypto/__tests__/crypto.test.ts`):
- 加密/解密功能
- Padding 驗證 (最小 256 bytes, 64-byte 邊界對齊)
- 不同 salt 產生不同密文
- Unicode 支援
- 邊界情況

**Transport 測試** (`src/lib/transport/__tests__/transport.test.ts`):
- WebTransport 支援偵測
- Transport 類型選擇邏輯
- WebSocket 連線/斷線
- 各種訊息類型發送
- 事件處理

### 後端測試 (29 tests)

**Password 測試** (`internal/pkg/password/argon2_test.go`):
- Hash/Verify 功能正確性
- 每次 hash 都用不同 salt
- Unicode 密碼
- 長密碼處理

**JWT 測試** (`internal/pkg/token/jwt_test.go`):
- Token 生成/驗證
- 過期處理
- 簽名驗證
- Algorithm confusion attack 防護 (none algorithm 拒絕)

**CircuitBreaker 測試** (`internal/pkg/circuitbreaker/breaker_test.go`):
- 狀態轉換 (Closed → Open → HalfOpen → Closed)
- 超時後重試
- 並發安全

---

## 執行測試指令

```bash
# 前端測試
cd frontend && pnpm test

# 後端測試
cd backend && go test ./internal/pkg/... -v

# 或用 Makefile
cd backend && make test
```

---

## 目前專案狀態

| Phase | 狀態 | 備註 |
|-------|------|------|
| Phase 1 - 環境建置 | ✅ 完成 | |
| Phase 2 - 基礎建設 | ✅ 完成 | PostgreSQL 容器待啟動 |
| Phase 3 - 後端 Domain | ✅ 完成 | |
| Phase 4 - 後端 Repository | ✅ 完成 | |
| Phase 5 - 後端 Service + Handler | ✅ 完成 | |
| Phase 6 - Transport 雙軌制 | ✅ 完成 | WebTransport 後端是 placeholder |
| Phase 7 - 前端加密 + 雙卡 UI | ✅ 完成 | |
| Phase 8 - 前端 Transport + 聊天 | ✅ 完成 | |
| Phase 9 - 測試 | ✅ 完成 | 單元測試 + E2E 測試 |

---

### E2E 測試 (15 tests) - 已完成

**測試檔案**: `backend/tests/e2e_test.go`

測試項目:
- `TestHealthEndpoint` - 健康檢查端點
- `TestCheckCard_NotFound` - 檢查不存在的卡片
- `TestDualCardRegistrationFlow` - 完整雙卡註冊流程
  - StartPair (掃描主卡開始配對)
  - CompletePair (掃描附卡完成配對)
  - Register (註冊帳號)
  - LoginWithPrimaryCard (主卡登入)
  - GetMe (驗證已登入狀態)
  - CheckPrimaryCard (檢查主卡狀態)
  - CheckBackupCard (檢查附卡狀態)
- `TestBackupCardRevocation` - 附卡登入撤銷主卡
  - BackupLoginWithoutConfirm (不確認應該失敗)
  - BackupLoginWithConfirm (確認撤銷)
  - PrimaryCardRevoked (驗證主卡已撤銷)
  - LoginWithRevokedCard (已撤銷卡片無法登入)
- `TestInvalidLogin` - 無效登入
  - NonexistentCard (不存在的卡片)
  - WrongPassword (錯誤密碼)
- `TestProtectedEndpointsRequireAuth` - 受保護端點認證驗證

**執行 E2E 測試**:
```bash
# 需要先啟動伺服器和資料庫
cd /Users/jimmy/project/Link && docker compose up -d
cd backend && make migrate-up
./bin/server &

# 執行測試
go test ./tests/... -v
```

**注意事項**:
- Rate Limiter 設定: 註冊 5 次/小時
- 測試間需加入 `time.Sleep(2 * time.Second)` 避免觸發限速
- 測試用 card_token 長度需 ≤ 32 字元 (VARCHAR(32) 限制)

---

## 重要檔案位置

```
frontend/
├── src/lib/transport/          # Transport 模組
│   ├── websocket.ts
│   ├── webtransport.ts
│   └── index.ts
├── src/lib/crypto/             # 加密模組
│   ├── keys.ts                 # 金鑰管理 (IndexedDB)
│   ├── encrypt.ts              # 加密 + padding
│   └── decrypt.ts              # 解密 + unpadding
├── src/lib/stores/             # Svelte 5 Stores
├── src/routes/chat/+page.svelte # 聊天頁面
└── vitest.config.ts            # 測試設定

backend/
├── internal/pkg/
│   ├── password/argon2.go      # Argon2id 密碼 hash
│   ├── token/jwt.go            # JWT 管理
│   └── circuitbreaker/         # 熔斷器
├── tests/e2e_test.go           # E2E 測試
└── cmd/server/main.go          # 入口點
```

---

## 下一步

Phase 9 測試已完成！可能的後續工作：

1. **WebTransport 後端實作** - 目前是 placeholder
2. **前端整合測試** - 測試完整使用者流程
3. **部署準備** - Docker 化、CI/CD 設定
4. **NFC 卡片整合** - NTAG 424 DNA 實際硬體測試
