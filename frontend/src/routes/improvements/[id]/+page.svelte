<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { improvements, auth, type ImprovementRequest, type ImprovementRequestComment } from '$lib/api/client';

	const id = $page.params.id;

	let loading = $state(true);
	let error = $state<string | null>(null);
	let actionLoading = $state(false);
	let currentUserId = $state<string | null>(null);

	let request = $state<ImprovementRequest | null>(null);

	// Comment form
	let commentContent = $state('');
	let commentInternal = $state(false);

	// Workflow modal
	let showApproveModal = $state(false);
	let showRejectModal = $state(false);
	let approveComment = $state('');
	let rejectReason = $state('');
	let approvedBudget = $state<number | null>(null);

	onMount(async () => {
		try {
			const user = await auth.getMe();
			currentUserId = user.id;
			await loadRequest();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Ошибка загрузки';
		} finally {
			loading = false;
		}
	});

	async function loadRequest() {
		request = await improvements.get(id);
		if (request?.approved_budget) {
			approvedBudget = request.approved_budget;
		}
	}

	async function handleSubmit() {
		if (!currentUserId || !request) return;
		actionLoading = true;
		try {
			await improvements.submit(id, currentUserId);
			await loadRequest();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Ошибка';
		} finally {
			actionLoading = false;
		}
	}

	async function handleApprove() {
		if (!currentUserId) return;
		actionLoading = true;
		try {
			await improvements.approve(id, {
				approver_id: currentUserId,
				comment: approveComment || undefined,
				approved_budget: approvedBudget || undefined,
			});
			showApproveModal = false;
			approveComment = '';
			await loadRequest();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Ошибка';
		} finally {
			actionLoading = false;
		}
	}

	async function handleReject() {
		if (!currentUserId || !rejectReason.trim()) {
			error = 'Укажите причину отклонения';
			return;
		}
		actionLoading = true;
		try {
			await improvements.reject(id, {
				rejector_id: currentUserId,
				reason: rejectReason.trim(),
			});
			showRejectModal = false;
			rejectReason = '';
			await loadRequest();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Ошибка';
		} finally {
			actionLoading = false;
		}
	}

	async function handleCreateProject() {
		if (!currentUserId) return;
		actionLoading = true;
		try {
			await improvements.createProject(id, currentUserId);
			await loadRequest();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Ошибка';
		} finally {
			actionLoading = false;
		}
	}

	async function handleAddComment() {
		if (!currentUserId || !commentContent.trim()) return;
		actionLoading = true;
		try {
			await improvements.addComment(id, {
				author_id: currentUserId,
				content: commentContent.trim(),
				is_internal: commentInternal,
			});
			commentContent = '';
			commentInternal = false;
			await loadRequest();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Ошибка';
		} finally {
			actionLoading = false;
		}
	}

	const statusLabels: Record<string, { label: string; color: string; step: number }> = {
		draft: { label: 'Черновик', color: 'bg-gray-400', step: 0 },
		submitted: { label: 'Подана', color: 'bg-blue-500', step: 1 },
		screening: { label: 'Скрининг', color: 'bg-cyan-500', step: 2 },
		evaluation: { label: 'Оценка', color: 'bg-indigo-500', step: 3 },
		manager_approval: { label: 'Согласование руководителя', color: 'bg-purple-500', step: 4 },
		committee_review: { label: 'Рассмотрение комитетом', color: 'bg-pink-500', step: 5 },
		budgeting: { label: 'Утверждение бюджета', color: 'bg-orange-500', step: 6 },
		project_created: { label: 'Проект создан', color: 'bg-green-500', step: 7 },
		in_progress: { label: 'В работе', color: 'bg-teal-500', step: 8 },
		completed: { label: 'Завершена', color: 'bg-emerald-500', step: 9 },
		rejected: { label: 'Отклонена', color: 'bg-red-500', step: -1 },
	};

	const workflowSteps = [
		{ status: 'draft', label: 'Черновик' },
		{ status: 'submitted', label: 'Подана' },
		{ status: 'screening', label: 'Скрининг' },
		{ status: 'evaluation', label: 'Оценка' },
		{ status: 'manager_approval', label: 'Согласование' },
		{ status: 'committee_review', label: 'Комитет' },
		{ status: 'budgeting', label: 'Бюджет' },
		{ status: 'project_created', label: 'Проект' },
	];

	function formatDate(dateStr: string | undefined): string {
		if (!dateStr) return '';
		return new Date(dateStr).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });
	}

	function formatDateTime(dateStr: string | undefined): string {
		if (!dateStr) return '';
		return new Date(dateStr).toLocaleString('ru-RU', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' });
	}

	function formatBudget(value: number | undefined): string {
		if (!value) return '';
		return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(value);
	}

	let currentStep = $derived(request ? statusLabels[request.status]?.step ?? 0 : 0);
	let canApprove = $derived(request && !['draft', 'project_created', 'in_progress', 'completed', 'rejected'].includes(request.status));
	let canReject = $derived(request && !['draft', 'rejected', 'completed'].includes(request.status));
</script>

<svelte:head>
	<title>{request?.title || 'Заявка'} - EKF Hub</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 py-8">
	<div class="max-w-4xl mx-auto px-6">
		{#if loading}
			<div class="flex items-center justify-center py-12">
				<div class="animate-spin w-8 h-8 border-4 border-ekf-red border-t-transparent rounded-full"></div>
			</div>
		{:else if !request}
			<div class="bg-white rounded-xl p-8 text-center">
				<p class="text-gray-500">Заявка не найдена</p>
				<a href="/improvements" class="text-ekf-red hover:underline mt-2 inline-block">Вернуться к списку</a>
			</div>
		{:else}
			<!-- Header -->
			<div class="mb-6">
				<a href="/improvements" class="text-sm text-gray-500 hover:text-ekf-red flex items-center gap-1 mb-2">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
					</svg>
					Все заявки
				</a>
				<div class="flex items-start justify-between">
					<div>
						<div class="flex items-center gap-3 mb-2">
							<span class="text-sm font-mono text-gray-500">{request.number}</span>
							<span class="px-3 py-1 rounded-full text-sm font-medium text-white {statusLabels[request.status]?.color || 'bg-gray-500'}">
								{statusLabels[request.status]?.label || request.status}
							</span>
						</div>
						<h1 class="text-2xl font-bold text-gray-900">{request.title}</h1>
					</div>
				</div>
			</div>

			{#if error}
				<div class="bg-red-50 text-red-700 p-4 rounded-xl mb-6">{error}</div>
			{/if}

			<!-- Workflow Progress -->
			{#if request.status !== 'rejected'}
				<div class="bg-white rounded-xl p-6 border border-gray-200 mb-6">
					<h2 class="font-medium text-gray-900 mb-4">Этапы согласования</h2>
					<div class="flex items-center justify-between">
						{#each workflowSteps as step, i}
							<div class="flex-1 relative">
								<div class="flex flex-col items-center">
									<div class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium transition-colors
										{i < currentStep ? 'bg-green-500 text-white' :
										 i === currentStep ? 'bg-ekf-red text-white' :
										 'bg-gray-200 text-gray-500'}">
										{#if i < currentStep}
											<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
												<path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
											</svg>
										{:else}
											{i + 1}
										{/if}
									</div>
									<span class="text-xs text-gray-500 mt-1 text-center">{step.label}</span>
								</div>
								{#if i < workflowSteps.length - 1}
									<div class="absolute top-4 left-1/2 w-full h-0.5 {i < currentStep ? 'bg-green-500' : 'bg-gray-200'}"></div>
								{/if}
							</div>
						{/each}
					</div>
				</div>
			{:else}
				<div class="bg-red-50 border border-red-200 rounded-xl p-6 mb-6">
					<h2 class="font-medium text-red-800 mb-2">Заявка отклонена</h2>
					<p class="text-red-700">{request.rejection_reason}</p>
					{#if request.rejected_at}
						<p class="text-sm text-red-600 mt-2">{formatDateTime(request.rejected_at)}</p>
					{/if}
				</div>
			{/if}

			<!-- Actions -->
			<div class="bg-white rounded-xl p-6 border border-gray-200 mb-6">
				<h2 class="font-medium text-gray-900 mb-4">Действия</h2>
				<div class="flex flex-wrap gap-3">
					{#if request.status === 'draft'}
						<button
							onclick={handleSubmit}
							disabled={actionLoading}
							class="px-4 py-2 bg-ekf-red text-white rounded-lg font-medium hover:bg-red-700 disabled:opacity-50 flex items-center gap-2"
						>
							{#if actionLoading}
								<div class="animate-spin w-4 h-4 border-2 border-white border-t-transparent rounded-full"></div>
							{/if}
							Подать на рассмотрение
						</button>
					{/if}

					{#if canApprove}
						<button
							onclick={() => showApproveModal = true}
							disabled={actionLoading}
							class="px-4 py-2 bg-green-600 text-white rounded-lg font-medium hover:bg-green-700 disabled:opacity-50"
						>
							Одобрить
						</button>
					{/if}

					{#if canReject}
						<button
							onclick={() => showRejectModal = true}
							disabled={actionLoading}
							class="px-4 py-2 bg-red-600 text-white rounded-lg font-medium hover:bg-red-700 disabled:opacity-50"
						>
							Отклонить
						</button>
					{/if}

					{#if request.status === 'project_created' && !request.project_id}
						<button
							onclick={handleCreateProject}
							disabled={actionLoading}
							class="px-4 py-2 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
						>
							{#if actionLoading}
								<div class="animate-spin w-4 h-4 border-2 border-white border-t-transparent rounded-full"></div>
							{/if}
							Создать проект
						</button>
					{/if}

					{#if request.project_id}
						<a
							href="/projects/{request.project_id}"
							class="px-4 py-2 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700"
						>
							Открыть проект
						</a>
					{/if}
				</div>
			</div>

			<!-- Details -->
			<div class="grid grid-cols-3 gap-6 mb-6">
				<div class="col-span-2 space-y-6">
					<!-- Description -->
					<div class="bg-white rounded-xl p-6 border border-gray-200">
						<h2 class="font-medium text-gray-900 mb-3">Описание</h2>
						<p class="text-gray-700 whitespace-pre-wrap">{request.description || 'Нет описания'}</p>
					</div>

					<!-- Business Value -->
					{#if request.business_value || request.expected_effect}
						<div class="bg-white rounded-xl p-6 border border-gray-200">
							<h2 class="font-medium text-gray-900 mb-4">Бизнес-обоснование</h2>
							{#if request.business_value}
								<div class="mb-4">
									<h3 class="text-sm font-medium text-gray-500 mb-1">Бизнес-ценность</h3>
									<p class="text-gray-700 whitespace-pre-wrap">{request.business_value}</p>
								</div>
							{/if}
							{#if request.expected_effect}
								<div>
									<h3 class="text-sm font-medium text-gray-500 mb-1">Ожидаемый эффект</h3>
									<p class="text-gray-700 whitespace-pre-wrap">{request.expected_effect}</p>
								</div>
							{/if}
						</div>
					{/if}

					<!-- Comments -->
					<div class="bg-white rounded-xl p-6 border border-gray-200">
						<h2 class="font-medium text-gray-900 mb-4">Комментарии</h2>

						{#if request.comments && request.comments.length > 0}
							<div class="space-y-4 mb-6">
								{#each request.comments as comment}
									<div class="flex gap-3 {comment.is_internal ? 'bg-yellow-50 -mx-3 px-3 py-3 rounded-lg' : ''}">
										{#if comment.author?.photo_base64}
											<img src="data:image/jpeg;base64,{comment.author.photo_base64}" alt="" class="w-8 h-8 rounded-full" />
										{:else}
											<div class="w-8 h-8 bg-gray-200 rounded-full flex items-center justify-center text-sm text-gray-500">
												{comment.author?.name?.charAt(0) || '?'}
											</div>
										{/if}
										<div class="flex-1">
											<div class="flex items-center gap-2">
												<span class="font-medium text-gray-900">{comment.author?.name || 'Неизвестно'}</span>
												{#if comment.is_internal}
													<span class="text-xs bg-yellow-200 text-yellow-800 px-1.5 py-0.5 rounded">Внутренний</span>
												{/if}
												<span class="text-sm text-gray-500">{formatDateTime(comment.created_at)}</span>
											</div>
											<p class="text-gray-700 mt-1">{comment.content}</p>
										</div>
									</div>
								{/each}
							</div>
						{/if}

						<div class="space-y-3">
							<textarea
								bind:value={commentContent}
								placeholder="Добавить комментарий..."
								rows="3"
								class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
							></textarea>
							<div class="flex items-center justify-between">
								<label class="flex items-center gap-2 text-sm">
									<input type="checkbox" bind:checked={commentInternal} class="rounded border-gray-300 text-ekf-red focus:ring-ekf-red" />
									Внутренний комментарий (только для рецензентов)
								</label>
								<button
									onclick={handleAddComment}
									disabled={actionLoading || !commentContent.trim()}
									class="px-4 py-2 bg-gray-900 text-white rounded-lg font-medium hover:bg-gray-800 disabled:opacity-50"
								>
									Отправить
								</button>
							</div>
						</div>
					</div>
				</div>

				<!-- Sidebar -->
				<div class="space-y-6">
					<!-- Meta -->
					<div class="bg-white rounded-xl p-6 border border-gray-200">
						<div class="space-y-4">
							<div>
								<span class="text-sm text-gray-500">Инициатор</span>
								{#if request.initiator}
									<div class="flex items-center gap-2 mt-1">
										{#if request.initiator.photo_base64}
											<img src="data:image/jpeg;base64,{request.initiator.photo_base64}" alt="" class="w-6 h-6 rounded-full" />
										{:else}
											<div class="w-6 h-6 bg-gray-200 rounded-full flex items-center justify-center text-xs">{request.initiator.name.charAt(0)}</div>
										{/if}
										<span class="text-gray-900">{request.initiator.name}</span>
									</div>
								{/if}
							</div>

							{#if request.sponsor}
								<div>
									<span class="text-sm text-gray-500">Спонсор</span>
									<div class="flex items-center gap-2 mt-1">
										{#if request.sponsor.photo_base64}
											<img src="data:image/jpeg;base64,{request.sponsor.photo_base64}" alt="" class="w-6 h-6 rounded-full" />
										{:else}
											<div class="w-6 h-6 bg-gray-200 rounded-full flex items-center justify-center text-xs">{request.sponsor.name.charAt(0)}</div>
										{/if}
										<span class="text-gray-900">{request.sponsor.name}</span>
									</div>
								</div>
							{/if}

							{#if request.type}
								<div>
									<span class="text-sm text-gray-500">Тип</span>
									<div class="text-gray-900 mt-1">{request.type.name}</div>
								</div>
							{/if}

							{#if request.estimated_budget}
								<div>
									<span class="text-sm text-gray-500">Предварительный бюджет</span>
									<div class="text-gray-900 mt-1">{formatBudget(request.estimated_budget)}</div>
								</div>
							{/if}

							{#if request.approved_budget}
								<div>
									<span class="text-sm text-gray-500">Утверждённый бюджет</span>
									<div class="text-green-600 font-medium mt-1">{formatBudget(request.approved_budget)}</div>
								</div>
							{/if}

							{#if request.estimated_start || request.estimated_end}
								<div>
									<span class="text-sm text-gray-500">Сроки</span>
									<div class="text-gray-900 mt-1">
										{formatDate(request.estimated_start)} - {formatDate(request.estimated_end)}
									</div>
								</div>
							{/if}

							<div>
								<span class="text-sm text-gray-500">Создана</span>
								<div class="text-gray-900 mt-1">{formatDateTime(request.created_at)}</div>
							</div>
						</div>
					</div>

					<!-- Activity -->
					{#if request.activity && request.activity.length > 0}
						<div class="bg-white rounded-xl p-6 border border-gray-200">
							<h3 class="font-medium text-gray-900 mb-4">История</h3>
							<div class="space-y-3">
								{#each request.activity.slice(0, 10) as activity}
									<div class="flex gap-2 text-sm">
										<div class="w-2 h-2 rounded-full bg-gray-300 mt-1.5"></div>
										<div>
											<span class="text-gray-900">{activity.action}</span>
											{#if activity.old_value && activity.new_value}
												<span class="text-gray-500">: {activity.old_value} → {activity.new_value}</span>
											{/if}
											<div class="text-xs text-gray-400">{formatDateTime(activity.created_at)}</div>
										</div>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- Approve Modal -->
{#if showApproveModal}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
		<div class="bg-white rounded-xl p-6 w-full max-w-md">
			<h2 class="text-lg font-bold text-gray-900 mb-4">Одобрить заявку</h2>

			{#if request?.status === 'budgeting'}
				<div class="mb-4">
					<label class="block text-sm font-medium text-gray-700 mb-1">Утверждённый бюджет (руб)</label>
					<input
						type="number"
						bind:value={approvedBudget}
						placeholder="0"
						class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
					/>
				</div>
			{/if}

			<div class="mb-6">
				<label class="block text-sm font-medium text-gray-700 mb-1">Комментарий (необязательно)</label>
				<textarea
					bind:value={approveComment}
					rows="3"
					placeholder="Ваш комментарий к одобрению..."
					class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
				></textarea>
			</div>

			<div class="flex gap-3">
				<button
					onclick={() => showApproveModal = false}
					class="flex-1 py-2.5 border border-gray-300 rounded-lg font-medium text-gray-700 hover:bg-gray-50"
				>
					Отмена
				</button>
				<button
					onclick={handleApprove}
					disabled={actionLoading}
					class="flex-1 py-2.5 bg-green-600 text-white rounded-lg font-medium hover:bg-green-700 disabled:opacity-50"
				>
					Одобрить
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Reject Modal -->
{#if showRejectModal}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
		<div class="bg-white rounded-xl p-6 w-full max-w-md">
			<h2 class="text-lg font-bold text-gray-900 mb-4">Отклонить заявку</h2>

			<div class="mb-6">
				<label class="block text-sm font-medium text-gray-700 mb-1">Причина отклонения <span class="text-red-500">*</span></label>
				<textarea
					bind:value={rejectReason}
					rows="4"
					placeholder="Укажите причину отклонения..."
					class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
				></textarea>
			</div>

			<div class="flex gap-3">
				<button
					onclick={() => showRejectModal = false}
					class="flex-1 py-2.5 border border-gray-300 rounded-lg font-medium text-gray-700 hover:bg-gray-50"
				>
					Отмена
				</button>
				<button
					onclick={handleReject}
					disabled={actionLoading || !rejectReason.trim()}
					class="flex-1 py-2.5 bg-red-600 text-white rounded-lg font-medium hover:bg-red-700 disabled:opacity-50"
				>
					Отклонить
				</button>
			</div>
		</div>
	</div>
{/if}
