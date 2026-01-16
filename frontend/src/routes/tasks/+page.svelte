<script lang="ts">
	import { onMount } from 'svelte';
	import { tasks as tasksApi } from '$lib/api/client';
	import type { KanbanBoard, Task } from '$lib/api/client';

	let kanban: KanbanBoard | null = $state(null);
	let loading = $state(true);
	let showCreateModal = $state(false);
	let newTask = $state({ title: '', description: '', priority: 'medium' });
	let draggedTask: Task | null = $state(null);

	const columns = [
		{ key: 'todo', label: 'К выполнению', color: 'bg-gray-100' },
		{ key: 'in_progress', label: 'В работе', color: 'bg-blue-50' },
		{ key: 'done', label: 'Выполнено', color: 'bg-green-50' }
	];

	onMount(async () => {
		await loadKanban();
	});

	async function loadKanban() {
		try {
			kanban = await tasksApi.getKanban();
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function createTask() {
		if (!newTask.title) return;
		try {
			await tasksApi.create({
				title: newTask.title,
				description: newTask.description,
				priority: newTask.priority,
				status: 'todo'
			});
			showCreateModal = false;
			newTask = { title: '', description: '', priority: 'medium' };
			await loadKanban();
		} catch (e) {
			console.error(e);
		}
	}

	function handleDragStart(task: Task) {
		draggedTask = task;
	}

	async function handleDrop(status: string) {
		if (!draggedTask || draggedTask.status === status) {
			draggedTask = null;
			return;
		}

		try {
			await tasksApi.moveKanban(draggedTask.id, status);
			await loadKanban();
		} catch (e) {
			console.error(e);
		} finally {
			draggedTask = null;
		}
	}

	function getTasksByStatus(status: string): Task[] {
		if (!kanban) return [];
		switch (status) {
			case 'todo': return kanban.todo || [];
			case 'in_progress': return kanban.in_progress || [];
			case 'done': return kanban.done || [];
			default: return [];
		}
	}

	function getPriorityColor(priority: string): string {
		switch (priority) {
			case 'high': return 'border-l-red-500';
			case 'medium': return 'border-l-yellow-500';
			case 'low': return 'border-l-green-500';
			default: return 'border-l-gray-300';
		}
	}
</script>

<svelte:head>
	<title>Задачи - EKF Team Hub</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-gray-900">Канбан-доска</h1>
		<button
			onclick={() => showCreateModal = true}
			class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors"
		>
			Новая задача
		</button>
	</div>

	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="text-gray-500">Загрузка...</div>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
			{#each columns as column}
				<div
					class="rounded-xl p-4 min-h-96 {column.color}"
					ondragover={(e) => e.preventDefault()}
					ondrop={() => handleDrop(column.key)}
				>
					<div class="flex items-center justify-between mb-4">
						<h3 class="font-semibold text-gray-900">{column.label}</h3>
						<span class="text-sm text-gray-500 bg-white px-2 py-1 rounded-full">
							{getTasksByStatus(column.key).length}
						</span>
					</div>
					<div class="space-y-3">
						{#each getTasksByStatus(column.key) as task}
							<div
								class="bg-white rounded-lg p-4 shadow-sm border-l-4 cursor-move hover:shadow-md transition-shadow {getPriorityColor(task.priority || 'medium')}"
								draggable="true"
								ondragstart={() => handleDragStart(task)}
							>
								<h4 class="font-medium text-gray-900 mb-1">{task.title}</h4>
								{#if task.description}
									<p class="text-sm text-gray-600 line-clamp-2">{task.description}</p>
								{/if}
								<div class="flex items-center justify-between mt-3">
									{#if task.due_date}
										<span class="text-xs text-gray-500">{task.due_date}</span>
									{:else}
										<span></span>
									{/if}
									{#if task.assignee_name}
										<span class="text-xs bg-gray-100 text-gray-600 px-2 py-1 rounded">
											{task.assignee_name}
										</span>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Create Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
		<div class="bg-white rounded-xl shadow-xl p-6 w-full max-w-md">
			<h2 class="text-xl font-bold text-gray-900 mb-4">Новая задача</h2>
			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Название *</label>
					<input
						type="text"
						bind:value={newTask.title}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						placeholder="Что нужно сделать?"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Описание</label>
					<textarea
						bind:value={newTask.description}
						rows="3"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						placeholder="Подробности задачи"
					></textarea>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Приоритет</label>
					<select
						bind:value={newTask.priority}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
					>
						<option value="low">Низкий</option>
						<option value="medium">Средний</option>
						<option value="high">Высокий</option>
					</select>
				</div>
			</div>
			<div class="flex gap-3 mt-6">
				<button
					onclick={() => showCreateModal = false}
					class="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
				>
					Отмена
				</button>
				<button
					onclick={createTask}
					class="flex-1 px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700"
				>
					Создать
				</button>
			</div>
		</div>
	</div>
{/if}
