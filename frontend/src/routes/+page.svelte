<script lang="ts">
	import { onMount } from 'svelte';
	import { user, subordinates } from '$lib/stores/auth';
	import { meetings as meetingsApi, tasks as tasksApi } from '$lib/api/client';
	import type { Meeting, Task } from '$lib/api/client';

	let recentMeetings: Meeting[] = $state([]);
	let tasks: Task[] = $state([]);
	let loading = $state(true);

	// Filter subordinates with departments
	const employeesWithDept = $derived($subordinates.filter(emp => emp.department));
	const pendingTasks = $derived(tasks.filter(t => t.status === 'pending' || t.status === 'in_progress' || t.status === 'todo'));
	const overdueTasks = $derived(tasks.filter(t => t.due_date && new Date(t.due_date) < new Date() && t.status !== 'done'));

	onMount(async () => {
		if ($user) {
			await fetchData();
		} else {
			loading = false;
		}
	});

	async function fetchData() {
		try {
			const [meetingsData, tasksData] = await Promise.all([
				meetingsApi.list().catch(() => []),
				tasksApi.list({ status: 'in_progress' }).catch(() => [])
			]);
			recentMeetings = meetingsData || [];
			tasks = tasksData || [];
		} catch (error) {
			console.error('Failed to fetch dashboard data:', error);
		} finally {
			loading = false;
		}
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
	}
</script>

<svelte:head>
	<title>Дашборд - EKF Team Hub</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center h-64">
		<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
	</div>
{:else}
	<div class="space-y-6">
		<!-- Header -->
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-bold text-gray-900">
					Добро пожаловать{$user ? `, ${$user.name.split(' ')[0]}` : ''}!
				</h1>
				<p class="text-gray-500">Обзор вашей команды и задач</p>
			</div>
		</div>

		<!-- Stats Cards -->
		<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
			<a href="/employees" class="bg-white p-6 rounded-lg shadow-sm border hover:border-ekf-red transition-colors">
				<div class="flex items-center gap-3">
					<div class="w-12 h-12 bg-orange-50 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
						</svg>
					</div>
					<div>
						<p class="text-2xl font-bold text-gray-900">{$subordinates.length}</p>
						<p class="text-gray-500 text-sm">Подчинённых</p>
					</div>
				</div>
			</a>

			<a href="/calendar" class="bg-white p-6 rounded-lg shadow-sm border hover:border-ekf-red transition-colors">
				<div class="flex items-center gap-3">
					<div class="w-12 h-12 bg-green-50 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
					</div>
					<div>
						<p class="text-2xl font-bold text-gray-900">{recentMeetings.length}</p>
						<p class="text-gray-500 text-sm">Встреч</p>
					</div>
				</div>
			</a>

			<a href="/tasks" class="bg-white p-6 rounded-lg shadow-sm border hover:border-ekf-red transition-colors">
				<div class="flex items-center gap-3">
					<div class="w-12 h-12 bg-yellow-50 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<div>
						<p class="text-2xl font-bold text-gray-900">{pendingTasks.length}</p>
						<p class="text-gray-500 text-sm">Открытых задач</p>
					</div>
				</div>
			</a>

			<div class="bg-white p-6 rounded-lg shadow-sm border">
				<div class="flex items-center gap-3">
					<div class="w-12 h-12 bg-red-50 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
						</svg>
					</div>
					<div>
						<p class="text-2xl font-bold text-gray-900">{overdueTasks.length}</p>
						<p class="text-gray-500 text-sm">Просрочено</p>
					</div>
				</div>
			</div>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Recent Meetings -->
			<div class="bg-white rounded-lg shadow-sm border">
				<div class="p-4 border-b flex items-center justify-between">
					<h2 class="text-lg font-semibold text-gray-900">Последние встречи</h2>
					<a href="/meetings" class="text-ekf-red text-sm hover:underline flex items-center gap-1">
						Все
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					</a>
				</div>
				<div class="divide-y">
					{#each recentMeetings.slice(0, 5) as meeting}
						<a href="/meetings/{meeting.id}" class="block p-4 hover:bg-gray-50 transition-colors">
							<div class="flex justify-between items-start">
								<div>
									<p class="font-medium text-gray-900">{meeting.title || meeting.employees?.name || 'Встреча'}</p>
									<p class="text-sm text-gray-500">{formatDate(meeting.date)}</p>
								</div>
								{#if meeting.mood_score}
									<span class="px-2 py-1 rounded text-sm font-medium
										{meeting.mood_score >= 7 ? 'bg-green-100 text-green-700' :
										meeting.mood_score >= 5 ? 'bg-yellow-100 text-yellow-700' : 'bg-red-100 text-red-700'}">
										{meeting.mood_score}/10
									</span>
								{/if}
							</div>
							{#if meeting.summary}
								<p class="text-sm text-gray-500 mt-2 line-clamp-2">{meeting.summary}</p>
							{/if}
						</a>
					{/each}
					{#if recentMeetings.length === 0}
						<div class="p-8 text-center">
							<svg class="w-8 h-8 mx-auto text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
							</svg>
							<p class="text-gray-500">Нет встреч</p>
							<a href="/calendar" class="text-ekf-red text-sm hover:underline">
								Синхронизировать календарь
							</a>
						</div>
					{/if}
				</div>
			</div>

			<!-- Team -->
			<div class="bg-white rounded-lg shadow-sm border">
				<div class="p-4 border-b flex items-center justify-between">
					<h2 class="text-lg font-semibold text-gray-900">Команда</h2>
					<a href="/employees" class="text-ekf-red text-sm hover:underline flex items-center gap-1">
						Все
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					</a>
				</div>
				<div class="divide-y">
					{#each (employeesWithDept.length > 0 ? employeesWithDept : $subordinates).slice(0, 6) as employee}
						<a href="/employees/{employee.id}" class="flex items-center gap-3 p-4 hover:bg-gray-50 transition-colors">
							{#if employee.photo_base64}
								<img src="data:image/jpeg;base64,{employee.photo_base64}" alt="" class="w-10 h-10 rounded-full object-cover" />
							{:else}
								<div class="w-10 h-10 bg-orange-50 rounded-full flex items-center justify-center">
									<svg class="w-5 h-5 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
									</svg>
								</div>
							{/if}
							<div class="flex-1 min-w-0">
								<p class="font-medium text-gray-900 truncate">{employee.name}</p>
								<p class="text-sm text-gray-500 truncate">{employee.position}</p>
							</div>
							<svg class="w-4 h-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
							</svg>
						</a>
					{/each}
					{#if $subordinates.length === 0}
						<div class="p-8 text-center">
							<svg class="w-8 h-8 mx-auto text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
							</svg>
							<p class="text-gray-500">Нет подчинённых</p>
						</div>
					{/if}
				</div>
			</div>
		</div>

		<!-- Quick Actions -->
		<div class="bg-white rounded-lg shadow-sm border p-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Быстрые действия</h2>
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				<a href="/calendar" class="p-4 border border-gray-200 rounded-lg hover:border-ekf-red hover:bg-orange-50 transition-colors text-center">
					<svg class="w-6 h-6 mx-auto text-ekf-red mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
					</svg>
					<p class="text-sm font-medium text-gray-900">Календарь</p>
				</a>
				<a href="/script" class="p-4 border border-gray-200 rounded-lg hover:border-ekf-red hover:bg-orange-50 transition-colors text-center">
					<svg class="w-6 h-6 mx-auto text-ekf-red mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
					</svg>
					<p class="text-sm font-medium text-gray-900">Скрипт встречи</p>
				</a>
				<a href="/upload" class="p-4 border border-gray-200 rounded-lg hover:border-ekf-red hover:bg-orange-50 transition-colors text-center">
					<svg class="w-6 h-6 mx-auto text-ekf-red mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
					</svg>
					<p class="text-sm font-medium text-gray-900">Загрузить запись</p>
				</a>
				<a href="/analytics" class="p-4 border border-gray-200 rounded-lg hover:border-ekf-red hover:bg-orange-50 transition-colors text-center">
					<svg class="w-6 h-6 mx-auto text-ekf-red mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
					</svg>
					<p class="text-sm font-medium text-gray-900">Аналитика</p>
				</a>
			</div>
		</div>
	</div>
{/if}
