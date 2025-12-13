#!/usr/bin/env python3
import asyncio
import json
import ssl
import websockets

# Test user credentials - we'll need to get a token first
WS_URL = "wss://127.0.0.1:9443/ws"

async def test_websocket(token):
    ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
    ssl_context.check_hostname = False
    ssl_context.verify_mode = ssl.CERT_NONE

    url = f"{WS_URL}?token={token}"
    print(f"Connecting to WebSocket...")

    async with websockets.connect(url, ssl=ssl_context) as ws:
        print("Connected!")

        # Send a test message
        msg = {
            "t": "msg",
            "p": {
                "to": "fcf454d3-d34a-4765-bc75-e9c3aa4bd9c3",
                "encrypted_content": '{"nonce":"test","ciphertext":"test"}',
                "temp_id": "test-temp-id-123"
            }
        }
        print(f"Sending: {json.dumps(msg)}")
        await ws.send(json.dumps(msg))

        # Wait for response
        print("Waiting for response...")
        try:
            response = await asyncio.wait_for(ws.recv(), timeout=5.0)
            print(f"Received: {response}")
        except asyncio.TimeoutError:
            print("No response received within 5 seconds")

if __name__ == "__main__":
    import sys
    if len(sys.argv) < 2:
        print("Usage: python3 test_ws.py <token>")
        sys.exit(1)

    token = sys.argv[1]
    asyncio.run(test_websocket(token))
