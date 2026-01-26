<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { serviceDesk, auth } from '$lib/api/client';
	import type { ServiceTicket } from '$lib/api/client';

	let ticket = $state<ServiceTicket | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let currentUserId = $state<string | null>(null);
	let newComment = $state('');
	let isInternal = $state(false);
	let submitting = $state(false);

	const ticketId = $page.params.id;

	onMount(async () => {
		try {
			const [ticketData, user] = await Promise.all([
				serviceDesk.getTicket(ticketId),
				auth.getMe()
			]);
			ticket = ticketData;
			currentUserId = user.id;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load ticket';
		} finally {
			loading = false;
		}
	});

	async function handleAddComment() {
		if (!newComment.trim() || !currentUserId || !ticket) return;

		submitting = true;
		try {
			const comment = await serviceDesk.addComment(ticket.id, {
				author_id: currentUserId,
				content: newComment.trim(),
				is_internal: isInternal,
			});

			ticket.comments = [...(ticket.comments || []), comment];
			newComment = '';
		} catch (e) {
			alert(e instanceof Error ? e.message : 'Failed to add comment');
		} finally {
			submitting = false;
		}
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'new': return 'bg-blue-100 text-blue-700';
			case 'in_progress': return 'bg-yellow-100 text-yellow-700';
			case 'pending': return 'bg-purple-100 text-purple-700';
			case 'resolved': return 'bg-green-100 text-green-700';
			case 'closed': return 'bg-gray-100 text-gray-700';
			default: return 'bg-gray-100 text-gray-700';
		}
	}

	function getStatusLabel(status: string): string {
		switch (status) {
			case 'new': return 'Новый';
			case 'in_progress': return 'В работе';
			case 'pending': return 'Ожидает';
			case 'resolved': return 'Решён';
			case 'closed': return 'Закрыт';
			default: return status;
		}
	}

	function getPriorityColor(priority: string): string {
		switch (priority) {
			case 'critical': return 'text-red-600';
			case 'high': return 'text-orange-600';
			case 'medium': return 'text-yellow-600';
			case 'low': return 'text-green-600';
			default: return 'text-gray-600';
		}
	}

	function getPriorityLabel(priority: string): string {
		switch (priority) {
			case 'critical': return 'Критический';
			case 'high': return 'Высокий';
			case 'medium': return 'Средний';
			case 'low': return 'Низкий';
			default: return priority;
		}
	}

	function getTypeLabel(type: string): string {
		switch (type) {
			case 'incident': return 'Инцидент';
			case 'service_request': return 'Запрос';
			case 'change': return 'Изменение';
			case 'problem': return 'Проблема';
			default: return type;
		}
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleDateString('ru-RU', {
			day: 'numeric',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getActivityIcon(action: string): string {
		switch (action) {
			case 'created': return 'M12 4v16m8-8H4';
			case 'status_changed': return 'M9 5l7 7-7 7';
			case 'assigned': return 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z';
			case 'comment_added':
			case 'internal_note_added': return 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z';
			default: return 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z';
		}
	}

	function getActivityLabel(action: string, newValue?: string, oldValue?: string): string {
		switch (action) {
			case 'created': return 'создал заявку';
			case 'status_changed':
				return `изменил статус: ${getStatusLabel(oldValue || '')} → ${getStatusLabel(newValue || '')}`;
			case 'assigned': return 'назначил исполнителя';
			case 'comment_added': return 'добавил комментарий';
			case 'internal_note_added': return 'добавил внутреннюю заметку';
			default: return action;
		}
	}
</script>

<svelte:head>
	<title>{ticket ? ticket.number : 'Заявка'} - Service Desk - EKF Hub</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 py-8">
	<div class="max-w-4xl mx-auto px-6">
		{#if loading}
			<div class="flex items-center justify-center py-20">
				<div class="animate-spin w-10 h-10 border-4 border-ekf-red border-t-transparent rounded-full"></div>
			</div>
		{:else if error}
			<div class="bg-red-50 text-red-700 p-4 rounded-xl">{error}</div>
		{:else if ticket}
			<!-- Back link -->
			<a href="/service-desk" class="text-sm text-gray-500 hover:text-ekf-red flex items-center gap-1 mb-4">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
				</svg>
				Назад к Service Desk
			</a>

			<!-- Header -->
			<div class="bg-white rounded-xl p-6 border border-gray-200 mb-6">
				<div class="flex items-start justify-between mb-4">
					<div>
						<div class="flex items-center gap-3 mb-2">
							<span class="text-sm text-gray-500">{ticket.number}</span>
							<span class="px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-700">
								{getTypeLabel(ticket.type)}
							</span>
							<span class="px-2 py-0.5 rounded-full text-xs font-medium {getStatusColor(ticket.status)}">
								{getStatusLabel(ticket.status)}
							</span>
						</div>
						<h1 class="text-xl font-bold text-gray-900">{ticket.title}</h1>
					</div>
					<div class="flex items-center gap-2 text-sm {getPriorityColor(ticket.priority)}">
						<div class="w-2 h-2 rounded-full bg-current"></div>
						{getPriorityLabel(ticket.priority)}
					</div>
				</div>

				{#if ticket.description}
					<p class="text-gray-600 mb-4">{ticket.description}</p>
				{/if}

				<!-- Meta -->
				<div class="grid grid-cols-3 gap-4 pt-4 border-t border-gray-100">
					<div>
						<div class="text-xs text-gray-500 mb-1">Заявитель</div>
						<div class="flex items-center gap-2">
							{#if ticket.requester?.photo_base64}
								<img src="data:image/jpeg;base64,{ticket.requester.photo_base64}" alt="" class="w-8 h-8 rounded-full object-cover">
							{:else}
								<div class="w-8 h-8 rounded-full bg-ekf-red flex items-center justify-center text-white text-sm">
									{ticket.requester?.name?.charAt(0) || '?'}
								</div>
							{/if}
							<div>
								<div class="font-medium text-sm">{ticket.requester?.name}</div>
								<div class="text-xs text-gray-500">{ticket.requester?.department}</div>
							</div>
						</div>
					</div>

					<div>
						<div class="text-xs text-gray-500 mb-1">Исполнитель</div>
						{#if ticket.assignee}
							<div class="flex items-center gap-2">
								{#if ticket.assignee.photo_base64}
									<img src="data:image/jpeg;base64,{ticket.assignee.photo_base64}" alt="" class="w-8 h-8 rounded-full object-cover">
								{:else}
									<div class="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white text-sm">
										{ticket.assignee.name.charAt(0)}
									</div>
								{/if}
								<div>
									<div class="font-medium text-sm">{ticket.assignee.name}</div>
									<div class="text-xs text-gray-500">{ticket.assignee.position}</div>
								</div>
							</div>
						{:else}
							<div class="text-sm text-gray-400">Не назначен</div>
						{/if}
					</div>

					<div>
						<div class="text-xs text-gray-500 mb-1">Создано</div>
						<div class="font-medium text-sm">{ticket.created_at ? formatDate(ticket.created_at) : '-'}</div>
						{#if ticket.sla_deadline}
							<div class="text-xs text-gray-500 mt-1">
								SLA: {formatDate(ticket.sla_deadline)}
							</div>
						{/if}
					</div>
				</div>
			</div>

			<!-- Comments & Activity -->
			<div class="grid grid-cols-3 gap-6">
				<!-- Comments -->
				<div class="col-span-2">
					<div class="bg-white rounded-xl p-6 border border-gray-200">
						<h2 class="font-semibold mb-4">Комментарии</h2>

						<!-- Add Comment -->
						<div class="mb-6">
							<textarea
								bind:value={newComment}
								placeholder="Добавьте комментарий..."
								rows="3"
								class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
							></textarea>
							<div class="flex items-center justify-between mt-2">
								<label class="flex items-center gap-2 text-sm text-gray-600">
									<input type="checkbox" bind:checked={isInternal} class="rounded border-gray-300 text-ekf-red focus:ring-ekf-red">
									Внутренняя заметка
								</label>
								<button
									onclick={handleAddComment}
									disabled={!newComment.trim() || submitting}
									class="px-4 py-2 bg-ekf-red text-white rounded-lg text-sm font-medium hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed"
								>
									{submitting ? 'Отправка...' : 'Отправить'}
								</button>
							</div>
						</div>

						<!-- Comments List -->
						{#if ticket.comments && ticket.comments.length > 0}
							<div class="space-y-4">
								{#each ticket.comments as comment}
									<div class="flex gap-3 {comment.is_internal ? 'bg-yellow-50 p-3 rounded-lg' : ''}">
										{#if comment.author?.photo_base64}
											<img src="data:image/jpeg;base64,{comment.author.photo_base64}" alt="" class="w-8 h-8 rounded-full object-cover flex-shrink-0">
										{:else}
											<div class="w-8 h-8 rounded-full bg-gray-300 flex items-center justify-center text-white text-sm flex-shrink-0">
												{comment.author?.name?.charAt(0) || '?'}
											</div>
										{/if}
										<div class="flex-1">
											<div class="flex items-center gap-2 mb-1">
												<span class="font-medium text-sm">{comment.author?.name}</span>
												{#if comment.is_internal}
													<span class="text-xs bg-yellow-200 text-yellow-800 px-2 py-0.5 rounded">Внутренняя</span>
												{/if}
												<span class="text-xs text-gray-400">{comment.created_at ? formatDate(comment.created_at) : ''}</span>
											</div>
											<p class="text-sm text-gray-600">{comment.content}</p>
										</div>
									</div>
								{/each}
							</div>
						{:else}
							<div class="text-sm text-gray-400 text-center py-4">Нет комментариев</div>
						{/if}
					</div>
				</div>

				<!-- Activity -->
				<div>
					<div class="bg-white rounded-xl p-5 border border-gray-200">
						<h2 class="font-semibold mb-4 text-sm">История</h2>
						{#if ticket.activity && ticket.activity.length > 0}
							<div class="space-y-3">
								{#each ticket.activity as activity}
									<div class="flex gap-3">
										<div class="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center flex-shrink-0">
											<svg class="w-3 h-3 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getActivityIcon(activity.action)}></path>
											</svg>
										</div>
										<div>
											<div class="text-xs">
												<span class="font-medium">{activity.actor?.name || 'Система'}</span>
												<span class="text-gray-500"> {getActivityLabel(activity.action, activity.new_value || undefined, activity.old_value || undefined)}</span>
											</div>
											<div class="text-xs text-gray-400 mt-0.5">{activity.created_at ? formatDate(activity.created_at) : ''}</div>
										</div>
									</div>
								{/each}
							</div>
						{:else}
							<div class="text-sm text-gray-400 text-center py-4">Нет истории</div>
						{/if}
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>
