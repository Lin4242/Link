<script lang="ts">
	import { onMount } from 'svelte';
	
	let status = $state('');
	let firstCard = $state('');
	let pairedToken = $state('');
	let cacheTime = $state('');
	
	function checkStatus() {
		firstCard = localStorage.getItem('register_first_card') || 'none';
		pairedToken = localStorage.getItem('register_paired_token') || 'none';
		cacheTime = localStorage.getItem('register_cache_time') || 'none';
		
		if (firstCard !== 'none' || pairedToken !== 'none') {
			status = '⚠️ Found cached registration data - this might block new registrations!';
		} else {
			status = '✅ No cached data - ready for new registration';
		}
	}
	
	function clearCache() {
		localStorage.removeItem('register_first_card');
		localStorage.removeItem('register_paired_token');
		localStorage.removeItem('register_cache_time');
		localStorage.removeItem('token');
		localStorage.removeItem('user');
		
		status = '🧹 Cache cleared successfully!';
		setTimeout(checkStatus, 1500);
	}
	
	onMount(() => {
		checkStatus();
	});
</script>

<div class="min-h-screen bg-gradient-to-b from-slate-900 to-slate-950 text-white p-4">
	<div class="max-w-lg mx-auto pt-8">
		<h1 class="text-2xl font-bold mb-6">Registration Debug Tool</h1>
		
		<div class="bg-slate-800/50 rounded-xl p-4 mb-6">
			<h2 class="text-lg font-semibold mb-3">Status</h2>
			<p class="text-sm {status.includes('⚠️') ? 'text-yellow-400' : 'text-green-400'}">
				{status}
			</p>
		</div>
		
		<div class="bg-slate-800/50 rounded-xl p-4 mb-6 font-mono text-sm">
			<p>First Card: <span class="text-blue-400">{firstCard}</span></p>
			<p>Paired Token: <span class="text-blue-400">{pairedToken}</span></p>
			<p>Cache Time: <span class="text-blue-400">{cacheTime}</span></p>
		</div>
		
		<div class="space-y-3">
			<button 
				onclick={clearCache}
				class="w-full py-3 bg-red-600 hover:bg-red-700 rounded-xl font-semibold transition-all"
			>
				Clear All Cache
			</button>
			
			<button 
				onclick={checkStatus}
				class="w-full py-3 bg-slate-700 hover:bg-slate-600 rounded-xl font-semibold transition-all"
			>
				Refresh Status
			</button>
			
			<a 
				href="/register"
				class="block w-full py-3 bg-blue-600 hover:bg-blue-700 rounded-xl font-semibold transition-all text-center"
			>
				Go to Register Page
			</a>
			
			<a 
				href="/register?reset=true"
				class="block w-full py-3 bg-orange-600 hover:bg-orange-700 rounded-xl font-semibold transition-all text-center"
			>
				Force Reset Registration
			</a>
			
			<a 
				href="/admin"
				class="block w-full py-3 bg-purple-600 hover:bg-purple-700 rounded-xl font-semibold transition-all text-center"
			>
				Go to Admin Panel
			</a>
		</div>
		
		<div class="mt-8 text-xs text-slate-500 text-center">
			This tool helps debug registration issues by clearing cached data
		</div>
	</div>
</div>