<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores';
	import { keysStore } from '$lib/stores/keys.svelte';
	import { goto } from '$app/navigation';
	import { generateKeyPair, saveSecretKey, loadSecretKey } from '$lib/crypto';
	import { encodeBase64 } from 'tweetnacl-util';
	import nacl from 'tweetnacl';
	
	let password = $state('');
	let status = $state('');
	let loading = $state(false);
	let currentPublicKey = $state('');
	let newPublicKey = $state('');
	
	onMount(async () => {
		authStore.init();
		if (!authStore.isAuthenticated) {
			status = '❌ 未登入，請先登入';
			setTimeout(() => goto('/'), 2000);
			return;
		}
		
		// 檢查當前公鑰狀態
		currentPublicKey = authStore.user?.public_key || '(無)';
		
		if (keysStore.secretKey) {
			// 從現有密鑰計算公鑰
			const keyPair = nacl.box.keyPair.fromSecretKey(keysStore.secretKey);
			newPublicKey = encodeBase64(keyPair.publicKey);
			status = '✅ 已有本地密鑰，可以更新公鑰';
		} else {
			status = '⚠️ 沒有本地密鑰，需要生成或載入';
		}
	});
	
	async function updatePublicKey() {
		if (!password) {
			status = '❌ 請輸入密碼';
			return;
		}
		
		loading = true;
		status = '🔄 處理中...';
		
		try {
			// 1. 嘗試載入或生成密鑰
			let secretKey = keysStore.secretKey;
			let publicKey = newPublicKey;
			
			if (!secretKey) {
				status = '🔑 嘗試載入現有密鑰...';
				const loadedKey = await loadSecretKey(password);
				
				if (loadedKey) {
					secretKey = loadedKey;
					const keyPair = nacl.box.keyPair.fromSecretKey(secretKey);
					publicKey = encodeBase64(keyPair.publicKey);
					status = '✅ 成功載入現有密鑰';
				} else {
					status = '🔑 生成新密鑰對...';
					const newKeyPair = generateKeyPair();
					secretKey = newKeyPair.secretKey;
					publicKey = newKeyPair.publicKey;
					
					// 儲存新密鑰
					try {
						await saveSecretKey(secretKey, password);
						status = '💾 新密鑰已儲存';
					} catch (e) {
						console.warn('Failed to save to IndexedDB:', e);
						sessionStorage.setItem('temp_secret_key', JSON.stringify(Array.from(secretKey)));
					}
				}
				
				// 更新 keysStore
				keysStore.secretKey = secretKey;
				keysStore.publicKey = publicKey;
			}
			
			// 2. 更新資料庫中的公鑰
			status = '📡 更新公鑰到伺服器...';
			const response = await fetch('/api/v1/users/me', {
				method: 'PATCH',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${authStore.token}`
				},
				body: JSON.stringify({ public_key: publicKey })
			});
			
			if (response.ok) {
				status = '✅ 公鑰更新成功！';
				newPublicKey = publicKey;
				
				// 更新 authStore 中的用戶資料
				if (authStore.user) {
					authStore.user.public_key = publicKey;
				}
				
				setTimeout(() => goto('/chat'), 1500);
			} else {
				const error = await response.json();
				status = `❌ 更新失敗: ${error.message || response.statusText}`;
			}
			
		} catch (error) {
			console.error('Update public key error:', error);
			status = `❌ 錯誤: ${error}`;
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen bg-gradient-to-b from-slate-900 to-slate-950 text-white flex items-center justify-center p-4">
	<div class="max-w-md w-full">
		<div class="bg-slate-800/50 rounded-2xl p-6 border border-white/10">
			<h1 class="text-2xl font-bold mb-4">更新公鑰</h1>
			
			<div class="mb-6 p-4 bg-slate-700/50 rounded-lg">
				<p class="text-sm mb-2">狀態：</p>
				<p class="font-mono text-sm">{status}</p>
			</div>
			
			<div class="mb-4 text-sm text-slate-400">
				<p class="mb-2">當前用戶：{authStore.user?.nickname}</p>
				<p class="mb-2">資料庫公鑰：{currentPublicKey.substring(0, 20)}...</p>
				{#if newPublicKey}
					<p>本地公鑰：{newPublicKey.substring(0, 20)}...</p>
				{/if}
			</div>
			
			<form onsubmit={(e) => { e.preventDefault(); updatePublicKey(); }} class="space-y-4">
				<div>
					<label for="password" class="block text-sm font-medium mb-2">
						輸入你的登入密碼
					</label>
					<input
						type="password"
						id="password"
						bind:value={password}
						class="w-full px-4 py-3 bg-slate-700/50 border border-white/10 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/50"
						placeholder="密碼"
						disabled={loading}
					/>
				</div>
				
				<button
					type="submit"
					disabled={loading || !password}
					class="w-full py-3 bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-xl font-medium disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-blue-500/20"
				>
					{loading ? '處理中...' : '更新公鑰'}
				</button>
			</form>
			
			<div class="mt-6 text-xs text-slate-500 text-center">
				<p>此工具會更新你的公鑰到伺服器</p>
				<p>讓其他用戶可以發送加密訊息給你</p>
			</div>
		</div>
	</div>
</div>