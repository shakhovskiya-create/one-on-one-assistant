<script lang="ts">
	import { onMount } from 'svelte';
	import { meetings as meetingsApi } from '$lib/api/client';
	import type { Meeting } from '$lib/api/client';

	let meetings: Meeting[] = $state([]);
	let loading = $state(true);
	let filter = $state('all');

	onMount(async () => {
		try {
			meetings = await meetingsApi.list();
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	const filteredMeetings = $derived(() => {
		if (filter === 'all') return meetings;
		return meetings.filter(m => m.category === filter);
	});

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });
	}
</script>

<svelte:head>
	<title>Встречи - EKF Team Hub</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-gray-900">Встречи</h1>
		<a
			href="/meetings/new"
			class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors"
		>
			Новая встреча
		</a>
	</div>

	<!-- Filters -->
	<div class="flex gap-2">
		{#each [
			{ key: 'all', label: 'Все' },
			{ key: 'one_on_one', label: '1-на-1' },
			{ key: 'project', label: 'Проектные' },
			{ key: 'team', label: 'Командные' }
		] as tab}
			<button
				onclick={() => filter = tab.key}
				class="px-4 py-2 rounded-lg text-sm font-medium transition-colors
					{filter === tab.key
						? 'bg-ekf-red text-white'
						: 'bg-white text-gray-600 hover:bg-gray-100'}"
			>
				{tab.label}
			</button>
		{/each}
	</div>

	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="text-gray-500">Загрузка...</div>
		</div>
	{:else if filteredMeetings().length === 0}
		<div class="bg-white rounded-xl shadow-sm p-12 text-center">
			<div class="text-gray-400 text-lg mb-4">Встреч пока нет</div>
			<a
				href="/meetings/new"
				class="inline-block px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors"
			>
				Провести первую встречу
			</a>
		</div>
	{:else}
		<div class="space-y-4">
			{#each filteredMeetings() as meeting}
				<a href="/meetings/{meeting.id}" class="block bg-white rounded-xl shadow-sm p-6 hover:shadow-md transition-shadow">
					<div class="flex items-start justify-between">
						<div class="flex-1">
							<div class="flex items-center gap-3">
								<h3 class="text-lg font-semibold text-gray-900">
									{meeting.title || 'Без названия'}
								</h3>
								{#if meeting.category}
									<span class="px-2 py-1 text-xs rounded-full bg-gray-100 text-gray-600">
										{meeting.category === 'one_on_one' ? '1-на-1' :
										meeting.category === 'project' ? 'Проект' : 'Команда'}
									</span>
								{/if}
							</div>
							<p class="text-sm text-gray-500 mt-1">{formatDate(meeting.date)}</p>
							{#if meeting.summary}
								<p class="text-gray-600 mt-3 line-clamp-2">{meeting.summary}</p>
							{/if}
						</div>
						<div class="flex flex-col items-end gap-2">
							{#if meeting.mood_score}
								<span class="text-lg font-bold
									{meeting.mood_score >= 7 ? 'text-green-600' :
									meeting.mood_score >= 5 ? 'text-yellow-600' : 'text-red-600'}">
									{meeting.mood_score}/10
								</span>
							{/if}
							{#if meeting.duration_minutes}
								<span class="text-sm text-gray-500">{meeting.duration_minutes} мин</span>
							{/if}
						</div>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>
