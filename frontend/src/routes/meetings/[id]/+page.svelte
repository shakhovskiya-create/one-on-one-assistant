<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { meetings as meetingsApi } from '$lib/api/client';
	import type { Meeting } from '$lib/api/client';

	let meeting: Meeting | null = $state(null);
	let loading = $state(true);

	const id = $page.params.id;

	onMount(async () => {
		try {
			meeting = await meetingsApi.get(id);
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleDateString('ru-RU', {
			day: 'numeric',
			month: 'long',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatActionItem(item: unknown): string {
		if (typeof item === 'string') return item;
		if (!item || typeof item !== 'object') return '';
		const data = item as { task?: string; improvement?: string; responsible?: string; deadline?: string | null };
		const label = data.task || data.improvement || '';
		const parts = [label];
		if (data.responsible) parts.push(`Ответственный: ${data.responsible}`);
		if (data.deadline) parts.push(`Срок: ${data.deadline}`);
		return parts.filter(Boolean).join(' · ');
	}

	function formatAgreement(item: unknown): string {
		if (typeof item === 'string') return item;
		if (!item || typeof item !== 'object') return '';
		const data = item as { task?: string; responsible?: string; deadline?: string | null };
		const label = data.task || '';
		const parts = [label];
		if (data.responsible) parts.push(`Ответственный: ${data.responsible}`);
		if (data.deadline) parts.push(`Срок: ${data.deadline}`);
		return parts.filter(Boolean).join(' · ');
	}
</script>

<svelte:head>
	<title>{meeting?.title || 'Встреча'} - EKF Team Hub</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center h-64">
		<div class="text-gray-500">Загрузка...</div>
	</div>
{:else if meeting}
	<div class="space-y-6">
		<!-- Header -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<div class="flex items-start justify-between">
				<div>
					<div class="flex items-center gap-3">
						<h1 class="text-2xl font-bold text-gray-900">{meeting.title || 'Без названия'}</h1>
						{#if meeting.meeting_categories?.code}
							<span class="px-3 py-1 text-sm rounded-full bg-gray-100 text-gray-600">
								{meeting.meeting_categories.code === 'one_on_one' ? '1-на-1' :
								meeting.meeting_categories.code === 'project' ? 'Проект' : 'Команда'}
							</span>
						{/if}
					</div>
					<p class="text-gray-500 mt-2">{formatDate(meeting.date)}</p>
					{#if meeting.duration_minutes}
						<p class="text-sm text-gray-400">Длительность: {meeting.duration_minutes} мин</p>
					{/if}
				</div>
				{#if meeting.mood_score}
					<div class="text-center">
						<div class="text-3xl font-bold
							{meeting.mood_score >= 7 ? 'text-green-600' :
							meeting.mood_score >= 5 ? 'text-yellow-600' : 'text-red-600'}">
							{meeting.mood_score}/10
						</div>
						<div class="text-sm text-gray-500">Настроение</div>
					</div>
				{/if}
			</div>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Main Content -->
			<div class="lg:col-span-2 space-y-6">
				<!-- Summary -->
				{#if meeting.summary}
					<div class="bg-white rounded-xl shadow-sm p-6">
						<h3 class="font-semibold text-gray-900 mb-3">Краткое содержание</h3>
						<p class="text-gray-600 whitespace-pre-wrap">{meeting.summary}</p>
					</div>
				{/if}

				<!-- Analysis -->
				{#if meeting.analysis}
					<div class="bg-white rounded-xl shadow-sm p-6">
						<h3 class="font-semibold text-gray-900 mb-4">Анализ встречи</h3>

						{#if meeting.analysis.key_topics}
							<div class="mb-4">
								<h4 class="text-sm font-medium text-gray-700 mb-2">Ключевые темы</h4>
								<div class="flex flex-wrap gap-2">
									{#each meeting.analysis.key_topics as topic}
										<span class="px-3 py-1 bg-blue-50 text-blue-700 rounded-full text-sm">{topic}</span>
									{/each}
								</div>
							</div>
						{/if}

						{#if meeting.analysis.action_items}
							<div class="mb-4">
								<h4 class="text-sm font-medium text-gray-700 mb-2">Задачи к выполнению</h4>
								<ul class="space-y-2">
										{#each meeting.analysis.action_items as item}
										<li class="flex items-start gap-2">
											<span class="text-ekf-red">•</span>
											<span class="text-gray-600">{formatActionItem(item)}</span>
										</li>
									{/each}
								</ul>
							</div>
						{/if}

						{#if meeting.analysis.agreements}
							<div class="mb-4">
								<h4 class="text-sm font-medium text-gray-700 mb-2">Договорённости</h4>
								<ul class="space-y-2">
										{#each meeting.analysis.agreements as agreement}
										<li class="flex items-start gap-2">
											<span class="text-green-600">✓</span>
											<span class="text-gray-600">{formatAgreement(agreement)}</span>
										</li>
									{/each}
								</ul>
							</div>
						{/if}

						{#if meeting.analysis.red_flags}
							<div>
								<h4 class="text-sm font-medium text-gray-700 mb-2">Красные флаги</h4>
								<div class="p-3 bg-red-50 rounded-lg">
									{#if meeting.analysis.red_flags.burnout_signs}
										<p class="text-red-700 text-sm mb-1">
											<strong>Признаки выгорания:</strong> {meeting.analysis.red_flags.burnout_signs}
										</p>
									{/if}
									{#if meeting.analysis.red_flags.turnover_risk && meeting.analysis.red_flags.turnover_risk !== 'low'}
										<p class="text-red-700 text-sm">
											<strong>Риск увольнения:</strong> {meeting.analysis.red_flags.turnover_risk}
										</p>
									{/if}
								</div>
							</div>
						{/if}
					</div>
				{/if}

				<!-- Transcript -->
				{#if meeting.transcript}
					<div class="bg-white rounded-xl shadow-sm p-6">
						<h3 class="font-semibold text-gray-900 mb-3">Транскрипт</h3>
						<div class="text-gray-600 whitespace-pre-wrap text-sm max-h-96 overflow-y-auto">
							{meeting.transcript}
						</div>
					</div>
				{/if}
			</div>

			<!-- Sidebar -->
			<div class="space-y-6">
				<!-- Participants -->
				<div class="bg-white rounded-xl shadow-sm p-6">
					<h3 class="font-semibold text-gray-900 mb-3">Участники</h3>
					{#if meeting.participants && meeting.participants.length > 0}
						<div class="space-y-2">
							{#each meeting.participants as participant}
								<a href="/employees/{participant.id}" class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50">
									<div class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center text-gray-600 font-medium">
										{participant.name?.charAt(0) || 'С'}
									</div>
									<span class="text-gray-900">{participant.name}</span>
								</a>
							{/each}
						</div>
					{:else if meeting.employee_id}
						<a href="/employees/{meeting.employee_id}" class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50">
							<div class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center text-gray-600 font-medium">
								С
							</div>
							<span class="text-gray-900">Сотрудник</span>
						</a>
					{:else}
						<p class="text-sm text-gray-500">Нет участников</p>
					{/if}
				</div>

				<!-- Related Project -->
				{#if meeting.project_id}
					<div class="bg-white rounded-xl shadow-sm p-6">
						<h3 class="font-semibold text-gray-900 mb-3">Проект</h3>
						<a href="/projects/{meeting.project_id}" class="text-ekf-red hover:underline">
							Перейти к проекту
						</a>
					</div>
				{/if}

				<!-- Actions -->
				<div class="bg-white rounded-xl shadow-sm p-6">
					<h3 class="font-semibold text-gray-900 mb-3">Действия</h3>
					<div class="space-y-2">
						<button class="w-full px-4 py-2 text-left text-gray-700 hover:bg-gray-50 rounded-lg transition-colors">
							Экспортировать в PDF
						</button>
						<button class="w-full px-4 py-2 text-left text-gray-700 hover:bg-gray-50 rounded-lg transition-colors">
							Отправить по email
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
{:else}
	<div class="text-center py-12">
		<div class="text-gray-400 text-lg">Встреча не найдена</div>
	</div>
{/if}
