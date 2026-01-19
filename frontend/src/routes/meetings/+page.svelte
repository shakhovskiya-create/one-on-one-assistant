<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { meetings as meetingsApi, projects as projectsApi } from '$lib/api/client';
	import { subordinates } from '$lib/stores/auth';
	import type { Meeting, MeetingCategory, Project } from '$lib/api/client';

	// List state
	let meetings: Meeting[] = $state([]);
	let categories: MeetingCategory[] = $state([]);
	let projectsList: Project[] = $state([]);
	let loading = $state(true);
	let filter = $state('all');
	let searchQuery = $state('');
	let viewMode: 'list' | 'grid' = $state('list');

	// Upload modal state
	let showUploadModal = $state(false);
	let selectedCategory = $state('one_on_one');
	let selectedEmployee = $state('');
	let selectedProject = $state('');
	let selectedParticipants: string[] = $state([]);
	let meetingTitle = $state('');
	let meetingDate = $state(new Date().toISOString().split('T')[0]);
	let file: File | null = $state(null);
	let dragActive = $state(false);
	type ProcessingStatus = 'idle' | 'uploading' | 'transcribing' | 'analyzing' | 'done' | 'error';
	let uploadStatus: ProcessingStatus = $state('idle');
	let uploadProgress = $state(0);
	let uploadError = $state('');
	let uploadResult: any = $state(null);

	// Script drawer state
	let showScriptDrawer = $state(false);
	interface Question { text: string; checked: boolean; notes: string; }
	interface Section { id: string; title: string; duration: number; questions: Question[]; expanded: boolean; }
	let scriptSections: Section[] = $state([]);
	let scriptEmployee = $state('');
	let currentSection = $state(0);
	let timer = $state(0);
	let totalTime = $state(0);
	let isRunning = $state(false);
	let timerInterval: ReturnType<typeof setInterval> | null = null;

	// Computed
	const isOneOnOne = $derived(selectedCategory === 'one_on_one');
	const needsProject = $derived(['team_meeting', 'planning', 'retro', 'kickoff', 'status', 'demo'].includes(selectedCategory));
	const needsParticipants = $derived(['team_meeting', 'planning', 'retro', 'kickoff', 'status', 'demo'].includes(selectedCategory));
	const canSubmit = $derived(() => {
		if (!file) return false;
		if (isOneOnOne && !selectedEmployee) return false;
		if (needsProject && !selectedProject) return false;
		return true;
	});
	const currentSectionData = $derived(scriptSections[currentSection]);
	const sectionTimeLimit = $derived((currentSectionData?.duration || 0) * 60);
	const isOverTime = $derived(timer > sectionTimeLimit);

	const filteredMeetings = $derived(() => {
		let result = meetings;

		// Filter by category
		if (filter !== 'all') {
			result = result.filter(m => {
				const catCode = m.meeting_categories?.code || m.category;
				return catCode === filter;
			});
		}

		// Filter by search
		if (searchQuery.trim()) {
			const query = searchQuery.toLowerCase();
			result = result.filter(m =>
				m.title?.toLowerCase().includes(query) ||
				m.summary?.toLowerCase().includes(query)
			);
		}

		return result;
	});

	onMount(async () => {
		try {
			const [meetingsData, categoriesData, projectsData] = await Promise.all([
				meetingsApi.list(),
				meetingsApi.getCategories(),
				projectsApi.list('active').catch(() => [])
			]);
			meetings = meetingsData || [];
			categories = categoriesData || [];
			projectsList = projectsData || [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
		scriptSections = getDefaultScript();
	});

	onDestroy(() => {
		if (timerInterval) clearInterval(timerInterval);
	});

	$effect(() => {
		if (isRunning && !timerInterval) {
			timerInterval = setInterval(() => { timer++; totalTime++; }, 1000);
		} else if (!isRunning && timerInterval) {
			clearInterval(timerInterval);
			timerInterval = null;
		}
	});

	function getCategoryName(meeting: Meeting): string {
		if (meeting.meeting_categories?.name) return meeting.meeting_categories.name;
		const code = meeting.category || '';
		const cat = categories.find(c => c.code === code);
		if (cat) return cat.name;
		switch (code) {
			case 'one_on_one': return '1-на-1';
			case 'project': return 'Проект';
			case 'team': return 'Команда';
			default: return code;
		}
	}

	function getCategoryColor(code: string): string {
		switch (code) {
			case 'one_on_one': return 'bg-blue-100 text-blue-700';
			case 'team_meeting': return 'bg-purple-100 text-purple-700';
			case 'planning': return 'bg-green-100 text-green-700';
			case 'retro': return 'bg-yellow-100 text-yellow-700';
			case 'status': return 'bg-indigo-100 text-indigo-700';
			default: return 'bg-gray-100 text-gray-700';
		}
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function formatRelativeDate(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const days = Math.floor(diff / (1000 * 60 * 60 * 24));

		if (days === 0) return 'Сегодня';
		if (days === 1) return 'Вчера';
		if (days < 7) return `${days} дн. назад`;
		return formatDate(dateStr);
	}

	// Upload handlers
	function handleDragEnter(e: DragEvent) { e.preventDefault(); e.stopPropagation(); dragActive = true; }
	function handleDragLeave(e: DragEvent) { e.preventDefault(); e.stopPropagation(); dragActive = false; }
	function handleDragOver(e: DragEvent) { e.preventDefault(); e.stopPropagation(); }
	function handleDrop(e: DragEvent) {
		e.preventDefault(); e.stopPropagation(); dragActive = false;
		if (e.dataTransfer?.files && e.dataTransfer.files[0]) file = e.dataTransfer.files[0];
	}
	function handleFileChange(e: Event) {
		const target = e.target as HTMLInputElement;
		if (target.files && target.files[0]) file = target.files[0];
	}
	function toggleParticipant(id: string) {
		if (selectedParticipants.includes(id)) {
			selectedParticipants = selectedParticipants.filter(p => p !== id);
		} else {
			selectedParticipants = [...selectedParticipants, id];
		}
	}

	async function processFile() {
		if (!canSubmit()) { uploadError = 'Заполните обязательные поля'; return; }
		uploadError = '';
		uploadStatus = 'uploading';
		uploadProgress = 10;
		try {
			const formData = new FormData();
			formData.append('file', file!);
			formData.append('category_code', selectedCategory);
			formData.append('meeting_date', meetingDate);
			if (meetingTitle) formData.append('title', meetingTitle);
			if (isOneOnOne && selectedEmployee) formData.append('employee_id', selectedEmployee);
			if (needsProject && selectedProject) formData.append('project_id', selectedProject);
			if (needsParticipants && selectedParticipants.length > 0) {
				formData.append('participant_ids', JSON.stringify(selectedParticipants));
			}
			uploadStatus = 'transcribing';
			uploadProgress = 30;
			const data = await meetingsApi.process(formData);
			uploadStatus = 'analyzing';
			uploadProgress = 90;
			if (data.error) throw new Error(data.error);
			uploadResult = data;
			uploadStatus = 'done';
			uploadProgress = 100;
			meetings = await meetingsApi.list();
		} catch (err) {
			uploadError = err instanceof Error ? err.message : 'Произошла ошибка';
			uploadStatus = 'error';
		}
	}

	function resetUpload() {
		file = null;
		uploadStatus = 'idle';
		uploadProgress = 0;
		uploadResult = null;
		uploadError = '';
		selectedCategory = 'one_on_one';
		selectedEmployee = '';
		selectedProject = '';
		selectedParticipants = [];
		meetingTitle = '';
		meetingDate = new Date().toISOString().split('T')[0];
	}

	function closeUploadModal() {
		if (uploadStatus === 'idle' || uploadStatus === 'done' || uploadStatus === 'error') {
			showUploadModal = false;
			resetUpload();
		}
	}

	function getUploadStatusText(): string {
		switch (uploadStatus) {
			case 'uploading': return 'Загрузка файла...';
			case 'transcribing': return 'Транскрибирование...';
			case 'analyzing': return 'AI-анализ встречи...';
			case 'done': return 'Готово!';
			case 'error': return 'Ошибка';
			default: return '';
		}
	}

	// Script handlers
	function getDefaultScript(): Section[] {
		return [
			{ id: 'checkin', title: 'Чекин', duration: 5, expanded: true, questions: [
				{ text: 'Как ты? Что нового?', checked: false, notes: '' },
				{ text: 'Как прошла неделя?', checked: false, notes: '' },
				{ text: 'Что занимает голову прямо сейчас?', checked: false, notes: '' },
			]},
			{ id: 'employee_agenda', title: 'Повестка сотрудника', duration: 20, expanded: false, questions: [
				{ text: 'С чем пришел? Что хочешь обсудить?', checked: false, notes: '' },
				{ text: 'Где нужна помощь или ресурс?', checked: false, notes: '' },
				{ text: 'Что буксует и почему?', checked: false, notes: '' },
			]},
			{ id: 'manager_agenda', title: 'Повестка руководителя', duration: 15, expanded: false, questions: [
				{ text: 'Статус по ключевым проектам', checked: false, notes: '' },
				{ text: 'Изменения в приоритетах', checked: false, notes: '' },
				{ text: 'Ожидания и сроки', checked: false, notes: '' },
			]},
			{ id: 'development', title: 'Развитие сотрудника', duration: 10, expanded: false, questions: [
				{ text: 'Как оцениваешь свою работу?', checked: false, notes: '' },
				{ text: 'Что получилось хорошо?', checked: false, notes: '' },
				{ text: 'Чему хочешь научиться?', checked: false, notes: '' },
			]},
			{ id: 'agreements', title: 'Договоренности', duration: 5, expanded: false, questions: [
				{ text: 'Фиксируем договоренности и сроки', checked: false, notes: '' },
			]},
		];
	}

	function formatTime(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
	}

	function toggleScriptSection(index: number) {
		scriptSections = scriptSections.map((s, i) => ({ ...s, expanded: i === index ? !s.expanded : s.expanded }));
		currentSection = index;
		timer = 0;
	}

	function toggleQuestion(sectionIndex: number, questionIndex: number) {
		scriptSections = scriptSections.map((s, si) =>
			si === sectionIndex
				? { ...s, questions: s.questions.map((q, qi) => qi === questionIndex ? { ...q, checked: !q.checked } : q) }
				: s
		);
	}

	function updateNotes(sectionIndex: number, questionIndex: number, notes: string) {
		scriptSections = scriptSections.map((s, si) =>
			si === sectionIndex
				? { ...s, questions: s.questions.map((q, qi) => qi === questionIndex ? { ...q, notes } : q) }
				: s
		);
	}

	function nextSection() {
		if (currentSection < scriptSections.length - 1) {
			scriptSections = scriptSections.map((s, i) => ({ ...s, expanded: i === currentSection + 1 }));
			currentSection++;
			timer = 0;
		}
	}

	function resetScript() {
		scriptSections = getDefaultScript();
		currentSection = 0;
		timer = 0;
		totalTime = 0;
		isRunning = false;
	}

	function getSectionProgress(section: Section): number {
		const checked = section.questions.filter(q => q.checked).length;
		return Math.round((checked / section.questions.length) * 100);
	}

	function getMoodColor(score: number): string {
		if (score >= 7) return 'text-green-600 bg-green-50';
		if (score >= 5) return 'text-yellow-600 bg-yellow-50';
		return 'text-red-600 bg-red-50';
	}
</script>

<svelte:head>
	<title>Встречи - EKF Hub</title>
</svelte:head>

<div class="space-y-4">
	<!-- Header -->
	<div class="flex items-center justify-between gap-4">
		<h1 class="text-xl font-bold text-gray-900">Встречи</h1>
		<div class="flex items-center gap-2">
			<button
				onclick={() => showScriptDrawer = !showScriptDrawer}
				class="px-3 py-2 text-sm font-medium rounded-lg transition-colors flex items-center gap-2
					{showScriptDrawer ? 'bg-ekf-red text-white' : 'bg-white text-gray-700 border hover:bg-gray-50'}"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
				</svg>
				Скрипт 1-на-1
			</button>
			<button
				onclick={() => showUploadModal = true}
				class="px-4 py-2 bg-ekf-red text-white text-sm font-medium rounded-lg hover:bg-red-700 transition-colors flex items-center gap-2"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
				</svg>
				Загрузить запись
			</button>
		</div>
	</div>

	<!-- Filters Row -->
	<div class="flex items-center gap-3 flex-wrap">
		<!-- Search -->
		<div class="relative flex-1 min-w-[200px] max-w-md">
			<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
			</svg>
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Поиск по встречам..."
				class="w-full pl-9 pr-4 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-ekf-red focus:border-transparent"
			/>
		</div>

		<!-- Category Filter Pills -->
		<div class="flex items-center gap-1.5 flex-wrap">
			<button
				onclick={() => filter = 'all'}
				class="px-3 py-1.5 rounded-full text-xs font-medium transition-colors
					{filter === 'all' ? 'bg-ekf-red text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}"
			>
				Все
			</button>
			{#each categories as cat}
				<button
					onclick={() => filter = cat.code}
					class="px-3 py-1.5 rounded-full text-xs font-medium transition-colors
						{filter === cat.code ? 'bg-ekf-red text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}"
				>
					{cat.name}
				</button>
			{/each}
		</div>

		<!-- View Toggle -->
		<div class="flex items-center gap-1 bg-gray-100 rounded-lg p-0.5">
			<button
				onclick={() => viewMode = 'list'}
				class="p-1.5 rounded-md transition-colors {viewMode === 'list' ? 'bg-white shadow-sm' : 'hover:bg-gray-200'}"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</button>
			<button
				onclick={() => viewMode = 'grid'}
				class="p-1.5 rounded-md transition-colors {viewMode === 'grid' ? 'bg-white shadow-sm' : 'hover:bg-gray-200'}"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
				</svg>
			</button>
		</div>
	</div>

	<!-- Content Area -->
	<div class="flex gap-4">
		<!-- Main Content -->
		<div class="flex-1 min-w-0">
			{#if loading}
				<div class="flex items-center justify-center h-48">
					<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
				</div>
			{:else if filteredMeetings().length === 0}
				<div class="bg-white rounded-xl border p-8 text-center">
					<div class="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
						<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
					</div>
					<p class="text-gray-600 mb-4">
						{searchQuery ? 'По вашему запросу ничего не найдено' : 'Встреч пока нет'}
					</p>
					<button
						onclick={() => showUploadModal = true}
						class="px-4 py-2 bg-ekf-red text-white text-sm rounded-lg hover:bg-red-700"
					>
						Загрузить первую запись
					</button>
				</div>
			{:else if viewMode === 'list'}
				<div class="space-y-2">
					{#each filteredMeetings() as meeting}
						<a href="/meetings/{meeting.id}" class="block bg-white rounded-lg border p-4 hover:shadow-md hover:border-ekf-red/30 transition-all group">
							<div class="flex items-start justify-between gap-4">
								<div class="flex-1 min-w-0">
									<div class="flex items-center gap-2 mb-1">
										{#if meeting.meeting_categories || meeting.category}
											<span class="px-2 py-0.5 text-xs rounded-full {getCategoryColor(meeting.meeting_categories?.code || meeting.category || '')}">
												{getCategoryName(meeting)}
											</span>
										{/if}
										<span class="text-xs text-gray-400">{formatRelativeDate(meeting.date)}</span>
									</div>
									<h3 class="font-medium text-gray-900 group-hover:text-ekf-red transition-colors truncate">
										{meeting.title || 'Без названия'}
									</h3>
									{#if meeting.summary}
										<p class="text-sm text-gray-500 mt-1 line-clamp-1">{meeting.summary}</p>
									{/if}
								</div>
								<div class="flex items-center gap-3 shrink-0">
									{#if meeting.duration_minutes}
										<span class="text-xs text-gray-400">{meeting.duration_minutes} мин</span>
									{/if}
									{#if meeting.mood_score}
										<span class="px-2 py-1 rounded text-sm font-semibold {getMoodColor(meeting.mood_score)}">
											{meeting.mood_score}
										</span>
									{/if}
									<svg class="w-4 h-4 text-gray-300 group-hover:text-ekf-red transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</div>
							</div>
						</a>
					{/each}
				</div>
			{:else}
				<div class="grid grid-cols-2 xl:grid-cols-3 gap-3">
					{#each filteredMeetings() as meeting}
						<a href="/meetings/{meeting.id}" class="bg-white rounded-lg border p-4 hover:shadow-md hover:border-ekf-red/30 transition-all group">
							<div class="flex items-center justify-between mb-2">
								{#if meeting.meeting_categories || meeting.category}
									<span class="px-2 py-0.5 text-xs rounded-full {getCategoryColor(meeting.meeting_categories?.code || meeting.category || '')}">
										{getCategoryName(meeting)}
									</span>
								{/if}
								{#if meeting.mood_score}
									<span class="px-2 py-1 rounded text-sm font-semibold {getMoodColor(meeting.mood_score)}">
										{meeting.mood_score}
									</span>
								{/if}
							</div>
							<h3 class="font-medium text-gray-900 group-hover:text-ekf-red transition-colors mb-1 line-clamp-2">
								{meeting.title || 'Без названия'}
							</h3>
							<div class="flex items-center gap-2 text-xs text-gray-400">
								<span>{formatDate(meeting.date)}</span>
								{#if meeting.duration_minutes}
									<span>•</span>
									<span>{meeting.duration_minutes} мин</span>
								{/if}
							</div>
							{#if meeting.summary}
								<p class="text-xs text-gray-500 mt-2 line-clamp-2">{meeting.summary}</p>
							{/if}
						</a>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Script Drawer -->
		{#if showScriptDrawer}
			<div class="w-80 shrink-0 bg-white rounded-xl border overflow-hidden flex flex-col max-h-[calc(100vh-180px)]">
				<div class="p-3 border-b bg-gray-50">
					<div class="flex items-center justify-between mb-2">
						<h3 class="font-semibold text-sm">Скрипт 1-на-1</h3>
						<button onclick={() => showScriptDrawer = false} class="p-1 hover:bg-gray-200 rounded">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
							</svg>
						</button>
					</div>
					<select bind:value={scriptEmployee}
						class="w-full border rounded px-2 py-1.5 text-sm focus:ring-1 focus:ring-ekf-red">
						<option value="">Выберите сотрудника</option>
						{#each $subordinates as emp}
							<option value={emp.id}>{emp.name}</option>
						{/each}
					</select>
				</div>

				<!-- Timer -->
				<div class="p-3 border-b bg-gradient-to-r from-gray-50 to-white">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-4">
							<div class="text-center">
								<p class="text-2xl font-mono font-bold {isOverTime ? 'text-red-600' : 'text-gray-900'}">{formatTime(timer)}</p>
								<p class="text-[10px] text-gray-400">секция</p>
							</div>
							<div class="text-center border-l pl-4">
								<p class="text-2xl font-mono font-bold text-gray-700">{formatTime(totalTime)}</p>
								<p class="text-[10px] text-gray-400">всего</p>
							</div>
						</div>
						<div class="flex items-center gap-1">
							<button onclick={() => isRunning = !isRunning}
								class="p-2 rounded-full transition-colors {isRunning ? 'bg-yellow-100 text-yellow-700' : 'bg-green-100 text-green-700'}">
								{#if isRunning}
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 9v6m4-6v6" />
									</svg>
								{:else}
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
									</svg>
								{/if}
							</button>
							<button onclick={resetScript} class="p-2 rounded-full bg-gray-100 text-gray-600 hover:bg-gray-200">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
								</svg>
							</button>
						</div>
					</div>
				</div>

				<!-- Sections -->
				<div class="flex-1 overflow-y-auto p-2 space-y-2">
					{#each scriptSections as section, sectionIndex}
						<div class="border rounded-lg overflow-hidden {sectionIndex === currentSection ? 'ring-1 ring-ekf-red' : ''}">
							<button onclick={() => toggleScriptSection(sectionIndex)}
								class="w-full p-2 flex items-center justify-between hover:bg-gray-50 text-left">
								<div class="flex items-center gap-2">
									<svg class="w-4 h-4 transform transition-transform {section.expanded ? 'rotate-90' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
									<span class="font-medium text-sm">{section.title}</span>
									<span class="text-[10px] text-gray-400">({section.duration} мин)</span>
								</div>
								<div class="flex items-center gap-2">
									<div class="w-12 h-1.5 bg-gray-200 rounded-full overflow-hidden">
										<div class="h-full bg-green-500 transition-all" style="width: {getSectionProgress(section)}%"></div>
									</div>
								</div>
							</button>
							{#if section.expanded}
								<div class="border-t p-2 space-y-2 bg-gray-50/50">
									{#each section.questions as question, questionIndex}
										<div class="space-y-1">
											<label class="flex items-start gap-2 cursor-pointer">
												<input type="checkbox" checked={question.checked}
													onchange={() => toggleQuestion(sectionIndex, questionIndex)}
													class="mt-0.5 h-4 w-4 rounded border-gray-300 text-ekf-red focus:ring-ekf-red" />
												<span class="text-xs {question.checked ? 'text-gray-400 line-through' : 'text-gray-700'}">{question.text}</span>
											</label>
											<textarea value={question.notes}
												oninput={(e) => updateNotes(sectionIndex, questionIndex, (e.target as HTMLTextAreaElement).value)}
												placeholder="Заметки..."
												class="w-full ml-6 p-1.5 text-xs border rounded resize-none focus:ring-1 focus:ring-ekf-red"
												rows="1"></textarea>
										</div>
									{/each}
									{#if sectionIndex < scriptSections.length - 1}
										<button onclick={nextSection}
											class="w-full py-1.5 bg-ekf-red text-white text-xs rounded hover:bg-red-700 flex items-center justify-center gap-1">
											Далее
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
											</svg>
										</button>
									{/if}
								</div>
							{/if}
						</div>
					{/each}
				</div>

				<!-- Progress -->
				<div class="p-2 border-t bg-gray-50">
					<div class="flex gap-1">
						{#each scriptSections as section, index}
							<div class="flex-1 h-1.5 rounded-full
								{getSectionProgress(section) === 100 ? 'bg-green-500' : index === currentSection ? 'bg-ekf-red' : 'bg-gray-200'}"></div>
						{/each}
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- Upload Modal -->
{#if showUploadModal}
	<div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" onclick={closeUploadModal}>
		<div class="bg-white rounded-xl max-w-lg w-full max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()}>
			<div class="p-4 border-b flex items-center justify-between">
				<h2 class="font-semibold">Загрузка записи встречи</h2>
				<button onclick={closeUploadModal} class="p-1 hover:bg-gray-100 rounded">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="p-4">
				{#if uploadStatus === 'idle'}
					<div class="space-y-4">
						<!-- Category Selection -->
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-2">Тип встречи</label>
							<div class="flex flex-wrap gap-1.5">
								{#each categories as cat}
									<button
										onclick={() => { selectedCategory = cat.code; selectedEmployee = ''; selectedProject = ''; selectedParticipants = []; }}
										class="px-3 py-1.5 rounded-full text-xs font-medium transition-colors
											{selectedCategory === cat.code ? 'bg-ekf-red text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
									>
										{cat.name}
									</button>
								{/each}
							</div>
						</div>

						<!-- Form Fields -->
						<div class="grid grid-cols-2 gap-3">
							<div class="col-span-2">
								<label class="block text-xs font-medium text-gray-700 mb-1">Название</label>
								<input type="text" bind:value={meetingTitle} placeholder="Название встречи"
									class="w-full border rounded px-3 py-2 text-sm focus:ring-1 focus:ring-ekf-red focus:border-transparent" />
							</div>
							{#if isOneOnOne}
								<div class="col-span-2">
									<label class="block text-xs font-medium text-gray-700 mb-1">Сотрудник *</label>
									<select bind:value={selectedEmployee}
										class="w-full border rounded px-3 py-2 text-sm focus:ring-1 focus:ring-ekf-red">
										<option value="">Выберите сотрудника</option>
										{#each $subordinates as emp}
											<option value={emp.id}>{emp.name} - {emp.position}</option>
										{/each}
									</select>
								</div>
							{/if}
							{#if needsProject}
								<div>
									<label class="block text-xs font-medium text-gray-700 mb-1">Проект *</label>
									<select bind:value={selectedProject}
										class="w-full border rounded px-3 py-2 text-sm focus:ring-1 focus:ring-ekf-red">
										<option value="">Выберите проект</option>
										{#each projectsList as proj}
											<option value={proj.id}>{proj.name}</option>
										{/each}
									</select>
								</div>
							{/if}
							<div class="{isOneOnOne || needsProject ? '' : 'col-span-2'}">
								<label class="block text-xs font-medium text-gray-700 mb-1">Дата</label>
								<input type="date" bind:value={meetingDate}
									class="w-full border rounded px-3 py-2 text-sm focus:ring-1 focus:ring-ekf-red" />
							</div>
						</div>

						{#if needsParticipants}
							<div>
								<label class="block text-xs font-medium text-gray-700 mb-2">Участники</label>
								<div class="flex flex-wrap gap-1.5">
									{#each $subordinates as emp}
										<button onclick={() => toggleParticipant(emp.id)}
											class="px-2 py-1 rounded-full text-xs transition-colors
												{selectedParticipants.includes(emp.id) ? 'bg-ekf-red/10 text-ekf-red border border-ekf-red' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}">
											{emp.name}
										</button>
									{/each}
								</div>
							</div>
						{/if}

						<!-- Drop Zone -->
						<div role="button" tabindex="0"
							ondragenter={handleDragEnter} ondragleave={handleDragLeave} ondragover={handleDragOver} ondrop={handleDrop}
							class="border-2 border-dashed rounded-lg p-6 text-center transition-colors
								{dragActive ? 'border-ekf-red bg-red-50' : file ? 'border-green-500 bg-green-50' : 'border-gray-300 hover:border-gray-400'}">
							{#if file}
								<div class="flex items-center justify-center gap-3">
									<svg class="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
									</svg>
									<div class="text-left">
										<p class="text-sm font-medium text-gray-900">{file.name}</p>
										<p class="text-xs text-gray-500">{(file.size / 1024 / 1024).toFixed(2)} MB</p>
									</div>
									<button onclick={() => file = null} class="text-xs text-red-600 hover:underline ml-2">Удалить</button>
								</div>
							{:else}
								<svg class="w-10 h-10 mx-auto text-gray-400 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
								</svg>
								<p class="text-sm text-gray-600">
									Перетащите файл или
									<label class="text-ekf-red hover:underline cursor-pointer">
										выберите
										<input type="file" accept="audio/*,video/*" onchange={handleFileChange} class="hidden" />
									</label>
								</p>
								<p class="text-xs text-gray-400 mt-1">MP3, WAV, MP4, WebM, M4A</p>
							{/if}
						</div>

						{#if uploadError}
							<div class="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-700">{uploadError}</div>
						{/if}

						<button onclick={processFile} disabled={!canSubmit()}
							class="w-full py-2.5 bg-ekf-red text-white rounded-lg hover:bg-red-700 disabled:bg-gray-300 disabled:cursor-not-allowed text-sm font-medium">
							Обработать запись
						</button>
					</div>
				{/if}

				{#if uploadStatus === 'uploading' || uploadStatus === 'transcribing' || uploadStatus === 'analyzing'}
					<div class="py-8 text-center space-y-4">
						<div class="animate-spin rounded-full h-10 w-10 border-b-2 border-ekf-red mx-auto"></div>
						<div>
							<p class="font-medium text-gray-900">{getUploadStatusText()}</p>
							<p class="text-xs text-gray-500 mt-1">Это может занять несколько минут</p>
						</div>
						<div class="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
							<div class="h-full bg-ekf-red transition-all duration-500" style="width: {uploadProgress}%"></div>
						</div>
					</div>
				{/if}

				{#if uploadStatus === 'done' && uploadResult}
					<div class="space-y-4">
						<div class="bg-green-50 border border-green-200 rounded-lg p-3 flex items-center gap-3">
							<svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
							<span class="text-green-700 text-sm font-medium">Анализ завершен!</span>
						</div>
						{#if uploadResult.analysis?.summary}
							<div class="bg-gray-50 rounded-lg p-3">
								<h4 class="text-xs font-medium text-gray-700 mb-1">Резюме</h4>
								<p class="text-sm text-gray-600">{uploadResult.analysis.summary}</p>
							</div>
						{/if}
						<div class="flex gap-2">
							<button onclick={closeUploadModal} class="flex-1 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 text-sm">
								Закрыть
							</button>
							<a href="/meetings/{uploadResult.meeting_id}" class="flex-1 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 text-sm text-center">
								Открыть встречу
							</a>
						</div>
					</div>
				{/if}

				{#if uploadStatus === 'error'}
					<div class="py-8 text-center space-y-4">
						<svg class="w-10 h-10 mx-auto text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<div>
							<p class="font-medium text-gray-900">Произошла ошибка</p>
							<p class="text-xs text-gray-500 mt-1">{uploadError}</p>
						</div>
						<button onclick={resetUpload} class="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 text-sm">
							Попробовать снова
						</button>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}
