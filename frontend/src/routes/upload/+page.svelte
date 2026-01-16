<script lang="ts">
	import { onMount } from 'svelte';
	import { subordinates } from '$lib/stores/auth';
	import { meetings, projects as projectsApi } from '$lib/api/client';
	import type { Project, MeetingCategory } from '$lib/api/client';

	// Form state
	let categories: MeetingCategory[] = $state([]);
	let projectsList: Project[] = $state([]);
	let selectedCategory = $state('one_on_one');
	let selectedEmployee = $state('');
	let selectedProject = $state('');
	let selectedParticipants: string[] = $state([]);
	let meetingTitle = $state('');
	let meetingDate = $state(new Date().toISOString().split('T')[0]);

	// File state
	let file: File | null = $state(null);
	let dragActive = $state(false);

	// Processing state
	type ProcessingStatus = 'idle' | 'uploading' | 'transcribing' | 'analyzing' | 'done' | 'error';
	let status: ProcessingStatus = $state('idle');
	let progress = $state(0);
	let error = $state('');
	let result: any = $state(null);
	let showTranscriptDetails = $state(false);

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

	onMount(async () => {
		try {
			const [cats, projs] = await Promise.all([
				meetings.getCategories().catch(() => []),
				projectsApi.list('active').catch(() => [])
			]);
			categories = cats || [];
			projectsList = projs || [];
		} catch (e) {
			console.error('Failed to load data:', e);
		}
	});

	function handleDragEnter(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		dragActive = true;
	}

	function handleDragLeave(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		dragActive = false;
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		dragActive = false;
		if (e.dataTransfer?.files && e.dataTransfer.files[0]) {
			file = e.dataTransfer.files[0];
		}
	}

	function handleFileChange(e: Event) {
		const target = e.target as HTMLInputElement;
		if (target.files && target.files[0]) {
			file = target.files[0];
		}
	}

	function toggleParticipant(id: string) {
		if (selectedParticipants.includes(id)) {
			selectedParticipants = selectedParticipants.filter(p => p !== id);
		} else {
			selectedParticipants = [...selectedParticipants, id];
		}
	}

	async function processFile() {
		if (!canSubmit()) {
			error = 'Заполните обязательные поля';
			return;
		}

		error = '';
		status = 'uploading';
		progress = 10;

		try {
			const formData = new FormData();
			formData.append('file', file!);
			formData.append('category_code', selectedCategory);
			formData.append('meeting_date', meetingDate);

			if (meetingTitle) {
				formData.append('title', meetingTitle);
			}

			if (isOneOnOne && selectedEmployee) {
				formData.append('employee_id', selectedEmployee);
			}

			if (needsProject && selectedProject) {
				formData.append('project_id', selectedProject);
			}

			if (needsParticipants && selectedParticipants.length > 0) {
				formData.append('participant_ids', JSON.stringify(selectedParticipants));
			}

			status = 'transcribing';
			progress = 30;

			const data = await meetings.process(formData);

			status = 'analyzing';
			progress = 90;

			if (data.error) {
				throw new Error(data.error);
			}

			result = data;
			status = 'done';
			progress = 100;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Произошла ошибка';
			status = 'error';
		}
	}

	function reset() {
		file = null;
		status = 'idle';
		progress = 0;
		result = null;
		error = '';
		showTranscriptDetails = false;
	}

	function getCategoryName(code: string): string {
		return categories.find(c => c.code === code)?.name || code;
	}

	function getStatusText(): string {
		switch (status) {
			case 'uploading': return 'Загрузка файла...';
			case 'transcribing': return 'Транскрибирование (Whisper + Yandex)...';
			case 'analyzing': return 'AI-анализ встречи...';
			case 'done': return 'Готово!';
			case 'error': return 'Ошибка';
			default: return '';
		}
	}
</script>

<svelte:head>
	<title>Загрузить запись - EKF Team Hub</title>
</svelte:head>

<div class="max-w-4xl mx-auto space-y-6">
	<h1 class="text-2xl font-bold text-gray-900">Загрузить запись встречи</h1>

	{#if status === 'idle'}
		<!-- Category Selection -->
		<div class="bg-white rounded-lg shadow-sm border p-6">
			<label class="block text-sm font-medium text-gray-700 mb-3">Тип встречи</label>
			<div class="grid grid-cols-4 gap-2">
				{#each categories as cat}
					<button
						onclick={() => {
							selectedCategory = cat.code;
							selectedEmployee = '';
							selectedProject = '';
							selectedParticipants = [];
						}}
						class="px-4 py-2 rounded-lg text-sm font-medium transition-colors
							{selectedCategory === cat.code
								? 'bg-ekf-red text-white'
								: 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
					>
						{cat.name}
					</button>
				{/each}
			</div>
		</div>

		<!-- Dynamic Settings -->
		<div class="bg-white rounded-lg shadow-sm border p-6 space-y-4">
			<div class="grid grid-cols-2 gap-4">
				<!-- Title -->
				<div class="col-span-2">
					<label class="block text-sm font-medium text-gray-700 mb-1">
						Название встречи (опционально)
					</label>
					<input
						type="text"
						bind:value={meetingTitle}
						placeholder="{getCategoryName(selectedCategory)} - {meetingDate}"
						class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent"
					/>
				</div>

				<!-- Employee for 1-on-1 -->
				{#if isOneOnOne}
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Сотрудник *</label>
						<select
							bind:value={selectedEmployee}
							class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						>
							<option value="">Выберите сотрудника</option>
							{#each $subordinates as emp}
								<option value={emp.id}>{emp.name} - {emp.position}</option>
							{/each}
						</select>
					</div>
				{/if}

				<!-- Project for team meetings -->
				{#if needsProject}
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Проект *</label>
						<select
							bind:value={selectedProject}
							class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						>
							<option value="">Выберите проект</option>
							{#each projectsList as proj}
								<option value={proj.id}>{proj.name}</option>
							{/each}
						</select>
					</div>
				{/if}

				<!-- Date -->
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Дата встречи</label>
					<input
						type="date"
						bind:value={meetingDate}
						class="w-full border rounded-lg px-3 py-2 focus:ring-2 focus:ring-ekf-red focus:border-transparent"
					/>
				</div>
			</div>

			<!-- Participants for team meetings -->
			{#if needsParticipants}
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-2">Участники</label>
					<div class="flex flex-wrap gap-2">
						{#each $subordinates as emp}
							<button
								onclick={() => toggleParticipant(emp.id)}
								class="flex items-center gap-1 px-3 py-1 rounded-full text-sm transition-colors
									{selectedParticipants.includes(emp.id)
										? 'bg-ekf-red/10 text-ekf-red border border-ekf-red'
										: 'bg-gray-100 text-gray-600 hover:bg-gray-200'}"
							>
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
								</svg>
								{emp.name}
							</button>
						{/each}
					</div>
					{#if selectedParticipants.length > 0}
						<p class="text-sm text-gray-500 mt-2">Выбрано: {selectedParticipants.length}</p>
					{/if}
				</div>
			{/if}
		</div>

		<!-- Drop Zone -->
		<div
			role="button"
			tabindex="0"
			ondragenter={handleDragEnter}
			ondragleave={handleDragLeave}
			ondragover={handleDragOver}
			ondrop={handleDrop}
			class="bg-white rounded-lg shadow-sm border-2 border-dashed p-12 text-center transition-colors
				{dragActive ? 'border-ekf-red bg-orange-50' : file ? 'border-green-500 bg-green-50' : 'border-gray-300'}"
		>
			{#if file}
				<div class="space-y-4">
					<svg class="w-12 h-12 mx-auto text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
					</svg>
					<div>
						<p class="font-medium text-gray-900">{file.name}</p>
						<p class="text-sm text-gray-500">{(file.size / 1024 / 1024).toFixed(2)} MB</p>
					</div>
					<button onclick={() => file = null} class="text-sm text-red-600 hover:underline">
						Удалить
					</button>
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
								<input
									type="file"
									accept="audio/*,video/*"
									onchange={handleFileChange}
									class="hidden"
								/>
							</label>
						</p>
						<p class="text-sm text-gray-500 mt-1">Поддерживаются: MP3, WAV, MP4, WebM, M4A, OGG</p>
					</div>
				</div>
			{/if}
		</div>

		{#if error}
			<div class="bg-red-50 border border-red-200 rounded-lg p-4 flex items-center gap-3">
				<svg class="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				<span class="text-red-700">{error}</span>
			</div>
		{/if}

		<button
			onclick={processFile}
			disabled={!canSubmit()}
			class="w-full py-3 bg-ekf-red text-white rounded-lg hover:bg-red-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-medium transition-colors"
		>
			Обработать запись
		</button>
	{/if}

	{#if status === 'uploading' || status === 'transcribing' || status === 'analyzing'}
		<div class="bg-white rounded-lg shadow-sm border p-12 text-center space-y-6">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-ekf-red mx-auto"></div>
			<div>
				<p class="font-medium text-gray-900">{getStatusText()}</p>
				<p class="text-sm text-gray-500 mt-1">Это может занять несколько минут</p>
			</div>
			<div class="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
				<div
					class="h-full bg-ekf-red transition-all duration-500"
					style="width: {progress}%"
				></div>
			</div>
		</div>
	{/if}

	{#if status === 'done' && result}
		<div class="space-y-6">
			<div class="bg-green-50 border border-green-200 rounded-lg p-4 flex items-center justify-between">
				<div class="flex items-center gap-3">
					<svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
					<span class="text-green-700 font-medium">Анализ завершен!</span>
				</div>
				<button onclick={reset} class="text-gray-600 hover:text-gray-800 text-xl">&times;</button>
			</div>

			<!-- Summary -->
			{#if result.analysis?.summary}
				<div class="bg-white rounded-lg shadow-sm border p-6">
					<h2 class="text-lg font-semibold mb-4">Резюме встречи</h2>
					<p class="text-gray-700">{result.analysis.summary}</p>

					{#if result.analysis.mood_score}
						<div class="mt-4 flex items-center gap-2">
							<span class="text-sm text-gray-500">Настроение:</span>
							<span class="px-3 py-1 rounded-full text-sm font-medium
								{result.analysis.mood_score >= 7 ? 'bg-green-100 text-green-800' :
								result.analysis.mood_score >= 5 ? 'bg-yellow-100 text-yellow-800' : 'bg-red-100 text-red-800'}">
								{result.analysis.mood_score}/10
							</span>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Agreements -->
			{#if result.analysis?.agreements?.length > 0}
				<div class="bg-white rounded-lg shadow-sm border p-6">
					<h3 class="font-semibold mb-4">Договоренности</h3>
					<div class="overflow-x-auto">
						<table class="w-full">
							<thead>
								<tr class="border-b">
									<th class="text-left py-2 px-3">Задача</th>
									<th class="text-left py-2 px-3">Ответственный</th>
									<th class="text-left py-2 px-3">Срок</th>
								</tr>
							</thead>
							<tbody>
								{#each result.analysis.agreements as item}
									<tr class="border-b">
										<td class="py-2 px-3">{item.task}</td>
										<td class="py-2 px-3">{item.responsible}</td>
										<td class="py-2 px-3">{item.deadline || '-'}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}

			<!-- Red Flags -->
			{#if result.analysis?.red_flags}
				<div class="bg-white rounded-lg shadow-sm border p-6">
					<h3 class="font-semibold mb-4">Красные флаги</h3>
					<div class="grid grid-cols-3 gap-4">
						<div class="p-4 rounded-lg {result.analysis.red_flags.burnout_signs ? 'bg-red-50 border border-red-200' : 'bg-green-50 border border-green-200'}">
							<p class="font-medium text-sm">Признаки выгорания</p>
							<p class="text-sm mt-1 {result.analysis.red_flags.burnout_signs ? 'text-red-700' : 'text-green-700'}">
								{result.analysis.red_flags.burnout_signs || 'Нет'}
							</p>
						</div>
						<div class="p-4 rounded-lg {result.analysis.red_flags.turnover_risk === 'high' ? 'bg-red-50 border border-red-200' : result.analysis.red_flags.turnover_risk === 'medium' ? 'bg-yellow-50 border border-yellow-200' : 'bg-green-50 border border-green-200'}">
							<p class="font-medium text-sm">Риск ухода</p>
							<p class="text-sm mt-1">
								{result.analysis.red_flags.turnover_risk === 'high' ? 'Высокий' :
								result.analysis.red_flags.turnover_risk === 'medium' ? 'Средний' : 'Низкий'}
							</p>
						</div>
						<div class="p-4 rounded-lg {result.analysis.red_flags.team_conflicts ? 'bg-red-50 border border-red-200' : 'bg-green-50 border border-green-200'}">
							<p class="font-medium text-sm">Конфликты</p>
							<p class="text-sm mt-1 {result.analysis.red_flags.team_conflicts ? 'text-red-700' : 'text-green-700'}">
								{result.analysis.red_flags.team_conflicts || 'Нет'}
							</p>
						</div>
					</div>
				</div>
			{/if}

			<!-- Transcript -->
			{#if result.transcript}
				<details class="bg-white rounded-lg shadow-sm border">
					<summary class="p-6 cursor-pointer font-semibold">Транскрипт встречи</summary>
					<div class="px-6 pb-6">
						<pre class="whitespace-pre-wrap text-sm text-gray-700 bg-gray-50 p-4 rounded-lg max-h-96 overflow-y-auto">{result.transcript.merged || result.transcript}</pre>
					</div>
				</details>
			{/if}

			<button
				onclick={reset}
				class="w-full py-3 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 font-medium"
			>
				Загрузить еще одну запись
			</button>
		</div>
	{/if}

	{#if status === 'error'}
		<div class="bg-white rounded-lg shadow-sm border p-12 text-center space-y-6">
			<svg class="w-12 h-12 mx-auto text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			<div>
				<p class="font-medium text-gray-900">Произошла ошибка</p>
				<p class="text-sm text-gray-500 mt-1">{error}</p>
			</div>
			<button
				onclick={reset}
				class="px-6 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200"
			>
				Попробовать снова
			</button>
		</div>
	{/if}
</div>
