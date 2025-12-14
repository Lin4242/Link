<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { authApi } from '$lib/api';
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
				goto(`/register?token=${cardToken}`);
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
			// 使用密碼推導金鑰 (v4.3)
			let keyLoaded = false;

			// 1. 先嘗試 IndexedDB
			try {
				const { loadSecretKey } = await import('$lib/crypto');
				const secretKey = await loadSecretKey(password);
				if (secretKey) {
					await keysStore.unlock(password);
					keyLoaded = true;
					console.log('✅ Secret key loaded from IndexedDB');
				}
			} catch (e) {
				console.warn('Failed to load from IndexedDB:', e);
			}

			// 2. 若無快取，嘗試密碼推導
			if (!keyLoaded && res.data.user.id) {
				try {
					const { deriveKeyPairFromPassword, saveSecretKey } = await import('$lib/crypto/keys');
					const { publicKey, secretKey } = await deriveKeyPairFromPassword(password, res.data.user.id);

					if (res.data.user.public_key === publicKey) {
						// 公鑰匹配，使用推導的金鑰
						await saveSecretKey(secretKey, password);
						keysStore.secretKey = secretKey;
						keysStore.publicKey = publicKey;
						keyLoaded = true;
						console.log('✅ Key derived from password + userId');
					} else {
						// 公鑰不匹配，更新伺服器
						const { updateMe } = await import('$lib/api/users');
						await updateMe({ public_key: publicKey });
						await saveSecretKey(secretKey, password);
						keysStore.secretKey = secretKey;
						keysStore.publicKey = publicKey;
						keyLoaded = true;
						console.log('✅ Server public key updated with derived key');
					}
				} catch (e) {
					console.warn('Failed to derive key:', e);
				}
			}

			if (!keyLoaded) {
				console.warn('⚠️ No secret key loaded');
			}

			authStore.login(res.data.user, res.data.token);
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

<div class="min-h-screen bg-gradient-to-b from-slate-900 via-slate-800 to-slate-900 flex items-center justify-center p-6">
	<div class="w-full max-w-sm">
		<div class="text-center mb-8">
			<div class="w-14 h-14 mx-auto mb-4 bg-gradient-to-br from-amber-400 to-orange-600 rounded-xl flex items-center justify-center shadow-lg shadow-orange-500/30">
				<svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
				</svg>
			</div>
			<h1 class="text-xl font-bold text-white tracking-tight">附卡登入</h1>
			<p class="text-slate-400 text-sm mt-1">Backup Card Login</p>
		</div>

		<div class="bg-slate-800/50 backdrop-blur rounded-xl p-6 border border-white/10">
			<div class="bg-amber-500/10 border border-amber-500/20 rounded-xl p-4 mb-6">
				<p class="text-amber-400 text-sm">
					您正在使用附卡（備援卡）。登入後，您的主卡將被<span class="font-bold">永久撤銷</span>。
				</p>
			</div>

			{#if loading && !showWarning}
				<div class="text-center py-8">
					<div class="animate-spin w-8 h-8 border-2 border-amber-400 border-t-transparent rounded-full mx-auto mb-4"></div>
					<p class="text-slate-400 text-sm">處理中...</p>
				</div>
			{:else}
				<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-5">
					{#if cardInfo?.nickname}
						<div class="text-center pb-4 border-b border-white/10">
							<div class="w-14 h-14 mx-auto mb-3 bg-gradient-to-br from-slate-600 to-slate-700 rounded-full flex items-center justify-center">
								<span class="text-xl font-medium text-white">{cardInfo.nickname[0]}</span>
							</div>
							<p class="text-white font-medium">{cardInfo.nickname}</p>
						</div>
					{/if}

					<div>
						<label for="password" class="block text-xs font-medium text-slate-400 mb-1.5">密碼</label>
						<input
							id="password"
							type="password"
							bind:value={password}
							class="w-full px-4 py-3 bg-slate-700/50 border border-white/10 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-amber-500/50 focus:border-amber-500/50"
							placeholder="輸入密碼"
						/>
					</div>

					<button
						type="submit"
						disabled={loading}
						class="w-full bg-gradient-to-r from-amber-500 to-orange-600 text-white py-3 rounded-xl font-medium hover:from-amber-600 hover:to-orange-700 transition-all disabled:opacity-50 shadow-lg shadow-orange-500/20"
					>
						繼續
					</button>

					<a
						href="/login"
						class="block text-center text-sm text-slate-500 hover:text-slate-300 transition-colors"
					>
						使用主卡登入
					</a>
				</form>
			{/if}

			{#if error}
				<div class="mt-4 p-3 bg-red-500/10 border border-red-500/20 text-red-400 rounded-xl text-sm">
					{error}
				</div>
			{/if}
		</div>
	</div>
</div>
