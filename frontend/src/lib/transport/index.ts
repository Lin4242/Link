import type { ITransport } from '$lib/types';
import { createWebSocketTransport, type WebSocketConfig } from './websocket';
import { createWebTransportClient, isWebTransportSupported, type WebTransportConfig } from './webtransport';

export { createWebSocketTransport, type WebSocketConfig } from './websocket';
export { createWebTransportClient, isWebTransportSupported, type WebTransportConfig } from './webtransport';

export type TransportType = 'websocket' | 'webtransport' | 'auto';

export interface TransportConfig {
	wsUrl: string;
	wtUrl?: string;
	token: string;
	preferredTransport?: TransportType;
	maxReconnects?: number;
	reconnectDelay?: number;
}

export interface TransportResult {
	transport: ITransport;
	type: TransportType;
}

/**
 * Create a transport with automatic fallback.
 * If preferredTransport is 'auto' or 'webtransport', it will try WebTransport first
 * and fall back to WebSocket if not supported or connection fails.
 */
export async function createTransport(config: TransportConfig): Promise<TransportResult> {
	const {
		wsUrl,
		wtUrl,
		token,
		preferredTransport = 'auto',
		maxReconnects,
		reconnectDelay,
	} = config;

	// If WebSocket is explicitly preferred, use it directly
	if (preferredTransport === 'websocket') {
		const transport = createWebSocketTransport({
			url: wsUrl,
			token,
			maxReconnects,
			reconnectDelay,
		});
		return { transport, type: 'websocket' };
	}

	// Try WebTransport if supported and URL is provided
	if (wtUrl && isWebTransportSupported()) {
		console.log('Trying WebTransport...', { wtUrl });
		try {
			const transport = createWebTransportClient({
				url: wtUrl,
				token,
				maxReconnects,
				reconnectDelay,
			});
			await transport.connect();
			console.log('WebTransport connected successfully');
			return { transport, type: 'webtransport' };
		} catch (e) {
			// WebTransport failed, fall back to WebSocket
			console.info('WebTransport connection failed, falling back to WebSocket', e);
		}
	}

	// Fall back to WebSocket
	console.log('Using WebSocket transport...', { wsUrl });
	const transport = createWebSocketTransport({
		url: wsUrl,
		token,
		maxReconnects,
		reconnectDelay,
	});
	return { transport, type: 'websocket' };
}

/**
 * Get the best available transport type without connecting
 */
export function getBestTransportType(wtUrl?: string): TransportType {
	if (wtUrl && isWebTransportSupported()) {
		return 'webtransport';
	}
	return 'websocket';
}
