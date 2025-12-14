import type { EncryptedMessage } from '$lib/types';
import { conversationsApi } from '$lib/api';
import { decryptFromString } from '$lib/crypto';
import { keysStore } from './keys.svelte';
import { deleteMessage as deleteMessageApi } from '$lib/api/conversations';

export interface DecryptedMessageItem {
	id: string;
	conversationId: string;
	senderId: string;
	content: string;
	createdAt: string;
	deliveredAt?: string;
	readAt?: string;
	pending?: boolean;
	tempId?: string;
}

function createMessagesStore() {
	let messagesByConversation = $state<Record<string, DecryptedMessageItem[]>>({});
	let loading = $state(false);

	async function loadMessages(
		conversationId: string,
		peerPublicKey: string,
		before?: string
	): Promise<void> {
		console.log('=== loadMessages called ===', {
			conversationId,
			peerPublicKey: peerPublicKey?.substring(0, 20) + '...',
			hasSecretKey: !!keysStore.secretKey,
			secretKeyLength: keysStore.secretKey?.length
		});

		loading = true;
		const res = await conversationsApi.getMessages(conversationId, 50, before);

		console.log('API response:', {
			hasData: !!res.data,
			messageCount: res.data?.length || 0,
			error: res.error
		});

		if (res.data && keysStore.secretKey) {
			let decryptSuccessCount = 0;
			let decryptFailCount = 0;

			const decrypted = res.data
				.map((m, index) => {
					const content = decryptFromString(m.encrypted_content, peerPublicKey, keysStore.secretKey!);
					if (!content) {
						decryptFailCount++;
						console.log(`Message ${index} decrypt FAILED:`, {
							id: m.id,
							senderId: m.sender_id,
							encryptedContentPreview: m.encrypted_content?.substring(0, 50) + '...'
						});
						return null;
					}
					decryptSuccessCount++;
					return {
						id: m.id,
						conversationId: m.conversation_id,
						senderId: m.sender_id,
						content,
						createdAt: m.created_at,
					} as DecryptedMessageItem;
				})
				.filter((m): m is DecryptedMessageItem => m !== null);

			console.log('Decryption results:', { decryptSuccessCount, decryptFailCount, totalDecrypted: decrypted.length });

			const existing = messagesByConversation[conversationId] || [];
			if (before) {
				messagesByConversation[conversationId] = [...decrypted, ...existing];
			} else {
				messagesByConversation[conversationId] = decrypted;
			}
		} else {
			console.log('Skipping decryption - no data or no secret key', {
				hasData: !!res.data,
				hasSecretKey: !!keysStore.secretKey
			});
		}
		loading = false;
	}

	function addMessage(conversationId: string, msg: DecryptedMessageItem): void {
		const existing = messagesByConversation[conversationId] || [];
		messagesByConversation[conversationId] = [...existing, msg];
	}

	function addPendingMessage(
		conversationId: string,
		tempId: string,
		senderId: string,
		content: string
	): void {
		addMessage(conversationId, {
			id: tempId,
			conversationId,
			senderId,
			content,
			createdAt: new Date().toISOString(),
			pending: true,
			tempId,
		});
	}

	function confirmMessage(tempId: string, realMessage: DecryptedMessageItem): void {
		for (const convId of Object.keys(messagesByConversation)) {
			messagesByConversation[convId] = messagesByConversation[convId].map((m) =>
				m.tempId === tempId ? { ...realMessage, pending: false } : m
			);
		}
	}

	function receiveMessage(
		msg: EncryptedMessage,
		peerPublicKey: string
	): DecryptedMessageItem | null {
		console.log('📥 Receiving message:', {
			msgId: msg.id,
			senderId: msg.sender_id,
			conversationId: msg.conversation_id,
			hasSecretKey: !!keysStore.secretKey,
			peerPublicKey: peerPublicKey?.substring(0, 10) + '...'
		});
		
		if (!keysStore.secretKey) {
			console.error('❌ No secret key available');
			return null;
		}
		
		const content = decryptFromString(msg.encrypted_content, peerPublicKey, keysStore.secretKey);
		if (!content) {
			console.error('❌ Failed to decrypt message:', {
				msgId: msg.id,
				encryptedContent: JSON.stringify(msg.encrypted_content).substring(0, 100) + '...'
			});
			return null;
		}
		
		console.log('✅ Message decrypted successfully:', {
			msgId: msg.id,
			contentLength: content.length
		});
		
		const decrypted: DecryptedMessageItem = {
			id: msg.id,
			conversationId: msg.conversation_id,
			senderId: msg.sender_id,
			content,
			createdAt: msg.created_at,
			deliveredAt: msg.delivered_at,
			readAt: msg.read_at,
		};
		addMessage(msg.conversation_id, decrypted);
		return decrypted;
	}

	function getMessages(conversationId: string): DecryptedMessageItem[] {
		return messagesByConversation[conversationId] || [];
	}

	function clear(): void {
		messagesByConversation = {};
	}

	async function deleteMessage(conversationId: string, messageId: string): Promise<boolean> {
		const res = await deleteMessageApi(messageId);
		if (res.error) {
			console.error('Failed to delete message:', res.error);
			return false;
		}
		// Remove from local store
		removeMessage(conversationId, messageId);
		return true;
	}

	function removeMessage(conversationId: string, messageId: string): void {
		const existing = messagesByConversation[conversationId];
		if (existing) {
			messagesByConversation[conversationId] = existing.filter((m) => m.id !== messageId);
		}
	}

	return {
		get messagesByConversation() {
			return messagesByConversation;
		},
		get loading() {
			return loading;
		},
		loadMessages,
		addMessage,
		addPendingMessage,
		confirmMessage,
		receiveMessage,
		getMessages,
		deleteMessage,
		removeMessage,
		clear,
	};
}

export const messagesStore = createMessagesStore();
