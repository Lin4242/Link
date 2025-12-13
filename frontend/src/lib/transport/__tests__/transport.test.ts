import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { isWebTransportSupported } from '../webtransport';
import { getBestTransportType } from '../index';

describe('Transport 模組', () => {
	describe('isWebTransportSupported', () => {
		const originalWebTransport = globalThis.WebTransport;

		afterEach(() => {
			if (originalWebTransport) {
				globalThis.WebTransport = originalWebTransport;
			} else {
				// @ts-expect-error - 清理測試環境
				delete globalThis.WebTransport;
			}
		});

		it('當 WebTransport 可用時應返回 true', () => {
			// @ts-expect-error - 模擬 WebTransport
			globalThis.WebTransport = class MockWebTransport {};
			expect(isWebTransportSupported()).toBe(true);
		});

		it('當 WebTransport 不可用時應返回 false', () => {
			// @ts-expect-error - 移除 WebTransport
			delete globalThis.WebTransport;
			expect(isWebTransportSupported()).toBe(false);
		});
	});

	describe('getBestTransportType', () => {
		const originalWebTransport = globalThis.WebTransport;

		afterEach(() => {
			if (originalWebTransport) {
				globalThis.WebTransport = originalWebTransport;
			} else {
				// @ts-expect-error - 清理測試環境
				delete globalThis.WebTransport;
			}
		});

		it('當沒有 WebTransport URL 時應返回 websocket', () => {
			expect(getBestTransportType()).toBe('websocket');
			expect(getBestTransportType(undefined)).toBe('websocket');
		});

		it('當有 WebTransport URL 但瀏覽器不支援時應返回 websocket', () => {
			// @ts-expect-error - 移除 WebTransport
			delete globalThis.WebTransport;
			expect(getBestTransportType('https://example.com/wt')).toBe('websocket');
		});

		it('當有 WebTransport URL 且瀏覽器支援時應返回 webtransport', () => {
			// @ts-expect-error - 模擬 WebTransport
			globalThis.WebTransport = class MockWebTransport {};
			expect(getBestTransportType('https://example.com/wt')).toBe('webtransport');
		});
	});

	describe('WebSocket Transport', () => {
		let mockWsInstance: {
			onopen: (() => void) | null;
			onclose: (() => void) | null;
			onmessage: ((event: { data: string }) => void) | null;
			onerror: ((event: unknown) => void) | null;
			readyState: number;
			send: ReturnType<typeof vi.fn>;
			close: ReturnType<typeof vi.fn>;
		};
		let originalWebSocket: typeof WebSocket;

		beforeEach(() => {
			mockWsInstance = {
				onopen: null,
				onclose: null,
				onmessage: null,
				onerror: null,
				readyState: 1, // OPEN
				send: vi.fn(),
				close: vi.fn(),
			};

			originalWebSocket = globalThis.WebSocket;

			// @ts-expect-error - 模擬 WebSocket 建構函數
			globalThis.WebSocket = class MockWebSocket {
				static OPEN = 1;
				constructor() {
					Object.assign(this, mockWsInstance);
					return mockWsInstance as unknown as WebSocket;
				}
			};
		});

		afterEach(() => {
			globalThis.WebSocket = originalWebSocket;
			vi.restoreAllMocks();
		});

		it('應該能建立 WebSocket 連線', async () => {
			const { createWebSocketTransport } = await import('../websocket');

			const transport = createWebSocketTransport({
				url: 'wss://test.com/ws',
				token: 'test-token',
			});

			const connectPromise = transport.connect();

			// 模擬連線成功
			mockWsInstance.onopen?.();

			await expect(connectPromise).resolves.toBeUndefined();
		});

		it('應該能發送訊息', async () => {
			const { createWebSocketTransport } = await import('../websocket');

			const transport = createWebSocketTransport({
				url: 'wss://test.com/ws',
				token: 'test-token',
			});

			const connectPromise = transport.connect();
			mockWsInstance.onopen?.();
			await connectPromise;

			await transport.sendMessage('user123', 'encrypted-content', 'temp-123');

			expect(mockWsInstance.send).toHaveBeenCalledWith(
				JSON.stringify({
					t: 'msg',
					p: { to: 'user123', encrypted_content: 'encrypted-content', temp_id: 'temp-123' },
				})
			);
		});

		it('應該能發送輸入中狀態', async () => {
			const { createWebSocketTransport } = await import('../websocket');

			const transport = createWebSocketTransport({
				url: 'wss://test.com/ws',
				token: 'test-token',
			});

			const connectPromise = transport.connect();
			mockWsInstance.onopen?.();
			await connectPromise;

			transport.sendTyping('user123', 'conv-456');

			expect(mockWsInstance.send).toHaveBeenCalledWith(
				JSON.stringify({
					t: 'typing',
					p: { to: 'user123', conversation_id: 'conv-456' },
				})
			);
		});

		it('應該能發送已讀回條', async () => {
			const { createWebSocketTransport } = await import('../websocket');

			const transport = createWebSocketTransport({
				url: 'wss://test.com/ws',
				token: 'test-token',
			});

			const connectPromise = transport.connect();
			mockWsInstance.onopen?.();
			await connectPromise;

			transport.sendRead('conv-456', 'msg-789');

			expect(mockWsInstance.send).toHaveBeenCalledWith(
				JSON.stringify({
					t: 'read',
					p: { conversation_id: 'conv-456', message_id: 'msg-789' },
				})
			);
		});

		it('應該能處理收到的訊息', async () => {
			const { createWebSocketTransport } = await import('../websocket');

			const transport = createWebSocketTransport({
				url: 'wss://test.com/ws',
				token: 'test-token',
			});

			const messageHandler = vi.fn();
			transport.onMessage = messageHandler;

			const connectPromise = transport.connect();
			mockWsInstance.onopen?.();
			await connectPromise;

			const testMessage = {
				id: 'msg-1',
				conversation_id: 'conv-1',
				sender_id: 'user-1',
				encrypted_content: 'encrypted',
				created_at: '2024-01-01T00:00:00Z',
			};

			mockWsInstance.onmessage?.({
				data: JSON.stringify({ t: 'msg', p: testMessage }),
			});

			expect(messageHandler).toHaveBeenCalledWith(testMessage);
		});

		it('應該能處理輸入中事件', async () => {
			const { createWebSocketTransport } = await import('../websocket');

			const transport = createWebSocketTransport({
				url: 'wss://test.com/ws',
				token: 'test-token',
			});

			const typingHandler = vi.fn();
			transport.onTyping = typingHandler;

			const connectPromise = transport.connect();
			mockWsInstance.onopen?.();
			await connectPromise;

			mockWsInstance.onmessage?.({
				data: JSON.stringify({ t: 'typing', p: { conversation_id: 'conv-1', from: 'user-1' } }),
			});

			expect(typingHandler).toHaveBeenCalledWith('conv-1', 'user-1');
		});

		it('應該能處理上線事件', async () => {
			const { createWebSocketTransport } = await import('../websocket');

			const transport = createWebSocketTransport({
				url: 'wss://test.com/ws',
				token: 'test-token',
			});

			const onlineHandler = vi.fn();
			transport.onOnline = onlineHandler;

			const connectPromise = transport.connect();
			mockWsInstance.onopen?.();
			await connectPromise;

			mockWsInstance.onmessage?.({
				data: JSON.stringify({ t: 'online', p: { user_id: 'user-1' } }),
			});

			expect(onlineHandler).toHaveBeenCalledWith('user-1');
		});

		it('應該能處理離線事件', async () => {
			const { createWebSocketTransport } = await import('../websocket');

			const transport = createWebSocketTransport({
				url: 'wss://test.com/ws',
				token: 'test-token',
			});

			const offlineHandler = vi.fn();
			transport.onOffline = offlineHandler;

			const connectPromise = transport.connect();
			mockWsInstance.onopen?.();
			await connectPromise;

			mockWsInstance.onmessage?.({
				data: JSON.stringify({ t: 'offline', p: { user_id: 'user-1' } }),
			});

			expect(offlineHandler).toHaveBeenCalledWith('user-1');
		});

		it('應該能斷開連線', async () => {
			const { createWebSocketTransport } = await import('../websocket');

			const transport = createWebSocketTransport({
				url: 'wss://test.com/ws',
				token: 'test-token',
			});

			const connectPromise = transport.connect();
			mockWsInstance.onopen?.();
			await connectPromise;

			transport.disconnect();

			expect(mockWsInstance.close).toHaveBeenCalled();
		});

		it('應該忽略無效的 JSON 訊息', async () => {
			const { createWebSocketTransport } = await import('../websocket');

			const transport = createWebSocketTransport({
				url: 'wss://test.com/ws',
				token: 'test-token',
			});

			const messageHandler = vi.fn();
			transport.onMessage = messageHandler;

			const connectPromise = transport.connect();
			mockWsInstance.onopen?.();
			await connectPromise;

			// 發送無效 JSON
			mockWsInstance.onmessage?.({
				data: 'not valid json',
			});

			expect(messageHandler).not.toHaveBeenCalled();
		});
	});
});
