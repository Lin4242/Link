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

	// Cookie helpers for cross-tab state (Safari localStorage issues with NFC)
	function setCookie(name: string, value: string, minutes = 10) {
		const expires = new Date(Date.now() + minutes * 60 * 1000).toUTCString();
		document.cookie = `${name}=${encodeURIComponent(value)}; expires=${expires}; path=/; SameSite=Lax`;
	}

	function getCookie(name: string): string | null {
		const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
		return match ? decodeURIComponent(match[2]) : null;
	}

	function clearRegistrationCookies() {
		document.cookie = 'register_first_card=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/';
		document.cookie = 'register_paired_token=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/';
	}

	onMount(async () => {
		const token = $page.url.searchParams.get('token');
		const forceReset = $page.url.searchParams.get('reset') === 'true';

		// Clear cache if force reset
		if (forceReset) {
			clearRegistrationCookies();
		}

		if (!token) {
			step = 'need_both';
			return;
		}

		loading = true;
		const res = await authApi.checkCard(token);
		loading = false;

		if (res.data?.status === 'can_register') {
			const pairedToken = res.data.paired_token || '';

			// Use cookies for cross-tab state (more reliable on iPhone Safari with NFC)
			const existingFirstCard = getCookie('register_first_card');
			const existingPairedToken = getCookie('register_paired_token');

			if (existingFirstCard && existingFirstCard !== token) {
				// Second card scanned - check if paired
				if (existingFirstCard === pairedToken || token === existingPairedToken) {
					primaryToken = existingFirstCard;
					backupToken = token;
					step = 'form';
					// Clear the registration state now that we have both cards
					clearRegistrationCookies();
				} else {
					// Not paired - start fresh with this card
					primaryToken = token;
					backupToken = '';
					step = 'scan_backup';
					setCookie('register_first_card', token);
					setCookie('register_paired_token', pairedToken);
				}
			} else {
				// First card - will become primary
				primaryToken = token;
				backupToken = '';
				step = 'scan_backup';
				setCookie('register_first_card', token);
				setCookie('register_paired_token', pairedToken);
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

		// 先用臨時金鑰註冊
		const tempKeyPair = generateKeyPair();

		const res = await authApi.register({
			primary_token: primaryToken,
			backup_token: backupToken,
			password,
			nickname: nickname.trim(),
			public_key: tempKeyPair.publicKey,
		});

		if (res.error) {
			error = res.error.message;
			loading = false;
			return;
		}

		if (res.data) {
			// Clear registration state
			clearRegistrationCookies();

			// 取得 userId 後，用密碼推導確定性金鑰
			const { deriveKeyPairFromPassword } = await import('$lib/crypto/keys');
			const derivedKeyPair = await deriveKeyPairFromPassword(password, res.data.user.id);
			console.log('✅ Derived keypair from password + userId');

			// 更新伺服器上的公鑰為推導的公鑰
			const { updateMe } = await import('$lib/api/users');
			await updateMe({ public_key: derivedKeyPair.publicKey });
			console.log('✅ Updated server with derived public key');

			// 儲存推導的私鑰
			try {
				await saveSecretKey(derivedKeyPair.secretKey, password);
				console.log('✅ Derived secret key saved to IndexedDB');
			} catch (e) {
				console.warn('Failed to save secret key to IndexedDB:', e);
			}

			authStore.login(res.data.user, res.data.token);
			window.location.replace('/chat');
		}
	}
</script>

<div class="min-h-screen bg-gradient-to-b from-slate-900 via-slate-800 to-slate-900 flex items-center justify-center p-6">
	<div class="w-full max-w-sm">
		<div class="text-center mb-8">
			<div class="w-14 h-14 mx-auto mb-4 rounded-xl flex items-center justify-center shadow-lg" style="background: linear-gradient(135deg, #3ACACA, #2BA3A3); box-shadow: 0 10px 15px -3px rgba(58, 202, 202, 0.3);">
				<svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
				</svg>
			</div>
			<h1 class="text-xl font-bold text-white tracking-tight">LINK</h1>
		</div>

		<div class="bg-slate-800/50 backdrop-blur rounded-xl p-6 border border-white/10">
			{#if step === 'loading'}
				<div class="text-center py-8">
					<div class="animate-spin w-8 h-8 border-2 border-t-transparent rounded-full mx-auto mb-4" style="border-color: #3ACACA; border-top-color: transparent;"></div>
					<p class="text-slate-400 text-sm">載入中...</p>
				</div>
			{:else if step === 'need_both'}
				<div class="text-center py-4">
					<div class="w-16 h-16 mx-auto mb-4 rounded-xl flex items-center justify-center" style="background-color: rgba(58, 202, 202, 0.1);">
						<svg class="w-8 h-8" style="color: #3ACACA;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A13.916 13.916 0 008 11a4 4 0 118 0c0 1.017-.07 2.019-.203 3m-2.118 6.844A21.88 21.88 0 0015.171 17m3.839 1.132c.645-2.266.99-4.659.99-7.132A8 8 0 008 4.07M3 15.364c.64-1.319 1-2.8 1-4.364 0-1.457.39-2.823 1.07-4" />
						</svg>
					</div>
					<p class="text-slate-200 font-medium mb-2">掃描 NFC 卡片</p>
					<p class="text-slate-500 text-sm">開始註冊您的帳號</p>
				</div>
			{:else if step === 'scan_backup'}
				<div class="text-center py-2">
					<div class="w-12 h-12 mx-auto mb-3 bg-emerald-500/20 rounded-xl flex items-center justify-center">
						<svg class="w-6 h-6 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
					</div>
					<p class="text-emerald-400 font-medium text-sm mb-4">主卡已掃描</p>

					<div class="w-14 h-14 mx-auto mb-4 bg-amber-500/10 rounded-xl flex items-center justify-center">
						<svg class="w-7 h-7 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A13.916 13.916 0 008 11a4 4 0 118 0c0 1.017-.07 2.019-.203 3m-2.118 6.844A21.88 21.88 0 0015.171 17m3.839 1.132c.645-2.266.99-4.659.99-7.132A8 8 0 008 4.07M3 15.364c.64-1.319 1-2.8 1-4.364 0-1.457.39-2.823 1.07-4" />
						</svg>
					</div>
					<p class="text-slate-200 text-sm mb-1">請掃描<span class="text-amber-400 font-medium">副卡</span></p>
					<p class="text-slate-500 text-xs mb-5">將副卡靠近手機感應區</p>

					<button
						onclick={skipBackupCard}
						class="w-full bg-slate-700/50 text-slate-300 py-3 rounded-xl text-sm hover:bg-slate-700 transition-colors border border-white/5"
					>
						不需要副卡
					</button>
					<p class="text-xs text-red-400/70 mt-2">沒有副卡，主卡遺失將無法恢復帳號</p>
				</div>
			{:else}
				<div class="flex justify-center gap-4 mb-5 pb-4 border-b border-white/10">
					<div class="flex items-center gap-2">
						<div class="w-3 h-3 bg-emerald-500 rounded-full"></div>
						<span class="text-xs text-slate-400">主卡</span>
					</div>
					<div class="flex items-center gap-2">
						<div class="w-3 h-3 rounded-full {backupToken ? 'bg-emerald-500' : 'bg-slate-600'}"></div>
						<span class="text-xs text-slate-400">{backupToken ? '副卡' : '無副卡'}</span>
					</div>
				</div>

				<form onsubmit={(e) => { e.preventDefault(); register(); }} class="space-y-4">
					<div>
						<label for="nickname" class="block text-xs font-medium text-slate-400 mb-1.5">暱稱</label>
						<input
							id="nickname"
							type="text"
							bind:value={nickname}
							class="w-full px-4 py-3 bg-slate-700/50 border border-white/10 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:ring-2"
							style="--tw-ring-color: rgba(58, 202, 202, 0.5);"
							placeholder="您的顯示名稱"
							maxlength="50"
						/>
					</div>

					<div>
						<label for="password" class="block text-xs font-medium text-slate-400 mb-1.5">密碼</label>
						<input
							id="password"
							type="password"
							bind:value={password}
							class="w-full px-4 py-3 bg-slate-700/50 border border-white/10 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:ring-2"
							style="--tw-ring-color: rgba(58, 202, 202, 0.5);"
							placeholder="設定您的密碼"
						/>
					</div>

					<div>
						<label for="confirmPassword" class="block text-xs font-medium text-slate-400 mb-1.5">確認密碼</label>
						<input
							id="confirmPassword"
							type="password"
							bind:value={confirmPassword}
							class="w-full px-4 py-3 bg-slate-700/50 border border-white/10 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:ring-2"
							style="--tw-ring-color: rgba(58, 202, 202, 0.5);"
							placeholder="再次輸入密碼"
						/>
					</div>

					<button
						type="submit"
						disabled={loading}
						class="w-full text-white py-3 rounded-xl font-medium transition-all disabled:opacity-50 shadow-lg mt-2"
						style="background: linear-gradient(to right, #3ACACA, #2BA3A3); box-shadow: 0 10px 15px -3px rgba(58, 202, 202, 0.2);"
						onmouseenter={(e) => !e.currentTarget.disabled && (e.currentTarget.style.background = 'linear-gradient(to right, #2BA3A3, #238B8B)')}
						onmouseleave={(e) => !e.currentTarget.disabled && (e.currentTarget.style.background = 'linear-gradient(to right, #3ACACA, #2BA3A3)')}
					>
						{loading ? '註冊中...' : '完成註冊'}
					</button>
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
