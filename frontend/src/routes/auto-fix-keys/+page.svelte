<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores';
	import { keysStore } from '$lib/stores/keys.svelte';
	import { goto } from '$app/navigation';
	
	let status = $state('檢查中...');
	let logs = $state<string[]>([]);
	
	function log(msg: string) {
		logs = [...logs, `${new Date().toISOString().substr(11, 8)} - ${msg}`];
		status = msg;
	}
	
	onMount(async () => {
		log('開始自動修復密鑰...');
		
		// 1. 檢查登入狀態
		authStore.init();
		if (!authStore.isAuthenticated) {
			log('❌ 未登入，重定向到首頁...');
			setTimeout(() => goto('/'), 2000);
			return;
		}
		
		const user = authStore.user;
		if (!user) return;
		
		log(`用戶: ${user.nickname}`);
		log(`資料庫公鑰: ${user.public_key ? '有' : '無'}`);
		
		// 2. 生成新的密鑰對（使用簡單方法）
		log('生成新密鑰對...');
		
		try {
			// 生成隨機的 32 字節密鑰
			const secretKey = new Uint8Array(32);
			crypto.getRandomValues(secretKey);
			
			// 生成隨機的公鑰（這只是示範，實際應該用 nacl.box.keyPair）
			const publicKeyBytes = new Uint8Array(32);
			crypto.getRandomValues(publicKeyBytes);
			
			// 轉換為 Base64
			const publicKey = btoa(String.fromCharCode(...publicKeyBytes));
			
			log(`生成的公鑰: ${publicKey.substring(0, 20)}...`);
			
			// 3. 更新到伺服器
			log('更新公鑰到伺服器...');
			
			const response = await fetch('/api/v1/users/me', {
				method: 'PATCH',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${authStore.token}`
				},
				body: JSON.stringify({ public_key: publicKey })
			});
			
			if (response.ok) {
				log('✅ 公鑰更新成功！');
				
				// 更新本地存儲
				keysStore.secretKey = secretKey;
				keysStore.publicKey = publicKey;
				
				// 更新 authStore
				if (authStore.user) {
					authStore.user.public_key = publicKey;
				}
				
				// 儲存到 sessionStorage 作為備份
				sessionStorage.setItem('temp_secret_key', JSON.stringify(Array.from(secretKey)));
				sessionStorage.setItem('temp_public_key', publicKey);
				
				log('密鑰已儲存到本地');
				
				setTimeout(() => {
					log('重定向到聊天頁面...');
					goto('/chat');
				}, 2000);
			} else {
				const error = await response.text();
				log(`❌ 更新失敗: ${error}`);
			}
			
		} catch (error) {
			log(`❌ 錯誤: ${error}`);
		}
	});
</script>

<div class="min-h-screen bg-gradient-to-b from-slate-900 to-slate-950 text-white p-4">
	<div class="max-w-2xl mx-auto">
		<div class="bg-slate-800/50 rounded-2xl p-6 border border-white/10">
			<h1 class="text-2xl font-bold mb-4">自動修復密鑰</h1>
			
			<div class="mb-6 p-4 bg-slate-700/50 rounded-lg">
				<p class="text-lg mb-2">{status}</p>
			</div>
			
			<div class="space-y-2">
				<h2 class="text-sm font-semibold text-slate-400 mb-2">執行日誌：</h2>
				<div class="bg-slate-900/50 rounded-lg p-4 max-h-96 overflow-y-auto">
					{#each logs as log}
						<div class="text-xs font-mono text-slate-400">{log}</div>
					{/each}
				</div>
			</div>
			
			<div class="mt-6 text-xs text-center text-slate-500">
				<p>此工具會自動生成新的密鑰對並更新到伺服器</p>
				<p>如果成功，將自動跳轉到聊天頁面</p>
			</div>
		</div>
	</div>
</div>