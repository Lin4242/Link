<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores';

	let ready = $state(false);

	onMount(() => {
		// Don't auto-redirect, just show the landing page
		// User will scan NFC card to proceed
		ready = true;
	});
</script>

<div class="min-h-screen bg-gradient-to-b from-slate-900 via-slate-800 to-slate-900 flex items-center justify-center p-6">
	<div class="text-center text-white max-w-sm">
		<div class="w-20 h-20 mx-auto mb-6 rounded-md flex items-center justify-center shadow-lg" style="background: linear-gradient(135deg, #3ACACA, #2BA3A3); box-shadow: 0 10px 15px -3px rgba(58, 202, 202, 0.3);">
			<svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
			</svg>
		</div>

		<h1 class="text-3xl font-bold tracking-tight mb-2">LINK</h1>
		<p class="text-slate-400 mb-10">端對端加密通訊</p>

		<div class="space-y-3 text-left mb-12">
			<div class="flex items-center gap-3">
				<svg class="w-5 h-5 text-slate-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
				</svg>
				<span class="text-slate-300 text-sm">NFC 雙卡安全認證</span>
			</div>
			<div class="flex items-center gap-3">
				<svg class="w-5 h-5 text-slate-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
				</svg>
				<span class="text-slate-300 text-sm">端對端加密訊息</span>
			</div>
			<div class="flex items-center gap-3">
				<svg class="w-5 h-5 text-slate-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				<span class="text-slate-300 text-sm">零知識驗證</span>
			</div>
		</div>

		<div class="pt-8 border-t border-white/10">
			<p class="text-slate-500 text-xs mb-4">使用 NFC 卡片開始</p>
			<div class="w-16 h-16 mx-auto border-2 border-dashed border-slate-600 rounded-md flex items-center justify-center">
				<svg class="w-7 h-7 text-slate-500 animate-pulse" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A13.916 13.916 0 008 11a4 4 0 118 0c0 1.017-.07 2.019-.203 3m-2.118 6.844A21.88 21.88 0 0015.171 17m3.839 1.132c.645-2.266.99-4.659.99-7.132A8 8 0 008 4.07M3 15.364c.64-1.319 1-2.8 1-4.364 0-1.457.39-2.823 1.07-4" />
				</svg>
			</div>
		</div>
	</div>
</div>
