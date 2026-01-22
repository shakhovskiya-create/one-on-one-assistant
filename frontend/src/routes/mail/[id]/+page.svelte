<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';
	import { mail } from '$lib/api/client';
	import type { EmailMessage, EmailAttachment } from '$lib/api/client';
	import AttachmentPreview from '$lib/components/AttachmentPreview.svelte';
	import {
		ArrowLeft,
		Reply,
		ReplyAll,
		Forward,
		Trash2,
		Download,
		Paperclip,
		Calendar,
		Check,
		X,
		HelpCircle,
		Eye
	} from 'lucide-svelte';

	// Get email ID from URL
	const emailId = $page.params.id;

	// State
	let email: EmailMessage | null = $state(null);
	let loading = $state(true);
	let error = $state('');
	let credentials = $state({ username: '', password: '' });
	let attachments: EmailAttachment[] = $state([]);
	let loadingAttachments = $state(false);
	let downloadingAttachment = $state<string | null>(null);

	// Attachment preview
	let previewAttachment: { name: string; contentType: string; content: string } | null = $state(null);

	// Meeting response
	let respondingToMeeting = $state<'Accept' | 'Decline' | 'Tentative' | null>(null);
	let meetingResponseSuccess = $state<string | null>(null);

	onMount(async () => {
		if (!browser) return;

		// Get credentials from session storage
		const stored = sessionStorage.getItem('ews_credentials');
		if (!stored) {
			goto('/mail');
			return;
		}

		try {
			credentials = JSON.parse(stored);
		} catch {
			goto('/mail');
			return;
		}

		await loadEmail();
	});

	async function loadEmail() {
		loading = true;
		error = '';

		try {
			// First get the email body
			const bodyRes = await mail.getEmailBody({
				username: credentials.username,
				password: credentials.password,
				item_id: emailId
			});

			// We need to get email metadata too - but for now, use body
			email = {
				id: emailId,
				subject: 'Загрузка...',
				from: null,
				to: [],
				received_at: '',
				body: bodyRes.body,
				is_read: true,
				has_attachments: false
			} as EmailMessage;

			// Mark as read
			await mail.markAsRead({
				username: credentials.username,
				password: credentials.password,
				item_id: emailId
			});

			// Load attachments
			await loadAttachments();
		} catch (e) {
			console.error('Failed to load email:', e);
			error = e instanceof Error ? e.message : 'Не удалось загрузить письмо';
		} finally {
			loading = false;
		}
	}

	async function loadAttachments() {
		loadingAttachments = true;
		try {
			const res = await mail.getAttachments({
				username: credentials.username,
				password: credentials.password,
				item_id: emailId
			});
			attachments = res.attachments || [];
		} catch (e) {
			console.error('Failed to load attachments:', e);
		} finally {
			loadingAttachments = false;
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

	async function previewOrDownloadAttachment(att: EmailAttachment) {
		downloadingAttachment = att.id;
		try {
			const res = await mail.getAttachmentContent({
				username: credentials.username,
				password: credentials.password,
				attachment_id: att.id
			});

			// Check if can preview
			if (canPreview(res.content_type)) {
				previewAttachment = {
					name: res.name || att.name,
					contentType: res.content_type,
					content: res.content
				};
			} else {
				// Download directly
				const byteCharacters = atob(res.content);
				const byteNumbers = new Array(byteCharacters.length);
				for (let i = 0; i < byteCharacters.length; i++) {
					byteNumbers[i] = byteCharacters.charCodeAt(i);
				}
				const byteArray = new Uint8Array(byteNumbers);
				const blob = new Blob([byteArray], { type: res.content_type });

				const url = URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = res.name || att.name;
				document.body.appendChild(a);
				a.click();
				document.body.removeChild(a);
				URL.revokeObjectURL(url);
			}
		} catch (e) {
			console.error('Failed to load attachment:', e);
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
		const a = document.createElement('a');
		a.href = url;
		a.download = previewAttachment.name;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	async function deleteEmail() {
		if (!confirm('Удалить это письмо?')) return;

		try {
			await mail.deleteEmail({
				username: credentials.username,
				password: credentials.password,
				item_id: emailId
			});
			goto('/mail');
		} catch (e) {
			console.error('Failed to delete email:', e);
			error = e instanceof Error ? e.message : 'Не удалось удалить письмо';
		}
	}

	async function respondToMeeting(response: 'Accept' | 'Decline' | 'Tentative') {
		respondingToMeeting = response;
		meetingResponseSuccess = null;

		try {
			await mail.respondToMeeting({
				username: credentials.username,
				password: credentials.password,
				item_id: emailId,
				response
			});
			meetingResponseSuccess = response === 'Accept' ? 'Принято' : response === 'Decline' ? 'Отклонено' : 'Под вопросом';
		} catch (e) {
			console.error('Failed to respond to meeting:', e);
			error = 'Не удалось ответить на приглашение';
		} finally {
			respondingToMeeting = null;
		}
	}

	function getAttachmentIcon(contentType: string): string {
		if (contentType.startsWith('image/')) return '🖼️';
		if (contentType.includes('pdf')) return '📄';
		if (contentType.includes('word') || contentType.includes('document')) return '📝';
		if (contentType.includes('excel') || contentType.includes('spreadsheet')) return '📊';
		if (contentType.includes('powerpoint') || contentType.includes('presentation')) return '📽️';
		if (contentType.startsWith('video/')) return '🎬';
		if (contentType.startsWith('audio/')) return '🎵';
		if (contentType.includes('zip') || contentType.includes('rar') || contentType.includes('7z')) return '📦';
		if (contentType.startsWith('text/')) return '📃';
		return '📎';
	}

	function formatFileSize(bytes: number): string {
		if (bytes < 1024) return bytes + ' Б';
		if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' КБ';
		return (bytes / (1024 * 1024)).toFixed(1) + ' МБ';
	}

	function goToReply() {
		// Store email data for reply and go to mail page
		if (browser && email) {
			sessionStorage.setItem('reply_email', JSON.stringify({ email, mode: 'reply' }));
			goto('/mail?compose=reply');
		}
	}

	function goToReplyAll() {
		if (browser && email) {
			sessionStorage.setItem('reply_email', JSON.stringify({ email, mode: 'replyAll' }));
			goto('/mail?compose=replyAll');
		}
	}

	function goToForward() {
		if (browser && email) {
			sessionStorage.setItem('reply_email', JSON.stringify({ email, mode: 'forward' }));
			goto('/mail?compose=forward');
		}
	}
</script>

<svelte:head>
	<title>{email?.subject || 'Письмо'} - Почта - EKF Hub</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 p-4">
	<div class="max-w-4xl mx-auto">
		<!-- Header -->
		<div class="flex items-center gap-4 mb-4">
			<button
				onclick={() => goto('/mail')}
				class="p-2 hover:bg-gray-200 rounded-lg transition-colors"
				title="Назад к списку"
			>
				<ArrowLeft size={20} />
			</button>
			<h1 class="text-xl font-semibold text-gray-800 flex-1 truncate">
				{email?.subject || 'Загрузка...'}
			</h1>
		</div>

		{#if loading}
			<div class="bg-white rounded-xl shadow-sm p-8 flex items-center justify-center">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
				<span class="ml-3 text-gray-600">Загрузка письма...</span>
			</div>
		{:else if error}
			<div class="bg-white rounded-xl shadow-sm p-8 text-center">
				<div class="text-red-500 mb-4">{error}</div>
				<button
					onclick={() => goto('/mail')}
					class="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300"
				>
					Вернуться к почте
				</button>
			</div>
		{:else if email}
			<div class="bg-white rounded-xl shadow-sm overflow-hidden">
				<!-- Action bar -->
				<div class="px-6 py-3 border-b border-gray-200 flex items-center gap-2 bg-gray-50">
					<button
						onclick={goToReply}
						class="flex items-center gap-2 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-200 rounded-lg transition-colors"
					>
						<Reply size={16} />
						Ответить
					</button>
					<button
						onclick={goToReplyAll}
						class="flex items-center gap-2 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-200 rounded-lg transition-colors"
					>
						<ReplyAll size={16} />
						Ответить всем
					</button>
					<button
						onclick={goToForward}
						class="flex items-center gap-2 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-200 rounded-lg transition-colors"
					>
						<Forward size={16} />
						Переслать
					</button>
					<div class="flex-1"></div>
					<button
						onclick={deleteEmail}
						class="flex items-center gap-2 px-3 py-1.5 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors"
					>
						<Trash2 size={16} />
						Удалить
					</button>
				</div>

				<!-- Email header info would go here if we had full metadata -->

				<!-- Meeting invite actions -->
				{#if email.is_meeting_request}
					<div class="px-6 py-4 border-b border-gray-200 bg-blue-50">
						<div class="flex items-center gap-2 mb-3">
							<Calendar size={20} class="text-blue-600" />
							<span class="font-medium text-blue-800">Приглашение на встречу</span>
						</div>
						{#if meetingResponseSuccess}
							<div class="text-green-600 font-medium">
								Ваш ответ: {meetingResponseSuccess}
							</div>
						{:else}
							<div class="flex gap-2">
								<button
									onclick={() => respondToMeeting('Accept')}
									disabled={respondingToMeeting !== null}
									class="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors disabled:opacity-50"
								>
									{#if respondingToMeeting === 'Accept'}
										<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
									{:else}
										<Check size={16} />
									{/if}
									Принять
								</button>
								<button
									onclick={() => respondToMeeting('Tentative')}
									disabled={respondingToMeeting !== null}
									class="flex items-center gap-2 px-4 py-2 bg-yellow-500 text-white rounded-lg hover:bg-yellow-600 transition-colors disabled:opacity-50"
								>
									{#if respondingToMeeting === 'Tentative'}
										<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
									{:else}
										<HelpCircle size={16} />
									{/if}
									Под вопросом
								</button>
								<button
									onclick={() => respondToMeeting('Decline')}
									disabled={respondingToMeeting !== null}
									class="flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50"
								>
									{#if respondingToMeeting === 'Decline'}
										<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
									{:else}
										<X size={16} />
									{/if}
									Отклонить
								</button>
							</div>
						{/if}
					</div>
				{/if}

				<!-- Attachments -->
				{#if attachments.length > 0}
					<div class="px-6 py-4 border-b border-gray-200 bg-gray-50">
						<div class="flex items-center gap-2 mb-3">
							<Paperclip size={16} class="text-gray-500" />
							<span class="text-sm font-medium text-gray-700">Вложения ({attachments.length})</span>
						</div>
						<div class="flex flex-wrap gap-2">
							{#each attachments as att}
								<button
									onclick={() => previewOrDownloadAttachment(att)}
									disabled={downloadingAttachment === att.id}
									class="flex items-center gap-2 px-3 py-2 bg-white border border-gray-200 rounded-lg hover:border-gray-300 hover:bg-gray-50 transition-colors disabled:opacity-50"
									title={canPreview(att.content_type) ? 'Предпросмотр' : 'Скачать'}
								>
									<span>{getAttachmentIcon(att.content_type)}</span>
									<span class="max-w-[200px] truncate text-sm">{att.name}</span>
									<span class="text-xs text-gray-400">({formatFileSize(att.size)})</span>
									{#if downloadingAttachment === att.id}
										<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-600"></div>
									{:else if canPreview(att.content_type)}
										<Eye size={14} class="text-gray-400" />
									{:else}
										<Download size={14} class="text-gray-400" />
									{/if}
								</button>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Email body -->
				<div class="p-6">
					<div class="prose max-w-none">
						{@html email.body}
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>

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

<style>
	:global(.prose) {
		line-height: 1.6;
	}

	:global(.prose img) {
		max-width: 100%;
		height: auto;
	}

	:global(.prose a) {
		color: #2563eb;
		text-decoration: underline;
	}

	:global(.prose table) {
		width: 100%;
		border-collapse: collapse;
	}

	:global(.prose td),
	:global(.prose th) {
		padding: 0.5rem;
		border: 1px solid #e5e7eb;
	}
</style>
