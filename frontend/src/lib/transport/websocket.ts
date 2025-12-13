import type { ITransport, EncryptedMessage } from '$lib/types';

export interface WebSocketConfig {
	url: string;
	token: string;
	maxReconnects?: number;
	reconnectDelay?: number;
}

export function createWebSocketTransport(config: WebSocketConfig): ITransport {
	const { url, token, maxReconnects = 5, reconnectDelay = 1000 } = config;

	let ws: WebSocket | null = null;
	let reconnectAttempts = 0;
	let intentionalClose = false;

	const transport: ITransport = {
		onMessage: null,
		onTyping: null,
		onOnline: null,
		onOffline: null,
		onDelivered: null,
		onDeleted: null,
		onConnected: null,

		async connect(): Promise<void> {
			return new Promise((resolve, reject) => {
				intentionalClose = false;
				const wsUrl = `${url}?token=${token.substring(0, 20)}...`;
				console.log('WebSocket connecting to:', wsUrl);
				ws = new WebSocket(`${url}?token=${token}`);

				ws.onopen = () => {
					console.log('WebSocket connection opened');
					reconnectAttempts = 0;
					transport.onConnected?.(true);
					resolve();
				};

				ws.onerror = (event) => {
					console.error('WebSocket error:', event);
					if (reconnectAttempts === 0) {
						reject(new Error('WebSocket connection failed'));
					}
				};

				ws.onclose = (event) => {
					console.log('WebSocket closed:', { code: event.code, reason: event.reason });
					transport.onConnected?.(false);
					if (!intentionalClose && reconnectAttempts < maxReconnects) {
						reconnectAttempts++;
						console.log(`WebSocket reconnecting (attempt ${reconnectAttempts}/${maxReconnects})...`);
						setTimeout(() => {
							transport.connect().catch(() => {});
						}, reconnectDelay * reconnectAttempts);
					}
				};

				ws.onmessage = (event) => {
					console.log('=== WebSocket RAW onmessage ===', event.data.substring(0, 80));
					try {
						const msg = JSON.parse(event.data);
						console.log('Parsed message type:', msg.t);
						handleMessage(msg);
					} catch (e) {
						console.error('WebSocket parse error:', e);
					}
				};
			});
		},

		disconnect(): void {
			intentionalClose = true;
			reconnectAttempts = maxReconnects;
			ws?.close();
			ws = null;
		},

		async sendMessage(to: string, encryptedContent: string, tempId: string): Promise<void> {
			if (ws?.readyState === WebSocket.OPEN) {
				ws.send(
					JSON.stringify({
						t: 'msg',
						p: { to, encrypted_content: encryptedContent, temp_id: tempId },
					})
				);
			}
		},

		sendTyping(to: string, conversationId: string): void {
			if (ws?.readyState === WebSocket.OPEN) {
				ws.send(
					JSON.stringify({
						t: 'typing',
						p: { to, conversation_id: conversationId },
					})
				);
			}
		},

		sendRead(conversationId: string, messageId: string): void {
			if (ws?.readyState === WebSocket.OPEN) {
				ws.send(
					JSON.stringify({
						t: 'read',
						p: { conversation_id: conversationId, message_id: messageId },
					})
				);
			}
		},
	};

	function handleMessage(msg: { t: string; p: unknown }): void {
		console.log('handleMessage called, type:', msg.t, 'onDelivered exists:', !!transport.onDelivered);
		switch (msg.t) {
			case 'msg':
				transport.onMessage?.(msg.p as EncryptedMessage);
				break;
			case 'typing': {
				const tp = msg.p as { conversation_id: string; from: string };
				transport.onTyping?.(tp.conversation_id, tp.from);
				break;
			}
			case 'online': {
				const op = msg.p as { user_id: string };
				transport.onOnline?.(op.user_id);
				break;
			}
			case 'offline': {
				const ofp = msg.p as { user_id: string };
				transport.onOffline?.(ofp.user_id);
				break;
			}
			case 'delivered': {
				console.log('=== DELIVERED MESSAGE RECEIVED ===');
				console.log('Raw payload:', JSON.stringify(msg.p));
				const dp = msg.p as { temp_id: string; message: EncryptedMessage };
				console.log('Parsed - temp_id:', dp.temp_id);
				console.log('onDelivered handler exists:', !!transport.onDelivered);
				if (transport.onDelivered) {
					console.log('Calling onDelivered handler NOW');
					transport.onDelivered(dp.temp_id, dp.message);
					console.log('onDelivered handler called');
				} else {
					console.error('NO onDelivered handler! Message will be lost!');
				}
				break;
			}
			case 'deleted': {
				const delp = msg.p as { id: string; conversation_id: string };
				transport.onDeleted?.(delp.id, delp.conversation_id);
				break;
			}
		}
	}

	return transport;
}
