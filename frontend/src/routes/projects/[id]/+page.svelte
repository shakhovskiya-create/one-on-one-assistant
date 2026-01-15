<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { projects as projectsApi, meetings as meetingsApi } from '$lib/api/client';
	import type { Project, Meeting } from '$lib/api/client';

	let project: Project | null = $state(null);
	let meetings: Meeting[] = $state([]);
	let loading = $state(true);

	const id = $page.params.id;

	onMount(async () => {
		try {
			[project, meetings] = await Promise.all([
				projectsApi.get(id),
				meetingsApi.list({ project_id: id })
			]);
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>{project?.name || 'Проект'} - EKF Team Hub</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center h-64">
		<div class="text-gray-500">Загрузка...</div>
	</div>
{:else if project}
	<div class="space-y-6">
		<!-- Header -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<div class="flex items-start justify-between">
				<div>
					<h1 class="text-2xl font-bold text-gray-900">{project.name}</h1>
					{#if project.description}
						<p class="text-gray-600 mt-2">{project.description}</p>
					{/if}
				</div>
				<a
					href="/projects/{id}/edit"
					class="px-4 py-2 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
				>
					Редактировать
				</a>
			</div>
		</div>

		<!-- Stats -->
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-2">Встречи</h3>
				<p class="text-3xl font-bold text-ekf-red">{meetings.length}</p>
			</div>
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-2">Участники</h3>
				<p class="text-3xl font-bold text-gray-900">
					{new Set(meetings.map(m => m.employee_id)).size}
				</p>
			</div>
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-2">Последняя встреча</h3>
				{#if meetings.length > 0}
					<p class="text-lg font-medium text-gray-900">{meetings[0].date}</p>
				{:else}
					<p class="text-gray-400">Нет встреч</p>
				{/if}
			</div>
		</div>

		<!-- Meetings List -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<div class="flex items-center justify-between mb-4">
				<h3 class="font-semibold text-gray-900">Встречи по проекту</h3>
				<a
					href="/meetings/new?project_id={id}"
					class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors text-sm"
				>
					Новая встреча
				</a>
			</div>
			{#if meetings.length === 0}
				<p class="text-gray-400 text-center py-8">Встреч пока нет</p>
			{:else}
				<div class="space-y-3">
					{#each meetings as meeting}
						<a href="/meetings/{meeting.id}" class="block p-4 rounded-lg hover:bg-gray-50 border border-gray-100">
							<div class="flex justify-between items-center">
								<div>
									<p class="font-medium text-gray-900">{meeting.title || 'Без названия'}</p>
									<p class="text-sm text-gray-500">{meeting.date}</p>
								</div>
								{#if meeting.mood_score}
									<span class="text-sm font-medium px-2 py-1 rounded-full
										{meeting.mood_score >= 7 ? 'bg-green-100 text-green-700' :
										meeting.mood_score >= 5 ? 'bg-yellow-100 text-yellow-700' : 'bg-red-100 text-red-700'}">
										{meeting.mood_score}/10
									</span>
								{/if}
							</div>
							{#if meeting.summary}
								<p class="text-sm text-gray-600 mt-2 line-clamp-2">{meeting.summary}</p>
							{/if}
						</a>
					{/each}
				</div>
			{/if}
		</div>
	</div>
{:else}
	<div class="text-center py-12">
		<div class="text-gray-400 text-lg">Проект не найден</div>
	</div>
{/if}
