# 2024-12-14 技術問題與解決方案

## 遇到的問題與解決

### 1. TLS 證書問題
**症狀**：
```
cannot load TLS key pair: open ../certs/localhost+2.pem: no such file or directory
```

**原因**：
- 開發環境使用 mkcert 生成的本地證書
- 生產環境沒有這些證書檔案

**解決方案**：
1. 短期：生成自簽證書
2. 長期：使用 Let's Encrypt 真實證書
3. 程式碼改進：根據環境自動選擇證書路徑

### 2. Admin Password 硬編碼
**症狀**：
- Admin 密碼寫死在程式碼中 ("42424242")
- 無法在不改程式碼的情況下修改密碼

**原因**：
- 初期開發時的便利做法
- 沒有考慮生產環境需求

**解決方案**：
```go
// 修改 config/config.go
type Config struct {
    // ...
    AdminPassword string
    BaseURL      string
}

// 修改 main.go
adminHandler := handler.NewAdminHandler(
    cardTokenGen, 
    cfg.AdminPassword,  // 從 config 讀取
    cfg.BaseURL        // 從 config 讀取
)
```

### 3. Git 同步問題
**症狀**：
```
error: Your local changes would be overwritten by merge
```

**原因**：
- 使用 scp 直接修改 server 檔案
- Server 和 GitHub 的程式碼不同步

**解決方案**：
1. 強制重設到 GitHub 版本：`git reset --hard HEAD`
2. 建立正確的部署流程
3. 禁止直接在 server 上修改

### 4. Frontend Build 環境變數
**症狀**：
- 前端 API URL 硬編碼
- 無法動態切換環境

**原因**：
- 使用硬編碼的 API_BASE

**解決方案**：
```javascript
const API_BASE = import.meta.env.VITE_API_URL ? 
    `${import.meta.env.VITE_API_URL}/api/v1` : 
    `${window.location.origin}/api/v1`;
```

### 5. pnpm 權限問題
**症狀**：
```
EACCES: permission denied, mkdir '/usr/local/lib/node_modules/pnpm'
```

**原因**：
- 全域安裝需要 root 權限

**解決方案**：
- 使用 `sudo npm install -g pnpm`
- 或設定 npm prefix 到使用者目錄

### 6. WebSocket 連線失敗
**症狀**：
- WebSocket 無法建立連線
- 錯誤：`WebSocket connection failed`

**原因**：
- Nginx 沒有正確配置 WebSocket 代理

**解決方案**：
```nginx
location /ws {
    proxy_pass http://localhost:8443;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}
```

### 7. CORS 問題
**症狀**：
- 前端無法呼叫 API
- 瀏覽器顯示 CORS 錯誤

**原因**：
- CORS_ORIGINS 設定不正確

**解決方案**：
- 更新 .env：`CORS_ORIGINS=https://link.mcphub.tw,http://link.mcphub.tw`
- 確保包含所有可能的來源

## 架構決策

### 為什麼選擇 VM 部署？
**優點**：
- 完全控制環境
- 簡單直接的部署流程  
- 適合 WebSocket/WebTransport
- 成本可預測

**缺點**：
- 需要手動管理和維護
- 沒有自動擴展
- 需要自己處理 SSL

### 為什麼使用 Nginx？
**優點**：
- 成熟穩定的反向代理
- 優秀的靜態檔案服務
- Let's Encrypt 整合良好
- WebSocket 支援完善

**替代方案**：
- Caddy (自動 HTTPS)
- Traefik (容器化環境)

### 為什麼選擇 PostgreSQL？
**優點**：
- ACID 特性保證資料一致性
- JSON 支援適合儲存加密訊息
- 成熟的生態系統
- 優秀的並發處理

**考慮過的替代方案**：
- SQLite：太簡單，不適合生產環境
- MongoDB：對於關聯資料不夠理想

## 安全考量

### 1. 密碼管理
- ✅ 使用環境變數
- ✅ 不 commit 到 Git
- ✅ 使用強密碼
- ⚠️ 考慮使用 Secret Manager

### 2. TLS/SSL
- ✅ 強制 HTTPS
- ✅ Let's Encrypt 證書
- ✅ 自動更新證書
- ⚠️ 考慮 HSTS

### 3. 資料庫安全
- ✅ 使用強密碼
- ✅ 限制連線來源
- ⚠️ 考慮加密連線
- ⚠️ 定期備份

### 4. API 安全
- ✅ JWT 認證
- ✅ Rate limiting
- ✅ CORS 設定
- ⚠️ 考慮 API Gateway

## 性能優化建議

1. **前端優化**
   - 實施程式碼分割
   - 使用 CDN
   - 圖片優化
   - Service Worker 快取

2. **後端優化**
   - 資料庫連線池
   - Redis 快取
   - 查詢優化
   - 使用索引

3. **基礎設施**
   - 負載均衡
   - 自動擴展
   - 監控告警
   - 日誌聚合

## 未來改進方向

1. **CI/CD 自動化**
   - GitHub Actions 自動部署
   - 自動測試
   - 程式碼品質檢查

2. **容器化**
   - Docker 化應用
   - Kubernetes 部署
   - 微服務架構

3. **監控與日誌**
   - Prometheus + Grafana
   - ELK Stack
   - 錯誤追蹤 (Sentry)

4. **備份策略**
   - 自動資料庫備份
   - 異地備份
   - 災難復原計劃