import { test, expect } from '@playwright/test';

const JWT_SECRET = 'adab22111811be224304bd27f82fa85b36424b9a4f2e0be16f4033a7e4e2b646';
const USER_ID = 'a7e5e31c-ade9-46cb-8b11-083460ae313c';
const TO_USER = 'fcf454d3-d34a-4765-bc75-e9c3aa4bd9c3';
const WS_URL = 'wss://127.0.0.1:9443/ws';

test('WebSocket delivered message test', async ({ page }) => {
  // Collect console logs
  const consoleLogs: string[] = [];
  page.on('console', msg => {
    consoleLogs.push(`[${msg.type()}] ${msg.text()}`);
    console.log(`[Browser ${msg.type()}] ${msg.text()}`);
  });

  // Navigate to a simple page (crypto.subtle needs secure context)
  await page.goto('https://127.0.0.1:9443/api/v1/health').catch(() => {});
  // Wait a bit
  await page.waitForTimeout(500);

  // Run WebSocket test in browser context
  const result = await page.evaluate(async ({ jwtSecret, userId, toUser, wsUrl }) => {
    // Generate JWT token
    async function generateToken() {
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

      return header + '.' + payload + '.' + sigB64;
    }

    const logs: string[] = [];
    const log = (msg: string) => {
      logs.push(msg);
      console.log(msg);
    };

    return new Promise(async (resolve) => {
      try {
        log('Generating token...');
        const token = await generateToken();
        log('Token generated: ' + token.substring(0, 30) + '...');

        log('Connecting WebSocket to ' + wsUrl + '...');
        const ws = new WebSocket(wsUrl + '?token=' + token);

        let deliveredReceived = false;
        let tempId = '';

        ws.onopen = () => {
          log('WebSocket CONNECTED!');

          // Send test message after 500ms
          setTimeout(() => {
            tempId = 'playwright-test-' + Date.now();
            const msg = {
              t: 'msg',
              p: {
                to: toUser,
                encrypted_content: '{"nonce":"test","ciphertext":"playwright test"}',
                temp_id: tempId
              }
            };
            log('SENDING message: ' + JSON.stringify(msg));
            ws.send(JSON.stringify(msg));
          }, 500);
        };

        ws.onmessage = (event) => {
          log('RECEIVED: ' + event.data);
          try {
            const msg = JSON.parse(event.data);
            log('Message type: ' + msg.t);
            if (msg.t === 'delivered') {
              log('*** DELIVERED CONFIRMATION RECEIVED! ***');
              log('temp_id: ' + msg.p?.temp_id);
              log('message.id: ' + msg.p?.message?.id);
              deliveredReceived = true;
            }
          } catch (e) {
            log('Parse error: ' + e);
          }
        };

        ws.onerror = (e) => {
          log('WebSocket ERROR: ' + JSON.stringify(e));
        };

        ws.onclose = (e) => {
          log('WebSocket CLOSED: code=' + e.code + ' reason=' + e.reason);
        };

        // Wait up to 10 seconds for delivered confirmation
        setTimeout(() => {
          ws.close();
          resolve({
            success: deliveredReceived,
            logs: logs,
            tempId: tempId
          });
        }, 5000);

      } catch (e) {
        resolve({
          success: false,
          logs: logs,
          error: String(e)
        });
      }
    });
  }, { jwtSecret: JWT_SECRET, userId: USER_ID, toUser: TO_USER, wsUrl: WS_URL });

  console.log('\n=== Test Result ===');
  console.log('Success:', result.success);
  console.log('Logs:');
  result.logs.forEach((log: string) => console.log('  ' + log));

  expect(result.success).toBe(true);
});
