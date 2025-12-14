<script>
	import { onMount } from 'svelte';
	
	onMount(() => {
		// Clear all localStorage
		localStorage.clear();
		
		// Clear all sessionStorage  
		sessionStorage.clear();
		
		// Clear all cookies
		document.cookie.split(";").forEach((c) => {
			document.cookie = c.replace(/^ +/, "").replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/");
		});
		
		// Redirect to register with card 9 token
		const card9Token = 'e894c8251988443a-1-ce6c9db0';
		setTimeout(() => {
			window.location.href = `/register?reset=true&token=${card9Token}`;
		}, 1000);
	});
</script>

<div class="min-h-screen bg-slate-900 text-white flex items-center justify-center">
	<div class="text-center">
		<div class="animate-spin w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full mx-auto mb-4"></div>
		<p class="text-xl">正在清除所有快取...</p>
		<p class="text-sm text-gray-400 mt-2">即將跳轉到卡片9註冊頁面</p>
	</div>
</div>