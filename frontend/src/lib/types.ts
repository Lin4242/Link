export interface User {
	id: string;
	nickname: string;
	public_key: string;
	avatar_url?: string;
	created_at: string;
	last_seen_at?: string;
}

export interface Card {
	id: string;
	user_id: string;
	card_token: string;
	card_type: 'primary' | 'backup';
	status: 'active' | 'revoked';
	created_at: string;
	activated_at?: string;
	revoked_at?: string;
}

export interface Friendship {
	id: string;
	requester_id: string;
	addressee_id: string;
	status: 'pending' | 'accepted' | 'rejected' | 'blocked';
	created_at: string;
	updated_at: string;
}

export interface Friend extends User {
	friendship_id: string;
	friendship_status: Friendship['status'];
}

export interface Conversation {
	id: string;
	participant_ids: string[];
	last_message_at?: string;
	created_at: string;
}

export interface Message {
	id: string;
	conversation_id: string;
	sender_id: string;
	encrypted_content: string;
	created_at: string;
}

export interface DecryptedMessage extends Omit<Message, 'encrypted_content'> {
	content: string;
}

export interface EncryptedPayload {
	nonce: string;
	ciphertext: string;
}

export interface EncryptedData {
	nonce: string;
	ciphertext: string;
}

export interface EncryptedMessage {
	id: string;
	conversation_id: string;
	sender_id: string;
	encrypted_content: string;
	created_at: string;
	delivered_at?: string;
	read_at?: string;
}

export interface WTMessage {
	t: string;
	p?: unknown;
}

export interface ITransport {
	connect(): Promise<void>;
	disconnect(): void;
	sendMessage(to: string, encryptedContent: string, tempId: string): Promise<void>;
	sendTyping(to: string, conversationId: string): void;
	sendRead(conversationId: string, messageId: string): void;
	onMessage: ((msg: EncryptedMessage) => void) | null;
	onTyping: ((convId: string, userId: string) => void) | null;
	onOnline: ((userId: string) => void) | null;
	onOffline: ((userId: string) => void) | null;
	onDelivered: ((tempId: string, msg: EncryptedMessage) => void) | null;
	onDeleted: ((messageId: string, conversationId: string) => void) | null;
	onConnected: ((connected: boolean) => void) | null;
}

export interface KeyPair {
	publicKey: Uint8Array;
	secretKey: Uint8Array;
}

export interface ApiResponse<T> {
	data?: T;
	error?: {
		code: string;
		message: string;
	};
}

export interface AuthTokens {
	access_token: string;
	expires_at: string;
}

export interface RegisterRequest {
	primary_token: string;
	backup_token: string;
	password: string;
	nickname: string;
	public_key: string;
}

export interface LoginRequest {
	card_token: string;
	password: string;
}

export interface CardCheckResponse {
	status: 'unregistered' | 'registered';
	card_type?: 'primary' | 'backup';
	card_status?: 'active' | 'revoked';
	has_pair?: boolean;
}

export type TransportType = 'webtransport' | 'websocket';

export interface TransportMessage {
	type: string;
	payload: unknown;
}
