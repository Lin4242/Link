import { test, expect } from '@playwright/test';

// Test the actual frontend chat page
test('Chat page message delivery', async ({ page }) => {
  const consoleLogs: string[] = [];
  page.on('console', msg => {
    const text = msg.text();
    consoleLogs.push(`[${msg.type()}] ${text}`);
    console.log(`[Browser ${msg.type()}] ${text}`);
  });

  // First we need to set up auth - simulate login by setting localStorage
  await page.goto('https://192.168.1.99:5173');
  await page.waitForTimeout(1000);

  // Generate a valid JWT token
  const token = await page.evaluate(async () => {
    const jwtSecret = 'adab22111811be224304bd27f82fa85b36424b9a4f2e0be16f4033a7e4e2b646';
    const userId = 'a7e5e31c-ade9-46cb-8b11-083460ae313c';

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

    // Set auth in localStorage
    localStorage.setItem('auth_token', token);
    localStorage.setItem('auth_user', JSON.stringify({
      id: userId,
      nickname: 'Test User'
    }));

    return token;
  });

  console.log('Token set:', token.substring(0, 30) + '...');

  // Navigate to chat page
  await page.goto('https://192.168.1.99:5173/chat');
  await page.waitForTimeout(3000);

  // Check console logs for transport connection
  const transportConnected = consoleLogs.some(log => log.includes('WebSocket connection opened'));
  console.log('Transport connected:', transportConnected);

  // Check for handler setup
  const handlerSet = consoleLogs.some(log => log.includes('onDelivered called') || log.includes('Setting transport.onDelivered'));
  console.log('Handler set up:', handlerSet);

  // Print all relevant logs
  console.log('\n=== Relevant Console Logs ===');
  consoleLogs.filter(log =>
    log.includes('WebSocket') ||
    log.includes('Transport') ||
    log.includes('onDelivered') ||
    log.includes('DELIVERED')
  ).forEach(log => console.log(log));
});
