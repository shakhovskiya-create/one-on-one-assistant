<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { calendar as calendarApi, meetings as meetingsApi, speech } from '$lib/api/client';
	import type { CalendarEvent, MeetingRoom } from '$lib/api/client';
	import { user, subordinates } from '$lib/stores/auth';

	let events: CalendarEvent[] = $state([]);
	let loading = $state(true);
	let syncing = $state(false);
	let showSyncDialog = $state(false);
	let showEventDialog = $state(false);
	let selectedEvent: CalendarEvent | null = $state(null);
	let showEventModal = $state(false);
	let modalEvent: CalendarEvent | null = $state(null);

	// Edit/Delete state
	let isEditingEvent = $state(false);
	let editingEvent = $state({
		subject: '',
		start: '',
		end: '',
		location: ''
	});
	let deletingEvent = $state(false);
	let updatingEvent = $state(false);

	// Transcription state
	let isRecording = $state(false);
	let recordingTime = $state(0);
	let mediaRecorder: MediaRecorder | null = $state(null);
	let audioChunks: Blob[] = $state([]);
	let recordingInterval: ReturnType<typeof setInterval> | null = null;
	let transcript = $state('');
	let transcribing = $state(false);
	let showTranscriptSection = $state(false);

	function openEventModal(event: CalendarEvent) {
		modalEvent = event;
		showEventModal = true;
		isEditingEvent = false;
		// Reset transcription state
		transcript = '';
		showTranscriptSection = false;
		isRecording = false;
		recordingTime = 0;
	}

	function startEditEvent() {
		if (!modalEvent) return;
		const startDate = modalEvent.start || modalEvent.start_time || '';
		const endDate = modalEvent.end || modalEvent.end_time || '';
		editingEvent = {
			subject: modalEvent.subject || modalEvent.title || '',
			start: startDate.slice(0, 16), // Format: YYYY-MM-DDTHH:MM
			end: endDate.slice(0, 16),
			location: modalEvent.location || ''
		};
		isEditingEvent = true;
	}

	function cancelEditEvent() {
		isEditingEvent = false;
	}

	async function saveEditEvent() {
		if (!modalEvent || !modalEvent.id) return;
		updatingEvent = true;
		try {
			await calendarApi.updateMeeting({
				item_id: modalEvent.id,
				subject: editingEvent.subject,
				start: editingEvent.start + ':00',
				end: editingEvent.end + ':00',
				location: editingEvent.location
			});
			await loadCalendar();
			showEventModal = false;
			modalEvent = null;
			isEditingEvent = false;
		} catch (e: any) {
			console.error('Failed to update event:', e);
			alert('Ошибка обновления события: ' + (e.message || 'Unknown error'));
		} finally {
			updatingEvent = false;
		}
	}

	async function deleteEvent(sendCancellations: boolean = true) {
		if (!modalEvent || !modalEvent.id) return;
		if (!confirm(sendCancellations
			? 'Удалить событие и отправить уведомления участникам?'
			: 'Удалить событие без уведомлений?')) {
			return;
		}
		deletingEvent = true;
		try {
			await calendarApi.deleteMeeting({
				item_id: modalEvent.id,
				send_cancellations: sendCancellations
			});
			await loadCalendar();
			showEventModal = false;
			modalEvent = null;
		} catch (e: any) {
			console.error('Failed to delete event:', e);
			alert('Ошибка удаления события: ' + (e.message || 'Unknown error'));
		} finally {
			deletingEvent = false;
		}
	}

	async function startRecording() {
		try {
			const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
			mediaRecorder = new MediaRecorder(stream);
			audioChunks = [];

			mediaRecorder.ondataavailable = (e) => {
				audioChunks.push(e.data);
			};

			mediaRecorder.onstop = async () => {
				const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
				stream.getTracks().forEach(track => track.stop());
				await transcribeAudio(audioBlob);
			};

			mediaRecorder.start();
			isRecording = true;
			recordingTime = 0;
			showTranscriptSection = true;

			recordingInterval = setInterval(() => {
				recordingTime++;
			}, 1000);
		} catch (err) {
			console.error('Error starting recording:', err);
			alert('Не удалось получить доступ к микрофону');
		}
	}

	function stopRecording() {
		if (mediaRecorder && isRecording) {
			mediaRecorder.stop();
			isRecording = false;
			if (recordingInterval) {
				clearInterval(recordingInterval);
				recordingInterval = null;
			}
		}
	}

	async function transcribeAudio(audioBlob: Blob) {
		transcribing = true;
		try {
			const result = await speech.transcribe(audioBlob, 'auto');
			transcript = result.transcript || '';
		} catch (err: any) {
			console.error('Transcription error:', err);
			alert('Ошибка транскрибирования: ' + (err.message || 'Unknown error'));
		} finally {
			transcribing = false;
		}
	}

	function formatRecordingTime(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
	}

	async function handleAudioFileUpload(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];
		if (file) {
			showTranscriptSection = true;
			await transcribeAudio(file);
		}
	}

	let currentDate = $state(new Date());
	let viewMode = $state<'day' | 'week' | 'month'>('week');

	// Drag-to-create state
	let isDragging = $state(false);
	let dragStartDate: Date | null = $state(null);
	let dragStartMinute = $state(0);
	let dragEndMinute = $state(0);
	let dragCurrentDate: Date | null = $state(null);

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
		participants: [] as string[],
		recurrence: 'none' as 'none' | 'daily' | 'weekly' | 'monthly' | 'yearly',
		recurrence_end: '' as string,
		recurrence_count: 10,
		// Extended recurrence options
		recurrence_interval: 1, // every N days/weeks/months/years
		recurrence_weekdays: [] as number[], // 0=Mon, 1=Tue, ..., 6=Sun for weekly
		recurrence_monthly_type: 'day' as 'day' | 'weekday', // day=on day X, weekday=on Nth weekday
	});

	// Recurrence options
	const recurrenceOptions = [
		{ id: 'none', name: 'Не повторять' },
		{ id: 'daily', name: 'Ежедневно' },
		{ id: 'weekly', name: 'Еженедельно' },
		{ id: 'monthly', name: 'Ежемесячно' },
		{ id: 'yearly', name: 'Ежегодно' }
	];

	// Weekday buttons for weekly recurrence
	const weekdayButtons = [
		{ id: 0, short: 'Пн', full: 'Понедельник' },
		{ id: 1, short: 'Вт', full: 'Вторник' },
		{ id: 2, short: 'Ср', full: 'Среда' },
		{ id: 3, short: 'Чт', full: 'Четверг' },
		{ id: 4, short: 'Пт', full: 'Пятница' },
		{ id: 5, short: 'Сб', full: 'Суббота' },
		{ id: 6, short: 'Вс', full: 'Воскресенье' }
	];

	function toggleWeekday(dayId: number) {
		if (newEvent.recurrence_weekdays.includes(dayId)) {
			// Don't allow removing the last weekday
			if (newEvent.recurrence_weekdays.length > 1) {
				newEvent.recurrence_weekdays = newEvent.recurrence_weekdays.filter(d => d !== dayId);
			}
		} else {
			newEvent.recurrence_weekdays = [...newEvent.recurrence_weekdays, dayId].sort((a, b) => a - b);
		}
	}

	function getWeekdayOrdinal(date: Date): { ordinal: number; weekday: number; weekdayName: string } {
		const day = date.getDate();
		const ordinal = Math.ceil(day / 7); // 1st, 2nd, 3rd, 4th, 5th
		const weekday = date.getDay();
		const weekdayIndex = weekday === 0 ? 6 : weekday - 1;
		const weekdayName = weekdayButtons[weekdayIndex].full;
		return { ordinal, weekday: weekdayIndex, weekdayName };
	}

	function getOrdinalText(n: number): string {
		switch (n) {
			case 1: return 'первый';
			case 2: return 'второй';
			case 3: return 'третий';
			case 4: return 'четвёртый';
			case 5: return 'пятый';
			default: return `${n}-й`;
		}
	}

	function getSelectedWeekdaysText(): string {
		if (newEvent.recurrence_weekdays.length === 0) return '';
		if (newEvent.recurrence_weekdays.length === 7) return 'каждый день';
		if (newEvent.recurrence_weekdays.length === 5 &&
			!newEvent.recurrence_weekdays.includes(5) &&
			!newEvent.recurrence_weekdays.includes(6)) {
			return 'по будням';
		}
		return newEvent.recurrence_weekdays.map(d => weekdayButtons[d].short).join(', ');
	}

	// Meeting rooms loaded from Exchange
	let meetingRooms: MeetingRoom[] = $state([]);
	let loadingRooms = $state(false);

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
		// Handle URL parameters for navigation
		const dateParam = $page.url.searchParams.get('date');
		if (dateParam) {
			const parsedDate = new Date(dateParam);
			if (!isNaN(parsedDate.getTime())) {
				currentDate = parsedDate;
			}
		}

		if ($user?.id) {
			await Promise.all([
				loadCalendar(),
				loadMeetingRooms()
			]);
		}
	});

	async function loadMeetingRooms() {
		if (!$user?.id) return;
		loadingRooms = true;
		try {
			const response = await calendarApi.getRooms($user.id);
			meetingRooms = response.rooms || [];
		} catch (e) {
			console.error('Failed to load meeting rooms:', e);
			// Fallback to empty - user can enter location manually
			meetingRooms = [];
		} finally {
			loadingRooms = false;
		}
	}

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

	function openNewEventDialog(date?: Date, hour?: number, startMinute?: number, endMinute?: number) {
		const d = date || new Date();

		// Calculate start and end times
		let startHour: number;
		let startMin: number;
		let endHour: number;
		let endMin: number;

		if (startMinute !== undefined && endMinute !== undefined) {
			// Use provided minute range (from drag selection)
			const minStart = Math.min(startMinute, endMinute);
			const maxEnd = Math.max(startMinute, endMinute);
			startHour = Math.floor(minStart / 60) + 7; // 7 is the first hour in the grid
			startMin = minStart % 60;
			endHour = Math.floor(maxEnd / 60) + 7;
			endMin = maxEnd % 60;
			// Round to nearest 15 minutes
			startMin = Math.round(startMin / 15) * 15;
			endMin = Math.round(endMin / 15) * 15;
			if (startMin === 60) { startHour++; startMin = 0; }
			if (endMin === 60) { endHour++; endMin = 0; }
			// Ensure at least 30 min duration
			if (startHour === endHour && startMin === endMin) {
				endMin += 30;
				if (endMin >= 60) { endHour++; endMin -= 60; }
			}
		} else if (hour !== undefined) {
			// Single click on hour
			startHour = hour;
			startMin = 0;
			endHour = hour + 1;
			endMin = 0;
		} else {
			// Default
			startHour = 9;
			startMin = 0;
			endHour = 10;
			endMin = 0;
		}

		// Get day of week for weekly recurrence default
		const dayOfWeek = d.getDay();
		const weekdayIndex = dayOfWeek === 0 ? 6 : dayOfWeek - 1; // Convert Sun=0 to Mon=0 format

		newEvent = {
			title: '',
			date: d.toISOString().split('T')[0],
			start_time: `${startHour.toString().padStart(2, '0')}:${startMin.toString().padStart(2, '0')}`,
			end_time: `${endHour.toString().padStart(2, '0')}:${endMin.toString().padStart(2, '0')}`,
			location: '',
			employee_id: $user?.id || '',
			room: '',
			is_online: false,
			online_service: 'teams',
			participants: [],
			recurrence: 'none',
			recurrence_end: '',
			recurrence_count: 10,
			recurrence_interval: 1,
			recurrence_weekdays: [weekdayIndex],
			recurrence_monthly_type: 'day'
		};
		participantSearch = '';
		showEventDialog = true;
	}

	// Drag-to-create handlers
	function getMinuteFromMouseY(e: MouseEvent, containerTop: number): number {
		const relY = e.clientY - containerTop;
		// Each hour is 48px, so minute = (relY / 48) * 60
		const minute = (relY / 48) * 60;
		return Math.max(0, Math.min(minute, hours.length * 60));
	}

	function handleDragStart(e: MouseEvent, date: Date, containerTop: number) {
		// Prevent text selection
		e.preventDefault();
		isDragging = true;
		dragStartDate = date;
		dragCurrentDate = date;
		const minute = getMinuteFromMouseY(e, containerTop);
		dragStartMinute = minute;
		dragEndMinute = minute;
	}

	function handleDragMove(e: MouseEvent, date: Date, containerTop: number) {
		if (!isDragging) return;
		dragCurrentDate = date;
		dragEndMinute = getMinuteFromMouseY(e, containerTop);
	}

	function handleDragEnd() {
		if (!isDragging || !dragStartDate) {
			isDragging = false;
			return;
		}

		// Check if we actually dragged (not just clicked)
		const timeDiff = Math.abs(dragEndMinute - dragStartMinute);
		if (timeDiff > 10) {
			// Open dialog with dragged time range
			openNewEventDialog(dragCurrentDate || dragStartDate, undefined, dragStartMinute, dragEndMinute);
		}

		isDragging = false;
		dragStartDate = null;
		dragCurrentDate = null;
	}

	function getDragSelectionStyle(date: Date): { top: number; height: number } | null {
		if (!isDragging || !dragStartDate) return null;

		// Check if this date matches the drag date
		const dateStr = date.toISOString().split('T')[0];
		const dragDateStr = (dragCurrentDate || dragStartDate).toISOString().split('T')[0];
		if (dateStr !== dragDateStr) return null;

		const minMinute = Math.min(dragStartMinute, dragEndMinute);
		const maxMinute = Math.max(dragStartMinute, dragEndMinute);

		return {
			top: (minMinute / 60) * 48,
			height: Math.max(((maxMinute - minMinute) / 60) * 48, 12)
		};
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
				const room = meetingRooms.find(r => r.email === newEvent.room);
				location = room?.name || newEvent.location;
			} else {
				location = newEvent.location;
			}

			// Create meeting in Exchange via EWS
			await calendarApi.createMeeting({
				subject: newEvent.title,
				start: `${newEvent.date}T${newEvent.start_time}:00`,
				end: `${newEvent.date}T${newEvent.end_time}:00`,
				location: location || undefined,
				required_attendees: newEvent.participants.length > 0 ? newEvent.participants : undefined,
				is_online_meeting: newEvent.is_online
			});

			// Reload calendar to show the new event
			await loadCalendar();
			showEventDialog = false;
		} catch (e: any) {
			console.error(e);
			alert('Ошибка создания события: ' + (e.message || 'Unknown error'));
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
	<title>Календарь - EKF Hub</title>
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
						ondblclick={() => openEventModal(event)}
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
				<div
					class="relative select-none"
					style="min-height: {hours.length * 48}px;"
					role="grid"
					onmouseup={handleDragEnd}
					onmouseleave={handleDragEnd}
				>
					<!-- Current time indicator -->
					{#if isToday(currentDate) && getCurrentTimePosition() > 0}
						<div
							class="absolute left-0 right-0 border-t-2 border-red-500 z-20 pointer-events-none"
							style="top: {getCurrentTimePosition()}px;"
						>
							<div class="absolute -left-1 -top-1.5 w-3 h-3 bg-red-500 rounded-full"></div>
						</div>
					{/if}

					<!-- Time grid rows -->
					{#each hours as hour}
						<div class="flex border-b border-gray-100" style="height: 48px;">
							<div class="w-16 flex-shrink-0 text-xs text-gray-400 text-right pr-2 pt-0.5">
								{hour.toString().padStart(2, '0')}:00
							</div>
							<div class="flex-1 border-l border-gray-200"></div>
						</div>
					{/each}

					<!-- Draggable overlay for day view -->
					<div
						class="absolute left-16 right-0 top-0 cursor-crosshair"
						style="height: {hours.length * 48}px;"
						role="button"
						tabindex="0"
						onmousedown={(e) => {
							const rect = e.currentTarget.getBoundingClientRect();
							handleDragStart(e, currentDate, rect.top);
						}}
						onmousemove={(e) => {
							if (isDragging) {
								const rect = e.currentTarget.getBoundingClientRect();
								handleDragMove(e, currentDate, rect.top);
							}
						}}
						onclick={(e) => {
							// Single click to create event at that hour
							if (!isDragging || Math.abs(dragEndMinute - dragStartMinute) <= 10) {
								const rect = e.currentTarget.getBoundingClientRect();
								const minute = getMinuteFromMouseY(e, rect.top);
								const hour = Math.floor(minute / 60) + 7;
								openNewEventDialog(currentDate, hour);
							}
						}}
					>
						<!-- Drag selection indicator -->
						{#if isDragging}
							{@const sel = getDragSelectionStyle(currentDate)}
							{@const minMin = Math.min(dragStartMinute, dragEndMinute)}
							{@const maxMin = Math.max(dragStartMinute, dragEndMinute)}
							{#if sel}
								<div
									class="absolute left-1 right-1 bg-blue-200/70 border-2 border-blue-400 rounded pointer-events-none z-10"
									style="top: {sel.top}px; height: {sel.height}px;"
								>
									<div class="px-2 py-1 text-xs text-blue-700 font-medium">
										{Math.floor(minMin / 60) + 7}:{(minMin % 60).toString().padStart(2, '0')} - {Math.floor(maxMin / 60) + 7}:{(maxMin % 60).toString().padStart(2, '0')}
									</div>
								</div>
							{/if}
						{/if}

						<!-- Events -->
						{#each getEventsForDate(currentDate) as event}
							{@const pos = getEventPosition(event)}
							<button
								onclick={(e) => { e.stopPropagation(); selectedEvent = event; }}
								ondblclick={(e) => { e.stopPropagation(); openEventModal(event); }}
								class="absolute left-1 right-1 px-2 py-1 rounded border-l-4 text-left overflow-hidden cursor-pointer {getEventColor(event)}"
								style="top: {pos.top}px; height: {pos.height}px; z-index: 5;"
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
				<div
					class="flex-1 overflow-auto"
					onmouseup={handleDragEnd}
					onmouseleave={handleDragEnd}
				>
					<div class="relative select-none" style="min-height: {hours.length * 48}px;">
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

						<!-- Time labels and grid lines -->
						{#each hours as hour}
							<div class="flex border-b border-gray-100" style="height: 48px;">
								<div class="w-16 flex-shrink-0 text-xs text-gray-400 text-right pr-2 pt-0.5">
									{hour.toString().padStart(2, '0')}:00
								</div>
								{#each getWeekDates() as date}
									<div class="flex-1 border-l border-gray-200 {isToday(date) ? 'bg-blue-50/30' : ''}"></div>
								{/each}
							</div>
						{/each}

						<!-- Draggable day columns overlay -->
						{#each getWeekDates() as date, dayIndex}
							<div
								class="absolute top-0 cursor-crosshair"
								style="left: calc(64px + {dayIndex} * (100% - 64px) / 7); width: calc((100% - 64px) / 7); height: {hours.length * 48}px;"
								role="button"
								tabindex="0"
								onmousedown={(e) => {
									const rect = e.currentTarget.getBoundingClientRect();
									handleDragStart(e, date, rect.top);
								}}
								onmousemove={(e) => {
									if (isDragging) {
										const rect = e.currentTarget.getBoundingClientRect();
										handleDragMove(e, date, rect.top);
									}
								}}
								onclick={(e) => {
									// Single click to create event at that hour
									if (!isDragging || Math.abs(dragEndMinute - dragStartMinute) <= 10) {
										const rect = e.currentTarget.getBoundingClientRect();
										const minute = getMinuteFromMouseY(e, rect.top);
										const hour = Math.floor(minute / 60) + 7;
										openNewEventDialog(date, hour);
									}
								}}
							>
								<!-- Drag selection indicator -->
								{#if isDragging}
									{@const sel = getDragSelectionStyle(date)}
									{@const minMin = Math.min(dragStartMinute, dragEndMinute)}
									{@const maxMin = Math.max(dragStartMinute, dragEndMinute)}
									{#if sel}
										<div
											class="absolute left-0.5 right-0.5 bg-blue-200/70 border-2 border-blue-400 rounded pointer-events-none z-10"
											style="top: {sel.top}px; height: {sel.height}px;"
										>
											<div class="px-1 py-0.5 text-xs text-blue-700 font-medium truncate">
												{Math.floor(minMin / 60) + 7}:{(minMin % 60).toString().padStart(2, '0')}-{Math.floor(maxMin / 60) + 7}:{(maxMin % 60).toString().padStart(2, '0')}
											</div>
										</div>
									{/if}
								{/if}

								<!-- Events -->
								{#each getEventsForDate(date) as event}
									{@const pos = getEventPosition(event)}
									<button
										onclick={(e) => { e.stopPropagation(); selectedEvent = event; }}
										ondblclick={(e) => { e.stopPropagation(); openEventModal(event); }}
										class="absolute left-0.5 right-0.5 px-1 py-0.5 rounded border-l-2 text-left overflow-hidden text-xs cursor-pointer {getEventColor(event)}"
										style="top: {pos.top}px; height: {pos.height}px; z-index: 5;"
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
						<div class="ml-8 max-h-48 overflow-y-auto space-y-2">
							{#each selectedEvent.attendees as attendee}
								<div class="flex items-center gap-2 text-sm group">
									<!-- Avatar with initials -->
									<div class="w-7 h-7 rounded-full bg-gradient-to-br from-blue-400 to-blue-600 flex items-center justify-center text-white text-xs font-medium flex-shrink-0">
										{(attendee.name || attendee.email).split(' ').map(n => n[0]).slice(0, 2).join('').toUpperCase()}
									</div>
									<!-- Response status indicator -->
									{#if attendee.response === 'Accept'}
										<span class="w-5 h-5 flex-shrink-0 rounded-full bg-green-100 text-green-600 flex items-center justify-center" title="Принял">
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
											</svg>
										</span>
									{:else if attendee.response === 'Decline'}
										<span class="w-5 h-5 flex-shrink-0 rounded-full bg-red-100 text-red-600 flex items-center justify-center" title="Отклонил">
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M6 18L18 6M6 6l12 12" />
											</svg>
										</span>
									{:else if attendee.response === 'Tentative'}
										<span class="w-5 h-5 flex-shrink-0 rounded-full bg-yellow-100 text-yellow-600 flex items-center justify-center" title="Возможно">
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M12 6v6m0 4h.01" />
											</svg>
										</span>
									{:else}
										<span class="w-5 h-5 flex-shrink-0 rounded-full bg-gray-100 text-gray-400 flex items-center justify-center" title="Ожидает ответа">
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01" />
											</svg>
										</span>
									{/if}
									<div class="flex-1 min-w-0">
										<span class="text-gray-700 truncate block" title={attendee.email}>
											{attendee.name || attendee.email}
										</span>
										{#if attendee.optional}
											<span class="text-xs text-gray-400">Необязательный</span>
										{/if}
									</div>
								</div>
							{/each}
						</div>
						<!-- Response summary -->
						{#if selectedEvent.attendees.length > 0}
							{@const accepted = selectedEvent.attendees.filter(a => a.response === 'Accept').length}
							{@const declined = selectedEvent.attendees.filter(a => a.response === 'Decline').length}
							{@const tentative = selectedEvent.attendees.filter(a => a.response === 'Tentative').length}
							{@const pending = selectedEvent.attendees.length - accepted - declined - tentative}
							<div class="ml-8 mt-2 flex flex-wrap gap-2 text-xs">
								{#if accepted > 0}
									<span class="px-2 py-0.5 bg-green-100 text-green-700 rounded-full">Принято: {accepted}</span>
								{/if}
								{#if declined > 0}
									<span class="px-2 py-0.5 bg-red-100 text-red-700 rounded-full">Отклонено: {declined}</span>
								{/if}
								{#if tentative > 0}
									<span class="px-2 py-0.5 bg-yellow-100 text-yellow-700 rounded-full">Возможно: {tentative}</span>
								{/if}
								{#if pending > 0}
									<span class="px-2 py-0.5 bg-gray-100 text-gray-600 rounded-full">Ожидает: {pending}</span>
								{/if}
							</div>
						{/if}
					</div>
				{/if}

				<!-- Recurring indicator -->
				{#if selectedEvent.is_recurring}
					<div class="flex items-center gap-3 text-gray-600">
						<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
						</svg>
						<span class="text-sm">Повторяющееся событие</span>
					</div>
				{/if}

				<!-- Cancelled indicator -->
				{#if selectedEvent.is_cancelled}
					<div class="flex items-center gap-3 text-red-600 bg-red-50 rounded-lg p-2 -mx-2">
						<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
						</svg>
						<span class="text-sm font-medium">Событие отменено</span>
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
								disabled={loadingRooms}
							>
								<option value="">{loadingRooms ? 'Загрузка переговорных...' : 'Выберите переговорную'}</option>
								{#each meetingRooms as room}
									<option value={room.email}>{room.name}{room.capacity ? ` (до ${room.capacity} чел.)` : ''}</option>
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

				<!-- Recurrence -->
				<div>
					<div class="flex items-center gap-3 mb-2">
						<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
						</svg>
						<span class="text-sm text-gray-700">Повторение</span>
					</div>
					<div class="ml-8 space-y-3">
						<!-- Recurrence type buttons -->
						<div class="flex flex-wrap gap-2">
							{#each recurrenceOptions as option}
								<button
									onclick={() => newEvent.recurrence = option.id as typeof newEvent.recurrence}
									class="px-3 py-1.5 text-sm rounded-lg border transition-colors
										{newEvent.recurrence === option.id ? 'border-ekf-red bg-red-50 text-ekf-red' : 'border-gray-200 hover:bg-gray-50'}"
								>
									{option.name}
								</button>
							{/each}
						</div>

						<!-- Extended recurrence options (only if recurrence is set) -->
						{#if newEvent.recurrence !== 'none'}
							<div class="bg-gray-50 rounded-lg p-3 space-y-4">
								<!-- Interval setting -->
								<div class="flex items-center gap-2 text-sm">
									<span class="text-gray-700">Повторять каждые</span>
									<input
										type="number"
										bind:value={newEvent.recurrence_interval}
										min="1"
										max="99"
										class="w-14 px-2 py-1 border border-gray-200 rounded text-sm text-center focus:outline-none focus:border-ekf-red"
									/>
									<span class="text-gray-600">
										{#if newEvent.recurrence === 'daily'}
											{newEvent.recurrence_interval === 1 ? 'день' : newEvent.recurrence_interval < 5 ? 'дня' : 'дней'}
										{:else if newEvent.recurrence === 'weekly'}
											{newEvent.recurrence_interval === 1 ? 'неделю' : newEvent.recurrence_interval < 5 ? 'недели' : 'недель'}
										{:else if newEvent.recurrence === 'monthly'}
											{newEvent.recurrence_interval === 1 ? 'месяц' : newEvent.recurrence_interval < 5 ? 'месяца' : 'месяцев'}
										{:else if newEvent.recurrence === 'yearly'}
											{newEvent.recurrence_interval === 1 ? 'год' : newEvent.recurrence_interval < 5 ? 'года' : 'лет'}
										{/if}
									</span>
								</div>

								<!-- Weekly: day of week selection -->
								{#if newEvent.recurrence === 'weekly'}
									<div class="space-y-2">
										<div class="text-xs text-gray-500 font-medium">Дни недели:</div>
										<div class="flex gap-1">
											{#each weekdayButtons as day}
												<button
													onclick={() => toggleWeekday(day.id)}
													class="w-8 h-8 text-xs rounded-lg border transition-colors
														{newEvent.recurrence_weekdays.includes(day.id)
															? 'border-ekf-red bg-red-50 text-ekf-red font-medium'
															: 'border-gray-200 hover:bg-gray-100 text-gray-600'}"
													title={day.full}
												>
													{day.short}
												</button>
											{/each}
										</div>
										<!-- Quick select buttons -->
										<div class="flex gap-2 mt-1">
											<button
												onclick={() => newEvent.recurrence_weekdays = [0, 1, 2, 3, 4]}
												class="text-xs text-blue-600 hover:text-blue-800"
											>
												Будни
											</button>
											<button
												onclick={() => newEvent.recurrence_weekdays = [5, 6]}
												class="text-xs text-blue-600 hover:text-blue-800"
											>
												Выходные
											</button>
											<button
												onclick={() => newEvent.recurrence_weekdays = [0, 1, 2, 3, 4, 5, 6]}
												class="text-xs text-blue-600 hover:text-blue-800"
											>
												Все дни
											</button>
										</div>
									</div>
								{/if}

								<!-- Monthly: type selection -->
								{#if newEvent.recurrence === 'monthly' && newEvent.date}
									{@const selectedDate = new Date(newEvent.date + 'T00:00:00')}
									{@const ordinalInfo = getWeekdayOrdinal(selectedDate)}
									<div class="space-y-2">
										<div class="text-xs text-gray-500 font-medium">Тип повторения:</div>
										<div class="space-y-2">
											<label class="flex items-center gap-2 text-sm cursor-pointer">
												<input
													type="radio"
													name="monthly_type"
													checked={newEvent.recurrence_monthly_type === 'day'}
													onchange={() => newEvent.recurrence_monthly_type = 'day'}
													class="text-ekf-red focus:ring-ekf-red"
												/>
												<span class="text-gray-700">
													Каждое {selectedDate.getDate()}-е число месяца
												</span>
											</label>
											<label class="flex items-center gap-2 text-sm cursor-pointer">
												<input
													type="radio"
													name="monthly_type"
													checked={newEvent.recurrence_monthly_type === 'weekday'}
													onchange={() => newEvent.recurrence_monthly_type = 'weekday'}
													class="text-ekf-red focus:ring-ekf-red"
												/>
												<span class="text-gray-700">
													Каждый {getOrdinalText(ordinalInfo.ordinal)} {ordinalInfo.weekdayName.toLowerCase()} месяца
												</span>
											</label>
										</div>
									</div>
								{/if}

								<div class="border-t border-gray-200 pt-3">
									<div class="text-xs text-gray-500 font-medium mb-2">Завершить повторение:</div>
									<div class="space-y-2">
										<label class="flex items-center gap-2 text-sm cursor-pointer">
											<input
												type="radio"
												name="recurrence_end_type"
												checked={!newEvent.recurrence_end && newEvent.recurrence_count > 0}
												onchange={() => { newEvent.recurrence_end = ''; newEvent.recurrence_count = 10; }}
												class="text-ekf-red focus:ring-ekf-red"
											/>
											<span class="text-gray-700">После</span>
											<input
												type="number"
												bind:value={newEvent.recurrence_count}
												min="1"
												max="365"
												class="w-16 px-2 py-1 border border-gray-200 rounded text-sm focus:outline-none focus:border-ekf-red"
											/>
											<span class="text-gray-600">повторений</span>
										</label>
										<label class="flex items-center gap-2 text-sm cursor-pointer">
											<input
												type="radio"
												name="recurrence_end_type"
												checked={!!newEvent.recurrence_end}
												onchange={() => {
													const d = new Date();
													d.setMonth(d.getMonth() + 1);
													newEvent.recurrence_end = d.toISOString().split('T')[0];
													newEvent.recurrence_count = 0;
												}}
												class="text-ekf-red focus:ring-ekf-red"
											/>
											<span class="text-gray-700">До даты</span>
											<input
												type="date"
												bind:value={newEvent.recurrence_end}
												min={newEvent.date}
												class="px-2 py-1 border border-gray-200 rounded text-sm focus:outline-none focus:border-ekf-red"
											/>
										</label>
									</div>
								</div>

								<!-- Recurrence summary -->
								<div class="text-xs text-gray-500 mt-2 p-2 bg-white rounded border border-gray-200">
									{#if newEvent.recurrence === 'daily'}
										Событие будет повторяться {newEvent.recurrence_interval === 1 ? 'каждый день' : `каждые ${newEvent.recurrence_interval} ${newEvent.recurrence_interval < 5 ? 'дня' : 'дней'}`}
									{:else if newEvent.recurrence === 'weekly'}
										Событие будет повторяться {newEvent.recurrence_interval === 1 ? 'каждую неделю' : `каждые ${newEvent.recurrence_interval} ${newEvent.recurrence_interval < 5 ? 'недели' : 'недель'}`}
										{#if newEvent.recurrence_weekdays.length > 0}
											<span class="font-medium"> ({getSelectedWeekdaysText()})</span>
										{/if}
									{:else if newEvent.recurrence === 'monthly' && newEvent.date}
										{@const selectedDate = new Date(newEvent.date + 'T00:00:00')}
										{@const ordinalInfo = getWeekdayOrdinal(selectedDate)}
										Событие будет повторяться {newEvent.recurrence_interval === 1 ? 'каждый месяц' : `каждые ${newEvent.recurrence_interval} ${newEvent.recurrence_interval < 5 ? 'месяца' : 'месяцев'}`}
										{#if newEvent.recurrence_monthly_type === 'day'}
											<span class="font-medium">{selectedDate.getDate()}-го числа</span>
										{:else}
											<span class="font-medium">в {getOrdinalText(ordinalInfo.ordinal)} {ordinalInfo.weekdayName.toLowerCase()}</span>
										{/if}
									{:else if newEvent.recurrence === 'yearly' && newEvent.date}
										Событие будет повторяться {newEvent.recurrence_interval === 1 ? 'каждый год' : `каждые ${newEvent.recurrence_interval} ${newEvent.recurrence_interval < 5 ? 'года' : 'лет'}`}
										<span class="font-medium">{new Date(newEvent.date + 'T00:00:00').toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' })}</span>
									{/if}
									{#if newEvent.recurrence_end}
										до {new Date(newEvent.recurrence_end + 'T00:00:00').toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })}
									{:else if newEvent.recurrence_count > 0}
										({newEvent.recurrence_count} {newEvent.recurrence_count === 1 ? 'повторение' : newEvent.recurrence_count < 5 ? 'повторения' : 'повторений'})
									{/if}
								</div>
							</div>
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

<!-- Event Details Modal (opens on double-click) -->
{#if showEventModal && modalEvent}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
		<div class="bg-white rounded-xl shadow-xl w-full max-w-lg max-h-[90vh] overflow-auto">
			<div class="p-6">
				<div class="flex items-start justify-between mb-4">
					{#if isEditingEvent}
						<input
							type="text"
							bind:value={editingEvent.subject}
							placeholder="Название события"
							class="flex-1 text-xl font-semibold text-gray-900 border-b-2 border-blue-500 outline-none bg-transparent"
						/>
					{:else}
						<h2 class="text-xl font-semibold text-gray-900">{modalEvent.subject || modalEvent.title}</h2>
					{/if}
					<button
						onclick={() => { showEventModal = false; modalEvent = null; isEditingEvent = false; }}
						class="p-1 hover:bg-gray-100 rounded ml-2"
					>
						<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>

				<div class="space-y-4">
					<!-- Time -->
					<div class="flex items-center gap-3 text-gray-600">
						<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						{#if isEditingEvent}
							<div class="flex gap-2 items-center">
								<input
									type="datetime-local"
									bind:value={editingEvent.start}
									class="text-sm border rounded px-2 py-1"
								/>
								<span class="text-gray-400">—</span>
								<input
									type="datetime-local"
									bind:value={editingEvent.end}
									class="text-sm border rounded px-2 py-1"
								/>
							</div>
						{:else}
							<div>
								<div class="text-sm font-medium text-gray-900">
									{new Date(modalEvent.start || modalEvent.start_time || '').toLocaleDateString('ru-RU', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
								</div>
								<div class="text-sm">{formatEventTime(modalEvent)}</div>
							</div>
						{/if}
					</div>

					<!-- Location -->
					{#if isEditingEvent}
						<div class="flex items-center gap-3 text-gray-600">
							<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
							</svg>
							<input
								type="text"
								bind:value={editingEvent.location}
								placeholder="Место проведения"
								class="flex-1 text-sm border rounded px-2 py-1"
							/>
						</div>
					{:else if modalEvent.location}
						<div class="flex items-center gap-3 text-gray-600">
							<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
							</svg>
							<span class="text-sm">{modalEvent.location}</span>
						</div>
					{/if}

					<!-- Organizer -->
					{#if modalEvent.organizer}
						<div class="flex items-center gap-3 text-gray-600">
							<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
							</svg>
							<div class="text-sm">
								<span class="text-gray-500">Организатор:</span>
								<span class="text-gray-900 font-medium">{modalEvent.organizer.name || modalEvent.organizer.email}</span>
							</div>
						</div>
					{/if}

					<!-- Attendees -->
					{#if modalEvent.attendees && modalEvent.attendees.length > 0}
						<div>
							<div class="flex items-center gap-3 text-gray-600 mb-3">
								<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
								</svg>
								<span class="text-sm font-medium">Участники ({modalEvent.attendees.length})</span>
							</div>
							<div class="ml-8 max-h-60 overflow-y-auto space-y-2">
								{#each modalEvent.attendees as attendee}
									<div class="flex items-center gap-2 text-sm">
										{#if attendee.response === 'Accept'}
											<span class="w-5 h-5 flex-shrink-0 rounded-full bg-green-100 text-green-600 flex items-center justify-center" title="Принял">
												<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
												</svg>
											</span>
										{:else if attendee.response === 'Decline'}
											<span class="w-5 h-5 flex-shrink-0 rounded-full bg-red-100 text-red-600 flex items-center justify-center" title="Отклонил">
												<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M6 18L18 6M6 6l12 12" />
												</svg>
											</span>
										{:else if attendee.response === 'Tentative'}
											<span class="w-5 h-5 flex-shrink-0 rounded-full bg-yellow-100 text-yellow-600 flex items-center justify-center" title="Возможно">
												<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 4h.01" />
												</svg>
											</span>
										{:else}
											<span class="w-5 h-5 flex-shrink-0 rounded-full bg-gray-100 text-gray-400 flex items-center justify-center" title="Ожидает ответа">
												<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01" />
												</svg>
											</span>
										{/if}
										<span class="text-gray-700" title={attendee.email}>
											{attendee.name || attendee.email}
										</span>
									</div>
								{/each}
							</div>
							<!-- Summary -->
							{#if modalEvent.attendees.length > 0}
								{@const accepted = modalEvent.attendees.filter(a => a.response === 'Accept').length}
								{@const declined = modalEvent.attendees.filter(a => a.response === 'Decline').length}
								{@const tentative = modalEvent.attendees.filter(a => a.response === 'Tentative').length}
								{@const pending = modalEvent.attendees.length - accepted - declined - tentative}
								<div class="ml-8 mt-3 flex flex-wrap gap-2 text-xs">
									{#if accepted > 0}
										<span class="px-2 py-1 bg-green-100 text-green-700 rounded-full">Принято: {accepted}</span>
									{/if}
									{#if declined > 0}
										<span class="px-2 py-1 bg-red-100 text-red-700 rounded-full">Отклонено: {declined}</span>
									{/if}
									{#if tentative > 0}
										<span class="px-2 py-1 bg-yellow-100 text-yellow-700 rounded-full">Возможно: {tentative}</span>
									{/if}
									{#if pending > 0}
										<span class="px-2 py-1 bg-gray-100 text-gray-600 rounded-full">Ожидает: {pending}</span>
									{/if}
								</div>
							{/if}
						</div>
					{/if}

					<!-- Recurring indicator -->
					{#if modalEvent.is_recurring}
						<div class="flex items-center gap-3 text-gray-600">
							<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
							</svg>
							<span class="text-sm">Повторяющееся событие</span>
						</div>
					{/if}

					<!-- Cancelled indicator -->
					{#if modalEvent.is_cancelled}
						<div class="flex items-center gap-3 text-red-600 bg-red-50 rounded-lg p-3">
							<svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
							</svg>
							<span class="text-sm font-medium">Событие отменено</span>
						</div>
					{/if}

					<!-- Transcription Section -->
					<div class="border-t pt-4 mt-4">
						<div class="flex items-center justify-between mb-3">
							<div class="flex items-center gap-2 text-gray-700">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
								</svg>
								<span class="text-sm font-medium">Транскрибирование</span>
							</div>
							<div class="flex items-center gap-2">
								{#if isRecording}
									<span class="text-red-500 text-sm font-medium animate-pulse">
										{formatRecordingTime(recordingTime)}
									</span>
									<button
										onclick={stopRecording}
										class="px-3 py-1.5 bg-red-500 text-white rounded-lg hover:bg-red-600 text-sm flex items-center gap-1"
									>
										<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
											<rect x="6" y="6" width="12" height="12" rx="2" />
										</svg>
										Стоп
									</button>
								{:else}
									<button
										onclick={startRecording}
										class="px-3 py-1.5 bg-blue-500 text-white rounded-lg hover:bg-blue-600 text-sm flex items-center gap-1"
									>
										<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
											<circle cx="12" cy="12" r="6" />
										</svg>
										Записать
									</button>
									<label class="px-3 py-1.5 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 text-sm cursor-pointer flex items-center gap-1">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
										</svg>
										Загрузить
										<input type="file" accept="audio/*" class="hidden" onchange={handleAudioFileUpload} />
									</label>
								{/if}
							</div>
						</div>

						{#if showTranscriptSection}
							<div class="bg-gray-50 rounded-lg p-3">
								{#if transcribing}
									<div class="flex items-center gap-2 text-gray-500">
										<svg class="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
											<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
											<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
										</svg>
										<span class="text-sm">Транскрибирование...</span>
									</div>
								{:else if transcript}
									<div class="text-sm text-gray-700 whitespace-pre-wrap max-h-40 overflow-y-auto">
										{transcript}
									</div>
								{:else if isRecording}
									<div class="flex items-center gap-2 text-red-500">
										<span class="w-2 h-2 bg-red-500 rounded-full animate-pulse"></span>
										<span class="text-sm">Идет запись...</span>
									</div>
								{:else}
									<p class="text-sm text-gray-400 italic">Нажмите "Записать" для начала записи или загрузите аудиофайл</p>
								{/if}
							</div>
						{/if}
					</div>
				</div>

				<div class="mt-6 flex justify-between">
					<div class="flex gap-2">
						{#if !isEditingEvent}
							<button
								onclick={startEditEvent}
								class="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 text-sm font-medium flex items-center gap-2"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
								</svg>
								Редактировать
							</button>
							<button
								onclick={() => deleteEvent(true)}
								disabled={deletingEvent}
								class="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 text-sm font-medium flex items-center gap-2 disabled:opacity-50"
							>
								{#if deletingEvent}
									<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
										<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
										<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
									</svg>
								{:else}
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
									</svg>
								{/if}
								Удалить
							</button>
						{:else}
							<button
								onclick={saveEditEvent}
								disabled={updatingEvent}
								class="px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 text-sm font-medium flex items-center gap-2 disabled:opacity-50"
							>
								{#if updatingEvent}
									<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
										<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
										<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
									</svg>
								{:else}
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
								{/if}
								Сохранить
							</button>
							<button
								onclick={cancelEditEvent}
								class="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 text-sm font-medium"
							>
								Отмена
							</button>
						{/if}
					</div>
					<button
						onclick={() => { showEventModal = false; modalEvent = null; isEditingEvent = false; }}
						class="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 text-sm font-medium"
					>
						Закрыть
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
