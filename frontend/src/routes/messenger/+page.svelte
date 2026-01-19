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
	let typingUsers: Record<string, { userId: string; name: string }> = $state({});
	let messagesContainer: HTMLDivElement;
	let searchQuery = $state('');
	let chatSearchQuery = $state('');

	// Reply feature
	let replyingTo: Message | null = $state(null);

	// Context menu
	let contextMenuMessage: Message | null = $state(null);
	let contextMenuPosition = $state({ x: 0, y: 0 });

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
					typingUsers = {
						...typingUsers,
						[data.conversation_id]: { userId: data.data.user_id, name: data.data.name || 'Кто-то' }
					};
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
		replyingTo = null;
		try {
			const result = await messenger.getConversation(conv.id, $user!.id);
			messages = result.messages || [];
			if (result.participants) {
				currentConversation = { ...conv, participants: result.participants };
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
		const replyToId = replyingTo?.id;
		newMessage = '';
		replyingTo = null;

		try {
			const msg = await messenger.sendMessage({
				conversation_id: currentConversation.id,
				sender_id: $user.id,
				content,
				reply_to_id: replyToId
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

	function getOtherParticipant(conv: Conversation): Employee | undefined {
		if (conv.type === 'direct' && conv.participants) {
			return conv.participants.find((p) => p.id !== $user?.id);
		}
		return undefined;
	}

	function formatTime(dateStr?: string): string {
		if (!dateStr) return '';
		const date = new Date(dateStr);
		return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
	}

	function formatDate(dateStr?: string): string {
		if (!dateStr) return '';
		const date = new Date(dateStr);
		const now = new Date();
		const isToday = date.toDateString() === now.toDateString();
		const yesterday = new Date(now);
		yesterday.setDate(yesterday.getDate() - 1);
		const isYesterday = date.toDateString() === yesterday.toDateString();

		if (isToday) return 'Сегодня';
		if (isYesterday) return 'Вчера';
		return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' });
	}

	function formatLastMessageTime(dateStr?: string): string {
		if (!dateStr) return '';
		const date = new Date(dateStr);
		const now = new Date();
		const isToday = date.toDateString() === now.toDateString();

		if (isToday) {
			return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
		}

		const yesterday = new Date(now);
		yesterday.setDate(yesterday.getDate() - 1);
		if (date.toDateString() === yesterday.toDateString()) {
			return 'Вчера';
		}

		const weekAgo = new Date(now);
		weekAgo.setDate(weekAgo.getDate() - 7);
		if (date > weekAgo) {
			return date.toLocaleDateString('ru-RU', { weekday: 'short' });
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

	function handleMessageContextMenu(event: MouseEvent, msg: Message) {
		event.preventDefault();
		contextMenuMessage = msg;
		contextMenuPosition = { x: event.clientX, y: event.clientY };
	}

	function closeContextMenu() {
		contextMenuMessage = null;
	}

	function replyToMessage(msg: Message) {
		replyingTo = msg;
		contextMenuMessage = null;
	}

	function cancelReply() {
		replyingTo = null;
	}

	function copyMessageText(msg: Message) {
		navigator.clipboard.writeText(msg.content);
		contextMenuMessage = null;
	}

	// Group messages by date
	function getGroupedMessages(): { date: string; messages: Message[] }[] {
		const groups: { date: string; messages: Message[] }[] = [];
		let currentDate = '';

		for (const msg of messages) {
			const msgDate = msg.created_at?.split('T')[0] || '';
			if (msgDate !== currentDate) {
				currentDate = msgDate;
				groups.push({ date: msgDate, messages: [msg] });
			} else {
				groups[groups.length - 1].messages.push(msg);
			}
		}

		return groups;
	}

	function getInitials(name: string): string {
		return name
			.split(' ')
			.map(n => n.charAt(0))
			.slice(0, 2)
			.join('')
			.toUpperCase();
	}

	function getAvatarColor(name: string): string {
		const colors = [
			'bg-red-500', 'bg-blue-500', 'bg-green-500', 'bg-yellow-500',
			'bg-purple-500', 'bg-pink-500', 'bg-indigo-500', 'bg-teal-500'
		];
		let hash = 0;
		for (let i = 0; i < name.length; i++) {
			hash = name.charCodeAt(i) + ((hash << 5) - hash);
		}
		return colors[Math.abs(hash) % colors.length];
	}

	let filteredEmployees = $derived(
		employees.filter((e) =>
			e.id !== $user?.id &&
			e.name.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	let filteredConversations = $derived(
		chatSearchQuery
			? conversations.filter(c =>
					getConversationName(c).toLowerCase().includes(chatSearchQuery.toLowerCase())
				)
			: conversations
	);
</script>

<svelte:head>
	<title>Мессенджер - EKF Team Hub</title>
</svelte:head>

<!-- Click outside to close context menu -->
<svelte:window onclick={closeContextMenu} />

<div class="h-[calc(100vh-100px)] flex bg-white rounded-xl shadow-sm overflow-hidden">
	<!-- Sidebar - Conversations List (Telegram-style) -->
	<div class="w-96 border-r border-gray-100 flex flex-col bg-white">
		<!-- Header -->
		<div class="p-3 flex items-center gap-2">
			<div class="flex-1 relative">
				<input
					type="text"
					bind:value={chatSearchQuery}
					placeholder="Поиск"
					class="w-full pl-10 pr-4 py-2 bg-gray-100 rounded-full text-sm focus:outline-none focus:bg-gray-200 transition-colors"
				/>
				<svg class="w-5 h-5 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
			</div>
			<button
				onclick={() => (showNewChat = !showNewChat)}
				class="p-2 text-gray-500 hover:text-ekf-red hover:bg-gray-100 rounded-full transition-colors"
				title="Новый чат"
			>
				<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
				</svg>
			</button>
		</div>

		<!-- Conversations List -->
		<div class="flex-1 overflow-y-auto">
			{#if loading}
				<div class="flex items-center justify-center h-32">
					<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-ekf-red"></div>
				</div>
			{:else if filteredConversations.length === 0}
				<div class="text-center py-12 px-4">
					<div class="w-20 h-20 mx-auto mb-4 bg-gray-100 rounded-full flex items-center justify-center">
						<svg class="w-10 h-10 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
						</svg>
					</div>
					<p class="text-gray-500 mb-2">Нет чатов</p>
					<button
						onclick={() => (showNewChat = true)}
						class="text-ekf-red hover:underline text-sm"
					>
						Начать новый чат
					</button>
				</div>
			{:else}
				{#each filteredConversations as conv}
					{@const otherUser = getOtherParticipant(conv)}
					<button
						onclick={() => selectConversation(conv)}
						class="w-full px-3 py-2 flex items-center gap-3 hover:bg-gray-50 transition-colors
							{currentConversation?.id === conv.id ? 'bg-blue-50' : ''}"
					>
						<!-- Avatar -->
						<div class="relative flex-shrink-0">
							{#if getConversationAvatar(conv)}
								<img
									src="data:image/jpeg;base64,{getConversationAvatar(conv)}"
									alt=""
									class="w-14 h-14 rounded-full object-cover"
								/>
							{:else}
								<div class="w-14 h-14 rounded-full {getAvatarColor(getConversationName(conv))} text-white flex items-center justify-center text-lg font-medium">
									{#if conv.type === 'group'}
										<svg class="w-7 h-7" fill="currentColor" viewBox="0 0 24 24">
											<path d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
										</svg>
									{:else}
										{getInitials(getConversationName(conv))}
									{/if}
								</div>
							{/if}
							<!-- Online indicator would go here -->
						</div>

						<!-- Chat info -->
						<div class="flex-1 min-w-0 text-left">
							<div class="flex items-center justify-between">
								<span class="font-medium text-gray-900 truncate">{getConversationName(conv)}</span>
								{#if conv.last_message?.created_at}
									<span class="text-xs text-gray-400 flex-shrink-0 ml-2">
										{formatLastMessageTime(conv.last_message.created_at)}
									</span>
								{/if}
							</div>
							<div class="flex items-center gap-1 mt-0.5">
								{#if conv.last_message}
									{#if conv.last_message.sender_id === $user?.id}
										<svg class="w-4 h-4 text-blue-500 flex-shrink-0" fill="currentColor" viewBox="0 0 24 24">
											<path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z" />
										</svg>
									{/if}
									<span class="text-sm text-gray-500 truncate">
										{#if conv.type === 'group' && conv.last_message.sender_id !== $user?.id}
											<span class="text-gray-600">{conv.last_message.sender?.name?.split(' ')[0] || ''}:</span>
										{/if}
										{conv.last_message.content}
									</span>
								{:else}
									<span class="text-sm text-gray-400 italic">Нет сообщений</span>
								{/if}
							</div>
						</div>
					</button>
				{/each}
			{/if}
		</div>
	</div>

	<!-- Main Chat Area -->
	<div class="flex-1 flex flex-col bg-[#e5ddd5]" style="background-image: url(&quot;data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23000000' fill-opacity='0.03'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E&quot;);">
		{#if showNewChat}
			<!-- New Chat Modal -->
			<div class="flex-1 bg-white p-6 overflow-auto">
				<div class="max-w-lg mx-auto">
					<div class="flex items-center gap-3 mb-6">
						<button
							onclick={() => {
								showNewChat = false;
								selectedParticipants = [];
								groupName = '';
							}}
							class="p-2 hover:bg-gray-100 rounded-full transition-colors"
						>
							<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
							</svg>
						</button>
						<h3 class="text-xl font-semibold text-gray-900">Новый чат</h3>
					</div>

					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Кому: Введите имя"
						class="w-full px-4 py-3 border-b border-gray-200 focus:outline-none focus:border-ekf-red text-lg"
					/>

					{#if selectedParticipants.length > 0}
						<div class="flex flex-wrap gap-2 py-3 border-b border-gray-200">
							{#each selectedParticipants as participantId}
								{@const emp = employees.find(e => e.id === participantId)}
								{#if emp}
									<span class="inline-flex items-center gap-1 px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm">
										{emp.name}
										<button onclick={() => toggleParticipant(participantId)} class="hover:text-blue-600">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
											</svg>
										</button>
									</span>
								{/if}
							{/each}
						</div>
					{/if}

					{#if selectedParticipants.length > 1}
						<input
							type="text"
							bind:value={groupName}
							placeholder="Название группы"
							class="w-full px-4 py-3 border-b border-gray-200 focus:outline-none focus:border-ekf-red"
						/>
					{/if}

					<div class="py-2">
						{#each filteredEmployees as emp}
							<button
								onclick={() => toggleParticipant(emp.id)}
								class="w-full p-3 flex items-center gap-3 hover:bg-gray-50 rounded-lg transition-colors"
							>
								{#if emp.photo_base64}
									<img
										src="data:image/jpeg;base64,{emp.photo_base64}"
										alt=""
										class="w-12 h-12 rounded-full object-cover"
									/>
								{:else}
									<div class="w-12 h-12 rounded-full {getAvatarColor(emp.name)} text-white flex items-center justify-center font-medium">
										{getInitials(emp.name)}
									</div>
								{/if}
								<div class="flex-1 text-left">
									<div class="font-medium text-gray-900">{emp.name}</div>
									<div class="text-sm text-gray-500">{emp.position || emp.department || ''}</div>
								</div>
								{#if selectedParticipants.includes(emp.id)}
									<div class="w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center">
										<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
									</div>
								{/if}
							</button>
						{/each}
					</div>

					{#if selectedParticipants.length > 0}
						<div class="fixed bottom-8 right-8">
							<button
								onclick={createConversation}
								class="w-14 h-14 bg-ekf-red text-white rounded-full shadow-lg hover:bg-red-700 transition-colors flex items-center justify-center"
							>
								<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
								</svg>
							</button>
						</div>
					{/if}
				</div>
			</div>
		{:else if currentConversation}
			<!-- Chat Header -->
			<div class="bg-white px-4 py-3 flex items-center gap-3 shadow-sm">
				<button
					onclick={() => currentConversation = null}
					class="p-2 hover:bg-gray-100 rounded-full transition-colors md:hidden"
				>
					<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
				</button>
				{#if getConversationAvatar(currentConversation)}
					<img
						src="data:image/jpeg;base64,{getConversationAvatar(currentConversation)}"
						alt=""
						class="w-10 h-10 rounded-full object-cover"
					/>
				{:else}
					<div class="w-10 h-10 rounded-full {getAvatarColor(getConversationName(currentConversation))} text-white flex items-center justify-center font-medium">
						{#if currentConversation.type === 'group'}
							<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
								<path d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
							</svg>
						{:else}
							{getInitials(getConversationName(currentConversation))}
						{/if}
					</div>
				{/if}
				<div class="flex-1">
					<div class="font-medium text-gray-900">{getConversationName(currentConversation)}</div>
					{#if typingUsers[currentConversation.id]}
						<div class="text-sm text-blue-500">печатает...</div>
					{:else if currentConversation.participants}
						<div class="text-sm text-gray-500">
							{currentConversation.type === 'group'
								? `${currentConversation.participants.length} участников`
								: 'был(а) недавно'}
						</div>
					{/if}
				</div>
				<button class="p-2 hover:bg-gray-100 rounded-full transition-colors" title="Поиск">
					<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
				</button>
				<button class="p-2 hover:bg-gray-100 rounded-full transition-colors" title="Ещё">
					<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
					</svg>
				</button>
			</div>

			<!-- Messages -->
			<div bind:this={messagesContainer} class="flex-1 overflow-y-auto px-4 py-2">
				{#if loadingMessages}
					<div class="flex items-center justify-center h-32">
						<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-ekf-red"></div>
					</div>
				{:else if messages.length === 0}
					<div class="text-center py-8">
						<div class="inline-block px-4 py-2 bg-white/80 rounded-full text-gray-500 text-sm">
							Нет сообщений. Начните диалог!
						</div>
					</div>
				{:else}
					{#each getGroupedMessages() as group}
						<!-- Date separator -->
						<div class="flex justify-center my-4">
							<span class="px-3 py-1 bg-white/80 rounded-full text-xs text-gray-500 shadow-sm">
								{formatDate(group.date)}
							</span>
						</div>

						{#each group.messages as msg, i}
							{@const isOwn = msg.sender_id === $user?.id}
							{@const showAvatar = !isOwn && (i === 0 || group.messages[i - 1]?.sender_id !== msg.sender_id)}
							{@const isLastInGroup = i === group.messages.length - 1 || group.messages[i + 1]?.sender_id !== msg.sender_id}

							<div
								class="flex mb-1 {isOwn ? 'justify-end' : 'justify-start'}"
								oncontextmenu={(e) => handleMessageContextMenu(e, msg)}
							>
								<div class="flex items-end gap-2 max-w-[75%] {isOwn ? 'flex-row-reverse' : ''}">
									<!-- Avatar -->
									{#if !isOwn && currentConversation.type === 'group'}
										<div class="w-8 flex-shrink-0">
											{#if showAvatar}
												{#if msg.sender?.photo_base64}
													<img
														src="data:image/jpeg;base64,{msg.sender.photo_base64}"
														alt=""
														class="w-8 h-8 rounded-full object-cover"
													/>
												{:else}
													<div class="w-8 h-8 rounded-full {getAvatarColor(msg.sender?.name || '')} text-white flex items-center justify-center text-xs font-medium">
														{getInitials(msg.sender?.name || '?')}
													</div>
												{/if}
											{/if}
										</div>
									{/if}

									<!-- Message bubble -->
									<div
										class="px-3 py-2 shadow-sm {isOwn
											? 'bg-[#dcf8c6] rounded-2xl rounded-br-sm'
											: 'bg-white rounded-2xl rounded-bl-sm'}"
									>
										<!-- Reply preview -->
										{#if msg.reply_to}
											<div class="border-l-2 border-blue-400 pl-2 mb-1 text-sm">
												<div class="font-medium text-blue-600">{msg.reply_to.sender?.name || 'Сообщение'}</div>
												<div class="text-gray-500 truncate">{msg.reply_to.content}</div>
											</div>
										{/if}

										<!-- Sender name for groups -->
										{#if !isOwn && currentConversation.type === 'group' && showAvatar}
											<div class="text-sm font-medium text-blue-600 mb-0.5">
												{msg.sender?.name}
											</div>
										{/if}

										<!-- Content -->
										<div class="break-words text-gray-900 leading-relaxed">{msg.content}</div>

										<!-- Time and status -->
										<div class="flex items-center justify-end gap-1 -mb-1 mt-1">
											<span class="text-xs text-gray-500">{formatTime(msg.created_at)}</span>
											{#if isOwn}
												<svg class="w-4 h-4 text-blue-500" fill="currentColor" viewBox="0 0 24 24">
													<path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z" />
												</svg>
											{/if}
										</div>
									</div>
								</div>
							</div>
						{/each}
					{/each}
				{/if}
			</div>

			<!-- Reply Preview -->
			{#if replyingTo}
				<div class="bg-white border-t border-gray-200 px-4 py-2 flex items-center gap-3">
					<div class="w-1 h-10 bg-blue-500 rounded"></div>
					<div class="flex-1 min-w-0">
						<div class="text-sm font-medium text-blue-600">{replyingTo.sender?.name || 'Сообщение'}</div>
						<div class="text-sm text-gray-500 truncate">{replyingTo.content}</div>
					</div>
					<button onclick={cancelReply} class="p-1 hover:bg-gray-100 rounded-full">
						<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			{/if}

			<!-- Message Input -->
			<div class="bg-white px-4 py-3">
				<form onsubmit={(e) => { e.preventDefault(); sendMessage(); }} class="flex items-end gap-2">
					<button type="button" class="p-2 hover:bg-gray-100 rounded-full transition-colors">
						<svg class="w-6 h-6 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</button>
					<button type="button" class="p-2 hover:bg-gray-100 rounded-full transition-colors">
						<svg class="w-6 h-6 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
						</svg>
					</button>
					<div class="flex-1">
						<input
							type="text"
							bind:value={newMessage}
							oninput={sendTyping}
							placeholder="Сообщение"
							class="w-full px-4 py-2 bg-gray-100 rounded-full focus:outline-none focus:bg-gray-200 transition-colors"
						/>
					</div>
					{#if newMessage.trim()}
						<button
							type="submit"
							class="p-2 bg-ekf-red text-white rounded-full hover:bg-red-700 transition-colors"
						>
							<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
							</svg>
						</button>
					{:else}
						<button type="button" class="p-2 hover:bg-gray-100 rounded-full transition-colors">
							<svg class="w-6 h-6 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
							</svg>
						</button>
					{/if}
				</form>
			</div>
		{:else}
			<!-- Empty State -->
			<div class="flex-1 flex items-center justify-center">
				<div class="text-center p-8 bg-white/80 rounded-2xl">
					<div class="w-24 h-24 mx-auto mb-6 bg-blue-100 rounded-full flex items-center justify-center">
						<svg class="w-12 h-12 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
						</svg>
					</div>
					<h3 class="text-xl font-medium text-gray-900 mb-2">EKF Team Hub</h3>
					<p class="text-gray-500 mb-4">Выберите чат из списка слева<br />или начните новый</p>
					<button
						onclick={() => showNewChat = true}
						class="px-6 py-2 bg-ekf-red text-white rounded-full hover:bg-red-700 transition-colors"
					>
						Новый чат
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- Context Menu -->
{#if contextMenuMessage}
	<div
		class="fixed bg-white rounded-xl shadow-xl py-2 z-50 min-w-48"
		style="left: {contextMenuPosition.x}px; top: {contextMenuPosition.y}px;"
		onclick={(e) => e.stopPropagation()}
	>
		<button
			onclick={() => replyToMessage(contextMenuMessage!)}
			class="w-full px-4 py-2 text-left text-gray-700 hover:bg-gray-100 flex items-center gap-3"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
			</svg>
			Ответить
		</button>
		<button
			onclick={() => copyMessageText(contextMenuMessage!)}
			class="w-full px-4 py-2 text-left text-gray-700 hover:bg-gray-100 flex items-center gap-3"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
			</svg>
			Копировать
		</button>
		<button
			class="w-full px-4 py-2 text-left text-gray-700 hover:bg-gray-100 flex items-center gap-3"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
			</svg>
			Переслать
		</button>
	</div>
{/if}
