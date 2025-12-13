<script lang="ts">
	import '../app.css';

	let { children } = $props();
</script>

<svelte:head>
	<title>LINK - 端對端加密通訊</title>
	<meta name="description" content="安全、私密的端對端加密即時通訊" />
	<meta name="theme-color" content="#2563eb" />
	<meta name="apple-mobile-web-app-capable" content="yes" />
	<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
</svelte:head>

{@render children()}
