<script lang="ts">
	import { page } from '$app/stores';
	import { authApi } from '$lib/api';
	import { generateKeyPair, saveSecretKey } from '$lib/crypto';
	import { authStore } from '$lib/stores';
	import { onMount } from 'svelte';

	let primaryToken = $state('');
	let backupToken = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let nickname = $state('');
	let step = $state<'loading' | 'need_both' | 'scan_backup' | 'form'>('loading');
	let loading = $state(false);
	let error = $state('');

	onMount(async () => {
		const token = $page.url.searchParams.get('token');

		if (!token) {
			step = 'need_both';
			return;
		}

		loading = true;
		const res = await authApi.checkCard(token);
		loading = false;

		if (res.data?.status === 'can_register') {
			const pairedToken = res.data.paired_token || '';

			// Use localStorage instead of sessionStorage (NFC opens new tabs on iPhone)
			const existingFirstCard = localStorage.getItem('register_first_card');
			const existingPairedToken = localStorage.getItem('register_paired_token');

			if (existingFirstCard && existingFirstCard !== token) {
				// Second card scanned - check if paired
				if (existingFirstCard === pairedToken || token === existingPairedToken) {
					primaryToken = existingFirstCard;
					backupToken = token;
					step = 'form';
					// Clear the registration state now that we have both cards
					localStorage.removeItem('register_first_card');
					localStorage.removeItem('register_paired_token');
				} else {
					// Not paired - start fresh with this card
					primaryToken = token;
					backupToken = '';
					step = 'scan_backup';
					localStorage.setItem('register_first_card', token);
					localStorage.setItem('register_paired_token', pairedToken);
				}
			} else {
				// First card - will become primary
				primaryToken = token;
				backupToken = '';
				step = 'scan_backup';
				localStorage.setItem('register_first_card', token);
				localStorage.setItem('register_paired_token', pairedToken);
			}
		} else if (res.data?.status === 'primary') {
			window.location.replace(`/login?token=${token}`);
			return;
		} else if (res.data?.status === 'backup') {
			window.location.replace(`/login/backup?token=${token}`);
			return;
		} else {
			error = res.data?.status === 'pair_already_registered'
				? '此卡片對已被註冊'
				: '無法使用此卡片';
			step = 'need_both';
		}
	});

	function skipBackupCard() {
		backupToken = '';
		step = 'form';
	}

	async function register() {
		if (password !== confirmPassword) {
			error = '密碼不一致';
			return;
		}
		if (!password) {
			error = '請輸入密碼';
			return;
		}
		if (!nickname.trim()) {
			error = '請輸入暱稱';
			return;
		}

		loading = true;
		error = '';

		const keyPair = generateKeyPair();

		const res = await authApi.register({
			primary_token: primaryToken,
			backup_token: backupToken,
			password,
			nickname: nickname.trim(),
			public_key: keyPair.publicKey,
		});

		if (res.error) {
			error = res.error.message;
			loading = false;
			return;
		}

		if (res.data) {
			// Clear registration state
			localStorage.removeItem('register_first_card');
			localStorage.removeItem('register_paired_token');

			// Try to save secret key (may fail on non-HTTPS)
			try {
				await saveSecretKey(keyPair.secretKey, password);
			} catch (e) {
				console.warn('Failed to save secret key (requires HTTPS):', e);
				// Store temporarily in sessionStorage for this session
				sessionStorage.setItem('temp_secret_key', JSON.stringify(Array.from(keyPair.secretKey)));
			}

			authStore.login(res.data.user, res.data.token);
			window.location.replace('/chat');
		}
	}
</script>

<div class="min-h-screen bg-gray-100 flex items-center justify-center p-4">
	<div class="bg-white rounded-lg shadow-lg max-w-md w-full p-8">
		<h1 class="text-2xl font-bold text-center mb-6">LINK 註冊</h1>

		{#if step === 'loading'}
			<div class="text-center">
				<div class="animate-spin w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full mx-auto"></div>
				<p class="mt-4 text-gray-600">載入中...</p>
			</div>
		{:else if step === 'need_both'}
			<div class="text-center">
				<p class="text-gray-600 mb-4">請掃描任一張 LINK 卡片開始註冊</p>
				<div class="w-24 h-24 mx-auto mb-6 bg-gray-100 rounded-full flex items-center justify-center">
					<svg class="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A13.916 13.916 0 008 11a4 4 0 118 0c0 1.017-.07 2.019-.203 3m-2.118 6.844A21.88 21.88 0 0015.171 17m3.839 1.132c.645-2.266.99-4.659.99-7.132A8 8 0 008 4.07M3 15.364c.64-1.319 1-2.8 1-4.364 0-1.457.39-2.823 1.07-4" />
					</svg>
				</div>
			</div>
		{:else if step === 'scan_backup'}
			<div class="text-center">
				<div class="w-16 h-16 mx-auto mb-4 bg-green-100 rounded-full flex items-center justify-center">
					<svg class="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
				</div>
				<p class="text-green-600 font-semibold mb-4">主卡已掃描</p>

				<div class="w-24 h-24 mx-auto mb-6 bg-yellow-100 rounded-full flex items-center justify-center">
					<svg class="w-12 h-12 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A13.916 13.916 0 008 11a4 4 0 118 0c0 1.017-.07 2.019-.203 3m-2.118 6.844A21.88 21.88 0 0015.171 17m3.839 1.132c.645-2.266.99-4.659.99-7.132A8 8 0 008 4.07M3 15.364c.64-1.319 1-2.8 1-4.364 0-1.457.39-2.823 1.07-4" />
					</svg>
				</div>
				<p class="text-gray-700 font-medium mb-2">請掃描<span class="text-yellow-600">副卡</span></p>
				<p class="text-sm text-gray-500 mb-6">將副卡靠近手機 NFC 感應區</p>

				<button
					onclick={skipBackupCard}
					class="w-full bg-gray-200 text-gray-700 py-3 px-4 rounded-lg hover:bg-gray-300 transition-colors"
				>
					不需要副卡
				</button>
				<p class="text-xs text-red-500 mt-2">沒有副卡，主卡遺失將無法恢復帳號</p>
			</div>
		{:else}
			<div class="text-center mb-6">
				<div class="flex justify-center gap-4 mb-4">
					<div class="flex items-center gap-2">
						<div class="w-4 h-4 bg-green-500 rounded-full"></div>
						<span class="text-sm text-gray-600">主卡</span>
					</div>
					<div class="flex items-center gap-2">
						<div class="w-4 h-4 rounded-full {backupToken ? 'bg-green-500' : 'bg-gray-300'}"></div>
						<span class="text-sm text-gray-600">{backupToken ? '副卡' : '無副卡'}</span>
					</div>
				</div>
			</div>

			<form onsubmit={(e) => { e.preventDefault(); register(); }} class="space-y-4">
				<div>
					<label for="nickname" class="block text-sm font-medium text-gray-700 mb-1">暱稱</label>
					<input
						id="nickname"
						type="text"
						bind:value={nickname}
						class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
						placeholder="您的顯示名稱"
						maxlength="50"
					/>
				</div>

				<div>
					<label for="password" class="block text-sm font-medium text-gray-700 mb-1">密碼</label>
					<input
						id="password"
						type="password"
						bind:value={password}
						class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
						placeholder="至少 8 個字元"
					/>
				</div>

				<div>
					<label for="confirmPassword" class="block text-sm font-medium text-gray-700 mb-1">確認密碼</label>
					<input
						id="confirmPassword"
						type="password"
						bind:value={confirmPassword}
						class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
						placeholder="再次輸入密碼"
					/>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="w-full bg-blue-600 text-white py-3 px-4 rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
				>
					{loading ? '註冊中...' : '完成註冊'}
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
