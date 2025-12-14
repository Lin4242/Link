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
	decryptFailed?: boolean;
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
						// 顯示解密失敗的訊息，而不是過濾掉
						return {
							id: m.id,
							conversationId: m.conversation_id,
							senderId: m.sender_id,
							content: '[無法解密此訊息]',
							createdAt: m.created_at,
							decryptFailed: true,
						} as DecryptedMessageItem;
					}
					decryptSuccessCount++;
					return {
						id: m.id,
						conversationId: m.conversation_id,
						senderId: m.sender_id,
						content,
						createdAt: m.created_at,
					} as DecryptedMessageItem;
				});

			console.log('Decryption results:', { decryptSuccessCount, decryptFailCount, totalDecrypted: decrypted.length });

			const existing = messagesByConversation[conversationId] || [];
			// FORCE REACTIVITY - must create new object
			if (before) {
				messagesByConversation = {
					...messagesByConversation,
					[conversationId]: [...decrypted, ...existing]
				};
			} else {
				messagesByConversation = {
					...messagesByConversation,
					[conversationId]: decrypted
				};
			}
			console.log('📚 Messages loaded and stored:', {
				conversationId,
				totalMessages: messagesByConversation[conversationId].length,
				firstMsg: messagesByConversation[conversationId][0]?.content?.substring(0, 20)
			});
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
		// Force reactivity by creating a new object
		messagesByConversation = {
			...messagesByConversation,
			[conversationId]: [...existing, msg]
		};
		console.log('📝 Message added to store:', {
			conversationId,
			totalMessages: messagesByConversation[conversationId].length,
			newMessageId: msg.id
		});
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
		const updated: Record<string, DecryptedMessageItem[]> = {};
		for (const convId of Object.keys(messagesByConversation)) {
			updated[convId] = messagesByConversation[convId].map((m) =>
				m.tempId === tempId ? { ...realMessage, pending: false } : m
			);
		}
		// Force reactivity
		messagesByConversation = updated;
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
		let decrypted: DecryptedMessageItem;

		if (!content) {
			console.error('❌ Failed to decrypt message:', {
				msgId: msg.id,
				encryptedContent: JSON.stringify(msg.encrypted_content).substring(0, 100) + '...'
			});
			// 顯示解密失敗的訊息
			decrypted = {
				id: msg.id,
				conversationId: msg.conversation_id,
				senderId: msg.sender_id,
				content: '[無法解密此訊息]',
				createdAt: msg.created_at,
				deliveredAt: msg.delivered_at,
				readAt: msg.read_at,
				decryptFailed: true,
			};
		} else {
			console.log('✅ Message decrypted successfully:', {
				msgId: msg.id,
				contentLength: content.length
			});
			decrypted = {
				id: msg.id,
				conversationId: msg.conversation_id,
				senderId: msg.sender_id,
				content,
				createdAt: msg.created_at,
				deliveredAt: msg.delivered_at,
				readAt: msg.read_at,
			};
		}

		addMessage(msg.conversation_id, decrypted);
		return decrypted;
	}

	function getMessages(conversationId: string): DecryptedMessageItem[] {
		const messages = messagesByConversation[conversationId] || [];
		console.log('🔍 getMessages called:', {
			conversationId,
			messagesCount: messages.length,
			allConversations: Object.keys(messagesByConversation),
			messagesInStore: Object.entries(messagesByConversation).map(([id, msgs]) => ({
				convId: id,
				count: msgs.length
			}))
		});
		return messages;
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
			// Force reactivity
			messagesByConversation = {
				...messagesByConversation,
				[conversationId]: existing.filter((m) => m.id !== messageId)
			};
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
