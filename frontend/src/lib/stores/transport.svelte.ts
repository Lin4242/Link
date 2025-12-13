import type { ITransport, EncryptedMessage } from '$lib/types';
import { createTransport, type TransportType } from '$lib/transport';

const WS_URL = import.meta.env.VITE_WS_URL || 'wss://192.168.1.99:9443/ws';
const WT_URL = import.meta.env.VITE_WT_URL || '';

function createTransportStore() {
	let connected = $state(false);
	let transportType = $state<TransportType>('websocket');
	let connecting = $state(false);

	// IMPORTANT: Don't use $state for transport - it creates a proxy that breaks handler assignment
	let transport: ITransport | null = null;

	async function connect(token: string): Promise<void> {
		console.log('Transport store: connect called', {
			connecting,
			hasTransport: !!transport,
			hasToken: !!token,
			wsUrl: WS_URL
		});

		if (connecting) {
			console.log('Transport store: already connecting, skipping');
			return;
		}

		// If already connected, skip
		if (transport && connected) {
			console.log('Transport store: already connected, skipping');
			return;
		}

		// If transport exists but not connected, disconnect and retry
		if (transport && !connected) {
			console.log('Transport store: transport exists but not connected, retrying...');
			transport.disconnect();
			transport = null;
		}

		connecting = true;
		console.log('Transport store: starting connection...');

		try {
			const result = await createTransport({
				wsUrl: WS_URL,
				wtUrl: WT_URL || undefined,
				token,
				preferredTransport: 'auto',
			});

			console.log('Transport store: transport created, type:', result.type);
			transport = result.transport;
			transportType = result.type;

			// Attach any handlers that were registered before transport was created
			attachHandlers();

			transport.onConnected = (isConnected) => {
				console.log('Transport store: onConnected callback, connected:', isConnected);
				connected = isConnected;
			};

			// If createTransport didn't already connect (for WebSocket path)
			if (!connected) {
				console.log('Transport store: calling transport.connect()...');
				await transport.connect();
				// Force set connected after successful connect
				connected = true;
				console.log('Transport store: connect() completed, connected:', connected);
			}
		} catch (e) {
			console.error('Transport store: connection error:', e);
			throw e;
		} finally {
			connecting = false;
		}
	}

	function disconnect(): void {
		transport?.disconnect();
		transport = null;
		connected = false;
	}

	// Store handlers so they can be attached when transport is created
	let messageHandler: ((msg: EncryptedMessage) => void) | null = null;
	let typingHandler: ((convId: string, userId: string) => void) | null = null;
	let onlineHandler: ((userId: string) => void) | null = null;
	let offlineHandler: ((userId: string) => void) | null = null;
	let deliveredHandler: ((tempId: string, msg: EncryptedMessage) => void) | null = null;

	function attachHandlers() {
		if (transport) {
			if (messageHandler) transport.onMessage = messageHandler;
			if (typingHandler) transport.onTyping = typingHandler;
			if (onlineHandler) transport.onOnline = onlineHandler;
			if (offlineHandler) transport.onOffline = offlineHandler;
			if (deliveredHandler) transport.onDelivered = deliveredHandler;
		}
	}

	function onMessage(handler: (msg: EncryptedMessage) => void): void {
		messageHandler = handler;
		if (transport) {
			transport.onMessage = handler;
		}
	}

	function onTyping(handler: (convId: string, userId: string) => void): void {
		typingHandler = handler;
		if (transport) {
			transport.onTyping = handler;
		}
	}

	function onOnline(handler: (userId: string) => void): void {
		onlineHandler = handler;
		if (transport) {
			transport.onOnline = handler;
		}
	}

	function onOffline(handler: (userId: string) => void): void {
		offlineHandler = handler;
		if (transport) {
			transport.onOffline = handler;
		}
	}

	function onDelivered(handler: (tempId: string, msg: EncryptedMessage) => void): void {
		console.log('transportStore.onDelivered called, transport exists:', !!transport);
		deliveredHandler = handler;
		if (transport) {
			console.log('Setting transport.onDelivered directly');
			transport.onDelivered = handler;
		}
	}

	async function sendMessage(to: string, encryptedContent: string, tempId: string): Promise<void> {
		await transport?.sendMessage(to, encryptedContent, tempId);
	}

	function sendTyping(to: string, conversationId: string): void {
		transport?.sendTyping(to, conversationId);
	}

	function sendRead(conversationId: string, messageId: string): void {
		transport?.sendRead(conversationId, messageId);
	}

	return {
		get connected() {
			return connected;
		},
		get transport() {
			return transport;
		},
		get transportType() {
			return transportType;
		},
		get connecting() {
			return connecting;
		},
		connect,
		disconnect,
		onMessage,
		onTyping,
		onOnline,
		onOffline,
		onDelivered,
		sendMessage,
		sendTyping,
		sendRead,
	};
}

export const transportStore = createTransportStore();
