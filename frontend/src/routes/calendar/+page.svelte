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
	let selectedEvent: CalendarEvent | null = $state(null);

	let currentDate = $state(new Date());
	let viewMode = $state<'day' | 'week' | 'month'>('week');

	// New event form
	let newEvent = $state({
		title: '',
		date: '',
		start_time: '',
		end_time: '',
		location: '',
		employee_id: '',
		room: '',
		is_online: false,
		online_service: 'teams',
		participants: [] as string[]
	});

	// Meeting rooms (could be loaded from backend)
	const meetingRooms = [
		{ id: 'conf-1', name: 'Переговорная 1', floor: '2 этаж', capacity: 6 },
		{ id: 'conf-2', name: 'Переговорная 2', floor: '2 этаж', capacity: 10 },
		{ id: 'conf-3', name: 'Большой зал', floor: '3 этаж', capacity: 20 },
		{ id: 'conf-4', name: 'Малая переговорная', floor: '1 этаж', capacity: 4 }
	];

	// Online services
	const onlineServices = [
		{ id: 'teams', name: 'Microsoft Teams', icon: '📱' },
		{ id: 'zoom', name: 'Zoom', icon: '🎥' },
		{ id: 'meet', name: 'Google Meet', icon: '📹' }
	];

	// Participant search
	let participantSearch = $state('');
	let showParticipantDropdown = $state(false);

	const weekDays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
	const weekDaysFull = ['Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота', 'Воскресенье'];
	const months = [
		'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
		'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
	];
	const monthsShort = [
		'янв', 'фев', 'мар', 'апр', 'май', 'июн',
		'июл', 'авг', 'сен', 'окт', 'ноя', 'дек'
	];

	// Hours for time grid (7:00 - 21:00)
	const hours = Array.from({ length: 15 }, (_, i) => i + 7);

	onMount(async () => {
		if ($user?.id) {
			await loadCalendar();
		}
	});

	async function loadCalendar() {
		if (!$user?.id) return;
		loading = true;
		try {
			const response = await calendarApi.get($user.id);
			events = Array.isArray(response) ? response : (response.events || []);
		} catch (e: any) {
			console.error('Failed to load calendar from Exchange:', e);
			try {
				const fallbackResponse = await calendarApi.getSimple($user.id);
				events = Array.isArray(fallbackResponse) ? fallbackResponse : (fallbackResponse.events || []);
			} catch (fallbackError) {
				console.error('Failed to load calendar from database:', fallbackError);
				events = [];
			}
		} finally {
			loading = false;
		}
	}

	async function syncCalendar() {
		if (!$user?.id) return;
		syncing = true;
		try {
			await calendarApi.sync({
				employee_id: $user.id,
				days_back: 30,
				days_forward: 60
			});
			await loadCalendar();
			showSyncDialog = false;
		} catch (e: any) {
			console.error(e);
			alert('Ошибка синхронизации: ' + (e.message || 'Unknown error'));
		} finally {
			syncing = false;
		}
	}

	function openNewEventDialog(date?: Date, hour?: number) {
		const d = date || new Date();
		const startHour = hour ?? 9;
		newEvent = {
			title: '',
			date: d.toISOString().split('T')[0],
			start_time: `${startHour.toString().padStart(2, '0')}:00`,
			end_time: `${(startHour + 1).toString().padStart(2, '0')}:00`,
			location: '',
			employee_id: $user?.id || '',
			room: '',
			is_online: false,
			online_service: 'teams',
			participants: []
		};
		participantSearch = '';
		showEventDialog = true;
	}

	function addParticipant(employeeId: string) {
		if (!newEvent.participants.includes(employeeId)) {
			newEvent.participants = [...newEvent.participants, employeeId];
		}
		participantSearch = '';
		showParticipantDropdown = false;
	}

	function removeParticipant(employeeId: string) {
		newEvent.participants = newEvent.participants.filter(id => id !== employeeId);
	}

	function getFilteredEmployees() {
		if (!participantSearch) return [];
		const search = participantSearch.toLowerCase();
		return $subordinates
			.filter(e =>
				e.id !== $user?.id &&
				!newEvent.participants.includes(e.id) &&
				(e.name?.toLowerCase().includes(search) || e.email?.toLowerCase().includes(search))
			)
			.slice(0, 5);
	}

	function getEmployeeById(id: string) {
		return $subordinates.find(e => e.id === id) || { name: 'Неизвестный', id };
	}

	async function createEvent() {
		if (!newEvent.title || !newEvent.date) return;
		try {
			// Build location string
			let location = '';
			if (newEvent.is_online) {
				const service = onlineServices.find(s => s.id === newEvent.online_service);
				location = service?.name || 'Онлайн';
			} else if (newEvent.room) {
				const room = meetingRooms.find(r => r.id === newEvent.room);
				location = room ? `${room.name} (${room.floor})` : newEvent.location;
			} else {
				location = newEvent.location;
			}

			await meetingsApi.create({
				title: newEvent.title,
				date: newEvent.date,
				start_time: `${newEvent.date}T${newEvent.start_time}:00`,
				end_time: `${newEvent.date}T${newEvent.end_time}:00`,
				location: location || undefined,
				employee_id: newEvent.employee_id || undefined,
				participant_ids: newEvent.participants.length > 0 ? newEvent.participants : undefined
			});
			await loadCalendar();
			showEventDialog = false;
		} catch (e) {
			console.error(e);
			alert('Ошибка создания события');
		}
	}

	function goToToday() {
		currentDate = new Date();
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

		for (let i = startDay - 1; i > 0; i--) {
			const d = new Date(year, month, 1 - i);
			currentWeek.push(d);
		}

		for (let day = 1; day <= lastDay.getDate(); day++) {
			currentWeek.push(new Date(year, month, day));
			if (currentWeek.length === 7) {
				weeks.push(currentWeek);
				currentWeek = [];
			}
		}

		if (currentWeek.length > 0) {
			let nextDay = 1;
			while (currentWeek.length < 7) {
				currentWeek.push(new Date(year, month + 1, nextDay++));
			}
			weeks.push(currentWeek);
		}

		return weeks;
	}

	// Mini calendar for sidebar
	function getMiniCalendarDates(): Date[][] {
		const year = currentDate.getFullYear();
		const month = currentDate.getMonth();
		const firstDay = new Date(year, month, 1);
		const lastDay = new Date(year, month + 1, 0);
		const startDay = firstDay.getDay() || 7;
		const weeks: Date[][] = [];
		let currentWeek: Date[] = [];

		for (let i = startDay - 1; i > 0; i--) {
			currentWeek.push(new Date(year, month, 1 - i));
		}

		for (let day = 1; day <= lastDay.getDate(); day++) {
			currentWeek.push(new Date(year, month, day));
			if (currentWeek.length === 7) {
				weeks.push(currentWeek);
				currentWeek = [];
			}
		}

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
		}).sort((a, b) => {
			const aTime = a.start || a.start_time || '';
			const bTime = b.start || b.start_time || '';
			return aTime.localeCompare(bTime);
		});
	}

	function getEventsForHour(date: Date, hour: number): CalendarEvent[] {
		const dateStr = date.toISOString().split('T')[0];
		return events.filter(e => {
			const eventDate = e.start?.split('T')[0] || e.date;
			if (eventDate !== dateStr) return false;
			const startTime = e.start || e.start_time || '';
			const eventHour = parseInt(startTime.split('T')[1]?.split(':')[0] || '0');
			return eventHour === hour;
		});
	}

	function getEventPosition(event: CalendarEvent): { top: number; height: number } {
		const startTime = event.start || event.start_time || '';
		const endTime = event.end || event.end_time || '';

		const startParts = startTime.split('T')[1]?.split(':') || ['9', '0'];
		const endParts = endTime.split('T')[1]?.split(':') || ['10', '0'];

		const startHour = parseInt(startParts[0]);
		const startMin = parseInt(startParts[1] || '0');
		const endHour = parseInt(endParts[0]);
		const endMin = parseInt(endParts[1] || '0');

		const startOffset = (startHour - 7) * 60 + startMin;
		const duration = (endHour - startHour) * 60 + (endMin - startMin);

		return {
			top: (startOffset / 60) * 48, // 48px per hour
			height: Math.max((duration / 60) * 48, 24) // minimum 24px
		};
	}

	function formatTime(isoString: string): string {
		if (!isoString) return '';
		const date = new Date(isoString);
		return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
	}

	function formatEventTime(event: CalendarEvent): string {
		const start = formatTime(event.start || event.start_time || '');
		const end = formatTime(event.end || event.end_time || '');
		return start && end ? `${start} - ${end}` : start || '';
	}

	function prevPeriod() {
		const newDate = new Date(currentDate);
		if (viewMode === 'day') {
			newDate.setDate(newDate.getDate() - 1);
		} else if (viewMode === 'week') {
			newDate.setDate(newDate.getDate() - 7);
		} else {
			newDate.setMonth(newDate.getMonth() - 1);
		}
		currentDate = newDate;
	}

	function nextPeriod() {
		const newDate = new Date(currentDate);
		if (viewMode === 'day') {
			newDate.setDate(newDate.getDate() + 1);
		} else if (viewMode === 'week') {
			newDate.setDate(newDate.getDate() + 7);
		} else {
			newDate.setMonth(newDate.getMonth() + 1);
		}
		currentDate = newDate;
	}

	function prevMonth() {
		const newDate = new Date(currentDate);
		newDate.setMonth(newDate.getMonth() - 1);
		currentDate = newDate;
	}

	function nextMonth() {
		const newDate = new Date(currentDate);
		newDate.setMonth(newDate.getMonth() + 1);
		currentDate = newDate;
	}

	function isToday(date: Date): boolean {
		const today = new Date();
		return date.toDateString() === today.toDateString();
	}

	function isCurrentMonth(date: Date): boolean {
		return date.getMonth() === currentDate.getMonth();
	}

	function isSelected(date: Date): boolean {
		return date.toDateString() === currentDate.toDateString();
	}

	function selectDate(date: Date) {
		currentDate = date;
		if (viewMode === 'month') {
			viewMode = 'day';
		}
	}

	function getEventColor(event: CalendarEvent): string {
		// Color based on event type or source
		if (event.location?.toLowerCase().includes('teams') || event.location?.toLowerCase().includes('online')) {
			return 'bg-purple-100 border-purple-400 text-purple-800';
		}
		if (event.location?.toLowerCase().includes('переговорн') || event.location?.toLowerCase().includes('комнат')) {
			return 'bg-green-100 border-green-400 text-green-800';
		}
		return 'bg-blue-100 border-blue-400 text-blue-800';
	}

	function getUpcomingEvents(): CalendarEvent[] {
		const now = new Date();
		const todayStr = now.toISOString().split('T')[0];
		return events
			.filter(e => {
				const eventDate = e.start?.split('T')[0] || e.date || '';
				return eventDate >= todayStr;
			})
			.sort((a, b) => {
				const aDate = a.start || a.start_time || '';
				const bDate = b.start || b.start_time || '';
				return aDate.localeCompare(bDate);
			})
			.slice(0, 5);
	}

	function getCurrentTimePosition(): number {
		const now = new Date();
		const hours = now.getHours();
		const minutes = now.getMinutes();
		if (hours < 7 || hours >= 22) return -1;
		return ((hours - 7) * 60 + minutes) / 60 * 48;
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

<div class="flex h-[calc(100vh-100px)] gap-4">
	<!-- Left Sidebar - Mini Calendar & Upcoming Events -->
	<div class="w-72 flex-shrink-0 flex flex-col gap-4">
		<!-- New Event Button -->
		<button
			onclick={() => openNewEventDialog()}
			class="w-full px-4 py-3 bg-ekf-red text-white rounded-xl hover:bg-red-700 transition-colors flex items-center justify-center gap-2 shadow-sm"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
			</svg>
			Новое событие
		</button>

		<!-- Mini Calendar -->
		<div class="bg-white rounded-xl shadow-sm p-4">
			<div class="flex items-center justify-between mb-3">
				<button onclick={prevMonth} class="p-1 hover:bg-gray-100 rounded">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
				</button>
				<span class="text-sm font-medium text-gray-900">
					{months[currentDate.getMonth()]} {currentDate.getFullYear()}
				</span>
				<button onclick={nextMonth} class="p-1 hover:bg-gray-100 rounded">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
					</svg>
				</button>
			</div>
			<div class="grid grid-cols-7 gap-0.5">
				{#each ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'] as day}
					<div class="text-center text-xs text-gray-500 py-1">{day}</div>
				{/each}
				{#each getMiniCalendarDates() as week}
					{#each week as date}
						<button
							onclick={() => selectDate(date)}
							class="text-center text-xs py-1 rounded hover:bg-gray-100 transition-colors
								{isToday(date) ? 'bg-ekf-red text-white hover:bg-red-700' : ''}
								{isSelected(date) && !isToday(date) ? 'bg-blue-100 text-blue-700' : ''}
								{!isCurrentMonth(date) ? 'text-gray-300' : 'text-gray-700'}
								{getEventsForDate(date).length > 0 && !isToday(date) && !isSelected(date) ? 'font-bold' : ''}"
						>
							{date.getDate()}
						</button>
					{/each}
				{/each}
			</div>
		</div>

		<!-- Upcoming Events -->
		<div class="bg-white rounded-xl shadow-sm p-4 flex-1 overflow-auto">
			<h3 class="text-sm font-semibold text-gray-900 mb-3">Ближайшие события</h3>
			<div class="space-y-2">
				{#each getUpcomingEvents() as event}
					<button
						onclick={() => selectedEvent = event}
						class="w-full text-left p-2 rounded-lg hover:bg-gray-50 transition-colors border-l-2 {getEventColor(event)}"
					>
						<div class="text-sm font-medium truncate">{event.subject || event.title}</div>
						<div class="text-xs text-gray-500 mt-0.5">
							{new Date(event.start || event.start_time || '').toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })}
							{formatTime(event.start || event.start_time || '')}
						</div>
					</button>
				{/each}
				{#if getUpcomingEvents().length === 0}
					<p class="text-sm text-gray-400">Нет предстоящих событий</p>
				{/if}
			</div>
		</div>
	</div>

	<!-- Main Calendar Area -->
	<div class="flex-1 flex flex-col bg-white rounded-xl shadow-sm overflow-hidden">
		<!-- Header -->
		<div class="flex items-center justify-between p-4 border-b bg-gray-50">
			<div class="flex items-center gap-3">
				<button
					onclick={goToToday}
					class="px-3 py-1.5 text-sm border border-gray-300 rounded-lg hover:bg-white transition-colors"
				>
					Сегодня
				</button>
				<div class="flex items-center">
					<button onclick={prevPeriod} class="p-1.5 hover:bg-gray-200 rounded-lg transition-colors">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
						</svg>
					</button>
					<button onclick={nextPeriod} class="p-1.5 hover:bg-gray-200 rounded-lg transition-colors">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					</button>
				</div>
				<h2 class="text-lg font-semibold text-gray-900">
					{#if viewMode === 'day'}
						{currentDate.toLocaleDateString('ru-RU', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
					{:else if viewMode === 'week'}
						{#each [getWeekDates()] as wd}
							{wd[0].getDate()} {monthsShort[wd[0].getMonth()]} - {wd[6].getDate()} {monthsShort[wd[6].getMonth()]} {wd[6].getFullYear()}
						{/each}
					{:else}
						{months[currentDate.getMonth()]} {currentDate.getFullYear()}
					{/if}
				</h2>
			</div>
			<div class="flex items-center gap-2">
				<div class="flex rounded-lg border border-gray-300 overflow-hidden bg-white">
					<button
						onclick={() => viewMode = 'day'}
						class="px-3 py-1.5 text-sm {viewMode === 'day' ? 'bg-ekf-red text-white' : 'text-gray-600 hover:bg-gray-50'}"
					>
						День
					</button>
					<button
						onclick={() => viewMode = 'week'}
						class="px-3 py-1.5 text-sm border-x border-gray-300 {viewMode === 'week' ? 'bg-ekf-red text-white' : 'text-gray-600 hover:bg-gray-50'}"
					>
						Неделя
					</button>
					<button
						onclick={() => viewMode = 'month'}
						class="px-3 py-1.5 text-sm {viewMode === 'month' ? 'bg-ekf-red text-white' : 'text-gray-600 hover:bg-gray-50'}"
					>
						Месяц
					</button>
				</div>
				<button
					onclick={() => showSyncDialog = true}
					disabled={syncing}
					class="p-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
					title="Синхронизировать с Exchange"
				>
					<svg class="w-5 h-5 {syncing ? 'animate-spin' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
					</svg>
				</button>
			</div>
		</div>

		<!-- Calendar Content -->
		{#if loading}
			<div class="flex-1 flex items-center justify-center">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
			</div>
		{:else if viewMode === 'day'}
			<!-- Day View -->
			<div class="flex-1 overflow-auto">
				<div class="relative" style="min-height: {hours.length * 48}px;">
					<!-- Current time indicator -->
					{#if isToday(currentDate) && getCurrentTimePosition() > 0}
						<div
							class="absolute left-0 right-0 border-t-2 border-red-500 z-20 pointer-events-none"
							style="top: {getCurrentTimePosition()}px;"
						>
							<div class="absolute -left-1 -top-1.5 w-3 h-3 bg-red-500 rounded-full"></div>
						</div>
					{/if}

					{#each hours as hour}
						<div class="flex border-b border-gray-100" style="height: 48px;">
							<div class="w-16 flex-shrink-0 text-xs text-gray-400 text-right pr-2 pt-0.5">
								{hour.toString().padStart(2, '0')}:00
							</div>
							<button
								onclick={() => openNewEventDialog(currentDate, hour)}
								class="flex-1 border-l border-gray-200 hover:bg-blue-50 transition-colors relative"
							>
							</button>
						</div>
					{/each}

					<!-- Events overlay -->
					<div class="absolute left-16 right-0 top-0" style="height: {hours.length * 48}px;">
						{#each getEventsForDate(currentDate) as event}
							{@const pos = getEventPosition(event)}
							<button
								onclick={() => selectedEvent = event}
								class="absolute left-1 right-1 px-2 py-1 rounded border-l-4 text-left overflow-hidden {getEventColor(event)}"
								style="top: {pos.top}px; height: {pos.height}px;"
							>
								<div class="text-xs font-medium truncate">{event.subject || event.title}</div>
								{#if pos.height > 30}
									<div class="text-xs opacity-75 truncate">{formatEventTime(event)}</div>
								{/if}
								{#if pos.height > 50 && event.location}
									<div class="text-xs opacity-60 truncate">{event.location}</div>
								{/if}
							</button>
						{/each}
					</div>
				</div>
			</div>
		{:else if viewMode === 'week'}
			<!-- Week View -->
			<div class="flex-1 flex flex-col overflow-hidden">
				<!-- Week header -->
				<div class="flex border-b bg-gray-50">
					<div class="w-16 flex-shrink-0"></div>
					{#each getWeekDates() as date, i}
						<div class="flex-1 p-2 text-center border-l border-gray-200 {isToday(date) ? 'bg-blue-50' : ''}">
							<div class="text-xs text-gray-500">{weekDays[i]}</div>
							<button
								onclick={() => { currentDate = date; viewMode = 'day'; }}
								class="text-lg font-semibold mt-0.5 w-8 h-8 rounded-full inline-flex items-center justify-center hover:bg-gray-200 transition-colors
									{isToday(date) ? 'bg-ekf-red text-white hover:bg-red-700' : 'text-gray-900'}"
							>
								{date.getDate()}
							</button>
						</div>
					{/each}
				</div>

				<!-- Time grid -->
				<div class="flex-1 overflow-auto">
					<div class="relative" style="min-height: {hours.length * 48}px;">
						<!-- Current time indicator -->
						{#each [getWeekDates()] as weekDatesArr}
							{@const todayIndex = weekDatesArr.findIndex(d => isToday(d))}
							{#if todayIndex >= 0 && getCurrentTimePosition() > 0}
								<div
									class="absolute border-t-2 border-red-500 z-20 pointer-events-none"
									style="top: {getCurrentTimePosition()}px; left: calc(64px + {todayIndex} * (100% - 64px) / 7); width: calc((100% - 64px) / 7);"
								>
									<div class="absolute -left-1 -top-1.5 w-3 h-3 bg-red-500 rounded-full"></div>
								</div>
							{/if}
						{/each}

						{#each hours as hour}
							<div class="flex border-b border-gray-100" style="height: 48px;">
								<div class="w-16 flex-shrink-0 text-xs text-gray-400 text-right pr-2 pt-0.5">
									{hour.toString().padStart(2, '0')}:00
								</div>
								{#each getWeekDates() as date, i}
									<button
										onclick={() => openNewEventDialog(date, hour)}
										class="flex-1 border-l border-gray-200 hover:bg-blue-50 transition-colors {isToday(date) ? 'bg-blue-50/30' : ''}"
									>
									</button>
								{/each}
							</div>
						{/each}

						<!-- Events overlay -->
						{#each getWeekDates() as date, dayIndex}
							<div
								class="absolute top-0"
								style="left: calc(64px + {dayIndex} * (100% - 64px) / 7); width: calc((100% - 64px) / 7); height: {hours.length * 48}px;"
							>
								{#each getEventsForDate(date) as event}
									{@const pos = getEventPosition(event)}
									<button
										onclick={() => selectedEvent = event}
										class="absolute left-0.5 right-0.5 px-1 py-0.5 rounded border-l-2 text-left overflow-hidden text-xs {getEventColor(event)}"
										style="top: {pos.top}px; height: {pos.height}px;"
									>
										<div class="font-medium truncate">{event.subject || event.title}</div>
										{#if pos.height > 24}
											<div class="opacity-75 truncate">{formatEventTime(event)}</div>
										{/if}
									</button>
								{/each}
							</div>
						{/each}
					</div>
				</div>
			</div>
		{:else}
			<!-- Month View -->
			<div class="flex-1 flex flex-col overflow-hidden">
				<div class="grid grid-cols-7 border-b bg-gray-50">
					{#each weekDays as day}
						<div class="p-2 text-center text-sm font-medium text-gray-600 border-r last:border-r-0">
							{day}
						</div>
					{/each}
				</div>
				<div class="flex-1 grid grid-rows-6 overflow-hidden">
					{#each getMonthDates() as week}
						<div class="grid grid-cols-7 border-b last:border-b-0 min-h-0">
							{#each week as date}
								<button
									onclick={() => selectDate(date)}
									class="p-1 border-r last:border-r-0 text-left overflow-hidden hover:bg-gray-50 transition-colors flex flex-col
										{isToday(date) ? 'bg-blue-50' : ''}
										{!isCurrentMonth(date) ? 'bg-gray-50' : ''}"
								>
									<div class="text-sm font-medium mb-0.5 flex items-center gap-1
										{isToday(date) ? 'text-ekf-red' : ''}
										{!isCurrentMonth(date) ? 'text-gray-400' : 'text-gray-900'}">
										{#if isToday(date)}
											<span class="w-6 h-6 rounded-full bg-ekf-red text-white flex items-center justify-center text-xs">
												{date.getDate()}
											</span>
										{:else}
											{date.getDate()}
										{/if}
									</div>
									<div class="flex-1 overflow-hidden space-y-0.5">
										{#each getEventsForDate(date).slice(0, 3) as event}
											<div class="text-xs px-1 py-0.5 rounded truncate {getEventColor(event)}">
												{event.subject || event.title}
											</div>
										{/each}
										{#if getEventsForDate(date).length > 3}
											<div class="text-xs text-gray-500 px-1">+{getEventsForDate(date).length - 3} ещё</div>
										{/if}
									</div>
								</button>
							{/each}
						</div>
					{/each}
				</div>
			</div>
		{/if}
	</div>

	<!-- Right Sidebar - Event Details -->
	{#if selectedEvent}
		<div class="w-80 flex-shrink-0 bg-white rounded-xl shadow-sm p-4 overflow-auto">
			<div class="flex items-start justify-between mb-4">
				<h3 class="text-lg font-semibold text-gray-900">{selectedEvent.subject || selectedEvent.title}</h3>
				<button
					onclick={() => selectedEvent = null}
					class="p-1 hover:bg-gray-100 rounded"
				>
					<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="space-y-4">
				<div class="flex items-center gap-3 text-gray-600">
					<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
					<div>
						<div class="text-sm font-medium text-gray-900">
							{new Date(selectedEvent.start || selectedEvent.start_time || '').toLocaleDateString('ru-RU', { weekday: 'long', day: 'numeric', month: 'long' })}
						</div>
						<div class="text-sm">{formatEventTime(selectedEvent)}</div>
					</div>
				</div>

				{#if selectedEvent.location}
					<div class="flex items-center gap-3 text-gray-600">
						<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
						</svg>
						<span class="text-sm">{selectedEvent.location}</span>
					</div>
				{/if}

				{#if selectedEvent.organizer}
					<div class="flex items-center gap-3 text-gray-600">
						<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
						</svg>
						<div class="text-sm">
							<span class="text-gray-500">Организатор:</span>
							<span class="text-gray-900">{selectedEvent.organizer.name || selectedEvent.organizer.email}</span>
						</div>
					</div>
				{/if}

				{#if selectedEvent.attendees && selectedEvent.attendees.length > 0}
					<div>
						<div class="flex items-center gap-3 text-gray-600 mb-2">
							<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
							</svg>
							<span class="text-sm font-medium">Участники ({selectedEvent.attendees.length})</span>
						</div>
						<div class="ml-8 space-y-1">
							{#each selectedEvent.attendees.slice(0, 5) as attendee}
								<div class="text-sm text-gray-600">{attendee.name || attendee.email}</div>
							{/each}
							{#if selectedEvent.attendees.length > 5}
								<div class="text-sm text-gray-400">+{selectedEvent.attendees.length - 5} ещё</div>
							{/if}
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<!-- Sync Dialog -->
{#if showSyncDialog}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
		<div class="bg-white rounded-xl p-6 w-full max-w-md shadow-xl">
			<h3 class="text-lg font-semibold text-gray-900 mb-4">Синхронизация с Exchange</h3>
			<p class="text-gray-600 mb-6">
				Будет выполнена синхронизация вашего календаря из Exchange Server (последние 30 дней и 60 дней вперёд).
			</p>
			<div class="flex justify-end gap-3">
				<button
					onclick={() => showSyncDialog = false}
					class="px-4 py-2 border border-gray-200 rounded-lg hover:bg-gray-50"
				>
					Отмена
				</button>
				<button
					onclick={syncCalendar}
					disabled={syncing}
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
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onclick={() => showEventDialog = false}>
		<div class="bg-white rounded-xl shadow-xl w-full max-w-lg max-h-[90vh] overflow-hidden flex flex-col" onclick={(e) => e.stopPropagation()}>
			<div class="p-4 border-b flex items-center justify-between">
				<h3 class="text-lg font-semibold text-gray-900">Новое событие</h3>
				<button onclick={() => showEventDialog = false} class="p-1 hover:bg-gray-100 rounded">
					<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>
			<div class="flex-1 overflow-y-auto p-4 space-y-4">
				<!-- Title -->
				<div>
					<input
						type="text"
						bind:value={newEvent.title}
						placeholder="Добавить название"
						class="w-full px-3 py-2 text-lg font-medium border-0 border-b border-gray-200 focus:outline-none focus:border-ekf-red"
					/>
				</div>

				<!-- Date & Time -->
				<div class="flex items-center gap-3 text-sm">
					<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
					</svg>
					<input
						type="date"
						bind:value={newEvent.date}
						class="px-2 py-1 border border-gray-200 rounded focus:outline-none focus:border-ekf-red"
					/>
					<input
						type="time"
						bind:value={newEvent.start_time}
						class="px-2 py-1 border border-gray-200 rounded focus:outline-none focus:border-ekf-red"
					/>
					<span class="text-gray-400">—</span>
					<input
						type="time"
						bind:value={newEvent.end_time}
						class="px-2 py-1 border border-gray-200 rounded focus:outline-none focus:border-ekf-red"
					/>
				</div>

				<!-- Participants -->
				<div>
					<div class="flex items-center gap-3 mb-2">
						<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
						</svg>
						<span class="text-sm text-gray-700">Участники</span>
					</div>
					<!-- Selected participants -->
					{#if newEvent.participants.length > 0}
						<div class="flex flex-wrap gap-2 mb-2 ml-8">
							{#each newEvent.participants as pid}
								{@const emp = getEmployeeById(pid)}
								<span class="inline-flex items-center gap-1 px-2 py-1 bg-gray-100 rounded-full text-sm">
									<span class="w-5 h-5 rounded-full bg-ekf-red text-white flex items-center justify-center text-xs">
										{emp.name?.charAt(0) || '?'}
									</span>
									{emp.name}
									<button onclick={() => removeParticipant(pid)} class="ml-1 text-gray-400 hover:text-gray-600">
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
										</svg>
									</button>
								</span>
							{/each}
						</div>
					{/if}
					<!-- Search input -->
					<div class="relative ml-8">
						<input
							type="text"
							bind:value={participantSearch}
							onfocus={() => showParticipantDropdown = true}
							placeholder="Добавить участника..."
							class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-ekf-red"
						/>
						{#if showParticipantDropdown && getFilteredEmployees().length > 0}
							<div class="absolute top-full left-0 right-0 mt-1 bg-white border border-gray-200 rounded-lg shadow-lg z-10 max-h-40 overflow-auto">
								{#each getFilteredEmployees() as emp}
									<button
										onclick={() => addParticipant(emp.id)}
										class="w-full px-3 py-2 text-left hover:bg-gray-50 flex items-center gap-2 text-sm"
									>
										<div class="w-6 h-6 rounded-full bg-gray-200 flex items-center justify-center text-xs">
											{emp.name?.charAt(0) || '?'}
										</div>
										<div>
											<div class="font-medium">{emp.name}</div>
											{#if emp.position}
												<div class="text-xs text-gray-400">{emp.position}</div>
											{/if}
										</div>
									</button>
								{/each}
							</div>
						{/if}
					</div>
				</div>

				<!-- Location Type Toggle -->
				<div>
					<div class="flex items-center gap-3 mb-2">
						<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
						</svg>
						<span class="text-sm text-gray-700">Место проведения</span>
					</div>
					<div class="ml-8 space-y-3">
						<!-- Location type buttons -->
						<div class="flex gap-2">
							<button
								onclick={() => { newEvent.is_online = false; newEvent.room = ''; }}
								class="px-3 py-1.5 text-sm rounded-lg border transition-colors
									{!newEvent.is_online && !newEvent.room ? 'border-ekf-red bg-red-50 text-ekf-red' : 'border-gray-200 hover:bg-gray-50'}"
							>
								Без места
							</button>
							<button
								onclick={() => { newEvent.is_online = false; }}
								class="px-3 py-1.5 text-sm rounded-lg border transition-colors
									{!newEvent.is_online && newEvent.room ? 'border-ekf-red bg-red-50 text-ekf-red' : 'border-gray-200 hover:bg-gray-50'}"
							>
								Переговорная
							</button>
							<button
								onclick={() => { newEvent.is_online = true; newEvent.room = ''; }}
								class="px-3 py-1.5 text-sm rounded-lg border transition-colors
									{newEvent.is_online ? 'border-ekf-red bg-red-50 text-ekf-red' : 'border-gray-200 hover:bg-gray-50'}"
							>
								Онлайн
							</button>
						</div>

						<!-- Meeting room selector -->
						{#if !newEvent.is_online}
							<select
								bind:value={newEvent.room}
								class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-ekf-red"
							>
								<option value="">Выберите переговорную</option>
								{#each meetingRooms as room}
									<option value={room.id}>{room.name} — {room.floor} (до {room.capacity} чел.)</option>
								{/each}
							</select>
						{/if}

						<!-- Online service selector -->
						{#if newEvent.is_online}
							<div class="space-y-2">
								<p class="text-xs text-gray-500">Выберите сервис для видеоконференции:</p>
								<div class="flex gap-2">
									{#each onlineServices as service}
										<button
											onclick={() => newEvent.online_service = service.id}
											class="flex-1 px-3 py-2 text-sm rounded-lg border transition-colors flex items-center justify-center gap-2
												{newEvent.online_service === service.id ? 'border-ekf-red bg-red-50 text-ekf-red' : 'border-gray-200 hover:bg-gray-50'}"
										>
											<span>{service.icon}</span>
											<span>{service.name}</span>
										</button>
									{/each}
								</div>
								<p class="text-xs text-gray-400">Ссылка на встречу будет сгенерирована автоматически</p>
							</div>
						{/if}

						<!-- Custom location input (if no room selected and not online) -->
						{#if !newEvent.is_online && !newEvent.room}
							<input
								type="text"
								bind:value={newEvent.location}
								placeholder="Или укажите место вручную..."
								class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-ekf-red"
							/>
						{/if}
					</div>
				</div>
			</div>
			<div class="p-4 border-t flex justify-end gap-3">
				<button
					onclick={() => showEventDialog = false}
					class="px-4 py-2 text-gray-600 hover:text-gray-900 text-sm"
				>
					Отмена
				</button>
				<button
					onclick={createEvent}
					disabled={!newEvent.title || !newEvent.date}
					class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 disabled:opacity-50 text-sm"
				>
					Создать событие
				</button>
			</div>
		</div>
	</div>
{/if}
