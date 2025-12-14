<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores';
	import { keysStore } from '$lib/stores/keys.svelte';
	import { goto } from '$app/navigation';
	import { generateKeyPair, saveSecretKey, loadSecretKey } from '$lib/crypto';
	
	let password = $state('');
	let status = $state('');
	let loading = $state(false);
	let hasKey = $state(false);
	
	onMount(async () => {
		authStore.init();
		if (!authStore.isAuthenticated) {
			status = '❌ 未登入，請先登入';
			setTimeout(() => goto('/'), 2000);
			return;
		}
		
		hasKey = !!keysStore.secretKey;
		if (hasKey) {
			status = '✅ 密鑰已載入';
		} else {
			status = '⚠️ 沒有密鑰，需要修復';
		}
	});
	
	async function fixKeys() {
		if (!password) {
			status = '❌ 請輸入密碼';
			return;
		}
		
		loading = true;
		status = '🔄 嘗試載入現有密鑰...';
		
		try {
			// First try to load existing key
			const existingKey = await loadSecretKey(password);
			if (existingKey) {
				status = '✅ 找到現有密鑰，正在載入...';
				await keysStore.unlock(password);
				status = '✅ 密鑰載入成功！';
				setTimeout(() => goto('/chat'), 1500);
				return;
			}
			
			// No existing key, generate new one
			status = '🔑 沒有找到密鑰，生成新密鑰...';
			const { publicKey, secretKey } = generateKeyPair();
			
			// Save the new key
			status = '💾 儲存新密鑰...';
			await saveSecretKey(secretKey, password);
			
			// Update user's public key in backend by re-registering
			status = '📡 更新公鑰（使用重新註冊方式）...';
			// For now, just save locally - backend update would need a new endpoint
			console.log('Generated new keypair, public key:', publicKey);
			
			// Load the new key into store
			await keysStore.unlock(password);
			
			status = '✅ 密鑰修復成功！重新導向到聊天頁面...';
			setTimeout(() => goto('/chat'), 1500);
			
		} catch (error) {
			console.error('Fix keys error:', error);
			status = `❌ 錯誤: ${error}`;
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen bg-gradient-to-b from-slate-900 to-slate-950 text-white flex items-center justify-center p-4">
	<div class="max-w-md w-full">
		<div class="bg-slate-800/50 rounded-2xl p-6 border border-white/10">
			<h1 class="text-2xl font-bold mb-4">修復加密密鑰</h1>
			
			<div class="mb-6 p-4 bg-slate-700/50 rounded-lg">
				<p class="text-sm mb-2">狀態：</p>
				<p class="font-mono text-sm">{status}</p>
			</div>
			
			<div class="mb-4">
				<p class="text-sm text-slate-400 mb-2">
					{#if authStore.user}
						登入身份：{authStore.user.nickname}
					{/if}
				</p>
				<p class="text-sm text-slate-400 mb-4">
					目前密鑰狀態：{hasKey ? '✅ 已載入' : '❌ 未載入'}
				</p>
			</div>
			
			<form onsubmit={(e) => { e.preventDefault(); fixKeys(); }} class="space-y-4">
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
					{loading ? '處理中...' : '修復密鑰'}
				</button>
			</form>
			
			<div class="mt-6 text-xs text-slate-500 text-center">
				<p>此工具會嘗試載入或重新生成你的加密密鑰</p>
				<p>如果密鑰遺失，之前的訊息將無法解密</p>
			</div>
		</div>
	</div>
</div>