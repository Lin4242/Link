#!/bin/bash

SERVER_IP="34.136.217.56"
SERVER_USER="rocketmantw5516"

echo "檢查 update-public-key 錯誤..."
echo "================================"

ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd ~/Link

echo "1. 檢查 Nginx 錯誤日誌："
sudo tail -20 /var/log/nginx/error.log | grep -i "update\|public" || echo "沒有相關錯誤"

echo ""
echo "2. 檢查瀏覽器控制台錯誤 (從 systemd journal)："
sudo journalctl -u link-backend -n 20 --no-pager | grep -i "error\|fail" || echo "沒有後端錯誤"

echo ""
echo "3. 檢查前端構建輸出是否有錯誤："
ls -la frontend/build/update-public-key* 2>/dev/null || echo "頁面文件不存在"

echo ""
echo "4. 檢查頁面是否正確編譯："
grep -l "update-public-key" frontend/build/_app/immutable/nodes/* 2>/dev/null | head -5 || echo "未找到編譯文件"

echo ""
echo "5. 檢查實際的編譯輸出："
if [ -f frontend/src/routes/update-public-key/+page.svelte ]; then
    echo "源文件存在 ✓"
    head -30 frontend/src/routes/update-public-key/+page.svelte
else
    echo "源文件不存在 ✗"
fi

echo ""
echo "6. 測試頁面訪問："
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" https://link.mcphub.tw/update-public-key

EOF