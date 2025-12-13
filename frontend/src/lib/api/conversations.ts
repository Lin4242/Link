import { get } from './client';
import type { User, Message } from '$lib/types';

export interface ConversationWithPeer {
	id: string;
	participant_1: string;
	participant_2: string;
	last_message_at?: string;
	created_at: string;
	peer: User;
	unread_count: number;
}

export async function getConversations() {
	return get<ConversationWithPeer[]>('/conversations');
}

export async function getMessages(conversationId: string, limit = 50, before?: string) {
	let url = `/conversations/${conversationId}/messages?limit=${limit}`;
	if (before) {
		url += `&before=${encodeURIComponent(before)}`;
	}
	return get<Message[]>(url);
}
