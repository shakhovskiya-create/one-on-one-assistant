<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { user, subordinates } from '$lib/stores/auth';
	import type { TaskDependency } from '$lib/api/client';

	// Types
	interface Task {
		id: string;
		title: string;
		description?: string;
		status: 'backlog' | 'todo' | 'in_progress' | 'review' | 'done';
		priority?: number;
		story_points?: number;
		assignee_id?: string;
		assignee?: any;
		project_id?: string;
		project?: any;
		tags?: { name: string; color: string }[];
		due_date?: string;
		parent_id?: string;
		sprint?: string;
		created_at: string;
	}

	interface Project {
		id: string;
		name: string;
	}

	// State
	let tasks: Task[] = $state([]);
	let projects: Project[] = $state([]);
	let employees: any[] = $state([]);
	let loading = $state(true);

	// View mode
	let viewMode = $state<'list' | 'kanban'>('list');

	// Filters
	let filterProject = $state('');
	let filterAssignee = $state('');
	let filterStatus = $state('');
	let searchQuery = $state('');

	// Modal state
	let showTaskModal = $state(false);
	let editingTask: Partial<Task> | null = $state(null);
	let selectedTask: Task | null = $state(null);

	// Dependencies state
	let taskDependencies: TaskDependency[] = $state([]);
	let taskDependents: TaskDependency[] = $state([]);
	let taskBlockers: Task[] = $state([]);
	let isTaskBlocked = $state(false);
	let showDependencyPicker = $state(false);
	let dependencySearch = $state('');

	// Status columns for Kanban with WIP limits
	const statusColumns = [
		{ id: 'backlog', label: 'Backlog', color: 'bg-gray-100', wipLimit: 0 },
		{ id: 'todo', label: 'К выполнению', color: 'bg-blue-50', wipLimit: 10 },
		{ id: 'in_progress', label: 'В работе', color: 'bg-yellow-50', wipLimit: 5 },
		{ id: 'review', label: 'На проверке', color: 'bg-purple-50', wipLimit: 3 },
		{ id: 'done', label: 'Готово', color: 'bg-green-50', wipLimit: 0 }
	];

	const priorityLabels: Record<number, { label: string; color: string }> = {
		1: { label: 'Критический', color: 'text-red-600 bg-red-50' },
		2: { label: 'Высокий', color: 'text-orange-600 bg-orange-50' },
		3: { label: 'Средний', color: 'text-yellow-600 bg-yellow-50' },
		4: { label: 'Низкий', color: 'text-blue-600 bg-blue-50' },
		5: { label: 'Минимальный', color: 'text-gray-600 bg-gray-50' }
	};

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		loading = true;
		try {
			const [tasksRes, projectsRes, employeesRes] = await Promise.all([
				api.tasks.list(),
				api.projects.list().catch(() => []),
				api.employees.list().catch(() => [])
			]);
			tasks = tasksRes || [];
			projects = projectsRes || [];
			employees = employeesRes || [];
		} catch (e) {
			console.error('Error loading data:', e);
		}
		loading = false;
	}

	function openNewTask() {
		editingTask = {
			title: '',
			description: '',
			status: 'todo',
			priority: 3,
			story_points: undefined,
			assignee_id: $user?.id || '',
			project_id: filterProject || '',
			due_date: ''
		};
		// Reset dependencies for new task
		taskDependencies = [];
		taskDependents = [];
		taskBlockers = [];
		isTaskBlocked = false;
		showTaskModal = true;
	}

	// Story points options (Fibonacci-like)
	const storyPointsOptions = [1, 2, 3, 5, 8, 13, 21];

	async function openEditTask(task: Task) {
		editingTask = { ...task };
		showTaskModal = true;
		// Load dependencies for existing task
		if (task.id) {
			await loadTaskDependencies(task.id);
		}
	}

	async function saveTask() {
		if (!editingTask?.title?.trim()) return;
		try {
			// Clean up empty string values to undefined
			const taskData = {
				...editingTask,
				assignee_id: editingTask.assignee_id || undefined,
				project_id: editingTask.project_id || undefined,
				due_date: editingTask.due_date || undefined,
				priority: Number(editingTask.priority) || 3,
				story_points: editingTask.story_points ? Number(editingTask.story_points) : undefined
			};

			if (editingTask.id) {
				const updated = await api.tasks.update(editingTask.id, taskData);
				tasks = tasks.map(t => t.id === editingTask!.id ? updated : t);
			} else {
				const created = await api.tasks.create(taskData);
				tasks = [created, ...tasks];
			}
			showTaskModal = false;
			editingTask = null;
		} catch (e) {
			console.error('Error saving task:', e);
		}
	}

	// Dependencies functions
	async function loadTaskDependencies(taskId: string) {
		try {
			const result = await api.tasks.getDependencies(taskId);
			taskDependencies = result.dependencies || [];
			taskDependents = result.dependents || [];

			// Check if blocked
			const blockedResult = await api.tasks.isBlocked(taskId);
			isTaskBlocked = blockedResult.blocked;
			taskBlockers = blockedResult.blockers || [];
		} catch (e) {
			console.error('Error loading dependencies:', e);
			taskDependencies = [];
			taskDependents = [];
		}
	}

	async function addDependency(dependsOnTaskId: string) {
		if (!editingTask?.id) return;
		try {
			const dep = await api.tasks.addDependency(editingTask.id, dependsOnTaskId);
			taskDependencies = [...taskDependencies, dep];
			showDependencyPicker = false;
			dependencySearch = '';
			// Reload blocked status
			const blockedResult = await api.tasks.isBlocked(editingTask.id);
			isTaskBlocked = blockedResult.blocked;
			taskBlockers = blockedResult.blockers || [];
		} catch (e: any) {
			alert(e.message || 'Не удалось добавить зависимость');
		}
	}

	async function removeDependency(depId: string) {
		if (!editingTask?.id) return;
		try {
			await api.tasks.removeDependency(editingTask.id, depId);
			taskDependencies = taskDependencies.filter(d => d.id !== depId);
			// Reload blocked status
			const blockedResult = await api.tasks.isBlocked(editingTask.id);
			isTaskBlocked = blockedResult.blocked;
			taskBlockers = blockedResult.blockers || [];
		} catch (e) {
			console.error('Error removing dependency:', e);
		}
	}

	// Filter tasks for dependency picker (exclude current task and already added)
	$effect(() => {
		// This will reactively filter when dependencySearch changes
	});

	function getAvailableDependencies(): Task[] {
		const currentId = editingTask?.id;
		const existingIds = new Set(taskDependencies.map(d => d.depends_on_task_id));
		return tasks.filter(t =>
			t.id !== currentId &&
			!existingIds.has(t.id) &&
			(dependencySearch === '' || t.title.toLowerCase().includes(dependencySearch.toLowerCase()))
		);
	}

	async function updateTaskStatus(task: Task, newStatus: string) {
		try {
			const updated = await api.tasks.update(task.id, { status: newStatus });
			tasks = tasks.map(t => t.id === task.id ? updated : t);
		} catch (e) {
			console.error('Error updating task:', e);
		}
	}

	async function deleteTask(id: string) {
		if (!confirm('Удалить задачу?')) return;
		try {
			await api.tasks.delete(id);
			tasks = tasks.filter(t => t.id !== id);
		} catch (e) {
			console.error('Error deleting task:', e);
		}
	}

	// Filtered tasks
	let filteredTasks = $derived.by(() => {
		let result = tasks;

		if (filterProject) {
			result = result.filter(t => t.project_id === filterProject);
		}
		if (filterAssignee) {
			result = result.filter(t => t.assignee_id === filterAssignee);
		}
		if (filterStatus) {
			result = result.filter(t => t.status === filterStatus);
		}
		if (searchQuery) {
			const q = searchQuery.toLowerCase();
			result = result.filter(t =>
				t.title.toLowerCase().includes(q) ||
				t.description?.toLowerCase().includes(q)
			);
		}

		return result.sort((a, b) => {
			// Sort by priority first, then by created_at
			if ((a.priority || 3) !== (b.priority || 3)) {
				return (a.priority || 3) - (b.priority || 3);
			}
			return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
		});
	});

	function getTasksByStatus(status: string) {
		return filteredTasks.filter(t => t.status === status);
	}

	function getStoryPointsByStatus(status: string): number {
		return getTasksByStatus(status).reduce((sum, t) => sum + (t.story_points || 0), 0);
	}

	function getTotalStoryPoints(): number {
		return filteredTasks.reduce((sum, t) => sum + (t.story_points || 0), 0);
	}

	function isWipLimitExceeded(status: string): boolean {
		const column = statusColumns.find(c => c.id === status);
		if (!column || column.wipLimit === 0) return false;
		return getTasksByStatus(status).length > column.wipLimit;
	}

	function getCompletedTasksCount(): number {
		return filteredTasks.filter(t => t.status === 'done').length;
	}

	function getInProgressTasksCount(): number {
		return filteredTasks.filter(t => t.status === 'in_progress').length;
	}

	// Quick add task in column
	let quickAddColumn: string | null = $state(null);
	let quickAddTitle = $state('');

	async function quickAddTask(status: string) {
		if (!quickAddTitle.trim()) {
			quickAddColumn = null;
			return;
		}
		try {
			const taskData = {
				title: quickAddTitle,
				status,
				priority: 3,
				assignee_id: $user?.id || undefined,
				project_id: filterProject || undefined
			};
			const created = await api.tasks.create(taskData);
			tasks = [created, ...tasks];
			quickAddTitle = '';
			quickAddColumn = null;
		} catch (e) {
			console.error('Error creating task:', e);
		}
	}

	function getEmployeeName(id: string): string {
		if (!id) return 'Не назначен';
		const emp = employees.find(e => e.id === id);
		return emp?.name || 'Неизвестный';
	}

	function getProjectName(id: string): string {
		if (!id) return '';
		const proj = projects.find(p => p.id === id);
		return proj?.name || '';
	}

	function formatDate(dateStr: string | undefined): string {
		if (!dateStr) return '';
		const date = new Date(dateStr);
		const today = new Date();
		const tomorrow = new Date(today);
		tomorrow.setDate(tomorrow.getDate() + 1);

		if (date.toDateString() === today.toDateString()) return 'Сегодня';
		if (date.toDateString() === tomorrow.toDateString()) return 'Завтра';
		return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
	}

	function isOverdue(dateStr: string | undefined): boolean {
		if (!dateStr) return false;
		return new Date(dateStr) < new Date();
	}

	// Drag and drop for Kanban
	let draggedTask: Task | null = $state(null);

	function handleDragStart(e: DragEvent, task: Task) {
		draggedTask = task;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
		}
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		if (e.dataTransfer) {
			e.dataTransfer.dropEffect = 'move';
		}
	}

	function handleDrop(e: DragEvent, newStatus: string) {
		e.preventDefault();
		if (draggedTask && draggedTask.status !== newStatus) {
			updateTaskStatus(draggedTask, newStatus);
		}
		draggedTask = null;
	}
</script>

<svelte:head>
	<title>Задачи - EKF Hub</title>
</svelte:head>

<div class="space-y-4">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-gray-900">Задачи</h1>
			<div class="flex items-center gap-3 text-sm text-gray-500">
				<span>{filteredTasks.length} задач</span>
				{#if getTotalStoryPoints() > 0}
					<span class="text-indigo-600 font-medium">{getTotalStoryPoints()} SP</span>
				{/if}
				<span class="flex items-center gap-1">
					<span class="w-2 h-2 rounded-full bg-yellow-400"></span>
					{getInProgressTasksCount()} в работе
				</span>
				<span class="flex items-center gap-1">
					<span class="w-2 h-2 rounded-full bg-green-500"></span>
					{getCompletedTasksCount()} выполнено
				</span>
			</div>
		</div>
		<button
			onclick={openNewTask}
			class="px-3 py-1.5 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors flex items-center gap-1.5 text-sm"
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
			</svg>
			Создать
		</button>
	</div>

	<!-- Filters & View Toggle -->
	<div class="bg-white rounded-lg shadow-sm p-3 flex flex-wrap items-center gap-3">
		<!-- Search -->
		<div class="relative flex-1 min-w-[200px] max-w-xs">
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Поиск задач..."
				class="w-full pl-8 pr-3 py-1.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-1 focus:ring-ekf-red focus:border-ekf-red"
			/>
			<svg class="w-4 h-4 absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
			</svg>
		</div>

		<!-- Filters -->
		<select bind:value={filterProject} class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-1 focus:ring-ekf-red">
			<option value="">Все проекты</option>
			{#each projects as project}
				<option value={project.id}>{project.name}</option>
			{/each}
		</select>

		<select bind:value={filterAssignee} class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-1 focus:ring-ekf-red">
			<option value="">Все исполнители</option>
			<option value={$user?.id}>Я</option>
			{#each $subordinates as sub}
				<option value={sub.id}>{sub.name}</option>
			{/each}
		</select>

		<select bind:value={filterStatus} class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-1 focus:ring-ekf-red">
			<option value="">Все статусы</option>
			{#each statusColumns as col}
				<option value={col.id}>{col.label}</option>
			{/each}
		</select>

		<!-- View Toggle -->
		<div class="flex rounded-lg border border-gray-200 overflow-hidden ml-auto">
			<button
				onclick={() => viewMode = 'list'}
				class="px-3 py-1.5 text-sm transition-colors {viewMode === 'list' ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}"
			>
				Список
			</button>
			<button
				onclick={() => viewMode = 'kanban'}
				class="px-3 py-1.5 text-sm transition-colors {viewMode === 'kanban' ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}"
			>
				Kanban
			</button>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center h-48">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
		</div>
	{:else if viewMode === 'list'}
		<!-- List View -->
		<div class="bg-white rounded-lg shadow-sm overflow-hidden">
			{#if filteredTasks.length === 0}
				<div class="text-center py-12">
					<svg class="w-12 h-12 mx-auto text-gray-300 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
					</svg>
					<p class="text-gray-500 mb-3">Нет задач</p>
					<button onclick={openNewTask} class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 text-sm">
						Создать первую задачу
					</button>
				</div>
			{:else}
				<table class="w-full text-sm">
					<thead class="bg-gray-50 border-b">
						<tr>
							<th class="px-4 py-2 text-left font-medium text-gray-500">Задача</th>
							<th class="px-4 py-2 text-left font-medium text-gray-500 w-32">Статус</th>
							<th class="px-4 py-2 text-left font-medium text-gray-500 w-20">Приор.</th>
							<th class="px-4 py-2 text-center font-medium text-gray-500 w-12" title="Story Points">SP</th>
							<th class="px-4 py-2 text-left font-medium text-gray-500 w-36">Исполнитель</th>
							<th class="px-4 py-2 text-left font-medium text-gray-500 w-24">Срок</th>
							<th class="px-4 py-2 w-16"></th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100">
						{#each filteredTasks as task (task.id)}
							<tr class="hover:bg-gray-50 transition-colors">
								<td class="px-4 py-2">
									<button onclick={() => openEditTask(task)} class="text-left group">
										<div class="font-medium text-gray-900 group-hover:text-ekf-red">{task.title}</div>
										{#if task.project_id}
											<div class="text-xs text-gray-400">{getProjectName(task.project_id)}</div>
										{/if}
									</button>
								</td>
								<td class="px-4 py-2">
									<select
										value={task.status}
										onchange={(e) => updateTaskStatus(task, (e.target as HTMLSelectElement).value)}
										class="w-full px-2 py-1 text-xs border border-gray-200 rounded focus:outline-none focus:ring-1 focus:ring-ekf-red"
									>
										{#each statusColumns as col}
											<option value={col.id}>{col.label}</option>
										{/each}
									</select>
								</td>
								<td class="px-4 py-2">
									<span class="px-2 py-0.5 text-xs rounded {priorityLabels[task.priority || 3].color}">
										P{task.priority || 3}
									</span>
								</td>
								<td class="px-4 py-2 text-center">
									{#if task.story_points}
										<span class="px-1.5 py-0.5 text-xs rounded bg-indigo-50 text-indigo-600 font-medium">
											{task.story_points}
										</span>
									{:else}
										<span class="text-gray-300">—</span>
									{/if}
								</td>
								<td class="px-4 py-2 text-gray-600 text-xs">{getEmployeeName(task.assignee_id || '')}</td>
								<td class="px-4 py-2">
									{#if task.due_date}
										<span class="{isOverdue(task.due_date) && task.status !== 'done' ? 'text-red-600 font-medium' : 'text-gray-600'}">
											{formatDate(task.due_date)}
										</span>
									{:else}
										<span class="text-gray-300">—</span>
									{/if}
								</td>
								<td class="px-4 py-2">
									<button
										onclick={() => deleteTask(task.id)}
										class="p-1 text-gray-400 hover:text-red-600 rounded transition-colors"
										title="Удалить"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
										</svg>
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>
	{:else}
		<!-- Kanban View -->
		<div class="flex gap-3 overflow-x-auto pb-4" style="min-height: calc(100vh - 280px);">
			{#each statusColumns as column}
				<div
					class="flex-shrink-0 w-72 flex flex-col rounded-lg {column.color} {isWipLimitExceeded(column.id) ? 'ring-2 ring-red-400' : ''}"
					ondragover={handleDragOver}
					ondrop={(e) => handleDrop(e, column.id)}
				>
					<div class="p-3 font-medium text-gray-700 border-b border-gray-200/50">
						<div class="flex items-center justify-between">
							<span>{column.label}</span>
							<div class="flex items-center gap-1.5">
								{#if getStoryPointsByStatus(column.id) > 0}
									<span class="text-xs bg-indigo-100 text-indigo-600 px-1.5 py-0.5 rounded font-medium" title="Story Points">
										{getStoryPointsByStatus(column.id)} SP
									</span>
								{/if}
								<span class="text-xs px-1.5 py-0.5 rounded {isWipLimitExceeded(column.id) ? 'bg-red-100 text-red-600 font-bold' : 'bg-white/80'}">
									{getTasksByStatus(column.id).length}{#if column.wipLimit > 0}/{column.wipLimit}{/if}
								</span>
							</div>
						</div>
						{#if isWipLimitExceeded(column.id)}
							<div class="mt-1 text-xs text-red-600 flex items-center gap-1">
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
								</svg>
								WIP лимит превышен!
							</div>
						{/if}
						<!-- Quick Add Button -->
						<button
							onclick={() => quickAddColumn = quickAddColumn === column.id ? null : column.id}
							class="mt-2 w-full py-1.5 text-xs text-gray-500 hover:text-gray-700 hover:bg-white/50 rounded border border-dashed border-gray-300 flex items-center justify-center gap-1 transition-colors"
						>
							<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
							</svg>
							Добавить
						</button>
						{#if quickAddColumn === column.id}
							<div class="mt-2">
								<input
									type="text"
									bind:value={quickAddTitle}
									placeholder="Название задачи"
									class="w-full px-2 py-1.5 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-ekf-red"
									onkeydown={(e) => e.key === 'Enter' && quickAddTask(column.id)}
								/>
								<div class="flex gap-1 mt-1">
									<button
										onclick={() => quickAddTask(column.id)}
										class="flex-1 py-1 bg-ekf-red text-white text-xs rounded hover:bg-red-700"
									>
										Создать
									</button>
									<button
										onclick={() => { quickAddColumn = null; quickAddTitle = ''; }}
										class="px-2 py-1 text-gray-500 text-xs hover:bg-gray-100 rounded"
									>
										Отмена
									</button>
								</div>
							</div>
						{/if}
					</div>
					<div class="flex-1 p-2 space-y-2 overflow-y-auto">
						{#each getTasksByStatus(column.id) as task (task.id)}
							<div
								class="bg-white rounded-lg shadow-sm p-3 cursor-pointer hover:shadow-md transition-shadow border-l-4 {task.priority === 1 ? 'border-red-500' : task.priority === 2 ? 'border-orange-500' : task.priority === 4 ? 'border-blue-300' : task.priority === 5 ? 'border-gray-300' : 'border-yellow-400'}"
								draggable="true"
								ondragstart={(e) => handleDragStart(e, task)}
								onclick={() => openEditTask(task)}
							>
								<!-- Task ID / Project -->
								{#if task.project_id}
									<div class="text-xs text-gray-400 mb-1 flex items-center gap-1">
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
										</svg>
										{getProjectName(task.project_id)}
									</div>
								{/if}
								<div class="font-medium text-gray-900 text-sm mb-1">{task.title}</div>
								{#if task.description}
									<p class="text-xs text-gray-500 line-clamp-2 mb-2">{task.description}</p>
								{/if}
								<div class="flex items-center justify-between text-xs">
									<div class="flex items-center gap-1.5">
										<span class="px-1.5 py-0.5 rounded {priorityLabels[task.priority || 3].color}">
											P{task.priority || 3}
										</span>
										{#if task.story_points}
											<span class="px-1.5 py-0.5 rounded bg-indigo-50 text-indigo-600 font-medium" title="Story Points">
												{task.story_points}
											</span>
										{/if}
									</div>
									<div class="flex items-center gap-2">
										{#if task.due_date}
											<span class="{isOverdue(task.due_date) && task.status !== 'done' ? 'text-red-600 font-medium' : 'text-gray-400'}" title="Срок">
												{formatDate(task.due_date)}
											</span>
										{/if}
										{#if task.assignee_id}
											<div class="w-5 h-5 rounded-full bg-gradient-to-br from-gray-400 to-gray-500 text-white flex items-center justify-center text-[10px] font-medium shadow-sm" title={getEmployeeName(task.assignee_id)}>
												{getEmployeeName(task.assignee_id).charAt(0)}
											</div>
										{/if}
									</div>
								</div>
							</div>
						{/each}
						{#if getTasksByStatus(column.id).length === 0}
							<div class="text-center py-8 text-gray-400 text-sm">
								Перетащите задачу сюда
							</div>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Task Modal -->
{#if showTaskModal && editingTask}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onclick={() => showTaskModal = false}>
		<div class="bg-white rounded-lg shadow-xl w-full max-w-md" onclick={(e) => e.stopPropagation()}>
			<div class="p-4 border-b flex items-center justify-between">
				<h2 class="font-bold text-gray-900">{editingTask.id ? 'Редактировать задачу' : 'Новая задача'}</h2>
				<button onclick={() => showTaskModal = false} class="p-1 hover:bg-gray-100 rounded">
					<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>
			<form onsubmit={(e) => { e.preventDefault(); saveTask(); }} class="p-4 space-y-3">
				<div>
					<label class="block text-xs font-medium text-gray-500 mb-1">Название *</label>
					<input
						type="text"
						bind:value={editingTask.title}
						required
						class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-1 focus:ring-ekf-red"
						placeholder="Что нужно сделать?"
					/>
				</div>
				<div>
					<label class="block text-xs font-medium text-gray-500 mb-1">Описание</label>
					<textarea
						bind:value={editingTask.description}
						rows="2"
						class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm resize-none focus:outline-none focus:ring-1 focus:ring-ekf-red"
						placeholder="Подробности..."
					></textarea>
				</div>
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">Проект</label>
						<select bind:value={editingTask.project_id} class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm">
							<option value="">Без проекта</option>
							{#each projects as project}
								<option value={project.id}>{project.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">Исполнитель</label>
						<select bind:value={editingTask.assignee_id} class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm">
							<option value="">Не назначен</option>
							<option value={$user?.id}>Я ({$user?.name})</option>
							{#each $subordinates as sub}
								{#if sub.id !== $user?.id}
									<option value={sub.id}>{sub.name}</option>
								{/if}
							{/each}
						</select>
					</div>
				</div>
				<div class="grid grid-cols-4 gap-3">
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">Статус</label>
						<select bind:value={editingTask.status} class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm">
							{#each statusColumns as col}
								<option value={col.id}>{col.label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">Приоритет</label>
						<select bind:value={editingTask.priority} class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm">
							<option value={1}>P1 - Критический</option>
							<option value={2}>P2 - Высокий</option>
							<option value={3}>P3 - Средний</option>
							<option value={4}>P4 - Низкий</option>
							<option value={5}>P5 - Минимальный</option>
						</select>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">Story Points</label>
						<select bind:value={editingTask.story_points} class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm">
							<option value={undefined}>—</option>
							{#each storyPointsOptions as sp}
								<option value={sp}>{sp}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">Срок</label>
						<input
							type="date"
							bind:value={editingTask.due_date}
							class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm"
						/>
					</div>
				</div>

				<!-- Dependencies Section (only for existing tasks) -->
				{#if editingTask.id}
					<div class="border-t pt-3 mt-3">
						<div class="flex items-center justify-between mb-2">
							<label class="block text-xs font-medium text-gray-500">Зависимости (блокирующие задачи)</label>
							<button
								type="button"
								onclick={() => showDependencyPicker = !showDependencyPicker}
								class="text-xs text-ekf-red hover:text-red-700"
							>
								+ Добавить
							</button>
						</div>

						<!-- Blocked warning -->
						{#if isTaskBlocked}
							<div class="mb-2 p-2 bg-yellow-50 border border-yellow-200 rounded-lg text-xs text-yellow-800">
								<span class="font-medium">⚠️ Задача заблокирована.</span> Ожидает завершения:
								<ul class="mt-1 list-disc list-inside">
									{#each taskBlockers as blocker}
										<li>{blocker.title}</li>
									{/each}
								</ul>
							</div>
						{/if}

						<!-- Current dependencies -->
						{#if taskDependencies.length > 0}
							<div class="space-y-1 mb-2">
								{#each taskDependencies as dep}
									<div class="flex items-center justify-between p-2 bg-gray-50 rounded text-xs">
										<div class="flex items-center gap-2">
											<span class={dep.depends_on_task?.status === 'done' ? 'text-green-600' : 'text-orange-600'}>
												{dep.depends_on_task?.status === 'done' ? '✓' : '○'}
											</span>
											<span class="truncate">{dep.depends_on_task?.title || 'Неизвестно'}</span>
										</div>
										<button
											type="button"
											onclick={() => removeDependency(dep.id)}
											class="text-gray-400 hover:text-red-500"
										>
											×
										</button>
									</div>
								{/each}
							</div>
						{:else}
							<p class="text-xs text-gray-400 mb-2">Нет зависимостей</p>
						{/if}

						<!-- Dependency picker -->
						{#if showDependencyPicker}
							<div class="border rounded-lg p-2 bg-white">
								<input
									type="text"
									bind:value={dependencySearch}
									placeholder="Поиск задачи..."
									class="w-full px-2 py-1 border border-gray-200 rounded text-xs mb-2"
								/>
								<div class="max-h-32 overflow-y-auto space-y-1">
									{#each getAvailableDependencies().slice(0, 10) as task}
										<button
											type="button"
											onclick={() => addDependency(task.id)}
											class="w-full text-left px-2 py-1 hover:bg-gray-100 rounded text-xs truncate"
										>
											{task.title}
										</button>
									{:else}
										<p class="text-xs text-gray-400 p-2">Нет доступных задач</p>
									{/each}
								</div>
							</div>
						{/if}

						<!-- Dependents info -->
						{#if taskDependents.length > 0}
							<div class="mt-3 pt-2 border-t">
								<p class="text-xs text-gray-500 mb-1">Блокирует другие задачи ({taskDependents.length}):</p>
								<div class="text-xs text-gray-400">
									{#each taskDependents.slice(0, 3) as dep}
										<span class="mr-2">• {dep.depends_on_task?.title}</span>
									{/each}
									{#if taskDependents.length > 3}
										<span>и ещё {taskDependents.length - 3}...</span>
									{/if}
								</div>
							</div>
						{/if}
					</div>
				{/if}

				<div class="flex justify-end gap-2 pt-2">
					<button
						type="button"
						onclick={() => showTaskModal = false}
						class="px-4 py-2 text-gray-600 hover:text-gray-900 text-sm"
					>
						Отмена
					</button>
					<button
						type="submit"
						class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 text-sm"
					>
						{editingTask.id ? 'Сохранить' : 'Создать'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
