<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	// Redirect to new registration page
	onMount(() => {
		const token = $page.url.searchParams.get('token');
		if (token) {
			goto(`/register?token=${token}`);
		} else {
			goto('/register');
		}
	});
</script>

<div class="min-h-screen bg-gray-100 flex items-center justify-center p-4">
	<div class="animate-spin w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
</div>
