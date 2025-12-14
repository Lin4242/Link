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

		if (res.data?.status === 'not_found' || res.data?.status === 'can_register') {
			goto(`/register/start?token=${cardToken}`);
			return;
		}

		if (res.data?.status === 'revoked') {
			error = '此卡片已被撤銷';
			loading = false;
			return;
		}

		if (res.data?.status === 'pair_already_registered') {
			error = '此卡片組已被註冊';
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
			// Try to load secret key from multiple sources
			let keyLoaded = false;
			
			// 1. Try IndexedDB first
			try {
				const secretKey = await loadSecretKey(password);
				if (secretKey) {
					await keysStore.unlock(password);
					keyLoaded = true;
					console.log('✅ Secret key loaded from IndexedDB');
				}
			} catch (e) {
				console.warn('Failed to load from IndexedDB:', e);
			}
			
			// 2. Try temporary storage if IndexedDB failed
			if (!keyLoaded) {
				try {
					// Check sessionStorage first
					let tempKey = sessionStorage.getItem('temp_secret_key');
					// Then check localStorage with user ID
					if (!tempKey && res.data.user.id) {
						tempKey = localStorage.getItem(`temp_key_${res.data.user.id}`);
					}
					
					if (tempKey) {
						const keyArray = new Uint8Array(JSON.parse(tempKey));
						keysStore.secretKey = keyArray;
						keysStore.publicKey = res.data.user.public_key || '';
						keyLoaded = true;
						console.log('⚠️ Secret key loaded from temporary storage');
						
						// Try to persist it properly
						try {
							const { saveSecretKey } = await import('$lib/crypto');
							await saveSecretKey(keyArray, password);
							console.log('✅ Key migrated to IndexedDB');
						} catch (saveErr) {
							console.warn('Could not migrate key to IndexedDB:', saveErr);
						}
					}
				} catch (e) {
					console.warn('Failed to load from temporary storage:', e);
				}
			}
			
			if (!keyLoaded) {
				console.warn('❌ No secret key found - redirecting to fix-keys');
			}

			authStore.login(res.data.user, res.data.token);

			// Always go to chat - it will handle key unlock/regeneration
			if (!keyLoaded) {
				console.warn('⚠️ No secret key found - chat page will prompt for key setup');
			}
			goto('/chat');
		}
	}
</script>

<div class="min-h-screen flex items-center justify-center p-6" style="background: #0f172a; background: linear-gradient(to bottom, #0f172a, #1e293b, #0f172a);">
	<div class="w-full max-w-sm">
		<div class="text-center mb-8">
			<div class="w-14 h-14 mx-auto mb-4 rounded-xl flex items-center justify-center shadow-lg" style="background: linear-gradient(135deg, #3ACACA, #2BA3A3); box-shadow: 0 10px 15px -3px rgba(58, 202, 202, 0.3);">
				<svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
				</svg>
			</div>
			<h1 class="text-xl font-bold text-white tracking-tight">LINK</h1>
		</div>

		<div class="rounded-xl p-6 border" style="background-color: rgba(30, 41, 59, 0.5); border-color: rgba(255, 255, 255, 0.1);">
			{#if !cardToken}
				<div class="text-center py-4">
					<div class="w-16 h-16 mx-auto mb-4 rounded-xl flex items-center justify-center" style="background-color: rgba(58, 202, 202, 0.1);">
						<svg class="w-8 h-8" style="color: #3ACACA;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A13.916 13.916 0 008 11a4 4 0 118 0c0 1.017-.07 2.019-.203 3m-2.118 6.844A21.88 21.88 0 0015.171 17m3.839 1.132c.645-2.266.99-4.659.99-7.132A8 8 0 008 4.07M3 15.364c.64-1.319 1-2.8 1-4.364 0-1.457.39-2.823 1.07-4" />
						</svg>
					</div>
					<p class="text-slate-200 font-medium mb-2">掃描 NFC 卡片登入</p>
					<p class="text-slate-500 text-sm">將卡片靠近手機感應區</p>
				</div>
			{:else if loading && !cardInfo}
				<div class="text-center py-8">
					<div class="animate-spin w-8 h-8 border-2 border-t-transparent rounded-full mx-auto mb-4" style="border-color: #3ACACA; border-top-color: transparent;"></div>
					<p class="text-slate-400 text-sm">驗證卡片中...</p>
				</div>
			{:else}
				<form onsubmit={(e) => { e.preventDefault(); login(); }} class="space-y-5">
					{#if cardInfo?.nickname}
						<div class="text-center pb-4 border-b border-white/10">
							<div class="w-14 h-14 mx-auto mb-3 bg-gradient-to-br from-slate-600 to-slate-700 rounded-full flex items-center justify-center">
								<span class="text-xl font-medium text-white">{cardInfo.nickname[0]}</span>
							</div>
							<p class="text-white font-medium">{cardInfo.nickname}</p>
						</div>
					{/if}

					<div>
						<input
							id="password"
							type="password"
							bind:value={password}
							class="w-full px-4 py-3 bg-slate-700/50 border border-white/10 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:ring-2"
							style="--tw-ring-color: rgba(58, 202, 202, 0.5); --tw-ring-offset-color: rgba(58, 202, 202, 0.5);"
							placeholder="輸入密碼"
							autofocus
						/>
					</div>

					<button
						type="submit"
						disabled={loading}
						class="w-full text-white py-3 rounded-xl font-medium transition-all disabled:opacity-50 shadow-lg"
						style="background: linear-gradient(to right, #3ACACA, #2BA3A3); box-shadow: 0 10px 15px -3px rgba(58, 202, 202, 0.2);"
						onmouseenter={(e) => !e.currentTarget.disabled && (e.currentTarget.style.background = 'linear-gradient(to right, #2BA3A3, #238B8B)')}
						onmouseleave={(e) => !e.currentTarget.disabled && (e.currentTarget.style.background = 'linear-gradient(to right, #3ACACA, #2BA3A3)')}
					>
						{loading ? '登入中...' : '登入'}
					</button>
				</form>
			{/if}

			{#if error}
				<div class="mt-4 p-3 text-red-400 rounded-xl text-sm" style="background-color: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.2);">
					{error}
				</div>
			{/if}
		</div>
	</div>
</div>
