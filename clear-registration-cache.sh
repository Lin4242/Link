#!/bin/bash

echo "=== 檢查並清理註冊快取問題 ==="

# 產生一個測試頁面來清理 localStorage
cat > /tmp/clear-cache.html << 'HTML'
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Clear Registration Cache</title>
    <style>
        body {
            font-family: system-ui;
            padding: 20px;
            max-width: 500px;
            margin: 0 auto;
            background: #1e293b;
            color: white;
        }
        button {
            display: block;
            width: 100%;
            padding: 15px;
            margin: 10px 0;
            background: #475569;
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            cursor: pointer;
        }
        button:hover {
            background: #64748b;
        }
        .status {
            margin: 20px 0;
            padding: 15px;
            background: #334155;
            border-radius: 8px;
            font-family: monospace;
            white-space: pre-wrap;
        }
    </style>
</head>
<body>
    <h2>Registration Cache Debug</h2>
    <div class="status" id="status">Loading...</div>
    
    <button onclick="clearCache()">Clear Registration Cache</button>
    <button onclick="checkStatus()">Check Current Status</button>
    <button onclick="window.location.href='https://link.mcphub.tw/register'">Go to Register Page</button>
    
    <script>
        function checkStatus() {
            const firstCard = localStorage.getItem('register_first_card');
            const pairedToken = localStorage.getItem('register_paired_token');
            const status = document.getElementById('status');
            
            if (firstCard || pairedToken) {
                status.innerHTML = `Found cached registration data:
First Card: ${firstCard || 'none'}
Paired Token: ${pairedToken || 'none'}

This might be blocking new registrations!`;
                status.style.background = '#7f1d1d';
            } else {
                status.innerHTML = 'No cached registration data found.\nYou should be able to register new cards.';
                status.style.background = '#14532d';
            }
        }
        
        function clearCache() {
            localStorage.removeItem('register_first_card');
            localStorage.removeItem('register_paired_token');
            // Also clear any auth tokens
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            
            const status = document.getElementById('status');
            status.innerHTML = 'Cache cleared successfully!';
            status.style.background = '#14532d';
            
            setTimeout(checkStatus, 1500);
        }
        
        // Check on load
        checkStatus();
    </script>
</body>
</html>
HTML

echo "已產生清理頁面: /tmp/clear-cache.html"
echo ""
echo "請在手機瀏覽器打開此檔案，或訪問："
echo "https://link.mcphub.tw/clear-cache.html"
echo ""
echo "=== 檢查最新卡片狀態 ==="
curl -s https://link.mcphub.tw/api/v1/auth/check-card/e894c8251988443a-1-ce6c9db0 | python3 -m json.tool