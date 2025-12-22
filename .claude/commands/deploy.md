# 部署流程

執行完整的前後端部署流程。每一步都要確認成功才能繼續。

## 步驟 1: Git 檢查
```bash
git status
git diff --stat HEAD~1
```
確認：
- 沒有未提交的重要修改
- 知道這次部署包含什麼變更

## 步驟 2: Push 到 GitHub
```bash
git push origin main
```

## 步驟 3: 前端 Build
```bash
cd /Users/jimmy/project/Link/frontend
PATH="/opt/homebrew/bin:$PATH" pnpm build
```
確認：build 成功，沒有錯誤

## 步驟 4: 檢查 Build 產物
```bash
ls -la /Users/jimmy/project/Link/frontend/build/_app/immutable/assets/
```
記下 CSS 檔案的 hash（如 `0.D4uDu4_Q.css`），部署後要確認伺服器上的 hash 相同

## 步驟 5: 上傳前端到伺服器
```bash
scp -r /Users/jimmy/project/Link/frontend/build/* jimmy@link.mcphub.tw:/tmp/link-frontend/
```

## 步驟 6: 部署前端
```bash
ssh jimmy@link.mcphub.tw 'sudo rm -rf /home/rocketmantw5516/Link/frontend/build/* && sudo cp -r /tmp/link-frontend/* /home/rocketmantw5516/Link/frontend/build/'
```

## 步驟 7: 確認前端部署成功
```bash
ssh jimmy@link.mcphub.tw 'ls /home/rocketmantw5516/Link/frontend/build/_app/immutable/assets/'
```
確認：hash 與步驟 4 相同

## 步驟 8: 後端 Pull + Build + Restart
```bash
ssh jimmy@link.mcphub.tw 'cd /home/rocketmantw5516/Link && sudo -u rocketmantw5516 git pull && cd backend && sudo -u rocketmantw5516 /usr/local/go/bin/go build -o bin/server ./cmd/server && sudo systemctl restart link-backend'
```

## 步驟 9: 確認後端運行正常
```bash
ssh jimmy@link.mcphub.tw 'sudo systemctl status link-backend --no-pager'
```
確認：Active: active (running)

## 步驟 10: 健康檢查
```bash
curl -s https://link.mcphub.tw/health
```
確認：返回正常

## 完成
告訴用戶部署完成，並列出：
- 前端 CSS hash
- 後端狀態
- 健康檢查結果
