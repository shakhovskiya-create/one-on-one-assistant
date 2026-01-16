<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { user, subordinates } from '$lib/stores/auth';

	// Types
	interface Task {
		id: string;
		title: string;
		description?: string;
		status: 'backlog' | 'todo' | 'in_progress' | 'review' | 'done';
		priority?: 'low' | 'medium' | 'high' | 'urgent';
		assignee_id?: string;
		assignee?: any;
		project_id?: string;
		project?: any;
		epic_id?: string;
		epic?: string;
		tags?: string[];
		due_date?: string;
		parent_id?: string;
		created_at: string;
		comments?: Comment[];
	}

	interface Comment {
		id: string;
		task_id: string;
		user_id: string;
		user?: any;
		content: string;
		created_at: string;
	}

	interface Project {
		id: string;
		name: string;
		color?: string;
	}

	// State
	let tasks: Task[] = $state([]);
	let projects: Project[] = $state([]);
	let employees: any[] = $state([]);
	let loading = $state(true);
	let view = $state<'board' | 'list' | 'timeline'>('board');

	// Filters
	let filterProject = $state('all');
	let filterAssignee = $state('all');
	let filterPriority = $state('all');
	let filterEpic = $state('all');
	let searchQuery = $state('');

	// Modals
	let showCreateModal = $state(false);
	let showDetailModal = $state(false);
	let selectedTask: Task | null = $state(null);

	// New task form
	let newTask = $state({
		title: '',
		description: '',
		status: 'backlog' as const,
		priority: 'medium' as const,
		assignee_id: '',
		project_id: '',
		epic: '',
		tags: [] as string[],
		due_date: ''
	});

	// Comment
	let newComment = $state('');

	// Available tags
	const availableTags = ['bug', 'feature', 'improvement', 'urgent', 'blocked', 'needs-review', 'design', 'frontend', 'backend', 'devops'];

	// Columns for kanban
	const columns = [
		{ id: 'backlog', label: 'Backlog', color: 'bg-gray-100' },
		{ id: 'todo', label: 'К выполнению', color: 'bg-blue-50' },
		{ id: 'in_progress', label: 'В работе', color: 'bg-yellow-50' },
		{ id: 'review', label: 'На проверке', color: 'bg-purple-50' },
		{ id: 'done', label: 'Готово', color: 'bg-green-50' }
	];

	// Epics (derived from tasks)
	let epics = $derived([...new Set(tasks.map(t => t.epic).filter(Boolean))]);

	// Filtered tasks
	let filteredTasks = $derived(() => {
		return tasks.filter(t => {
			if (filterProject !== 'all' && t.project_id !== filterProject) return false;
			if (filterAssignee !== 'all' && t.assignee_id !== filterAssignee) return false;
			if (filterPriority !== 'all' && t.priority !== filterPriority) return false;
			if (filterEpic !== 'all' && t.epic !== filterEpic) return false;
			if (searchQuery) {
				const q = searchQuery.toLowerCase();
				const inTitle = t.title.toLowerCase().includes(q);
				const inDesc = t.description?.toLowerCase().includes(q);
				const inTags = t.tags?.some(tag => tag.toLowerCase().includes(q));
				if (!inTitle && !inDesc && !inTags) return false;
			}
			return true;
		});
	});

	// Tasks without epic (derived)
	let noEpicTasks = $derived(() => filteredTasks().filter(t => !t.epic));

	// Tasks by column
	function getTasksByStatus(status: string) {
		return filteredTasks().filter(t => t.status === status);
	}

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

	async function createTask() {
		try {
			const task = await api.tasks.create(newTask);
			tasks = [task, ...tasks];
			showCreateModal = false;
			resetNewTask();
		} catch (e) {
			console.error('Error creating task:', e);
		}
	}

	async function updateTask(task: Task) {
		try {
			const updated = await api.tasks.update(task.id, task);
			tasks = tasks.map(t => t.id === task.id ? updated : t);
		} catch (e) {
			console.error('Error updating task:', e);
		}
	}

	async function moveTask(taskId: string, newStatus: string) {
		const task = tasks.find(t => t.id === taskId);
		if (task) {
			task.status = newStatus as Task['status'];
			await updateTask(task);
		}
	}

	async function addComment() {
		if (!selectedTask || !newComment.trim()) return;
		const comment: Comment = {
			id: Date.now().toString(),
			task_id: selectedTask.id,
			user_id: $user?.id || '',
			user: $user,
			content: newComment,
			created_at: new Date().toISOString()
		};
		selectedTask.comments = [...(selectedTask.comments || []), comment];
		newComment = '';
	}

	function resetNewTask() {
		newTask = {
			title: '',
			description: '',
			status: 'backlog',
			priority: 'medium',
			assignee_id: '',
			project_id: '',
			epic: '',
			tags: [],
			due_date: ''
		};
	}

	function openTaskDetail(task: Task) {
		selectedTask = { ...task, comments: task.comments || [] };
		showDetailModal = true;
	}

	function toggleTag(tag: string) {
		if (newTask.tags.includes(tag)) {
			newTask.tags = newTask.tags.filter(t => t !== tag);
		} else {
			newTask.tags = [...newTask.tags, tag];
		}
	}

	function getAssigneeName(id: string) {
		const emp = employees.find(e => e.id === id);
		return emp?.name || 'Не назначен';
	}

	function getProjectName(id: string) {
		const proj = projects.find(p => p.id === id);
		return proj?.name || '';
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString('ru-RU', {
			day: 'numeric',
			month: 'short'
		});
	}

	function formatDateTime(dateStr: string) {
		return new Date(dateStr).toLocaleString('ru-RU', {
			day: 'numeric',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function parseContent(content: string) {
		return content.replace(/@(\w+)/g, '<span class="text-blue-600 font-medium">@$1</span>');
	}

	// Drag and drop
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

	function handleDrop(e: DragEvent, status: string) {
		e.preventDefault();
		if (draggedTask && draggedTask.status !== status) {
			moveTask(draggedTask.id, status);
		}
		draggedTask = null;
	}

	function getPriorityColor(priority: string) {
		switch (priority) {
			case 'urgent': return 'bg-red-500';
			case 'high': return 'bg-orange-500';
			case 'medium': return 'bg-yellow-500';
			case 'low': return 'bg-green-500';
			default: return 'bg-gray-400';
		}
	}
</script>

<svelte:head>
	<title>Задачи - EKF Team Hub</title>
</svelte:head>

<div class="h-full flex flex-col -m-6">
	<!-- Header -->
	<div class="bg-white border-b px-6 py-4">
		<div class="flex items-center justify-between mb-4">
			<div class="flex items-center gap-4">
				<h1 class="text-2xl font-bold text-gray-900">Задачи</h1>
				<div class="flex bg-gray-100 rounded-lg p-1">
					<button
						onclick={() => view = 'board'}
						class="px-3 py-1.5 text-sm rounded-md transition-colors {view === 'board' ? 'bg-white shadow text-gray-900' : 'text-gray-600 hover:text-gray-900'}"
					>
						Доска
					</button>
					<button
						onclick={() => view = 'list'}
						class="px-3 py-1.5 text-sm rounded-md transition-colors {view === 'list' ? 'bg-white shadow text-gray-900' : 'text-gray-600 hover:text-gray-900'}"
					>
						Список
					</button>
					<button
						onclick={() => view = 'timeline'}
						class="px-3 py-1.5 text-sm rounded-md transition-colors {view === 'timeline' ? 'bg-white shadow text-gray-900' : 'text-gray-600 hover:text-gray-900'}"
					>
						Эпики
					</button>
				</div>
			</div>
			<button
				onclick={() => showCreateModal = true}
				class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors flex items-center gap-2"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
				</svg>
				Создать задачу
			</button>
		</div>

		<!-- Filters -->
		<div class="flex flex-wrap gap-3">
			<div class="relative">
				<svg class="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
				<input
					type="text"
					placeholder="Поиск задач..."
					bind:value={searchQuery}
					class="pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent w-64"
				/>
			</div>
			<select bind:value={filterProject} class="px-3 py-2 border border-gray-200 rounded-lg">
				<option value="all">Все проекты</option>
				{#each projects as project}
					<option value={project.id}>{project.name}</option>
				{/each}
			</select>
			<select bind:value={filterEpic} class="px-3 py-2 border border-gray-200 rounded-lg">
				<option value="all">Все эпики</option>
				{#each epics as epic}
					<option value={epic}>{epic}</option>
				{/each}
			</select>
			<select bind:value={filterAssignee} class="px-3 py-2 border border-gray-200 rounded-lg">
				<option value="all">Все исполнители</option>
				<option value={$user?.id}>Мои задачи</option>
				{#each employees as emp}
					<option value={emp.id}>{emp.name}</option>
				{/each}
			</select>
			<select bind:value={filterPriority} class="px-3 py-2 border border-gray-200 rounded-lg">
				<option value="all">Все приоритеты</option>
				<option value="urgent">Срочный</option>
				<option value="high">Высокий</option>
				<option value="medium">Средний</option>
				<option value="low">Низкий</option>
			</select>
		</div>
	</div>

	<!-- Content -->
	{#if loading}
		<div class="flex-1 flex items-center justify-center">
			<div class="text-gray-500">Загрузка задач...</div>
		</div>
	{:else if view === 'board'}
		<div class="flex-1 overflow-x-auto p-6">
			<div class="flex gap-4 h-full min-w-max">
				{#each columns as column}
					<div
						class="w-80 flex flex-col rounded-xl {column.color}"
						ondragover={handleDragOver}
						ondrop={(e) => handleDrop(e, column.id)}
					>
						<div class="p-4 font-semibold text-gray-700 flex items-center justify-between">
							<span>{column.label}</span>
							<span class="text-sm bg-white/50 px-2 py-0.5 rounded-full">{getTasksByStatus(column.id).length}</span>
						</div>
						<div class="flex-1 overflow-y-auto p-2 space-y-2">
							{#each getTasksByStatus(column.id) as task (task.id)}
								<div
									class="bg-white rounded-lg shadow-sm p-4 cursor-pointer hover:shadow-md transition-shadow"
									draggable="true"
									ondragstart={(e) => handleDragStart(e, task)}
									onclick={() => openTaskDetail(task)}
								>
									<div class="flex items-start justify-between mb-2">
										<div class="w-2 h-2 rounded-full {getPriorityColor(task.priority || 'medium')}"></div>
										{#if task.epic}
											<span class="text-xs bg-purple-100 text-purple-700 px-2 py-0.5 rounded-full">{task.epic}</span>
										{/if}
									</div>
									<h4 class="font-medium text-gray-900 mb-2">{task.title}</h4>
									{#if task.description}
										<p class="text-sm text-gray-500 mb-3 line-clamp-2">{task.description}</p>
									{/if}
									{#if task.tags && task.tags.length > 0}
										<div class="flex flex-wrap gap-1 mb-3">
											{#each task.tags as tag}
												<span class="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded">{tag}</span>
											{/each}
										</div>
									{/if}
									<div class="flex items-center justify-between text-sm">
										<div class="flex items-center gap-2">
											{#if task.assignee_id}
												<div class="w-6 h-6 rounded-full bg-gray-200 flex items-center justify-center text-xs font-medium">
													{getAssigneeName(task.assignee_id).charAt(0)}
												</div>
											{/if}
											{#if task.project_id}
												<span class="text-gray-500">{getProjectName(task.project_id)}</span>
											{/if}
										</div>
										{#if task.due_date}
											<span class="text-gray-400">{formatDate(task.due_date)}</span>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{:else if view === 'list'}
		<div class="flex-1 overflow-y-auto p-6">
			<div class="bg-white rounded-xl shadow-sm overflow-hidden">
				<table class="w-full">
					<thead class="bg-gray-50 border-b">
						<tr>
							<th class="px-4 py-3 text-left text-sm font-semibold text-gray-900">Задача</th>
							<th class="px-4 py-3 text-left text-sm font-semibold text-gray-900">Статус</th>
							<th class="px-4 py-3 text-left text-sm font-semibold text-gray-900">Приоритет</th>
							<th class="px-4 py-3 text-left text-sm font-semibold text-gray-900">Исполнитель</th>
							<th class="px-4 py-3 text-left text-sm font-semibold text-gray-900">Проект</th>
							<th class="px-4 py-3 text-left text-sm font-semibold text-gray-900">Срок</th>
						</tr>
					</thead>
					<tbody class="divide-y">
						{#each filteredTasks() as task (task.id)}
							<tr class="hover:bg-gray-50 cursor-pointer" onclick={() => openTaskDetail(task)}>
								<td class="px-4 py-3">
									<div class="font-medium text-gray-900">{task.title}</div>
									{#if task.tags && task.tags.length > 0}
										<div class="flex gap-1 mt-1">
											{#each task.tags.slice(0, 3) as tag}
												<span class="text-xs bg-gray-100 text-gray-600 px-1.5 py-0.5 rounded">{tag}</span>
											{/each}
										</div>
									{/if}
								</td>
								<td class="px-4 py-3">
									<span class="px-2 py-1 rounded text-xs font-medium
										{task.status === 'done' ? 'bg-green-100 text-green-700' : ''}
										{task.status === 'in_progress' ? 'bg-yellow-100 text-yellow-700' : ''}
										{task.status === 'todo' ? 'bg-blue-100 text-blue-700' : ''}
										{task.status === 'backlog' ? 'bg-gray-100 text-gray-700' : ''}
										{task.status === 'review' ? 'bg-purple-100 text-purple-700' : ''}
									">
										{columns.find(c => c.id === task.status)?.label || task.status}
									</span>
								</td>
								<td class="px-4 py-3">
									<div class="flex items-center gap-2">
										<div class="w-2 h-2 rounded-full {getPriorityColor(task.priority || 'medium')}"></div>
										<span class="text-sm text-gray-600 capitalize">{task.priority || 'medium'}</span>
									</div>
								</td>
								<td class="px-4 py-3 text-sm text-gray-600">{task.assignee_id ? getAssigneeName(task.assignee_id) : '-'}</td>
								<td class="px-4 py-3 text-sm text-gray-600">{task.project_id ? getProjectName(task.project_id) : '-'}</td>
								<td class="px-4 py-3 text-sm text-gray-500">{task.due_date ? formatDate(task.due_date) : '-'}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{:else}
		<div class="flex-1 overflow-y-auto p-6">
			<div class="space-y-4">
				{#each epics as epic}
					<div class="bg-white rounded-xl shadow-sm overflow-hidden">
						<div class="p-4 bg-purple-50 border-b border-purple-100">
							<h3 class="font-semibold text-purple-900">{epic}</h3>
						</div>
						<div class="p-4 space-y-2">
							{#each filteredTasks().filter(t => t.epic === epic) as task (task.id)}
								<div class="flex items-center gap-4 p-3 rounded-lg hover:bg-gray-50 cursor-pointer" onclick={() => openTaskDetail(task)}>
									<div class="w-2 h-2 rounded-full {getPriorityColor(task.priority || 'medium')}"></div>
									<div class="flex-1">
										<div class="font-medium text-gray-900">{task.title}</div>
										<div class="text-sm text-gray-500">{getAssigneeName(task.assignee_id || '')}</div>
									</div>
									<span class="px-2 py-1 rounded text-xs font-medium
										{task.status === 'done' ? 'bg-green-100 text-green-700' : ''}
										{task.status === 'in_progress' ? 'bg-yellow-100 text-yellow-700' : ''}
										{task.status === 'todo' ? 'bg-blue-100 text-blue-700' : ''}
										{task.status === 'backlog' ? 'bg-gray-100 text-gray-700' : ''}
										{task.status === 'review' ? 'bg-purple-100 text-purple-700' : ''}
									">
										{columns.find(c => c.id === task.status)?.label || task.status}
									</span>
								</div>
							{/each}
						</div>
					</div>
				{/each}
				{#if noEpicTasks().length > 0}
					<div class="bg-white rounded-xl shadow-sm overflow-hidden">
						<div class="p-4 bg-gray-50 border-b">
							<h3 class="font-semibold text-gray-700">Без эпика</h3>
						</div>
						<div class="p-4 space-y-2">
							{#each noEpicTasks() as task (task.id)}
								<div class="flex items-center gap-4 p-3 rounded-lg hover:bg-gray-50 cursor-pointer" onclick={() => openTaskDetail(task)}>
									<div class="w-2 h-2 rounded-full {getPriorityColor(task.priority || 'medium')}"></div>
									<div class="flex-1">
										<div class="font-medium text-gray-900">{task.title}</div>
										<div class="text-sm text-gray-500">{getAssigneeName(task.assignee_id || '')}</div>
									</div>
									<span class="px-2 py-1 rounded text-xs font-medium
										{task.status === 'done' ? 'bg-green-100 text-green-700' : ''}
										{task.status === 'in_progress' ? 'bg-yellow-100 text-yellow-700' : ''}
										{task.status === 'todo' ? 'bg-blue-100 text-blue-700' : ''}
										{task.status === 'backlog' ? 'bg-gray-100 text-gray-700' : ''}
										{task.status === 'review' ? 'bg-purple-100 text-purple-700' : ''}
									">
										{columns.find(c => c.id === task.status)?.label || task.status}
									</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<!-- Create Task Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onclick={() => showCreateModal = false}>
		<div class="bg-white rounded-xl shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()}>
			<div class="p-6 border-b">
				<h2 class="text-xl font-bold text-gray-900">Создать задачу</h2>
			</div>
			<form onsubmit={(e) => { e.preventDefault(); createTask(); }} class="p-6 space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Название</label>
					<input type="text" bind:value={newTask.title} required class="w-full px-3 py-2 border border-gray-300 rounded-lg" placeholder="Введите название задачи" />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Описание</label>
					<textarea bind:value={newTask.description} rows="3" class="w-full px-3 py-2 border border-gray-300 rounded-lg" placeholder="Описание задачи... (поддерживается @упоминание)"></textarea>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Проект</label>
						<select bind:value={newTask.project_id} class="w-full px-3 py-2 border border-gray-300 rounded-lg">
							<option value="">Выберите проект</option>
							{#each projects as project}
								<option value={project.id}>{project.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Эпик</label>
						<input type="text" bind:value={newTask.epic} list="epics-list" class="w-full px-3 py-2 border border-gray-300 rounded-lg" placeholder="Название эпика" />
						<datalist id="epics-list">
							{#each epics as epic}
								<option value={epic}></option>
							{/each}
						</datalist>
					</div>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Исполнитель</label>
						<select bind:value={newTask.assignee_id} class="w-full px-3 py-2 border border-gray-300 rounded-lg">
							<option value="">Не назначен</option>
							<option value={$user?.id}>Я ({$user?.name})</option>
							{#each employees as emp}
								<option value={emp.id}>{emp.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Приоритет</label>
						<select bind:value={newTask.priority} class="w-full px-3 py-2 border border-gray-300 rounded-lg">
							<option value="low">Низкий</option>
							<option value="medium">Средний</option>
							<option value="high">Высокий</option>
							<option value="urgent">Срочный</option>
						</select>
					</div>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Срок выполнения</label>
					<input type="date" bind:value={newTask.due_date} class="w-full px-3 py-2 border border-gray-300 rounded-lg" />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-2">Теги</label>
					<div class="flex flex-wrap gap-2">
						{#each availableTags as tag}
							<button type="button" onclick={() => toggleTag(tag)} class="px-3 py-1 rounded-full text-sm transition-colors {newTask.tags.includes(tag) ? 'bg-ekf-red text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}">
								{tag}
							</button>
						{/each}
					</div>
				</div>
				<div class="flex justify-end gap-3 pt-4 border-t">
					<button type="button" onclick={() => showCreateModal = false} class="px-4 py-2 text-gray-600 hover:text-gray-900">Отмена</button>
					<button type="submit" class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700">Создать</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Task Detail Modal -->
{#if showDetailModal && selectedTask}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onclick={() => showDetailModal = false}>
		<div class="bg-white rounded-xl shadow-xl w-full max-w-3xl max-h-[90vh] overflow-hidden flex flex-col" onclick={(e) => e.stopPropagation()}>
			<div class="p-6 border-b flex items-start justify-between">
				<div class="flex-1">
					<div class="flex items-center gap-3 mb-2">
						<div class="w-3 h-3 rounded-full {getPriorityColor(selectedTask.priority || 'medium')}"></div>
						{#if selectedTask.epic}
							<span class="text-sm bg-purple-100 text-purple-700 px-2 py-0.5 rounded-full">{selectedTask.epic}</span>
						{/if}
						<span class="px-2 py-1 rounded text-xs font-medium
							{selectedTask.status === 'done' ? 'bg-green-100 text-green-700' : ''}
							{selectedTask.status === 'in_progress' ? 'bg-yellow-100 text-yellow-700' : ''}
							{selectedTask.status === 'todo' ? 'bg-blue-100 text-blue-700' : ''}
							{selectedTask.status === 'backlog' ? 'bg-gray-100 text-gray-700' : ''}
							{selectedTask.status === 'review' ? 'bg-purple-100 text-purple-700' : ''}
						">{columns.find(c => c.id === selectedTask.status)?.label || selectedTask.status}</span>
					</div>
					<h2 class="text-xl font-bold text-gray-900">{selectedTask.title}</h2>
				</div>
				<button onclick={() => showDetailModal = false} class="p-2 hover:bg-gray-100 rounded-lg">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>
			<div class="flex-1 overflow-y-auto">
				<div class="p-6 grid grid-cols-3 gap-6">
					<div class="col-span-2 space-y-6">
						{#if selectedTask.description}
							<div>
								<h3 class="text-sm font-semibold text-gray-700 mb-2">Описание</h3>
								<p class="text-gray-600 whitespace-pre-wrap">{@html parseContent(selectedTask.description)}</p>
							</div>
						{/if}
						{#if selectedTask.tags && selectedTask.tags.length > 0}
							<div>
								<h3 class="text-sm font-semibold text-gray-700 mb-2">Теги</h3>
								<div class="flex flex-wrap gap-2">
									{#each selectedTask.tags as tag}
										<span class="px-2 py-1 bg-gray-100 text-gray-600 rounded text-sm">{tag}</span>
									{/each}
								</div>
							</div>
						{/if}
						<div>
							<h3 class="text-sm font-semibold text-gray-700 mb-3">Комментарии</h3>
							<div class="space-y-4">
								{#if selectedTask.comments && selectedTask.comments.length > 0}
									{#each selectedTask.comments as comment}
										<div class="flex gap-3">
											<div class="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center text-sm font-medium flex-shrink-0">{comment.user?.name?.charAt(0) || '?'}</div>
											<div class="flex-1">
												<div class="flex items-center gap-2 mb-1">
													<span class="font-medium text-gray-900">{comment.user?.name || 'Пользователь'}</span>
													<span class="text-xs text-gray-400">{formatDateTime(comment.created_at)}</span>
												</div>
												<p class="text-gray-600">{@html parseContent(comment.content)}</p>
											</div>
										</div>
									{/each}
								{:else}
									<p class="text-gray-400 text-sm">Нет комментариев</p>
								{/if}
								<div class="flex gap-3 pt-4 border-t">
									<div class="w-8 h-8 rounded-full bg-ekf-red text-white flex items-center justify-center text-sm font-medium flex-shrink-0">{$user?.name?.charAt(0) || '?'}</div>
									<div class="flex-1">
										<textarea bind:value={newComment} rows="2" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" placeholder="Добавить комментарий... (@упоминание)"></textarea>
										<div class="flex justify-end mt-2">
											<button onclick={addComment} disabled={!newComment.trim()} class="px-3 py-1.5 bg-ekf-red text-white rounded text-sm hover:bg-red-700 disabled:opacity-50">Отправить</button>
										</div>
									</div>
								</div>
							</div>
						</div>
					</div>
					<div class="space-y-4">
						<div>
							<label class="block text-sm font-medium text-gray-500 mb-1">Статус</label>
							<select bind:value={selectedTask.status} onchange={() => updateTask(selectedTask!)} class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
								{#each columns as col}
									<option value={col.id}>{col.label}</option>
								{/each}
							</select>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-500 mb-1">Исполнитель</label>
							<div class="text-gray-900">{selectedTask.assignee_id ? getAssigneeName(selectedTask.assignee_id) : 'Не назначен'}</div>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-500 mb-1">Проект</label>
							<div class="text-gray-900">{selectedTask.project_id ? getProjectName(selectedTask.project_id) : '-'}</div>
						</div>
						{#if selectedTask.due_date}
							<div>
								<label class="block text-sm font-medium text-gray-500 mb-1">Срок</label>
								<div class="text-gray-900">{formatDate(selectedTask.due_date)}</div>
							</div>
						{/if}
						<div>
							<label class="block text-sm font-medium text-gray-500 mb-1">Создана</label>
							<div class="text-gray-600 text-sm">{formatDateTime(selectedTask.created_at)}</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}
