# Claude 交接文件

> 每次重啟 Claude 後閱讀此文件快速恢復上下文
> **敏感資訊 (密碼/token)**: 見 `local_note.md` (gitignored)

---

## 專案狀態 (2024-12-15 06:05 更新)

### 當前狀態
- **線上環境**: ✅ 已部署最新版本
- **Seed 資料**: 小安 + 阿詠 + F + Demo 卡片
- **自動加好友**: 新用戶註冊後自動與小安成為好友
- **E2EE 通訊**: ✅ 全部正常

### 待辦事項 (下次繼續)

1. **泡泡尖角調整** - 已改成 SVG 曲線，但太尖了，需要更胖更圓潤
   - 檔案: `frontend/src/routes/chat/+page.svelte` 第 622-634 行
   - 目前用 SVG path，需調整曲線讓尖角更圓

2. **未讀計數功能** - 小安看到 F 有 9 個未讀，但實際沒那麼多
   - 前端有 `unreadCount` 但後端可能沒正確計算
   - 相關檔案: `frontend/src/lib/stores/conversations.svelte.ts`

### 測試帳號狀態

| 帳號 | 狀態 | public_key | 備註 |
|------|------|------------|------|
| 小安 | ✅ 正常 | XdsZAOR4KfQiF4P5SpiaV7dJj+Hz96BW+V0bzomiuyk= | 服務帳號 |
| 阿詠 | ✅ 正常 | qmlu7ws5LLCG3uuMcaH4UzWsG1YYdXw5HaEv7PFkmRg= | 測試帳號 |
| F | ✅ 正常 | xMmbmPw5xuHtui9P4UKpHjl9Wnx2qGXPeZAQtce8pUM= | 已修復 |

### 好友關係
- 小安 ↔ 阿詠: ✅ 雙向通訊正常
- 小安 ↔ F: ✅ 雙向通訊正常 (2024-12-15 05:21 測試通過)

---

## ✅ 已修復: F ↔ 小安 通訊問題

### 修復內容 (2024-12-15)

1. **金鑰解鎖邏輯改為強制驗證** (`chat/+page.svelte`)
   - 改為永遠從密碼推導金鑰，不再依賴 IndexedDB
   - 推導後與伺服器公鑰比對
   - 若伺服器是 placeholder，自動更新

2. **F 的公鑰已同步**
   - DB 公鑰: `xMmbmPw5xuHtui9P4UKpHjl9Wnx2qGXPeZAQtce8pUM=`
   - 與 F 密碼 (000000) 推導結果一致

3. **測試結果**
   - F → 小安: "測試訊息from F" ✅ 解密成功
   - 小安 → F: "回覆from小安！通訊成功" ✅ 解密成功
   - 即時推送: ✅ WebSocket 即時收到

### 舊訊息無法解密是正常的

修復前發送的訊息使用了錯誤的公鑰加密，這些訊息會顯示 "[無法解密此訊息]"。這是預期行為，無法修復。

---

## 歷史問題分析 (供參考)

### 原始問題

金鑰解鎖邏輯優先從 IndexedDB 載入，若成功則跳過伺服器公鑰驗證：

```typescript
// 舊邏輯 (有 Bug)
let success = await keysStore.unlock(pwd);
if (!success && authStore.user?.id) {
    // 只有 IndexedDB 失敗才會驗證伺服器公鑰
}
```

### 修復後邏輯

```typescript
// 新邏輯 - 永遠驗證
const { publicKey, secretKey } = await deriveKeyPairFromPassword(pwd, userId);
if (serverPublicKey === publicKey) {
    // 匹配 → 使用
} else if (serverPublicKey?.startsWith('placeholder')) {
    // placeholder → 更新伺服器
} else {
    // 不匹配 → 報錯
}
```

---

## 解決方案 (5 種)

### 方案 A: 手動清除 IndexedDB (最快，需用戶操作)

讓 F 執行以下步驟：
1. 打開 Chrome DevTools → Application → Storage → IndexedDB
2. 找到 `link-keys` 資料庫，右鍵刪除
3. 重新整理頁面
4. 輸入密碼 → 會觸發密碼推導 → 更新伺服器公鑰
5. 小安也要重新整理頁面，獲取 F 的新公鑰

### 方案 B: 修改程式碼 - 強制驗證公鑰 (推薦)

修改 `frontend/src/routes/chat/+page.svelte` 的金鑰解鎖邏輯：

```typescript
// 找到第 349 行的 form onsubmit handler，改成：
form onsubmit={async (e) => {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const pwd = (form.elements.namedItem('unlockPwd') as HTMLInputElement).value;
    if (!pwd) {
        alert('請輸入密碼');
        return;
    }

    // 無論 IndexedDB 有沒有，都要從密碼推導並驗證
    if (!authStore.user?.id) {
        alert('用戶資料不完整，請重新登入');
        return;
    }

    const { deriveKeyPairFromPassword, saveSecretKey } = await import('$lib/crypto/keys');
    const { publicKey, secretKey } = await deriveKeyPairFromPassword(pwd, authStore.user.id);

    const serverPublicKey = authStore.user.public_key;
    console.log('Derived public key:', publicKey);
    console.log('Server public key:', serverPublicKey);

    if (serverPublicKey === publicKey) {
        // 密碼正確，金鑰匹配
        console.log('✅ Derived key matches server!');
        await saveSecretKey(secretKey, pwd);
        await keysStore.save(secretKey, pwd);
    } else if (serverPublicKey?.startsWith('placeholder')) {
        // 首次登入或需要重新同步，自動設定公鑰
        console.log('📝 Setting public key on server');
        await saveSecretKey(secretKey, pwd);
        await keysStore.save(secretKey, pwd);
        const { updateMe } = await import('$lib/api/users');
        await updateMe({ public_key: publicKey });
        // 更新本地 authStore
        authStore.setUser({ ...authStore.user, public_key: publicKey });
    } else {
        // 公鑰不匹配且不是 placeholder - 密碼錯誤
        console.error('❌ Key mismatch');
        alert('密碼錯誤或金鑰不匹配。請確認密碼正確。');
        return;
    }

    // 重新載入訊息
    if (activeConversation) {
        await messagesStore.loadMessages(
            activeConversation.id,
            activeConversation.peer.public_key
        );
    }
}}
```

### 方案 C: 新增「重置金鑰」按鈕

在聊天頁面設定或側邊欄加一個按鈕：

```svelte
<button onclick={async () => {
    if (!confirm('確定要重置加密金鑰嗎？這會清除本地儲存的金鑰。')) return;
    await keysStore.clear();  // 清除 IndexedDB
    window.location.reload();  // 重新載入頁面觸發重新解鎖
}}>
    重置金鑰
</button>
```

### 方案 D: 登入時自動檢查並同步

在 `/login` 成功後，自動檢查公鑰是否需要同步：

```typescript
// frontend/src/routes/login/+page.svelte 登入成功後
async function onLoginSuccess(user) {
    // 如果伺服器公鑰是 placeholder，標記需要重新設定
    if (user.public_key?.startsWith('placeholder')) {
        // 清除 IndexedDB，強制重新推導
        const { clearSecretKey } = await import('$lib/crypto/keys');
        await clearSecretKey();
    }
    goto('/chat');
}
```

### 方案 E: 後端 API 強制重置

新增 API 端點讓管理員重置用戶公鑰：

```go
// POST /api/v1/admin/users/:id/reset-key
func (h *AdminHandler) ResetUserKey(c *fiber.Ctx) error {
    userID := c.Params("id")
    _, err := h.pool.Exec(c.Context(),
        "UPDATE users SET public_key = 'placeholder' WHERE id = $1", userID)
    if err != nil {
        return err
    }
    return c.JSON(fiber.Map{"message": "Public key reset to placeholder"})
}
```

---

## 推薦行動順序

1. **立即** - 先讓 F 手動清除 IndexedDB (方案 A)，測試是否修復
2. **然後** - 實作方案 B，避免未來再發生同樣問題
3. **可選** - 方案 C 或 D 作為用戶友好的補救措施

---

## 今日已完成 (2024-12-15)

### UI 改動
- [x] **藍色系改為 #3ACACA** - 亮藍色都換成青綠色
- [x] **縮小圓角** - `rounded-xl` → `rounded-md`
- [x] **手機 safe-area 支援** - 修復 iPhone 輸入區被截斷
- [x] **金鑰措辭修正** - "解鎖" 改為 "載入"
- [x] **訊息泡泡顏色調暗** - `#3ACACA` → `#2A9A9A` (對比度更好)
- [x] **時間戳顏色修正** - 自己訊息的時間改為白色半透明 `rgba(255,255,255,0.6)`

### 後端改動
- [x] **在線狀態通知修復** - Hub 新增 onConnect/onDisconnect callbacks

### 重大 Bug 修復
- [x] **F ↔ 小安 通訊** - 金鑰解鎖邏輯改為強制驗證伺服器公鑰

### 測試結果 (2024-12-15 05:51 最終確認)
- [x] 小安 ↔ 阿詠: ✅ 雙向通訊正常 (兩個獨立瀏覽器測試)
- [x] 小安 ↔ F: ✅ 雙向通訊正常 (電腦 ↔ 手機跨裝置測試)
- [x] 在線狀態: ✅ 即時更新
- [x] 手機端測試: ✅ F 用手機發訊息給小安，解密成功

### 測試帳號快速登入

| 帳號 | Token | 密碼 |
|------|-------|------|
| 小安 | `c002bb3026ed5e21-1-f0cbd314` | 424242 |
| 阿詠 | `dccab8bf83cad66c-1-13bb9dc9` | 123456 |
| F | `e1ae970143db444f-2-f99d181d` | 000000 |

登入 URL 格式: `https://link.mcphub.tw/login?token=TOKEN`

---

## Playwright 開新瀏覽器技巧

當需要同時登入兩個帳號測試即時通訊時，用 `browser_run_code` 建立獨立的 browser context：

### 開新瀏覽器並登入
```javascript
async (page) => {
  const browser = page.context().browser();
  const newContext = await browser.newContext();
  const newPage = await newContext.newPage();
  await newPage.goto('https://link.mcphub.tw/login?token=TOKEN');
  return 'New browser context opened';
}
```

### 操作新瀏覽器中的頁面
```javascript
async (page) => {
  const browser = page.context().browser();
  const contexts = browser.contexts();
  for (const ctx of contexts) {
    for (const p of ctx.pages()) {
      if (p !== page && p.url().includes('chat')) {
        // 操作另一個瀏覽器的頁面
        await p.fill('input[placeholder="輸入訊息..."]', '測試訊息');
        await p.click('button[aria-label="發送訊息"]');
        return 'Message sent from other browser';
      }
    }
  }
}
```

**注意**: 新的 context 不會出現在 `browser_tabs` 列表中，需用 `browser_run_code` 操作

---

## 重要檔案路徑

| 用途 | 路徑 | 關鍵行號 |
|------|------|----------|
| 金鑰解鎖邏輯 | `frontend/src/routes/chat/+page.svelte` | 349-408 |
| keysStore | `frontend/src/lib/stores/keys.svelte.ts` | 13-27 (unlock) |
| 金鑰推導/儲存 | `frontend/src/lib/crypto/keys.ts` | 16-45, 61-85 |
| Hub | `backend/internal/transport/hub.go` | 全部 |
| 後端入口 | `backend/cmd/server/main.go` | 69-84 (online notify) |

---

## E2EE 金鑰系統圖解

```
┌─────────────────────────────────────────────────────────────────┐
│                     用戶輸入密碼                                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│           目前邏輯 (有 Bug)                                       │
│  1. keysStore.unlock(pwd) ─── IndexedDB 有舊 key ──→ 直接使用    │
│                         │                           (跳過驗證!)  │
│                         └── IndexedDB 沒有 ──→ 2. 推導金鑰       │
│                                                                 │
│  2. deriveKeyPairFromPassword(pwd, userId)                      │
│     └── 比對伺服器公鑰 ──→ 匹配則成功                             │
│                        ──→ placeholder 則更新伺服器               │
│                        ──→ 不匹配則報錯                          │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│           修復後邏輯 (方案 B)                                     │
│  1. deriveKeyPairFromPassword(pwd, userId)  ← 永遠先推導         │
│  2. 比對伺服器公鑰                                               │
│     └── 匹配 ──→ 儲存到 IndexedDB + 使用                         │
│     └── placeholder ──→ 更新伺服器 + 儲存 + 使用                  │
│     └── 不匹配 ──→ 報錯 (密碼錯誤)                               │
└─────────────────────────────────────────────────────────────────┘
```

---

## 資料庫查詢

```sql
-- 查看用戶公鑰
SELECT id, nickname, public_key FROM users;

-- 查看好友關係
SELECT
  u1.nickname as requester,
  u2.nickname as addressee,
  f.status
FROM friendships f
JOIN users u1 ON f.requester_id = u1.id
JOIN users u2 ON f.addressee_id = u2.id;

-- 重設 F 的公鑰為 placeholder (觸發重新同步)
UPDATE users SET public_key = 'placeholder' WHERE nickname = 'F';
```

---

## 部署方式

### 方式 1: Claude `/deploy` 命令 (推薦)
直接輸入 `/deploy`，Claude 會照著 `.claude/commands/deploy.md` 的清單執行每一步。

### 方式 2: GitHub Actions 自動部署
Push 到 main 後自動部署。需要先設定：

1. 到 GitHub repo → Settings → Secrets and variables → Actions
2. 新增 secret: `SSH_PRIVATE_KEY`（jimmy 的 SSH 私鑰）

設定完成後，每次 push 到 main 都會自動：
- Build 前端
- SSH 到伺服器 pull + build 後端
- 上傳前端 build
- 重啟服務
- 健康檢查

### 方式 3: 手動部署
見下方常用指令。

---

## 常用指令

```bash
# 重啟後端
ssh jimmy@link.mcphub.tw 'sudo systemctl restart link-backend'

# 查看 logs
ssh jimmy@link.mcphub.tw 'sudo journalctl -u link-backend -f'

# Pull + Build + Restart 後端
ssh jimmy@link.mcphub.tw 'cd /home/rocketmantw5516/Link && sudo -u rocketmantw5516 git pull && cd backend && sudo -u rocketmantw5516 /usr/local/go/bin/go build -o bin/server ./cmd/server && sudo systemctl restart link-backend'

# 部署前端 (完整步驟)
# 1. Build (注意: 需要用完整路徑或設定 PATH)
PATH="/opt/homebrew/bin:$PATH" pnpm build
# 2. Upload
scp -r frontend/build/* jimmy@link.mcphub.tw:/tmp/link-frontend/
# 3. Deploy (清除舊檔案，複製新檔案)
ssh jimmy@link.mcphub.tw 'sudo rm -rf /home/rocketmantw5516/Link/frontend/build/* && sudo cp -r /tmp/link-frontend/* /home/rocketmantw5516/Link/frontend/build/'

# 資料庫查詢
ssh jimmy@link.mcphub.tw 'sudo -u postgres psql -d link -c "SELECT id, nickname, public_key FROM users;"'
```

### 前端部署重要說明

SvelteKit 使用 content hash 作為檔案名稱（如 `0.D4uDu4_Q.css`），所以：
- **不需要** 手動加 timestamp 或 cache busting
- **不需要** 設定 Cache-Control headers（因為檔案名本身就會變）
- **需要** 每次修改後重新 `pnpm build` 並部署

如果前端修改沒生效：
1. 確認有執行 `pnpm build`
2. 確認有上傳到伺服器
3. 確認檔案 hash 有變化（用 `ls` 檢查 `build/_app/immutable/assets/`）

---

## 伺服器資訊

| 項目 | 值 |
|------|---|
| SSH | `ssh jimmy@link.mcphub.tw` |
| App 路徑 | `/home/rocketmantw5516/Link/` |
| 網址 | https://link.mcphub.tw |

---

## 已知地雷

| 問題 | 原因 | 解法 |
|------|------|------|
| SSH 權限 | jimmy 不是 app owner | `sudo -u rocketmantw5516` |
| Go 找不到 | PATH 問題 | `/usr/local/go/bin/go` |
| pnpm/npm 找不到 | Claude 的 shell 沒有 PATH | `PATH="/opt/homebrew/bin:$PATH" pnpm ...` |
| 前端修改沒生效 | 沒有 build 或沒有部署 | 每次改前端都要 build + deploy |
| 公鑰不同步 | IndexedDB 載入跳過驗證 | 實作方案 B |
| 解密失敗 | 公鑰 cache 過期 | 重新載入頁面 |
