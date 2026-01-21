<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { admin, auth, workflows } from '$lib/api/client';
	import type { AdminStats, AdminUser, SystemSetting, AuditLog, DepartmentInfo, WorkflowMode } from '$lib/api/client';

	let isAdmin = $state(false);
	let loading = $state(true);
	let activeTab = $state<'dashboard' | 'users' | 'departments' | 'settings' | 'audit'>('dashboard');
	let error = $state('');

	// Dashboard data
	let stats = $state<AdminStats | null>(null);

	// Users data
	let users = $state<AdminUser[]>([]);
	let userSearch = $state('');
	let editingUserId = $state<string | null>(null);
	let editingRole = $state('');

	// Departments data
	let departments = $state<DepartmentInfo[]>([]);
	let workflowModes = $state<WorkflowMode[]>([]);

	// Settings data
	let settings = $state<SystemSetting[]>([]);
	let editingSettingKey = $state<string | null>(null);
	let editingSettingValue = $state('');

	// Audit logs data
	let auditLogs = $state<AuditLog[]>([]);
	let auditLoading = $state(false);

	onMount(async () => {
		try {
			const { role } = await auth.getRole();
			if (role !== 'admin' && role !== 'super_admin') {
				goto('/');
				return;
			}
			isAdmin = true;
			await loadDashboard();
		} catch (e) {
			console.error(e);
			goto('/');
		} finally {
			loading = false;
		}
	});

	async function loadDashboard() {
		try {
			stats = await admin.getStats();
		} catch (e: any) {
			error = e.message;
		}
	}

	async function loadUsers() {
		try {
			users = await admin.listUsers();
		} catch (e: any) {
			error = e.message;
		}
	}

	async function loadDepartments() {
		try {
			const [depts, modes] = await Promise.all([
				admin.getDepartments(),
				workflows.list()
			]);
			departments = depts || [];
			workflowModes = modes || [];
		} catch (e: any) {
			error = e.message;
		}
	}

	async function loadSettings() {
		try {
			settings = await admin.getSettings();
		} catch (e: any) {
			error = e.message;
		}
	}

	async function loadAuditLogs() {
		auditLoading = true;
		try {
			auditLogs = await admin.getAuditLogs({ limit: 100 });
		} catch (e: any) {
			error = e.message;
		} finally {
			auditLoading = false;
		}
	}

	async function switchTab(tab: typeof activeTab) {
		activeTab = tab;
		error = '';
		if (tab === 'dashboard' && !stats) await loadDashboard();
		if (tab === 'users' && users.length === 0) await loadUsers();
		if (tab === 'departments' && departments.length === 0) await loadDepartments();
		if (tab === 'settings' && settings.length === 0) await loadSettings();
		if (tab === 'audit' && auditLogs.length === 0) await loadAuditLogs();
	}

	async function updateUserRole(userId: string, role: string) {
		try {
			await admin.updateUserRole(userId, role);
			editingUserId = null;
			await loadUsers();
		} catch (e: any) {
			error = e.message;
		}
	}

	async function updateDepartmentWorkflow(department: string, workflowModeId: string) {
		try {
			await workflows.setDepartmentWorkflow(department, workflowModeId);
			await loadDepartments();
		} catch (e: any) {
			error = e.message;
		}
	}

	async function updateSetting(key: string, value: string) {
		try {
			// Parse value if it looks like JSON
			let parsedValue: any = value;
			try {
				parsedValue = JSON.parse(value);
			} catch {
				// Keep as string
			}
			await admin.updateSetting(key, parsedValue);
			editingSettingKey = null;
			await loadSettings();
		} catch (e: any) {
			error = e.message;
		}
	}

	function getRoleBadgeColor(role: string): string {
		switch (role) {
			case 'super_admin': return 'bg-red-100 text-red-800';
			case 'admin': return 'bg-purple-100 text-purple-800';
			default: return 'bg-gray-100 text-gray-800';
		}
	}

	function getRoleLabel(role: string): string {
		switch (role) {
			case 'super_admin': return 'Супер-админ';
			case 'admin': return 'Админ';
			default: return 'Пользователь';
		}
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleString('ru-RU');
	}

	$effect(() => {
		if (userSearch) {
			// Filter handled in template
		}
	});

	const filteredUsers = $derived(
		userSearch
			? users.filter(u =>
				u.name.toLowerCase().includes(userSearch.toLowerCase()) ||
				u.email.toLowerCase().includes(userSearch.toLowerCase()) ||
				(u.department || '').toLowerCase().includes(userSearch.toLowerCase())
			)
			: users
	);
</script>

<svelte:head>
	<title>Админ-панель - EKF Hub</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center h-64">
		<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
	</div>
{:else if !isAdmin}
	<div class="flex items-center justify-center h-64">
		<p class="text-gray-500">Доступ запрещён</p>
	</div>
{:else}
	<div class="max-w-7xl mx-auto space-y-6">
		<div class="flex items-center justify-between">
			<h1 class="text-2xl font-bold text-gray-900">Админ-панель</h1>
		</div>

		{#if error}
			<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
				{error}
			</div>
		{/if}

		<!-- Tabs -->
		<div class="border-b border-gray-200">
			<nav class="-mb-px flex space-x-8">
				<button
					onclick={() => switchTab('dashboard')}
					class="py-2 px-1 border-b-2 font-medium text-sm transition-colors {activeTab === 'dashboard' ? 'border-ekf-red text-ekf-red' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
				>
					Обзор
				</button>
				<button
					onclick={() => switchTab('users')}
					class="py-2 px-1 border-b-2 font-medium text-sm transition-colors {activeTab === 'users' ? 'border-ekf-red text-ekf-red' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
				>
					Пользователи
				</button>
				<button
					onclick={() => switchTab('departments')}
					class="py-2 px-1 border-b-2 font-medium text-sm transition-colors {activeTab === 'departments' ? 'border-ekf-red text-ekf-red' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
				>
					Департаменты
				</button>
				<button
					onclick={() => switchTab('settings')}
					class="py-2 px-1 border-b-2 font-medium text-sm transition-colors {activeTab === 'settings' ? 'border-ekf-red text-ekf-red' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
				>
					Настройки
				</button>
				<button
					onclick={() => switchTab('audit')}
					class="py-2 px-1 border-b-2 font-medium text-sm transition-colors {activeTab === 'audit' ? 'border-ekf-red text-ekf-red' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
				>
					Аудит
				</button>
			</nav>
		</div>

		<!-- Dashboard Tab -->
		{#if activeTab === 'dashboard'}
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				<div class="bg-white rounded-xl shadow-sm p-6">
					<div class="flex items-center gap-3">
						<div class="w-12 h-12 rounded-lg bg-blue-100 flex items-center justify-center">
							<svg class="w-6 h-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
							</svg>
						</div>
						<div>
							<div class="text-2xl font-bold text-gray-900">{stats?.total_users || 0}</div>
							<div class="text-sm text-gray-500">Пользователей</div>
						</div>
					</div>
				</div>

				<div class="bg-white rounded-xl shadow-sm p-6">
					<div class="flex items-center gap-3">
						<div class="w-12 h-12 rounded-lg bg-purple-100 flex items-center justify-center">
							<svg class="w-6 h-6 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
							</svg>
						</div>
						<div>
							<div class="text-2xl font-bold text-gray-900">{stats?.admin_count || 0}</div>
							<div class="text-sm text-gray-500">Админов</div>
						</div>
					</div>
				</div>

				<div class="bg-white rounded-xl shadow-sm p-6">
					<div class="flex items-center gap-3">
						<div class="w-12 h-12 rounded-lg bg-green-100 flex items-center justify-center">
							<svg class="w-6 h-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
							</svg>
						</div>
						<div>
							<div class="text-2xl font-bold text-gray-900">{stats?.total_tasks || 0}</div>
							<div class="text-sm text-gray-500">Задач</div>
						</div>
					</div>
				</div>

				<div class="bg-white rounded-xl shadow-sm p-6">
					<div class="flex items-center gap-3">
						<div class="w-12 h-12 rounded-lg bg-amber-100 flex items-center justify-center">
							<svg class="w-6 h-6 text-amber-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
							</svg>
						</div>
						<div>
							<div class="text-2xl font-bold text-gray-900">{stats?.departments_count || 0}</div>
							<div class="text-sm text-gray-500">Департаментов</div>
						</div>
					</div>
				</div>
			</div>

			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<div class="bg-white rounded-xl shadow-sm p-6">
					<h3 class="text-lg font-semibold text-gray-900 mb-4">Задачи</h3>
					<div class="space-y-3">
						<div class="flex justify-between items-center">
							<span class="text-gray-600">Всего задач</span>
							<span class="font-semibold">{stats?.total_tasks || 0}</span>
						</div>
						<div class="flex justify-between items-center">
							<span class="text-gray-600">Выполнено</span>
							<span class="font-semibold text-green-600">{stats?.completed_tasks || 0}</span>
						</div>
						<div class="w-full bg-gray-200 rounded-full h-2">
							<div
								class="bg-green-500 h-2 rounded-full transition-all"
								style="width: {stats?.total_tasks ? (stats.completed_tasks / stats.total_tasks * 100) : 0}%"
							></div>
						</div>
					</div>
				</div>

				<div class="bg-white rounded-xl shadow-sm p-6">
					<h3 class="text-lg font-semibold text-gray-900 mb-4">Активность</h3>
					<div class="space-y-3">
						<div class="flex justify-between items-center">
							<span class="text-gray-600">Встречи</span>
							<span class="font-semibold">{stats?.total_meetings || 0}</span>
						</div>
						<div class="flex justify-between items-center">
							<span class="text-gray-600">Сообщения</span>
							<span class="font-semibold">{stats?.total_messages || 0}</span>
						</div>
					</div>
				</div>
			</div>
		{/if}

		<!-- Users Tab -->
		{#if activeTab === 'users'}
			<div class="bg-white rounded-xl shadow-sm">
				<div class="p-4 border-b border-gray-200">
					<input
						type="text"
						bind:value={userSearch}
						placeholder="Поиск по имени, email или департаменту..."
						class="w-full md:w-96 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
					/>
				</div>
				<div class="overflow-x-auto">
					<table class="w-full">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Пользователь</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Департамент</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Должность</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Роль</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Действия</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-200">
							{#each filteredUsers as user (user.id)}
								<tr class="hover:bg-gray-50">
									<td class="px-6 py-4 whitespace-nowrap">
										<div>
											<div class="font-medium text-gray-900">{user.name}</div>
											<div class="text-sm text-gray-500">{user.email}</div>
										</div>
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
										{user.department || '-'}
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
										{user.position}
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										{#if editingUserId === user.id}
											<select
												bind:value={editingRole}
												class="px-2 py-1 border border-gray-300 rounded text-sm"
											>
												<option value="user">Пользователь</option>
												<option value="admin">Админ</option>
												<option value="super_admin">Супер-админ</option>
											</select>
										{:else}
											<span class="px-2 py-1 text-xs font-medium rounded-full {getRoleBadgeColor(user.role || 'user')}">
												{getRoleLabel(user.role || 'user')}
											</span>
										{/if}
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm">
										{#if editingUserId === user.id}
											<button
												onclick={() => updateUserRole(user.id, editingRole)}
												class="text-green-600 hover:text-green-800 mr-2"
											>
												Сохранить
											</button>
											<button
												onclick={() => editingUserId = null}
												class="text-gray-600 hover:text-gray-800"
											>
												Отмена
											</button>
										{:else}
											<button
												onclick={() => { editingUserId = user.id; editingRole = user.role || 'user'; }}
												class="text-blue-600 hover:text-blue-800"
											>
												Изменить роль
											</button>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}

		<!-- Departments Tab -->
		{#if activeTab === 'departments'}
			<div class="bg-white rounded-xl shadow-sm">
				<div class="p-4 border-b border-gray-200">
					<h3 class="text-lg font-semibold text-gray-900">Workflow по департаментам</h3>
					<p class="text-sm text-gray-500 mt-1">Настройка режимов работы для каждого департамента</p>
				</div>
				<div class="overflow-x-auto">
					<table class="w-full">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Департамент</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Сотрудников</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Workflow</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-200">
							{#each departments as dept (dept.name)}
								<tr class="hover:bg-gray-50">
									<td class="px-6 py-4 whitespace-nowrap">
										<div class="font-medium text-gray-900">{dept.name}</div>
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
										{dept.employee_count}
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<select
											value={dept.workflow_mode_id || ''}
											onchange={(e) => updateDepartmentWorkflow(dept.name, e.currentTarget.value)}
											class="px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-ekf-red focus:border-transparent"
										>
											<option value="">По умолчанию (simple)</option>
											{#each workflowModes as mode (mode.id)}
												<option value={mode.id}>{mode.name} - {mode.description}</option>
											{/each}
										</select>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>

			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">Доступные Workflow режимы</h3>
				<div class="space-y-4">
					{#each workflowModes as mode (mode.id)}
						<div class="p-4 border border-gray-200 rounded-lg">
							<div class="flex items-center justify-between">
								<div>
									<div class="font-medium text-gray-900">{mode.name}</div>
									<div class="text-sm text-gray-500">{mode.description}</div>
								</div>
								{#if mode.is_default}
									<span class="px-2 py-1 text-xs font-medium bg-green-100 text-green-800 rounded-full">По умолчанию</span>
								{/if}
							</div>
							<div class="mt-3 flex flex-wrap gap-2">
								{#each mode.statuses as status}
									<span class="px-2 py-1 text-xs rounded {status.color} text-gray-700">
										{status.label}
									</span>
								{/each}
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Settings Tab -->
		{#if activeTab === 'settings'}
			<div class="bg-white rounded-xl shadow-sm">
				<div class="p-4 border-b border-gray-200">
					<h3 class="text-lg font-semibold text-gray-900">Системные настройки</h3>
				</div>
				<div class="divide-y divide-gray-200">
					{#each settings as setting (setting.id)}
						<div class="p-4 hover:bg-gray-50">
							<div class="flex items-center justify-between">
								<div class="flex-1">
									<div class="font-medium text-gray-900">{setting.key}</div>
									<div class="text-sm text-gray-500">{setting.description || 'Нет описания'}</div>
								</div>
								<div class="flex items-center gap-4">
									{#if editingSettingKey === setting.key}
										<input
											type="text"
											bind:value={editingSettingValue}
											class="px-3 py-1.5 border border-gray-300 rounded-lg text-sm w-48"
										/>
										<button
											onclick={() => updateSetting(setting.key, editingSettingValue)}
											class="text-green-600 hover:text-green-800"
										>
											Сохранить
										</button>
										<button
											onclick={() => editingSettingKey = null}
											class="text-gray-600 hover:text-gray-800"
										>
											Отмена
										</button>
									{:else}
										<code class="px-2 py-1 bg-gray-100 rounded text-sm">
											{JSON.stringify(setting.value)}
										</code>
										<button
											onclick={() => { editingSettingKey = setting.key; editingSettingValue = JSON.stringify(setting.value); }}
											class="text-blue-600 hover:text-blue-800 text-sm"
										>
											Изменить
										</button>
									{/if}
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Audit Tab -->
		{#if activeTab === 'audit'}
			<div class="bg-white rounded-xl shadow-sm">
				<div class="p-4 border-b border-gray-200 flex items-center justify-between">
					<h3 class="text-lg font-semibold text-gray-900">Журнал аудита</h3>
					<button
						onclick={() => loadAuditLogs()}
						disabled={auditLoading}
						class="px-3 py-1.5 text-sm bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors disabled:opacity-50"
					>
						{auditLoading ? 'Загрузка...' : 'Обновить'}
					</button>
				</div>
				<div class="overflow-x-auto">
					<table class="w-full">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Время</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Пользователь</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Действие</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Объект</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Изменения</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-200">
							{#each auditLogs as log (log.id)}
								<tr class="hover:bg-gray-50">
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
										{formatDate(log.created_at)}
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm">
										{log.user?.name || 'Система'}
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<span class="px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded-full">
											{log.action}
										</span>
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
										{log.entity_type || '-'}
									</td>
									<td class="px-6 py-4 text-sm text-gray-500 max-w-xs truncate">
										{#if log.old_value || log.new_value}
											<span class="text-red-600">{JSON.stringify(log.old_value)}</span>
											<span class="mx-1">→</span>
											<span class="text-green-600">{JSON.stringify(log.new_value)}</span>
										{:else}
											-
										{/if}
									</td>
								</tr>
							{/each}
							{#if auditLogs.length === 0}
								<tr>
									<td colspan="5" class="px-6 py-8 text-center text-gray-500">
										Нет записей в журнале аудита
									</td>
								</tr>
							{/if}
						</tbody>
					</table>
				</div>
			</div>
		{/if}
	</div>
{/if}
