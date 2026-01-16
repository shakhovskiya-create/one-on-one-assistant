<script lang="ts">
	import { onMount } from 'svelte';
	import { employees as employeesApi } from '$lib/api/client';
	import type { Employee } from '$lib/api/client';
	import { notifications } from '$lib/stores/app';

	let employees: Employee[] = $state([]);
	let loading = $state(true);
	let searchQuery = $state('');

	let filteredEmployees = $derived(
		employees.filter(e =>
			e.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
			e.position?.toLowerCase().includes(searchQuery.toLowerCase()) ||
			e.department?.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	onMount(async () => {
		try {
			employees = await employeesApi.list();
		} catch (e) {
			notifications.add({
				type: 'error',
				message: 'Ошибка загрузки сотрудников'
			});
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Сотрудники - EKF Team Hub</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-gray-900">Сотрудники</h1>
		<a
			href="/employees/new"
			class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors"
		>
			+ Добавить
		</a>
	</div>

	<!-- Search -->
	<div class="relative">
		<input
			type="text"
			bind:value={searchQuery}
			placeholder="Поиск сотрудников..."
			class="w-full px-4 py-3 pl-10 rounded-lg border border-gray-200 focus:outline-none focus:ring-2 focus:ring-ekf-red/20 focus:border-ekf-red"
		/>
		<svg class="w-5 h-5 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
		</svg>
	</div>

	<!-- List -->
	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="text-gray-500">Загрузка...</div>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each filteredEmployees as employee (employee.id)}
				<a
					href="/employees/{employee.id}"
					class="bg-white rounded-xl shadow-sm p-4 hover:shadow-md transition-shadow"
				>
					<div class="flex items-center gap-4">
						<div class="w-12 h-12 rounded-full bg-gray-200 flex items-center justify-center text-xl font-medium text-gray-600">
							{#if employee.photo_base64}
								<img
									src="data:image/jpeg;base64,{employee.photo_base64}"
									alt={employee.name}
									class="w-full h-full rounded-full object-cover"
								/>
							{:else}
								{employee.name.charAt(0)}
							{/if}
						</div>
						<div class="flex-1 min-w-0">
							<p class="font-semibold text-gray-900 truncate">{employee.name}</p>
							<p class="text-sm text-gray-500 truncate">{employee.position}</p>
							{#if employee.department}
								<p class="text-xs text-gray-400 truncate">{employee.department}</p>
							{/if}
						</div>
					</div>
				</a>
			{/each}
		</div>

		{#if filteredEmployees.length === 0}
			<div class="text-center py-12">
				<div class="text-gray-400 text-lg">Сотрудники не найдены</div>
			</div>
		{/if}
	{/if}
</div>
