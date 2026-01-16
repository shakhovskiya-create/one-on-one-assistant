<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { meetings as meetingsApi, projects as projectsApi } from '$lib/api/client';
	import { subordinates } from '$lib/stores/auth';
	import type { Meeting, MeetingCategory, Project } from '$lib/api/client';

	// Tab state
	let activeTab: 'list' | 'upload' | 'script' = $state('list');

	// List state
	let meetings: Meeting[] = $state([]);
	let categories: MeetingCategory[] = $state([]);
	let loading = $state(true);
	let filter = $state('all');

	// Upload state
	let projectsList: Project[] = $state([]);
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

	// Script state
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
		if (filter === 'all') return meetings;
		return meetings.filter(m => {
			const catCode = m.meeting_categories?.code || m.category;
			return catCode === filter;
		});
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

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });
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
			// Refresh meetings list
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
</script>

<svelte:head>
	<title>Встречи - EKF Team Hub</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-gray-900">Встречи</h1>
	</div>

	<!-- Tabs -->
	<div class="border-b border-gray-200">
		<nav class="flex gap-8">
			{#each [
				{ key: 'list', label: 'Все встречи', icon: 'list' },
				{ key: 'upload', label: 'Загрузить запись', icon: 'upload' },
				{ key: 'script', label: 'Скрипт 1-на-1', icon: 'script' }
			] as tab}
				<button
					onclick={() => activeTab = tab.key as typeof activeTab}
					class="py-4 px-1 border-b-2 font-medium text-sm transition-colors flex items-center gap-2
						{activeTab === tab.key
							? 'border-ekf-red text-ekf-red'
							: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
				>
					{#if tab.icon === 'list'}
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
						</svg>
					{:else if tab.icon === 'upload'}
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
						</svg>
					{:else}
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
						</svg>
					{/if}
					{tab.label}
				</button>
			{/each}
		</nav>
	</div>

	<!-- List Tab -->
	{#if activeTab === 'list'}
		<!-- Filters -->
		<div class="flex gap-2 flex-wrap">
			<button
				onclick={() => filter = 'all'}
				class="px-4 py-2 rounded-lg text-sm font-medium transition-colors
					{filter === 'all' ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-100'}"
			>
				Все
			</button>
			{#each categories as cat}
				<button
					onclick={() => filter = cat.code}
					class="px-4 py-2 rounded-lg text-sm font-medium transition-colors
						{filter === cat.code ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-100'}"
				>
					{cat.name}
				</button>
			{/each}
		</div>

		{#if loading}
			<div class="flex items-center justify-center h-64">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
			</div>
		{:else if filteredMeetings().length === 0}
			<div class="bg-white rounded-xl shadow-sm p-12 text-center">
				<div class="text-gray-400 text-lg mb-4">Встреч пока нет</div>
				<button
					onclick={() => activeTab = 'upload'}
					class="inline-block px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors"
				>
					Загрузить первую запись
				</button>
			</div>
		{:else}
			<div class="space-y-4">
				{#each filteredMeetings() as meeting}
					<a href="/meetings/{meeting.id}" class="block bg-white rounded-xl shadow-sm p-6 hover:shadow-md transition-shadow">
						<div class="flex items-start justify-between">
							<div class="flex-1">
								<div class="flex items-center gap-3">
									<h3 class="text-lg font-semibold text-gray-900">{meeting.title || 'Без названия'}</h3>
									{#if meeting.meeting_categories || meeting.category}
										<span class="px-2 py-1 text-xs rounded-full bg-gray-100 text-gray-600">{getCategoryName(meeting)}</span>
									{/if}
								</div>
								<p class="text-sm text-gray-500 mt-1">{formatDate(meeting.date)}</p>
								{#if meeting.summary}
									<p class="text-gray-600 mt-3 line-clamp-2">{meeting.summary}</p>
								{/if}
							</div>
							<div class="flex flex-col items-end gap-2">
								{#if meeting.mood_score}
									<span class="text-lg font-bold {meeting.mood_score >= 7 ? 'text-green-600' : meeting.mood_score >= 5 ? 'text-yellow-600' : 'text-red-600'}">
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
	{/if}

	<!-- Upload Tab -->
	{#if activeTab === 'upload'}
		<div class="max-w-4xl space-y-6">
			{#if uploadStatus === 'idle'}
				<!-- Category Selection -->
				<div class="bg-white rounded-lg shadow-sm border p-6">
					<label class="block text-sm font-medium text-gray-700 mb-3">Тип встречи</label>
					<div class="grid grid-cols-4 gap-2">
						{#each categories as cat}
							<button
								onclick={() => { selectedCategory = cat.code; selectedEmployee = ''; selectedProject = ''; selectedParticipants = []; }}
								class="px-4 py-2 rounded-lg text-sm font-medium transition-colors
									{selectedCategory === cat.code ? 'bg-ekf-red text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
							>
								{cat.name}
							</button>
						{/each}
					</div>
				</div>

				<!-- Dynamic Settings -->
				<div class="bg-white rounded-lg shadow-sm border p-6 space-y-4">
					<div class="grid grid-cols-2 gap-4">
						<div class="col-span-2">
							<label class="block text-sm font-medium text-gray-700 mb-1">Название (опционально)</label>
							<input type="text" bind:value={meetingTitle} placeholder="Название встречи"
								class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent" />
						</div>
						{#if isOneOnOne}
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Сотрудник *</label>
								<select bind:value={selectedEmployee}
									class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent">
									<option value="">Выберите сотрудника</option>
									{#each $subordinates as emp}
										<option value={emp.id}>{emp.name} - {emp.position}</option>
									{/each}
								</select>
							</div>
						{/if}
						{#if needsProject}
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Проект *</label>
								<select bind:value={selectedProject}
									class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent">
									<option value="">Выберите проект</option>
									{#each projectsList as proj}
										<option value={proj.id}>{proj.name}</option>
									{/each}
								</select>
							</div>
						{/if}
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Дата встречи</label>
							<input type="date" bind:value={meetingDate}
								class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent" />
						</div>
					</div>
					{#if needsParticipants}
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-2">Участники</label>
							<div class="flex flex-wrap gap-2">
								{#each $subordinates as emp}
									<button onclick={() => toggleParticipant(emp.id)}
										class="flex items-center gap-1 px-3 py-1 rounded-full text-sm transition-colors
											{selectedParticipants.includes(emp.id) ? 'bg-ekf-red/10 text-ekf-red border border-ekf-red' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}">
										{emp.name}
									</button>
								{/each}
							</div>
						</div>
					{/if}
				</div>

				<!-- Drop Zone -->
				<div role="button" tabindex="0"
					ondragenter={handleDragEnter} ondragleave={handleDragLeave} ondragover={handleDragOver} ondrop={handleDrop}
					class="bg-white rounded-lg shadow-sm border-2 border-dashed p-12 text-center transition-colors
						{dragActive ? 'border-ekf-red bg-orange-50' : file ? 'border-green-500 bg-green-50' : 'border-gray-300'}">
					{#if file}
						<div class="space-y-4">
							<svg class="w-12 h-12 mx-auto text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
							</svg>
							<div>
								<p class="font-medium text-gray-900">{file.name}</p>
								<p class="text-sm text-gray-500">{(file.size / 1024 / 1024).toFixed(2)} MB</p>
							</div>
							<button onclick={() => file = null} class="text-sm text-red-600 hover:underline">Удалить</button>
						</div>
					{:else}
						<div class="space-y-4">
							<svg class="w-12 h-12 mx-auto text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
							</svg>
							<div>
								<p class="font-medium text-gray-900">
									Перетащите файл сюда или
									<label class="text-ekf-red hover:underline cursor-pointer">
										выберите
										<input type="file" accept="audio/*,video/*" onchange={handleFileChange} class="hidden" />
									</label>
								</p>
								<p class="text-sm text-gray-500 mt-1">MP3, WAV, MP4, WebM, M4A, OGG</p>
							</div>
						</div>
					{/if}
				</div>

				{#if uploadError}
					<div class="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700">{uploadError}</div>
				{/if}

				<button onclick={processFile} disabled={!canSubmit()}
					class="w-full py-3 bg-ekf-red text-white rounded-lg hover:bg-red-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-medium transition-colors">
					Обработать запись
				</button>
			{/if}

			{#if uploadStatus === 'uploading' || uploadStatus === 'transcribing' || uploadStatus === 'analyzing'}
				<div class="bg-white rounded-lg shadow-sm border p-12 text-center space-y-6">
					<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-ekf-red mx-auto"></div>
					<div>
						<p class="font-medium text-gray-900">{getUploadStatusText()}</p>
						<p class="text-sm text-gray-500 mt-1">Это может занять несколько минут</p>
					</div>
					<div class="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
						<div class="h-full bg-ekf-red transition-all duration-500" style="width: {uploadProgress}%"></div>
					</div>
				</div>
			{/if}

			{#if uploadStatus === 'done' && uploadResult}
				<div class="space-y-6">
					<div class="bg-green-50 border border-green-200 rounded-lg p-4 flex items-center justify-between">
						<div class="flex items-center gap-3">
							<svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
							<span class="text-green-700 font-medium">Анализ завершен!</span>
						</div>
						<a href="/meetings/{uploadResult.id}" class="text-ekf-red hover:underline">Открыть встречу</a>
					</div>
					{#if uploadResult.analysis?.summary}
						<div class="bg-white rounded-lg shadow-sm border p-6">
							<h2 class="text-lg font-semibold mb-4">Резюме встречи</h2>
							<p class="text-gray-700">{uploadResult.analysis.summary}</p>
						</div>
					{/if}
					<button onclick={resetUpload} class="w-full py-3 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 font-medium">
						Загрузить еще одну запись
					</button>
				</div>
			{/if}

			{#if uploadStatus === 'error'}
				<div class="bg-white rounded-lg shadow-sm border p-12 text-center space-y-6">
					<svg class="w-12 h-12 mx-auto text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
					<div>
						<p class="font-medium text-gray-900">Произошла ошибка</p>
						<p class="text-sm text-gray-500 mt-1">{uploadError}</p>
					</div>
					<button onclick={resetUpload} class="px-6 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200">
						Попробовать снова
					</button>
				</div>
			{/if}
		</div>
	{/if}

	<!-- Script Tab -->
	{#if activeTab === 'script'}
		<div class="max-w-4xl space-y-6">
			<div class="flex justify-between items-center">
				<select bind:value={scriptEmployee}
					class="border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent">
					<option value="">Выберите сотрудника</option>
					{#each $subordinates as emp}
						<option value={emp.id}>{emp.name}</option>
					{/each}
				</select>
			</div>

			<!-- Timer Panel -->
			<div class="bg-white rounded-lg shadow-sm border p-4">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-6">
						<div class="text-center">
							<p class="text-sm text-gray-500">Секция</p>
							<p class="text-3xl font-mono font-bold {isOverTime ? 'text-red-600' : 'text-gray-900'}">{formatTime(timer)}</p>
							<p class="text-xs text-gray-400">из {currentSectionData?.duration || 0} мин</p>
						</div>
						<div class="text-center border-l pl-6">
							<p class="text-sm text-gray-500">Всего</p>
							<p class="text-3xl font-mono font-bold text-gray-900">{formatTime(totalTime)}</p>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<button onclick={() => isRunning = !isRunning}
							class="p-3 rounded-full transition-colors {isRunning ? 'bg-yellow-100 text-yellow-700 hover:bg-yellow-200' : 'bg-green-100 text-green-700 hover:bg-green-200'}">
							{#if isRunning}
								<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
								</svg>
							{:else}
								<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
								</svg>
							{/if}
						</button>
						<button onclick={resetScript} class="p-3 rounded-full bg-gray-100 text-gray-700 hover:bg-gray-200">
							<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
							</svg>
						</button>
					</div>
				</div>
			</div>

			<!-- Sections -->
			<div class="space-y-4">
				{#each scriptSections as section, sectionIndex}
					<div class="bg-white rounded-lg shadow-sm border overflow-hidden {sectionIndex === currentSection ? 'ring-2 ring-ekf-red' : ''}">
						<button onclick={() => toggleScriptSection(sectionIndex)}
							class="w-full p-4 flex items-center justify-between hover:bg-gray-50">
							<div class="flex items-center gap-3">
								<svg class="w-5 h-5 transform transition-transform {section.expanded ? 'rotate-90' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
								</svg>
								<span class="font-semibold">{section.title}</span>
								<span class="text-sm text-gray-500">({section.duration} мин)</span>
							</div>
							<div class="flex items-center gap-3">
								<div class="w-24 h-2 bg-gray-200 rounded-full overflow-hidden">
									<div class="h-full bg-green-500 transition-all" style="width: {getSectionProgress(section)}%"></div>
								</div>
								<span class="text-sm text-gray-500 w-10">{getSectionProgress(section)}%</span>
							</div>
						</button>
						{#if section.expanded}
							<div class="border-t p-4 space-y-3">
								{#each section.questions as question, questionIndex}
									<div class="space-y-2">
										<label class="flex items-start gap-3 cursor-pointer">
											<input type="checkbox" checked={question.checked}
												onchange={() => toggleQuestion(sectionIndex, questionIndex)}
												class="mt-1 h-5 w-5 rounded border-gray-300 text-ekf-red focus:ring-ekf-red" />
											<span class="flex-1 {question.checked ? 'text-gray-400 line-through' : 'text-gray-700'}">{question.text}</span>
										</label>
										<textarea value={question.notes}
											oninput={(e) => updateNotes(sectionIndex, questionIndex, (e.target as HTMLTextAreaElement).value)}
											placeholder="Заметки..."
											class="w-full ml-8 p-2 text-sm border rounded resize-none focus:ring-1 focus:ring-ekf-red"
											rows="2"></textarea>
									</div>
								{/each}
								{#if sectionIndex < scriptSections.length - 1}
									<button onclick={nextSection}
										class="mt-4 w-full py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 flex items-center justify-center gap-2">
										Следующая секция
										<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
										</svg>
									</button>
								{/if}
							</div>
						{/if}
					</div>
				{/each}
			</div>

			<!-- Progress Overview -->
			<div class="bg-white rounded-lg shadow-sm border p-4">
				<h3 class="font-semibold mb-3">Прогресс встречи</h3>
				<div class="flex gap-2">
					{#each scriptSections as section, index}
						<div class="flex-1 h-2 rounded-full
							{getSectionProgress(section) === 100 ? 'bg-green-500' : index === currentSection ? 'bg-ekf-red' : 'bg-gray-200'}"></div>
					{/each}
				</div>
			</div>
		</div>
	{/if}
</div>
