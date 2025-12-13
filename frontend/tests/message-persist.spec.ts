import { test, expect } from '@playwright/test';

const JWT_SECRET = 'adab22111811be224304bd27f82fa85b36424b9a4f2e0be16f4033a7e4e2b646';
const USER_ID = 'a7e5e31c-ade9-46cb-8b11-083460ae313c';
const BASE_URL = 'https://192.168.1.99:5173';
const TEST_PASSWORD = 'test123';

test('Send message and verify persistence', async ({ page, context }) => {
  // Collect all console logs
  page.on('console', msg => {
    console.log(`[Browser] ${msg.text()}`);
  });

  // Handle dialogs
  page.on('dialog', async dialog => {
    console.log('Dialog:', dialog.message());
    await dialog.accept();
  });

  // Step 1: Setup auth
  console.log('\n=== Step 1: Setting up auth ===');
  await page.goto(BASE_URL);
  await page.waitForTimeout(1000);

  await page.evaluate(async ({ jwtSecret, userId }) => {
    const header = btoa(JSON.stringify({alg: 'HS256', typ: 'JWT'})).replace(/=/g, '');
    const payload = btoa(JSON.stringify({
      uid: userId,
      exp: Math.floor(Date.now()/1000) + 3600,
      iat: Math.floor(Date.now()/1000),
      nbf: Math.floor(Date.now()/1000)
    })).replace(/=/g, '');

    const enc = new TextEncoder();
    const key = await crypto.subtle.importKey('raw', enc.encode(jwtSecret), {name: 'HMAC', hash: 'SHA-256'}, false, ['sign']);
    const sig = await crypto.subtle.sign('HMAC', key, enc.encode(header + '.' + payload));
    const sigB64 = btoa(String.fromCharCode(...new Uint8Array(sig))).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');

    const token = header + '.' + payload + '.' + sigB64;
    localStorage.setItem('link_auth', JSON.stringify({
      token: token,
      user: { id: userId, nickname: 'Test User' }
    }));
  }, { jwtSecret: JWT_SECRET, userId: USER_ID });

  // Step 2: Go to chat
  console.log('\n=== Step 2: Navigate to chat ===');
  await page.goto(BASE_URL + '/chat');
  await page.waitForTimeout(2000);

  // Step 3: Unlock key
  console.log('\n=== Step 3: Unlock key ===');
  const unlockModal = page.locator('text=解鎖加密金鑰');
  if (await unlockModal.isVisible()) {
    await page.fill('input[name="unlockPwd"]', TEST_PASSWORD);
    await page.click('button[type="submit"]');
    await page.waitForTimeout(3000);
  }

  // Step 4: Select conversation with 小安
  console.log('\n=== Step 4: Select conversation ===');
  const convButton = page.locator('button:has-text("小安")');
  if (await convButton.count() > 0) {
    await convButton.first().click();
    await page.waitForTimeout(2000);
  }

  // Step 5: Send a test message
  console.log('\n=== Step 5: Send test message ===');
  const testMessage = `Test ${Date.now()}`;
  await page.fill('input[placeholder="輸入訊息..."]', testMessage);
  await page.click('button[aria-label="發送訊息"]');
  await page.waitForTimeout(2000);

  // Verify message appears
  const sentMessage = page.locator(`.break-words:has-text("${testMessage}")`);
  const messageVisible = await sentMessage.isVisible().catch(() => false);
  console.log('Message sent and visible:', messageVisible);

  // Take screenshot
  await page.screenshot({ path: '/tmp/after-send.png' });

  // Step 6: Refresh page
  console.log('\n=== Step 6: Refresh page ===');
  await page.reload();
  await page.waitForTimeout(2000);

  // Step 7: Unlock key again (same password)
  console.log('\n=== Step 7: Unlock key again ===');
  const unlockModal2 = page.locator('text=解鎖加密金鑰');
  if (await unlockModal2.isVisible()) {
    await page.fill('input[name="unlockPwd"]', TEST_PASSWORD);
    await page.click('button[type="submit"]');
    await page.waitForTimeout(3000);
  }

  // Step 8: Select conversation again
  console.log('\n=== Step 8: Select conversation again ===');
  const convButton2 = page.locator('button:has-text("小安")');
  if (await convButton2.count() > 0) {
    await convButton2.first().click();
    await page.waitForTimeout(2000);
  }

  // Step 9: Check if message persisted
  console.log('\n=== Step 9: Check message persistence ===');
  const persistedMessage = page.locator(`.break-words:has-text("${testMessage}")`);
  const persisted = await persistedMessage.isVisible().catch(() => false);
  console.log('Message persisted after refresh:', persisted);

  // Take screenshot
  await page.screenshot({ path: '/tmp/after-refresh.png' });

  // Print final result
  console.log('\n=== RESULT ===');
  console.log('Message sent:', messageVisible);
  console.log('Message persisted:', persisted);

  expect(persisted).toBe(true);
});
