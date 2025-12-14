<script lang="ts">
	import { onMount } from 'svelte';

	interface CardPair {
		id: number;
		first_token: string;
		second_token: string;
		first_url: string;
		second_url: string;
	}

	let password = $state('');
	let isAuthenticated = $state(false);
	let cardPairs = $state<CardPair[]>([]);
	let loading = $state(false);
	let error = $state('');
	let copiedId = $state<string | null>(null);

	const API_BASE = import.meta.env.VITE_API_URL ? 
		`${import.meta.env.VITE_API_URL}/api/v1` : 
		`${window.location.origin}/api/v1`;

	async function login() {
		error = '';
		loading = true;

		try {
			const res = await fetch(`${API_BASE}/admin/cards`, {
				headers: { 'X-Admin-Password': password }
			});

			if (res.ok) {
				isAuthenticated = true;
				const data = await res.json();
				cardPairs = data.data || [];
			} else {
				error = '密碼錯誤';
			}
		} catch (e) {
			error = '連線失敗';
		}

		loading = false;
	}

	async function generatePair() {
		loading = true;
		error = '';

		try {
			const res = await fetch(`${API_BASE}/admin/cards/generate`, {
				method: 'POST',
				headers: { 'X-Admin-Password': password }
			});

			if (res.ok) {
				const data = await res.json();
				cardPairs = [...cardPairs, data.data];
			} else {
				error = '產生失敗';
			}
		} catch (e) {
			error = '連線失敗';
		}

		loading = false;
	}

	async function deletePair(id: number) {
		try {
			const res = await fetch(`${API_BASE}/admin/cards/${id}`, {
				method: 'DELETE',
				headers: { 'X-Admin-Password': password }
			});

			if (res.ok) {
				cardPairs = cardPairs.filter(p => p.id !== id);
			}
		} catch (e) {
			error = '刪除失敗';
		}
	}

	async function copyToClipboard(text: string, id: string) {
		try {
			await navigator.clipboard.writeText(text);
			copiedId = id;
			setTimeout(() => copiedId = null, 2000);
		} catch (e) {
			// Fallback for mobile
			const input = document.createElement('input');
			input.value = text;
			document.body.appendChild(input);
			input.select();
			document.execCommand('copy');
			document.body.removeChild(input);
			copiedId = id;
			setTimeout(() => copiedId = null, 2000);
		}
	}
</script>

<div class="min-h-screen bg-gradient-to-b from-slate-950 via-slate-900 to-slate-950 text-white flex flex-col">
	<div class="max-w-4xl w-full mx-auto p-4 pb-20 flex-1">
		<h1 class="text-lg font-light tracking-wider text-slate-400 mb-8 mt-2">LINK</h1>

		{#if !isAuthenticated}
			<div class="flex items-center justify-center min-h-[60vh]">
				<div class="w-full max-w-sm">
					<form onsubmit={(e) => { e.preventDefault(); login(); }} class="space-y-4">
						<input
							type="password"
							bind:value={password}
							placeholder=""
							class="w-full px-4 py-3 bg-slate-800/30 border border-slate-700/50 rounded-lg text-white placeholder-slate-600 focus:border-slate-600 focus:outline-none focus:ring-1 focus:ring-slate-600/50 transition-all"
							inputmode="numeric"
							style="background-color: rgba(30, 41, 59, 0.3); -webkit-text-fill-color: #fff;"
							autofocus
						/>
						<button
							type="submit"
							disabled={loading}
							class="w-full py-3 bg-slate-800/50 hover:bg-slate-700/50 disabled:opacity-30 rounded-lg font-normal transition-all duration-200 flex items-center justify-center"
						>
							{#if loading}
								<div class="animate-spin w-5 h-5 border-2 border-slate-600 border-t-transparent rounded-full"></div>
							{:else}
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 7l5 5m0 0l-5 5m5-5H6" />
								</svg>
							{/if}
						</button>
					</form>
				</div>
			</div>
		{:else}
			<div class="mb-6">
				<button
					onclick={generatePair}
					disabled={loading}
					class="w-full bg-green-600 px-6 py-4 rounded-lg font-bold text-lg hover:bg-green-700 disabled:opacity-50 border-2 border-green-400"
				>
					{loading ? '產生中...' : '+ 產生新卡片對'}
				</button>
			</div>

			{#if cardPairs.length === 0}
				<div class="text-center text-gray-400 py-8">
					點擊上方綠色按鈕產生卡片
				</div>
			{:else}
				<div class="space-y-4">
					{#each cardPairs as pair (pair.id)}
						<div class="bg-gray-800 rounded-lg p-4">
							<div class="flex justify-between items-start mb-3">
								<div class="flex items-center gap-2">
									<span class="text-sm text-gray-400">卡片對 #{pair.id}</span>
									{#if pair.is_activated}
										<span class="flex items-center gap-1 text-xs px-2 py-1 bg-green-600/20 text-green-400 rounded-full">
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
											</svg>
											已開卡
										</span>
									{:else}
										<span class="text-xs px-2 py-1 bg-yellow-600/20 text-yellow-400 rounded-full">
											未開卡
										</span>
									{/if}
								</div>
								<button
									onclick={() => deletePair(pair.id)}
									class="text-red-400 hover:text-red-300 text-sm"
								>
									刪除
								</button>
							</div>

							<!-- Card A -->
							<div class="mb-4">
								<div class="flex items-center gap-2 mb-2">
									<span class="bg-purple-600 text-xs px-2 py-1 rounded">卡片 A</span>
								</div>
								<div class="bg-gray-700 rounded p-3 flex items-center gap-2">
									<input
										type="text"
										readonly
										value={pair.first_url}
										class="flex-1 bg-transparent text-sm text-gray-300 outline-none"
									/>
									<button
										onclick={() => copyToClipboard(pair.first_url, `a-${pair.id}`)}
										class="px-3 py-1 bg-purple-600 rounded text-sm hover:bg-purple-700 whitespace-nowrap"
									>
										{copiedId === `a-${pair.id}` ? '已複製' : '複製'}
									</button>
								</div>
							</div>

							<!-- Card B -->
							<div>
								<div class="flex items-center gap-2 mb-2">
									<span class="bg-purple-600 text-xs px-2 py-1 rounded">卡片 B</span>
								</div>
								<div class="bg-gray-700 rounded p-3 flex items-center gap-2">
									<input
										type="text"
										readonly
										value={pair.second_url}
										class="flex-1 bg-transparent text-sm text-gray-300 outline-none"
									/>
									<button
										onclick={() => copyToClipboard(pair.second_url, `b-${pair.id}`)}
										class="px-3 py-1 bg-purple-600 rounded text-sm hover:bg-purple-700 whitespace-nowrap"
									>
										{copiedId === `b-${pair.id}` ? '已複製' : '複製'}
									</button>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		{/if}

		{#if error}
			<div class="fixed bottom-8 left-1/2 -translate-x-1/2 px-4 py-2 bg-slate-800/80 backdrop-blur text-slate-400 rounded-lg text-sm">
				{error}
			</div>
		{/if}
	</div>
</div>
