<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { messenger, employees as employeesApi } from '$lib/api/client';
	import type { Conversation, Message, Employee } from '$lib/api/client';
	import { user, subordinates as userSubordinates } from '$lib/stores/auth';

	// State
	let conversations: Conversation[] = $state([]);
	let currentConversation: Conversation | null = $state(null);
	let messages: Message[] = $state([]);
	let employees: Employee[] = $state([]);
	let newMessage = $state('');
	let loading = $state(true);
	let loadingMessages = $state(false);
	let ws: WebSocket | null = null;
	let typingUsers: Record<string, { userId: string; name: string }> = $state({});
	let messagesContainer: HTMLDivElement;
	let searchQuery = $state('');

	// Tabs: chats, contacts, channels
	let activeTab: 'chats' | 'contacts' | 'channels' = $state('chats');

	// New chat modal
	let showNewChat = $state(false);
	let newChatType: 'direct' | 'group' | 'channel' = $state('direct');
	let selectedParticipants: string[] = $state([]);
	let groupName = $state('');
	let channelDescription = $state('');

	// Contacts - expanded departments
	let expandedDepartments: Set<string> = $state(new Set());

	// Reply feature
	let replyingTo: Message | null = $state(null);

	// Context menu
	let contextMenuMessage: Message | null = $state(null);
	let contextMenuPosition = $state({ x: 0, y: 0 });

	// Build department hierarchy from employees
	let departmentHierarchy = $derived.by(() => {
		const depts: Record<string, { name: string; employees: Employee[]; subDepts: string[] }> = {};

		employees.forEach(emp => {
			const dept = emp.department || 'Без отдела';
			if (!depts[dept]) {
				depts[dept] = { name: dept, employees: [], subDepts: [] };
			}
			depts[dept].employees.push(emp);
		});

		// Sort employees within departments
		Object.values(depts).forEach(d => {
			d.employees.sort((a, b) => a.name.localeCompare(b.name, 'ru'));
		});

		return depts;
	});

	// Filter conversations by type
	let chats = $derived(
		conversations.filter(c => c.type === 'direct' || c.type === 'group')
	);

	let channels = $derived(
		conversations.filter(c => c.type === 'channel')
	);

	let filteredChats = $derived(
		searchQuery
			? chats.filter(c => getConversationName(c).toLowerCase().includes(searchQuery.toLowerCase()))
			: chats
	);

	let filteredChannels = $derived(
		searchQuery
			? channels.filter(c => getConversationName(c).toLowerCase().includes(searchQuery.toLowerCase()))
			: channels
	);

	let filteredEmployees = $derived(
		employees.filter((e) =>
			e.id !== $user?.id &&
			e.name.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

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

		ws.onopen = () => console.log('WebSocket connected');

		ws.onmessage = (event) => {
			const data = JSON.parse(event.data);
			handleWSMessage(data);
		};

		ws.onclose = () => {
			console.log('WebSocket disconnected');
			setTimeout(connectWebSocket, 3000);
		};

		ws.onerror = (err) => console.error('WebSocket error:', err);
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

	async function startDirectChat(employee: Employee) {
		// Check if conversation already exists
		const existing = conversations.find(c =>
			c.type === 'direct' &&
			c.participants?.some(p => p.id === employee.id)
		);

		if (existing) {
			selectConversation(existing);
			activeTab = 'chats';
			return;
		}

		// Create new direct conversation
		try {
			const conv = await messenger.createConversation({
				type: 'direct',
				participants: [employee.id, $user!.id]
			});
			// Add participant info directly since we know who it is
			const convWithParticipants = {
				...conv,
				participants: [employee, $user as Employee]
			};
			conversations = [convWithParticipants, ...conversations];
			selectConversation(convWithParticipants);
			activeTab = 'chats';
		} catch (e) {
			console.error('Failed to create conversation:', e);
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

		try {
			const conv = await messenger.createConversation({
				type: newChatType,
				name: newChatType !== 'direct' ? groupName : undefined,
				participants
			});
			// Add participant info from selected employees
			const participantEmployees = employees.filter(e =>
				selectedParticipants.includes(e.id) || e.id === $user?.id
			);
			const convWithParticipants = {
				...conv,
				participants: participantEmployees
			};
			conversations = [convWithParticipants, ...conversations];
			showNewChat = false;
			selectedParticipants = [];
			groupName = '';
			channelDescription = '';
			selectConversation(convWithParticipants);
			activeTab = newChatType === 'channel' ? 'channels' : 'chats';
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

		return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
	}

	function toggleParticipant(id: string) {
		if (selectedParticipants.includes(id)) {
			selectedParticipants = selectedParticipants.filter((p) => p !== id);
		} else {
			selectedParticipants = [...selectedParticipants, id];
		}
	}

	function toggleDepartment(dept: string) {
		const newSet = new Set(expandedDepartments);
		if (newSet.has(dept)) {
			newSet.delete(dept);
		} else {
			newSet.add(dept);
		}
		expandedDepartments = newSet;
	}

	function handleMessageContextMenu(event: MouseEvent, msg: Message) {
		event.preventDefault();
		contextMenuMessage = msg;

		// Calculate position to keep menu on screen
		const menuWidth = 160; // min-w-40 = 160px
		const menuHeight = 120; // approximate height for 3 buttons
		let x = event.clientX;
		let y = event.clientY;

		// Adjust if would go off right edge
		if (x + menuWidth > window.innerWidth) {
			x = window.innerWidth - menuWidth - 10;
		}
		// Adjust if would go off bottom edge
		if (y + menuHeight > window.innerHeight) {
			y = window.innerHeight - menuHeight - 10;
		}

		contextMenuPosition = { x, y };
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

	function openNewChat(type: 'direct' | 'group' | 'channel') {
		showNewChat = true;
		newChatType = type;
		selectedParticipants = [];
		groupName = '';
		channelDescription = '';
	}
</script>

<svelte:head>
	<title>Мессенджер - EKF Hub</title>
</svelte:head>

<svelte:window onclick={closeContextMenu} />

<div class="h-[calc(100vh-100px)] flex bg-white rounded-xl shadow-sm overflow-hidden">
	<!-- Sidebar -->
	<div class="w-80 border-r border-gray-200 flex flex-col bg-white">
		<!-- Header with search -->
		<div class="p-3 border-b border-gray-200">
			<div class="relative">
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Поиск"
					class="w-full pl-10 pr-4 py-2 bg-gray-100 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
				/>
				<svg class="w-5 h-5 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
			</div>
		</div>

		<!-- Tabs -->
		<div class="flex border-b border-gray-200">
			<button
				onclick={() => activeTab = 'chats'}
				class="flex-1 py-3 text-sm font-medium transition-colors relative
					{activeTab === 'chats' ? 'text-ekf-red' : 'text-gray-500 hover:text-gray-700'}"
			>
				Чаты
				{#if activeTab === 'chats'}
					<div class="absolute bottom-0 left-0 right-0 h-0.5 bg-ekf-red"></div>
				{/if}
			</button>
			<button
				onclick={() => activeTab = 'contacts'}
				class="flex-1 py-3 text-sm font-medium transition-colors relative
					{activeTab === 'contacts' ? 'text-ekf-red' : 'text-gray-500 hover:text-gray-700'}"
			>
				Контакты
				{#if activeTab === 'contacts'}
					<div class="absolute bottom-0 left-0 right-0 h-0.5 bg-ekf-red"></div>
				{/if}
			</button>
			<button
				onclick={() => activeTab = 'channels'}
				class="flex-1 py-3 text-sm font-medium transition-colors relative
					{activeTab === 'channels' ? 'text-ekf-red' : 'text-gray-500 hover:text-gray-700'}"
			>
				Каналы
				{#if activeTab === 'channels'}
					<div class="absolute bottom-0 left-0 right-0 h-0.5 bg-ekf-red"></div>
				{/if}
			</button>
		</div>

		<!-- Tab content -->
		<div class="flex-1 overflow-y-auto">
			{#if activeTab === 'chats'}
				<!-- Chats list -->
				{#if loading}
					<div class="flex items-center justify-center h-32">
						<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-ekf-red"></div>
					</div>
				{:else if filteredChats.length === 0}
					<div class="text-center py-12 px-4">
						<div class="w-16 h-16 mx-auto mb-4 bg-gray-100 rounded-full flex items-center justify-center">
							<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
							</svg>
						</div>
						<p class="text-gray-500 text-sm">Нет чатов</p>
					</div>
				{:else}
					{#each filteredChats as conv}
						<button
							onclick={() => selectConversation(conv)}
							class="w-full px-3 py-2.5 flex items-center gap-3 hover:bg-gray-50 transition-colors
								{currentConversation?.id === conv.id ? 'bg-ekf-red/5' : ''}"
						>
							{#if getConversationAvatar(conv)}
								<img src="data:image/jpeg;base64,{getConversationAvatar(conv)}" alt="" class="w-12 h-12 rounded-full object-cover" />
							{:else}
								<div class="w-12 h-12 rounded-full {getAvatarColor(getConversationName(conv))} text-white flex items-center justify-center font-medium">
									{#if conv.type === 'group'}
										<svg class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
											<path d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
										</svg>
									{:else}
										{getInitials(getConversationName(conv))}
									{/if}
								</div>
							{/if}
							<div class="flex-1 min-w-0 text-left">
								<div class="flex items-center justify-between">
									<span class="font-medium text-gray-900 truncate text-sm">{getConversationName(conv)}</span>
									{#if conv.last_message?.created_at}
										<span class="text-xs text-gray-400 ml-2">{formatLastMessageTime(conv.last_message.created_at)}</span>
									{/if}
								</div>
								{#if conv.last_message}
									<p class="text-sm text-gray-500 truncate mt-0.5">{conv.last_message.content}</p>
								{:else}
									<p class="text-sm text-gray-400 italic mt-0.5">Нет сообщений</p>
								{/if}
							</div>
						</button>
					{/each}
				{/if}

			{:else if activeTab === 'contacts'}
				<!-- Contacts with department hierarchy -->
				<div class="py-2">
					{#each Object.entries(departmentHierarchy).sort((a, b) => a[0].localeCompare(b[0], 'ru')) as [dept, data]}
						<div class="mb-1">
							<button
								onclick={() => toggleDepartment(dept)}
								class="w-full px-4 py-2 flex items-center gap-2 hover:bg-gray-50 text-left"
							>
								<svg
									class="w-4 h-4 text-gray-400 transition-transform {expandedDepartments.has(dept) ? 'rotate-90' : ''}"
									fill="none" stroke="currentColor" viewBox="0 0 24 24"
								>
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
								</svg>
								<div class="w-8 h-8 rounded-lg bg-ekf-red/10 flex items-center justify-center">
									<svg class="w-4 h-4 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
									</svg>
								</div>
								<div class="flex-1">
									<span class="text-sm font-medium text-gray-700">{dept}</span>
									<span class="text-xs text-gray-400 ml-1">({data.employees.length})</span>
								</div>
							</button>

							{#if expandedDepartments.has(dept)}
								<div class="ml-6 border-l border-gray-200">
									{#each data.employees as emp}
										{#if emp.id !== $user?.id}
											<button
												onclick={() => startDirectChat(emp)}
												class="w-full px-4 py-2 flex items-center gap-3 hover:bg-gray-50 transition-colors"
											>
												{#if emp.photo_base64}
													<img src="data:image/jpeg;base64,{emp.photo_base64}" alt="" class="w-10 h-10 rounded-full object-cover" />
												{:else}
													<div class="w-10 h-10 rounded-full {getAvatarColor(emp.name)} text-white flex items-center justify-center text-sm font-medium">
														{getInitials(emp.name)}
													</div>
												{/if}
												<div class="flex-1 text-left">
													<div class="text-sm font-medium text-gray-900">{emp.name}</div>
													<div class="text-xs text-gray-500">{emp.position || ''}</div>
												</div>
											</button>
										{/if}
									{/each}
								</div>
							{/if}
						</div>
					{/each}
				</div>

			{:else if activeTab === 'channels'}
				<!-- Channels list -->
				{#if filteredChannels.length === 0}
					<div class="text-center py-12 px-4">
						<div class="w-16 h-16 mx-auto mb-4 bg-gray-100 rounded-full flex items-center justify-center">
							<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
							</svg>
						</div>
						<p class="text-gray-500 text-sm mb-2">Нет каналов</p>
						<button
							onclick={() => openNewChat('channel')}
							class="text-ekf-red hover:underline text-sm"
						>
							Создать канал
						</button>
					</div>
				{:else}
					{#each filteredChannels as conv}
						<button
							onclick={() => selectConversation(conv)}
							class="w-full px-3 py-2.5 flex items-center gap-3 hover:bg-gray-50 transition-colors
								{currentConversation?.id === conv.id ? 'bg-ekf-red/5' : ''}"
						>
							<div class="w-12 h-12 rounded-full bg-blue-500 text-white flex items-center justify-center">
								<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
								</svg>
							</div>
							<div class="flex-1 min-w-0 text-left">
								<div class="font-medium text-gray-900 truncate text-sm">{conv.name || 'Канал'}</div>
								<p class="text-sm text-gray-500 truncate mt-0.5">
									{conv.participants?.length || 0} подписчиков
								</p>
							</div>
						</button>
					{/each}
				{/if}
			{/if}
		</div>

		<!-- New chat button -->
		<div class="p-3 border-t border-gray-200">
			<div class="flex gap-2">
				<button
					onclick={() => openNewChat('direct')}
					class="flex-1 py-2 px-3 bg-ekf-red text-white rounded-lg text-sm font-medium hover:bg-red-700 transition-colors flex items-center justify-center gap-2"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					Чат
				</button>
				<button
					onclick={() => openNewChat('group')}
					class="py-2 px-3 bg-gray-100 text-gray-700 rounded-lg text-sm font-medium hover:bg-gray-200 transition-colors"
					title="Создать группу"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
					</svg>
				</button>
				<button
					onclick={() => openNewChat('channel')}
					class="py-2 px-3 bg-gray-100 text-gray-700 rounded-lg text-sm font-medium hover:bg-gray-200 transition-colors"
					title="Создать канал"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
					</svg>
				</button>
			</div>
		</div>
	</div>

	<!-- Main Chat Area -->
	<div class="flex-1 flex flex-col bg-gray-50">
		{#if showNewChat}
			<!-- New Chat/Group/Channel Modal -->
			<div class="flex-1 bg-white p-6 overflow-auto">
				<div class="max-w-lg mx-auto">
					<div class="flex items-center gap-3 mb-6">
						<button
							onclick={() => showNewChat = false}
							class="p-2 hover:bg-gray-100 rounded-full transition-colors"
						>
							<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
							</svg>
						</button>
						<h3 class="text-xl font-semibold text-gray-900">
							{#if newChatType === 'direct'}Новый чат{:else if newChatType === 'group'}Новая группа{:else}Новый канал{/if}
						</h3>
					</div>

					{#if newChatType !== 'direct'}
						<div class="mb-4">
							<input
								type="text"
								bind:value={groupName}
								placeholder={newChatType === 'channel' ? 'Название канала' : 'Название группы'}
								class="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
							/>
						</div>
						{#if newChatType === 'channel'}
							<div class="mb-4">
								<textarea
									bind:value={channelDescription}
									placeholder="Описание канала (необязательно)"
									rows="2"
									class="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20 resize-none"
								></textarea>
							</div>
						{/if}
					{/if}

					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Поиск сотрудников"
						class="w-full px-4 py-3 border-b border-gray-200 focus:outline-none focus:border-ekf-red mb-2"
					/>

					{#if selectedParticipants.length > 0}
						<div class="flex flex-wrap gap-2 py-3">
							{#each selectedParticipants as participantId}
								{@const emp = employees.find(e => e.id === participantId)}
								{#if emp}
									<span class="inline-flex items-center gap-1 px-3 py-1 bg-ekf-red/10 text-ekf-red rounded-full text-sm">
										{emp.name}
										<button onclick={() => toggleParticipant(participantId)} class="hover:text-red-700">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
											</svg>
										</button>
									</span>
								{/if}
							{/each}
						</div>
					{/if}

					<div class="py-2 max-h-80 overflow-y-auto">
						{#each filteredEmployees as emp}
							<button
								onclick={() => toggleParticipant(emp.id)}
								class="w-full p-3 flex items-center gap-3 hover:bg-gray-50 rounded-lg transition-colors"
							>
								{#if emp.photo_base64}
									<img src="data:image/jpeg;base64,{emp.photo_base64}" alt="" class="w-10 h-10 rounded-full object-cover" />
								{:else}
									<div class="w-10 h-10 rounded-full {getAvatarColor(emp.name)} text-white flex items-center justify-center text-sm font-medium">
										{getInitials(emp.name)}
									</div>
								{/if}
								<div class="flex-1 text-left">
									<div class="font-medium text-gray-900 text-sm">{emp.name}</div>
									<div class="text-xs text-gray-500">{emp.position || emp.department || ''}</div>
								</div>
								{#if selectedParticipants.includes(emp.id)}
									<div class="w-6 h-6 bg-ekf-red rounded-full flex items-center justify-center">
										<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
										</svg>
									</div>
								{/if}
							</button>
						{/each}
					</div>

					{#if selectedParticipants.length > 0}
						<div class="mt-4">
							<button
								onclick={createConversation}
								disabled={newChatType !== 'direct' && !groupName.trim()}
								class="w-full py-3 bg-ekf-red text-white rounded-lg font-medium hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
							>
								{#if newChatType === 'direct'}Начать чат{:else if newChatType === 'group'}Создать группу{:else}Создать канал{/if}
							</button>
						</div>
					{/if}
				</div>
			</div>
		{:else if currentConversation}
			<!-- Chat Header -->
			<div class="bg-white px-4 py-3 flex items-center gap-3 border-b border-gray-200">
				<button onclick={() => currentConversation = null} class="p-2 hover:bg-gray-100 rounded-full md:hidden">
					<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
				</button>
				{#if getConversationAvatar(currentConversation)}
					<img src="data:image/jpeg;base64,{getConversationAvatar(currentConversation)}" alt="" class="w-10 h-10 rounded-full object-cover" />
				{:else}
					<div class="w-10 h-10 rounded-full {getAvatarColor(getConversationName(currentConversation))} text-white flex items-center justify-center font-medium">
						{#if currentConversation.type === 'group'}
							<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
								<path d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
							</svg>
						{:else if currentConversation.type === 'channel'}
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
							</svg>
						{:else}
							{getInitials(getConversationName(currentConversation))}
						{/if}
					</div>
				{/if}
				<div class="flex-1">
					<div class="font-medium text-gray-900">{getConversationName(currentConversation)}</div>
					{#if typingUsers[currentConversation.id]}
						<div class="text-sm text-ekf-red">печатает...</div>
					{:else if currentConversation.participants}
						<div class="text-sm text-gray-500">
							{currentConversation.type === 'group' || currentConversation.type === 'channel'
								? `${currentConversation.participants.length} участников`
								: 'был(а) недавно'}
						</div>
					{/if}
				</div>
				<button class="p-2 hover:bg-gray-100 rounded-full transition-colors">
					<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
				</button>
				<button class="p-2 hover:bg-gray-100 rounded-full transition-colors">
					<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
					</svg>
				</button>
			</div>

			<!-- Messages -->
			<div bind:this={messagesContainer} class="flex-1 overflow-y-auto px-4 py-2 bg-gray-50">
				{#if loadingMessages}
					<div class="flex items-center justify-center h-32">
						<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-ekf-red"></div>
					</div>
				{:else if messages.length === 0}
					<div class="text-center py-8">
						<div class="inline-block px-4 py-2 bg-white rounded-lg text-gray-500 text-sm shadow-sm">
							Нет сообщений. Начните диалог!
						</div>
					</div>
				{:else}
					{#each getGroupedMessages() as group}
						<div class="flex justify-center my-4">
							<span class="px-3 py-1 bg-white rounded-lg text-xs text-gray-500 shadow-sm">
								{formatDate(group.date)}
							</span>
						</div>

						{#each group.messages as msg, i}
							{@const isOwn = msg.sender_id === $user?.id}
							{@const showAvatar = !isOwn && (i === 0 || group.messages[i - 1]?.sender_id !== msg.sender_id)}

							<div
								class="flex mb-2 {isOwn ? 'justify-end' : 'justify-start'}"
								oncontextmenu={(e) => handleMessageContextMenu(e, msg)}
							>
								<div class="flex items-end gap-2 max-w-[70%] {isOwn ? 'flex-row-reverse' : ''}">
									{#if !isOwn && (currentConversation.type === 'group' || currentConversation.type === 'channel')}
										<div class="w-8 flex-shrink-0">
											{#if showAvatar}
												{#if msg.sender?.photo_base64}
													<img src="data:image/jpeg;base64,{msg.sender.photo_base64}" alt="" class="w-8 h-8 rounded-full object-cover" />
												{:else}
													<div class="w-8 h-8 rounded-full {getAvatarColor(msg.sender?.name || '')} text-white flex items-center justify-center text-xs font-medium">
														{getInitials(msg.sender?.name || '?')}
													</div>
												{/if}
											{/if}
										</div>
									{/if}

									<div class="px-3 py-2 rounded-2xl shadow-sm {isOwn ? 'bg-ekf-red text-white rounded-br-sm' : 'bg-white rounded-bl-sm'}">
										{#if msg.reply_to}
											<div class="border-l-2 {isOwn ? 'border-white/50' : 'border-ekf-red'} pl-2 mb-1 text-sm {isOwn ? 'text-white/80' : ''}">
												<div class="font-medium {isOwn ? 'text-white' : 'text-ekf-red'}">{msg.reply_to.sender?.name || 'Сообщение'}</div>
												<div class="{isOwn ? 'text-white/70' : 'text-gray-500'} truncate">{msg.reply_to.content}</div>
											</div>
										{/if}

										{#if !isOwn && (currentConversation.type === 'group' || currentConversation.type === 'channel') && showAvatar}
											<div class="text-sm font-medium text-ekf-red mb-0.5">{msg.sender?.name}</div>
										{/if}

										<div class="break-words leading-relaxed text-sm">{msg.content}</div>

										<div class="flex items-center justify-end gap-1 mt-1">
											<span class="text-xs {isOwn ? 'text-white/70' : 'text-gray-400'}">{formatTime(msg.created_at)}</span>
											{#if isOwn}
												<svg class="w-4 h-4 text-white/70" fill="currentColor" viewBox="0 0 24 24">
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
					<div class="w-1 h-10 bg-ekf-red rounded"></div>
					<div class="flex-1 min-w-0">
						<div class="text-sm font-medium text-ekf-red">{replyingTo.sender?.name || 'Сообщение'}</div>
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
			<div class="bg-white px-4 py-3 border-t border-gray-200">
				<form onsubmit={(e) => { e.preventDefault(); sendMessage(); }} class="flex items-center gap-2">
					<button type="button" class="p-2 hover:bg-gray-100 rounded-full transition-colors">
						<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
						</svg>
					</button>
					<div class="flex-1">
						<input
							type="text"
							bind:value={newMessage}
							oninput={sendTyping}
							placeholder="Сообщение"
							class="w-full px-4 py-2 bg-gray-100 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20 text-sm"
						/>
					</div>
					{#if newMessage.trim()}
						<button type="submit" class="p-2 bg-ekf-red text-white rounded-full hover:bg-red-700 transition-colors">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
							</svg>
						</button>
					{:else}
						<button type="button" class="p-2 hover:bg-gray-100 rounded-full transition-colors">
							<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
							</svg>
						</button>
					{/if}
				</form>
			</div>
		{:else}
			<!-- Empty State -->
			<div class="flex-1 flex items-center justify-center">
				<div class="text-center p-8">
					<div class="w-20 h-20 mx-auto mb-4 bg-ekf-red/10 rounded-full flex items-center justify-center">
						<svg class="w-10 h-10 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
						</svg>
					</div>
					<h3 class="text-lg font-medium text-gray-900 mb-2">EKF Hub Messenger</h3>
					<p class="text-gray-500 text-sm mb-4">Выберите чат или начните новый</p>
					<button onclick={() => openNewChat('direct')} class="px-6 py-2 bg-ekf-red text-white rounded-lg text-sm font-medium hover:bg-red-700 transition-colors">
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
		class="fixed bg-white rounded-lg shadow-lg py-1 z-50 min-w-40 border border-gray-200"
		style="left: {contextMenuPosition.x}px; top: {contextMenuPosition.y}px;"
		onclick={(e) => e.stopPropagation()}
	>
		<button onclick={() => replyToMessage(contextMenuMessage!)} class="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
			</svg>
			Ответить
		</button>
		<button onclick={() => copyMessageText(contextMenuMessage!)} class="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
			</svg>
			Копировать
		</button>
		<button class="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
			</svg>
			Переслать
		</button>
	</div>
{/if}
