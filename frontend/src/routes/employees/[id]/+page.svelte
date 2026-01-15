<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { employees as employeesApi, analytics } from '$lib/api/client';
	import type { Employee, EmployeeDossier, EmployeeAnalytics } from '$lib/api/client';

	let employee: Employee | null = $state(null);
	let dossier: EmployeeDossier | null = $state(null);
	let employeeAnalytics: EmployeeAnalytics | null = $state(null);
	let loading = $state(true);
	let activeTab = $state('overview');

	const id = $page.params.id;

	onMount(async () => {
		try {
			[employee, dossier, employeeAnalytics] = await Promise.all([
				employeesApi.get(id),
				employeesApi.getDossier(id),
				analytics.getEmployee(id)
			]);
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>{employee?.name || 'Сотрудник'} - EKF Team Hub</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center h-64">
		<div class="text-gray-500">Загрузка...</div>
	</div>
{:else if employee}
	<div class="space-y-6">
		<!-- Header -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<div class="flex items-start gap-6">
				<div class="w-24 h-24 rounded-full bg-gray-200 flex items-center justify-center text-3xl font-medium text-gray-600 flex-shrink-0">
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
				<div class="flex-1">
					<h1 class="text-2xl font-bold text-gray-900">{employee.name}</h1>
					<p class="text-lg text-gray-600">{employee.position}</p>
					{#if employee.department}
						<p class="text-gray-500">{employee.department}</p>
					{/if}
					{#if employee.email}
						<a href="mailto:{employee.email}" class="text-ekf-red hover:underline">{employee.email}</a>
					{/if}
				</div>
				<a
					href="/employees/{id}/edit"
					class="px-4 py-2 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
				>
					Редактировать
				</a>
			</div>
		</div>

		<!-- Tabs -->
		<div class="border-b border-gray-200">
			<nav class="flex gap-8">
				{#each [
					{ key: 'overview', label: 'Обзор' },
					{ key: 'meetings', label: 'Встречи' },
					{ key: 'tasks', label: 'Задачи' },
					{ key: 'analytics', label: 'Аналитика' }
				] as tab}
					<button
						onclick={() => activeTab = tab.key}
						class="py-4 px-1 border-b-2 font-medium text-sm transition-colors
							{activeTab === tab.key
								? 'border-ekf-red text-ekf-red'
								: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
					>
						{tab.label}
					</button>
				{/each}
			</nav>
		</div>

		<!-- Content -->
		{#if activeTab === 'overview' && dossier}
			<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
				<div class="bg-white rounded-xl shadow-sm p-6">
					<h3 class="font-semibold text-gray-900 mb-2">Встречи 1-на-1</h3>
					<p class="text-3xl font-bold text-ekf-red">{dossier.one_on_one_count}</p>
				</div>
				<div class="bg-white rounded-xl shadow-sm p-6">
					<h3 class="font-semibold text-gray-900 mb-2">Задачи</h3>
					<p class="text-3xl font-bold text-gray-900">{dossier.tasks.total}</p>
					<p class="text-sm text-gray-500">
						{dossier.tasks.done} выполнено, {dossier.tasks.in_progress} в работе
					</p>
				</div>
				<div class="bg-white rounded-xl shadow-sm p-6">
					<h3 class="font-semibold text-gray-900 mb-2">Последнее настроение</h3>
					{#if dossier.mood_history.length > 0}
						{@const lastMood = dossier.mood_history[dossier.mood_history.length - 1]}
						<p class="text-3xl font-bold {lastMood.score >= 7 ? 'text-green-600' : lastMood.score >= 5 ? 'text-yellow-600' : 'text-red-600'}">
							{lastMood.score}/10
						</p>
						<p class="text-sm text-gray-500">{lastMood.date}</p>
					{:else}
						<p class="text-gray-400">Нет данных</p>
					{/if}
				</div>
			</div>

			<!-- Recent Meetings -->
			{#if dossier.recent_meetings.length > 0}
				<div class="bg-white rounded-xl shadow-sm p-6">
					<h3 class="font-semibold text-gray-900 mb-4">Последние встречи</h3>
					<div class="space-y-3">
						{#each dossier.recent_meetings as meeting}
							<a href="/meetings/{meeting.id}" class="block p-3 rounded-lg hover:bg-gray-50">
								<div class="flex justify-between items-center">
									<div>
										<p class="font-medium">{meeting.title || 'Без названия'}</p>
										<p class="text-sm text-gray-500">{meeting.date}</p>
									</div>
									{#if meeting.mood_score}
										<span class="text-sm font-medium">{meeting.mood_score}/10</span>
									{/if}
								</div>
								{#if meeting.summary}
									<p class="text-sm text-gray-600 mt-2 line-clamp-2">{meeting.summary}</p>
								{/if}
							</a>
						{/each}
					</div>
				</div>
			{/if}
		{/if}

		{#if activeTab === 'analytics' && employeeAnalytics}
			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<!-- Task Stats -->
				<div class="bg-white rounded-xl shadow-sm p-6">
					<h3 class="font-semibold text-gray-900 mb-4">Статистика задач</h3>
					<div class="space-y-3">
						<div class="flex justify-between">
							<span class="text-gray-600">Всего</span>
							<span class="font-medium">{employeeAnalytics.task_stats.total}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-gray-600">Выполнено</span>
							<span class="font-medium text-green-600">{employeeAnalytics.task_stats.done}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-gray-600">В работе</span>
							<span class="font-medium text-blue-600">{employeeAnalytics.task_stats.in_progress}</span>
						</div>
					</div>
				</div>

				<!-- Agreement Stats -->
				<div class="bg-white rounded-xl shadow-sm p-6">
					<h3 class="font-semibold text-gray-900 mb-4">Договорённости</h3>
					<div class="space-y-3">
						<div class="flex justify-between">
							<span class="text-gray-600">Всего</span>
							<span class="font-medium">{employeeAnalytics.agreement_stats.total}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-gray-600">Выполнено</span>
							<span class="font-medium text-green-600">{employeeAnalytics.agreement_stats.completed}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-gray-600">В ожидании</span>
							<span class="font-medium">{employeeAnalytics.agreement_stats.pending}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-gray-600">Просрочено</span>
							<span class="font-medium text-red-600">{employeeAnalytics.agreement_stats.overdue}</span>
						</div>
					</div>
				</div>

				<!-- Mood History -->
				<div class="bg-white rounded-xl shadow-sm p-6 md:col-span-2">
					<h3 class="font-semibold text-gray-900 mb-4">История настроения</h3>
					{#if employeeAnalytics.mood_history.length > 0}
						<div class="flex items-end gap-2 h-32">
							{#each employeeAnalytics.mood_history.slice(-12) as mood}
								<div class="flex-1 flex flex-col items-center gap-1">
									<div
										class="w-full rounded-t transition-all
											{mood.score >= 7 ? 'bg-green-500' : mood.score >= 5 ? 'bg-yellow-500' : 'bg-red-500'}"
										style="height: {mood.score * 10}%"
									></div>
									<span class="text-xs text-gray-500">{mood.score}</span>
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-gray-400">Нет данных о настроении</p>
					{/if}
				</div>
			</div>
		{/if}
	</div>
{:else}
	<div class="text-center py-12">
		<div class="text-gray-400 text-lg">Сотрудник не найден</div>
	</div>
{/if}
