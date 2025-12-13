import { get, post, del } from './client';
import type { User, Friendship } from '$lib/types';

export interface FriendWithUser {
	id: string;
	requester_id: string;
	addressee_id: string;
	status: Friendship['status'];
	created_at: string;
	updated_at: string;
	friend: User;
}

export async function getFriends() {
	return get<FriendWithUser[]>('/friends');
}

export async function getPendingRequests() {
	return get<FriendWithUser[]>('/friends/requests');
}

export async function sendFriendRequest(addresseeId: string) {
	return post<Friendship>('/friends/request', { addressee_id: addresseeId });
}

export async function acceptFriendRequest(friendshipId: string) {
	return post<{ message: string }>(`/friends/${friendshipId}/accept`);
}

export async function rejectFriendRequest(friendshipId: string) {
	return post<{ message: string }>(`/friends/${friendshipId}/reject`);
}

export async function removeFriend(friendshipId: string) {
	return del<{ message: string }>(`/friends/${friendshipId}`);
}
