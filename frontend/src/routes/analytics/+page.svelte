<script lang="ts">
	import { onMount } from 'svelte';
	import { analytics as analyticsApi } from '$lib/api/client';
	import type { DashboardData } from '$lib/api/client';

	let dashboard: DashboardData | null = $state(null);
	let loading = $state(true);
	let selectedPeriod = $state('month');

	onMount(async () => {
		try {
			dashboard = await analyticsApi.getDashboard();
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Аналитика - EKF Team Hub</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-gray-900">Аналитика</h1>
		<div class="flex gap-2">
			{#each [
				{ key: 'week', label: 'Неделя' },
				{ key: 'month', label: 'Месяц' },
				{ key: 'quarter', label: 'Квартал' },
				{ key: 'year', label: 'Год' }
			] as period}
				<button
					onclick={() => selectedPeriod = period.key}
					class="px-4 py-2 rounded-lg text-sm font-medium transition-colors
						{selectedPeriod === period.key
							? 'bg-ekf-red text-white'
							: 'bg-white text-gray-600 hover:bg-gray-100'}"
				>
					{period.label}
				</button>
			{/each}
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="text-gray-500">Загрузка...</div>
		</div>
	{:else if dashboard}
		<!-- Overview Stats -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="text-sm font-medium text-gray-500">Всего сотрудников</h3>
				<p class="text-3xl font-bold text-gray-900 mt-2">{dashboard.total_employees}</p>
			</div>
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="text-sm font-medium text-gray-500">Встречи за период</h3>
				<p class="text-3xl font-bold text-ekf-red mt-2">{dashboard.meetings_this_month}</p>
			</div>
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="text-sm font-medium text-gray-500">Среднее настроение</h3>
				<p class="text-3xl font-bold mt-2
					{dashboard.average_mood >= 7 ? 'text-green-600' :
					dashboard.average_mood >= 5 ? 'text-yellow-600' : 'text-red-600'}">
					{dashboard.average_mood.toFixed(1)}/10
				</p>
			</div>
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="text-sm font-medium text-gray-500">Задач выполнено</h3>
				<p class="text-3xl font-bold text-green-600 mt-2">{dashboard.tasks_completed}</p>
			</div>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Mood Trend -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-4">Динамика настроения команды</h3>
				{#if dashboard.mood_trend && dashboard.mood_trend.length > 0}
					<div class="flex items-end gap-2 h-40">
						{#each dashboard.mood_trend as point}
							<div class="flex-1 flex flex-col items-center gap-1">
								<div
									class="w-full rounded-t transition-all
										{point.score >= 7 ? 'bg-green-500' : point.score >= 5 ? 'bg-yellow-500' : 'bg-red-500'}"
									style="height: {point.score * 10}%"
								></div>
								<span class="text-xs text-gray-500">{point.score.toFixed(1)}</span>
							</div>
						{/each}
					</div>
				{:else}
					<p class="text-gray-400 text-center py-8">Недостаточно данных</p>
				{/if}
			</div>

			<!-- Risk Employees -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-4">Требуют внимания</h3>
				{#if dashboard.employees_needing_attention && dashboard.employees_needing_attention.length > 0}
					<div class="space-y-3">
						{#each dashboard.employees_needing_attention as emp}
							<a href="/employees/{emp.id}" class="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50">
								<div class="flex items-center gap-3">
									<div class="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center text-red-600 font-medium">
										{emp.name.charAt(0)}
									</div>
									<div>
										<p class="font-medium text-gray-900">{emp.name}</p>
										<p class="text-sm text-gray-500">{emp.reason}</p>
									</div>
								</div>
								<span class="text-sm font-medium text-red-600">
									{emp.days_since_meeting} дней
								</span>
							</a>
						{/each}
					</div>
				{:else}
					<p class="text-gray-400 text-center py-8">Все сотрудники в порядке</p>
				{/if}
			</div>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Meeting Categories -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-4">Типы встреч</h3>
				{#if dashboard.meetings_by_category}
					<div class="space-y-3">
						{#each Object.entries(dashboard.meetings_by_category) as [category, count]}
							<div class="flex items-center justify-between">
								<span class="text-gray-600">
									{category === 'one_on_one' ? '1-на-1' :
									category === 'project' ? 'Проектные' : 'Командные'}
								</span>
								<span class="font-medium text-gray-900">{count}</span>
							</div>
						{/each}
					</div>
				{:else}
					<p class="text-gray-400 text-center py-4">Нет данных</p>
				{/if}
			</div>

			<!-- Task Stats -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-4">Статистика задач</h3>
				<div class="space-y-3">
					<div class="flex items-center justify-between">
						<span class="text-gray-600">К выполнению</span>
						<span class="font-medium text-gray-900">{dashboard.tasks_todo || 0}</span>
					</div>
					<div class="flex items-center justify-between">
						<span class="text-gray-600">В работе</span>
						<span class="font-medium text-blue-600">{dashboard.tasks_in_progress || 0}</span>
					</div>
					<div class="flex items-center justify-between">
						<span class="text-gray-600">Выполнено</span>
						<span class="font-medium text-green-600">{dashboard.tasks_completed || 0}</span>
					</div>
				</div>
			</div>

			<!-- Agreements -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-4">Договорённости</h3>
				<div class="space-y-3">
					<div class="flex items-center justify-between">
						<span class="text-gray-600">Всего</span>
						<span class="font-medium text-gray-900">{dashboard.agreements_total || 0}</span>
					</div>
					<div class="flex items-center justify-between">
						<span class="text-gray-600">Выполнено</span>
						<span class="font-medium text-green-600">{dashboard.agreements_completed || 0}</span>
					</div>
					<div class="flex items-center justify-between">
						<span class="text-gray-600">Просрочено</span>
						<span class="font-medium text-red-600">{dashboard.agreements_overdue || 0}</span>
					</div>
				</div>
			</div>
		</div>

		<!-- Recent Meetings -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h3 class="font-semibold text-gray-900 mb-4">Последние встречи</h3>
			{#if dashboard.recent_meetings && dashboard.recent_meetings.length > 0}
				<div class="space-y-3">
					{#each dashboard.recent_meetings as meeting}
						<a href="/meetings/{meeting.id}" class="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50">
							<div>
								<p class="font-medium text-gray-900">{meeting.title || 'Без названия'}</p>
								<p class="text-sm text-gray-500">{meeting.date} • {meeting.employee_name}</p>
							</div>
							{#if meeting.mood_score}
								<span class="text-sm font-medium
									{meeting.mood_score >= 7 ? 'text-green-600' :
									meeting.mood_score >= 5 ? 'text-yellow-600' : 'text-red-600'}">
									{meeting.mood_score}/10
								</span>
							{/if}
						</a>
					{/each}
				</div>
			{:else}
				<p class="text-gray-400 text-center py-8">Нет встреч за выбранный период</p>
			{/if}
		</div>
	{:else}
		<div class="text-center py-12">
			<div class="text-gray-400 text-lg">Не удалось загрузить данные</div>
		</div>
	{/if}
</div>
