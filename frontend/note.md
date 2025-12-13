# Link 開發筆記

## 2025-12-14 開發經驗與踩坑紀錄

### 1. WebSocket 訊息確認 (Delivered) 收不到

**問題：** 發送訊息後，前端一直顯示「發送中」，無法收到 delivered 確認。

**根本原因：** Svelte 5 的 `$state()` 會創建 Proxy 物件，導致 handler 被設定在 Proxy 上而非原始物件。

**錯誤的寫法：**
```typescript
let transport = $state<ITransport | null>(null);
```

**正確的寫法：**
```typescript
// IMPORTANT: Don't use $state for transport - it creates a proxy that breaks handler assignment
let transport: ITransport | null = null;
```

**教訓：** 當物件需要直接設定 callback handlers 時，不要使用 `$state()`。

---

### 2. API 回傳欄位名稱不一致 (PascalCase vs snake_case)

**問題：** 前端無法讀取訊息內容，`m.id`, `m.sender_id` 全部是 `undefined`。

**根本原因：** Go 的 Message struct 沒有加 JSON tags，導致 JSON 序列化時使用 PascalCase。

**錯誤的 Go struct：**
```go
type Message struct {
    ID               string
    ConversationID   string
    SenderID         string
    EncryptedContent string
    CreatedAt        time.Time
}
```

**正確的 Go struct：**
```go
type Message struct {
    ID               string     `json:"id"`
    ConversationID   string     `json:"conversation_id"`
    SenderID         string     `json:"sender_id"`
    EncryptedContent string     `json:"encrypted_content"`
    CreatedAt        time.Time  `json:"created_at"`
    DeliveredAt      *time.Time `json:"delivered_at"`
    ReadAt           *time.Time `json:"read_at"`
}
```

**教訓：** Go struct 一定要加 JSON tags，確保前後端欄位名稱一致。

---

### 3. 訊息解密失敗 - 對方公鑰無效

**問題：** 所有訊息解密都失敗，`nacl.box.open` 返回 `null`。

**根本原因：** 對方用戶（小安）的 public_key 是全零 `AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=`，這是無效的公鑰。

**解決方案：**
1. 為小安生成有效的 NaCl key pair
2. 更新資料庫中的 public_key
3. 刪除用無效公鑰加密的舊訊息（無法解密）

```sql
-- 更新小安的公鑰
UPDATE users SET
  nickname = '小安',
  public_key = 'UA78a/QTUz7TLZc2fEimp76v4sTNdmFdnnhEWqRIbiI='
WHERE id = 'fcf454d3-d34a-4765-bc75-e9c3aa4bd9c3';

-- 刪除無法解密的舊訊息
DELETE FROM messages WHERE conversation_id = '...';
```

**教訓：**
- 系統預設用戶（如小安）必須在資料庫初始化時就設定有效的公鑰
- 可以考慮加入公鑰驗證，拒絕全零或明顯無效的公鑰

---

### 4. 金鑰解鎖流程問題

**問題：** 每次重新整理後訊息都消失。

**根本原因：**
1. 用戶每次輸入不同密碼 → 解鎖失敗 → 產生新金鑰 → 舊訊息無法解密
2. 解鎖流程沒有警告用戶產生新金鑰的後果

**解決方案：**
在產生新金鑰前加入確認對話框：
```typescript
if (!success) {
    const confirmed = confirm('密碼錯誤或金鑰不存在。是否要產生新的金鑰？\n\n警告：這將導致舊訊息無法解密！');
    if (!confirmed) {
        return;
    }
    // Generate new keys...
}
```

**教訓：** 端對端加密的金鑰管理非常重要，必須清楚提示用戶：
- 密碼必須和第一次設定的一致
- 產生新金鑰會導致舊訊息無法解密
- 考慮加入密碼恢復機制或金鑰備份功能

---

### 5. JWT Token 的 claim 名稱

**問題：** Go WebSocket 測試連線成功但 user_id 是空的。

**根本原因：** JWT payload 使用 `sub` 作為 user ID 的 claim 名稱，但後端 Claims struct 期望的是 `uid`。

**錯誤：**
```go
claims := jwt.MapClaims{
    "sub": userID,  // 錯誤
}
```

**正確：**
```go
claims := jwt.MapClaims{
    "uid": userID,  // 正確，對應後端 Claims struct
}
```

**教訓：** 確認 JWT 的 claim 名稱與後端解析邏輯一致。

---

### 6. NaCl Box 加解密原理

**重要概念：**
- 加密：`nacl.box(plaintext, nonce, theirPublicKey, mySecretKey)`
- 解密：`nacl.box.open(ciphertext, nonce, theirPublicKey, mySecretKey)`

**關鍵：** NaCl box 是對稱的 - A 和 B 都可以用自己的 secret key + 對方的 public key 來解密訊息。

```
A 發送給 B：
- A 加密：A_secret + B_public
- A 解密自己的訊息：A_secret + B_public ✓
- B 解密：B_secret + A_public ✓

訊息解密時，使用：mySecretKey + peerPublicKey
對於自己發的訊息和收到的訊息，都用同樣的 key 組合。
```

---

### 7. IndexedDB 金鑰儲存

**流程：**
1. 用戶輸入密碼
2. 用 PBKDF2 從密碼派生加密金鑰 (100,000 iterations)
3. 用 NaCl secretbox 加密 secret key
4. 儲存 { salt, nonce, encrypted } 到 IndexedDB

**解鎖流程：**
1. 從 IndexedDB 讀取 { salt, nonce, encrypted }
2. 用相同密碼 + salt 派生金鑰
3. 解密 secret key
4. 如果解密失敗（密碼錯誤），返回 null

**教訓：** IndexedDB 在同一個 origin 下是持久的，但 Playwright 測試每次都是新的 browser context。

---

### 8. Playwright E2E 測試技巧

**處理 confirm/alert 對話框：**
```typescript
page.on('dialog', async dialog => {
    console.log('Dialog:', dialog.message());
    await dialog.accept(); // 或 dialog.dismiss()
});
```

**等待元素可見：**
```typescript
await page.locator('text=某文字').isVisible();
```

**模態框阻擋點擊：**
如果有模態框遮住元素，Playwright 會報錯 "intercepts pointer events"。必須先關閉模態框。

---

### 9. 除錯技巧

**加入詳細 console.log：**
```typescript
console.log('=== loadMessages called ===', {
    conversationId,
    peerPublicKey: peerPublicKey?.substring(0, 20) + '...',
    hasSecretKey: !!keysStore.secretKey,
    secretKeyLength: keysStore.secretKey?.length
});
```

**API 回應檢查：**
```bash
curl -sk 'https://..../api/v1/messages' -H 'Authorization: Bearer TOKEN'
```

**資料庫直接查詢：**
```bash
PGPASSWORD=secret psql -h localhost -U app -d link -c "SELECT * FROM messages;"
```

---

### 10. 常見錯誤總結

| 問題 | 原因 | 解決方案 |
|------|------|----------|
| Handler 不被調用 | Svelte 5 $state proxy | 不對 transport 使用 $state |
| API 欄位 undefined | Go struct 沒有 json tags | 加上 `json:"field_name"` |
| 解密失敗 | 公鑰無效或金鑰不匹配 | 驗證公鑰、確保同密碼 |
| JWT user_id 空 | claim 名稱錯誤 | 使用 `uid` 而非 `sub` |
| 訊息消失 | 每次產生新金鑰 | 警告用戶、用同密碼 |

---

### 11. 隱私保護遮罩

**需求：** 切換分頁或 app 時，不要讓系統截圖看到聊天內容。

**解決方案：** 監聽 `visibilitychange` 和 `blur` 事件，在離開頁面時顯示遮罩。

```typescript
// blur 比 visibilitychange 更早觸發
function handleWindowBlur() {
    privacyScreen = true;
}

function handleVisibilityChange() {
    if (document.hidden) {
        privacyScreen = true;
    }
}

onMount(() => {
    document.addEventListener('visibilitychange', handleVisibilityChange);
    window.addEventListener('blur', handleWindowBlur);
});
```

**遮罩 UI：**
- 藍色漸層背景
- 顯示 Link logo 和「端對端加密通訊」
- 點擊任意處解除遮罩
- z-index 設為 100 確保覆蓋所有內容

**注意事項：**
- `blur` 事件更早觸發，確保在系統截圖前顯示遮罩
- 回到頁面時不自動解除，需要用戶主動點擊
- 這是安全功能，防止旁人偷看或系統截圖洩露隱私

---

## 待辦事項

- [x] 分頁切換時隱藏聊天內容（隱私保護）
- [ ] 金鑰備份/恢復機制
- [ ] 小安自動回覆功能
- [ ] 密碼強度驗證
- [ ] 忘記密碼處理流程
- [ ] iOS PWA 支援（Add to Home Screen）
