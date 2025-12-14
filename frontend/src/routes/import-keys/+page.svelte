<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores';
	import { keysStore } from '$lib/stores/keys.svelte';
	import { goto } from '$app/navigation';
	import { saveSecretKey } from '$lib/crypto';
	import { decodeBase64 } from 'tweetnacl-util';
	
	let password = $state('');
	let status = $state('');
	let loading = $state(false);
	
	// 預設的密鑰（F 和 N 的）
	const keys = {
		'F': {
			publicKey: '491IZL9EeBgQER+zM8q1DjZxisq+1F4ONmvucU4Xxmc=',
			secretKey: 'VpRfgl9QTuNwxtHJ++EjyeOZTqODz0I2pekKLfJQ1tg=',
			password: '000000'
		},
		'N': {
			publicKey: 'pNEEV/aCJ4K0XUS5FGFX9hXV5+eh/IJeAEL76h/sqzs=',
			secretKey: '2ar9400RoPfWMGXDlACx3x25hl2JCkIo44Adsuo5YgI=',
			password: '999999'
		}
	};
	
	onMount(async () => {
		authStore.init();
		if (!authStore.isAuthenticated) {
			status = '❌ 未登入，請先登入';
			setTimeout(() => goto('/'), 2000);
			return;
		}
		
		const nickname = authStore.user?.nickname;
		if (nickname && (nickname === 'F' || nickname === 'N')) {
			status = `歡迎 ${nickname}！請輸入密碼導入你的密鑰`;
			// 自動填入對應的密碼提示
			if (nickname === 'F') {
				status += ' (提示: 000000)';
			} else if (nickname === 'N') {
				status += ' (提示: 999999)';
			}
		} else {
			status = '此工具只供 F 和 N 使用';
		}
	});
	
	async function importKeys() {
		const nickname = authStore.user?.nickname;
		if (!nickname || !(nickname === 'F' || nickname === 'N')) {
			status = '❌ 只有 F 和 N 可以使用此工具';
			return;
		}
		
		if (!password) {
			status = '❌ 請輸入密碼';
			return;
		}
		
		// 檢查密碼是否正確
		const userKeys = keys[nickname as 'F' | 'N'];
		if (password !== userKeys.password) {
			status = '❌ 密碼錯誤';
			return;
		}
		
		loading = true;
		status = '🔄 導入密鑰中...';
		
		try {
			// 解碼私鑰
			const secretKey = decodeBase64(userKeys.secretKey);
			
			// 儲存到 IndexedDB
			try {
				await saveSecretKey(secretKey, password);
				status = '✅ 密鑰已儲存到 IndexedDB';
			} catch (e) {
				console.warn('無法儲存到 IndexedDB，使用備用存儲:', e);
				// 儲存到 sessionStorage 和 localStorage
				const keyData = JSON.stringify(Array.from(secretKey));
				sessionStorage.setItem('temp_secret_key', keyData);
				localStorage.setItem(`temp_key_${authStore.user?.id}`, keyData);
				status = '⚠️ 密鑰已儲存到臨時存儲';
			}
			
			// 更新 keysStore
			keysStore.secretKey = secretKey;
			keysStore.publicKey = userKeys.publicKey;
			
			status = '✅ 密鑰導入成功！即將跳轉...';
			
			setTimeout(() => {
				goto('/chat');
			}, 2000);
			
		} catch (error) {
			console.error('導入密鑰錯誤:', error);
			status = `❌ 錯誤: ${error}`;
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen bg-gradient-to-b from-slate-900 to-slate-950 text-white flex items-center justify-center p-4">
	<div class="max-w-md w-full">
		<div class="bg-slate-800/50 rounded-2xl p-6 border border-white/10">
			<h1 class="text-2xl font-bold mb-4">導入密鑰</h1>
			
			<div class="mb-6 p-4 bg-slate-700/50 rounded-lg">
				<p class="text-sm">{status}</p>
			</div>
			
			{#if authStore.user?.nickname === 'F' || authStore.user?.nickname === 'N'}
				<form onsubmit={(e) => { e.preventDefault(); importKeys(); }} class="space-y-4">
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
						{loading ? '導入中...' : '導入密鑰'}
					</button>
				</form>
			{:else}
				<div class="text-center text-slate-400">
					<p>此工具只供 F 和 N 使用</p>
				</div>
			{/if}
			
			<div class="mt-6 text-xs text-slate-500 text-center">
				<p>此工具會導入預設的密鑰對</p>
				<p>確保你可以正常收發加密訊息</p>
			</div>
		</div>
	</div>
</div>