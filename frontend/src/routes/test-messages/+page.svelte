<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores';
	import { messagesStore } from '$lib/stores/messages.svelte';
	import { conversationsStore } from '$lib/stores/conversations.svelte';
	import { transportStore } from '$lib/stores/transport.svelte';
	import { keysStore } from '$lib/stores/keys.svelte';
	
	let diagnostics = $state({
		auth: false,
		hasSecretKey: false,
		conversations: 0,
		activeConv: '',
		messages: 0,
		transport: false,
		logs: [] as string[]
	});
	
	function log(msg: string) {
		diagnostics.logs = [...diagnostics.logs, `${new Date().toISOString().substr(11, 8)} - ${msg}`];
	}
	
	onMount(async () => {
		// Check auth
		authStore.init();
		diagnostics.auth = authStore.isAuthenticated;
		log(`Auth status: ${diagnostics.auth}`);
		
		if (!diagnostics.auth) {
			log('Not authenticated - cannot continue');
			return;
		}
		
		// Check keys
		diagnostics.hasSecretKey = !!keysStore.secretKey;
		log(`Has secret key: ${diagnostics.hasSecretKey}`);
		
		// Load conversations
		log('Loading conversations...');
		await conversationsStore.loadConversations();
		diagnostics.conversations = conversationsStore.conversations.length;
		log(`Loaded ${diagnostics.conversations} conversations`);
		
		// Check transport
		diagnostics.transport = transportStore.connected;
		log(`Transport connected: ${diagnostics.transport}`);
		
		// If not connected, try to connect
		if (!diagnostics.transport && authStore.token) {
			log('Attempting to connect transport...');
			try {
				await transportStore.connect(authStore.token);
				diagnostics.transport = transportStore.connected;
				log(`Transport connected: ${diagnostics.transport}`);
			} catch (e) {
				log(`Transport connection failed: ${e}`);
			}
		}
		
		// Check first conversation
		if (conversationsStore.conversations.length > 0) {
			const conv = conversationsStore.conversations[0];
			diagnostics.activeConv = `${conv.peer.nickname} (${conv.id})`;
			log(`Checking conversation with ${conv.peer.nickname}`);
			
			// Load messages
			log('Loading messages...');
			await messagesStore.loadMessages(conv.id, conv.peer.public_key);
			
			// Get messages
			const messages = messagesStore.getMessages(conv.id);
			diagnostics.messages = messages.length;
			log(`Found ${messages.length} messages`);
			
			// Log first few messages
			messages.slice(0, 3).forEach((msg, i) => {
				log(`Message ${i + 1}: ${msg.senderId === authStore.user?.id ? 'Me' : conv.peer.nickname} - "${msg.content.substring(0, 30)}..."`);
			});
			
			// Check message store state
			const allConvs = Object.keys(messagesStore.messagesByConversation);
			log(`Message store has ${allConvs.length} conversations`);
			allConvs.forEach(convId => {
				const msgs = messagesStore.messagesByConversation[convId];
				log(`  Conv ${convId.substring(0, 8)}... has ${msgs?.length || 0} messages`);
			});
		}
	});
</script>

<div class="min-h-screen bg-slate-900 text-white p-4">
	<div class="max-w-4xl mx-auto">
		<h1 class="text-2xl font-bold mb-6">Message System Diagnostics</h1>
		
		<div class="space-y-4">
			<div class="bg-slate-800 rounded-lg p-4">
				<h2 class="font-semibold mb-2">System Status</h2>
				<div class="space-y-1 text-sm">
					<div>Auth: {diagnostics.auth ? '✅' : '❌'}</div>
					<div>Secret Key: {diagnostics.hasSecretKey ? '✅' : '❌'}</div>
					<div>Transport: {diagnostics.transport ? '✅' : '❌'}</div>
					<div>Conversations: {diagnostics.conversations}</div>
					<div>Active Conv: {diagnostics.activeConv || 'None'}</div>
					<div>Messages: {diagnostics.messages}</div>
				</div>
			</div>
			
			<div class="bg-slate-800 rounded-lg p-4">
				<h2 class="font-semibold mb-2">Diagnostic Logs</h2>
				<div class="space-y-1 text-xs font-mono">
					{#each diagnostics.logs as log}
						<div class="text-gray-400">{log}</div>
					{/each}
				</div>
			</div>
			
			<div class="bg-slate-800 rounded-lg p-4">
				<h2 class="font-semibold mb-2">Actions</h2>
				<div class="space-y-2">
					<button 
						onclick={() => window.location.reload()} 
						class="px-4 py-2 bg-blue-600 rounded hover:bg-blue-700"
					>
						Reload Page
					</button>
					<button 
						onclick={() => {
							localStorage.clear();
							sessionStorage.clear();
							window.location.href = '/';
						}} 
						class="px-4 py-2 bg-red-600 rounded hover:bg-red-700 ml-2"
					>
						Clear All & Logout
					</button>
				</div>
			</div>
		</div>
	</div>
</div>