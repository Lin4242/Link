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
```bash
cd backend && make dev      # 後端開發
cd backend && make test     # 後端測試
cd frontend && pnpm dev     # 前端開發
cd frontend && pnpm test    # 前端測試
```
