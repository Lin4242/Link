import type { User } from '$lib/types';
import { friendsApi } from '$lib/api';

export interface FriendItem {
	id: string;
	friendshipId: string;
	user: User;
	isOnline: boolean;
}

function createFriendsStore() {
	let friends = $state<FriendItem[]>([]);
	let pendingRequests = $state<FriendItem[]>([]);
	let loading = $state(false);

	async function loadFriends(): Promise<void> {
		loading = true;
		const res = await friendsApi.getFriends();
		if (res.data) {
			friends = res.data.map((f) => ({
				id: f.friend.id,
				friendshipId: f.id,
				user: f.friend,
				isOnline: false,
			}));
		}
		loading = false;
	}

	async function loadPendingRequests(): Promise<void> {
		const res = await friendsApi.getPendingRequests();
		if (res.data) {
			pendingRequests = res.data.map((f) => ({
				id: f.friend.id,
				friendshipId: f.id,
				user: f.friend,
				isOnline: false,
			}));
		}
	}

	function setOnline(userId: string, online: boolean): void {
		friends = friends.map((f) => (f.id === userId ? { ...f, isOnline: online } : f));
	}

	function removeFriend(friendshipId: string): void {
		friends = friends.filter((f) => f.friendshipId !== friendshipId);
	}

	function removeRequest(friendshipId: string): void {
		pendingRequests = pendingRequests.filter((r) => r.friendshipId !== friendshipId);
	}

	function addFriend(item: FriendItem): void {
		if (!friends.find((f) => f.id === item.id)) {
			friends = [...friends, item];
		}
	}

	return {
		get friends() {
			return friends;
		},
		get pendingRequests() {
			return pendingRequests;
		},
		get loading() {
			return loading;
		},
		loadFriends,
		loadPendingRequests,
		setOnline,
		removeFriend,
		removeRequest,
		addFriend,
	};
}

export const friendsStore = createFriendsStore();
