<script lang="ts">
	import { onMount } from 'svelte';
	import { calendar as calendarApi, meetings as meetingsApi } from '$lib/api/client';
	import type { CalendarEvent } from '$lib/api/client';
	import { user, subordinates } from '$lib/stores/auth';

	let events: CalendarEvent[] = $state([]);
	let loading = $state(true);
	let syncing = $state(false);
	let showSyncDialog = $state(false);
	let showEventDialog = $state(false);

	let currentDate = $state(new Date());
	let viewMode = $state<'week' | 'month'>('month');

	// Exchange credentials (stored in localStorage)
	let exchangeUser = $state('');
	let exchangePass = $state('');

	// New event form
	let newEvent = $state({
		title: '',
		date: '',
		start_time: '',
		end_time: '',
		location: '',
		employee_id: ''
	});

	const weekDays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
	const months = [
		'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
		'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
	];

	onMount(async () => {
		// Load saved credentials
		exchangeUser = localStorage.getItem('exchange_user') || '';
		if ($user?.id) {
			await loadCalendar();
		}
	});

	async function loadCalendar() {
		if (!$user?.id) return;
		loading = true;
		try {
			const response = await calendarApi.getSimple($user.id);
			events = Array.isArray(response) ? response : (response.events || []);
		} catch (e) {
			console.error(e);
			events = [];
		} finally {
			loading = false;
		}
	}

	function openSyncDialog() {
		showSyncDialog = true;
	}

	async function syncCalendar() {
		if (!$user?.id || !exchangeUser || !exchangePass) return;

		syncing = true;
		try {
			localStorage.setItem('exchange_user', exchangeUser);
			await calendarApi.sync({
				employee_id: $user.id,
				username: exchangeUser,
				password: exchangePass,
				days_back: 30,
				days_forward: 60
			});
			await loadCalendar();
			showSyncDialog = false;
		} catch (e: any) {
			console.error(e);
			alert('Ошибка синхронизации: ' + (e.message || 'Проверьте учётные данные'));
		} finally {
			syncing = false;
		}
	}

	function openNewEventDialog(date?: Date) {
		const d = date || new Date();
		newEvent = {
			title: '',
			date: d.toISOString().split('T')[0],
			start_time: '09:00',
			end_time: '10:00',
			location: '',
			employee_id: $user?.id || ''
		};
		showEventDialog = true;
	}

	async function createEvent() {
		if (!newEvent.title || !newEvent.date) return;

		try {
			await meetingsApi.create({
				title: newEvent.title,
				date: newEvent.date,
				start_time: `${newEvent.date}T${newEvent.start_time}:00`,
				end_time: `${newEvent.date}T${newEvent.end_time}:00`,
				location: newEvent.location || undefined,
				employee_id: newEvent.employee_id || undefined
			});
			await loadCalendar();
			showEventDialog = false;
		} catch (e) {
			console.error(e);
			alert('Ошибка создания события');
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

	function getMonthDates(): Date[][] {
		const year = currentDate.getFullYear();
		const month = currentDate.getMonth();
		const firstDay = new Date(year, month, 1);
		const lastDay = new Date(year, month + 1, 0);

		const startDay = firstDay.getDay() || 7;
		const weeks: Date[][] = [];
		let currentWeek: Date[] = [];

		// Fill in days from previous month
		for (let i = startDay - 1; i > 0; i--) {
			const d = new Date(year, month, 1 - i);
			currentWeek.push(d);
		}

		// Fill in days of current month
		for (let day = 1; day <= lastDay.getDate(); day++) {
			currentWeek.push(new Date(year, month, day));
			if (currentWeek.length === 7) {
				weeks.push(currentWeek);
				currentWeek = [];
			}
		}

		// Fill remaining days from next month
		if (currentWeek.length > 0) {
			let nextDay = 1;
			while (currentWeek.length < 7) {
				currentWeek.push(new Date(year, month + 1, nextDay++));
			}
			weeks.push(currentWeek);
		}

		return weeks;
	}

	function getEventsForDate(date: Date): CalendarEvent[] {
		const dateStr = date.toISOString().split('T')[0];
		return events.filter(e => {
			const eventDate = e.start?.split('T')[0] || e.date;
			return eventDate === dateStr;
		});
	}

	function formatTime(isoString: string): string {
		if (!isoString) return '';
		const date = new Date(isoString);
		return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
	}

	function prevPeriod() {
		const newDate = new Date(currentDate);
		if (viewMode === 'week') {
			newDate.setDate(newDate.getDate() - 7);
		} else {
			newDate.setMonth(newDate.getMonth() - 1);
		}
		currentDate = newDate;
	}

	function nextPeriod() {
		const newDate = new Date(currentDate);
		if (viewMode === 'week') {
			newDate.setDate(newDate.getDate() + 7);
		} else {
			newDate.setMonth(newDate.getMonth() + 1);
		}
		currentDate = newDate;
	}

	function isToday(date: Date): boolean {
		const today = new Date();
		return date.toDateString() === today.toDateString();
	}

	function isCurrentMonth(date: Date): boolean {
		return date.getMonth() === currentDate.getMonth();
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
		<div class="flex items-center gap-3">
			<button
				onclick={() => openNewEventDialog()}
				class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors flex items-center gap-2"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Новая встреча
			</button>
			<button
				onclick={openSyncDialog}
				class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors flex items-center gap-2"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				Синхронизировать
			</button>
		</div>
	</div>

	<!-- Navigation -->
	<div class="bg-white rounded-xl shadow-sm p-4">
		<div class="flex items-center justify-between">
			<button
				onclick={prevPeriod}
				class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
			</button>
			<div class="flex items-center gap-4">
				<div class="text-lg font-semibold text-gray-900">
					{months[currentDate.getMonth()]} {currentDate.getFullYear()}
				</div>
				<div class="flex rounded-lg border border-gray-200 overflow-hidden">
					<button
						onclick={() => viewMode = 'week'}
						class="px-3 py-1 text-sm {viewMode === 'week' ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}"
					>
						Неделя
					</button>
					<button
						onclick={() => viewMode = 'month'}
						class="px-3 py-1 text-sm {viewMode === 'month' ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}"
					>
						Месяц
					</button>
				</div>
			</div>
			<button
				onclick={nextPeriod}
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
	{:else if viewMode === 'week'}
		<div class="bg-white rounded-xl shadow-sm overflow-hidden">
			<div class="grid grid-cols-7 border-b">
				{#each getWeekDates() as date, i}
					<div class="p-4 text-center border-r last:border-r-0 {isToday(date) ? 'bg-ekf-red/5' : ''}">
						<div class="text-sm text-gray-500">{weekDays[i]}</div>
						<div class="text-lg font-semibold mt-1 {isToday(date) ? 'text-ekf-red' : 'text-gray-900'}">
							{date.getDate()}
						</div>
					</div>
				{/each}
			</div>
			<div class="grid grid-cols-7 min-h-96">
				{#each getWeekDates() as date}
					<button
						onclick={() => openNewEventDialog(date)}
						class="p-2 border-r last:border-r-0 border-b text-left hover:bg-gray-50 {isToday(date) ? 'bg-ekf-red/5' : ''}"
					>
						<div class="space-y-1">
							{#each getEventsForDate(date) as event}
								<div class="p-2 rounded text-xs bg-blue-50 border-l-2 border-blue-500">
									<div class="font-medium text-gray-900 truncate">{event.subject || event.title}</div>
									<div class="text-gray-500">
										{formatTime(event.start || event.start_time)} - {formatTime(event.end || event.end_time)}
									</div>
								</div>
							{/each}
						</div>
					</button>
				{/each}
			</div>
		</div>
	{:else}
		<div class="bg-white rounded-xl shadow-sm overflow-hidden">
			<div class="grid grid-cols-7 border-b bg-gray-50">
				{#each weekDays as day}
					<div class="p-3 text-center text-sm font-medium text-gray-600 border-r last:border-r-0">
						{day}
					</div>
				{/each}
			</div>
			{#each getMonthDates() as week}
				<div class="grid grid-cols-7 border-b last:border-b-0">
					{#each week as date}
						<button
							onclick={() => openNewEventDialog(date)}
							class="p-2 min-h-24 border-r last:border-r-0 text-left hover:bg-gray-50 transition-colors
								{isToday(date) ? 'bg-ekf-red/5' : ''}
								{!isCurrentMonth(date) ? 'bg-gray-50 text-gray-400' : ''}"
						>
							<div class="text-sm font-medium mb-1 {isToday(date) ? 'text-ekf-red' : ''}">
								{date.getDate()}
							</div>
							<div class="space-y-1">
								{#each getEventsForDate(date).slice(0, 3) as event}
									<div class="text-xs p-1 rounded bg-blue-100 text-blue-800 truncate">
										{event.subject || event.title}
									</div>
								{/each}
								{#if getEventsForDate(date).length > 3}
									<div class="text-xs text-gray-500">+{getEventsForDate(date).length - 3} ещё</div>
								{/if}
							</div>
						</button>
					{/each}
				</div>
			{/each}
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

<!-- Sync Dialog -->
{#if showSyncDialog}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
		<div class="bg-white rounded-xl p-6 w-full max-w-md">
			<h3 class="text-lg font-semibold text-gray-900 mb-4">Синхронизация с Exchange</h3>
			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Логин Exchange</label>
					<input
						type="text"
						bind:value={exchangeUser}
						placeholder="domain\username или email"
						class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Пароль</label>
					<input
						type="password"
						bind:value={exchangePass}
						class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
					/>
				</div>
			</div>
			<div class="mt-6 flex justify-end gap-3">
				<button
					onclick={() => showSyncDialog = false}
					class="px-4 py-2 border border-gray-200 rounded-lg hover:bg-gray-50"
				>
					Отмена
				</button>
				<button
					onclick={syncCalendar}
					disabled={!exchangeUser || !exchangePass || syncing}
					class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 disabled:opacity-50"
				>
					{syncing ? 'Синхронизация...' : 'Синхронизировать'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- New Event Dialog -->
{#if showEventDialog}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
		<div class="bg-white rounded-xl p-6 w-full max-w-md">
			<h3 class="text-lg font-semibold text-gray-900 mb-4">Новая встреча</h3>
			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Название</label>
					<input
						type="text"
						bind:value={newEvent.title}
						placeholder="Тема встречи"
						class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Дата</label>
					<input
						type="date"
						bind:value={newEvent.date}
						class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
					/>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Начало</label>
						<input
							type="time"
							bind:value={newEvent.start_time}
							class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Конец</label>
						<input
							type="time"
							bind:value={newEvent.end_time}
							class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
						/>
					</div>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Место</label>
					<input
						type="text"
						bind:value={newEvent.location}
						placeholder="Переговорная, онлайн..."
						class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red/20"
					/>
				</div>
			</div>
			<div class="mt-6 flex justify-end gap-3">
				<button
					onclick={() => showEventDialog = false}
					class="px-4 py-2 border border-gray-200 rounded-lg hover:bg-gray-50"
				>
					Отмена
				</button>
				<button
					onclick={createEvent}
					disabled={!newEvent.title || !newEvent.date}
					class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 disabled:opacity-50"
				>
					Создать
				</button>
			</div>
		</div>
	</div>
{/if}
