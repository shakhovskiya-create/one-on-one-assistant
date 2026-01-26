<script lang="ts">
	import { onMount } from 'svelte';
	import { mail, employees } from '$lib/api/client';
	import type { MailFolder, EmailMessage, EmailPerson, EmailAttachment, Employee } from '$lib/api/client';
	import { user } from '$lib/stores/auth';
	import { browser } from '$app/environment';
	import RichTextEditor from '$lib/components/RichTextEditor.svelte';
	import AttachmentPreview from '$lib/components/AttachmentPreview.svelte';
	import { sanitizeEmailHtml } from '$lib/utils/sanitize';

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
	let rememberMe = $state(true); // Remember credentials by default

	// Compose
	let showCompose = $state(false);
	let composeTo = $state('');
	let composeCc = $state('');
	let composeSubject = $state('');
	let composeBody = $state('');
	let sending = $state(false);
	let composeMode: 'new' | 'reply' | 'replyAll' | 'forward' = $state('new');
	let composeAttachments: { name: string; content: string; size: number }[] = $state([]);

	// Full email view modal
	let showEmailModal = $state(false);

	// Attachments
	let attachments: EmailAttachment[] = $state([]);
	let loadingAttachments = $state(false);
	let downloadingAttachment = $state<string | null>(null);
	let attachmentError = $state('');

	// Attachment preview
	let previewAttachment: { name: string; contentType: string; content: string } | null = $state(null);

	// Meeting response
	let respondingToMeeting = $state<'Accept' | 'Decline' | 'Tentative' | null>(null);
	let meetingResponseSuccess = $state<string | null>(null);

	// Search & Filter
	let searchQuery = $state('');
	let showOnlyUnread = $state(false);

	// Threading
	let showThreaded = $state(true);
	let expandedThreads = $state<Set<string>>(new Set());

	// Sidebar collapsed state
	let sidebarCollapsed = $state(false);

	// Employees for photo lookup
	let employeeList: Employee[] = $state([]);
	let employeesByEmail = $state<Map<string, Employee>>(new Map());

	// Auto-refresh interval (60 seconds)
	let refreshInterval: ReturnType<typeof setInterval> | null = null;

	// Load employees for photo lookup
	async function loadEmployees() {
		try {
			const list = await employees.list();
			employeeList = list;
			// Build email -> employee map
			const map = new Map<string, Employee>();
			for (const emp of list) {
				if (emp.email) {
					map.set(emp.email.toLowerCase(), emp);
				}
			}
			employeesByEmail = map;
		} catch (err) {
			console.error('Failed to load employees for photos:', err);
		}
	}

	// Get employee photo by email
	function getEmployeePhoto(email: string | undefined): string | null {
		if (!email) return null;
		const emp = employeesByEmail.get(email.toLowerCase());
		return emp?.photo_base64 || null;
	}

	function startAutoRefresh() {
		if (refreshInterval) clearInterval(refreshInterval);
		refreshInterval = setInterval(() => {
			if (selectedFolder && !loadingEmails) {
				refreshEmails();
			}
		}, 60000); // 60 seconds
	}

	async function refreshEmails() {
		if (!selectedFolder) return;
		try {
			const newEmails = await mail.getEmails(credentials.username, credentials.password, selectedFolder.id, 50);
			emails = newEmails;
			// Update folder unread count
			const unreadCount = newEmails.filter(e => !e.is_read).length;
			folders = folders.map(f =>
				f.id === selectedFolder?.id ? { ...f, unread_count: unreadCount } : f
			);
		} catch (e) {
			console.error('Auto-refresh failed:', e);
		}
	}

	// Check for saved credentials from main login
	onMount(() => {
		// Load employees for photo lookup (async, don't block)
		loadEmployees();

		if (browser) {
			// First try localStorage (persistent)
			const localCreds = localStorage.getItem('ews_credentials');
			if (localCreds) {
				try {
					credentials = JSON.parse(localCreds);
					showLogin = false;
					loadFolders();
					startAutoRefresh();
					loading = false;
					return;
				} catch {
					localStorage.removeItem('ews_credentials');
				}
			}
			// Then try sessionStorage (temporary)
			const sessionCreds = sessionStorage.getItem('ews_credentials');
			if (sessionCreds) {
				try {
					credentials = JSON.parse(sessionCreds);
					showLogin = false;
					loadFolders();
					startAutoRefresh();
					loading = false;
					return;
				} catch {
					sessionStorage.removeItem('ews_credentials');
				}
			}
		}
		loading = false;

		// Cleanup on unmount
		return () => {
			if (refreshInterval) clearInterval(refreshInterval);
		};
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
				// Store in localStorage if "Remember me" is checked, otherwise sessionStorage
				if (rememberMe) {
					localStorage.setItem('ews_credentials', JSON.stringify(credentials));
					sessionStorage.removeItem('ews_credentials');
				} else {
					sessionStorage.setItem('ews_credentials', JSON.stringify(credentials));
					localStorage.removeItem('ews_credentials');
				}
			}

			showLogin = false;
			startAutoRefresh();

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

	let loadingBody = $state(false);
	let bodyError = $state('');

	// Replace cid: references with data URIs for inline images
	async function replaceInlineImages(body: string, inlineAttachments: EmailAttachment[]): Promise<string> {
		if (!body || inlineAttachments.length === 0) return body;

		let updatedBody = body;
		const cidMap = new Map<string, EmailAttachment>();

		// Build map of content_id -> attachment
		for (const att of inlineAttachments) {
			if (att.content_id) {
				// Content-ID can have angle brackets: <image001.png@01DB0A5C.90F4D140>
				const cleanCid = att.content_id.replace(/^<|>$/g, '');
				cidMap.set(cleanCid, att);
			}
		}

		// Find all cid: references in the body
		const cidRegex = /src=["']cid:([^"']+)["']/gi;
		const matches = [...body.matchAll(cidRegex)];

		for (const match of matches) {
			const cidRef = match[1]; // The content ID from the cid: reference
			const attachment = cidMap.get(cidRef);

			if (attachment) {
				try {
					const result = await mail.getAttachmentContent({
						username: credentials.username,
						password: credentials.password,
						attachment_id: attachment.id
					});

					if (result.content) {
						// Create data URI
						const dataUri = `data:${result.content_type};base64,${result.content}`;
						updatedBody = updatedBody.replace(
							new RegExp(`src=["']cid:${escapeRegExp(cidRef)}["']`, 'gi'),
							`src="${dataUri}"`
						);
					}
				} catch (err) {
					console.error('Failed to load inline image:', err);
				}
			}
		}

		return updatedBody;
	}

	// Helper to escape regex special characters
	function escapeRegExp(string: string): string {
		return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
	}

	async function selectEmail(email: EmailMessage) {
		selectedEmail = email;
		bodyError = '';
		attachments = [];
		attachmentError = '';

		// Load attachments first if email has them (needed for inline images)
		let inlineAttachments: EmailAttachment[] = [];
		if (email.has_attachments) {
			loadingAttachments = true;
			try {
				const result = await mail.getAttachments({
					username: credentials.username,
					password: credentials.password,
					item_id: email.id,
					change_key: email.change_key
				});
				attachments = result.attachments || [];
				inlineAttachments = attachments.filter(a => a.is_inline && a.content_id);
				if (attachments.length === 0) {
					attachmentError = 'Вложения не найдены (возможно, они были удалены)';
				}
			} catch (err) {
				console.error('Failed to load attachments:', err);
				attachmentError = err instanceof Error ? err.message : 'Ошибка загрузки вложений';
			} finally {
				loadingAttachments = false;
			}
		}

		// Load email body if not already loaded
		if (!email.body || email.body === '') {
			loadingBody = true;
			try {
				const result = await mail.getEmailBody({
					username: credentials.username,
					password: credentials.password,
					item_id: email.id
				});
				// Update the email with body
				let bodyContent = result.body || '';

				// Replace cid: references with actual image data
				if (bodyContent && inlineAttachments.length > 0) {
					bodyContent = await replaceInlineImages(bodyContent, inlineAttachments);
				}

				const updatedEmail = { ...email, body: bodyContent };
				selectedEmail = updatedEmail;
				emails = emails.map(e => e.id === email.id ? updatedEmail : e);

				if (!bodyContent) {
					bodyError = 'Письмо не содержит текста';
				}
			} catch (err) {
				console.error('Failed to load email body:', err);
				bodyError = err instanceof Error ? err.message : 'Ошибка загрузки письма';
			} finally {
				loadingBody = false;
			}
		} else if (email.body && inlineAttachments.length > 0) {
			// Body is already loaded but we need to update inline images
			loadingBody = true;
			try {
				const updatedBody = await replaceInlineImages(email.body, inlineAttachments);
				const updatedEmail = { ...email, body: updatedBody };
				selectedEmail = updatedEmail;
				emails = emails.map(e => e.id === email.id ? updatedEmail : e);
			} finally {
				loadingBody = false;
			}
		}

		if (!email.is_read) {
			// Mark as read in UI immediately for better UX
			emails = emails.map(e =>
				e.id === email.id ? { ...e, is_read: true } : e
			);
			selectedEmail = { ...selectedEmail, is_read: true };

			// Update folder unread count
			if (selectedFolder) {
				folders = folders.map(f =>
					f.id === selectedFolder?.id
						? { ...f, unread_count: Math.max(0, f.unread_count - 1) }
						: f
				);
			}

			// Mark as read on server
			mail.markAsRead({
				username: credentials.username,
				password: credentials.password,
				item_id: email.id,
				change_key: email.change_key
			}).catch(err => {
				console.error('Failed to mark as read:', err);
				// Revert on error
				emails = emails.map(e =>
					e.id === email.id ? { ...e, is_read: false } : e
				);
				if (selectedEmail) {
					selectedEmail = { ...selectedEmail, is_read: false };
				}
			});
		}
	}

	async function deleteEmail(email: EmailMessage) {
		try {
			await mail.deleteEmail({
				username: credentials.username,
				password: credentials.password,
				item_id: email.id,
				change_key: email.change_key
			});
			emails = emails.filter(e => e.id !== email.id);
			if (selectedEmail?.id === email.id) {
				selectedEmail = null;
			}
		} catch (e) {
			console.error('Failed to delete email:', e);
		}
	}

	async function downloadAttachment(attachment: EmailAttachment) {
		downloadingAttachment = attachment.id;
		try {
			const result = await mail.getAttachmentContent({
				username: credentials.username,
				password: credentials.password,
				attachment_id: attachment.id
			});

			// Decode base64 and create download
			const byteCharacters = atob(result.content);
			const byteNumbers = new Array(byteCharacters.length);
			for (let i = 0; i < byteCharacters.length; i++) {
				byteNumbers[i] = byteCharacters.charCodeAt(i);
			}
			const byteArray = new Uint8Array(byteNumbers);
			const blob = new Blob([byteArray], { type: result.content_type });

			// Create download link
			const url = URL.createObjectURL(blob);
			const link = document.createElement('a');
			link.href = url;
			link.download = result.name || attachment.name;
			document.body.appendChild(link);
			link.click();
			document.body.removeChild(link);
			URL.revokeObjectURL(url);
		} catch (err) {
			console.error('Failed to download attachment:', err);
		} finally {
			downloadingAttachment = null;
		}
	}

	// Check if attachment can be previewed
	function canPreview(contentType: string): boolean {
		return contentType.startsWith('image/') ||
			contentType.includes('pdf') ||
			contentType.startsWith('text/') ||
			contentType.includes('json') ||
			contentType.includes('xml') ||
			contentType.startsWith('video/') ||
			contentType.startsWith('audio/');
	}

	async function previewOrDownloadAttachment(attachment: EmailAttachment) {
		downloadingAttachment = attachment.id;
		try {
			const result = await mail.getAttachmentContent({
				username: credentials.username,
				password: credentials.password,
				attachment_id: attachment.id
			});

			// Check if can preview
			if (canPreview(result.content_type)) {
				previewAttachment = {
					name: result.name || attachment.name,
					contentType: result.content_type,
					content: result.content
				};
			} else {
				// Download directly
				const byteCharacters = atob(result.content);
				const byteNumbers = new Array(byteCharacters.length);
				for (let i = 0; i < byteCharacters.length; i++) {
					byteNumbers[i] = byteCharacters.charCodeAt(i);
				}
				const byteArray = new Uint8Array(byteNumbers);
				const blob = new Blob([byteArray], { type: result.content_type });

				const url = URL.createObjectURL(blob);
				const link = document.createElement('a');
				link.href = url;
				link.download = result.name || attachment.name;
				document.body.appendChild(link);
				link.click();
				document.body.removeChild(link);
				URL.revokeObjectURL(url);
			}
		} catch (err) {
			console.error('Failed to preview attachment:', err);
		} finally {
			downloadingAttachment = null;
		}
	}

	function downloadPreviewedAttachment() {
		if (!previewAttachment) return;

		const byteCharacters = atob(previewAttachment.content);
		const byteNumbers = new Array(byteCharacters.length);
		for (let i = 0; i < byteCharacters.length; i++) {
			byteNumbers[i] = byteCharacters.charCodeAt(i);
		}
		const byteArray = new Uint8Array(byteNumbers);
		const blob = new Blob([byteArray], { type: previewAttachment.contentType });

		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = url;
		link.download = previewAttachment.name;
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
		URL.revokeObjectURL(url);
	}

	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
	}

	function getFileIcon(contentType: string, name: string): string {
		const ext = name.split('.').pop()?.toLowerCase() || '';
		if (contentType.startsWith('image/')) return '🖼️';
		if (contentType.includes('pdf') || ext === 'pdf') return '📄';
		if (contentType.includes('word') || ['doc', 'docx'].includes(ext)) return '📝';
		if (contentType.includes('excel') || contentType.includes('spreadsheet') || ['xls', 'xlsx'].includes(ext)) return '📊';
		if (contentType.includes('powerpoint') || contentType.includes('presentation') || ['ppt', 'pptx'].includes(ext)) return '📊';
		if (contentType.includes('zip') || contentType.includes('archive') || ['zip', 'rar', '7z'].includes(ext)) return '📦';
		if (contentType.startsWith('text/')) return '📃';
		return '📎';
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
				body: composeBody,
				attachments: composeAttachments.map(a => ({ name: a.name, content: a.content }))
			});

			showCompose = false;
			composeTo = '';
			composeCc = '';
			composeSubject = '';
			composeBody = '';
			composeAttachments = [];

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

	function handleAttachmentSelect(event: Event) {
		const input = event.target as HTMLInputElement;
		if (!input.files) return;

		for (const file of Array.from(input.files)) {
			const reader = new FileReader();
			reader.onload = () => {
				const base64 = (reader.result as string).split(',')[1];
				composeAttachments = [...composeAttachments, {
					name: file.name,
					content: base64,
					size: file.size
				}];
			};
			reader.readAsDataURL(file);
		}
		input.value = ''; // Reset input for re-selection
	}

	function removeComposeAttachment(index: number) {
		composeAttachments = composeAttachments.filter((_, i) => i !== index);
	}

	function logout() {
		if (browser) {
			// Clear credentials from both storages
			localStorage.removeItem('ews_credentials');
			sessionStorage.removeItem('ews_credentials');
			sessionStorage.removeItem('mail_credentials');
		}
		if (refreshInterval) {
			clearInterval(refreshInterval);
			refreshInterval = null;
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

	function getFolderPriority(name: string): number {
		const lower = name.toLowerCase();
		if (lower === 'inbox' || lower === 'входящие') return 0;
		if (lower.includes('sent') || lower.includes('отправленные')) return 1;
		if (lower.includes('draft') || lower.includes('черновик')) return 2;
		if (lower.includes('deleted') || lower.includes('удаленные')) return 3;
		if (lower.includes('spam') || lower.includes('junk') || lower.includes('нежелательн')) return 4;
		if (lower.includes('archive') || lower.includes('архив')) return 5;
		return 10; // Other folders at the end
	}

	let sortedFolders = $derived(
		[...folders].sort((a, b) => getFolderPriority(a.display_name) - getFolderPriority(b.display_name))
	);

	function getPersonDisplay(person?: EmailPerson): string {
		if (!person) return 'Неизвестный';
		return person.name || person.email || 'Неизвестный';
	}

	function getAvatarColor(name: string): string {
		const colors = [
			'bg-red-500', 'bg-orange-500', 'bg-amber-500', 'bg-yellow-500',
			'bg-lime-500', 'bg-green-500', 'bg-emerald-500', 'bg-teal-500',
			'bg-cyan-500', 'bg-sky-500', 'bg-blue-500', 'bg-indigo-500',
			'bg-violet-500', 'bg-purple-500', 'bg-fuchsia-500', 'bg-pink-500'
		];
		let hash = 0;
		for (let i = 0; i < name.length; i++) {
			hash = name.charCodeAt(i) + ((hash << 5) - hash);
		}
		return colors[Math.abs(hash) % colors.length];
	}

	function getInitials(name: string): string {
		const parts = name.trim().split(/\s+/);
		if (parts.length >= 2) {
			return (parts[0][0] + parts[1][0]).toUpperCase();
		}
		return name.substring(0, 2).toUpperCase();
	}

	// Meeting invite detection
	function isMeetingInvite(email: EmailMessage): boolean {
		// IPM.Schedule.Meeting.Request - приглашение на встречу
		// IPM.Schedule.Meeting.Resp.* - ответы на приглашение
		// IPM.Schedule.Meeting.Canceled - отмена встречи
		return email.item_class?.includes('IPM.Schedule.Meeting') || false;
	}

	function getMeetingInviteType(email: EmailMessage): 'request' | 'response' | 'cancel' | null {
		if (!email.item_class) return null;
		if (email.item_class.includes('Meeting.Request')) return 'request';
		if (email.item_class.includes('Meeting.Resp')) return 'response';
		if (email.item_class.includes('Meeting.Canceled')) return 'cancel';
		return null;
	}

	// Try to extract meeting date from email content
	function extractMeetingDate(email: EmailMessage): string | null {
		const content = email.body || email.body_preview || email.subject || '';

		// Try common date patterns
		const patterns = [
			// ISO format: 2024-01-20
			/(\d{4}-\d{2}-\d{2})/,
			// Russian format: 20.01.2024 or 20/01/2024
			/(\d{1,2})[./](\d{1,2})[./](\d{4})/,
			// English format: Jan 20, 2024 or January 20, 2024
			/((?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*)\s+(\d{1,2}),?\s+(\d{4})/i,
			// Russian month names: 20 января 2024
			/(\d{1,2})\s+(январ[яь]|феврал[яь]|март[а]?|апрел[яь]|ма[яй]|июн[яь]|июл[яь]|август[а]?|сентябр[яь]|октябр[яь]|ноябр[яь]|декабр[яь])\s+(\d{4})/i
		];

		for (const pattern of patterns) {
			const match = content.match(pattern);
			if (match) {
				try {
					// ISO format
					if (pattern === patterns[0]) {
						const date = new Date(match[1]);
						if (!isNaN(date.getTime())) {
							return match[1];
						}
					}
					// Russian/European format: DD.MM.YYYY
					else if (pattern === patterns[1]) {
						const day = parseInt(match[1]);
						const month = parseInt(match[2]) - 1;
						const year = parseInt(match[3]);
						const date = new Date(year, month, day);
						if (!isNaN(date.getTime())) {
							return date.toISOString().split('T')[0];
						}
					}
					// English month format
					else if (pattern === patterns[2]) {
						const date = new Date(`${match[1]} ${match[2]}, ${match[3]}`);
						if (!isNaN(date.getTime())) {
							return date.toISOString().split('T')[0];
						}
					}
					// Russian month names
					else if (pattern === patterns[3]) {
						const monthMap: Record<string, number> = {
							'январ': 0, 'феврал': 1, 'март': 2, 'апрел': 3,
							'ма': 4, 'июн': 5, 'июл': 6, 'август': 7,
							'сентябр': 8, 'октябр': 9, 'ноябр': 10, 'декабр': 11
						};
						const monthKey = Object.keys(monthMap).find(k => match[2].toLowerCase().startsWith(k));
						if (monthKey !== undefined) {
							const day = parseInt(match[1]);
							const month = monthMap[monthKey];
							const year = parseInt(match[3]);
							const date = new Date(year, month, day);
							if (!isNaN(date.getTime())) {
								return date.toISOString().split('T')[0];
							}
						}
					}
				} catch {
					continue;
				}
			}
		}

		// Fallback: use email received date
		const receivedDate = new Date(email.received_at);
		if (!isNaN(receivedDate.getTime())) {
			return receivedDate.toISOString().split('T')[0];
		}

		return null;
	}

	function getCalendarLink(email: EmailMessage): string {
		const date = extractMeetingDate(email);
		if (date) {
			return `/calendar?date=${date}`;
		}
		return '/calendar';
	}

	async function respondToMeeting(response: 'Accept' | 'Decline' | 'Tentative') {
		if (!selectedEmail) return;

		respondingToMeeting = response;
		meetingResponseSuccess = null;

		try {
			await mail.respondToMeeting({
				username: credentials.username,
				password: credentials.password,
				item_id: selectedEmail.id,
				change_key: selectedEmail.change_key,
				response
			});

			const responseText = response === 'Accept' ? 'принято' : response === 'Decline' ? 'отклонено' : 'под вопросом';
			meetingResponseSuccess = `Приглашение ${responseText}`;

			// Remove the meeting request from the list after successful response
			setTimeout(() => {
				if (selectedEmail) {
					emails = emails.filter(e => e.id !== selectedEmail?.id);
					selectedEmail = null;
					showEmailModal = false;
					meetingResponseSuccess = null;
				}
			}, 2000);
		} catch (err) {
			console.error('Failed to respond to meeting:', err);
			error = err instanceof Error ? err.message : 'Не удалось ответить на приглашение';
		} finally {
			respondingToMeeting = null;
		}
	}

	// Thread type for grouping emails
	interface EmailThread {
		conversationId: string;
		emails: EmailMessage[];
		latestDate: string;
		hasUnread: boolean;
		subject: string;
	}

	// Normalize subject for grouping (remove Re:, Fwd:, FW:, etc.)
	function normalizeSubject(subject: string): string {
		return subject
			.replace(/^(Re|Fwd|FW|RE|AW|Ответ|Пересылка):\s*/gi, '')
			.replace(/^(Re|Fwd|FW|RE|AW|Ответ|Пересылка)\[\d+\]:\s*/gi, '')
			.trim()
			.toLowerCase();
	}

	function toggleThread(conversationId: string) {
		const newSet = new Set(expandedThreads);
		if (newSet.has(conversationId)) {
			newSet.delete(conversationId);
		} else {
			newSet.add(conversationId);
		}
		expandedThreads = newSet;
	}

	// Handle thread header click - expand/collapse and select latest email
	function handleThreadClick(thread: EmailThread) {
		const latestEmail = thread.emails[thread.emails.length - 1];

		if (thread.emails.length === 1) {
			// Single email - just select it
			selectEmail(latestEmail);
		} else {
			// Multi-email thread
			const isExpanded = expandedThreads.has(thread.conversationId);
			if (!isExpanded) {
				// Expand the thread and select latest email
				const newSet = new Set(expandedThreads);
				newSet.add(thread.conversationId);
				expandedThreads = newSet;
			}
			// Always select the latest email when clicking thread header
			selectEmail(latestEmail);
		}
	}

	// Get current thread for selected email
	function getCurrentThread(): EmailThread | null {
		if (!selectedEmail || !showThreaded) return null;
		const threads = groupedEmails();
		if (!threads) return null;
		return threads.find(t => t.emails.some(e => e.id === selectedEmail?.id)) || null;
	}

	// Navigate within thread
	function navigateThread(direction: 'first' | 'prev' | 'next' | 'last') {
		const thread = getCurrentThread();
		if (!thread || !selectedEmail) return;

		const currentIdx = thread.emails.findIndex(e => e.id === selectedEmail.id);
		let newIdx = currentIdx;

		switch (direction) {
			case 'first': newIdx = 0; break;
			case 'prev': newIdx = Math.max(0, currentIdx - 1); break;
			case 'next': newIdx = Math.min(thread.emails.length - 1, currentIdx + 1); break;
			case 'last': newIdx = thread.emails.length - 1; break;
		}

		if (newIdx !== currentIdx) {
			selectEmail(thread.emails[newIdx]);
		}
	}

	// Check if selected email belongs to this thread
	function isThreadSelected(thread: EmailThread): boolean {
		if (!selectedEmail) return false;
		return thread.emails.some(e => e.id === selectedEmail?.id);
	}

	let filteredEmails = $derived(() => {
		let result = emails;

		// Filter by unread if enabled
		if (showOnlyUnread) {
			result = result.filter(e => !e.is_read);
		}

		// Filter by search query
		if (searchQuery) {
			const query = searchQuery.toLowerCase();
			result = result.filter(e =>
				e.subject.toLowerCase().includes(query) ||
				(e.from?.name || '').toLowerCase().includes(query) ||
				(e.from?.email || '').toLowerCase().includes(query)
			);
		}

		return result;
	});

	let groupedEmails = $derived(() => {
		if (!showThreaded) return null;

		const filtered = filteredEmails();
		const threads = new Map<string, EmailThread>();
		// Secondary map for subject-based grouping when no conversation_id
		const subjectToConvId = new Map<string, string>();

		for (const email of filtered) {
			let convId: string;

			if (email.conversation_id) {
				// Use EWS conversation_id when available
				convId = email.conversation_id;
				// Store normalized subject -> convId mapping for emails without conv_id
				const normalizedSubj = normalizeSubject(email.subject);
				if (!subjectToConvId.has(normalizedSubj)) {
					subjectToConvId.set(normalizedSubj, convId);
				}
			} else {
				// Fallback: try to find matching thread by normalized subject
				const normalizedSubj = normalizeSubject(email.subject);
				if (subjectToConvId.has(normalizedSubj)) {
					convId = subjectToConvId.get(normalizedSubj)!;
				} else {
					// Create new thread based on normalized subject
					convId = `subj_${normalizedSubj}`;
					subjectToConvId.set(normalizedSubj, convId);
				}
			}

			if (!threads.has(convId)) {
				threads.set(convId, {
					conversationId: convId,
					emails: [],
					latestDate: email.received_at,
					hasUnread: false,
					subject: email.subject
				});
			}
			const thread = threads.get(convId)!;
			thread.emails.push(email);
			if (!email.is_read) thread.hasUnread = true;
			if (email.received_at > thread.latestDate) {
				thread.latestDate = email.received_at;
				// Update thread subject to latest email's subject
				thread.subject = email.subject;
			}
		}

		// Sort threads by latest date
		const sortedThreads = Array.from(threads.values())
			.filter(t => t.emails.length > 0)
			.sort((a, b) => b.latestDate.localeCompare(a.latestDate));

		// Sort emails within each thread by date (oldest first)
		for (const thread of sortedThreads) {
			thread.emails.sort((a, b) => a.received_at.localeCompare(b.received_at));
		}

		return sortedThreads;
	});

	function openEmailModal() {
		if (selectedEmail) {
			showEmailModal = true;
		}
	}

	// Open email directly in modal (used for double-click)
	async function openEmailInModal(email: EmailMessage) {
		await selectEmail(email);
		showEmailModal = true;
	}

	function replyToEmail(mode: 'reply' | 'replyAll' = 'reply') {
		if (!selectedEmail) return;
		composeMode = mode;

		// Set recipient
		if (mode === 'reply') {
			composeTo = selectedEmail.from?.email || '';
		} else {
			// Reply all - include sender and all recipients except self
			const recipients = [selectedEmail.from?.email || ''];
			if (selectedEmail.to) {
				recipients.push(...selectedEmail.to.filter(p => p.email !== credentials.username).map(p => p.email));
			}
			composeTo = recipients.filter(Boolean).join(', ');
			if (selectedEmail.cc) {
				composeCc = selectedEmail.cc.filter(p => p.email !== credentials.username).map(p => p.email).join(', ');
			}
		}

		// Set subject with Re: prefix
		composeSubject = selectedEmail.subject.startsWith('Re:')
			? selectedEmail.subject
			: `Re: ${selectedEmail.subject}`;

		// Quote original message with HTML formatting
		const originalDate = new Date(selectedEmail.received_at).toLocaleString('ru-RU');
		const originalFrom = selectedEmail.from?.name || selectedEmail.from?.email || 'Неизвестный';
		composeBody = `<p><br></p><hr><p><strong>${originalDate}, ${originalFrom} написал(а):</strong></p><blockquote>${selectedEmail.body || ''}</blockquote>`;

		composeAttachments = [];
		showCompose = true;
	}

	function forwardEmail() {
		if (!selectedEmail) return;
		composeMode = 'forward';
		composeTo = '';
		composeCc = '';

		// Set subject with Fwd: prefix
		composeSubject = selectedEmail.subject.startsWith('Fwd:') || selectedEmail.subject.startsWith('FW:')
			? selectedEmail.subject
			: `Fwd: ${selectedEmail.subject}`;

		// Include original message with HTML formatting
		const originalDate = new Date(selectedEmail.received_at).toLocaleString('ru-RU');
		const originalFrom = selectedEmail.from?.name || selectedEmail.from?.email || 'Неизвестный';
		const originalTo = selectedEmail.to?.map(p => p.name || p.email).join(', ') || '';
		composeBody = `<p><br></p><hr><p><strong>---------- Пересылаемое сообщение ----------</strong></p>
<p><strong>От:</strong> ${originalFrom}<br>
<strong>Дата:</strong> ${originalDate}<br>
<strong>Тема:</strong> ${selectedEmail.subject}<br>
<strong>Кому:</strong> ${originalTo}</p>
<div>${selectedEmail.body || ''}</div>`;

		composeAttachments = [];
		showCompose = true;
	}

	function stripHtml(html: string): string {
		const doc = new DOMParser().parseFromString(html, 'text/html');
		return doc.body.textContent || '';
	}
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
				<div class="mb-4">
					<label class="block text-sm font-medium text-gray-700 mb-1">Пароль</label>
					<input
						type="password"
						bind:value={credentials.password}
						class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
					/>
				</div>
				<div class="mb-6 flex items-center">
					<input
						type="checkbox"
						id="rememberMe"
						bind:checked={rememberMe}
						class="w-4 h-4 text-ekf-red border-gray-300 rounded focus:ring-ekf-red"
					/>
					<label for="rememberMe" class="ml-2 text-sm text-gray-600">Запомнить меня</label>
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
		<div class="{sidebarCollapsed ? 'w-16' : 'w-60'} border-r border-gray-200 flex flex-col bg-gray-50 transition-all duration-200">
			<div class="p-3 flex {sidebarCollapsed ? 'justify-center' : 'gap-2'}">
				<button
					onclick={() => { composeMode = 'new'; composeTo = ''; composeCc = ''; composeSubject = ''; composeBody = ''; composeAttachments = []; showCompose = true; }}
					class="{sidebarCollapsed ? 'w-10 h-10 p-0 justify-center' : 'flex-1 py-2 px-4'} bg-ekf-red text-white rounded-lg font-medium hover:bg-red-700 transition-colors flex items-center gap-2"
					title={sidebarCollapsed ? 'Написать' : ''}
				>
					<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					{#if !sidebarCollapsed}
						<span>Написать</span>
					{/if}
				</button>
			</div>

			<nav class="flex-1 overflow-y-auto px-2">
				{#each sortedFolders as folder}
					<button
						onclick={() => selectFolder(folder)}
						class="w-full {sidebarCollapsed ? 'px-2 justify-center' : 'px-3'} py-2 flex items-center gap-3 rounded-lg text-left text-sm transition-colors mb-1 relative
							{selectedFolder?.id === folder.id ? 'bg-ekf-red/10 text-ekf-red' : 'text-gray-700 hover:bg-gray-100'}"
						title={sidebarCollapsed ? folder.display_name : ''}
					>
						<div class="relative flex-shrink-0">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
							{#if sidebarCollapsed && folder.unread_count > 0}
								<span class="absolute -top-2 -right-2 min-w-[16px] h-4 text-[10px] bg-ekf-red text-white rounded-full flex items-center justify-center px-1">{folder.unread_count > 9 ? '9+' : folder.unread_count}</span>
							{/if}
						</div>
						{#if !sidebarCollapsed}
							<span class="flex-1 truncate">{folder.display_name}</span>
							{#if folder.unread_count > 0}
								<span class="text-xs bg-ekf-red text-white px-1.5 py-0.5 rounded-full">{folder.unread_count}</span>
							{/if}
						{/if}
					</button>
				{/each}
			</nav>

			<!-- Collapse toggle button -->
			<div class="p-2 border-t border-gray-200">
				<button
					onclick={() => sidebarCollapsed = !sidebarCollapsed}
					class="w-full p-2 flex items-center justify-center gap-2 text-gray-500 hover:bg-gray-100 rounded-lg transition-colors"
					title={sidebarCollapsed ? 'Развернуть' : 'Свернуть'}
				>
					<svg class="w-5 h-5 transition-transform {sidebarCollapsed ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
					</svg>
					{#if !sidebarCollapsed}
						<span class="text-sm">Свернуть</span>
					{/if}
				</button>
			</div>
		</div>

		<!-- Email list -->
		<div class="w-80 border-r border-gray-200 flex flex-col">
			<div class="p-3 border-b border-gray-200 space-y-2">
				<div class="flex gap-2">
					<div class="relative flex-1">
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
					<button
						onclick={() => selectedFolder && selectFolder(selectedFolder)}
						disabled={loadingEmails}
						class="p-2 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors disabled:opacity-50"
						title="Обновить"
					>
						<svg class="w-5 h-5 text-gray-600 {loadingEmails ? 'animate-spin' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
						</svg>
					</button>
				</div>
				<div class="flex gap-1">
					<button
						onclick={() => showOnlyUnread = !showOnlyUnread}
						class="flex-1 flex items-center justify-center gap-1 px-2 py-1.5 text-xs rounded-lg transition-colors
							{showOnlyUnread ? 'bg-ekf-red text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
						title={showOnlyUnread ? 'Показать все' : 'Только непрочитанные'}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2 2v10a2 2 0 002 2z" />
						</svg>
						Непрочит.
					</button>
					<button
						onclick={() => showThreaded = !showThreaded}
						class="flex-1 flex items-center justify-center gap-1 px-2 py-1.5 text-xs rounded-lg transition-colors
							{showThreaded ? 'bg-ekf-red text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
						title={showThreaded ? 'Список' : 'По цепочкам'}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
						</svg>
						Цепочки
					</button>
				</div>
			</div>

			<div class="flex-1 overflow-y-auto">
				{#if loadingEmails}
					<div class="flex items-center justify-center h-32">
						<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-ekf-red"></div>
					</div>
				{:else if filteredEmails().length === 0}
					<div class="text-center py-12 text-gray-500 text-sm">
						{showOnlyUnread ? 'Нет непрочитанных писем' : 'Нет писем'}
					</div>
				{:else if showThreaded && groupedEmails()}
					<!-- Threaded view -->
					{#each groupedEmails() as thread}
						{@const isExpanded = expandedThreads.has(thread.conversationId)}
						{@const latestEmail = thread.emails[thread.emails.length - 1]}
						{@const senderName = getPersonDisplay(latestEmail.from)}
						{@const senderPhoto = getEmployeePhoto(latestEmail.from?.email)}
						<div class="border-b border-gray-100 relative">
							<!-- Thread header -->
							<button
								onclick={() => handleThreadClick(thread)}
								ondblclick={() => openEmailInModal(latestEmail)}
								class="w-full px-4 py-3 {thread.emails.length > 1 ? 'pr-10' : ''} text-left hover:bg-gray-50 transition-colors
									{isThreadSelected(thread) ? 'bg-ekf-red/5' : ''}
									{thread.hasUnread ? 'bg-blue-50/50' : ''}"
							>
								<div class="flex items-start gap-3">
									<!-- Avatar -->
									<div class="w-10 h-10 rounded-full {senderPhoto ? '' : getAvatarColor(senderName)} flex items-center justify-center flex-shrink-0 relative overflow-hidden">
										{#if senderPhoto}
											<img src="data:image/jpeg;base64,{senderPhoto}" alt="" class="w-full h-full object-cover" />
										{:else}
											<span class="text-white text-sm font-medium">{getInitials(senderName)}</span>
										{/if}
										{#if thread.hasUnread}
											<div class="absolute -top-0.5 -right-0.5 w-3 h-3 bg-ekf-red rounded-full border-2 border-white"></div>
										{/if}
									</div>
									<div class="flex-1 min-w-0">
										<div class="flex items-center gap-2 mb-1">
											<span class="text-sm {thread.hasUnread ? 'font-semibold' : ''} text-gray-900 truncate">
												{senderName}
											</span>
											{#if thread.emails.length > 1}
												<span class="px-1.5 py-0.5 bg-gray-200 text-gray-600 text-xs rounded-full">
													{thread.emails.length}
												</span>
											{/if}
											<span class="text-xs text-gray-400 flex-shrink-0 ml-auto">
												{formatDate(thread.latestDate)}
											</span>
										</div>
										<div class="text-sm {thread.hasUnread ? 'font-medium' : ''} text-gray-800 truncate">{thread.subject}</div>
										{#if latestEmail.body_preview}
											<div class="text-xs text-gray-500 truncate mt-0.5">{latestEmail.body_preview}</div>
										{/if}
									</div>
									{#if thread.emails.some(e => e.has_attachments)}
										<svg class="w-4 h-4 text-gray-400 flex-shrink-0 mt-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
										</svg>
									{/if}
								</div>
							</button>
							{#if thread.emails.length > 1}
								<!-- Separate toggle button for expand/collapse -->
								<button
									onclick={(e) => { e.stopPropagation(); toggleThread(thread.conversationId); }}
									class="absolute right-2 top-1/2 -translate-y-1/2 p-1 hover:bg-gray-200 rounded transition-colors"
									title={isExpanded ? 'Свернуть' : 'Развернуть'}
								>
									<svg class="w-4 h-4 text-gray-400 transition-transform {isExpanded ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
									</svg>
								</button>
							{/if}
							<!-- Expanded thread emails -->
							{#if isExpanded && thread.emails.length > 1}
								<div class="bg-gray-50 border-t border-gray-200">
									{#each thread.emails as email, idx}
										{@const emailSender = getPersonDisplay(email.from)}
										<button
											onclick={() => selectEmail(email)}
											ondblclick={() => openEmailInModal(email)}
											class="w-full pl-12 pr-4 py-2 text-left hover:bg-gray-100 transition-colors border-b border-gray-100 last:border-b-0
												{selectedEmail?.id === email.id ? 'bg-ekf-red/10' : ''}
												{!email.is_read ? 'bg-blue-50/30' : ''}"
										>
											<div class="flex items-center gap-2">
												<div class="w-6 h-6 rounded-full {getAvatarColor(emailSender)} flex items-center justify-center flex-shrink-0">
													<span class="text-white text-xs">{getInitials(emailSender).charAt(0)}</span>
												</div>
												<span class="text-xs {!email.is_read ? 'font-semibold' : ''} text-gray-700 truncate flex-1">
													{emailSender}
												</span>
												<span class="text-xs text-gray-400 flex-shrink-0">
													{formatDate(email.received_at)}
												</span>
												{#if email.has_attachments}
													<svg class="w-3 h-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
													</svg>
												{/if}
											</div>
										</button>
									{/each}
								</div>
							{/if}
						</div>
					{/each}
				{:else}
					<!-- Flat list view -->
					{#each filteredEmails() as email}
						{@const senderName = getPersonDisplay(email.from)}
						{@const senderPhoto = getEmployeePhoto(email.from?.email)}
						<button
							onclick={() => selectEmail(email)}
							ondblclick={() => openEmailInModal(email)}
							class="w-full px-4 py-3 text-left border-b border-gray-100 hover:bg-gray-50 transition-colors
								{selectedEmail?.id === email.id ? 'bg-ekf-red/5' : ''}
								{!email.is_read ? 'bg-blue-50/50' : ''}"
						>
							<div class="flex items-start gap-3">
								<!-- Avatar -->
								<div class="w-10 h-10 rounded-full {senderPhoto ? '' : getAvatarColor(senderName)} flex items-center justify-center flex-shrink-0 relative overflow-hidden">
									{#if senderPhoto}
										<img src="data:image/jpeg;base64,{senderPhoto}" alt="" class="w-full h-full object-cover" />
									{:else}
										<span class="text-white text-sm font-medium">{getInitials(senderName)}</span>
									{/if}
									{#if !email.is_read}
										<div class="absolute -top-0.5 -right-0.5 w-3 h-3 bg-ekf-red rounded-full border-2 border-white"></div>
									{/if}
								</div>
								<div class="flex-1 min-w-0">
									<div class="flex items-center gap-2 mb-1">
										<span class="text-sm {!email.is_read ? 'font-semibold' : ''} text-gray-900 truncate">
											{senderName}
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
								<div class="flex items-center gap-1 flex-shrink-0 mt-1">
									{#if isMeetingInvite(email)}
										<svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" title="Приглашение на встречу">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
										</svg>
									{/if}
									{#if email.has_attachments}
										<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
										</svg>
									{/if}
								</div>
							</div>
						</button>
					{/each}
				{/if}
			</div>
		</div>

		<!-- Email content -->
		<div class="flex-1 flex flex-col bg-gray-50">
			{#if selectedEmail}
				{@const selectedPhoto = getEmployeePhoto(selectedEmail.from?.email)}
				<div class="bg-white border-b border-gray-200 px-6 py-4">
					<div class="flex items-start justify-between mb-3">
						<button onclick={openEmailModal} class="text-lg font-medium text-gray-900 hover:text-ekf-red transition-colors text-left flex items-center gap-2">
							{selectedEmail.subject}
							<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
							</svg>
						</button>
						<div class="flex items-center gap-1">
							<!-- Thread navigation -->
							{#if showThreaded}
								{@const currentThread = getCurrentThread()}
								{#if currentThread && currentThread.emails.length > 1}
									{@const currentIdx = currentThread.emails.findIndex(e => e.id === selectedEmail?.id)}
								<div class="flex items-center gap-0.5 mr-2 px-2 py-1 bg-gray-100 rounded-lg">
									<button
										onclick={() => navigateThread('first')}
										disabled={currentIdx === 0}
										class="p-1 hover:bg-gray-200 rounded disabled:opacity-30 disabled:cursor-not-allowed"
										title="Первое письмо (оригинал)"
									>
										<svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
										</svg>
									</button>
									<button
										onclick={() => navigateThread('prev')}
										disabled={currentIdx === 0}
										class="p-1 hover:bg-gray-200 rounded disabled:opacity-30 disabled:cursor-not-allowed"
										title="Предыдущее"
									>
										<svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
										</svg>
									</button>
									<span class="text-xs text-gray-500 px-1">{currentIdx + 1}/{currentThread.emails.length}</span>
									<button
										onclick={() => navigateThread('next')}
										disabled={currentIdx === currentThread.emails.length - 1}
										class="p-1 hover:bg-gray-200 rounded disabled:opacity-30 disabled:cursor-not-allowed"
										title="Следующее"
									>
										<svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
										</svg>
									</button>
									<button
										onclick={() => navigateThread('last')}
										disabled={currentIdx === currentThread.emails.length - 1}
										class="p-1 hover:bg-gray-200 rounded disabled:opacity-30 disabled:cursor-not-allowed"
										title="Последнее письмо"
									>
										<svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
										</svg>
									</button>
								</div>
								{/if}
							{/if}
							<button onclick={() => replyToEmail('reply')} class="p-2 hover:bg-gray-100 rounded-lg transition-colors" title="Ответить">
								<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
								</svg>
							</button>
							<button onclick={() => replyToEmail('replyAll')} class="p-2 hover:bg-gray-100 rounded-lg transition-colors" title="Ответить всем">
								<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h7a8 8 0 018 8v2M3 10l5 5m-5-5l5-5M10 10h7a5 5 0 015 5v2" />
								</svg>
							</button>
							<button onclick={forwardEmail} class="p-2 hover:bg-gray-100 rounded-lg transition-colors" title="Переслать">
								<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
								</svg>
							</button>
							<button onclick={() => window.open(`/mail/${selectedEmail?.id}`, '_blank')} class="p-2 hover:bg-gray-100 rounded-lg transition-colors" title="Открыть в новом окне">
								<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
								</svg>
							</button>
							<div class="w-px h-5 bg-gray-200 mx-1"></div>
							<button onclick={() => deleteEmail(selectedEmail!)} class="p-2 hover:bg-red-50 rounded-lg transition-colors" title="Удалить">
								<svg class="w-5 h-5 text-gray-500 hover:text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
								</svg>
							</button>
						</div>
					</div>
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-full {selectedPhoto ? '' : 'bg-ekf-red/10'} flex items-center justify-center overflow-hidden">
							{#if selectedPhoto}
								<img src="data:image/jpeg;base64,{selectedPhoto}" alt="" class="w-full h-full object-cover" />
							{:else}
								<span class="text-ekf-red font-medium">
									{(selectedEmail.from?.name || selectedEmail.from?.email || '?').charAt(0).toUpperCase()}
								</span>
							{/if}
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

				<!-- Meeting Invite Banner -->
				{#if isMeetingInvite(selectedEmail)}
					<div class="mx-4 mb-4 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl border border-blue-200">
						<div class="flex items-center gap-3">
							<div class="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
								<svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
								</svg>
							</div>
							<div class="flex-1">
								{#if getMeetingInviteType(selectedEmail) === 'request'}
									<div class="text-sm font-medium text-blue-900">Приглашение на встречу</div>
									<div class="text-xs text-blue-700">Вас приглашают на встречу</div>
								{:else if getMeetingInviteType(selectedEmail) === 'cancel'}
									<div class="text-sm font-medium text-red-900">Встреча отменена</div>
									<div class="text-xs text-red-700">Организатор отменил встречу</div>
								{:else}
									<div class="text-sm font-medium text-blue-900">Ответ на приглашение</div>
									<div class="text-xs text-blue-700">Ответ от участника</div>
								{/if}
							</div>
							{#if getMeetingInviteType(selectedEmail) === 'request'}
								{#if meetingResponseSuccess}
									<div class="px-3 py-1.5 bg-green-100 text-green-700 text-xs font-medium rounded-lg">
										{meetingResponseSuccess}
									</div>
								{:else}
									<div class="flex items-center gap-2">
										<button
											onclick={() => respondToMeeting('Accept')}
											disabled={respondingToMeeting !== null}
											class="px-3 py-1.5 bg-green-500 text-white text-xs font-medium rounded-lg hover:bg-green-600 transition-colors disabled:opacity-50 flex items-center gap-1"
										>
											{#if respondingToMeeting === 'Accept'}
												<div class="animate-spin rounded-full h-3 w-3 border-b-2 border-white"></div>
											{/if}
											Принять
										</button>
										<button
											onclick={() => respondToMeeting('Tentative')}
											disabled={respondingToMeeting !== null}
											class="px-3 py-1.5 bg-yellow-500 text-white text-xs font-medium rounded-lg hover:bg-yellow-600 transition-colors disabled:opacity-50 flex items-center gap-1"
										>
											{#if respondingToMeeting === 'Tentative'}
												<div class="animate-spin rounded-full h-3 w-3 border-b-2 border-white"></div>
											{/if}
											Под вопросом
										</button>
										<button
											onclick={() => respondToMeeting('Decline')}
											disabled={respondingToMeeting !== null}
											class="px-3 py-1.5 bg-red-500 text-white text-xs font-medium rounded-lg hover:bg-red-600 transition-colors disabled:opacity-50 flex items-center gap-1"
										>
											{#if respondingToMeeting === 'Decline'}
												<div class="animate-spin rounded-full h-3 w-3 border-b-2 border-white"></div>
											{/if}
											Отклонить
										</button>
									</div>
								{/if}
							{/if}
							<a
								href={getCalendarLink(selectedEmail)}
								class="flex items-center gap-1 px-3 py-1.5 bg-blue-500 text-white text-xs font-medium rounded-lg hover:bg-blue-600 transition-colors"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
								</svg>
								Календарь
							</a>
						</div>
					</div>
				{/if}

				<div class="flex-1 overflow-y-auto p-6">
					<div class="bg-white rounded-lg p-6 shadow-sm">
						{#if loadingBody}
							<div class="flex items-center justify-center py-8">
								<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-ekf-red"></div>
								<span class="ml-3 text-gray-500 text-sm">Загрузка письма...</span>
							</div>
						{:else if bodyError}
							<div class="text-center py-8">
								<svg class="w-12 h-12 mx-auto text-red-300 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
								</svg>
								<p class="text-red-500">{bodyError}</p>
							</div>
						{:else if selectedEmail.body}
							<div class="prose max-w-none">
								{@html sanitizeEmailHtml(selectedEmail.body)}
							</div>
						{:else}
							<p class="text-gray-500">Нет содержимого</p>
						{/if}

						<!-- Attachments -->
						{#if selectedEmail.has_attachments}
							<div class="mt-4 pt-4 border-t border-gray-200">
								<h4 class="text-sm font-medium text-gray-700 mb-2 flex items-center gap-2">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
									</svg>
									Вложения
									{#if loadingAttachments}
										<span class="animate-spin inline-block w-4 h-4 border-2 border-gray-300 border-t-ekf-red rounded-full"></span>
									{/if}
								</h4>
								{#if attachmentError}
									<div class="p-2 bg-red-50 border border-red-200 rounded-lg">
										<p class="text-sm text-red-600">{attachmentError}</p>
									</div>
								{:else if attachments.length > 0}
									<div class="grid gap-2">
										{#each attachments.filter(a => !a.is_inline) as attachment}
											<button
												onclick={() => previewOrDownloadAttachment(attachment)}
												disabled={downloadingAttachment === attachment.id}
												class="flex items-center gap-3 p-2 rounded-lg border border-gray-200 hover:bg-gray-50 text-left transition-colors disabled:opacity-50"
											>
												<span class="text-2xl">{getFileIcon(attachment.content_type, attachment.name)}</span>
												<div class="flex-1 min-w-0">
													<div class="text-sm font-medium text-gray-900 truncate">{attachment.name}</div>
													<div class="text-xs text-gray-500">{formatFileSize(attachment.size)}</div>
												</div>
												{#if downloadingAttachment === attachment.id}
													<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-ekf-red"></div>
												{:else}
													<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
													</svg>
												{/if}
											</button>
										{/each}
									</div>
								{:else if !loadingAttachments}
									<p class="text-sm text-gray-500">Вложения не найдены</p>
								{/if}
							</div>
						{/if}
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

	<!-- Full Email Modal -->
	{#if showEmailModal && selectedEmail}
		{@const modalPhoto = getEmployeePhoto(selectedEmail.from?.email)}
		<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
			<div class="bg-white rounded-xl shadow-xl w-full max-w-4xl max-h-[90vh] flex flex-col m-4">
				<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
					<h3 class="text-lg font-semibold text-gray-900 truncate flex-1 mr-4">{selectedEmail.subject}</h3>
					<div class="flex items-center gap-2">
						<button onclick={() => replyToEmail('reply')} class="p-2 hover:bg-gray-100 rounded-lg" title="Ответить">
							<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
							</svg>
						</button>
						<button onclick={() => replyToEmail('replyAll')} class="p-2 hover:bg-gray-100 rounded-lg" title="Ответить всем">
							<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h7a8 8 0 018 8v2M3 10l5 5m-5-5l5-5M10 10h7a5 5 0 015 5v2" />
							</svg>
						</button>
						<button onclick={forwardEmail} class="p-2 hover:bg-gray-100 rounded-lg" title="Переслать">
							<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
							</svg>
						</button>
						<button onclick={() => window.open(`/mail/${selectedEmail?.id}`, '_blank')} class="p-2 hover:bg-gray-100 rounded-lg" title="Открыть в новом окне">
							<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
							</svg>
						</button>
						<button onclick={() => { deleteEmail(selectedEmail!); showEmailModal = false; }} class="p-2 hover:bg-red-50 rounded-lg" title="Удалить">
							<svg class="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
							</svg>
						</button>
						<button onclick={() => showEmailModal = false} class="p-2 hover:bg-gray-100 rounded-lg ml-2">
							<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
							</svg>
						</button>
					</div>
				</div>
				<div class="px-6 py-4 border-b border-gray-200 bg-gray-50">
					<div class="flex items-center gap-4">
						<div class="w-12 h-12 rounded-full {modalPhoto ? '' : getAvatarColor(selectedEmail.from?.name || selectedEmail.from?.email || '')} flex items-center justify-center flex-shrink-0 overflow-hidden">
							{#if modalPhoto}
								<img src="data:image/jpeg;base64,{modalPhoto}" alt="" class="w-full h-full object-cover" />
							{:else}
								<span class="text-white text-lg font-medium">{getInitials(selectedEmail.from?.name || selectedEmail.from?.email || '?')}</span>
							{/if}
						</div>
						<div class="flex-1 min-w-0">
							<div class="font-medium text-gray-900">{getPersonDisplay(selectedEmail.from)}</div>
							<div class="text-sm text-gray-500">
								Кому: {selectedEmail.to?.map(p => getPersonDisplay(p)).join(', ') || 'Вам'}
								{#if selectedEmail.cc && selectedEmail.cc.length > 0}
									<span class="ml-2">| Копия: {selectedEmail.cc.map(p => getPersonDisplay(p)).join(', ')}</span>
								{/if}
							</div>
						</div>
						<div class="text-sm text-gray-400">
							{new Date(selectedEmail.received_at).toLocaleString('ru-RU')}
						</div>
					</div>
					{#if selectedEmail.has_attachments}
						<div class="mt-3 pt-3 border-t border-gray-200">
							<div class="flex items-center gap-2 text-sm text-gray-600">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
								</svg>
								<span>Вложения (загрузка вложений скоро будет доступна)</span>
							</div>
						</div>
					{/if}
				</div>

				<!-- Meeting Invite Banner in Modal -->
				{#if isMeetingInvite(selectedEmail)}
					<div class="mx-6 my-4 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl border border-blue-200">
						<div class="flex items-center gap-3">
							<div class="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center flex-shrink-0">
								<svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
								</svg>
							</div>
							<div class="flex-1 min-w-0">
								{#if getMeetingInviteType(selectedEmail) === 'request'}
									<div class="text-sm font-medium text-blue-900">Приглашение на встречу</div>
									<div class="text-xs text-blue-700">Вас приглашают на встречу</div>
								{:else if getMeetingInviteType(selectedEmail) === 'cancel'}
									<div class="text-sm font-medium text-red-900">Встреча отменена</div>
									<div class="text-xs text-red-700">Организатор отменил встречу</div>
								{:else}
									<div class="text-sm font-medium text-blue-900">Ответ на приглашение</div>
									<div class="text-xs text-blue-700">Ответ от участника</div>
								{/if}
							</div>
							{#if getMeetingInviteType(selectedEmail) === 'request'}
								{#if meetingResponseSuccess}
									<div class="px-3 py-1.5 bg-green-100 text-green-700 text-xs font-medium rounded-lg">
										{meetingResponseSuccess}
									</div>
								{:else}
									<div class="flex items-center gap-2">
										<button
											onclick={() => respondToMeeting('Accept')}
											disabled={respondingToMeeting !== null}
											class="px-3 py-1.5 bg-green-500 text-white text-xs font-medium rounded-lg hover:bg-green-600 transition-colors disabled:opacity-50 flex items-center gap-1"
										>
											{#if respondingToMeeting === 'Accept'}
												<div class="animate-spin rounded-full h-3 w-3 border-b-2 border-white"></div>
											{/if}
											Принять
										</button>
										<button
											onclick={() => respondToMeeting('Tentative')}
											disabled={respondingToMeeting !== null}
											class="px-3 py-1.5 bg-yellow-500 text-white text-xs font-medium rounded-lg hover:bg-yellow-600 transition-colors disabled:opacity-50 flex items-center gap-1"
										>
											{#if respondingToMeeting === 'Tentative'}
												<div class="animate-spin rounded-full h-3 w-3 border-b-2 border-white"></div>
											{/if}
											Под вопросом
										</button>
										<button
											onclick={() => respondToMeeting('Decline')}
											disabled={respondingToMeeting !== null}
											class="px-3 py-1.5 bg-red-500 text-white text-xs font-medium rounded-lg hover:bg-red-600 transition-colors disabled:opacity-50 flex items-center gap-1"
										>
											{#if respondingToMeeting === 'Decline'}
												<div class="animate-spin rounded-full h-3 w-3 border-b-2 border-white"></div>
											{/if}
											Отклонить
										</button>
									</div>
								{/if}
							{/if}
							<a
								href={getCalendarLink(selectedEmail)}
								class="flex items-center gap-1 px-3 py-1.5 bg-blue-500 text-white text-xs font-medium rounded-lg hover:bg-blue-600 transition-colors flex-shrink-0"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
								</svg>
								Календарь
							</a>
						</div>
					</div>
				{/if}

				<div class="flex-1 overflow-y-auto p-6">
					{#if loadingBody}
						<div class="flex items-center justify-center py-8">
							<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-ekf-red"></div>
							<span class="ml-3 text-gray-500 text-sm">Загрузка письма...</span>
						</div>
					{:else if selectedEmail.body}
						<div class="prose max-w-none bg-white rounded-lg p-4">
							{@html sanitizeEmailHtml(selectedEmail.body)}
						</div>
					{:else}
						<p class="text-gray-500 text-center py-8">Нет содержимого</p>
					{/if}

					<!-- Attachments in modal -->
					{#if selectedEmail.has_attachments && attachments.length > 0}
						<div class="mt-6 pt-4 border-t border-gray-200">
							<h4 class="text-sm font-medium text-gray-700 mb-3 flex items-center gap-2">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
								</svg>
								Вложения ({attachments.filter(a => !a.is_inline).length})
							</h4>
							<div class="flex flex-wrap gap-2">
								{#each attachments.filter(a => !a.is_inline) as attachment}
									<button
										onclick={() => previewOrDownloadAttachment(attachment)}
										disabled={downloadingAttachment === attachment.id}
										class="flex items-center gap-2 px-3 py-2 rounded-lg border border-gray-200 hover:bg-gray-50 text-left transition-colors disabled:opacity-50"
									>
										<span class="text-xl">{getFileIcon(attachment.content_type, attachment.name)}</span>
										<div class="min-w-0">
											<div class="text-sm font-medium text-gray-900 truncate max-w-[200px]">{attachment.name}</div>
											<div class="text-xs text-gray-500">{formatFileSize(attachment.size)}</div>
										</div>
										{#if downloadingAttachment === attachment.id}
											<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-ekf-red"></div>
										{/if}
									</button>
								{/each}
							</div>
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}

	<!-- Compose Modal -->
	{#if showCompose}
		<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
			<div class="bg-white rounded-xl shadow-xl w-full max-w-2xl max-h-[90vh] flex flex-col">
				<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
					<h3 class="text-lg font-semibold">
						{#if composeMode === 'reply'}Ответ{:else if composeMode === 'replyAll'}Ответ всем{:else if composeMode === 'forward'}Пересылка{:else}Новое письмо{/if}
					</h3>
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
							<RichTextEditor
								content={composeBody}
								placeholder="Введите текст письма..."
								onchange={(html) => composeBody = html}
							/>
						</div>
						<!-- Attachments section -->
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-2">Вложения</label>
							{#if composeAttachments.length > 0}
								<div class="flex flex-wrap gap-2 mb-2">
									{#each composeAttachments as att, i}
										<div class="flex items-center gap-2 px-3 py-1.5 bg-gray-100 rounded-lg text-sm">
											<span class="text-gray-600">📎</span>
											<span class="max-w-[150px] truncate">{att.name}</span>
											<span class="text-gray-400 text-xs">({formatFileSize(att.size)})</span>
											<button
												type="button"
												onclick={() => removeComposeAttachment(i)}
												class="text-gray-400 hover:text-red-500 ml-1"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
												</svg>
											</button>
										</div>
									{/each}
								</div>
							{/if}
							<label class="inline-flex items-center gap-2 px-3 py-2 text-sm text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
								</svg>
								Прикрепить файл
								<input type="file" multiple class="hidden" onchange={handleAttachmentSelect} />
							</label>
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

	<!-- Attachment Preview Modal -->
	{#if previewAttachment}
		<AttachmentPreview
			name={previewAttachment.name}
			contentType={previewAttachment.contentType}
			content={previewAttachment.content}
			onclose={() => previewAttachment = null}
			ondownload={downloadPreviewedAttachment}
		/>
	{/if}
{/if}
