<script lang="ts">
	import { onMount } from 'svelte';
	import { user, subordinates } from '$lib/stores/auth';
	import { meetings as meetingsApi, tasks as tasksApi, analytics } from '$lib/api/client';
	import type { Meeting, Task, TeamMemberStats } from '$lib/api/client';

	let recentMeetings: Meeting[] = $state([]);
	let tasks: Task[] = $state([]);
	let teamStats: TeamMemberStats[] = $state([]);
	let loading = $state(true);

	// Filter subordinates with departments
	const employeesWithDept = $derived($subordinates.filter(emp => emp.department));
	const pendingTasks = $derived(tasks.filter(t => ['backlog', 'todo', 'in_progress', 'review'].includes(t.status)));
	const overdueTasks = $derived(tasks.filter(t => t.due_date && new Date(t.due_date) < new Date() && t.status !== 'done'));

	// Extract greeting name (first name or patronymic from "Фамилия Имя Отчество")
	function getGreetingName(fullName: string): string {
		const parts = fullName.trim().split(/\s+/);
		// Russian names are typically "Фамилия Имя Отчество"
		if (parts.length >= 3) {
			// Return "Имя Отчество"
			return `${parts[1]} ${parts[2]}`;
		} else if (parts.length === 2) {
			// Return first name
			return parts[1];
		}
		return parts[0];
	}

	onMount(async () => {
		if ($user) {
			await fetchData();
		} else {
			loading = false;
		}
	});

	async function fetchData() {
		try {
			const [meetingsData, tasksData, teamData] = await Promise.all([
				meetingsApi.list().catch(() => []),
				tasksApi.list().catch(() => []),
				$user ? analytics.getTeamStats($user.id).catch(() => []) : Promise.resolve([])
			]);
			recentMeetings = meetingsData || [];
			tasks = tasksData || [];
			teamStats = teamData || [];
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
	<title>Дашборд - EKF Hub</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center h-48">
		<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
	</div>
{:else}
	<div class="space-y-4">
		<!-- Header -->
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-xl font-bold text-gray-900">
					Добро пожаловать{$user ? `, ${getGreetingName($user.name)}` : ''}!
				</h1>
				<p class="text-sm text-gray-500">Обзор вашей команды и задач</p>
			</div>
		</div>

		<!-- Stats Cards -->
		<div class="grid grid-cols-2 md:grid-cols-4 gap-3">
			<a href="/employees" class="bg-white p-4 rounded-lg border hover:border-ekf-red transition-colors">
				<div class="flex items-center gap-3">
					<div class="w-10 h-10 bg-orange-50 rounded-lg flex items-center justify-center">
						<svg class="w-5 h-5 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
						</svg>
					</div>
					<div>
						<p class="text-xl font-bold text-gray-900">{$subordinates.length}</p>
						<p class="text-gray-500 text-xs">Подчинённых</p>
					</div>
				</div>
			</a>

			<a href="/calendar" class="bg-white p-4 rounded-lg border hover:border-ekf-red transition-colors">
				<div class="flex items-center gap-3">
					<div class="w-10 h-10 bg-green-50 rounded-lg flex items-center justify-center">
						<svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
					</div>
					<div>
						<p class="text-xl font-bold text-gray-900">{recentMeetings.length}</p>
						<p class="text-gray-500 text-xs">Встреч</p>
					</div>
				</div>
			</a>

			<a href="/tasks" class="bg-white p-4 rounded-lg border hover:border-ekf-red transition-colors">
				<div class="flex items-center gap-3">
					<div class="w-10 h-10 bg-yellow-50 rounded-lg flex items-center justify-center">
						<svg class="w-5 h-5 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<div>
						<p class="text-xl font-bold text-gray-900">{pendingTasks.length}</p>
						<p class="text-gray-500 text-xs">Открытых задач</p>
					</div>
				</div>
			</a>

			<div class="bg-white p-4 rounded-lg border">
				<div class="flex items-center gap-3">
					<div class="w-10 h-10 bg-red-50 rounded-lg flex items-center justify-center">
						<svg class="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
						</svg>
					</div>
					<div>
						<p class="text-xl font-bold text-gray-900">{overdueTasks.length}</p>
						<p class="text-gray-500 text-xs">Просрочено</p>
					</div>
				</div>
			</div>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
			<!-- Recent Meetings -->
			<div class="bg-white rounded-lg border">
				<div class="p-3 border-b flex items-center justify-between">
					<h2 class="font-semibold text-gray-900">Последние встречи</h2>
					<a href="/meetings" class="text-ekf-red text-xs hover:underline flex items-center gap-1">
						Все
						<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					</a>
				</div>
				<div class="divide-y">
					{#each recentMeetings.slice(0, 5) as meeting}
						<a href="/meetings/{meeting.id}" class="block p-3 hover:bg-gray-50 transition-colors">
							<div class="flex justify-between items-start">
								<div class="min-w-0 flex-1">
									<p class="font-medium text-sm text-gray-900 truncate">{meeting.title || meeting.employees?.name || 'Встреча'}</p>
									<p class="text-xs text-gray-500">{formatDate(meeting.date)}</p>
								</div>
								{#if meeting.mood_score}
									<span class="ml-2 px-1.5 py-0.5 rounded text-xs font-medium
										{meeting.mood_score >= 7 ? 'bg-green-100 text-green-700' :
										meeting.mood_score >= 5 ? 'bg-yellow-100 text-yellow-700' : 'bg-red-100 text-red-700'}">
										{meeting.mood_score}
									</span>
								{/if}
							</div>
						</a>
					{/each}
					{#if recentMeetings.length === 0}
						<div class="p-6 text-center">
							<svg class="w-8 h-8 mx-auto text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
							</svg>
							<p class="text-gray-500 text-sm">Нет встреч</p>
							<a href="/calendar" class="text-ekf-red text-xs hover:underline">
								Синхронизировать календарь
							</a>
						</div>
					{/if}
				</div>
			</div>

			<!-- Team -->
			<div class="bg-white rounded-lg border">
				<div class="p-3 border-b flex items-center justify-between">
					<h2 class="font-semibold text-gray-900">Команда</h2>
					<a href="/employees" class="text-ekf-red text-xs hover:underline flex items-center gap-1">
						Все
						<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					</a>
				</div>
				<div class="divide-y">
					{#each (teamStats.length > 0 ? teamStats : $subordinates).slice(0, 6) as member}
						<a href="/employees/{member.id}" class="block p-3 hover:bg-gray-50 transition-colors">
							<div class="flex items-center gap-2 mb-2">
								{#if member.photo_base64}
									<img src="data:image/jpeg;base64,{member.photo_base64}" alt="" class="w-10 h-10 rounded-full object-cover" />
								{:else}
									<div class="w-10 h-10 bg-orange-50 rounded-full flex items-center justify-center">
										<svg class="w-5 h-5 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
										</svg>
									</div>
								{/if}
								<div class="flex-1 min-w-0">
									<p class="font-medium text-sm text-gray-900 truncate">{member.name}</p>
									<p class="text-xs text-gray-500 truncate">{member.position}</p>
								</div>
							</div>
							{#if 'subordinates' in member}
								<div class="flex gap-3 text-xs ml-12">
									<div class="flex items-center gap-1">
										<svg class="w-3.5 h-3.5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
										</svg>
										<span class="text-gray-600">{member.subordinates}</span>
									</div>
									<div class="flex items-center gap-1">
										<svg class="w-3.5 h-3.5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
										</svg>
										<span class="text-gray-600">{member.meetings}</span>
									</div>
									<div class="flex items-center gap-1">
										<svg class="w-3.5 h-3.5 text-yellow-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
										</svg>
										<span class="text-gray-600">{member.open_tasks}</span>
									</div>
									{#if member.overdue_tasks > 0}
										<div class="flex items-center gap-1">
											<svg class="w-3.5 h-3.5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
											</svg>
											<span class="text-red-600 font-medium">{member.overdue_tasks}</span>
										</div>
									{/if}
								</div>
							{/if}
						</a>
					{/each}
					{#if $subordinates.length === 0 && teamStats.length === 0}
						<div class="p-6 text-center">
							<svg class="w-8 h-8 mx-auto text-gray-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
							</svg>
							<p class="text-gray-500 text-sm">Нет подчинённых</p>
						</div>
					{/if}
				</div>
			</div>
		</div>

		<!-- Quick Actions -->
		<div class="bg-white rounded-lg border p-4">
			<h2 class="font-semibold text-gray-900 mb-3">Быстрые действия</h2>
			<div class="grid grid-cols-2 md:grid-cols-4 gap-2">
				<a href="/calendar" class="p-3 border border-gray-200 rounded-lg hover:border-ekf-red hover:bg-orange-50 transition-colors text-center">
					<svg class="w-5 h-5 mx-auto text-ekf-red mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
					</svg>
					<p class="text-xs font-medium text-gray-900">Календарь</p>
				</a>
				<a href="/meetings" class="p-3 border border-gray-200 rounded-lg hover:border-ekf-red hover:bg-orange-50 transition-colors text-center">
					<svg class="w-5 h-5 mx-auto text-ekf-red mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
					</svg>
					<p class="text-xs font-medium text-gray-900">Встречи</p>
				</a>
				<a href="/tasks" class="p-3 border border-gray-200 rounded-lg hover:border-ekf-red hover:bg-orange-50 transition-colors text-center">
					<svg class="w-5 h-5 mx-auto text-ekf-red mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
					</svg>
					<p class="text-xs font-medium text-gray-900">Задачи</p>
				</a>
				<a href="/analytics" class="p-3 border border-gray-200 rounded-lg hover:border-ekf-red hover:bg-orange-50 transition-colors text-center">
					<svg class="w-5 h-5 mx-auto text-ekf-red mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
					</svg>
					<p class="text-xs font-medium text-gray-900">Аналитика</p>
				</a>
			</div>
		</div>
	</div>
{/if}
