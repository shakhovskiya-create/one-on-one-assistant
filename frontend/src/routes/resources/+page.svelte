<script lang="ts">
	import { onMount } from 'svelte';
	import { resources, projects, type ResourceCapacity, type ResourceStats, type ResourceAllocation, type Project } from '$lib/api/client';

	let capacities = $state<ResourceCapacity[]>([]);
	let stats = $state<ResourceStats | null>(null);
	let allocations = $state<ResourceAllocation[]>([]);
	let projectList = $state<Project[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	// Filters
	let selectedProject = $state<string>('');

	// Modal state
	let showAllocationModal = $state(false);
	let allocationForm = $state({
		employee_id: '',
		project_id: '',
		task_id: '',
		role: '',
		allocated_hours_per_week: 8,
		period_start: new Date().toISOString().split('T')[0],
		period_end: '',
		notes: ''
	});

	async function loadData() {
		loading = true;
		error = null;
		try {
			const params = selectedProject ? { project_id: selectedProject } : undefined;
			const [capacityRes, statsRes, allocRes, projRes] = await Promise.all([
				resources.getCapacity(params),
				resources.getStats(params),
				resources.listAllocations(params),
				projects.list()
			]);
			capacities = capacityRes || [];
			stats = statsRes;
			allocations = allocRes || [];
			projectList = projRes || [];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load data';
		} finally {
			loading = false;
		}
	}

	async function createAllocation() {
		if (!allocationForm.employee_id || !allocationForm.period_start) {
			alert('Employee and start date are required');
			return;
		}
		try {
			await resources.createAllocation({
				employee_id: allocationForm.employee_id,
				project_id: allocationForm.project_id || undefined,
				task_id: allocationForm.task_id || undefined,
				role: allocationForm.role || undefined,
				allocated_hours_per_week: allocationForm.allocated_hours_per_week,
				period_start: allocationForm.period_start,
				period_end: allocationForm.period_end || undefined,
				notes: allocationForm.notes || undefined
			});
			showAllocationModal = false;
			allocationForm = {
				employee_id: '',
				project_id: '',
				task_id: '',
				role: '',
				allocated_hours_per_week: 8,
				period_start: new Date().toISOString().split('T')[0],
				period_end: '',
				notes: ''
			};
			await loadData();
		} catch (e) {
			alert(e instanceof Error ? e.message : 'Failed to create allocation');
		}
	}

	async function deleteAllocation(id: string) {
		if (!confirm('Delete this allocation?')) return;
		try {
			await resources.deleteAllocation(id);
			await loadData();
		} catch (e) {
			alert(e instanceof Error ? e.message : 'Failed to delete allocation');
		}
	}

	function getUtilizationColor(pct: number): string {
		if (pct > 100) return 'bg-red-500';
		if (pct > 80) return 'bg-yellow-500';
		if (pct > 50) return 'bg-green-500';
		return 'bg-gray-300';
	}

	function getUtilizationTextColor(pct: number): string {
		if (pct > 100) return 'text-red-600';
		if (pct > 80) return 'text-yellow-600';
		return 'text-green-600';
	}

	onMount(loadData);

	$effect(() => {
		if (selectedProject !== undefined) {
			loadData();
		}
	});
</script>

<svelte:head>
	<title>Resource Planning | EKF Hub</title>
</svelte:head>

<div class="p-6">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Планирование ресурсов</h1>
			<p class="text-gray-600 mt-1">Управление загрузкой и распределением сотрудников</p>
		</div>
		<button
			onclick={() => showAllocationModal = true}
			class="px-4 py-2 bg-[#E53935] text-white rounded-lg hover:bg-red-600 transition-colors"
		>
			+ Добавить аллокацию
		</button>
	</div>

	<!-- Stats Cards -->
	{#if stats}
		<div class="grid grid-cols-1 md:grid-cols-5 gap-4 mb-6">
			<div class="bg-white rounded-lg shadow p-4">
				<div class="text-sm text-gray-500">Сотрудников</div>
				<div class="text-2xl font-bold">{stats.total_employees}</div>
			</div>
			<div class="bg-white rounded-lg shadow p-4">
				<div class="text-sm text-gray-500">Аллокаций</div>
				<div class="text-2xl font-bold">{stats.total_allocations}</div>
			</div>
			<div class="bg-white rounded-lg shadow p-4">
				<div class="text-sm text-gray-500">Перегружены</div>
				<div class="text-2xl font-bold text-red-600">{stats.overloaded_count}</div>
			</div>
			<div class="bg-white rounded-lg shadow p-4">
				<div class="text-sm text-gray-500">Недозагружены</div>
				<div class="text-2xl font-bold text-yellow-600">{stats.underutilized_count}</div>
			</div>
			<div class="bg-white rounded-lg shadow p-4">
				<div class="text-sm text-gray-500">Ср. загрузка</div>
				<div class="text-2xl font-bold">{stats.avg_utilization.toFixed(0)}%</div>
			</div>
		</div>
	{/if}

	<!-- Filters -->
	<div class="bg-white rounded-lg shadow p-4 mb-6">
		<div class="flex items-center gap-4">
			<label class="flex items-center gap-2">
				<span class="text-sm text-gray-600">Проект:</span>
				<select
					bind:value={selectedProject}
					class="border rounded px-3 py-1.5 text-sm"
				>
					<option value="">Все проекты</option>
					{#each projectList as project}
						<option value={project.id}>{project.name}</option>
					{/each}
				</select>
			</label>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-[#E53935]"></div>
		</div>
	{:else if error}
		<div class="bg-red-50 text-red-700 p-4 rounded-lg">{error}</div>
	{:else}
		<!-- Capacity Table -->
		<div class="bg-white rounded-lg shadow overflow-hidden mb-6">
			<div class="px-6 py-4 border-b">
				<h2 class="text-lg font-semibold">Загрузка сотрудников</h2>
			</div>
			<div class="overflow-x-auto">
				<table class="w-full">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-4 py-3 text-left text-sm font-medium text-gray-600">Сотрудник</th>
							<th class="px-4 py-3 text-left text-sm font-medium text-gray-600">Должность</th>
							<th class="px-4 py-3 text-center text-sm font-medium text-gray-600">Норма ч/нед</th>
							<th class="px-4 py-3 text-center text-sm font-medium text-gray-600">Доступность</th>
							<th class="px-4 py-3 text-center text-sm font-medium text-gray-600">Доступно</th>
							<th class="px-4 py-3 text-center text-sm font-medium text-gray-600">Выделено</th>
							<th class="px-4 py-3 text-center text-sm font-medium text-gray-600">Свободно</th>
							<th class="px-4 py-3 text-left text-sm font-medium text-gray-600">Загрузка</th>
						</tr>
					</thead>
					<tbody class="divide-y">
						{#each capacities as cap}
							<tr class="hover:bg-gray-50" class:bg-red-50={cap.overloaded}>
								<td class="px-4 py-3">
									<div class="font-medium text-gray-900">{cap.employee_name}</div>
								</td>
								<td class="px-4 py-3 text-gray-600">{cap.position || '—'}</td>
								<td class="px-4 py-3 text-center">{cap.weekly_hours}</td>
								<td class="px-4 py-3 text-center">{cap.availability_pct}%</td>
								<td class="px-4 py-3 text-center">{cap.available_hours.toFixed(0)}</td>
								<td class="px-4 py-3 text-center font-medium">{cap.allocated_hours}</td>
								<td class="px-4 py-3 text-center" class:text-red-600={cap.free_hours < 0}>
									{cap.free_hours.toFixed(0)}
								</td>
								<td class="px-4 py-3">
									<div class="flex items-center gap-2">
										<div class="w-24 h-2 bg-gray-200 rounded-full overflow-hidden">
											<div
												class="{getUtilizationColor(cap.utilization_percent)} h-full transition-all"
												style="width: {Math.min(cap.utilization_percent, 100)}%"
											></div>
										</div>
										<span class="{getUtilizationTextColor(cap.utilization_percent)} text-sm font-medium">
											{cap.utilization_percent.toFixed(0)}%
										</span>
										{#if cap.overloaded}
											<span class="text-red-600 text-xs">Перегружен!</span>
										{/if}
									</div>
								</td>
							</tr>
						{:else}
							<tr>
								<td colspan="8" class="px-4 py-8 text-center text-gray-500">
									Нет данных о сотрудниках
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>

		<!-- Allocations List -->
		<div class="bg-white rounded-lg shadow overflow-hidden">
			<div class="px-6 py-4 border-b">
				<h2 class="text-lg font-semibold">Текущие аллокации</h2>
			</div>
			<div class="divide-y">
				{#each allocations as alloc}
					<div class="px-6 py-4 flex items-center justify-between hover:bg-gray-50">
						<div class="flex-1">
							<div class="font-medium">{alloc.employee?.name || alloc.employee_id}</div>
							<div class="text-sm text-gray-600">
								{alloc.project?.name || 'Без проекта'}
								{#if alloc.role}
									<span class="text-gray-400">• {alloc.role}</span>
								{/if}
							</div>
							<div class="text-xs text-gray-500 mt-1">
								{alloc.period_start}
								{#if alloc.period_end}
									— {alloc.period_end}
								{:else}
									— бессрочно
								{/if}
							</div>
						</div>
						<div class="flex items-center gap-4">
							<div class="text-right">
								<div class="text-lg font-bold">{alloc.allocated_hours_per_week}</div>
								<div class="text-xs text-gray-500">ч/нед</div>
							</div>
							<button
								onclick={() => deleteAllocation(alloc.id)}
								class="text-red-600 hover:text-red-800 p-1"
								title="Удалить"
							>
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
								</svg>
							</button>
						</div>
					</div>
				{:else}
					<div class="px-6 py-8 text-center text-gray-500">
						Нет аллокаций
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

<!-- Allocation Modal -->
{#if showAllocationModal}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
		<div class="bg-white rounded-lg shadow-xl w-full max-w-md p-6">
			<h3 class="text-lg font-semibold mb-4">Новая аллокация</h3>

			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Сотрудник *</label>
					<select
						bind:value={allocationForm.employee_id}
						class="w-full border rounded-lg px-3 py-2"
					>
						<option value="">Выберите сотрудника</option>
						{#each capacities as cap}
							<option value={cap.employee_id}>{cap.employee_name}</option>
						{/each}
					</select>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Проект</label>
					<select
						bind:value={allocationForm.project_id}
						class="w-full border rounded-lg px-3 py-2"
					>
						<option value="">Без проекта</option>
						{#each projectList as project}
							<option value={project.id}>{project.name}</option>
						{/each}
					</select>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Роль</label>
					<input
						type="text"
						bind:value={allocationForm.role}
						placeholder="Например: Backend Developer"
						class="w-full border rounded-lg px-3 py-2"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Часов в неделю *</label>
					<input
						type="number"
						bind:value={allocationForm.allocated_hours_per_week}
						min="1"
						max="80"
						class="w-full border rounded-lg px-3 py-2"
					/>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Начало *</label>
						<input
							type="date"
							bind:value={allocationForm.period_start}
							class="w-full border rounded-lg px-3 py-2"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Окончание</label>
						<input
							type="date"
							bind:value={allocationForm.period_end}
							class="w-full border rounded-lg px-3 py-2"
						/>
					</div>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Примечание</label>
					<textarea
						bind:value={allocationForm.notes}
						rows="2"
						class="w-full border rounded-lg px-3 py-2"
					></textarea>
				</div>
			</div>

			<div class="flex justify-end gap-3 mt-6">
				<button
					onclick={() => showAllocationModal = false}
					class="px-4 py-2 text-gray-600 hover:text-gray-800"
				>
					Отмена
				</button>
				<button
					onclick={createAllocation}
					class="px-4 py-2 bg-[#E53935] text-white rounded-lg hover:bg-red-600"
				>
					Создать
				</button>
			</div>
		</div>
	</div>
{/if}
