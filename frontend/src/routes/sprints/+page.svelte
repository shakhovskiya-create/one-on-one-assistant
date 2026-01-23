<script lang="ts">
	import { onMount } from 'svelte';
	import { sprints as sprintsApi, projects as projectsApi } from '$lib/api/client';
	import type { Sprint } from '$lib/api/client';

	interface Project {
		id: string;
		name: string;
	}

	let sprints: Sprint[] = $state([]);
	let projects: Project[] = $state([]);
	let loading = $state(true);
	let error = $state('');

	// Modal state
	let showModal = $state(false);
	let editingSprint: Partial<Sprint> | null = $state(null);
	let saving = $state(false);

	// Filter
	let filterStatus = $state('');
	let filterProject = $state('');

	onMount(() => {
		loadData();
	});

	async function loadData() {
		loading = true;
		error = '';
		try {
			const [sprintsRes, projectsRes] = await Promise.all([
				sprintsApi.list(),
				projectsApi.list().catch(() => [])
			]);
			sprints = sprintsRes || [];
			projects = projectsRes || [];
		} catch (e) {
			error = 'Ошибка загрузки данных';
			console.error(e);
		}
		loading = false;
	}

	function openNewSprint() {
		const today = new Date();
		const twoWeeksLater = new Date(today);
		twoWeeksLater.setDate(twoWeeksLater.getDate() + 14);

		editingSprint = {
			name: `Sprint ${sprints.length + 1}`,
			goal: '',
			start_date: today.toISOString().split('T')[0],
			end_date: twoWeeksLater.toISOString().split('T')[0],
			project_id: filterProject || undefined
		};
		showModal = true;
	}

	function openEditSprint(sprint: Sprint) {
		editingSprint = { ...sprint };
		showModal = true;
	}

	async function saveSprint() {
		if (!editingSprint?.name?.trim()) return;
		saving = true;
		try {
			if (editingSprint.id) {
				const updated = await sprintsApi.update(editingSprint.id, editingSprint);
				sprints = sprints.map(s => s.id === editingSprint!.id ? updated : s);
			} else {
				const created = await sprintsApi.create(editingSprint);
				sprints = [created, ...sprints];
			}
			showModal = false;
			editingSprint = null;
		} catch (e) {
			console.error('Error saving sprint:', e);
		}
		saving = false;
	}

	async function startSprint(sprint: Sprint) {
		try {
			const updated = await sprintsApi.start(sprint.id);
			sprints = sprints.map(s => s.id === sprint.id ? updated : s);
		} catch (e) {
			console.error('Error starting sprint:', e);
		}
	}

	async function completeSprint(sprint: Sprint) {
		if (!confirm(`Завершить спринт "${sprint.name}"? Velocity будет рассчитан автоматически.`)) return;
		try {
			const updated = await sprintsApi.complete(sprint.id);
			sprints = sprints.map(s => s.id === sprint.id ? updated : s);
		} catch (e) {
			console.error('Error completing sprint:', e);
		}
	}

	async function deleteSprint(sprint: Sprint) {
		if (!confirm(`Удалить спринт "${sprint.name}"? Задачи останутся без спринта.`)) return;
		try {
			await sprintsApi.delete(sprint.id);
			sprints = sprints.filter(s => s.id !== sprint.id);
		} catch (e) {
			console.error('Error deleting sprint:', e);
		}
	}

	let filteredSprints = $derived.by(() => {
		let result = sprints;
		if (filterStatus) {
			result = result.filter(s => s.status === filterStatus);
		}
		if (filterProject) {
			result = result.filter(s => s.project_id === filterProject);
		}
		return [...result].sort((a, b) => {
			// Active first, then by start_date desc
			if (a.status === 'active' && b.status !== 'active') return -1;
			if (b.status === 'active' && a.status !== 'active') return 1;
			return new Date(b.start_date).getTime() - new Date(a.start_date).getTime();
		});
	});

	function getProjectName(id: string | undefined): string {
		if (!id) return '';
		const project = projects.find(p => p.id === id);
		return project?.name || '';
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
	}

	function getStatusBadge(status: string) {
		switch (status) {
			case 'active':
				return { label: 'Активный', class: 'bg-green-100 text-green-700' };
			case 'completed':
				return { label: 'Завершён', class: 'bg-gray-100 text-gray-600' };
			default:
				return { label: 'Планирование', class: 'bg-blue-100 text-blue-700' };
		}
	}

	function getDaysLeft(endDate: string): number {
		const end = new Date(endDate);
		const today = new Date();
		const diff = end.getTime() - today.getTime();
		return Math.ceil(diff / (1000 * 60 * 60 * 24));
	}
</script>

<svelte:head>
	<title>Спринты | EKF Hub</title>
</svelte:head>

<div class="p-6 max-w-7xl mx-auto">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Спринты</h1>
			<p class="text-sm text-gray-500 mt-1">Управление Scrum спринтами</p>
		</div>
		<button
			onclick={openNewSprint}
			class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 flex items-center gap-2"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
			</svg>
			Новый спринт
		</button>
	</div>

	<!-- Filters -->
	<div class="flex gap-3 mb-6">
		<select bind:value={filterStatus} class="px-3 py-2 border border-gray-200 rounded-lg text-sm">
			<option value="">Все статусы</option>
			<option value="planning">Планирование</option>
			<option value="active">Активные</option>
			<option value="completed">Завершённые</option>
		</select>
		<select bind:value={filterProject} class="px-3 py-2 border border-gray-200 rounded-lg text-sm">
			<option value="">Все проекты</option>
			{#each projects as project}
				<option value={project.id}>{project.name}</option>
			{/each}
		</select>
	</div>

	{#if loading}
		<div class="flex items-center justify-center h-48">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
		</div>
	{:else if error}
		<div class="text-center py-8 text-red-600">{error}</div>
	{:else if filteredSprints.length === 0}
		<div class="text-center py-12 text-gray-500">
			<svg class="w-16 h-16 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z" />
			</svg>
			<p>Спринтов пока нет</p>
			<button onclick={openNewSprint} class="mt-4 text-ekf-red hover:underline">Создать первый спринт</button>
		</div>
	{:else}
		<div class="grid gap-4">
			{#each filteredSprints as sprint}
				{@const status = getStatusBadge(sprint.status)}
				{@const daysLeft = getDaysLeft(sprint.end_date)}
				<div class="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-shadow">
					<div class="flex items-start justify-between">
						<div class="flex-1">
							<div class="flex items-center gap-3 mb-2">
								<h3 class="font-semibold text-gray-900">{sprint.name}</h3>
								<span class="px-2 py-0.5 rounded text-xs font-medium {status.class}">{status.label}</span>
								{#if sprint.project_id}
									<span class="text-xs text-gray-500">{getProjectName(sprint.project_id)}</span>
								{/if}
							</div>

							{#if sprint.goal}
								<p class="text-sm text-gray-600 mb-3">{sprint.goal}</p>
							{/if}

							<div class="flex items-center gap-6 text-sm">
								<div class="flex items-center gap-1 text-gray-500">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
									</svg>
									{formatDate(sprint.start_date)} — {formatDate(sprint.end_date)}
								</div>

								{#if sprint.status === 'active' && daysLeft >= 0}
									<span class="text-blue-600 font-medium">{daysLeft} дн. осталось</span>
								{:else if sprint.status === 'active' && daysLeft < 0}
									<span class="text-red-600 font-medium">Просрочен на {Math.abs(daysLeft)} дн.</span>
								{/if}

								<div class="flex items-center gap-1">
									<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
									</svg>
									<span>{sprint.tasks_done || 0}/{sprint.tasks_count || 0} задач</span>
								</div>

								<div class="flex items-center gap-1">
									<span class="text-indigo-600 font-medium">{sprint.completed_points || 0}/{sprint.total_points || 0} SP</span>
								</div>

								{#if sprint.status === 'completed' && sprint.velocity}
									<div class="flex items-center gap-1 text-green-600">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
										</svg>
										Velocity: {sprint.velocity}
									</div>
								{/if}
							</div>

							<!-- Progress bar -->
							{#if sprint.tasks_count && sprint.tasks_count > 0}
								<div class="mt-3 w-full bg-gray-100 rounded-full h-2">
									<div
										class="h-2 rounded-full transition-all {sprint.status === 'completed' ? 'bg-green-500' : 'bg-ekf-red'}"
										style="width: {sprint.progress || 0}%"
									></div>
								</div>
							{/if}
						</div>

						<!-- Actions -->
						<div class="flex items-center gap-2 ml-4">
							{#if sprint.status === 'planning'}
								<button
									onclick={() => startSprint(sprint)}
									class="px-3 py-1.5 bg-green-600 text-white text-sm rounded hover:bg-green-700"
								>
									Старт
								</button>
							{:else if sprint.status === 'active'}
								<button
									onclick={() => completeSprint(sprint)}
									class="px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
								>
									Завершить
								</button>
							{/if}

							<button
								onclick={() => openEditSprint(sprint)}
								class="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded"
								title="Редактировать"
							>
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
								</svg>
							</button>

							<button
								onclick={() => deleteSprint(sprint)}
								class="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded"
								title="Удалить"
							>
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
								</svg>
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Sprint Modal -->
{#if showModal && editingSprint}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onclick={() => showModal = false}>
		<div class="bg-white rounded-lg shadow-xl w-full max-w-md" onclick={(e) => e.stopPropagation()}>
			<div class="p-4 border-b flex items-center justify-between">
				<h2 class="font-bold text-gray-900">{editingSprint.id ? 'Редактировать спринт' : 'Новый спринт'}</h2>
				<button onclick={() => showModal = false} class="p-1 hover:bg-gray-100 rounded">
					<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<form onsubmit={(e) => { e.preventDefault(); saveSprint(); }} class="p-4 space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Название</label>
					<input
						type="text"
						bind:value={editingSprint.name}
						class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red"
						placeholder="Sprint 1"
						required
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Цель спринта</label>
					<textarea
						bind:value={editingSprint.goal}
						rows="2"
						class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red"
						placeholder="Что мы хотим достичь в этом спринте?"
					></textarea>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Начало</label>
						<input
							type="date"
							bind:value={editingSprint.start_date}
							class="w-full px-3 py-2 border border-gray-200 rounded-lg"
							required
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Окончание</label>
						<input
							type="date"
							bind:value={editingSprint.end_date}
							class="w-full px-3 py-2 border border-gray-200 rounded-lg"
							required
						/>
					</div>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Проект</label>
					<select bind:value={editingSprint.project_id} class="w-full px-3 py-2 border border-gray-200 rounded-lg">
						<option value="">Без проекта</option>
						{#each projects as project}
							<option value={project.id}>{project.name}</option>
						{/each}
					</select>
				</div>

				<div class="flex justify-end gap-2 pt-4 border-t">
					<button
						type="button"
						onclick={() => showModal = false}
						class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
					>
						Отмена
					</button>
					<button
						type="submit"
						disabled={saving}
						class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 disabled:opacity-50"
					>
						{saving ? 'Сохранение...' : (editingSprint.id ? 'Сохранить' : 'Создать')}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
