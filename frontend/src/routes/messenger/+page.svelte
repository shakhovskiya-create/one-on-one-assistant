<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { messenger, employees as employeesApi } from '$lib/api/client';
	import type { Conversation, Message, Employee } from '$lib/api/client';
	import { user } from '$lib/stores/auth';

	let conversations: Conversation[] = $state([]);
	let currentConversation: Conversation | null = $state(null);
	let messages: Message[] = $state([]);
	let employees: Employee[] = $state([]);
	let newMessage = $state('');
	let loading = $state(true);
	let loadingMessages = $state(false);
	let showNewChat = $state(false);
	let selectedParticipants: string[] = $state([]);
	let groupName = $state('');
	let ws: WebSocket | null = null;
	let typingUsers: Record<string, string> = $state({});
	let messagesContainer: HTMLDivElement;
	let searchQuery = $state('');

	$effect(() => {
		if ($user?.id) {
			loadConversations();
			loadEmployees();
			connectWebSocket();
		}
	});

	onDestroy(() => {
		if (ws) {
			ws.close();
		}
	});

	function connectWebSocket() {
		if (!$user?.id) return;

		const wsUrl = messenger.getWebSocketUrl($user.id);
		ws = new WebSocket(wsUrl);

		ws.onopen = () => {
			console.log('WebSocket connected');
		};

		ws.onmessage = (event) => {
			const data = JSON.parse(event.data);
			handleWSMessage(data);
		};

		ws.onclose = () => {
			console.log('WebSocket disconnected');
			setTimeout(connectWebSocket, 3000);
		};

		ws.onerror = (err) => {
			console.error('WebSocket error:', err);
		};
	}

	function handleWSMessage(data: { type: string; conversation_id?: string; message?: Message; data?: any }) {
		switch (data.type) {
			case 'new_message':
				if (data.message) {
					// Avoid duplicates - don't add if message already exists
					if (currentConversation?.id === data.conversation_id) {
						const exists = messages.some(m => m.id === data.message!.id);
						if (!exists) {
							messages = [...messages, data.message];
							scrollToBottom();
						}
					}
					updateConversationLastMessage(data.conversation_id!, data.message);
				}
				break;
			case 'typing':
				if (data.conversation_id && data.data?.user_id && data.data.user_id !== $user?.id) {
					typingUsers = { ...typingUsers, [data.conversation_id]: data.data.user_id };
					setTimeout(() => {
						const { [data.conversation_id!]: _, ...rest } = typingUsers;
						typingUsers = rest;
					}, 3000);
				}
				break;
		}
	}

	function updateConversationLastMessage(convId: string, message: Message) {
		conversations = conversations.map((c) =>
			c.id === convId ? { ...c, last_message: message, updated_at: message.created_at } : c
		);
		conversations = [...conversations].sort((a, b) =>
			new Date(b.updated_at || 0).getTime() - new Date(a.updated_at || 0).getTime()
		);
	}

	async function loadConversations() {
		if (!$user?.id) return;
		try {
			conversations = await messenger.listConversations($user.id);
		} catch (e) {
			console.error('Failed to load conversations:', e);
		} finally {
			loading = false;
		}
	}

	async function loadEmployees() {
		try {
			employees = await employeesApi.list();
		} catch (e) {
			console.error('Failed to load employees:', e);
		}
	}

	async function selectConversation(conv: Conversation) {
		currentConversation = conv;
		loadingMessages = true;
		try {
			const result = await messenger.getConversation(conv.id);
			messages = result.messages || [];
			// Update currentConversation with participants from API
			if (result.participants) {
				currentConversation = { ...conv, participants: result.participants };
				// Also update in the list
				conversations = conversations.map(c =>
					c.id === conv.id ? { ...c, participants: result.participants } : c
				);
			}
			scrollToBottom();
		} catch (e) {
			console.error('Failed to load messages:', e);
			messages = [];
		} finally {
			loadingMessages = false;
		}
	}

	async function sendMessage() {
		if (!newMessage.trim() || !currentConversation || !$user?.id) return;

		const content = newMessage;
		newMessage = '';

		try {
			const msg = await messenger.sendMessage({
				conversation_id: currentConversation.id,
				sender_id: $user.id,
				content
			});
			messages = [...messages, msg];
			scrollToBottom();
		} catch (e) {
			console.error('Failed to send message:', e);
			newMessage = content;
		}
	}

	function sendTyping() {
		if (!ws || !currentConversation) return;
		ws.send(JSON.stringify({
			type: 'typing',
			conversation_id: currentConversation.id
		}));
	}

	async function createConversation() {
		if (selectedParticipants.length === 0 || !$user?.id) return;

		const participants = [...selectedParticipants, $user.id];
		const isGroup = participants.length > 2;

		try {
			const conv = await messenger.createConversation({
				type: isGroup ? 'group' : 'direct',
				name: isGroup ? groupName : undefined,
				participants
			});
			conversations = [conv, ...conversations];
			showNewChat = false;
			selectedParticipants = [];
			groupName = '';
			selectConversation(conv);
		} catch (e) {
			console.error('Failed to create conversation:', e);
		}
	}

	function scrollToBottom() {
		setTimeout(() => {
			if (messagesContainer) {
				messagesContainer.scrollTop = messagesContainer.scrollHeight;
			}
		}, 50);
	}

	function getConversationName(conv: Conversation): string {
		if (conv.name) return conv.name;
		if (conv.participants) {
			const others = conv.participants.filter((p) => p.id !== $user?.id);
			return others.map((p) => p.name).join(', ') || 'Чат';
		}
		return 'Чат';
	}

	function getConversationAvatar(conv: Conversation): string | null {
		if (conv.type === 'direct' && conv.participants) {
			const other = conv.participants.find((p) => p.id !== $user?.id);
			return other?.photo_base64 || null;
		}
		return null;
	}

	function formatTime(dateStr?: string): string {
		if (!dateStr) return '';
		const date = new Date(dateStr);
		const now = new Date();
		const isToday = date.toDateString() === now.toDateString();

		if (isToday) {
			return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
		}
		return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
	}

	function toggleParticipant(id: string) {
		if (selectedParticipants.includes(id)) {
			selectedParticipants = selectedParticipants.filter((p) => p !== id);
		} else {
			selectedParticipants = [...selectedParticipants, id];
		}
	}

	let filteredEmployees = $derived(
		employees.filter((e) =>
			e.id !== $user?.id &&
			e.name.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);
</script>

<svelte:head>
	<title>Мессенджер - EKF Team Hub</title>
</svelte:head>

<div class="h-[calc(100vh-8rem)] flex bg-white rounded-xl shadow-sm overflow-hidden">
	<!-- Sidebar - Conversations List -->
	<div class="w-80 border-r border-gray-200 flex flex-col">
		<div class="p-4 border-b border-gray-200">
			<div class="flex items-center justify-between mb-3">
				<h2 class="text-lg font-semibold text-gray-900">Сообщения</h2>
				<button
					onclick={() => (showNewChat = !showNewChat)}
					class="p-2 text-ekf-red hover:bg-red-50 rounded-lg transition-colors"
					title="Новый чат"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
				</button>
			</div>
			<input
				type="text"
				placeholder="Поиск..."
				class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
			/>
		</div>

		<div class="flex-1 overflow-y-auto">
			{#if loading}
				<div class="flex items-center justify-center h-32">
					<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-ekf-red"></div>
				</div>
			{:else if conversations.length === 0}
				<div class="text-center py-8 text-gray-500">
					<p>Нет чатов</p>
					<button
						onclick={() => (showNewChat = true)}
						class="text-ekf-red hover:underline mt-2"
					>
						Начать новый чат
					</button>
				</div>
			{:else}
				{#each conversations as conv}
					<button
						onclick={() => selectConversation(conv)}
						class="w-full p-4 flex items-center gap-3 hover:bg-gray-50 transition-colors border-b border-gray-100
							{currentConversation?.id === conv.id ? 'bg-red-50' : ''}"
					>
						{#if getConversationAvatar(conv)}
							<img
								src="data:image/jpeg;base64,{getConversationAvatar(conv)}"
								alt=""
								class="w-12 h-12 rounded-full object-cover"
							/>
						{:else}
							<div class="w-12 h-12 rounded-full bg-ekf-red text-white flex items-center justify-center font-medium">
								{getConversationName(conv).charAt(0)}
							</div>
						{/if}
						<div class="flex-1 min-w-0 text-left">
							<div class="font-medium text-gray-900 truncate">{getConversationName(conv)}</div>
							{#if conv.last_message}
								<div class="text-sm text-gray-500 truncate">
									{conv.last_message.content}
								</div>
							{/if}
						</div>
						{#if conv.last_message?.created_at}
							<div class="text-xs text-gray-400">
								{formatTime(conv.last_message.created_at)}
							</div>
						{/if}
					</button>
				{/each}
			{/if}
		</div>
	</div>

	<!-- Main Chat Area -->
	<div class="flex-1 flex flex-col">
		{#if showNewChat}
			<!-- New Chat Modal -->
			<div class="flex-1 p-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">Новый чат</h3>

				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Поиск сотрудников..."
					class="w-full px-4 py-2 border border-gray-200 rounded-lg mb-4 focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
				/>

				{#if selectedParticipants.length > 1}
					<input
						type="text"
						bind:value={groupName}
						placeholder="Название группы (необязательно)"
						class="w-full px-4 py-2 border border-gray-200 rounded-lg mb-4 focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
					/>
				{/if}

				<div class="space-y-2 max-h-96 overflow-y-auto">
					{#each filteredEmployees as emp}
						<button
							onclick={() => toggleParticipant(emp.id)}
							class="w-full p-3 flex items-center gap-3 rounded-lg border transition-colors
								{selectedParticipants.includes(emp.id) ? 'border-ekf-red bg-red-50' : 'border-gray-200 hover:bg-gray-50'}"
						>
							{#if emp.photo_base64}
								<img
									src="data:image/jpeg;base64,{emp.photo_base64}"
									alt=""
									class="w-10 h-10 rounded-full object-cover"
								/>
							{:else}
								<div class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center text-gray-600 font-medium">
									{emp.name.charAt(0)}
								</div>
							{/if}
							<div class="flex-1 text-left">
								<div class="font-medium text-gray-900">{emp.name}</div>
								<div class="text-sm text-gray-500">{emp.position}</div>
							</div>
							{#if selectedParticipants.includes(emp.id)}
								<svg class="w-5 h-5 text-ekf-red" fill="currentColor" viewBox="0 0 20 20">
									<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
								</svg>
							{/if}
						</button>
					{/each}
				</div>

				<div class="mt-6 flex gap-3">
					<button
						onclick={() => {
							showNewChat = false;
							selectedParticipants = [];
							groupName = '';
						}}
						class="px-4 py-2 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
					>
						Отмена
					</button>
					<button
						onclick={createConversation}
						disabled={selectedParticipants.length === 0}
						class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Создать чат
					</button>
				</div>
			</div>
		{:else if currentConversation}
			<!-- Chat Header -->
			<div class="p-4 border-b border-gray-200 flex items-center gap-3">
				{#if getConversationAvatar(currentConversation)}
					<img
						src="data:image/jpeg;base64,{getConversationAvatar(currentConversation)}"
						alt=""
						class="w-10 h-10 rounded-full object-cover"
					/>
				{:else}
					<div class="w-10 h-10 rounded-full bg-ekf-red text-white flex items-center justify-center font-medium">
						{getConversationName(currentConversation).charAt(0)}
					</div>
				{/if}
				<div>
					<div class="font-medium text-gray-900">{getConversationName(currentConversation)}</div>
					{#if typingUsers[currentConversation.id]}
						<div class="text-sm text-gray-500">печатает...</div>
					{:else if currentConversation.participants}
						<div class="text-sm text-gray-500">
							{currentConversation.participants.length} участников
						</div>
					{/if}
				</div>
			</div>

			<!-- Messages -->
			<div bind:this={messagesContainer} class="flex-1 overflow-y-auto p-4 space-y-4">
				{#if loadingMessages}
					<div class="flex items-center justify-center h-32">
						<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-ekf-red"></div>
					</div>
				{:else if messages.length === 0}
					<div class="text-center py-8 text-gray-500">
						Нет сообщений. Начните диалог!
					</div>
				{:else}
					{#each messages as msg}
						{@const isOwn = msg.sender_id === $user?.id}
						<div class="flex {isOwn ? 'justify-end' : 'justify-start'}">
							<div class="flex items-end gap-2 max-w-[70%] {isOwn ? 'flex-row-reverse' : ''}">
								{#if !isOwn}
									{#if msg.sender?.photo_base64}
										<img
											src="data:image/jpeg;base64,{msg.sender.photo_base64}"
											alt=""
											class="w-8 h-8 rounded-full object-cover flex-shrink-0"
										/>
									{:else}
										<div class="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center text-gray-600 text-sm font-medium flex-shrink-0">
											{msg.sender?.name?.charAt(0) || '?'}
										</div>
									{/if}
								{/if}
								<div class="{isOwn ? 'bg-ekf-red text-white' : 'bg-gray-100 text-gray-900'} rounded-2xl px-4 py-2">
									{#if !isOwn && currentConversation.type === 'group'}
										<div class="text-xs font-medium mb-1 {isOwn ? 'text-red-200' : 'text-gray-500'}">
											{msg.sender?.name}
										</div>
									{/if}
									<div class="break-words">{msg.content}</div>
									<div class="text-xs mt-1 {isOwn ? 'text-red-200' : 'text-gray-400'}">
										{formatTime(msg.created_at)}
									</div>
								</div>
							</div>
						</div>
					{/each}
				{/if}
			</div>

			<!-- Message Input -->
			<div class="p-4 border-t border-gray-200">
				<form onsubmit={(e) => { e.preventDefault(); sendMessage(); }} class="flex gap-3">
					<input
						type="text"
						bind:value={newMessage}
						oninput={sendTyping}
						placeholder="Введите сообщение..."
						class="flex-1 px-4 py-2 border border-gray-200 rounded-full focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
					/>
					<button
						type="submit"
						disabled={!newMessage.trim()}
						class="px-4 py-2 bg-ekf-red text-white rounded-full hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
						</svg>
					</button>
				</form>
			</div>
		{:else}
			<!-- Empty State -->
			<div class="flex-1 flex items-center justify-center text-gray-500">
				<div class="text-center">
					<svg class="w-16 h-16 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
					</svg>
					<p class="text-lg">Выберите чат или начните новый</p>
				</div>
			</div>
		{/if}
	</div>
</div>
