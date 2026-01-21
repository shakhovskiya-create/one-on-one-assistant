<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { messenger, employees as employeesApi, speech } from '$lib/api/client';
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

	// Edit feature
	let editingMessage: Message | null = $state(null);

	// Context menu
	let contextMenuMessage: Message | null = $state(null);
	let contextMenuPosition = $state({ x: 0, y: 0 });

	// Emoji picker
	let showEmojiPicker = $state(false);
	const emojiCategories = {
		'Смайлики': ['😀', '😃', '😄', '😁', '😆', '😅', '🤣', '😂', '🙂', '😊', '😇', '🥰', '😍', '🤩', '😘', '😗', '😚', '😋', '😛', '😜', '🤪', '😝', '🤑', '🤗', '🤭', '🤫', '🤔', '🤐', '🤨', '😐', '😑', '😶', '😏', '😒', '🙄', '😬', '🤥', '😌', '😔', '😪', '🤤', '😴', '😷'],
		'Жесты': ['👍', '👎', '👌', '🤌', '✌️', '🤞', '🤟', '🤘', '👊', '✊', '🤛', '🤜', '👏', '🙌', '👐', '🤲', '🤝', '🙏', '💪', '🦾'],
		'Сердца': ['❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '🤍', '🤎', '💔', '❣️', '💕', '💞', '💓', '💗', '💖', '💘', '💝'],
		'Работа': ['💼', '📁', '📂', '📅', '📆', '📊', '📈', '📉', '📋', '📌', '📍', '✏️', '📝', '💻', '🖥️', '⌨️', '🖱️', '📱', '☎️', '📞', '✉️', '📧', '📨'],
		'Другое': ['✅', '❌', '❓', '❗', '💯', '🔥', '⭐', '🌟', '✨', '💡', '🎉', '🎊', '🏆', '🥇', '🎯', '🚀', '⏰', '⏳', '🔔', '🔕']
	};

	// Voice recording
	let isRecording = $state(false);
	let recordingTime = $state(0);
	let recordingInterval: ReturnType<typeof setInterval> | null = null;
	let mediaRecorder: MediaRecorder | null = null;
	let audioChunks: Blob[] = [];

	// Voice message storage (local - in production would be uploaded to server)
	let voiceMessages: Record<string, { url: string; duration: number; blob?: Blob }> = $state({});
	let playingVoiceId: string | null = $state(null);
	let audioElement: HTMLAudioElement | null = null;

	// Voice transcription state
	let voiceTranscriptions: Record<string, { text: string; loading: boolean; error?: string }> = $state({});

	// Call state
	let showCallModal = $state(false);
	let callType: 'audio' | 'video' = $state('audio');
	let callStatus: 'calling' | 'connected' | 'ended' = $state('calling');

	// Video circle recording
	let isRecordingVideo = $state(false);
	let videoRecordingTime = $state(0);
	let videoRecordingInterval: ReturnType<typeof setInterval> | null = null;
	let videoMediaRecorder: MediaRecorder | null = null;
	let videoChunks: Blob[] = [];
	let videoStream: MediaStream | null = null;
	let videoPreviewElement: HTMLVideoElement | null = $state(null);
	let showVideoRecorder = $state(false);

	// Video message storage
	let videoMessages: Record<string, { url: string; duration: number; blob?: Blob }> = $state({});
	let playingVideoId: string | null = $state(null);

	// Telegram bot configuration
	let showTelegramConfig = $state(false);
	let telegramBotToken = $state('');
	let telegramChatId = $state('');
	let telegramEnabled = $state(false);
	let telegramWebhookUrl = $state('');
	let telegramLoading = $state(false);

	async function openTelegramConfig() {
		if (!currentConversation || currentConversation.type !== 'channel') return;
		if (currentConversation.created_by !== $user?.id) return;

		telegramLoading = true;
		showTelegramConfig = true;

		try {
			const config = await messenger.getTelegramConfig(currentConversation.id);
			telegramEnabled = config.enabled;
			telegramChatId = config.chat_id?.toString() || '';
			telegramWebhookUrl = config.webhook_url;
		} catch (e) {
			console.error('Failed to load Telegram config:', e);
		} finally {
			telegramLoading = false;
		}
	}

	async function saveTelegramConfig() {
		if (!currentConversation) return;

		telegramLoading = true;
		try {
			const result = await messenger.configureTelegram(currentConversation.id, {
				bot_token: telegramBotToken || undefined,
				chat_id: telegramChatId ? parseInt(telegramChatId) : undefined,
				enabled: telegramEnabled
			});
			telegramWebhookUrl = result.webhook_url;
			alert(result.message);
			showTelegramConfig = false;

			// Update local conversation state
			if (currentConversation) {
				currentConversation = { ...currentConversation, telegram_enabled: telegramEnabled };
			}
		} catch (e) {
			console.error('Failed to save Telegram config:', e);
			alert('Ошибка сохранения настроек Telegram');
		} finally {
			telegramLoading = false;
		}
	}

	async function startRecording() {
		try {
			const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
			mediaRecorder = new MediaRecorder(stream);
			audioChunks = [];

			mediaRecorder.ondataavailable = (e) => {
				audioChunks.push(e.data);
			};

			mediaRecorder.onstop = () => {
				const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
				// В реальном приложении здесь бы отправлялось голосовое сообщение
				console.log('Голосовое сообщение записано:', audioBlob.size, 'байт');
				// Показать превью или отправить
				sendVoiceMessage(audioBlob);
				stream.getTracks().forEach(track => track.stop());
			};

			mediaRecorder.start();
			isRecording = true;
			recordingTime = 0;
			recordingInterval = setInterval(() => {
				recordingTime++;
			}, 1000);
		} catch (err) {
			console.error('Ошибка доступа к микрофону:', err);
			alert('Не удалось получить доступ к микрофону');
		}
	}

	function stopRecording() {
		if (mediaRecorder && isRecording) {
			mediaRecorder.stop();
			isRecording = false;
			if (recordingInterval) {
				clearInterval(recordingInterval);
				recordingInterval = null;
			}
		}
	}

	function cancelRecording() {
		if (mediaRecorder && isRecording) {
			mediaRecorder.stop();
			isRecording = false;
			audioChunks = [];
			if (recordingInterval) {
				clearInterval(recordingInterval);
				recordingInterval = null;
			}
			if (mediaRecorder.stream) {
				mediaRecorder.stream.getTracks().forEach(track => track.stop());
			}
		}
	}

	function sendVoiceMessage(audioBlob: Blob) {
		const duration = recordingTime;
		const voiceId = `voice_${Date.now()}`;

		// Create object URL for playback
		const audioUrl = URL.createObjectURL(audioBlob);
		voiceMessages[voiceId] = { url: audioUrl, duration, blob: audioBlob };

		// Send message with voice marker
		const msg = `[VOICE:${voiceId}:${duration}]`;
		if (currentConversation && $user) {
			newMessage = msg;
			sendMessage();
		}
	}

	async function transcribeVoice(voiceId: string) {
		const voice = voiceMessages[voiceId];
		if (!voice?.blob) {
			console.error('No audio blob available for transcription');
			return;
		}

		// Mark as loading
		voiceTranscriptions[voiceId] = { text: '', loading: true };
		voiceTranscriptions = { ...voiceTranscriptions };

		try {
			const result = await speech.transcribe(voice.blob);
			voiceTranscriptions[voiceId] = {
				text: result.transcript,
				loading: false
			};
		} catch (err) {
			console.error('Transcription failed:', err);
			voiceTranscriptions[voiceId] = {
				text: '',
				loading: false,
				error: err instanceof Error ? err.message : 'Transcription failed'
			};
		}
		voiceTranscriptions = { ...voiceTranscriptions };
	}

	function isVoiceMessage(content: string): boolean {
		return content.startsWith('[VOICE:');
	}

	function getVoiceInfo(content: string): { id: string; duration: number } | null {
		const match = content.match(/\[VOICE:([^:]+):(\d+)\]/);
		if (match) {
			return { id: match[1], duration: parseInt(match[2]) };
		}
		return null;
	}

	function playVoice(voiceId: string) {
		const voice = voiceMessages[voiceId];
		if (!voice) return;

		if (playingVoiceId === voiceId && audioElement) {
			// Toggle pause/play
			if (audioElement.paused) {
				audioElement.play();
			} else {
				audioElement.pause();
			}
			return;
		}

		// Stop current audio
		if (audioElement) {
			audioElement.pause();
			audioElement = null;
		}

		// Play new audio
		audioElement = new Audio(voice.url);
		playingVoiceId = voiceId;

		audioElement.onended = () => {
			playingVoiceId = null;
			audioElement = null;
		};

		audioElement.play();
	}

	function formatRecordingTime(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}

	// Video circle recording functions
	async function openVideoRecorder() {
		try {
			videoStream = await navigator.mediaDevices.getUserMedia({
				video: { facingMode: 'user', width: 480, height: 480 },
				audio: true
			});
			showVideoRecorder = true;

			// Wait for DOM to update, then set video source
			setTimeout(() => {
				if (videoPreviewElement && videoStream) {
					videoPreviewElement.srcObject = videoStream;
					videoPreviewElement.play();
				}
			}, 100);
		} catch (err) {
			console.error('Failed to access camera:', err);
			alert('Не удалось получить доступ к камере');
		}
	}

	function closeVideoRecorder() {
		if (videoStream) {
			videoStream.getTracks().forEach(track => track.stop());
			videoStream = null;
		}
		if (videoMediaRecorder && isRecordingVideo) {
			videoMediaRecorder.stop();
		}
		showVideoRecorder = false;
		isRecordingVideo = false;
		videoRecordingTime = 0;
		if (videoRecordingInterval) {
			clearInterval(videoRecordingInterval);
			videoRecordingInterval = null;
		}
	}

	function startVideoRecording() {
		if (!videoStream) return;

		videoMediaRecorder = new MediaRecorder(videoStream, { mimeType: 'video/webm' });
		videoChunks = [];

		videoMediaRecorder.ondataavailable = (e) => {
			if (e.data.size > 0) {
				videoChunks.push(e.data);
			}
		};

		videoMediaRecorder.onstop = () => {
			const videoBlob = new Blob(videoChunks, { type: 'video/webm' });
			sendVideoMessage(videoBlob);
			closeVideoRecorder();
		};

		videoMediaRecorder.start();
		isRecordingVideo = true;
		videoRecordingTime = 0;
		videoRecordingInterval = setInterval(() => {
			videoRecordingTime++;
			// Auto-stop after 60 seconds
			if (videoRecordingTime >= 60) {
				stopVideoRecording();
			}
		}, 1000);
	}

	function stopVideoRecording() {
		if (videoMediaRecorder && isRecordingVideo) {
			videoMediaRecorder.stop();
			isRecordingVideo = false;
			if (videoRecordingInterval) {
				clearInterval(videoRecordingInterval);
				videoRecordingInterval = null;
			}
		}
	}

	function sendVideoMessage(videoBlob: Blob) {
		const duration = videoRecordingTime;
		const videoId = `video_${Date.now()}`;

		const videoUrl = URL.createObjectURL(videoBlob);
		videoMessages[videoId] = { url: videoUrl, duration, blob: videoBlob };

		const msg = `[VIDEO:${videoId}:${duration}]`;
		if (currentConversation && $user) {
			newMessage = msg;
			sendMessage();
		}
	}

	function isVideoMessage(content: string): boolean {
		return content.startsWith('[VIDEO:');
	}

	function getVideoInfo(content: string): { id: string; duration: number } | null {
		const match = content.match(/\[VIDEO:([^:]+):(\d+)\]/);
		if (match) {
			return { id: match[1], duration: parseInt(match[2]) };
		}
		return null;
	}

	function playVideo(videoId: string) {
		if (playingVideoId === videoId) {
			playingVideoId = null;
		} else {
			playingVideoId = videoId;
		}
	}

	function startCall(type: 'audio' | 'video') {
		callType = type;
		callStatus = 'calling';
		showCallModal = true;
		// Симуляция подключения через 2 секунды
		setTimeout(() => {
			if (showCallModal) {
				callStatus = 'connected';
			}
		}, 2000);
	}

	function endCall() {
		callStatus = 'ended';
		setTimeout(() => {
			showCallModal = false;
		}, 500);
	}

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
			// Update last_message in sidebar
			updateConversationLastMessage(currentConversation.id, msg);
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
		// For channels, we can create without selecting participants
		// For chats, we need at least one participant
		if (newChatType !== 'channel' && selectedParticipants.length === 0) return;
		if (!$user?.id) return;

		const participants = [...selectedParticipants, $user.id];

		try {
			const conv = await messenger.createConversation({
				type: newChatType,
				name: newChatType !== 'direct' ? groupName : undefined,
				description: newChatType === 'channel' ? channelDescription : undefined,
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

	// Check if current user can post in conversation (for channels, only creator/admins can post)
	function canPostInConversation(conv: Conversation): boolean {
		if (conv.type !== 'channel') return true;
		// For channels, only the creator can post (in production, check is_admin flag)
		return conv.created_by === $user?.id;
	}

	// Get subscriber/participant count text
	function getParticipantCountText(conv: Conversation): string {
		const count = conv.participants?.length || 0;
		if (conv.type === 'channel') {
			// Russian pluralization for subscribers
			if (count === 1) return '1 подписчик';
			if (count >= 2 && count <= 4) return `${count} подписчика`;
			return `${count} подписчиков`;
		}
		// For groups
		if (count === 1) return '1 участник';
		if (count >= 2 && count <= 4) return `${count} участника`;
		return `${count} участников`;
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
		const menuWidth = 200; // actual width
		const menuHeight = 280; // reactions row + 4-5 buttons
		let x = event.clientX;
		let y = event.clientY;

		// Adjust if would go off right edge
		if (x + menuWidth > window.innerWidth) {
			x = window.innerWidth - menuWidth - 16;
		}
		// Adjust if would go off left edge
		if (x < 16) {
			x = 16;
		}
		// Adjust if would go off bottom edge
		if (y + menuHeight > window.innerHeight) {
			y = window.innerHeight - menuHeight - 16;
		}
		// Adjust if would go off top edge
		if (y < 16) {
			y = 16;
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

	function startEditMessage(msg: Message) {
		editingMessage = msg;
		newMessage = msg.content;
		contextMenuMessage = null;
	}

	function cancelEdit() {
		editingMessage = null;
		newMessage = '';
	}

	async function saveEdit() {
		if (!editingMessage || !newMessage.trim()) return;

		const updatedContent = newMessage.trim();
		// Update locally (in a real app this would be synced with backend)
		messages = messages.map(m =>
			m.id === editingMessage!.id
				? { ...m, content: updatedContent, edited_at: new Date().toISOString() }
				: m
		);

		editingMessage = null;
		newMessage = '';
	}

	function copyMessageText(msg: Message) {
		navigator.clipboard.writeText(msg.content);
		contextMenuMessage = null;
	}

	function insertEmoji(emoji: string) {
		newMessage += emoji;
		showEmojiPicker = false;
	}

	// Store reactions locally (in a real app this would be synced with backend)
	let messageReactions: Record<string, { emoji: string; users: string[] }[]> = $state({});

	function addReaction(msg: Message, emoji: string) {
		const msgId = msg.id;
		if (!messageReactions[msgId]) {
			messageReactions[msgId] = [];
		}

		const existingReaction = messageReactions[msgId].find(r => r.emoji === emoji);
		if (existingReaction) {
			if (existingReaction.users.includes($user!.id)) {
				// Remove user's reaction
				existingReaction.users = existingReaction.users.filter(u => u !== $user!.id);
				if (existingReaction.users.length === 0) {
					messageReactions[msgId] = messageReactions[msgId].filter(r => r.emoji !== emoji);
				}
			} else {
				// Add user to existing reaction
				existingReaction.users.push($user!.id);
			}
		} else {
			// Add new reaction
			messageReactions[msgId].push({ emoji, users: [$user!.id] });
		}

		// Force reactivity
		messageReactions = { ...messageReactions };
		contextMenuMessage = null;
	}

	function getMessageReactions(msgId: string) {
		return messageReactions[msgId] || [];
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
									{getParticipantCountText(conv)}
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
				<div class="flex-1 min-w-0">
					<div class="font-medium text-gray-900 truncate">{getConversationName(currentConversation)}</div>
					{#if typingUsers[currentConversation.id]}
						<div class="text-sm text-ekf-red">печатает...</div>
					{:else if currentConversation.type === 'channel'}
						<div class="text-sm text-gray-500">
							{getParticipantCountText(currentConversation)}
							{#if currentConversation.description}
								<span class="text-gray-400"> · </span>
								<span class="text-gray-400 truncate">{currentConversation.description}</span>
							{/if}
						</div>
					{:else if currentConversation.type === 'group' && currentConversation.participants}
						<div class="text-sm text-gray-500">{getParticipantCountText(currentConversation)}</div>
					{:else if currentConversation.participants}
						<div class="text-sm text-gray-500">был(а) недавно</div>
					{/if}
				</div>
				<!-- Кнопки звонков -->
				{#if currentConversation.type === 'direct'}
					<button
						onclick={() => startCall('audio')}
						class="p-2 hover:bg-gray-100 rounded-full transition-colors"
						title="Аудиозвонок"
					>
						<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
						</svg>
					</button>
					<button
						onclick={() => startCall('video')}
						class="p-2 hover:bg-gray-100 rounded-full transition-colors"
						title="Видеозвонок"
					>
						<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
						</svg>
					</button>
				{:else if currentConversation.type === 'channel' && currentConversation.created_by === $user?.id}
					<!-- Telegram settings button for channel creator -->
					<button
						onclick={openTelegramConfig}
						class="p-2 hover:bg-gray-100 rounded-full transition-colors relative"
						title="Настройки Telegram бота"
					>
						<svg class="w-5 h-5 {currentConversation.telegram_enabled ? 'text-blue-500' : 'text-gray-500'}" fill="currentColor" viewBox="0 0 24 24">
							<path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.894 8.221l-1.97 9.28c-.145.658-.537.818-1.084.508l-3-2.21-1.446 1.394c-.14.18-.357.223-.548.223l.188-2.85 5.18-4.686c.223-.198-.054-.308-.346-.11l-6.4 4.02-2.76-.918c-.6-.187-.612-.6.125-.89l10.782-4.156c.5-.18.94.12.78.89z"/>
						</svg>
						{#if currentConversation.telegram_enabled}
							<span class="absolute -top-0.5 -right-0.5 w-2 h-2 bg-green-500 rounded-full"></span>
						{/if}
					</button>
				{/if}
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

									<div class="flex flex-col {isOwn ? 'items-end' : 'items-start'}">
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

										{#if isVideoMessage(msg.content)}
									{@const videoInfo = getVideoInfo(msg.content)}
									{#if videoInfo && videoMessages[videoInfo.id]}
										<div class="relative">
											<button
												onclick={() => playVideo(videoInfo.id)}
												class="relative w-48 h-48 rounded-full overflow-hidden bg-black flex items-center justify-center group"
											>
												<video
													src={videoMessages[videoInfo.id].url}
													class="w-full h-full object-cover"
													loop
													muted={playingVideoId !== videoInfo.id}
													autoplay={playingVideoId === videoInfo.id}
													playsinline
												></video>
												{#if playingVideoId !== videoInfo.id}
													<div class="absolute inset-0 flex items-center justify-center bg-black/30 group-hover:bg-black/40 transition-colors">
														<svg class="w-12 h-12 text-white" fill="currentColor" viewBox="0 0 24 24">
															<path d="M8 5v14l11-7z"/>
														</svg>
													</div>
												{/if}
												<div class="absolute bottom-2 right-2 px-1.5 py-0.5 bg-black/60 rounded text-white text-xs">
													{formatRecordingTime(videoInfo.duration)}
												</div>
											</button>
										</div>
									{:else}
										<div class="w-48 h-48 rounded-full bg-gray-200 flex items-center justify-center">
											<svg class="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
											</svg>
										</div>
									{/if}
								{:else if isVoiceMessage(msg.content)}
									{@const voiceInfo = getVoiceInfo(msg.content)}
									{#if voiceInfo}
										<div class="space-y-2">
											<button
												onclick={() => playVoice(voiceInfo.id)}
												class="flex items-center gap-3 py-1 w-full"
											>
												<div class="w-10 h-10 rounded-full {isOwn ? 'bg-white/20' : 'bg-ekf-red/10'} flex items-center justify-center flex-shrink-0">
													{#if playingVoiceId === voiceInfo.id}
														<svg class="w-5 h-5 {isOwn ? 'text-white' : 'text-ekf-red'}" fill="currentColor" viewBox="0 0 24 24">
															<path d="M6 4h4v16H6V4zm8 0h4v16h-4V4z"/>
														</svg>
													{:else}
														<svg class="w-5 h-5 {isOwn ? 'text-white' : 'text-ekf-red'}" fill="currentColor" viewBox="0 0 24 24">
															<path d="M8 5v14l11-7z"/>
														</svg>
													{/if}
												</div>
												<div class="flex-1">
													<div class="flex items-center gap-1">
														{#each Array(12) as _, i}
															<div class="w-1 rounded-full {isOwn ? 'bg-white/50' : 'bg-gray-300'}" style="height: {4 + Math.random() * 12}px"></div>
														{/each}
													</div>
													<div class="text-xs {isOwn ? 'text-white/70' : 'text-gray-500'} mt-1">
														{formatRecordingTime(voiceInfo.duration)}
													</div>
												</div>
											</button>

											<!-- Transcription section -->
											{#if voiceTranscriptions[voiceInfo.id]?.text}
												<div class="text-sm {isOwn ? 'text-white/90' : 'text-gray-700'} px-2 py-1.5 {isOwn ? 'bg-white/10' : 'bg-gray-100'} rounded-lg">
													<div class="flex items-center gap-1 mb-1">
														<svg class="w-3 h-3 {isOwn ? 'text-white/60' : 'text-gray-400'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
														</svg>
														<span class="text-xs {isOwn ? 'text-white/60' : 'text-gray-400'}">Расшифровка</span>
													</div>
													{voiceTranscriptions[voiceInfo.id].text}
												</div>
											{:else if voiceTranscriptions[voiceInfo.id]?.loading}
												<div class="flex items-center gap-2 text-xs {isOwn ? 'text-white/70' : 'text-gray-500'}">
													<div class="w-3 h-3 border-2 {isOwn ? 'border-white/30 border-t-white' : 'border-gray-300 border-t-gray-600'} rounded-full animate-spin"></div>
													Расшифровка...
												</div>
											{:else if voiceTranscriptions[voiceInfo.id]?.error}
												<div class="text-xs text-red-500">
													{voiceTranscriptions[voiceInfo.id].error}
												</div>
											{:else if voiceMessages[voiceInfo.id]?.blob}
												<button
													onclick={(e) => { e.stopPropagation(); transcribeVoice(voiceInfo.id); }}
													class="flex items-center gap-1 text-xs {isOwn ? 'text-white/70 hover:text-white' : 'text-gray-500 hover:text-gray-700'} transition-colors"
												>
													<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
													</svg>
													Расшифровать
												</button>
											{/if}
										</div>
									{/if}
								{:else}
									<div class="break-words leading-relaxed text-sm">{msg.content}</div>
								{/if}

										<div class="flex items-center justify-end gap-1 mt-1">
											{#if msg.edited_at}
												<span class="text-xs {isOwn ? 'text-white/50' : 'text-gray-400'} italic">ред.</span>
											{/if}
											<span class="text-xs {isOwn ? 'text-white/70' : 'text-gray-400'}">{formatTime(msg.created_at)}</span>
											{#if isOwn}
												<svg class="w-4 h-4 text-white/70" fill="currentColor" viewBox="0 0 24 24">
													<path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z" />
												</svg>
											{/if}
										</div>
									</div>
									<!-- Reactions -->
									{#if getMessageReactions(msg.id).length > 0}
										<div class="flex flex-wrap gap-1 mt-1 {isOwn ? 'mr-1' : 'ml-1'}">
											{#each getMessageReactions(msg.id) as reaction}
												<button
													onclick={() => addReaction(msg, reaction.emoji)}
													class="inline-flex items-center gap-1 px-2 py-0.5 bg-white rounded-full shadow-sm border border-gray-100 text-xs hover:bg-gray-50 transition-colors"
												>
													<span>{reaction.emoji}</span>
													<span class="text-gray-500">{reaction.users.length}</span>
												</button>
											{/each}
										</div>
									{/if}
								</div>
							</div>
						</div>
						{/each}
					{/each}
				{/if}
			</div>

			<!-- Edit Preview -->
			{#if editingMessage}
				<div class="bg-white border-t border-gray-200 px-4 py-2 flex items-center gap-3">
					<div class="w-1 h-10 bg-blue-500 rounded"></div>
					<div class="flex-1 min-w-0">
						<div class="text-sm font-medium text-blue-500">Редактирование</div>
						<div class="text-sm text-gray-500 truncate">{editingMessage.content}</div>
					</div>
					<button onclick={cancelEdit} class="p-1 hover:bg-gray-100 rounded-full">
						<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			{:else if replyingTo}
			<!-- Reply Preview -->
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
			{#if canPostInConversation(currentConversation)}
			<div class="bg-white px-4 py-3 border-t border-gray-200 relative">
				<!-- Emoji Picker -->
				{#if showEmojiPicker}
					<div class="absolute bottom-full left-0 right-0 mb-2 mx-4 bg-white rounded-xl shadow-lg border border-gray-200 max-h-64 overflow-y-auto z-10">
						<div class="p-3">
							{#each Object.entries(emojiCategories) as [category, emojis]}
								<div class="mb-3">
									<div class="text-xs font-medium text-gray-500 mb-2">{category}</div>
									<div class="flex flex-wrap gap-1">
										{#each emojis as emoji}
											<button
												type="button"
												onclick={() => insertEmoji(emoji)}
												class="w-8 h-8 flex items-center justify-center hover:bg-gray-100 rounded text-lg transition-colors"
											>
												{emoji}
											</button>
										{/each}
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<form onsubmit={(e) => { e.preventDefault(); editingMessage ? saveEdit() : sendMessage(); }} class="flex items-center gap-2">
					<button type="button" class="p-2 hover:bg-gray-100 rounded-full transition-colors">
						<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
						</svg>
					</button>
					<button
						type="button"
						onclick={() => showEmojiPicker = !showEmojiPicker}
						class="p-2 hover:bg-gray-100 rounded-full transition-colors {showEmojiPicker ? 'bg-gray-100' : ''}"
					>
						<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</button>
					<div class="flex-1">
						<input
							type="text"
							bind:value={newMessage}
							oninput={sendTyping}
							onfocus={() => showEmojiPicker = false}
							placeholder="Сообщение"
							class="w-full px-4 py-2 bg-gray-100 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20 text-sm"
						/>
					</div>
					{#if isRecording}
						<!-- Recording UI -->
						<div class="flex items-center gap-2">
							<button
								type="button"
								onclick={cancelRecording}
								class="p-2 hover:bg-gray-100 rounded-full transition-colors"
								title="Отменить"
							>
								<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
								</svg>
							</button>
							<div class="flex items-center gap-2 px-3 py-1 bg-red-50 rounded-full">
								<div class="w-2 h-2 bg-red-500 rounded-full animate-pulse"></div>
								<span class="text-red-600 text-sm font-medium">{formatRecordingTime(recordingTime)}</span>
							</div>
							<button
								type="button"
								onclick={stopRecording}
								class="p-2 bg-ekf-red text-white rounded-full hover:bg-red-700 transition-colors"
								title="Отправить"
							>
								<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
									<path d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
								</svg>
							</button>
						</div>
					{:else if newMessage.trim()}
						<button type="submit" class="p-2 bg-ekf-red text-white rounded-full hover:bg-red-700 transition-colors">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
							</svg>
						</button>
					{:else}
						<button
							type="button"
							onclick={openVideoRecorder}
							class="p-2 hover:bg-gray-100 rounded-full transition-colors"
							title="Записать видео-кружок"
						>
							<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
							</svg>
						</button>
						<button
							type="button"
							onclick={startRecording}
							class="p-2 hover:bg-gray-100 rounded-full transition-colors"
							title="Записать голосовое сообщение"
						>
							<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
							</svg>
						</button>
					{/if}
				</form>
			</div>
			{:else}
			<!-- Channel read-only notice for non-admins -->
			<div class="bg-gray-100 px-4 py-3 border-t border-gray-200 text-center">
				<div class="flex items-center justify-center gap-2 text-gray-500 text-sm">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
					</svg>
					<span>Только администраторы могут публиковать в канале</span>
				</div>
			</div>
			{/if}
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
		class="fixed bg-white rounded-xl shadow-lg z-50 border border-gray-200 overflow-hidden min-w-48 max-h-80"
		style="left: {contextMenuPosition.x}px; top: {contextMenuPosition.y}px; max-width: calc(100vw - 32px);"
		onclick={(e) => e.stopPropagation()}
	>
		<!-- Quick Reactions -->
		<div class="flex items-center gap-1 px-3 py-2 border-b border-gray-100">
			{#each ['👍', '❤️', '😂', '😮', '😢', '🔥'] as emoji}
				<button
					onclick={() => addReaction(contextMenuMessage!, emoji)}
					class="w-8 h-8 flex items-center justify-center hover:bg-gray-100 rounded-full text-lg transition-transform hover:scale-110"
				>
					{emoji}
				</button>
			{/each}
		</div>
		<!-- Menu items -->
		<div class="py-1">
			<button onclick={() => replyToMessage(contextMenuMessage!)} class="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
				</svg>
				Ответить
			</button>
			{#if contextMenuMessage.sender_id === $user?.id}
				<button onclick={() => startEditMessage(contextMenuMessage!)} class="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
					</svg>
					Редактировать
				</button>
			{/if}
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
	</div>
{/if}

<!-- Video Recorder Modal -->
{#if showVideoRecorder}
	<div class="fixed inset-0 bg-black/80 flex items-center justify-center z-50">
		<div class="bg-gray-900 rounded-2xl w-96 p-6 text-center shadow-2xl">
			<h3 class="text-lg font-semibold text-white mb-4">Видео-кружок</h3>

			<!-- Circular Video Preview -->
			<div class="relative mx-auto mb-6">
				<div class="w-64 h-64 rounded-full overflow-hidden mx-auto border-4 {isRecordingVideo ? 'border-red-500' : 'border-gray-700'}">
					<video
						bind:this={videoPreviewElement}
						autoplay
						playsinline
						muted
						class="w-full h-full object-cover scale-x-[-1]"
					></video>
				</div>

				{#if isRecordingVideo}
					<!-- Recording indicator -->
					<div class="absolute top-2 left-1/2 -translate-x-1/2 flex items-center gap-2 px-3 py-1 bg-red-500 rounded-full">
						<div class="w-2 h-2 bg-white rounded-full animate-pulse"></div>
						<span class="text-white text-sm font-medium">{formatRecordingTime(videoRecordingTime)}</span>
					</div>

					<!-- Progress ring -->
					<svg class="absolute inset-0 w-full h-full -rotate-90" viewBox="0 0 100 100">
						<circle
							cx="50"
							cy="50"
							r="48"
							fill="none"
							stroke="rgba(239, 68, 68, 0.3)"
							stroke-width="2"
						/>
						<circle
							cx="50"
							cy="50"
							r="48"
							fill="none"
							stroke="#ef4444"
							stroke-width="2"
							stroke-dasharray="{(videoRecordingTime / 60) * 301.59} 301.59"
							stroke-linecap="round"
						/>
					</svg>
				{/if}
			</div>

			<!-- Timer info -->
			<p class="text-gray-400 text-sm mb-4">
				{#if isRecordingVideo}
					Максимум 60 секунд
				{:else}
					Нажмите для записи
				{/if}
			</p>

			<!-- Controls -->
			<div class="flex items-center justify-center gap-4">
				<button
					onclick={closeVideoRecorder}
					class="w-12 h-12 bg-gray-700 hover:bg-gray-600 rounded-full flex items-center justify-center transition-colors"
					title="Отмена"
				>
					<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>

				{#if isRecordingVideo}
					<button
						onclick={stopVideoRecording}
						class="w-16 h-16 bg-red-500 hover:bg-red-600 rounded-full flex items-center justify-center transition-colors shadow-lg"
						title="Остановить и отправить"
					>
						<svg class="w-8 h-8 text-white" fill="currentColor" viewBox="0 0 24 24">
							<path d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
						</svg>
					</button>
				{:else}
					<button
						onclick={startVideoRecording}
						class="w-16 h-16 bg-red-500 hover:bg-red-600 rounded-full flex items-center justify-center transition-colors shadow-lg"
						title="Начать запись"
					>
						<div class="w-6 h-6 bg-white rounded-full"></div>
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}

<!-- Call Modal -->
{#if showCallModal}
	<div class="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
		<div class="bg-gradient-to-b from-gray-800 to-gray-900 rounded-2xl w-80 p-6 text-center shadow-2xl">
			<!-- Avatar -->
			<div class="mb-4">
				{#if currentConversation && getConversationAvatar(currentConversation)}
					<img
						src="data:image/jpeg;base64,{getConversationAvatar(currentConversation)}"
						alt=""
						class="w-24 h-24 rounded-full mx-auto object-cover border-4 border-white/20"
					/>
				{:else}
					<div class="w-24 h-24 rounded-full mx-auto {getAvatarColor(currentConversation ? getConversationName(currentConversation) : '')} text-white flex items-center justify-center text-3xl font-medium border-4 border-white/20">
						{currentConversation ? getInitials(getConversationName(currentConversation)) : '?'}
					</div>
				{/if}
			</div>

			<!-- Name -->
			<h3 class="text-xl font-semibold text-white mb-1">
				{currentConversation ? getConversationName(currentConversation) : ''}
			</h3>

			<!-- Status -->
			<div class="text-gray-300 text-sm mb-8">
				{#if callStatus === 'calling'}
					<div class="flex items-center justify-center gap-2">
						<span>Вызов</span>
						<span class="flex gap-1">
							<span class="w-1.5 h-1.5 bg-white rounded-full animate-bounce" style="animation-delay: 0ms"></span>
							<span class="w-1.5 h-1.5 bg-white rounded-full animate-bounce" style="animation-delay: 150ms"></span>
							<span class="w-1.5 h-1.5 bg-white rounded-full animate-bounce" style="animation-delay: 300ms"></span>
						</span>
					</div>
				{:else if callStatus === 'connected'}
					<span class="text-green-400">{callType === 'video' ? 'Видеозвонок' : 'Аудиозвонок'} активен</span>
				{:else}
					<span class="text-red-400">Звонок завершён</span>
				{/if}
			</div>

			<!-- Video preview placeholder -->
			{#if callType === 'video' && callStatus === 'connected'}
				<div class="bg-gray-700 rounded-xl h-32 mb-6 flex items-center justify-center">
					<svg class="w-12 h-12 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
					</svg>
				</div>
			{/if}

			<!-- Controls -->
			<div class="flex items-center justify-center gap-4">
				{#if callStatus === 'connected'}
					<!-- Mute button -->
					<button class="w-12 h-12 bg-gray-700 hover:bg-gray-600 rounded-full flex items-center justify-center transition-colors">
						<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
						</svg>
					</button>

					{#if callType === 'video'}
						<!-- Camera toggle -->
						<button class="w-12 h-12 bg-gray-700 hover:bg-gray-600 rounded-full flex items-center justify-center transition-colors">
							<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
							</svg>
						</button>
					{/if}
				{/if}

				<!-- End call button -->
				<button
					onclick={endCall}
					class="w-14 h-14 bg-red-500 hover:bg-red-600 rounded-full flex items-center justify-center transition-colors shadow-lg"
				>
					<svg class="w-7 h-7 text-white transform rotate-135" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
					</svg>
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Telegram Bot Configuration Modal -->
{#if showTelegramConfig}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
		<div class="bg-white rounded-2xl w-full max-w-md mx-4 shadow-2xl">
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<div class="flex items-center gap-3">
					<div class="w-10 h-10 bg-blue-500 rounded-full flex items-center justify-center">
						<svg class="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
							<path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.894 8.221l-1.97 9.28c-.145.658-.537.818-1.084.508l-3-2.21-1.446 1.394c-.14.18-.357.223-.548.223l.188-2.85 5.18-4.686c.223-.198-.054-.308-.346-.11l-6.4 4.02-2.76-.918c-.6-.187-.612-.6.125-.89l10.782-4.156c.5-.18.94.12.78.89z"/>
						</svg>
					</div>
					<div>
						<h3 class="text-lg font-semibold text-gray-900">Telegram бот</h3>
						<p class="text-sm text-gray-500">Настройка интеграции с Telegram</p>
					</div>
				</div>
				<button onclick={() => showTelegramConfig = false} class="p-2 hover:bg-gray-100 rounded-full transition-colors">
					<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			{#if telegramLoading}
				<div class="p-6 flex items-center justify-center">
					<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
				</div>
			{:else}
				<div class="p-6 space-y-4">
					<!-- Enable toggle -->
					<div class="flex items-center justify-between">
						<div>
							<div class="font-medium text-gray-900">Включить интеграцию</div>
							<div class="text-sm text-gray-500">Сообщения из Telegram будут появляться в канале</div>
						</div>
						<button
							onclick={() => telegramEnabled = !telegramEnabled}
							class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {telegramEnabled ? 'bg-blue-500' : 'bg-gray-200'}"
						>
							<span class="inline-block h-4 w-4 transform rounded-full bg-white transition {telegramEnabled ? 'translate-x-6' : 'translate-x-1'}"></span>
						</button>
					</div>

					{#if telegramEnabled}
						<!-- Bot Token -->
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Bot Token</label>
							<input
								type="password"
								bind:value={telegramBotToken}
								placeholder="1234567890:ABCdef..."
								class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/20"
							/>
							<p class="text-xs text-gray-500 mt-1">Получите токен у @BotFather в Telegram</p>
						</div>

						<!-- Chat ID -->
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Chat ID</label>
							<input
								type="text"
								bind:value={telegramChatId}
								placeholder="-1001234567890"
								class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500/20"
							/>
							<p class="text-xs text-gray-500 mt-1">ID группы или канала в Telegram (используйте @userinfobot)</p>
						</div>

						<!-- Webhook URL -->
						{#if telegramWebhookUrl}
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Webhook URL</label>
								<div class="flex gap-2">
									<input
										type="text"
										value={telegramWebhookUrl}
										readonly
										class="flex-1 px-4 py-2 bg-gray-100 border border-gray-200 rounded-lg text-sm text-gray-600"
									/>
									<button
										onclick={() => navigator.clipboard.writeText(telegramWebhookUrl)}
										class="px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
										title="Копировать"
									>
										<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
										</svg>
									</button>
								</div>
								<p class="text-xs text-gray-500 mt-1">Установите этот URL как webhook в настройках бота</p>
							</div>
						{/if}

						<!-- Instructions -->
						<div class="bg-blue-50 rounded-lg p-4">
							<h4 class="font-medium text-blue-900 mb-2">Инструкция по настройке</h4>
							<ol class="text-sm text-blue-800 list-decimal list-inside space-y-1">
								<li>Создайте бота через @BotFather и получите токен</li>
								<li>Добавьте бота в вашу Telegram группу/канал как администратора</li>
								<li>Получите Chat ID группы/канала</li>
								<li>Введите данные выше и сохраните</li>
								<li>Установите webhook URL в настройках бота</li>
							</ol>
						</div>
					{/if}
				</div>

				<!-- Footer -->
				<div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
					<button
						onclick={() => showTelegramConfig = false}
						class="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
					>
						Отмена
					</button>
					<button
						onclick={saveTelegramConfig}
						class="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
					>
						Сохранить
					</button>
				</div>
			{/if}
		</div>
	</div>
{/if}
