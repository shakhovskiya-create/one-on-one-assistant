<script lang="ts">
	import { onMount } from 'svelte';
	import { employees as employeesApi } from '$lib/api/client';
	import type { Employee } from '$lib/api/client';
	import { notifications } from '$lib/stores/app';
	import { user, subordinates } from '$lib/stores/auth';

	let allEmployees: Employee[] = $state([]);
	let loading = $state(true);
	let searchQuery = $state('');
	let viewMode = $state<'my' | 'all'>('my');
	let displayMode = $state<'grid' | 'list' | 'hierarchy' | 'department'>('grid');
	let collapsedDepts: Set<string> = $state(new Set());
	let collapsedNodes: Set<string> = $state(new Set());

	// Helper: get all descendants recursively
	function getAllDescendants(managerId: string, allEmps: Employee[]): Employee[] {
		const direct = allEmps.filter(e => e.manager_id === managerId);
		const all: Employee[] = [...direct];
		for (const emp of direct) {
			all.push(...getAllDescendants(emp.id, allEmps));
		}
		return all;
	}

	// Filter employees based on view mode and department presence
	let visibleEmployees = $derived(() => {
		let employees: Employee[];

		if (viewMode === 'my') {
			// Show full tree: user's subordinates + their subordinates recursively
			if ($user?.id) {
				employees = getAllDescendants($user.id, allEmployees);
			} else {
				employees = $subordinates;
			}
		} else {
			// Show all employees with department
			employees = allEmployees.filter(e => e.department && e.department.trim() !== '');
		}

		// Apply search filter
		if (searchQuery.trim()) {
			const query = searchQuery.toLowerCase();
			employees = employees.filter(e =>
				e.name.toLowerCase().includes(query) ||
				e.position?.toLowerCase().includes(query) ||
				e.department?.toLowerCase().includes(query)
			);
		}

		return employees;
	});

	// Group employees by department for hierarchy view
	let employeesByDepartment = $derived(() => {
		const grouped: Record<string, Employee[]> = {};
		const employees = visibleEmployees();

		for (const emp of employees) {
			const dept = emp.department || 'Без отдела';
			if (!grouped[dept]) {
				grouped[dept] = [];
			}
			grouped[dept].push(emp);
		}

		// Sort departments alphabetically
		const sorted: Record<string, Employee[]> = {};
		Object.keys(grouped).sort().forEach(key => {
			sorted[key] = grouped[key].sort((a, b) => a.name.localeCompare(b.name));
		});

		return sorted;
	});

	// Build hierarchy tree (manager -> subordinates)
	let hierarchyTree = $derived(() => {
		const employees = visibleEmployees();
		const byId = new Map(employees.map(e => [e.id, e]));
		const children: Map<string, Employee[]> = new Map();

		// Build children map for all employees
		for (const emp of allEmployees) {
			if (emp.manager_id) {
				if (!children.has(emp.manager_id)) {
					children.set(emp.manager_id, []);
				}
				children.get(emp.manager_id)!.push(emp);
			}
		}

		// Sort children by name
		children.forEach(list => list.sort((a, b) => a.name.localeCompare(b.name)));

		// Find roots: employees whose manager is not in the visible set or is the current user
		const roots: Employee[] = [];
		for (const emp of employees) {
			const isDirectReport = emp.manager_id === $user?.id;
			const managerNotVisible = !emp.manager_id || !byId.has(emp.manager_id);
			if (isDirectReport || managerNotVisible) {
				roots.push(emp);
			}
		}

		return { roots: roots.sort((a, b) => a.name.localeCompare(b.name)), children };
	});

	function toggleDept(dept: string) {
		if (collapsedDepts.has(dept)) {
			collapsedDepts.delete(dept);
		} else {
			collapsedDepts.add(dept);
		}
		collapsedDepts = new Set(collapsedDepts);
	}

	function toggleNode(id: string) {
		if (collapsedNodes.has(id)) {
			collapsedNodes.delete(id);
		} else {
			collapsedNodes.add(id);
		}
		collapsedNodes = new Set(collapsedNodes);
	}

	onMount(async () => {
		try {
			allEmployees = await employeesApi.list();
			// Start with all departments collapsed
			const depts = new Set(allEmployees.map(e => e.department || 'Без отдела'));
			collapsedDepts = depts;
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
	<div class="flex items-center justify-between flex-wrap gap-4">
		<h1 class="text-2xl font-bold text-gray-900">Сотрудники</h1>
		<div class="flex items-center gap-4">
			<!-- My/All Toggle -->
			<div class="flex rounded-lg border border-gray-200 overflow-hidden">
				<button
					onclick={() => viewMode = 'my'}
					class="px-4 py-2 text-sm font-medium transition-colors
						{viewMode === 'my' ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}"
				>
					Мои ({$subordinates.length})
				</button>
				<button
					onclick={() => viewMode = 'all'}
					class="px-4 py-2 text-sm font-medium transition-colors
						{viewMode === 'all' ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}"
				>
					Все ({allEmployees.filter(e => e.department).length})
				</button>
			</div>

			<!-- View Mode Toggle -->
			<div class="flex rounded-lg border border-gray-200 overflow-hidden">
				<button
					onclick={() => displayMode = 'grid'}
					class="p-2 transition-colors {displayMode === 'grid' ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}"
					title="Сетка"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
					</svg>
				</button>
				<button
					onclick={() => displayMode = 'list'}
					class="p-2 transition-colors {displayMode === 'list' ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}"
					title="Список"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
					</svg>
				</button>
				<button
					onclick={() => displayMode = 'hierarchy'}
					class="p-2 transition-colors {displayMode === 'hierarchy' ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}"
					title="Оргструктура"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
					</svg>
				</button>
				<button
					onclick={() => displayMode = 'department'}
					class="p-2 transition-colors {displayMode === 'department' ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}"
					title="По департаментам"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
					</svg>
				</button>
			</div>

			<a
				href="/employees/new"
				class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors"
			>
				+ Добавить
			</a>
		</div>
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

	<!-- Content -->
	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
		</div>
	{:else if viewMode === 'my' && $subordinates.length === 0}
		<div class="text-center py-12 bg-white rounded-xl shadow-sm">
			<svg class="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
			</svg>
			<p class="text-gray-500 text-lg">У вас нет подчинённых</p>
			<button
				onclick={() => viewMode = 'all'}
				class="mt-4 text-ekf-red hover:underline"
			>
				Посмотреть всех сотрудников
			</button>
		</div>
	{:else if displayMode === 'grid'}
		<!-- Grid View -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each visibleEmployees() as employee (employee.id)}
				<a
					href="/employees/{employee.id}"
					class="bg-white rounded-xl shadow-sm p-4 hover:shadow-md transition-shadow"
				>
					<div class="flex items-center gap-4">
						<div class="w-12 h-12 rounded-full bg-gray-200 flex items-center justify-center text-xl font-medium text-gray-600 flex-shrink-0">
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
	{:else if displayMode === 'list'}
		<!-- List View -->
		<div class="bg-white rounded-xl shadow-sm overflow-hidden">
			<table class="w-full">
				<thead class="bg-gray-50 border-b">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Сотрудник</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Должность</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Отдел</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Email</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100">
					{#each visibleEmployees() as employee (employee.id)}
						<tr class="hover:bg-gray-50 cursor-pointer" onclick={() => window.location.href = `/employees/${employee.id}`}>
							<td class="px-6 py-4">
								<div class="flex items-center gap-3">
									<div class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center text-sm font-medium text-gray-600 flex-shrink-0">
										{#if employee.photo_base64}
											<img
												src="data:image/jpeg;base64,{employee.photo_base64}"
												alt=""
												class="w-full h-full rounded-full object-cover"
											/>
										{:else}
											{employee.name.charAt(0)}
										{/if}
									</div>
									<span class="font-medium text-gray-900">{employee.name}</span>
								</div>
							</td>
							<td class="px-6 py-4 text-sm text-gray-600">{employee.position || '-'}</td>
							<td class="px-6 py-4 text-sm text-gray-600">{employee.department || '-'}</td>
							<td class="px-6 py-4 text-sm text-gray-600">
								{#if employee.email}
									<a href="mailto:{employee.email}" class="text-ekf-red hover:underline" onclick={(e) => e.stopPropagation()}>
										{employee.email}
									</a>
								{:else}
									-
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{:else if displayMode === 'hierarchy'}
		<!-- Hierarchy View by Manager -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<div class="space-y-1">
				{#each hierarchyTree().roots as employee (employee.id)}
					{@render employeeNode(employee, 0)}
				{/each}
			</div>
		</div>
	{:else if displayMode === 'department'}
		<!-- Department View with Collapsible Groups -->
		<div class="space-y-3">
			{#each Object.entries(employeesByDepartment()) as [department, employees]}
				<div class="bg-white rounded-xl shadow-sm overflow-hidden">
					<button
						onclick={() => toggleDept(department)}
						class="w-full bg-gray-50 px-6 py-3 border-b flex items-center justify-between hover:bg-gray-100 transition-colors"
					>
						<h3 class="font-semibold text-gray-900 flex items-center gap-2">
							<svg class="w-5 h-5 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
							</svg>
							{department}
							<span class="text-sm font-normal text-gray-500">({employees.length})</span>
						</h3>
						<svg
							class="w-5 h-5 text-gray-400 transition-transform {collapsedDepts.has(department) ? '' : 'rotate-180'}"
							fill="none" stroke="currentColor" viewBox="0 0 24 24"
						>
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
						</svg>
					</button>
					{#if !collapsedDepts.has(department)}
						<div class="divide-y divide-gray-100">
							{#each employees as employee (employee.id)}
								<a
									href="/employees/{employee.id}"
									class="flex items-center gap-4 px-6 py-4 hover:bg-gray-50 transition-colors"
								>
									<div class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center text-sm font-medium text-gray-600 flex-shrink-0">
										{#if employee.photo_base64}
											<img src="data:image/jpeg;base64,{employee.photo_base64}" alt="" class="w-full h-full rounded-full object-cover" />
										{:else}
											{employee.name.charAt(0)}
										{/if}
									</div>
									<div class="flex-1 min-w-0">
										<p class="font-medium text-gray-900">{employee.name}</p>
										<p class="text-sm text-gray-500">{employee.position}</p>
									</div>
									{#if employee.email}
										<span class="text-sm text-gray-400 hidden md:block">{employee.email}</span>
									{/if}
								</a>
							{/each}
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}

	{#if !loading && visibleEmployees().length === 0 && (viewMode === 'all' || $subordinates.length > 0)}
		<div class="text-center py-12">
			<div class="text-gray-400 text-lg">Сотрудники не найдены</div>
		</div>
	{/if}
</div>

{#snippet employeeNode(employee: Employee, level: number)}
	{@const hasChildren = hierarchyTree().children.get(employee.id)?.length ?? 0}
	{@const isCollapsed = collapsedNodes.has(employee.id)}
	<div class="relative" style="margin-left: {level * 28}px">
		{#if level > 0}
			<div class="absolute -left-5 top-6 w-4 border-t-2 border-gray-200"></div>
			<div class="absolute -left-5 top-0 h-6 border-l-2 border-gray-200"></div>
		{/if}
		<div class="flex items-center gap-2 p-2 rounded-lg hover:bg-gray-50 transition-colors group">
			{#if hasChildren > 0}
				<button
					onclick={(e) => { e.preventDefault(); e.stopPropagation(); toggleNode(employee.id); }}
					class="w-6 h-6 flex items-center justify-center rounded hover:bg-gray-200 transition-colors flex-shrink-0"
				>
					<svg class="w-4 h-4 text-gray-500 transition-transform {isCollapsed ? '' : 'rotate-90'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
					</svg>
				</button>
			{:else}
				<div class="w-6 h-6 flex-shrink-0"></div>
			{/if}
			<a href="/employees/{employee.id}" class="flex items-center gap-3 flex-1">
				<div class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center text-sm font-medium text-gray-600 flex-shrink-0">
					{#if employee.photo_base64}
						<img src="data:image/jpeg;base64,{employee.photo_base64}" alt="" class="w-full h-full rounded-full object-cover" />
					{:else}
						{employee.name.charAt(0)}
					{/if}
				</div>
				<div class="flex-1 min-w-0">
					<p class="font-medium text-gray-900 group-hover:text-ekf-red transition-colors">{employee.name}</p>
					<p class="text-sm text-gray-500">{employee.position}</p>
				</div>
				{#if hasChildren > 0}
					<span class="text-xs text-gray-400 bg-gray-100 px-2 py-1 rounded">{hasChildren}</span>
				{/if}
			</a>
		</div>
		{#if hasChildren > 0 && !isCollapsed}
			<div class="relative">
				{#if level > 0}
					<div class="absolute -left-5 top-0 h-full border-l-2 border-gray-200"></div>
				{/if}
				{#each hierarchyTree().children.get(employee.id) || [] as child (child.id)}
					{@render employeeNode(child, level + 1)}
				{/each}
			</div>
		{/if}
	</div>
{/snippet}
