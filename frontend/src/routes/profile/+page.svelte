<script lang="ts">
	import { user, subordinates } from '$lib/stores/auth';
	import { api } from '$lib/api/client';

	let meetings: any[] = $state([]);
	let tasks: any[] = $state([]);
	let loading = $state(true);

	$effect(() => {
		loadUserData();
	});

	async function loadUserData() {
		loading = true;
		try {
			const [meetingsRes, tasksRes] = await Promise.all([
				api.meetings.list().catch(() => []),
				api.tasks.list().catch(() => [])
			]);
			meetings = meetingsRes.slice(0, 5);
			tasks = tasksRes.filter((t: any) => t.assignee_id === $user?.id).slice(0, 5);
		} catch (e) {
			console.error('Error loading user data:', e);
		}
		loading = false;
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString('ru-RU', {
			day: 'numeric',
			month: 'short',
			year: 'numeric'
		});
	}
</script>

<svelte:head>
	<title>Мой профиль - EKF Hub</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
	<!-- Header -->
	<div class="bg-white rounded-xl shadow-sm overflow-hidden mb-6">
		<div class="h-32 bg-gradient-to-r from-ekf-red to-red-600"></div>
		<div class="px-6 pb-6">
			<div class="flex items-end -mt-16 mb-4">
				{#if $user?.photo_base64}
					<img
						src="data:image/jpeg;base64,{$user.photo_base64}"
						alt=""
						class="w-32 h-32 rounded-full border-4 border-white object-cover shadow-lg"
					/>
				{:else}
					<div class="w-32 h-32 rounded-full border-4 border-white bg-ekf-red text-white flex items-center justify-center text-4xl font-bold shadow-lg">
						{$user?.name?.charAt(0) || '?'}
					</div>
				{/if}
				<div class="ml-6 mb-2">
					<h1 class="text-2xl font-bold text-gray-900">{$user?.name || 'Пользователь'}</h1>
					<p class="text-gray-600">{$user?.position || ''}</p>
				</div>
			</div>

			<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
				{#if $user?.email}
					<div class="flex items-center gap-2 text-gray-600">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
						</svg>
						<span>{$user.email}</span>
					</div>
				{/if}
				{#if $user?.department}
					<div class="flex items-center gap-2 text-gray-600">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
						</svg>
						<span>{$user.department}</span>
					</div>
				{/if}
				{#if $subordinates.length > 0}
					<div class="flex items-center gap-2 text-gray-600">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
						</svg>
						<span>{$subordinates.length} подчинённых</span>
					</div>
				{/if}
			</div>
		</div>
	</div>

	<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
		<!-- Subordinates -->
		{#if $subordinates.length > 0}
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h2 class="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
					<svg class="w-5 h-5 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
					</svg>
					Мои подчинённые
				</h2>
				<div class="space-y-3 max-h-64 overflow-y-auto">
					{#each $subordinates as sub}
						<a href="/employees/{sub.id}" class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 transition-colors">
							{#if sub.photo_base64}
								<img src="data:image/jpeg;base64,{sub.photo_base64}" alt="" class="w-10 h-10 rounded-full object-cover" />
							{:else}
								<div class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center text-gray-600 font-medium">
									{sub.name.charAt(0)}
								</div>
							{/if}
							<div class="flex-1 min-w-0">
								<div class="font-medium text-gray-900 truncate">{sub.name}</div>
								<div class="text-sm text-gray-500 truncate">{sub.position || ''}</div>
							</div>
						</a>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Recent Meetings -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
				<svg class="w-5 h-5 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
				</svg>
				Последние встречи
			</h2>
			{#if loading}
				<div class="text-center py-8 text-gray-500">Загрузка...</div>
			{:else if meetings.length === 0}
				<div class="text-center py-8 text-gray-500">Нет встреч</div>
			{:else}
				<div class="space-y-3">
					{#each meetings as meeting}
						<a href="/meetings/{meeting.id}" class="block p-3 rounded-lg border border-gray-100 hover:border-ekf-red transition-colors">
							<div class="font-medium text-gray-900">{meeting.title || 'Встреча'}</div>
							<div class="text-sm text-gray-500">{formatDate(meeting.date || meeting.created_at)}</div>
						</a>
					{/each}
				</div>
				<a href="/meetings" class="block text-center text-ekf-red hover:underline mt-4 text-sm">
					Все встречи
				</a>
			{/if}
		</div>

		<!-- My Tasks -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
				<svg class="w-5 h-5 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
				</svg>
				Мои задачи
			</h2>
			{#if loading}
				<div class="text-center py-8 text-gray-500">Загрузка...</div>
			{:else if tasks.length === 0}
				<div class="text-center py-8 text-gray-500">Нет задач</div>
			{:else}
				<div class="space-y-3">
					{#each tasks as task}
						<div class="p-3 rounded-lg border border-gray-100">
							<div class="font-medium text-gray-900">{task.title}</div>
							<div class="text-sm text-gray-500 flex items-center gap-2 mt-1">
								<span class="px-2 py-0.5 rounded text-xs
									{task.status === 'done' ? 'bg-green-100 text-green-700' : ''}
									{task.status === 'in_progress' ? 'bg-blue-100 text-blue-700' : ''}
									{task.status === 'todo' ? 'bg-gray-100 text-gray-700' : ''}
									{task.status === 'backlog' ? 'bg-gray-100 text-gray-600' : ''}
									{task.status === 'review' ? 'bg-yellow-100 text-yellow-700' : ''}
								">
									{task.status === 'done' ? 'Готово' : ''}
									{task.status === 'in_progress' ? 'В работе' : ''}
									{task.status === 'todo' ? 'К выполнению' : ''}
									{task.status === 'backlog' ? 'Backlog' : ''}
									{task.status === 'review' ? 'На проверке' : ''}
								</span>
							</div>
						</div>
					{/each}
				</div>
				<a href="/tasks" class="block text-center text-ekf-red hover:underline mt-4 text-sm">
					Все задачи
				</a>
			{/if}
		</div>

		<!-- Quick Actions -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Быстрые действия</h2>
			<div class="space-y-2">
				<a href="/calendar" class="flex items-center gap-3 p-3 rounded-lg border border-gray-100 hover:border-ekf-red transition-colors">
					<svg class="w-6 h-6 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
					</svg>
					<span>Мой календарь</span>
				</a>
				<a href="/meetings" class="flex items-center gap-3 p-3 rounded-lg border border-gray-100 hover:border-ekf-red transition-colors">
					<svg class="w-6 h-6 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
					</svg>
					<span>Создать встречу</span>
				</a>
				<a href="/analytics" class="flex items-center gap-3 p-3 rounded-lg border border-gray-100 hover:border-ekf-red transition-colors">
					<svg class="w-6 h-6 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
					</svg>
					<span>Моя аналитика</span>
				</a>
			</div>
		</div>
	</div>
</div>
