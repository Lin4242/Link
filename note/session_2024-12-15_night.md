# Session 記錄 (2024-12-15 晚上 ~ 12-16 下午)

> 這是開發記錄，供下次繼續使用
> 最後更新: 2024-12-16 下午 02:10

---

## 最後狀態

### 已完成
- [x] F ↔ 小安 E2EE 通訊修復並測試通過
- [x] 小安 ↔ 阿詠 通訊測試通過
- [x] 手機 ↔ 電腦 跨裝置測試通過
- [x] 訊息泡泡顏色調暗 (#3ACACA → #2A9A9A)
- [x] 時間戳顏色修正 (白色半透明)
- [x] 泡泡尖角改為 SVG 曲線 (但還需調整)
- [x] E 帳號重置為未開卡狀態 (供 demo 用)

### 待辦 (下次繼續) - 重要！

1. **🔴 公鑰同步問題** - 兩邊都要重新整理才能正常通訊
   - 這是最重要的 bug！
   - 問題: 當一方更新公鑰後，另一方還在用舊的公鑰
   - 可能原因: conversation 載入時緩存了 peer.public_key，沒有即時更新
   - 需要檢查: `conversationsStore` 和 `messagesStore` 的公鑰獲取邏輯

2. **泡泡尖角調整** - 目前太尖，需要更胖更圓潤 (參考 LINE)
   - 檔案: `frontend/src/routes/chat/+page.svelte` 第 622-634 行
   - 目前用 SVG path，曲線需要調整

3. **未讀計數功能** - 顯示錯誤的未讀數量
   - 前端有 `unreadCount` 但後端可能沒正確計算
   - 相關檔案: `frontend/src/lib/stores/conversations.svelte.ts`

---

## 剛剛發生的問題

### 🔴 公鑰同步問題 (最重要的 Bug)

**症狀**:
- 新用戶註冊後發訊息，對方無法解密
- 必須兩邊都重新整理頁面才能正常通訊
- 這不合理，應該要即時同步

**根本原因** (待確認):
- conversation 載入時緩存了 `peer.public_key`
- 當對方更新公鑰後，本地還在用舊的公鑰加密/解密
- WebSocket 沒有通知公鑰更新事件

**可能的解法**:
1. 加載訊息時重新獲取對方的最新公鑰
2. 新增 WebSocket 事件通知公鑰更新
3. 定期刷新 conversation 列表

**相關檔案**:
- `frontend/src/lib/stores/conversations.svelte.ts` - conversation 緩存
- `frontend/src/lib/stores/messages.svelte.ts` - 訊息加解密
- `frontend/src/routes/chat/+page.svelte` - 載入邏輯

### F/E 公鑰重設問題

**症狀**: F/E 發送的新訊息，小安顯示 "[無法解密此訊息]"

**原因**: F/E 那邊的 keypair 跟 DB 裡的公鑰不匹配

**解法**:
```sql
-- 把公鑰設成 placeholder，強制重新設定
UPDATE users SET public_key = 'placeholder' WHERE nickname = 'F';
UPDATE users SET public_key = 'placeholder' WHERE nickname = 'E';
```

**後續步驟**:
1. 用戶重新整理頁面
2. 輸入密碼解鎖
3. 系統會自動更新公鑰到 DB
4. 對方也要重新整理頁面 (這就是上面說的 bug)

### E 帳號已重置 (供 Demo 用)

E 用戶已刪除，新的 card_pair 已創建：
- **主卡 Token**: `dd203445f74726fc-1-15a93882`
- **備用卡 Token**: `dd203445f74726fc-2-2b88b66d`
- **有效期**: 30 天

可以用這兩張卡來 demo 完整的註冊流程。

---

## 金鑰系統原理

### 密碼推導金鑰 (確定性)
```typescript
// 同一組 password + userId = 同一組 keypair（任何裝置）
const salt = new TextEncoder().encode(`link-e2e-${userId}`);
const bits = await crypto.subtle.deriveBits(
  { name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' },
  keyMaterial, 256
);
const keyPair = nacl.box.keyPair.fromSecretKey(new Uint8Array(bits));
```

### 金鑰解鎖流程 (chat/+page.svelte 第 349-408 行)
```
用戶輸入密碼
    ↓
deriveKeyPairFromPassword(pwd, userId)
    ↓
比對伺服器公鑰
    ├── 匹配 → 儲存到 IndexedDB + 使用
    ├── placeholder → 更新伺服器 + 儲存 + 使用
    └── 不匹配 → 報錯 (密碼錯誤)
```

### 常見問題
| 問題 | 原因 | 解法 |
|------|------|------|
| 新訊息無法解密 | 公鑰不同步 | 重設為 placeholder |
| 舊訊息無法解密 | 用錯誤公鑰加密 | 無法修復，預期行為 |
| 換裝置無法解密 | IndexedDB 不同步 | 輸入密碼重新推導 |

---

## Playwright 測試技巧

### 開新瀏覽器 (不是新分頁)
用 `browser_run_code` 建立獨立的 browser context：

```javascript
async (page) => {
  const browser = page.context().browser();
  const newContext = await browser.newContext();
  const newPage = await newContext.newPage();
  await newPage.goto('https://link.mcphub.tw/login?token=TOKEN');
  return 'New browser context opened';
}
```

### 操作其他瀏覽器的頁面
```javascript
async (page) => {
  const browser = page.context().browser();
  const contexts = browser.contexts();
  for (const ctx of contexts) {
    for (const p of ctx.pages()) {
      if (p !== page && p.url().includes('chat')) {
        await p.fill('input[placeholder="輸入訊息..."]', '測試訊息');
        await p.click('button[aria-label="發送訊息"]');
        return 'Message sent from other browser';
      }
    }
  }
}
```

**注意**: 新的 context 不會出現在 `browser_tabs` 列表中

---

## 測試帳號

| 帳號 | Token | 密碼 | 狀態 |
|------|-------|------|------|
| 小安 | `c002bb3026ed5e21-1-f0cbd314` | 424242 | ✅ 正常 |
| 阿詠 | `dccab8bf83cad66c-1-13bb9dc9` | 123456 | ✅ 正常 |
| F | `e1ae970143db444f-2-f99d181d` | 000000 | ⚠️ 公鑰是 placeholder，需重新登入 |
| E | (已刪除，供 demo 重新註冊) | - | 🔄 待重新註冊 |

**E 的新 card_pair** (供 demo):
- 主卡: `dd203445f74726fc-1-15a93882`
- 備用卡: `dd203445f74726fc-2-2b88b66d`

登入 URL: `https://link.mcphub.tw/login?token=TOKEN`

---

## 泡泡尖角 SVG 代碼 (需調整)

目前的代碼在 `chat/+page.svelte` 第 622-634 行：

```svelte
<!-- 泡泡尖角 (LINE 風格圓滑曲線) -->
{#if !msg.decryptFailed}
  <svg
    class="absolute bottom-0 w-3 h-4 {isOwn ? '-right-2' : '-left-2'}"
    viewBox="0 0 12 16"
    style={isOwn ? 'transform: scaleX(1);' : 'transform: scaleX(-1);'}
  >
    <path
      d="M0 0 L0 12 Q0 16, 4 16 L12 16 Q4 16, 4 8 Q4 0, 0 0 Z"
      fill={isOwn ? '#1E8080' : '#1e293b'}
    />
  </svg>
{/if}
```

**問題**: 太尖，需要更胖更圓潤
**參考**: LINE 的聊天泡泡尖角

---

## 常用指令

```bash
# 前端 build + 部署
cd frontend && PATH="/opt/homebrew/bin:$PATH" pnpm build
scp -r frontend/build/* jimmy@link.mcphub.tw:/tmp/link-frontend/
ssh jimmy@link.mcphub.tw 'sudo rm -rf /home/rocketmantw5516/Link/frontend/build/* && sudo cp -r /tmp/link-frontend/* /home/rocketmantw5516/Link/frontend/build/'

# 查看/修改用戶公鑰
ssh jimmy@link.mcphub.tw 'sudo -u postgres psql -d link -c "SELECT id, nickname, public_key FROM users;"'
ssh jimmy@link.mcphub.tw 'sudo -u postgres psql -d link -c "UPDATE users SET public_key = '\''placeholder'\'' WHERE nickname = '\''F'\'';"'

# 後端 logs
ssh jimmy@link.mcphub.tw 'sudo journalctl -u link-backend -f'
```

---

## 重要檔案

| 用途 | 路徑 |
|------|------|
| 聊天頁面 + 金鑰解鎖 | `frontend/src/routes/chat/+page.svelte` |
| 金鑰推導/儲存 | `frontend/src/lib/crypto/keys.ts` |
| 訊息加密 | `frontend/src/lib/crypto/encrypt.ts` |
| 訊息解密 | `frontend/src/lib/crypto/decrypt.ts` |
| 對話 store | `frontend/src/lib/stores/conversations.svelte.ts` |
| 訊息 store | `frontend/src/lib/stores/messages.svelte.ts` |
| 交接文件 | `note/handoff.md` |
