# LINK 專案開發進度

## Phase 1 - 環境建置
- [x] 安裝 Homebrew
- [x] 安裝 Go 1.25
- [x] 安裝 Node.js 25
- [x] 安裝 pnpm
- [x] 安裝 PostgreSQL 15
- [x] 安裝 golangci-lint
- [x] 安裝 air (Go 熱重載)
- [x] 安裝 mkcert
- [x] 生成本地 TLS 證書
- [x] 安裝 Docker Desktop

## Phase 2 - 基礎建設
- [x] 創建完整目錄結構
- [x] docker-compose.yml
- [x] scripts/setup-certs.sh
- [x] 初始化前端 (pnpm create svelte@latest)
- [x] 安裝前端依賴 (tweetnacl, tweetnacl-util)
- [x] 創建 CLAUDE.md
- [x] 設定 svelte.config.js (SPA 模式)
- [ ] 啟動 PostgreSQL 容器

## Phase 3 - 後端核心 + Domain
- [x] internal/config/config.go
- [x] internal/domain/errors.go
- [x] internal/domain/user.go
- [x] internal/domain/card.go (雙卡機制)
- [x] internal/domain/session.go
- [x] internal/domain/friendship.go
- [x] internal/domain/conversation.go
- [x] internal/domain/message.go
- [x] internal/pkg/password/argon2.go
- [x] internal/pkg/token/jwt.go
- [x] internal/pkg/circuitbreaker/breaker.go
- [x] migrations/*.sql

## Phase 4 - 後端 Repository
- [x] internal/repository/postgres/user.go
- [x] internal/repository/postgres/card.go
- [x] internal/repository/postgres/session.go
- [x] internal/repository/postgres/friendship.go
- [x] internal/repository/postgres/conversation.go
- [x] internal/repository/postgres/message.go

## Phase 5 - 後端 Service + Handler
- [x] internal/service/auth.go (含雙卡登入)
- [x] internal/service/card.go
- [x] internal/service/user.go
- [x] internal/service/friendship.go
- [x] internal/service/message.go
- [x] internal/handler/auth.go
- [x] internal/handler/user.go
- [x] internal/handler/friendship.go
- [x] internal/handler/conversation.go
- [x] internal/handler/response.go
- [x] internal/handler/routes.go
- [x] internal/middleware/*.go
- [x] Makefile + .air.toml

## Phase 6 - Transport 雙軌制
- [x] internal/transport/protocol.go
- [x] internal/transport/hub.go
- [x] internal/transport/handler.go
- [ ] internal/transport/client.go (WebTransport)
- [x] internal/transport/ws_client.go (WebSocket)
- [x] internal/transport/server.go
- [x] cmd/server/main.go

## Phase 7 - 前端加密 + 雙卡 UI
- [x] src/lib/types.ts
- [x] src/lib/crypto/*.ts (含 padding)
- [x] src/lib/api/*.ts
- [x] src/lib/stores/*.svelte.ts
- [x] src/lib/components/SecurityWarning.svelte
- [x] src/lib/components/BackupCardWarning.svelte
- [x] src/routes/register/start/+page.svelte
- [x] src/routes/register/pair/+page.svelte
- [x] src/routes/login/+page.svelte
- [x] src/routes/login/backup/+page.svelte

## Phase 8 - 前端 Transport + 聊天
- [x] src/lib/transport/webtransport.ts
- [x] src/lib/transport/websocket.ts
- [x] src/lib/transport/index.ts
- [x] src/routes/chat/+page.svelte
- [x] src/routes/+layout.svelte
- [x] biome.json (已有)
- [x] vitest.config.ts
- [x] 加密測試 + padding 測試

## Phase 9 - 測試與驗收
- [x] 後端單元測試 (password, token, circuitbreaker - 29 tests)
- [x] 前端單元測試 (crypto, transport - 29 tests)
- [ ] E2E 測試：雙卡註冊流程
- [ ] E2E 測試：主卡登入
- [ ] E2E 測試：附卡撤銷
- [ ] E2E 測試：加密聊天
- [ ] E2E 測試：Transport fallback
