<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { authApi } from '$lib/api';
	import { loadSecretKey } from '$lib/crypto';
	import { authStore, keysStore } from '$lib/stores';
	import { BackupCardWarning } from '$lib/components';
	import { onMount } from 'svelte';

	let cardToken = $state('');
	let password = $state('');
	let loading = $state(false);
	let error = $state('');
	let showWarning = $state(false);
	let cardInfo = $state<{ nickname?: string } | null>(null);

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

		const res = await authApi.checkCard(cardToken);
		if (res.error) {
			error = res.error.message;
			loading = false;
			return;
		}

		if (res.data?.status !== 'backup') {
			if (res.data?.status === 'primary') {
				goto(`/login?token=${cardToken}`);
			} else {
				goto(`/register/start?token=${cardToken}`);
			}
			return;
		}

		cardInfo = { nickname: res.data?.nickname };
		loading = false;
	}

	function handleSubmit() {
		if (!password) {
			error = '請輸入密碼';
			return;
		}
		showWarning = true;
	}

	async function confirmLogin() {
		loading = true;
		error = '';

		const res = await authApi.loginWithBackup(cardToken, password, true);
		if (res.error) {
			error = res.error.message;
			loading = false;
			showWarning = false;
			return;
		}

		if (res.data) {
			const secretKey = await loadSecretKey(password);
			if (!secretKey) {
				error = '無法載入私鑰';
				loading = false;
				showWarning = false;
				return;
			}

			authStore.login(res.data.user, res.data.token);
			await keysStore.unlock(password);
			goto('/chat');
		}
	}
</script>

<BackupCardWarning
	show={showWarning}
	onConfirm={confirmLogin}
	onCancel={() => (showWarning = false)}
	{loading}
/>

<div class="min-h-screen bg-gray-100 flex items-center justify-center p-4">
	<div class="bg-white rounded-lg shadow-lg max-w-md w-full p-8">
		<div class="flex items-center gap-3 mb-6">
			<div class="w-10 h-10 bg-yellow-100 rounded-full flex items-center justify-center">
				<svg class="w-6 h-6 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
					/>
				</svg>
			</div>
			<h1 class="text-xl font-bold text-yellow-700">附卡登入</h1>
		</div>

		<div class="bg-yellow-50 p-4 rounded-lg mb-6">
			<p class="text-sm text-yellow-700">
				您正在使用附卡（備援卡）。登入後，您的主卡將被永久撤銷。
			</p>
		</div>

		{#if loading && !showWarning}
			<div class="text-center py-8">
				<div class="animate-spin w-8 h-8 border-4 border-yellow-600 border-t-transparent rounded-full mx-auto mb-4"></div>
				<p class="text-gray-600">處理中...</p>
			</div>
		{:else}
			<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-4">
				{#if cardInfo?.nickname}
					<div class="text-center mb-4">
						<p class="font-medium text-gray-700">帳號：{cardInfo.nickname}</p>
					</div>
				{/if}

				<div>
					<label for="password" class="block text-sm font-medium text-gray-700 mb-1">密碼</label>
					<input
						id="password"
						type="password"
						bind:value={password}
						class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-yellow-500 focus:border-yellow-500"
						placeholder="輸入您的密碼"
					/>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="w-full bg-yellow-600 text-white py-3 px-4 rounded-lg hover:bg-yellow-700 transition-colors disabled:opacity-50"
				>
					繼續（將顯示警告）
				</button>

				<a
					href="/login"
					class="block text-center text-sm text-gray-500 hover:text-gray-700"
				>
					使用主卡登入
				</a>
			</form>
		{/if}

		{#if error}
			<div class="mt-4 p-3 bg-red-100 text-red-700 rounded-lg text-sm">
				{error}
			</div>
		{/if}
	</div>
</div>
