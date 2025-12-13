import type { User } from '$lib/types';
import { conversationsApi } from '$lib/api';

export interface ConversationItem {
	id: string;
	peer: User;
	lastMessageAt?: string;
	unreadCount: number;
}

function createConversationsStore() {
	let conversations = $state<ConversationItem[]>([]);
	let activeConversationId = $state<string | null>(null);
	let loading = $state(false);

	async function loadConversations(): Promise<void> {
		loading = true;
		const res = await conversationsApi.getConversations();
		if (res.data) {
			conversations = res.data.map((c) => ({
				id: c.id,
				peer: c.peer,
				lastMessageAt: c.last_message_at,
				unreadCount: c.unread_count,
			}));
		}
		loading = false;
	}

	function setActive(id: string | null): void {
		activeConversationId = id;
		if (id) {
			conversations = conversations.map((c) =>
				c.id === id ? { ...c, unreadCount: 0 } : c
			);
		}
	}

	function updateLastMessage(conversationId: string, timestamp: string): void {
		conversations = conversations.map((c) =>
			c.id === conversationId ? { ...c, lastMessageAt: timestamp } : c
		);
		conversations = [...conversations].sort((a, b) => {
			const aTime = a.lastMessageAt ? new Date(a.lastMessageAt).getTime() : 0;
			const bTime = b.lastMessageAt ? new Date(b.lastMessageAt).getTime() : 0;
			return bTime - aTime;
		});
	}

	function incrementUnread(conversationId: string): void {
		if (conversationId !== activeConversationId) {
			conversations = conversations.map((c) =>
				c.id === conversationId ? { ...c, unreadCount: c.unreadCount + 1 } : c
			);
		}
	}

	function addOrUpdate(item: ConversationItem): void {
		const existing = conversations.find((c) => c.id === item.id);
		if (existing) {
			conversations = conversations.map((c) => (c.id === item.id ? item : c));
		} else {
			conversations = [item, ...conversations];
		}
	}

	return {
		get conversations() {
			return conversations;
		},
		get activeConversationId() {
			return activeConversationId;
		},
		get loading() {
			return loading;
		},
		loadConversations,
		setActive,
		updateLastMessage,
		incrementUnread,
		addOrUpdate,
	};
}

export const conversationsStore = createConversationsStore();
