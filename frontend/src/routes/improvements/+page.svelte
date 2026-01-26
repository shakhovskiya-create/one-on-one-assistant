<script lang="ts">
	import { onMount } from 'svelte';
	import { improvements, auth, type ImprovementRequest, type ImprovementRequestType, type ImprovementRequestStats } from '$lib/api/client';

	let loading = $state(true);
	let error = $state<string | null>(null);
	let currentUserId = $state<string | null>(null);

	let requests = $state<ImprovementRequest[]>([]);
	let types = $state<ImprovementRequestType[]>([]);
	let stats = $state<ImprovementRequestStats | null>(null);

	// Filters
	let statusFilter = $state<string>('');
	let typeFilter = $state<string>('');
	let showMine = $state(false);

	onMount(async () => {
		try {
			const [user, typesData, statsData] = await Promise.all([
				auth.getMe(),
				improvements.getTypes(),
				improvements.getStats()
			]);
			currentUserId = user.id;
			types = typesData;
			stats = statsData;
			await loadRequests();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Ошибка загрузки';
		} finally {
			loading = false;
		}
	});

	async function loadRequests() {
		try {
			const params: Record<string, string> = {};
			if (statusFilter) params.status = statusFilter;
			if (typeFilter) params.type_id = typeFilter;
			if (showMine && currentUserId) params.initiator_id = currentUserId;
			requests = await improvements.list(params);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Ошибка загрузки';
		}
	}

	$effect(() => {
		if (!loading) {
			loadRequests();
		}
	});

	const statusLabels: Record<string, { label: string; color: string }> = {
		draft: { label: 'Черновик', color: 'bg-gray-100 text-gray-700' },
		submitted: { label: 'Подана', color: 'bg-blue-100 text-blue-700' },
		screening: { label: 'Скрининг', color: 'bg-cyan-100 text-cyan-700' },
		evaluation: { label: 'Оценка', color: 'bg-indigo-100 text-indigo-700' },
		manager_approval: { label: 'Согласование', color: 'bg-purple-100 text-purple-700' },
		committee_review: { label: 'Комитет', color: 'bg-pink-100 text-pink-700' },
		budgeting: { label: 'Бюджет', color: 'bg-orange-100 text-orange-700' },
		project_created: { label: 'Проект создан', color: 'bg-green-100 text-green-700' },
		in_progress: { label: 'В работе', color: 'bg-teal-100 text-teal-700' },
		completed: { label: 'Завершена', color: 'bg-emerald-100 text-emerald-700' },
		rejected: { label: 'Отклонена', color: 'bg-red-100 text-red-700' },
	};

	const priorityLabels: Record<string, { label: string; color: string }> = {
		low: { label: 'Низкий', color: 'bg-green-100 text-green-700' },
		medium: { label: 'Средний', color: 'bg-yellow-100 text-yellow-700' },
		high: { label: 'Высокий', color: 'bg-orange-100 text-orange-700' },
		critical: { label: 'Критический', color: 'bg-red-100 text-red-700' },
	};

	function formatDate(dateStr: string | undefined): string {
		if (!dateStr) return '';
		return new Date(dateStr).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function formatBudget(value: number | undefined): string {
		if (!value) return '';
		return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(value);
	}
</script>

<svelte:head>
	<title>Заявки на улучшение - EKF Hub</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 py-8">
	<div class="max-w-7xl mx-auto px-6">
		<!-- Header -->
		<div class="flex items-center justify-between mb-6">
			<div>
				<h1 class="text-2xl font-bold text-gray-900">Заявки на улучшение</h1>
				<p class="text-gray-500 mt-1">Управление инициативами и улучшениями</p>
			</div>
			<a
				href="/improvements/create"
				class="flex items-center gap-2 px-4 py-2.5 bg-ekf-red text-white rounded-xl font-medium hover:bg-red-700 transition-colors"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Новая заявка
			</a>
		</div>

		{#if error}
			<div class="bg-red-50 text-red-700 p-4 rounded-xl mb-6">{error}</div>
		{/if}

		<!-- Stats -->
		{#if stats}
			<div class="grid grid-cols-5 gap-4 mb-6">
				<div class="bg-white rounded-xl p-4 border border-gray-200">
					<div class="text-2xl font-bold text-gray-900">{stats.total}</div>
					<div class="text-sm text-gray-500">Всего заявок</div>
				</div>
				<div class="bg-white rounded-xl p-4 border border-gray-200">
					<div class="text-2xl font-bold text-gray-400">{stats.draft}</div>
					<div class="text-sm text-gray-500">Черновики</div>
				</div>
				<div class="bg-white rounded-xl p-4 border border-gray-200">
					<div class="text-2xl font-bold text-blue-600">{stats.pending}</div>
					<div class="text-sm text-gray-500">На рассмотрении</div>
				</div>
				<div class="bg-white rounded-xl p-4 border border-gray-200">
					<div class="text-2xl font-bold text-green-600">{stats.approved}</div>
					<div class="text-sm text-gray-500">Одобрено</div>
				</div>
				<div class="bg-white rounded-xl p-4 border border-gray-200">
					<div class="text-2xl font-bold text-red-600">{stats.rejected}</div>
					<div class="text-sm text-gray-500">Отклонено</div>
				</div>
			</div>
		{/if}

		<!-- Filters -->
		<div class="bg-white rounded-xl p-4 border border-gray-200 mb-6">
			<div class="flex items-center gap-4">
				<label class="flex items-center gap-2">
					<input
						type="checkbox"
						bind:checked={showMine}
						class="w-4 h-4 rounded border-gray-300 text-ekf-red focus:ring-ekf-red"
					/>
					<span class="text-sm text-gray-700">Только мои</span>
				</label>

				<select
					bind:value={statusFilter}
					class="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
				>
					<option value="">Все статусы</option>
					{#each Object.entries(statusLabels) as [value, { label }]}
						<option {value}>{label}</option>
					{/each}
				</select>

				<select
					bind:value={typeFilter}
					class="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
				>
					<option value="">Все типы</option>
					{#each types as type}
						<option value={type.id}>{type.name}</option>
					{/each}
				</select>
			</div>
		</div>

		<!-- List -->
		{#if loading}
			<div class="flex items-center justify-center py-12">
				<div class="animate-spin w-8 h-8 border-4 border-ekf-red border-t-transparent rounded-full"></div>
			</div>
		{:else if requests.length === 0}
			<div class="bg-white rounded-xl p-12 border border-gray-200 text-center">
				<svg class="w-12 h-12 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
				</svg>
				<h3 class="text-lg font-medium text-gray-900 mb-2">Заявок не найдено</h3>
				<p class="text-gray-500 mb-4">Создайте первую заявку на улучшение</p>
				<a
					href="/improvements/create"
					class="inline-flex items-center gap-2 px-4 py-2 bg-ekf-red text-white rounded-lg font-medium hover:bg-red-700"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					Создать заявку
				</a>
			</div>
		{:else}
			<div class="space-y-4">
				{#each requests as request}
					<a
						href="/improvements/{request.id}"
						class="block bg-white rounded-xl p-5 border border-gray-200 hover:border-ekf-red hover:shadow-lg transition-all"
					>
						<div class="flex items-start justify-between">
							<div class="flex-1">
								<div class="flex items-center gap-3 mb-2">
									<span class="text-sm font-mono text-gray-500">{request.number}</span>
									<span class="px-2 py-0.5 rounded-full text-xs font-medium {statusLabels[request.status]?.color || 'bg-gray-100 text-gray-700'}">
										{statusLabels[request.status]?.label || request.status}
									</span>
									{#if request.priority}
										<span class="px-2 py-0.5 rounded-full text-xs font-medium {priorityLabels[request.priority]?.color || 'bg-gray-100 text-gray-700'}">
											{priorityLabels[request.priority]?.label || request.priority}
										</span>
									{/if}
									{#if request.type}
										<span class="px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-600">
											{request.type.name}
										</span>
									{/if}
								</div>
								<h3 class="text-lg font-semibold text-gray-900 mb-2">{request.title}</h3>
								{#if request.description}
									<p class="text-gray-600 text-sm line-clamp-2">{request.description}</p>
								{/if}
								<div class="flex items-center gap-6 mt-3 text-sm text-gray-500">
									{#if request.initiator}
										<div class="flex items-center gap-2">
											{#if request.initiator.photo_base64}
												<img src="data:image/jpeg;base64,{request.initiator.photo_base64}" alt="" class="w-5 h-5 rounded-full" />
											{:else}
												<div class="w-5 h-5 bg-gray-200 rounded-full flex items-center justify-center text-xs text-gray-500">
													{request.initiator.name.charAt(0)}
												</div>
											{/if}
											{request.initiator.name}
										</div>
									{/if}
									{#if request.estimated_budget}
										<span>{formatBudget(request.estimated_budget)}</span>
									{/if}
									{#if request.created_at}
										<span>{formatDate(request.created_at)}</span>
									{/if}
								</div>
							</div>
							<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
							</svg>
						</div>
					</a>
				{/each}
			</div>
		{/if}
	</div>
</div>
