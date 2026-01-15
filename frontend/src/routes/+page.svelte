<script lang="ts">
	import { onMount } from 'svelte';
	import { analytics, connector } from '$lib/api/client';
	import type { DashboardData, ConnectorStatus } from '$lib/api/client';

	let dashboard: DashboardData | null = $state(null);
	let connectorStatus: ConnectorStatus | null = $state(null);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			[dashboard, connectorStatus] = await Promise.all([
				analytics.getDashboard(),
				connector.status()
			]);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Ошибка загрузки';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Дашборд - EKF Team Hub</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center h-64">
		<div class="text-gray-500">Загрузка...</div>
	</div>
{:else if error}
	<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
		{error}
	</div>
{:else if dashboard}
	<div class="space-y-6">
		<!-- Stats Cards -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
			<div class="bg-white rounded-xl shadow-sm p-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm text-gray-500">Сотрудники</p>
						<p class="text-2xl font-bold text-gray-900">{dashboard.employees.length}</p>
					</div>
					<div class="text-3xl">👥</div>
				</div>
			</div>

			<div class="bg-white rounded-xl shadow-sm p-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm text-gray-500">Активные проекты</p>
						<p class="text-2xl font-bold text-gray-900">{dashboard.projects.length}</p>
					</div>
					<div class="text-3xl">📁</div>
				</div>
			</div>

			<div class="bg-white rounded-xl shadow-sm p-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm text-gray-500">Задачи в работе</p>
						<p class="text-2xl font-bold text-gray-900">{dashboard.task_summary.in_progress}</p>
					</div>
					<div class="text-3xl">✅</div>
				</div>
			</div>

			<div class="bg-white rounded-xl shadow-sm p-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm text-gray-500">Просрочено</p>
						<p class="text-2xl font-bold {dashboard.task_summary.overdue > 0 ? 'text-red-600' : 'text-gray-900'}">
							{dashboard.task_summary.overdue}
						</p>
					</div>
					<div class="text-3xl">⚠️</div>
				</div>
			</div>
		</div>

		<!-- Connector Status -->
		{#if connectorStatus}
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h2 class="text-lg font-semibold mb-4">Статус подключений</h2>
				<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
					<div class="flex items-center gap-3">
						<div class="w-3 h-3 rounded-full {connectorStatus.connected ? 'bg-green-500' : 'bg-red-500'}"></div>
						<span class="text-sm">
							Коннектор: {connectorStatus.connected ? 'Подключен' : 'Отключен'}
						</span>
					</div>
					<div class="flex items-center gap-3">
						<div class="w-3 h-3 rounded-full {connectorStatus.ad_status === 'connected' ? 'bg-green-500' : 'bg-yellow-500'}"></div>
						<span class="text-sm">
							Active Directory: {connectorStatus.employee_count} сотрудников
						</span>
					</div>
					<div class="flex items-center gap-3">
						<div class="w-3 h-3 rounded-full bg-green-500"></div>
						<span class="text-sm">
							Exchange (EWS): {connectorStatus.ews_url ? 'Настроен' : 'Не настроен'}
						</span>
					</div>
				</div>
			</div>
		{/if}

		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Recent Meetings -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h2 class="text-lg font-semibold mb-4">Последние встречи</h2>
				{#if dashboard.recent_meetings.length === 0}
					<p class="text-gray-500 text-sm">Нет встреч</p>
				{:else}
					<div class="space-y-3">
						{#each dashboard.recent_meetings.slice(0, 5) as meeting}
							<a
								href="/meetings/{meeting.id}"
								class="block p-3 rounded-lg hover:bg-gray-50 transition-colors"
							>
								<div class="flex items-center justify-between">
									<div>
										<p class="font-medium text-gray-900">{meeting.title || 'Без названия'}</p>
										<p class="text-sm text-gray-500">
											{meeting.employees?.name || 'Общая встреча'} • {meeting.date}
										</p>
									</div>
									{#if meeting.mood_score}
										<div class="text-sm font-medium {meeting.mood_score >= 7 ? 'text-green-600' : meeting.mood_score >= 5 ? 'text-yellow-600' : 'text-red-600'}">
											{meeting.mood_score}/10
										</div>
									{/if}
								</div>
							</a>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Red Flags -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
					<span>⚠️</span> Сигналы внимания
				</h2>
				{#if dashboard.red_flags.length === 0}
					<p class="text-gray-500 text-sm">Нет активных сигналов</p>
				{:else}
					<div class="space-y-3">
						{#each dashboard.red_flags as flag}
							<div class="p-3 rounded-lg bg-red-50 border border-red-100">
								<div class="font-medium text-red-800">{flag.employee}</div>
								<div class="text-sm text-red-600">{flag.date}</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}
