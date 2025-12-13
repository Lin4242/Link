<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { authApi } from '$lib/api';
	import { loadSecretKey } from '$lib/crypto';
	import { authStore, keysStore } from '$lib/stores';
	import { onMount } from 'svelte';

	let cardToken = $state('');
	let password = $state('');
	let loading = $state(false);
	let error = $state('');
	let cardInfo = $state<{ nickname?: string; warning?: string } | null>(null);

	onMount(async () => {
		const token = $page.url.searchParams.get('token');
		if (token) {
			cardToken = token;
			await checkCard();
		}
	});

	async function checkCard() {
		if (!cardToken) return;
		loading = true;
		error = '';

		const res = await authApi.checkCard(cardToken);
		if (res.error) {
			error = res.error.message;
			loading = false;
			return;
		}

		if (res.data?.status === 'backup') {
			goto(`/login/backup?token=${cardToken}`);
			return;
		}

		if (res.data?.status === 'not_found') {
			goto(`/register/start?token=${cardToken}`);
			return;
		}

		if (res.data?.status === 'revoked') {
			error = '此卡片已被撤銷';
			loading = false;
			return;
		}

		cardInfo = {
			nickname: res.data?.nickname,
			warning: res.data?.warning,
		};
		loading = false;
	}

	async function login() {
		if (!password) {
			error = '請輸入密碼';
			return;
		}

		loading = true;
		error = '';

		const res = await authApi.login(cardToken, password);
		if (res.error) {
			error = res.error.message;
			loading = false;
			return;
		}

		if (res.data) {
			// Try to load secret key (may fail on HTTP or if key wasn't saved)
			try {
				const secretKey = await loadSecretKey(password);
				if (secretKey) {
					await keysStore.unlock(password);
				} else {
					console.warn('No secret key found - E2E encryption disabled');
				}
			} catch (e) {
				console.warn('Failed to load secret key (requires HTTPS):', e);
			}

			authStore.login(res.data.user, res.data.token);
			goto('/chat');
		}
	}
</script>

<div class="min-h-screen bg-gray-100 flex items-center justify-center p-4">
	<div class="bg-white rounded-lg shadow-lg max-w-md w-full p-8">
		<h1 class="text-2xl font-bold text-center mb-6">LINK 登入</h1>

		{#if !cardToken}
			<div class="text-center">
				<div class="w-24 h-24 mx-auto mb-6 bg-blue-100 rounded-full flex items-center justify-center">
					<svg class="w-12 h-12 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A13.916 13.916 0 008 11a4 4 0 118 0c0 1.017-.07 2.019-.203 3m-2.118 6.844A21.88 21.88 0 0015.171 17m3.839 1.132c.645-2.266.99-4.659.99-7.132A8 8 0 008 4.07M3 15.364c.64-1.319 1-2.8 1-4.364 0-1.457.39-2.823 1.07-4"
						/>
					</svg>
				</div>
				<p class="text-gray-600 mb-6">請使用 NFC 卡片掃描以登入</p>
				<p class="text-sm text-gray-500">將卡片靠近手機的 NFC 感應區域</p>
			</div>
		{:else if loading && !cardInfo}
			<div class="text-center py-8">
				<div class="animate-spin w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full mx-auto mb-4"></div>
				<p class="text-gray-600">驗證卡片中...</p>
			</div>
		{:else}
			<form onsubmit={(e) => { e.preventDefault(); login(); }} class="space-y-4">
				{#if cardInfo?.nickname}
					<div class="text-center mb-4">
						<div class="w-16 h-16 mx-auto mb-2 bg-gray-200 rounded-full flex items-center justify-center">
							<span class="text-2xl">{cardInfo.nickname[0]}</span>
						</div>
						<p class="font-medium">{cardInfo.nickname}</p>
					</div>
				{/if}

				<div>
					<label for="password" class="block text-sm font-medium text-gray-700 mb-1">密碼</label>
					<input
						id="password"
						type="password"
						bind:value={password}
						class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
						placeholder="輸入您的密碼"
						autofocus
					/>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="w-full bg-blue-600 text-white py-3 px-4 rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
				>
					{loading ? '登入中...' : '登入'}
				</button>
			</form>
		{/if}

		{#if error}
			<div class="mt-4 p-3 bg-red-100 text-red-700 rounded-lg text-sm">
				{error}
			</div>
		{/if}
	</div>
</div>
