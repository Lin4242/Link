# LINK

端對端加密即時通訊應用程式。

## 功能特色

- **端對端加密** - 使用 NaCl Box 加密，只有通訊雙方能解密訊息
- **實體卡片認證** - 使用 NFC 卡片進行身份驗證
- **隱私保護** - 切換應用程式時自動顯示遮罩，防止截圖洩露
- **即時通訊** - 基於 WebSocket 的即時訊息傳遞
- **金鑰保護** - 私鑰使用密碼加密儲存於 IndexedDB

## 技術棧

- **Frontend**: SvelteKit 5, TypeScript, TailwindCSS
- **Backend**: Go, Fiber, PostgreSQL
- **加密**: TweetNaCl (NaCl Box for E2E encryption)
- **傳輸**: WebSocket

## 開發

```bash
# 安裝依賴
npm install

# 啟動開發伺服器
npm run dev

# 建置
npm run build
```

## 環境變數

```env
VITE_WS_URL=wss://your-server:9443/ws
VITE_API_URL=https://your-server:9443/api/v1
```

## 安全設計

1. **金鑰生成** - 每個用戶註冊時生成 NaCl key pair
2. **金鑰儲存** - 私鑰使用 PBKDF2 派生的密碼加密後存入 IndexedDB
3. **訊息加密** - 發送前使用收件者公鑰 + 發送者私鑰加密
4. **隱私遮罩** - 離開頁面時立即顯示遮罩，需輸入密碼解鎖

## 授權

Private - All rights reserved
