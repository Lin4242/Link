<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		authStore,
		keysStore,
		conversationsStore,
		messagesStore,
		friendsStore,
		transportStore,
	} from '$lib/stores';
	import { encryptToString } from '$lib/crypto';
	import type { EncryptedMessage } from '$lib/types';

	let messageInput = $state('');
	let sending = $state(false);
	let messagesContainer: HTMLDivElement | undefined = undefined;
	let typingUsers = $state<Record<string, number>>({});
	let privacyScreen = $state(false); // 隱私保護遮罩

	const activeConversation = $derived(
		conversationsStore.conversations.find((c) => c.id === conversationsStore.activeConversationId)
	);

	const messages = $derived(
		conversationsStore.activeConversationId
			? messagesStore.getMessages(conversationsStore.activeConversationId)
			: []
	);

	const isTyping = $derived(
		activeConversation && typingUsers[activeConversation.peer.id] ? true : false
	);

	// 隱私保護：切換分頁或切換 app 時顯示遮罩
	function handleVisibilityChange() {
		if (document.hidden) {
			privacyScreen = true;
		}
		// 注意：回來時不自動關閉，需要點擊才能解除
	}

	function handleWindowBlur() {
		// blur 事件比 visibilitychange 更早觸發，可以更快顯示遮罩
		privacyScreen = true;
	}

	onMount(async () => {
		authStore.init();

		if (!authStore.isAuthenticated || !authStore.token) {
			localStorage.removeItem('link_auth');
			window.location.replace('/');
			return;
		}

		// 監聽頁面可見性變化和視窗失焦
		document.addEventListener('visibilitychange', handleVisibilityChange);
		window.addEventListener('blur', handleWindowBlur);

		// Load data
		await Promise.all([
			conversationsStore.loadConversations(),
			friendsStore.loadFriends().catch(() => {}), // Ignore errors for now
			friendsStore.loadPendingRequests().catch(() => {}),
		]);

		// Connect to transport
		try {
			await transportStore.connect(authStore.token);
			setupTransportHandlers();
		} catch (e) {
			console.error('Transport connection failed:', e);
		}
	});

	onDestroy(() => {
		document.removeEventListener('visibilitychange', handleVisibilityChange);
		window.removeEventListener('blur', handleWindowBlur);
		transportStore.disconnect();
	});

	function setupTransportHandlers() {
		transportStore.onMessage((msg: EncryptedMessage) => {
			const conv = conversationsStore.conversations.find((c) => c.id === msg.conversation_id);
			if (conv) {
				const decrypted = messagesStore.receiveMessage(msg, conv.peer.public_key);
				if (decrypted) {
					conversationsStore.updateLastMessage(msg.conversation_id, msg.created_at);
					if (msg.conversation_id !== conversationsStore.activeConversationId) {
						conversationsStore.incrementUnread(msg.conversation_id);
					} else {
						transportStore.sendRead(msg.conversation_id, msg.id);
					}
					scrollToBottom();
				}
			}
		});

		transportStore.onTyping((convId: string, userId: string) => {
			typingUsers[userId] = Date.now();
			setTimeout(() => {
				if (Date.now() - typingUsers[userId] >= 2000) {
					delete typingUsers[userId];
					typingUsers = { ...typingUsers };
				}
			}, 3000);
		});

		transportStore.onOnline((userId: string) => {
			friendsStore.setOnline(userId, true);
		});

		transportStore.onOffline((userId: string) => {
			friendsStore.setOnline(userId, false);
		});

		transportStore.onDelivered((tempId: string, msg: EncryptedMessage) => {
			console.log('onDelivered callback received:', { tempId, msg });
			const conv = conversationsStore.conversations.find((c) => c.id === msg.conversation_id);
			console.log('Found conversation:', conv?.id, 'hasSecretKey:', !!keysStore.secretKey);
			if (conv && keysStore.secretKey) {
				const content = messages.find((m) => m.tempId === tempId)?.content || '';
				console.log('Confirming message with content:', content);
				messagesStore.confirmMessage(tempId, {
					id: msg.id,
					conversationId: msg.conversation_id,
					senderId: msg.sender_id,
					content,
					createdAt: msg.created_at,
					deliveredAt: msg.delivered_at,
				});
				conversationsStore.updateLastMessage(msg.conversation_id, msg.created_at);
			}
		});
	}

	async function selectConversation(id: string) {
		conversationsStore.setActive(id);
		const conv = conversationsStore.conversations.find((c) => c.id === id);
		if (conv) {
			await messagesStore.loadMessages(id, conv.peer.public_key);
			scrollToBottom();
		}
	}

	async function sendMessage() {
		if (!messageInput.trim() || !activeConversation || !keysStore.secretKey || sending) {
			alert('Cannot send: ' + JSON.stringify({
				hasInput: !!messageInput.trim(),
				hasConv: !!activeConversation,
				hasKey: !!keysStore.secretKey,
				sending
			}));
			return;
		}

		sending = true;
		const content = messageInput.trim();
		messageInput = '';

		try {
			const tempId = crypto.randomUUID();
			const peerPubKey = activeConversation.peer.public_key;

			if (!peerPubKey) {
				alert('Peer has no public key!');
				sending = false;
				return;
			}

			const encrypted = encryptToString(content, peerPubKey, keysStore.secretKey);

			messagesStore.addPendingMessage(
				activeConversation.id,
				tempId,
				authStore.user!.id,
				content
			);
			scrollToBottom();

			await transportStore.sendMessage(activeConversation.peer.id, encrypted, tempId);
		} catch (e) {
			alert('Send error: ' + (e instanceof Error ? e.message : String(e)));
		}
		sending = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			sendMessage();
		}
	}

	let typingTimeout: ReturnType<typeof setTimeout>;
	function handleTyping() {
		if (!activeConversation) return;
		clearTimeout(typingTimeout);
		typingTimeout = setTimeout(() => {
			transportStore.sendTyping(activeConversation.peer.id, activeConversation.id);
		}, 300);
	}

	function scrollToBottom() {
		setTimeout(() => {
			if (messagesContainer) {
				messagesContainer.scrollTop = messagesContainer.scrollHeight;
			}
		}, 50);
	}

	function formatTime(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleTimeString('zh-TW', { hour: '2-digit', minute: '2-digit' });
	}

	function logout() {
		transportStore.disconnect();
		keysStore.lock();
		messagesStore.clear();
		authStore.logout();
		goto('/');
	}
</script>

<!-- Privacy screen - 隱私保護遮罩 -->
{#if privacyScreen}
<div class="fixed inset-0 bg-gradient-to-br from-blue-600 to-blue-800 flex items-center justify-center z-[100]">
	<div class="text-center text-white w-full max-w-xs px-6">
		<svg class="w-16 h-16 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
		</svg>
		<h2 class="text-2xl font-bold mb-1">LINK</h2>
		<p class="text-blue-200 text-sm mb-6">端對端加密通訊</p>
		<form onsubmit={async (e) => {
			e.preventDefault();
			const form = e.target as HTMLFormElement;
			const pwd = (form.elements.namedItem('privacyPwd') as HTMLInputElement).value;
			if (!pwd) return;
			const success = await keysStore.unlock(pwd);
			if (success) {
				privacyScreen = false;
			} else {
				alert('密碼錯誤');
			}
			(form.elements.namedItem('privacyPwd') as HTMLInputElement).value = '';
		}}>
			<input
				type="password"
				name="privacyPwd"
				placeholder="輸入密碼解鎖"
				class="w-full px-4 py-3 rounded-lg bg-white/20 text-white placeholder-blue-200 border border-white/30 focus:outline-none focus:ring-2 focus:ring-white/50 mb-3"
			/>
			<button type="submit" class="w-full py-3 bg-white text-blue-600 font-medium rounded-lg">
				解鎖
			</button>
		</form>
	</div>
</div>
{/if}

<!-- Key unlock modal -->
{#if transportStore.connected && !keysStore.secretKey}
<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
	<div class="bg-white rounded-lg p-6 max-w-sm w-full">
		<h2 class="text-lg font-bold mb-4">解鎖加密金鑰</h2>
		<p class="text-sm text-gray-600 mb-4">請輸入密碼來解鎖您的加密金鑰。如果密碼錯誤，將產生新的金鑰。</p>
		<form onsubmit={async (e) => {
			e.preventDefault();
			const form = e.target as HTMLFormElement;
			const pwd = (form.elements.namedItem('unlockPwd') as HTMLInputElement).value;
			if (!pwd) {
				alert('請輸入密碼');
				return;
			}
			console.log('Attempting to unlock key with password...');
			// Try to unlock first
			let success = await keysStore.unlock(pwd);
			console.log('Unlock result:', success);
			if (!success) {
				// Warn user that old messages will be lost
				const confirmed = confirm('密碼錯誤或金鑰不存在。是否要產生新的金鑰？\n\n警告：這將導致舊訊息無法解密！');
				if (!confirmed) {
					return;
				}
				// Generate new keys
				const { generateKeyPair, saveSecretKey } = await import('$lib/crypto/keys');
				const { publicKey, secretKey } = generateKeyPair();
				await saveSecretKey(secretKey, pwd);
				await keysStore.save(secretKey, pwd);
				// Update public key on server
				const { updateMe } = await import('$lib/api/users');
				await updateMe({ public_key: publicKey });
				alert('金鑰已重新產生。舊訊息將無法解密。');
			} else {
				console.log('Key unlocked successfully!');
			}
			// Reload messages if there's an active conversation
			if (activeConversation) {
				console.log('Reloading messages for active conversation:', activeConversation.id);
				await messagesStore.loadMessages(
					activeConversation.id,
					activeConversation.peer.public_key
				);
			}
		}}>
			<input
				type="password"
				name="unlockPwd"
				placeholder="輸入密碼"
				class="w-full px-4 py-2 border rounded-lg mb-4"
				autofocus
			/>
			<button type="submit" class="w-full bg-blue-600 text-white py-2 rounded-lg">
				解鎖 / 產生金鑰
			</button>
		</form>
	</div>
</div>
{/if}

<div class="h-[100dvh] flex bg-gray-100 overflow-hidden {!transportStore.connected ? 'pt-8' : ''}">
	<!-- Sidebar -->
	<div class="w-full md:w-80 bg-white border-r flex flex-col flex-shrink-0 {activeConversation ? 'hidden md:flex' : 'flex'}">
		<!-- Header -->
		<div class="p-4 border-b flex items-center justify-between">
			<div class="flex items-center gap-3">
				<div class="w-10 h-10 bg-blue-600 rounded-full flex items-center justify-center text-white font-medium">
					{authStore.user?.nickname?.[0] || '?'}
				</div>
				<div>
					<p class="font-medium">{authStore.user?.nickname}</p>
					<p class="text-xs text-gray-500 flex items-center gap-1">
						<span class="w-2 h-2 rounded-full {transportStore.connected ? 'bg-green-500' : 'bg-red-500'}"></span>
						{transportStore.connected ? '已連線' : '離線'}
						{#if transportStore.transportType === 'webtransport'}
							<span class="text-blue-500">(WT)</span>
						{/if}
					</p>
				</div>
			</div>
			<button onclick={logout} class="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg" title="登出">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
				</svg>
			</button>
		</div>

		<!-- Conversation List -->
		<div class="flex-1 overflow-y-auto">
			{#if conversationsStore.loading}
				<div class="p-4 text-center text-gray-500">載入中...</div>
			{:else if conversationsStore.conversations.length === 0}
				<div class="p-4 text-center text-gray-500">
					<p class="mb-2">還沒有對話</p>
					<p class="text-sm">新增好友開始聊天</p>
				</div>
			{:else}
				{#each conversationsStore.conversations as conv}
					<button
						onclick={() => selectConversation(conv.id)}
						class="w-full p-4 flex items-center gap-3 hover:bg-gray-50 transition-colors border-b
							{conversationsStore.activeConversationId === conv.id ? 'bg-blue-50' : ''}"
					>
						<div class="relative">
							<div class="w-12 h-12 bg-gray-200 rounded-full flex items-center justify-center text-lg font-medium">
								{conv.peer.nickname[0]}
							</div>
							{#if friendsStore.friends.find((f) => f.id === conv.peer.id)?.isOnline}
								<div class="absolute bottom-0 right-0 w-3 h-3 bg-green-500 border-2 border-white rounded-full"></div>
							{/if}
						</div>
						<div class="flex-1 text-left min-w-0">
							<div class="flex items-center justify-between">
								<p class="font-medium truncate">{conv.peer.nickname}</p>
								{#if conv.lastMessageAt}
									<span class="text-xs text-gray-500">{formatTime(conv.lastMessageAt)}</span>
								{/if}
							</div>
							<div class="flex items-center justify-between">
								<p class="text-sm text-gray-500 truncate">
									{#if typingUsers[conv.peer.id]}
										<span class="text-blue-500">正在輸入...</span>
									{:else}
										點擊開始聊天
									{/if}
								</p>
								{#if conv.unreadCount > 0}
									<span class="bg-blue-600 text-white text-xs px-2 py-0.5 rounded-full min-w-[20px] text-center">
										{conv.unreadCount > 99 ? '99+' : conv.unreadCount}
									</span>
								{/if}
							</div>
						</div>
					</button>
				{/each}
			{/if}
		</div>

		<!-- Pending Requests -->
		{#if friendsStore.pendingRequests.length > 0}
			<div class="border-t p-2">
				<p class="text-xs text-gray-500 px-2 mb-1">好友請求 ({friendsStore.pendingRequests.length})</p>
			</div>
		{/if}
	</div>

	<!-- Main Chat Area -->
	<div class="flex-1 flex flex-col min-w-0 {activeConversation ? 'flex' : 'hidden md:flex'}">
		{#if !activeConversation}
			<div class="flex-1 flex items-center justify-center text-gray-500">
				<div class="text-center">
					<svg class="w-16 h-16 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
					</svg>
					<p>選擇一個對話開始聊天</p>
				</div>
			</div>
		{:else}
			<!-- Chat Header -->
			<div class="p-4 border-b bg-white flex items-center gap-3">
				<!-- Back button for mobile -->
				<button
					onclick={() => conversationsStore.setActive(null)}
					class="md:hidden p-2 -ml-2 text-gray-500 hover:text-gray-700"
				>
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
				</button>
				<div class="relative">
					<div class="w-10 h-10 bg-gray-200 rounded-full flex items-center justify-center font-medium">
						{activeConversation.peer.nickname[0]}
					</div>
					{#if friendsStore.friends.find((f) => f.id === activeConversation.peer.id)?.isOnline}
						<div class="absolute bottom-0 right-0 w-2.5 h-2.5 bg-green-500 border-2 border-white rounded-full"></div>
					{/if}
				</div>
				<div>
					<p class="font-medium">{activeConversation.peer.nickname}</p>
					<p class="text-xs text-gray-500">
						{#if isTyping}
							<span class="text-blue-500">正在輸入...</span>
						{:else if friendsStore.friends.find((f) => f.id === activeConversation.peer.id)?.isOnline}
							在線上
						{:else}
							離線
						{/if}
					</p>
				</div>
				<div class="ml-auto flex items-center gap-1 text-xs text-gray-400">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
					</svg>
					端對端加密
				</div>
			</div>

			<!-- Messages -->
			<div bind:this={messagesContainer} class="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-50">
				{#if messagesStore.loading}
					<div class="text-center text-gray-500">載入訊息中...</div>
				{:else if messages.length === 0}
					<div class="text-center text-gray-500 py-8">
						<svg class="w-12 h-12 mx-auto mb-2 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
						</svg>
						<p class="text-sm">訊息已加密保護</p>
						<p class="text-xs">發送第一則訊息開始聊天</p>
					</div>
				{:else}
					{#each messages as msg}
						{@const isOwn = msg.senderId === authStore.user?.id}
						<div class="flex {isOwn ? 'justify-end' : 'justify-start'}">
							<div class="max-w-[70%] {isOwn ? 'bg-blue-600 text-white' : 'bg-white'} rounded-2xl px-4 py-2 shadow-sm
								{msg.pending ? 'opacity-70' : ''}">
								<p class="break-words whitespace-pre-wrap">{msg.content}</p>
								<p class="text-xs {isOwn ? 'text-blue-200' : 'text-gray-400'} mt-1 text-right">
									{formatTime(msg.createdAt)}
									{#if isOwn && msg.pending}
										<span class="ml-1">發送中</span>
									{/if}
								</p>
							</div>
						</div>
					{/each}
				{/if}
			</div>

			<!-- Message Input -->
			<div class="p-4 bg-white border-t">
				{#if !transportStore.connected}
					<div class="text-center text-sm text-yellow-600 mb-2">
						未連線 - 訊息功能暫時無法使用
					</div>
				{:else if !keysStore.secretKey}
					<div class="text-center text-sm text-yellow-600 mb-2">
						加密金鑰未載入 - 需要 HTTPS 環境
					</div>
				{/if}
				<form onsubmit={(e) => { e.preventDefault(); sendMessage(); }} class="flex gap-2">
					<input
						type="text"
						bind:value={messageInput}
						onkeydown={handleKeydown}
						oninput={handleTyping}
						placeholder="輸入訊息..."
						class="flex-1 px-4 py-2 border border-gray-300 rounded-full focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
						disabled={sending}
					/>
					<button
						type="submit"
						disabled={!messageInput.trim() || sending || !transportStore.connected || !keysStore.secretKey}
						class="w-10 h-10 bg-blue-600 text-white rounded-full flex items-center justify-center hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
						aria-label="發送訊息"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
						</svg>
					</button>
				</form>
			</div>
		{/if}
	</div>
</div>
