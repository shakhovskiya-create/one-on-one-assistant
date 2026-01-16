<script lang="ts">
	import { onMount } from 'svelte';
	import { calendar as calendarApi } from '$lib/api/client';
	import type { CalendarEvent } from '$lib/api/client';
	import { user, subordinates } from '$lib/stores/auth';

	let events: CalendarEvent[] = $state([]);
	let loading = $state(true);
	let syncing = $state(false);

	let currentDate = $state(new Date());
	let viewMode = $state<'week' | 'month'>('week');

	const weekDays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
	const months = [
		'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
		'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
	];

	onMount(async () => {
		if ($user?.id) {
			await loadCalendar();
		}
	});

	async function loadCalendar() {
		if (!$user?.id) return;
		loading = true;
		try {
			const response = await calendarApi.getSimple($user.id);
			// API returns { events: [...] } or just array
			events = Array.isArray(response) ? response : (response.events || []);
		} catch (e) {
			console.error(e);
			events = [];
		} finally {
			loading = false;
		}
	}

	async function syncCalendar() {
		syncing = true;
		try {
			await calendarApi.sync();
			await loadCalendar();
		} catch (e) {
			console.error(e);
		} finally {
			syncing = false;
		}
	}

	function getWeekDates(): Date[] {
		const start = new Date(currentDate);
		const day = start.getDay();
		const diff = start.getDate() - day + (day === 0 ? -6 : 1);
		start.setDate(diff);

		return Array.from({ length: 7 }, (_, i) => {
			const date = new Date(start);
			date.setDate(start.getDate() + i);
			return date;
		});
	}

	function getEventsForDate(date: Date): CalendarEvent[] {
		const dateStr = date.toISOString().split('T')[0];
		return events.filter(e => e.start.startsWith(dateStr));
	}

	function formatTime(isoString: string): string {
		const date = new Date(isoString);
		return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
	}

	function prevWeek() {
		const newDate = new Date(currentDate);
		newDate.setDate(newDate.getDate() - 7);
		currentDate = newDate;
	}

	function nextWeek() {
		const newDate = new Date(currentDate);
		newDate.setDate(newDate.getDate() + 7);
		currentDate = newDate;
	}

	function isToday(date: Date): boolean {
		const today = new Date();
		return date.toDateString() === today.toDateString();
	}

	$effect(() => {
		if ($user?.id) {
			loadCalendar();
		}
	});
</script>

<svelte:head>
	<title>Календарь - EKF Team Hub</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Мой календарь</h1>
			{#if $user}
				<p class="text-gray-500">{$user.name}</p>
			{/if}
		</div>
		<button
			onclick={syncCalendar}
			disabled={syncing}
			class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50 flex items-center gap-2"
		>
			<svg class="w-5 h-5 {syncing ? 'animate-spin' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
			</svg>
			{#if syncing}
				Синхронизация...
			{:else}
				Синхронизировать
			{/if}
		</button>
	</div>

	<!-- Navigation -->
	<div class="bg-white rounded-xl shadow-sm p-4">
		<div class="flex items-center justify-between">
			<button
				onclick={prevWeek}
				class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
			</button>
			<div class="text-lg font-semibold text-gray-900">
				{months[currentDate.getMonth()]} {currentDate.getFullYear()}
			</div>
			<button
				onclick={nextWeek}
				class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
				</svg>
			</button>
		</div>
	</div>

	<!-- Calendar Grid -->
	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="text-gray-500">Загрузка...</div>
		</div>
	{:else}
		<div class="bg-white rounded-xl shadow-sm overflow-hidden">
			<!-- Header -->
			<div class="grid grid-cols-7 border-b">
				{#each getWeekDates() as date, i}
					<div class="p-4 text-center border-r last:border-r-0
						{isToday(date) ? 'bg-ekf-red/5' : ''}">
						<div class="text-sm text-gray-500">{weekDays[i]}</div>
						<div class="text-lg font-semibold mt-1
							{isToday(date) ? 'text-ekf-red' : 'text-gray-900'}">
							{date.getDate()}
						</div>
					</div>
				{/each}
			</div>

			<!-- Events -->
			<div class="grid grid-cols-7 min-h-96">
				{#each getWeekDates() as date}
					<div class="p-2 border-r last:border-r-0 border-b
						{isToday(date) ? 'bg-ekf-red/5' : ''}">
						<div class="space-y-1">
							{#each getEventsForDate(date) as event}
								<div class="p-2 rounded text-xs bg-blue-50 border-l-2 border-blue-500">
									<div class="font-medium text-gray-900 truncate">{event.subject}</div>
									<div class="text-gray-500">
										{formatTime(event.start)} - {formatTime(event.end)}
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Free Slots -->
	<div class="bg-white rounded-xl shadow-sm p-6">
		<h3 class="font-semibold text-gray-900 mb-4">Свободные слоты для встреч</h3>
		<p class="text-gray-500 text-sm">
			Выберите сотрудника для просмотра свободных временных слотов.
		</p>
	</div>
</div>
