# LINK 故障排除指南

## 快速診斷

### 🔴 網站完全無法訪問
```bash
# 1. 檢查服務狀態
ssh rocketmantw5516@34.136.217.56
sudo systemctl status link-backend
sudo systemctl status nginx

# 2. 如果服務停止
sudo systemctl restart link-backend
sudo systemctl restart nginx

# 3. 檢查防火牆
sudo ufw status
```

### 🟡 API 回應 502 Bad Gateway
```bash
# 1. 檢查後端是否運行
sudo systemctl status link-backend

# 2. 查看錯誤日誌
sudo journalctl -u link-backend -n 50

# 3. 檢查 .env 設定
cat ~/Link/backend/.env

# 4. 重啟服務
sudo systemctl restart link-backend
```

### 🟡 WebSocket 連線失敗
```bash
# 1. 檢查 Nginx 配置
sudo nginx -t
cat /etc/nginx/sites-available/link | grep -A5 "/ws"

# 2. 檢查 CORS 設定
grep CORS ~/Link/backend/.env

# 3. 重新載入 Nginx
sudo systemctl reload nginx
```

## 常見問題

### 1. Admin Panel 密碼錯誤
**症狀**：輸入密碼後顯示「密碼錯誤」

**檢查步驟**：
```bash
# 檢查環境變數
grep ADMIN_PASSWORD ~/Link/backend/.env

# 確認服務有載入新設定
sudo systemctl restart link-backend

# 測試 API
curl -X POST https://link.mcphub.tw/api/v1/admin/cards/generate \
  -H "X-Admin-Password: YOUR_PASSWORD" \
  -H "Content-Type: application/json"
```

### 2. 資料庫連線失敗
**症狀**：`failed to connect to database`

**檢查步驟**：
```bash
# 檢查 PostgreSQL 狀態
sudo systemctl status postgresql

# 測試連線
psql -U link -d link -h localhost

# 檢查連線字串
grep DATABASE_URL ~/Link/backend/.env
```

**解決方案**：
```bash
# 重啟資料庫
sudo systemctl restart postgresql

# 重設密碼（如需要）
sudo -u postgres psql
ALTER USER link WITH PASSWORD 'new_password';
\q
```

### 3. 前端頁面空白
**症狀**：訪問網站只看到空白頁

**檢查步驟**：
```bash
# 檢查前端檔案
ls -la ~/Link/frontend/build/

# 檢查 Nginx 錯誤
sudo tail -f /var/log/nginx/error.log

# 重新建置前端
cd ~/Link/frontend
pnpm build
```

### 4. Git Pull 失敗
**症狀**：`Your local changes would be overwritten`

**解決方案**：
```bash
cd ~/Link

# 方法 1：保存本地變更
git stash
git pull origin main
git stash pop

# 方法 2：放棄本地變更
git reset --hard HEAD
git pull origin main

# 方法 3：使用修復腳本
./fix-server-conflicts.sh
```

### 5. 證書過期
**症狀**：瀏覽器顯示證書錯誤

**解決方案**：
```bash
# 更新證書
sudo certbot renew

# 強制更新
sudo certbot certonly --nginx -d link.mcphub.tw --force-renewal

# 重啟 Nginx
sudo systemctl restart nginx
```

## 日誌位置

```bash
# 後端日誌
sudo journalctl -u link-backend -f

# Nginx 訪問日誌
sudo tail -f /var/log/nginx/access.log

# Nginx 錯誤日誌
sudo tail -f /var/log/nginx/error.log

# PostgreSQL 日誌
sudo tail -f /var/log/postgresql/postgresql-*.log

# 系統日誌
sudo tail -f /var/log/syslog
```

## 效能問題

### 回應緩慢
```bash
# 檢查 CPU 和記憶體
top
htop

# 檢查磁碟空間
df -h

# 檢查資料庫連線數
sudo -u postgres psql -c "SELECT count(*) FROM pg_stat_activity;"

# 檢查慢查詢
sudo -u postgres psql -d link -c "SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 5;"
```

### 記憶體不足
```bash
# 清理不必要的檔案
cd ~/Link/frontend
rm -rf node_modules .svelte-kit
pnpm install
pnpm build

# 清理 Docker（如有使用）
docker system prune -a

# 重啟服務釋放記憶體
sudo systemctl restart link-backend
sudo systemctl restart postgresql
```

## 緊急復原

### 完整重新部署
```bash
# 1. 備份重要資料
cd ~
cp ~/Link/backend/.env ~/env-backup.txt
pg_dump -U link link > ~/link-backup.sql

# 2. 重新部署
cd ~/Link
git fetch origin
git reset --hard origin/main
./deploy-from-github.sh

# 3. 恢復設定
cp ~/env-backup.txt ~/Link/backend/.env
sudo systemctl restart link-backend
```

### 資料庫復原
```bash
# 從備份復原
psql -U link link < ~/link-backup.sql

# 重建資料庫（會清除所有資料！）
sudo -u postgres psql
DROP DATABASE link;
CREATE DATABASE link OWNER link;
\q
cd ~/Link/backend
psql -U link -d link -f migrations/001_init.up.sql
```

## 監控檢查

### 健康檢查腳本
```bash
#!/bin/bash
# health-check.sh

echo "=== LINK Health Check ==="
echo ""

# 1. 服務狀態
echo "1. Service Status:"
systemctl is-active link-backend
systemctl is-active nginx
systemctl is-active postgresql

# 2. API 健康
echo ""
echo "2. API Health:"
curl -s https://link.mcphub.tw/health || echo "API Failed"

# 3. 磁碟空間
echo ""
echo "3. Disk Space:"
df -h / | tail -1

# 4. 記憶體使用
echo ""
echo "4. Memory Usage:"
free -h | grep Mem

# 5. 最近錯誤
echo ""
echo "5. Recent Errors:"
sudo journalctl -u link-backend -p err -n 5 --no-pager
```

## 聯絡支援

如果以上方法都無法解決問題：

1. 收集診斷資訊：
```bash
./health-check.sh > diagnosis.txt
sudo journalctl -u link-backend -n 100 >> diagnosis.txt
```

2. 檢查 GitHub Issues：
https://github.com/Lin4242/Link/issues

3. 記錄問題詳情：
- 發生時間
- 錯誤訊息
- 最近的變更
- 嘗試過的解決方法