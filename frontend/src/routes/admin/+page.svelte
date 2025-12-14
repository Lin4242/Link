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

<div class="bg-gray-900 text-white">
	<div class="max-w-4xl mx-auto p-4 pb-20">
		<h1 class="text-2xl font-bold mb-6">LINK Admin</h1>

		{#if !isAuthenticated}
			<div class="bg-gray-800 rounded-lg p-6 max-w-sm mx-auto">
				<h2 class="text-lg font-semibold mb-4">輸入密碼</h2>
				<form onsubmit={(e) => { e.preventDefault(); login(); }}>
					<input
						type="password"
						bind:value={password}
						placeholder="Admin 密碼"
						class="w-full px-4 py-3 bg-black border-2 border-gray-500 rounded-lg mb-4 text-white placeholder-gray-400 focus:border-blue-500 focus:outline-none"
						inputmode="numeric"
						style="background-color: #000 !important; color: #fff !important; -webkit-text-fill-color: #fff;"
					/>
					<button
						type="submit"
						disabled={loading}
						class="w-full bg-blue-600 py-3 rounded-lg font-semibold hover:bg-blue-700 disabled:opacity-50"
					>
						{loading ? '驗證中...' : '登入'}
					</button>
				</form>
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
								<span class="text-sm text-gray-400">卡片對 #{pair.id}</span>
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
			<div class="mt-4 p-3 bg-red-900/50 text-red-300 rounded-lg text-sm">
				{error}
			</div>
		{/if}
	</div>
</div>
