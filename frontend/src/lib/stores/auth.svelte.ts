import type { User } from '$lib/types';
import { setAuthToken } from '$lib/api/client';

const AUTH_KEY = 'link_auth';

interface AuthState {
	user: User | null;
	token: string | null;
	isAuthenticated: boolean;
}

function createAuthStore() {
	let user = $state<User | null>(null);
	let token = $state<string | null>(null);

	function init() {
		if (typeof window === 'undefined') return;
		const stored = localStorage.getItem(AUTH_KEY);
		if (stored) {
			try {
				const data = JSON.parse(stored);
				user = data.user;
				token = data.token;
				setAuthToken(token);
			} catch {
				localStorage.removeItem(AUTH_KEY);
			}
		}
	}

	function login(userData: User, authToken: string) {
		user = userData;
		token = authToken;
		setAuthToken(authToken);
		localStorage.setItem(AUTH_KEY, JSON.stringify({ user: userData, token: authToken }));
	}

	function logout() {
		user = null;
		token = null;
		setAuthToken(null);
		localStorage.removeItem(AUTH_KEY);
	}

	function updateUser(userData: Partial<User>) {
		if (user) {
			user = { ...user, ...userData };
			localStorage.setItem(AUTH_KEY, JSON.stringify({ user, token }));
		}
	}

	return {
		get user() {
			return user;
		},
		get token() {
			return token;
		},
		get isAuthenticated() {
			return !!token && !!user;
		},
		init,
		login,
		logout,
		updateUser,
	};
}

export const authStore = createAuthStore();
