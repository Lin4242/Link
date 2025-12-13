import { get, post } from './client';
import type { User } from '$lib/types';

export interface CardCheckResult {
	status: 'can_register' | 'invalid_token' | 'pair_already_registered' | 'primary' | 'backup' | 'revoked';
	user_id?: string;
	card_type?: 'primary' | 'backup';
	warning?: string;
	paired_token?: string;
}

export interface AuthResponse {
	user: User;
	token: string;
}

export async function checkCard(token: string) {
	return get<CardCheckResult>(`/auth/check-card/${token}`);
}

export async function register(data: {
	primary_token: string;
	backup_token: string;
	password: string;
	nickname: string;
	public_key: string;
}) {
	return post<AuthResponse>('/auth/register', data);
}

export async function login(cardToken: string, password: string) {
	return post<AuthResponse>('/auth/login', {
		card_token: cardToken,
		password,
	});
}

export async function loginWithBackup(cardToken: string, password: string, confirm: boolean) {
	return post<AuthResponse>('/auth/login/backup', {
		card_token: cardToken,
		password,
		confirm,
	});
}

export async function logout() {
	return post<{ message: string }>('/auth/logout');
}
