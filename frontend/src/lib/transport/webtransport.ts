import type { ITransport, EncryptedMessage } from '$lib/types';

export interface WebTransportConfig {
	url: string;
	token: string;
	maxReconnects?: number;
	reconnectDelay?: number;
}

// Check if WebTransport is available in the browser
export function isWebTransportSupported(): boolean {
	return typeof WebTransport !== 'undefined';
}

export function createWebTransportClient(config: WebTransportConfig): ITransport {
	const { url, token, maxReconnects = 5, reconnectDelay = 1000 } = config;

	let wt: WebTransport | null = null;
	let writer: WritableStreamDefaultWriter<Uint8Array> | null = null;
	let reconnectAttempts = 0;
	let intentionalClose = false;

	const encoder = new TextEncoder();
	const decoder = new TextDecoder();

	const transport: ITransport = {
		onMessage: null,
		onTyping: null,
		onOnline: null,
		onOffline: null,
		onDelivered: null,
		onConnected: null,

		async connect(): Promise<void> {
			if (!isWebTransportSupported()) {
				throw new Error('WebTransport is not supported in this browser');
			}

			intentionalClose = false;

			// Append token to URL
			const wtUrl = new URL(url);
			wtUrl.searchParams.set('token', token);

			wt = new WebTransport(wtUrl.toString());

			try {
				await wt.ready;
				reconnectAttempts = 0;
				transport.onConnected?.(true);

				// Get a bidirectional stream for communication
				const stream = await wt.createBidirectionalStream();
				writer = stream.writable.getWriter();

				// Start reading from the stream
				readStream(stream.readable.getReader());

				// Handle connection close
				wt.closed
					.then(() => {
						handleClose();
					})
					.catch(() => {
						handleClose();
					});
			} catch (error) {
				if (reconnectAttempts === 0) {
					throw error;
				}
				handleClose();
			}
		},

		disconnect(): void {
			intentionalClose = true;
			reconnectAttempts = maxReconnects;
			writer?.close().catch(() => {});
			wt?.close();
			wt = null;
			writer = null;
		},

		async sendMessage(to: string, encryptedContent: string, tempId: string): Promise<void> {
			await send({
				t: 'msg',
				p: { to, encrypted_content: encryptedContent, temp_id: tempId },
			});
		},

		sendTyping(to: string, conversationId: string): void {
			send({
				t: 'typing',
				p: { to, conversation_id: conversationId },
			}).catch(() => {});
		},

		sendRead(conversationId: string, messageId: string): void {
			send({
				t: 'read',
				p: { conversation_id: conversationId, message_id: messageId },
			}).catch(() => {});
		},
	};

	async function send(data: object): Promise<void> {
		if (!writer) return;

		const json = JSON.stringify(data);
		const bytes = encoder.encode(json + '\n'); // Newline-delimited JSON
		await writer.write(bytes);
	}

	async function readStream(reader: ReadableStreamDefaultReader<Uint8Array>): Promise<void> {
		let buffer = '';

		try {
			while (true) {
				const { done, value } = await reader.read();
				if (done) break;

				buffer += decoder.decode(value, { stream: true });

				// Process complete messages (newline-delimited)
				const lines = buffer.split('\n');
				buffer = lines.pop() || '';

				for (const line of lines) {
					if (line.trim()) {
						try {
							const msg = JSON.parse(line);
							handleMessage(msg);
						} catch {
							// Ignore parse errors
						}
					}
				}
			}
		} catch {
			// Stream closed or error
		}
	}

	function handleMessage(msg: { t: string; p: unknown }): void {
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
				const dp = msg.p as { temp_id: string; message: EncryptedMessage };
				transport.onDelivered?.(dp.temp_id, dp.message);
				break;
			}
		}
	}

	function handleClose(): void {
		transport.onConnected?.(false);
		writer = null;
		wt = null;

		if (!intentionalClose && reconnectAttempts < maxReconnects) {
			reconnectAttempts++;
			setTimeout(() => {
				transport.connect().catch(() => {});
			}, reconnectDelay * reconnectAttempts);
		}
	}

	return transport;
}
