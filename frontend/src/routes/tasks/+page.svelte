<script lang="ts">
	import { onMount } from 'svelte';
	import { subordinates, user } from '$lib/stores/auth';
	import { tasks as tasksApi, projects as projectsApi } from '$lib/api/client';
	import type { KanbanBoard, Task, Project } from '$lib/api/client';

	// View state
	type ViewMode = 'board' | 'list' | 'table';
	let viewMode: ViewMode = $state('board');

	// Data
	let kanban: KanbanBoard | null = $state(null);
	let tasksList: Task[] = $state([]);
	let projects: Project[] = $state([]);
	let loading = $state(true);

	// Filters
	let filterEmployee = $state('');
	let filterProject = $state('');
	let filterStatus = $state('');
	let filterPriority = $state('');
	let searchQuery = $state('');

	// Drag state
	let draggedTask: Task | null = $state(null);

	// Modal state
	let showCreateModal = $state(false);
	let showDetailModal = $state(false);
	let selectedTask: Task | null = $state(null);
	let newTask = $state({
		title: '',
		description: '',
		priority: 'medium',
		status: 'todo',
		assignee_id: '',
		project_id: '',
		due_date: '',
		flag_color: ''
	});

	// Columns configuration
	const columns = [
		{ key: 'backlog', label: 'Бэклог', color: 'bg-gray-100', icon: '📋' },
		{ key: 'todo', label: 'К выполнению', color: 'bg-blue-50', icon: '📝' },
		{ key: 'in_progress', label: 'В работе', color: 'bg-yellow-50', icon: '🔄' },
		{ key: 'review', label: 'На проверке', color: 'bg-purple-50', icon: '👀' },
		{ key: 'done', label: 'Выполнено', color: 'bg-green-50', icon: '✅' }
	];

	const priorities = [
		{ key: 'high', label: 'Высокий', color: 'text-red-600 bg-red-50' },
		{ key: 'medium', label: 'Средний', color: 'text-yellow-600 bg-yellow-50' },
		{ key: 'low', label: 'Низкий', color: 'text-green-600 bg-green-50' }
	];

	const flagColors = [
		{ key: '', label: 'Без флага', color: 'bg-gray-200' },
		{ key: 'red', label: 'Красный', color: 'bg-red-500' },
		{ key: 'orange', label: 'Оранжевый', color: 'bg-orange-500' },
		{ key: 'yellow', label: 'Жёлтый', color: 'bg-yellow-500' },
		{ key: 'green', label: 'Зелёный', color: 'bg-green-500' },
		{ key: 'blue', label: 'Синий', color: 'bg-blue-500' },
		{ key: 'purple', label: 'Фиолетовый', color: 'bg-purple-500' }
	];

	// Filtered tasks for list/table view
	const filteredTasks = $derived(() => {
		let result = tasksList;
		if (filterEmployee) {
			result = result.filter(t => t.assignee_id === filterEmployee);
		}
		if (filterProject) {
			result = result.filter(t => t.project_id === filterProject);
		}
		if (filterStatus) {
			result = result.filter(t => t.status === filterStatus);
		}
		if (filterPriority) {
			result = result.filter(t => t.priority === filterPriority);
		}
		if (searchQuery) {
			const query = searchQuery.toLowerCase();
			result = result.filter(t =>
				t.title.toLowerCase().includes(query) ||
				t.description?.toLowerCase().includes(query)
			);
		}
		return result;
	});

	onMount(async () => {
		await Promise.all([loadData(), loadProjects()]);
	});

	async function loadData() {
		loading = true;
		try {
			const params: { assignee_id?: string; project_id?: string } = {};
			if (filterEmployee) params.assignee_id = filterEmployee;
			if (filterProject) params.project_id = filterProject;

			if (viewMode === 'board') {
				kanban = await tasksApi.getKanban(params);
			} else {
				tasksList = await tasksApi.list({ ...params, status: filterStatus || undefined });
			}
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function loadProjects() {
		try {
			projects = await projectsApi.list();
		} catch (e) {
			console.error(e);
		}
	}

	function getTasksByStatus(status: string): Task[] {
		if (!kanban) return [];
		return (kanban as any)[status] || [];
	}

	function handleDragStart(task: Task) {
		draggedTask = task;
	}

	function handleDragEnd() {
		draggedTask = null;
	}

	async function handleDrop(newStatus: string) {
		if (!draggedTask || draggedTask.status === newStatus) {
			draggedTask = null;
			return;
		}

		const taskId = draggedTask.id;
		draggedTask = null;

		try {
			await tasksApi.moveKanban(taskId, newStatus);
			await loadData();
		} catch (e) {
			console.error(e);
		}
	}

	async function createTask() {
		if (!newTask.title.trim()) return;

		try {
			await tasksApi.create({
				title: newTask.title,
				description: newTask.description,
				priority: newTask.priority,
				status: newTask.status,
				assignee_id: newTask.assignee_id || undefined,
				project_id: newTask.project_id || undefined,
				due_date: newTask.due_date || undefined,
				flag_color: newTask.flag_color || undefined
			});
			showCreateModal = false;
			resetNewTask();
			await loadData();
		} catch (e) {
			console.error(e);
		}
	}

	async function updateTask(task: Task) {
		try {
			await tasksApi.update(task.id, task);
			await loadData();
		} catch (e) {
			console.error(e);
		}
	}

	async function deleteTask(taskId: string) {
		if (!confirm('Удалить задачу?')) return;
		try {
			await tasksApi.delete(taskId);
			showDetailModal = false;
			selectedTask = null;
			await loadData();
		} catch (e) {
			console.error(e);
		}
	}

	function resetNewTask() {
		newTask = {
			title: '',
			description: '',
			priority: 'medium',
			status: 'todo',
			assignee_id: '',
			project_id: '',
			due_date: '',
			flag_color: ''
		};
	}

	function openTaskDetail(task: Task) {
		selectedTask = { ...task };
		showDetailModal = true;
	}

	function getPriorityClass(priority: string): string {
		switch (priority) {
			case 'high': return 'border-l-red-500';
			case 'medium': return 'border-l-yellow-500';
			case 'low': return 'border-l-green-500';
			default: return 'border-l-gray-300';
		}
	}

	function getPriorityLabel(priority: string): string {
		return priorities.find(p => p.key === priority)?.label || priority;
	}

	function getStatusLabel(status: string): string {
		return columns.find(c => c.key === status)?.label || status;
	}

	function formatDate(dateStr: string | undefined): string {
		if (!dateStr) return '';
		return new Date(dateStr).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
	}

	function isOverdue(task: Task): boolean {
		if (!task.due_date || task.status === 'done') return false;
		return new Date(task.due_date) < new Date();
	}

	$effect(() => {
		loadData();
	});
</script>

<svelte:head>
	<title>Задачи - EKF Team Hub</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex items-center justify-between flex-wrap gap-4">
		<h1 class="text-2xl font-bold text-gray-900">Задачи</h1>
		<div class="flex items-center gap-3">
			<!-- View toggle -->
			<div class="flex bg-gray-100 rounded-lg p-1">
				{#each [
					{ key: 'board', label: 'Доска', icon: '▤' },
					{ key: 'list', label: 'Список', icon: '≡' },
					{ key: 'table', label: 'Таблица', icon: '⊞' }
				] as view}
					<button
						onclick={() => { viewMode = view.key as ViewMode; loadData(); }}
						class="px-3 py-1.5 rounded text-sm font-medium transition-colors
							{viewMode === view.key ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-700'}"
					>
						<span class="mr-1">{view.icon}</span>
						{view.label}
					</button>
				{/each}
			</div>
			<button
				onclick={() => showCreateModal = true}
				class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors flex items-center gap-2"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Новая задача
			</button>
		</div>
	</div>

	<!-- Filters -->
	<div class="bg-white rounded-lg shadow-sm border p-4">
		<div class="flex flex-wrap gap-4">
			<!-- Search -->
			<div class="flex-1 min-w-[200px]">
				<div class="relative">
					<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Поиск задач..."
						class="w-full pl-10 pr-4 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
					/>
				</div>
			</div>

			<!-- Employee filter -->
			<select
				bind:value={filterEmployee}
				class="border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent"
			>
				<option value="">Все исполнители</option>
				{#each $subordinates as emp}
					<option value={emp.id}>{emp.name}</option>
				{/each}
			</select>

			<!-- Project filter -->
			<select
				bind:value={filterProject}
				class="border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent"
			>
				<option value="">Все проекты</option>
				{#each projects as project}
					<option value={project.id}>{project.name}</option>
				{/each}
			</select>

			<!-- Priority filter -->
			<select
				bind:value={filterPriority}
				class="border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent"
			>
				<option value="">Любой приоритет</option>
				{#each priorities as p}
					<option value={p.key}>{p.label}</option>
				{/each}
			</select>

			{#if viewMode !== 'board'}
				<!-- Status filter (only for list/table) -->
				<select
					bind:value={filterStatus}
					class="border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent"
				>
					<option value="">Все статусы</option>
					{#each columns as col}
						<option value={col.key}>{col.label}</option>
					{/each}
				</select>
			{/if}
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
		</div>
	{:else if viewMode === 'board'}
		<!-- Board View -->
		<div class="grid grid-cols-5 gap-4 min-h-[600px]">
			{#each columns as column}
				<div
					class="rounded-xl p-3 {column.color} min-h-full"
					ondragover={(e) => e.preventDefault()}
					ondrop={() => handleDrop(column.key)}
				>
					<div class="flex items-center justify-between mb-3 sticky top-0">
						<div class="flex items-center gap-2">
							<span>{column.icon}</span>
							<h3 class="font-semibold text-gray-900">{column.label}</h3>
						</div>
						<span class="text-sm text-gray-500 bg-white px-2 py-0.5 rounded-full shadow-sm">
							{getTasksByStatus(column.key).length}
						</span>
					</div>

					<div class="space-y-2">
						{#each getTasksByStatus(column.key) as task}
							<div
								class="bg-white rounded-lg p-3 shadow-sm border-l-4 cursor-pointer hover:shadow-md transition-all
									{getPriorityClass(task.priority || 'medium')}
									{draggedTask?.id === task.id ? 'opacity-50' : ''}"
								draggable="true"
								ondragstart={() => handleDragStart(task)}
								ondragend={handleDragEnd}
								onclick={() => openTaskDetail(task)}
								role="button"
								tabindex="0"
							>
								<!-- Flag -->
								{#if task.flag_color}
									<div class="w-full h-1 rounded-full mb-2 {flagColors.find(f => f.key === task.flag_color)?.color || 'bg-gray-200'}"></div>
								{/if}

								<h4 class="font-medium text-gray-900 text-sm mb-1 line-clamp-2">{task.title}</h4>

								{#if task.description}
									<p class="text-xs text-gray-500 line-clamp-2 mb-2">{task.description}</p>
								{/if}

								<div class="flex items-center justify-between text-xs">
									<div class="flex items-center gap-2">
										{#if task.due_date}
											<span class="flex items-center gap-1 {isOverdue(task) ? 'text-red-600' : 'text-gray-500'}">
												<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
												</svg>
												{formatDate(task.due_date)}
											</span>
										{/if}
									</div>
									{#if task.assignee_name || task.assignee?.name}
										<span class="bg-gray-100 text-gray-600 px-1.5 py-0.5 rounded text-xs truncate max-w-[80px]">
											{task.assignee_name || task.assignee?.name}
										</span>
									{/if}
								</div>

								{#if task.project?.name || task.tags?.length}
									<div class="flex flex-wrap gap-1 mt-2">
										{#if task.project?.name}
											<span class="text-xs bg-blue-50 text-blue-600 px-1.5 py-0.5 rounded">
												{task.project.name}
											</span>
										{/if}
										{#each task.tags || [] as tag}
											<span class="text-xs px-1.5 py-0.5 rounded" style="background-color: {tag.color}20; color: {tag.color}">
												{tag.name}
											</span>
										{/each}
									</div>
								{/if}
							</div>
						{/each}

						{#if getTasksByStatus(column.key).length === 0}
							<div class="text-center py-8 text-gray-400 text-sm">
								Нет задач
							</div>
						{/if}
					</div>
				</div>
			{/each}
		</div>

	{:else if viewMode === 'list'}
		<!-- List View -->
		<div class="bg-white rounded-lg shadow-sm border divide-y">
			{#each filteredTasks() as task}
				<div
					class="p-4 hover:bg-gray-50 cursor-pointer flex items-center gap-4"
					onclick={() => openTaskDetail(task)}
					role="button"
					tabindex="0"
				>
					<div class="w-1 h-10 rounded-full {getPriorityClass(task.priority || 'medium').replace('border-l-', 'bg-')}"></div>

					<div class="flex-1 min-w-0">
						<div class="flex items-center gap-2">
							{#if task.flag_color}
								<div class="w-2 h-2 rounded-full {flagColors.find(f => f.key === task.flag_color)?.color || ''}"></div>
							{/if}
							<h4 class="font-medium text-gray-900 truncate">{task.title}</h4>
						</div>
						{#if task.description}
							<p class="text-sm text-gray-500 truncate">{task.description}</p>
						{/if}
					</div>

					<div class="flex items-center gap-4 text-sm">
						<span class="px-2 py-1 rounded-full text-xs {columns.find(c => c.key === task.status)?.color || 'bg-gray-100'} text-gray-700">
							{getStatusLabel(task.status)}
						</span>

						{#if task.due_date}
							<span class="flex items-center gap-1 {isOverdue(task) ? 'text-red-600' : 'text-gray-500'}">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
								</svg>
								{formatDate(task.due_date)}
							</span>
						{/if}

						{#if task.assignee_name}
							<span class="text-gray-600">{task.assignee_name}</span>
						{/if}
					</div>
				</div>
			{:else}
				<div class="p-12 text-center text-gray-400">
					Задачи не найдены
				</div>
			{/each}
		</div>

	{:else if viewMode === 'table'}
		<!-- Table View -->
		<div class="bg-white rounded-lg shadow-sm border overflow-x-auto">
			<table class="w-full">
				<thead class="bg-gray-50 border-b">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Задача</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Статус</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Приоритет</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Исполнитель</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Проект</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Срок</th>
					</tr>
				</thead>
				<tbody class="divide-y">
					{#each filteredTasks() as task}
						<tr
							class="hover:bg-gray-50 cursor-pointer"
							onclick={() => openTaskDetail(task)}
						>
							<td class="px-4 py-3">
								<div class="flex items-center gap-2">
									{#if task.flag_color}
										<div class="w-2 h-2 rounded-full {flagColors.find(f => f.key === task.flag_color)?.color || ''}"></div>
									{/if}
									<div>
										<p class="font-medium text-gray-900">{task.title}</p>
										{#if task.description}
											<p class="text-sm text-gray-500 truncate max-w-xs">{task.description}</p>
										{/if}
									</div>
								</div>
							</td>
							<td class="px-4 py-3">
								<span class="px-2 py-1 rounded-full text-xs {columns.find(c => c.key === task.status)?.color || 'bg-gray-100'} text-gray-700">
									{getStatusLabel(task.status)}
								</span>
							</td>
							<td class="px-4 py-3">
								<span class="px-2 py-1 rounded text-xs {priorities.find(p => p.key === task.priority)?.color || ''}">
									{getPriorityLabel(task.priority || 'medium')}
								</span>
							</td>
							<td class="px-4 py-3 text-sm text-gray-600">
								{task.assignee_name || task.assignee?.name || '-'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-600">
								{task.project?.name || '-'}
							</td>
							<td class="px-4 py-3 text-sm {isOverdue(task) ? 'text-red-600' : 'text-gray-600'}">
								{task.due_date ? formatDate(task.due_date) : '-'}
							</td>
						</tr>
					{:else}
						<tr>
							<td colspan="6" class="px-4 py-12 text-center text-gray-400">
								Задачи не найдены
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<!-- Create Task Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onclick={() => showCreateModal = false}>
		<div class="bg-white rounded-xl shadow-xl p-6 w-full max-w-lg mx-4" onclick={(e) => e.stopPropagation()}>
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-xl font-bold text-gray-900">Новая задача</h2>
				<button onclick={() => showCreateModal = false} class="text-gray-400 hover:text-gray-600">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Название *</label>
					<input
						type="text"
						bind:value={newTask.title}
						class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						placeholder="Что нужно сделать?"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Описание</label>
					<textarea
						bind:value={newTask.description}
						rows="3"
						class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						placeholder="Подробности задачи..."
					></textarea>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Статус</label>
						<select bind:value={newTask.status} class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent">
							{#each columns as col}
								<option value={col.key}>{col.label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Приоритет</label>
						<select bind:value={newTask.priority} class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent">
							{#each priorities as p}
								<option value={p.key}>{p.label}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Исполнитель</label>
						<select bind:value={newTask.assignee_id} class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent">
							<option value="">Не назначен</option>
							{#each $subordinates as emp}
								<option value={emp.id}>{emp.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Проект</label>
						<select bind:value={newTask.project_id} class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent">
							<option value="">Без проекта</option>
							{#each projects as project}
								<option value={project.id}>{project.name}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Срок</label>
						<input
							type="date"
							bind:value={newTask.due_date}
							class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Флаг</label>
						<div class="flex gap-2">
							{#each flagColors as flag}
								<button
									type="button"
									onclick={() => newTask.flag_color = flag.key}
									class="w-6 h-6 rounded-full {flag.color} {newTask.flag_color === flag.key ? 'ring-2 ring-offset-2 ring-ekf-red' : ''}"
									title={flag.label}
								></button>
							{/each}
						</div>
					</div>
				</div>
			</div>

			<div class="flex gap-3 mt-6">
				<button
					onclick={() => showCreateModal = false}
					class="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
				>
					Отмена
				</button>
				<button
					onclick={createTask}
					disabled={!newTask.title.trim()}
					class="flex-1 px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors"
				>
					Создать
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Task Detail Modal -->
{#if showDetailModal && selectedTask}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onclick={() => showDetailModal = false}>
		<div class="bg-white rounded-xl shadow-xl w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()}>
			<div class="p-6 border-b sticky top-0 bg-white">
				<div class="flex items-start justify-between">
					<div class="flex-1">
						<input
							type="text"
							bind:value={selectedTask.title}
							class="text-xl font-bold text-gray-900 w-full border-0 p-0 focus:ring-0"
						/>
						{#if selectedTask.flag_color}
							<div class="w-16 h-1 rounded-full mt-2 {flagColors.find(f => f.key === selectedTask.flag_color)?.color || ''}"></div>
						{/if}
					</div>
					<button onclick={() => showDetailModal = false} class="text-gray-400 hover:text-gray-600 ml-4">
						<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>

			<div class="p-6 space-y-6">
				<!-- Status and Priority -->
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-500 mb-2">Статус</label>
						<select
							bind:value={selectedTask.status}
							onchange={() => updateTask(selectedTask!)}
							class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						>
							{#each columns as col}
								<option value={col.key}>{col.icon} {col.label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-500 mb-2">Приоритет</label>
						<select
							bind:value={selectedTask.priority}
							onchange={() => updateTask(selectedTask!)}
							class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						>
							{#each priorities as p}
								<option value={p.key}>{p.label}</option>
							{/each}
						</select>
					</div>
				</div>

				<!-- Description -->
				<div>
					<label class="block text-sm font-medium text-gray-500 mb-2">Описание</label>
					<textarea
						bind:value={selectedTask.description}
						onblur={() => updateTask(selectedTask!)}
						rows="4"
						class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						placeholder="Добавьте описание..."
					></textarea>
				</div>

				<!-- Assignee and Due Date -->
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-500 mb-2">Исполнитель</label>
						<select
							bind:value={selectedTask.assignee_id}
							onchange={() => updateTask(selectedTask!)}
							class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						>
							<option value="">Не назначен</option>
							{#each $subordinates as emp}
								<option value={emp.id}>{emp.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-500 mb-2">Срок</label>
						<input
							type="date"
							bind:value={selectedTask.due_date}
							onchange={() => updateTask(selectedTask!)}
							class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						/>
					</div>
				</div>

				<!-- Flag Color -->
				<div>
					<label class="block text-sm font-medium text-gray-500 mb-2">Флаг</label>
					<div class="flex gap-3">
						{#each flagColors as flag}
							<button
								type="button"
								onclick={() => { selectedTask!.flag_color = flag.key; updateTask(selectedTask!); }}
								class="w-8 h-8 rounded-full {flag.color} {selectedTask.flag_color === flag.key ? 'ring-2 ring-offset-2 ring-ekf-red' : ''} transition-transform hover:scale-110"
								title={flag.label}
							></button>
						{/each}
					</div>
				</div>

				<!-- Project -->
				<div>
					<label class="block text-sm font-medium text-gray-500 mb-2">Проект</label>
					<select
						bind:value={selectedTask.project_id}
						onchange={() => updateTask(selectedTask!)}
						class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
					>
						<option value="">Без проекта</option>
						{#each projects as project}
							<option value={project.id}>{project.name}</option>
						{/each}
					</select>
				</div>
			</div>

			<!-- Actions -->
			<div class="p-6 border-t bg-gray-50 flex justify-between">
				<button
					onclick={() => deleteTask(selectedTask!.id)}
					class="px-4 py-2 text-red-600 hover:bg-red-50 rounded-lg transition-colors"
				>
					Удалить задачу
				</button>
				<button
					onclick={() => showDetailModal = false}
					class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors"
				>
					Готово
				</button>
			</div>
		</div>
	</div>
{/if}
