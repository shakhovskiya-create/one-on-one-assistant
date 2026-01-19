<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { projects as projectsApi, tasks as tasksApi } from '$lib/api/client';
	import type { Project, Task } from '$lib/api/client';

	let project: Project | null = $state(null);
	let tasks: Task[] = $state([]);
	let loading = $state(true);
	let viewMode = $state<'gantt' | 'list' | 'kanban'>('gantt');
	let showTaskModal = $state(false);
	let editingTask: Partial<Task> | null = $state(null);

	// Gantt chart state
	let ganttStartDate = $state(new Date());
	let ganttDays = $state(90);
	let dayWidth = $state(24);
	let rowHeight = 40;

	const statusColors: Record<string, string> = {
		todo: 'bg-gray-400',
		in_progress: 'bg-blue-500',
		review: 'bg-yellow-500',
		done: 'bg-green-500'
	};

	const priorityColors: Record<number, string> = {
		1: 'border-red-500',
		2: 'border-orange-500',
		3: 'border-yellow-500',
		4: 'border-blue-500',
		5: 'border-gray-300'
	};

	onMount(async () => {
		const id = $page.params.id;
		try {
			const [projectData, tasksData] = await Promise.all([
				projectsApi.get(id),
				tasksApi.list({ project_id: id })
			]);
			project = projectData;
			tasks = tasksData || [];

			if (tasks.length > 0) {
				const dates = tasks
					.map(t => t.start_date || t.due_date)
					.filter(Boolean)
					.map(d => new Date(d!));
				if (dates.length > 0) {
					const minDate = new Date(Math.min(...dates.map(d => d.getTime())));
					minDate.setDate(minDate.getDate() - 7);
					ganttStartDate = minDate;
				}
			}
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	function getTaskPosition(task: Task): { left: number; width: number } {
		const start = task.start_date ? new Date(task.start_date) : new Date();
		const end = task.due_date ? new Date(task.due_date) : new Date(start.getTime() + 7 * 24 * 60 * 60 * 1000);
		const startDiff = Math.floor((start.getTime() - ganttStartDate.getTime()) / (24 * 60 * 60 * 1000));
		const duration = Math.max(1, Math.ceil((end.getTime() - start.getTime()) / (24 * 60 * 60 * 1000)));
		return { left: Math.max(0, startDiff * dayWidth), width: duration * dayWidth };
	}

	function getMonthHeaders(): { month: string; days: number; offset: number }[] {
		const headers: { month: string; days: number; offset: number }[] = [];
		const months = ['Янв', 'Фев', 'Мар', 'Апр', 'Май', 'Июн', 'Июл', 'Авг', 'Сен', 'Окт', 'Ноя', 'Дек'];
		let currentDate = new Date(ganttStartDate);
		let offset = 0;
		for (let i = 0; i < ganttDays;) {
			const month = currentDate.getMonth();
			const year = currentDate.getFullYear();
			const daysInMonth = new Date(year, month + 1, 0).getDate();
			const remainingDays = daysInMonth - currentDate.getDate() + 1;
			const days = Math.min(remainingDays, ganttDays - i);
			headers.push({ month: months[month] + ' ' + year, days, offset });
			offset += days * dayWidth;
			i += days;
			currentDate = new Date(year, month + 1, 1);
		}
		return headers;
	}

	function getDayHeaders(): { day: number; isWeekend: boolean; isToday: boolean }[] {
		const days: { day: number; isWeekend: boolean; isToday: boolean }[] = [];
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		for (let i = 0; i < ganttDays; i++) {
			const date = new Date(ganttStartDate);
			date.setDate(date.getDate() + i);
			const dayOfWeek = date.getDay();
			days.push({ day: date.getDate(), isWeekend: dayOfWeek === 0 || dayOfWeek === 6, isToday: date.getTime() === today.getTime() });
		}
		return days;
	}

	function openTaskModal(task?: Task) {
		if (task) {
			editingTask = { ...task };
		} else {
			editingTask = {
				title: '',
				description: '',
				status: 'todo',
				priority: 3,
				project_id: project?.id,
				start_date: new Date().toISOString().split('T')[0],
				due_date: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
			};
		}
		showTaskModal = true;
	}

	async function saveTask() {
		if (!editingTask?.title) return;
		try {
			if (editingTask.id) {
				await tasksApi.update(editingTask.id, editingTask);
				tasks = tasks.map(t => t.id === editingTask!.id ? { ...t, ...editingTask } as Task : t);
			} else {
				const created = await tasksApi.create(editingTask as Omit<Task, 'id'>);
				tasks = [...tasks, created];
			}
			showTaskModal = false;
			editingTask = null;
		} catch (e) {
			console.error(e);
		}
	}

	async function deleteTask(id: string) {
		if (!confirm('Удалить задачу?')) return;
		try {
			await tasksApi.delete(id);
			tasks = tasks.filter(t => t.id !== id);
		} catch (e) {
			console.error(e);
		}
	}

	async function updateTaskStatus(task: Task, status: string) {
		try {
			await tasksApi.update(task.id, { status });
			tasks = tasks.map(t => t.id === task.id ? { ...t, status } : t);
		} catch (e) {
			console.error(e);
		}
	}

	function formatDate(date: string | undefined): string {
		if (!date) return '';
		return new Date(date).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
	}

	let sortedTasks = $derived([...tasks].sort((a, b) => {
		if (a.is_epic !== b.is_epic) return a.is_epic ? -1 : 1;
		const aDate = a.start_date || a.due_date || '';
		const bDate = b.start_date || b.due_date || '';
		return aDate.localeCompare(bDate);
	}));
</script>

<svelte:head>
	<title>{project?.name || 'Проект'} - EKF Hub</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center h-64">
		<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
	</div>
{:else if !project}
	<div class="text-center py-12">
		<p class="text-gray-500">Проект не найден</p>
		<a href="/projects" class="text-ekf-red hover:underline mt-2 inline-block">Вернуться</a>
	</div>
{:else}
	<div class="space-y-6">
		<div class="flex items-center justify-between">
			<div>
				<div class="flex items-center gap-2 text-sm text-gray-500 mb-1">
					<a href="/projects" class="hover:text-ekf-red">Проекты</a>
					<span>/</span>
				</div>
				<h1 class="text-2xl font-bold text-gray-900">{project.name}</h1>
				{#if project.description}
					<p class="text-gray-600 mt-1">{project.description}</p>
				{/if}
			</div>
			<div class="flex items-center gap-3">
				<div class="flex rounded-lg border border-gray-200 overflow-hidden">
					{#each [{ key: 'gantt', label: 'Gantt' }, { key: 'list', label: 'Список' }, { key: 'kanban', label: 'Kanban' }] as mode}
						<button onclick={() => viewMode = mode.key as typeof viewMode}
							class="px-3 py-2 text-sm {viewMode === mode.key ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}">
							{mode.label}
						</button>
					{/each}
				</div>
				<button onclick={() => openTaskModal()} class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 flex items-center gap-2">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					Задача
				</button>
			</div>
		</div>

		{#if viewMode === 'gantt'}
			<div class="bg-white rounded-xl shadow-sm overflow-hidden">
				<div class="flex">
					<div class="w-72 flex-shrink-0 border-r border-gray-200">
						<div class="h-16 border-b border-gray-200 px-4 flex items-center">
							<span class="font-semibold text-gray-900">Задачи</span>
						</div>
						<div class="overflow-y-auto">
							{#each sortedTasks as task}
								<button onclick={() => openTaskModal(task)}
									class="w-full px-4 py-2 text-left hover:bg-gray-50 border-b border-gray-100 flex items-center gap-2"
									style="height: {rowHeight}px">
									{#if task.is_epic}
										<svg class="w-4 h-4 text-purple-600" fill="currentColor" viewBox="0 0 20 20">
											<path d="M10.394 2.08a1 1 0 00-.788 0l-7 3a1 1 0 000 1.84L5.25 8.051a.999.999 0 01.356-.257l4-1.714a1 1 0 11.788 1.838L7.667 9.088l1.94.831a1 1 0 00.787 0l7-3a1 1 0 000-1.838l-7-3z" />
										</svg>
									{/if}
									<span class="truncate flex-1 text-sm {task.status === 'done' ? 'line-through text-gray-400' : 'text-gray-900'}">{task.title}</span>
									<span class="w-2 h-2 rounded-full {statusColors[task.status] || 'bg-gray-400'}"></span>
								</button>
							{/each}
						</div>
					</div>
					<div class="flex-1 overflow-x-auto">
						<div class="h-8 border-b border-gray-200 flex relative bg-gray-50">
							{#each getMonthHeaders() as header}
								<div class="text-xs font-medium text-gray-600 px-2 border-r border-gray-200 flex items-center" style="width: {header.days * dayWidth}px">
									{header.month}
								</div>
							{/each}
						</div>
						<div class="h-8 border-b border-gray-200 flex">
							{#each getDayHeaders() as day}
								<div class="text-xs text-center border-r border-gray-100 flex items-center justify-center
									{day.isWeekend ? 'bg-gray-50 text-gray-400' : 'text-gray-600'}
									{day.isToday ? 'bg-ekf-red/10 font-bold text-ekf-red' : ''}" style="width: {dayWidth}px">{day.day}</div>
							{/each}
						</div>
						<div class="relative">
							{#each sortedTasks as task}
								{@const pos = getTaskPosition(task)}
								<div class="relative border-b border-gray-100" style="height: {rowHeight}px">
									{#each getDayHeaders() as day, dayIndex}
										<div class="absolute top-0 bottom-0 border-r border-gray-50 {day.isWeekend ? 'bg-gray-50' : ''} {day.isToday ? 'bg-ekf-red/5' : ''}"
											style="left: {dayIndex * dayWidth}px; width: {dayWidth}px"></div>
									{/each}
									<div class="absolute top-1.5 h-7 rounded cursor-pointer transition-all hover:opacity-80
										{task.is_epic ? 'bg-purple-500' : statusColors[task.status] || 'bg-gray-400'}
										{priorityColors[task.priority || 3]} border-l-4"
										style="left: {pos.left}px; width: {Math.max(pos.width, 20)}px"
										onclick={() => openTaskModal(task)} title="{task.title}">
										<div class="px-2 truncate text-xs text-white leading-7">{task.title}</div>
										{#if task.progress && task.progress > 0}
											<div class="absolute bottom-0 left-0 h-1 bg-white/50 rounded-b" style="width: {task.progress}%"></div>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</div>
		{/if}

		{#if viewMode === 'list'}
			<div class="bg-white rounded-xl shadow-sm overflow-hidden">
				<table class="w-full">
					<thead class="bg-gray-50 border-b">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Задача</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Статус</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Срок</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Прогресс</th>
							<th class="px-6 py-3"></th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100">
						{#each sortedTasks as task}
							<tr class="hover:bg-gray-50">
								<td class="px-6 py-4">
									<div class="flex items-center gap-2">
										{#if task.is_epic}<span class="text-purple-600 text-xs font-medium bg-purple-100 px-2 py-0.5 rounded">EPIC</span>{/if}
										<span class="font-medium text-gray-900">{task.title}</span>
									</div>
								</td>
								<td class="px-6 py-4">
									<select value={task.status} onchange={(e) => updateTaskStatus(task, (e.target as HTMLSelectElement).value)} class="text-sm border rounded px-2 py-1">
										<option value="todo">К выполнению</option>
										<option value="in_progress">В работе</option>
										<option value="review">На проверке</option>
										<option value="done">Выполнено</option>
									</select>
								</td>
								<td class="px-6 py-4 text-sm text-gray-600">{formatDate(task.due_date)}</td>
								<td class="px-6 py-4">
									<div class="flex items-center gap-2">
										<div class="w-24 h-2 bg-gray-200 rounded-full overflow-hidden">
											<div class="h-full bg-green-500" style="width: {task.progress || 0}%"></div>
										</div>
										<span class="text-xs text-gray-500">{task.progress || 0}%</span>
									</div>
								</td>
								<td class="px-6 py-4 text-right">
									<button onclick={() => openTaskModal(task)} class="text-gray-400 hover:text-gray-600 mr-2">Edit</button>
									<button onclick={() => deleteTask(task.id)} class="text-gray-400 hover:text-red-600">Del</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}

		{#if viewMode === 'kanban'}
			<div class="grid grid-cols-4 gap-4">
				{#each [{ key: 'todo', label: 'К выполнению', color: 'bg-gray-100' }, { key: 'in_progress', label: 'В работе', color: 'bg-blue-100' }, { key: 'review', label: 'На проверке', color: 'bg-yellow-100' }, { key: 'done', label: 'Выполнено', color: 'bg-green-100' }] as column}
					<div class="rounded-xl {column.color} p-4">
						<h3 class="font-semibold text-gray-900 mb-4 flex items-center justify-between">{column.label}
							<span class="text-sm font-normal text-gray-500">{tasks.filter(t => t.status === column.key).length}</span></h3>
						<div class="space-y-3">
							{#each tasks.filter(t => t.status === column.key) as task}
								<div onclick={() => openTaskModal(task)} class="bg-white rounded-lg shadow-sm p-4 cursor-pointer hover:shadow-md transition-shadow">
									<div class="flex items-start justify-between mb-2">
										<span class="font-medium text-gray-900 text-sm">{task.title}</span>
										{#if task.is_epic}<span class="text-xs bg-purple-100 text-purple-600 px-1.5 py-0.5 rounded">EPIC</span>{/if}
									</div>
									{#if task.due_date}<div class="text-xs text-gray-500">{formatDate(task.due_date)}</div>{/if}
									{#if task.progress && task.progress > 0}
										<div class="mt-2 h-1 bg-gray-200 rounded-full overflow-hidden"><div class="h-full bg-green-500" style="width: {task.progress}%"></div></div>
									{/if}
								</div>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
{/if}

{#if showTaskModal && editingTask}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
		<div class="bg-white rounded-xl shadow-xl p-6 w-full max-w-lg">
			<h2 class="text-xl font-bold text-gray-900 mb-4">{editingTask.id ? 'Редактировать' : 'Новая задача'}</h2>
			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Название *</label>
					<input type="text" bind:value={editingTask.title} class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent" />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Описание</label>
					<textarea bind:value={editingTask.description} rows="3" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"></textarea>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Дата начала</label>
						<input type="date" bind:value={editingTask.start_date} class="w-full px-3 py-2 border border-gray-300 rounded-lg" />
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Дедлайн</label>
						<input type="date" bind:value={editingTask.due_date} class="w-full px-3 py-2 border border-gray-300 rounded-lg" />
					</div>
				</div>
				<div class="grid grid-cols-3 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Статус</label>
						<select bind:value={editingTask.status} class="w-full px-3 py-2 border border-gray-300 rounded-lg">
							<option value="todo">К выполнению</option>
							<option value="in_progress">В работе</option>
							<option value="review">На проверке</option>
							<option value="done">Выполнено</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Приоритет</label>
						<select bind:value={editingTask.priority} class="w-full px-3 py-2 border border-gray-300 rounded-lg">
							<option value={1}>P1 - Критический</option>
							<option value={2}>P2 - Высокий</option>
							<option value={3}>P3 - Средний</option>
							<option value={4}>P4 - Низкий</option>
							<option value={5}>P5 - Минимальный</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Прогресс %</label>
						<input type="number" min="0" max="100" bind:value={editingTask.progress} class="w-full px-3 py-2 border border-gray-300 rounded-lg" />
					</div>
				</div>
				<label class="flex items-center gap-2">
					<input type="checkbox" bind:checked={editingTask.is_epic} class="rounded text-ekf-red" />
					<span class="text-sm font-medium text-gray-700">Это Epic</span>
				</label>
			</div>
			<div class="flex gap-3 mt-6">
				<button onclick={() => { showTaskModal = false; editingTask = null; }} class="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Отмена</button>
				<button onclick={saveTask} class="flex-1 px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700">{editingTask.id ? 'Сохранить' : 'Создать'}</button>
			</div>
		</div>
	</div>
{/if}
