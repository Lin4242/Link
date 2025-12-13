import type { ApiResponse } from '$lib/types';

const API_URL = import.meta.env.VITE_API_URL || 'https://192.168.1.99:9443';

let authToken: string | null = null;

export function setAuthToken(token: string | null): void {
	authToken = token;
}

export function getAuthToken(): string | null {
	return authToken;
}

async function request<T>(
	endpoint: string,
	options: RequestInit = {}
): Promise<ApiResponse<T>> {
	const headers: Record<string, string> = {
		'Content-Type': 'application/json',
		...(options.headers as Record<string, string>),
	};

	if (authToken) {
		headers['Authorization'] = `Bearer ${authToken}`;
	}

	try {
		const response = await fetch(`${API_URL}/api/v1${endpoint}`, {
			...options,
			headers,
		});

		const json = await response.json();

		if (!response.ok) {
			return { error: json.error };
		}

		return { data: json.data };
	} catch (err) {
		return {
			error: {
				code: 'NETWORK_ERROR',
				message: err instanceof Error ? err.message : '網路錯誤',
			},
		};
	}
}

export async function get<T>(endpoint: string): Promise<ApiResponse<T>> {
	return request<T>(endpoint, { method: 'GET' });
}

export async function post<T>(endpoint: string, body?: unknown): Promise<ApiResponse<T>> {
	return request<T>(endpoint, {
		method: 'POST',
		body: body ? JSON.stringify(body) : undefined,
	});
}

export async function patch<T>(endpoint: string, body?: unknown): Promise<ApiResponse<T>> {
	return request<T>(endpoint, {
		method: 'PATCH',
		body: body ? JSON.stringify(body) : undefined,
	});
}

export async function del<T>(endpoint: string): Promise<ApiResponse<T>> {
	return request<T>(endpoint, { method: 'DELETE' });
}
