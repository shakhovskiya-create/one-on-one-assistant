<script lang="ts">
	import { onMount } from 'svelte';
	import { mail } from '$lib/api/client';
	import type { MailFolder, EmailMessage, EmailPerson } from '$lib/api/client';
	import { user } from '$lib/stores/auth';
	import { browser } from '$app/environment';

	// State
	let folders: MailFolder[] = $state([]);
	let emails: EmailMessage[] = $state([]);
	let selectedFolder: MailFolder | null = $state(null);
	let selectedEmail: EmailMessage | null = $state(null);
	let loading = $state(true);
	let loadingEmails = $state(false);
	let error = $state('');

	// Credentials (from main login via ews_credentials)
	let credentials = $state({ username: '', password: '' });
	let showLogin = $state(true);

	// Compose
	let showCompose = $state(false);
	let composeTo = $state('');
	let composeCc = $state('');
	let composeSubject = $state('');
	let composeBody = $state('');
	let sending = $state(false);

	// Search
	let searchQuery = $state('');

	// Check for saved credentials from main login
	onMount(() => {
		if (browser) {
			// First try ews_credentials from main login
			const ewsCreds = sessionStorage.getItem('ews_credentials');
			if (ewsCreds) {
				try {
					credentials = JSON.parse(ewsCreds);
					showLogin = false;
					loadFolders();
					loading = false;
					return;
				} catch {
					// Fall through to show login
				}
			}
			// Fallback to mail_credentials for backwards compatibility
			const savedCreds = sessionStorage.getItem('mail_credentials');
			if (savedCreds) {
				try {
					credentials = JSON.parse(savedCreds);
					showLogin = false;
					loadFolders();
				} catch {
					showLogin = true;
				}
			}
		}
		loading = false;
	});

	async function handleLogin() {
		if (!credentials.username || !credentials.password) {
			error = 'Введите имя пользователя и пароль';
			return;
		}

		loading = true;
		error = '';

		try {
			const result = await mail.getFolders(credentials.username, credentials.password);
			folders = result;

			if (browser) {
				sessionStorage.setItem('ews_credentials', JSON.stringify(credentials));
			}

			showLogin = false;

			// Select inbox by default
			const inbox = folders.find(f => f.display_name.toLowerCase() === 'inbox' || f.display_name.toLowerCase() === 'входящие');
			if (inbox) {
				selectFolder(inbox);
			} else if (folders.length > 0) {
				selectFolder(folders[0]);
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Ошибка входа';
		} finally {
			loading = false;
		}
	}

	async function loadFolders() {
		try {
			folders = await mail.getFolders(credentials.username, credentials.password);
			const inbox = folders.find(f => f.display_name.toLowerCase() === 'inbox' || f.display_name.toLowerCase() === 'входящие');
			if (inbox) {
				selectFolder(inbox);
			}
		} catch (e) {
			console.error('Failed to load folders:', e);
		}
	}

	async function selectFolder(folder: MailFolder) {
		selectedFolder = folder;
		selectedEmail = null;
		loadingEmails = true;

		try {
			emails = await mail.getEmails(credentials.username, credentials.password, folder.id, 50);
		} catch (e) {
			console.error('Failed to load emails:', e);
			emails = [];
		} finally {
			loadingEmails = false;
		}
	}

	function selectEmail(email: EmailMessage) {
		selectedEmail = email;
		if (!email.is_read) {
			// Mark as read
			mail.markAsRead({
				username: credentials.username,
				password: credentials.password,
				item_id: email.id
			}).then(() => {
				email.is_read = true;
				emails = [...emails];
			}).catch(console.error);
		}
	}

	async function deleteEmail(email: EmailMessage) {
		try {
			await mail.deleteEmail({
				username: credentials.username,
				password: credentials.password,
				item_id: email.id
			});
			emails = emails.filter(e => e.id !== email.id);
			if (selectedEmail?.id === email.id) {
				selectedEmail = null;
			}
		} catch (e) {
			console.error('Failed to delete email:', e);
		}
	}

	async function sendEmail() {
		if (!composeTo.trim() || !composeSubject.trim()) {
			return;
		}

		sending = true;
		try {
			const toList = composeTo.split(',').map(e => e.trim()).filter(e => e);
			const ccList = composeCc ? composeCc.split(',').map(e => e.trim()).filter(e => e) : [];

			await mail.sendEmail({
				username: credentials.username,
				password: credentials.password,
				to: toList,
				cc: ccList,
				subject: composeSubject,
				body: composeBody
			});

			showCompose = false;
			composeTo = '';
			composeCc = '';
			composeSubject = '';
			composeBody = '';

			// Refresh sent folder if selected
			if (selectedFolder?.display_name.toLowerCase().includes('sent') || selectedFolder?.display_name.toLowerCase().includes('отправленные')) {
				selectFolder(selectedFolder);
			}
		} catch (e) {
			console.error('Failed to send email:', e);
			error = e instanceof Error ? e.message : 'Не удалось отправить';
		} finally {
			sending = false;
		}
	}

	function logout() {
		if (browser) {
			sessionStorage.removeItem('ews_credentials');
			sessionStorage.removeItem('mail_credentials');
		}
		credentials = { username: '', password: '' };
		showLogin = true;
		folders = [];
		emails = [];
		selectedFolder = null;
		selectedEmail = null;
	}

	function formatDate(dateStr: string): string {
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

	function getFolderIcon(name: string): string {
		const lower = name.toLowerCase();
		if (lower === 'inbox' || lower === 'входящие') return 'inbox';
		if (lower.includes('sent') || lower.includes('отправленные')) return 'send';
		if (lower.includes('draft') || lower.includes('черновик')) return 'draft';
		if (lower.includes('deleted') || lower.includes('удаленные')) return 'trash';
		if (lower.includes('spam') || lower.includes('junk') || lower.includes('нежелательн')) return 'spam';
		if (lower.includes('archive') || lower.includes('архив')) return 'archive';
		return 'folder';
	}

	function getPersonDisplay(person?: EmailPerson): string {
		if (!person) return 'Неизвестный';
		return person.name || person.email || 'Неизвестный';
	}

	let filteredEmails = $derived(
		searchQuery
			? emails.filter(e =>
					e.subject.toLowerCase().includes(searchQuery.toLowerCase()) ||
					(e.from?.name || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
					(e.from?.email || '').toLowerCase().includes(searchQuery.toLowerCase())
				)
			: emails
	);
</script>

<svelte:head>
	<title>Почта - EKF Hub</title>
</svelte:head>

{#if showLogin}
	<!-- Login Form -->
	<div class="flex items-center justify-center min-h-[calc(100vh-150px)]">
		<div class="bg-white rounded-xl shadow-sm p-8 w-full max-w-md">
			<div class="text-center mb-6">
				<div class="w-16 h-16 mx-auto mb-4 bg-ekf-red/10 rounded-full flex items-center justify-center">
					<svg class="w-8 h-8 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
					</svg>
				</div>
				<h2 class="text-xl font-semibold text-gray-900">Вход в почту</h2>
				<p class="text-sm text-gray-500 mt-1">Используйте учетные данные Exchange</p>
			</div>

			{#if error}
				<div class="mb-4 p-3 bg-red-50 text-red-700 rounded-lg text-sm">{error}</div>
			{/if}

			<form onsubmit={(e) => { e.preventDefault(); handleLogin(); }}>
				<div class="mb-4">
					<label class="block text-sm font-medium text-gray-700 mb-1">Имя пользователя</label>
					<input
						type="text"
						bind:value={credentials.username}
						placeholder="username или email"
						class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
					/>
				</div>
				<div class="mb-6">
					<label class="block text-sm font-medium text-gray-700 mb-1">Пароль</label>
					<input
						type="password"
						bind:value={credentials.password}
						class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
					/>
				</div>
				<button
					type="submit"
					disabled={loading}
					class="w-full py-2 bg-ekf-red text-white rounded-lg font-medium hover:bg-red-700 transition-colors disabled:opacity-50"
				>
					{loading ? 'Входим...' : 'Войти'}
				</button>
			</form>
		</div>
	</div>
{:else}
	<!-- Mail Interface -->
	<div class="h-[calc(100vh-100px)] flex bg-white rounded-xl shadow-sm overflow-hidden">
		<!-- Folders sidebar -->
		<div class="w-60 border-r border-gray-200 flex flex-col bg-gray-50">
			<div class="p-3">
				<button
					onclick={() => showCompose = true}
					class="w-full py-2 px-4 bg-ekf-red text-white rounded-lg font-medium hover:bg-red-700 transition-colors flex items-center justify-center gap-2"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					Написать
				</button>
			</div>

			<nav class="flex-1 overflow-y-auto px-2">
				{#each folders as folder}
					<button
						onclick={() => selectFolder(folder)}
						class="w-full px-3 py-2 flex items-center gap-3 rounded-lg text-left text-sm transition-colors mb-1
							{selectedFolder?.id === folder.id ? 'bg-ekf-red/10 text-ekf-red' : 'text-gray-700 hover:bg-gray-100'}"
					>
						<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							{#if getFolderIcon(folder.display_name) === 'inbox'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
							{:else if getFolderIcon(folder.display_name) === 'send'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
							{:else if getFolderIcon(folder.display_name) === 'draft'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
							{:else if getFolderIcon(folder.display_name) === 'trash'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
							{:else if getFolderIcon(folder.display_name) === 'spam'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
							{:else if getFolderIcon(folder.display_name) === 'archive'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
							{:else}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
							{/if}
						</svg>
						<span class="flex-1 truncate">{folder.display_name}</span>
						{#if folder.unread_count > 0}
							<span class="text-xs bg-ekf-red text-white px-1.5 py-0.5 rounded-full">{folder.unread_count}</span>
						{/if}
					</button>
				{/each}
			</nav>

			<div class="p-3 border-t border-gray-200">
				<button onclick={logout} class="w-full text-sm text-gray-500 hover:text-gray-700">
					Выйти
				</button>
			</div>
		</div>

		<!-- Email list -->
		<div class="w-80 border-r border-gray-200 flex flex-col">
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

			<div class="flex-1 overflow-y-auto">
				{#if loadingEmails}
					<div class="flex items-center justify-center h-32">
						<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-ekf-red"></div>
					</div>
				{:else if filteredEmails.length === 0}
					<div class="text-center py-12 text-gray-500 text-sm">Нет писем</div>
				{:else}
					{#each filteredEmails as email}
						<button
							onclick={() => selectEmail(email)}
							class="w-full px-4 py-3 text-left border-b border-gray-100 hover:bg-gray-50 transition-colors
								{selectedEmail?.id === email.id ? 'bg-ekf-red/5' : ''}
								{!email.is_read ? 'bg-blue-50/50' : ''}"
						>
							<div class="flex items-start gap-3">
								<div class="flex-1 min-w-0">
									<div class="flex items-center gap-2 mb-1">
										{#if !email.is_read}
											<div class="w-2 h-2 bg-ekf-red rounded-full flex-shrink-0"></div>
										{/if}
										<span class="text-sm {!email.is_read ? 'font-semibold' : ''} text-gray-900 truncate">
											{getPersonDisplay(email.from)}
										</span>
										<span class="text-xs text-gray-400 flex-shrink-0 ml-auto">
											{formatDate(email.received_at)}
										</span>
									</div>
									<div class="text-sm {!email.is_read ? 'font-medium' : ''} text-gray-800 truncate">{email.subject}</div>
									{#if email.body_preview}
										<div class="text-xs text-gray-500 truncate mt-0.5">{email.body_preview}</div>
									{/if}
								</div>
								{#if email.has_attachments}
									<svg class="w-4 h-4 text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
									</svg>
								{/if}
							</div>
						</button>
					{/each}
				{/if}
			</div>
		</div>

		<!-- Email content -->
		<div class="flex-1 flex flex-col bg-gray-50">
			{#if selectedEmail}
				<div class="bg-white border-b border-gray-200 px-6 py-4">
					<div class="flex items-start justify-between mb-3">
						<h2 class="text-lg font-medium text-gray-900">{selectedEmail.subject}</h2>
						<div class="flex items-center gap-2">
							<button class="p-2 hover:bg-gray-100 rounded-lg" title="Ответить">
								<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
								</svg>
							</button>
							<button class="p-2 hover:bg-gray-100 rounded-lg" title="Переслать">
								<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
								</svg>
							</button>
							<button onclick={() => deleteEmail(selectedEmail!)} class="p-2 hover:bg-gray-100 rounded-lg" title="Удалить">
								<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
								</svg>
							</button>
						</div>
					</div>
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-full bg-ekf-red/10 flex items-center justify-center">
							<span class="text-ekf-red font-medium">
								{(selectedEmail.from?.name || selectedEmail.from?.email || '?').charAt(0).toUpperCase()}
							</span>
						</div>
						<div>
							<div class="text-sm font-medium text-gray-900">{getPersonDisplay(selectedEmail.from)}</div>
							<div class="text-xs text-gray-500">
								Кому: {selectedEmail.to?.map(p => getPersonDisplay(p)).join(', ') || 'Вам'}
							</div>
						</div>
						<span class="ml-auto text-xs text-gray-400">
							{new Date(selectedEmail.received_at).toLocaleString('ru-RU')}
						</span>
					</div>
				</div>
				<div class="flex-1 overflow-y-auto p-6">
					<div class="bg-white rounded-lg p-6 shadow-sm">
						{@html selectedEmail.body || '<p class="text-gray-500">Нет содержимого</p>'}
					</div>
				</div>
			{:else}
				<div class="flex-1 flex items-center justify-center">
					<div class="text-center">
						<div class="w-20 h-20 mx-auto mb-4 bg-gray-100 rounded-full flex items-center justify-center">
							<svg class="w-10 h-10 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
							</svg>
						</div>
						<p class="text-gray-500 text-sm">Выберите письмо для просмотра</p>
					</div>
				</div>
			{/if}
		</div>
	</div>

	<!-- Compose Modal -->
	{#if showCompose}
		<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
			<div class="bg-white rounded-xl shadow-xl w-full max-w-2xl max-h-[90vh] flex flex-col">
				<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
					<h3 class="text-lg font-semibold">Новое письмо</h3>
					<button onclick={() => showCompose = false} class="p-1 hover:bg-gray-100 rounded">
						<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
				<div class="flex-1 overflow-y-auto p-6">
					<div class="space-y-4">
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Кому</label>
							<input
								type="text"
								bind:value={composeTo}
								placeholder="email@example.com (через запятую для нескольких)"
								class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
							/>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Копия</label>
							<input
								type="text"
								bind:value={composeCc}
								placeholder="email@example.com"
								class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
							/>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Тема</label>
							<input
								type="text"
								bind:value={composeSubject}
								class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
							/>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Сообщение</label>
							<textarea
								bind:value={composeBody}
								rows="10"
								class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20 resize-none"
							></textarea>
						</div>
					</div>
				</div>
				<div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
					<button onclick={() => showCompose = false} class="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg">
						Отмена
					</button>
					<button
						onclick={sendEmail}
						disabled={sending || !composeTo.trim() || !composeSubject.trim()}
						class="px-4 py-2 bg-ekf-red text-white rounded-lg font-medium hover:bg-red-700 transition-colors disabled:opacity-50 flex items-center gap-2"
					>
						{#if sending}
							<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
						{/if}
						Отправить
					</button>
				</div>
			</div>
		</div>
	{/if}
{/if}
