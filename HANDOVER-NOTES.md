# LINK 專案交接文檔 - 2024-12-14

## 🔴 核心問題：F 和 N 無法正常收發訊息

### 問題症狀
1. F 可以發訊息給 N，但 N 看不到（訊息框都沒有）
2. 原本錯誤 "peer has no public key" → 修復後變成 "invalid encoding" → 現在可以發送但對方看不到
3. 線上/離線狀態不準確

### 根本原因
1. **公鑰問題** ✅ 已修復
   - F 和 N 的公鑰在資料庫中原本是空的
   - 現已更新為正確的 Base64 編碼公鑰

2. **私鑰同步問題** ❌ 未完全解決
   - 客戶端（瀏覽器）沒有對應的私鑰
   - IndexedDB 在 HTTP 環境下可能無法正常工作
   - 密鑰沒有正確載入到 keysStore

3. **訊息解密失敗** ❌ 核心問題
   - 訊息已發送並儲存在資料庫
   - 但接收方無法解密（缺少正確的私鑰）
   - 導致訊息不顯示

## 📝 已嘗試的解決方案（撞牆記錄）

### 1. ❌ `/update-public-key` 頁面
- **問題**: TypeError - "attempted to assign to readonly property"
- **原因**: nacl.box.keyPair.fromSecretKey 類型錯誤
- **嘗試修復**: 添加 instanceof Uint8Array 檢查，但仍有問題

### 2. ❌ `/fix-keys` 頁面  
- **問題**: 可以生成新密鑰但無法更新公鑰到資料庫
- **原因**: 缺少後端 API 端點來更新公鑰

### 3. ❌ `/auto-fix-keys` 頁面
- **問題**: TypeError (同 update-public-key)
- **原因**: 使用了簡化的密鑰生成方法，格式不兼容

### 4. ⚠️ `/import-keys` 頁面
- **狀態**: 已部署但未測試
- **功能**: 讓 F 和 N 導入預設的私鑰
- **URL**: https://link.mcphub.tw/import-keys

### 5. ✅ 手動更新資料庫公鑰
- **成功**: 直接用 SQL 更新了 F 和 N 的公鑰
- **結果**: 解決了 "invalid encoding" 錯誤，可以發送訊息

## 🔑 當前密鑰資訊

### F 的帳號
- 暱稱: F
- 密碼: 000000
- 公鑰: `491IZL9EeBgQER+zM8q1DjZxisq+1F4ONmvucU4Xxmc=`
- 私鑰: `VpRfgl9QTuNwxtHJ++EjyeOZTqODz0I2pekKLfJQ1tg=`

### N 的帳號
- 暱稱: N  
- 密碼: 999999
- 公鑰: `pNEEV/aCJ4K0XUS5FGFX9hXV5+eh/IJeAEL76h/sqzs=`
- 私鑰: `2ar9400RoPfWMGXDlACx3x25hl2JCkIo44Adsuo5YgI=`

## 🚨 待解決的關鍵問題

1. **私鑰同步**: F 和 N 的瀏覽器端沒有正確的私鑰
2. **訊息解密**: 即使訊息發送成功，接收方無法解密顯示
3. **頁面 TypeError**: 多個密鑰修復頁面都有 JavaScript 錯誤

## 💡 建議的解決方案

### 方案 A：修復 import-keys 頁面
1. 檢查並修復 TypeError
2. 確保私鑰正確導入到 IndexedDB 和 keysStore
3. 測試訊息解密功能

### 方案 B：創建新的簡化版本
1. 不使用 nacl 庫，改用原生 crypto API
2. 直接將私鑰存到 localStorage（臨時方案）
3. 確保 keysStore 正確初始化

### 方案 C：後端生成並分發密鑰
1. 在後端生成密鑰對
2. 通過安全的 API 分發給客戶端
3. 客戶端直接使用，不需要本地生成

## 📂 重要文件位置

### 前端
- `/frontend/src/lib/crypto/keys.ts` - 密鑰管理核心
- `/frontend/src/lib/stores/keys.svelte.ts` - 密鑰狀態管理
- `/frontend/src/lib/stores/messages.svelte.ts` - 訊息加解密
- `/frontend/src/routes/import-keys/+page.svelte` - 密鑰導入工具

### 後端
- `/backend/internal/handler/user.go` - UpdateMe 可更新公鑰
- 資料庫: PostgreSQL, 表 `users`, 欄位 `public_key`

### 診斷工具
- https://link.mcphub.tw/test-messages - 檢查訊息系統狀態
- https://link.mcphub.tw/debug - 一般除錯資訊

## 🔧 伺服器資訊
- IP: 34.136.217.56
- 用戶: rocketmantw5516
- 域名: link.mcphub.tw
- 前端: SvelteKit (pnpm build)
- 後端: Go Fiber (systemctl restart link-backend)

## ⚠️ 已知陷阱

1. **HTTPS 問題**: IndexedDB 在 HTTP 環境可能無法使用
2. **Svelte 5 reactivity**: 必須創建新物件而非修改現有物件
3. **Base64 編碼**: 公私鑰必須是正確的 Base64 格式（44字元）
4. **瀏覽器快取**: 每次更新後用戶需要清除快取

## 📋 下一步行動計畫

1. **立即**: 讓 F 和 N 嘗試 https://link.mcphub.tw/import-keys
2. **調試**: 檢查瀏覽器控制台的具體 TypeError 訊息
3. **修復**: 根據錯誤訊息修正 JavaScript 代碼
4. **測試**: 確認訊息可以正確加密、發送、解密、顯示
5. **優化**: 改善密鑰管理流程，避免手動操作

## 🎯 成功標準
- [ ] F 可以發訊息給 N，N 能看到
- [ ] N 可以發訊息給 F，F 能看到  
- [ ] 線上/離線狀態正確顯示
- [ ] 密鑰自動管理，無需手動介入

---

最後更新: 2024-12-14 19:20 (UTC+8)
狀態: 🔴 部分解決，訊息仍無法顯示