# 2025-12-14 開發記錄 #01 - 刪除訊息 + 深色主題

## 本次完成工作

### 1. 小安帳號設定

為 AI 登入準備的帳號，設定有效的加密金鑰：

```
密碼: 4242
私鑰 (Base64): bHRKFupegk59jfvLJPMr/GS30qB4kTU7+hrPBsyDsd8=
```

**建立工具**: `backend/cmd/setup-xiaoan/main.go`
- 產生 Argon2 password hash
- 產生 NaCl Box keypair
- 輸出 SQL UPDATE 語句

### 2. UI 美化 - 深色主題

將整個前端從淺色主題改為深色 slate 主題：

| 頁面 | 變更 |
|------|------|
| `+page.svelte` (首頁) | slate-900 背景、漸層按鈕、玻璃效果卡片 |
| `login/+page.svelte` | 深色輸入框、藍色漸層按鈕 |
| `register/+page.svelte` | 一致的深色風格 |
| `chat/+page.svelte` | 深色側邊欄、訊息氣泡美化 |

**設計元素**:
- 背景: `bg-slate-900`
- 卡片: `bg-slate-800/50 backdrop-blur`
- 輸入框: `bg-slate-700/50 border-white/10`
- 主按鈕: `bg-gradient-to-r from-blue-500 to-blue-600`
- 圓角: `rounded-xl` / `rounded-2xl`

### 3. 首頁文字修改

```diff
- 即時低延遲通訊
+ 零知識驗證
```

### 4. 刪除訊息功能

**需求**: 短按訊息即刪除，無確認、不可恢復

#### 後端變更

| 檔案 | 變更 |
|------|------|
| `domain/message.go` | 新增 `FindByID`, `Delete` 介面 |
| `postgres/message.go` | 實作 `FindByID`, `Delete` |
| `service/message.go` | `Delete()` - 檢查只有發送者可刪除 |
| `handler/conversation.go` | `DeleteMessage` handler + WebSocket 通知 |
| `routes.go` | `DELETE /messages/:messageId` |
| `transport/protocol.go` | `TypeDeleted = "deleted"` |
| `transport/hub.go` | `SendTyped()` 供 HTTP handler 使用 |
| `cmd/server/main.go` | 傳遞 hub 到 ConversationHandler |

**API**:
```
DELETE /api/v1/messages/:messageId
Authorization: Bearer <token>

Response: { "data": { "id": "...", "conversation_id": "..." } }
```

**WebSocket 通知** (發送給對方):
```json
{ "t": "deleted", "p": { "id": "msg-id", "conversation_id": "conv-id" } }
```

#### 前端變更

| 檔案 | 變更 |
|------|------|
| `api/conversations.ts` | `deleteMessage(messageId)` |
| `stores/messages.svelte.ts` | `deleteMessage()`, `removeMessage()` |
| `types.ts` | `ITransport.onDeleted` |
| `transport/websocket.ts` | `onDeleted` handler + `deleted` case |
| `transport/webtransport.ts` | 同上 |
| `stores/transport.svelte.ts` | `onDeleted` handler 註冊 |
| `chat/+page.svelte` | 訊息 div → button，點擊觸發刪除 |

**UI 邏輯**:
- 只有自己發送的訊息 (藍色) 可刪除
- pending 狀態的訊息不可刪除
- 對方的訊息 (灰色) 無法點擊刪除

---

## 隱私保護功能 (之前實作)

當使用者切換分頁或 app 時，自動顯示隱私遮罩：
- 監聽 `visibilitychange` 和 `blur` 事件
- 需要輸入密碼才能解除
- 使用 `keysStore.unlock(pwd)` 驗證

---

## 重要檔案位置

```
backend/
├── cmd/setup-xiaoan/main.go     # 小安帳號設定工具
├── internal/
│   ├── domain/message.go        # Message 介面 (FindByID, Delete)
│   ├── repository/postgres/message.go  # 實作
│   ├── service/message.go       # Delete service
│   ├── handler/conversation.go  # DeleteMessage handler
│   └── transport/
│       ├── protocol.go          # TypeDeleted
│       └── hub.go               # SendTyped()

frontend/
├── src/lib/api/conversations.ts        # deleteMessage API
├── src/lib/stores/messages.svelte.ts   # deleteMessage, removeMessage
├── src/lib/stores/transport.svelte.ts  # onDeleted handler
├── src/lib/transport/websocket.ts      # deleted case
├── src/routes/
│   ├── +page.svelte             # 首頁 (深色主題)
│   ├── login/+page.svelte       # 登入頁
│   ├── register/+page.svelte    # 註冊頁
│   └── chat/+page.svelte        # 聊天頁 (刪除功能)
```

---

## Debug 提示

如果刪除功能不運作，檢查瀏覽器 Console：
```
Message clicked! { isOwn: true/false, pending: true/false, ... }
Deleting message: <id>
Delete result: true/false
```

常見問題:
- `isOwn: false` → 點了對方的訊息，無法刪除
- `pending: true` → 訊息還在發送中
- `Delete result: false` → 後端回傳錯誤，檢查 Network tab

---

## 下一步

- [ ] 測試刪除功能是否正常運作
- [ ] 小安 AI 自動回覆功能
- [ ] 訊息已讀狀態顯示
