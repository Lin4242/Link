# LINK 開發筆記

> 開發日誌命名規則: `YYYY-MM-DD-NN-主題.md`
> 例如: `2024-12-13-01-phase8-transport-chat.md`

---

## 環境資訊

**機器**: Apple Silicon Mac (arm64), macOS 15.5

**已安裝工具**:
| 工具 | 版本 | 路徑 |
|------|------|------|
| Go | 1.25.5 | /opt/homebrew/bin/go |
| Node.js | 25.2.1 | /opt/homebrew/bin/node |
| pnpm | 10.25.0 | /opt/homebrew/bin/pnpm |
| PostgreSQL | 15.15 | /opt/homebrew/opt/postgresql@15/bin |
| golangci-lint | 2.7.2 | /opt/homebrew/bin/golangci-lint |
| air | 1.63.4 | ~/go/bin/air |
| mkcert | 1.4.4 | /opt/homebrew/bin/mkcert |
| Docker | 29.1.2 | /usr/local/bin/docker |

**TLS 證書位置**: `/Users/jimmy/project/Link/certs/`
- localhost+2.pem (證書)
- localhost+2-key.pem (私鑰)
- 有效期至 2028-03-13

---

## 重要提醒

### 啟動前必做
1. **重開終端機** - 讓 PATH 設定生效
2. **啟動 Docker Desktop** - 在 Applications 開啟
3. **啟動 PostgreSQL**: `brew services start postgresql@15` (或用 Docker)

### PATH 設定 (~/.zshrc)
```bash
eval "$(/opt/homebrew/bin/brew shellenv)"
export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"
export PATH="$HOME/go/bin:$PATH"
```

---

## 已知地雷 (來自 SPEC)

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

## 開發指令

```bash
# 後端
cd backend && make dev      # 開發模式 (air 熱重載)
cd backend && make test     # 執行測試
cd backend && make lint     # 程式碼檢查

# 前端
cd frontend && pnpm dev     # 開發模式
cd frontend && pnpm test    # 執行測試
cd frontend && pnpm build   # 建置

# 資料庫
brew services start postgresql@15   # 啟動 (本地)
docker compose up -d                # 啟動 (Docker)
make migrate-up                     # 執行 migration
```

---

## 技術決策記錄

### 2024-12-13: 環境建置
- 選用 Homebrew 安裝所有工具
- PostgreSQL 用 brew 而非 Docker (開發方便)
- TLS 證書用 mkcert 生成本地信任的證書

---

## 待確認事項
- [ ] NFC 卡片硬體規格？
- [ ] 是否需要 iOS/Android App？
- [ ] 部署環境 (AWS/GCP/自建)？
