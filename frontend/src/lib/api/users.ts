import { get, patch } from './client';
import type { User, Card } from '$lib/types';

export async function getMe() {
	return get<User>('/users/me');
}

export async function getMyCards() {
	return get<Card[]>('/users/me/cards');
}

export async function updateMe(data: { nickname?: string; avatar_url?: string; public_key?: string }) {
	return patch<User>('/users/me', data);
}

export async function searchUsers(query: string) {
	return get<User[]>(`/users/search?q=${encodeURIComponent(query)}`);
}

export async function getPublicKey(userId: string) {
	return get<{ public_key: string }>(`/users/${userId}/public-key`);
}
