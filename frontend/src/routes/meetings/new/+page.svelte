<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { meetings as meetingsApi, employees as employeesApi, projects as projectsApi } from '$lib/api/client';
	import type { Employee, Project, MeetingCategory } from '$lib/api/client';

	let employees: Employee[] = $state([]);
	let projects: Project[] = $state([]);
	let categories: MeetingCategory[] = $state([]);
	let loading = $state(true);
	let processing = $state(false);
	let progress = $state('');

	let formData = $state({
		employee_id: $page.url.searchParams.get('employee_id') || '',
		project_id: $page.url.searchParams.get('project_id') || '',
		category: 'one_on_one',
		title: '',
		date: new Date().toISOString().split('T')[0]
	});

	let audioFile: File | null = $state(null);
	let audioUrl = $state('');

	onMount(async () => {
		try {
			[employees, projects, categories] = await Promise.all([
				employeesApi.list(),
				projectsApi.list(),
				meetingsApi.getCategories()
			]);
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	function handleFileSelect(event: Event) {
		const input = event.target as HTMLInputElement;
		if (input.files && input.files.length > 0) {
			audioFile = input.files[0];
			audioUrl = URL.createObjectURL(audioFile);
		}
	}

	async function processMeeting() {
		if (!audioFile || !formData.employee_id || !formData.category) {
			alert('Заполните все обязательные поля и загрузите аудиофайл');
			return;
		}

		processing = true;
		progress = 'Загрузка файла...';

		try {
			const data = new FormData();
			data.append('audio', audioFile);
			data.append('employee_id', formData.employee_id);
			data.append('category', formData.category);
			data.append('title', formData.title);
			data.append('date', formData.date);
			if (formData.project_id) {
				data.append('project_id', formData.project_id);
			}

			progress = 'Обработка аудио...';
			const result = await meetingsApi.process(data);

			progress = 'Готово!';
			await goto(`/meetings/${result.id}`);
		} catch (e) {
			console.error(e);
			alert('Ошибка обработки встречи');
		} finally {
			processing = false;
			progress = '';
		}
	}
</script>

<svelte:head>
	<title>Новая встреча - EKF Hub</title>
</svelte:head>

<div class="max-w-2xl mx-auto space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-gray-900">Новая встреча</h1>
		<a href="/meetings" class="text-gray-500 hover:text-gray-700">
			Отмена
		</a>
	</div>

	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="text-gray-500">Загрузка...</div>
		</div>
	{:else}
		<div class="bg-white rounded-xl shadow-sm p-6 space-y-6">
			<!-- Category -->
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-2">Тип встречи *</label>
				<div class="grid grid-cols-3 gap-3">
					{#each categories as cat}
						<button
							type="button"
							onclick={() => formData.category = cat.id}
							class="px-4 py-3 rounded-lg border-2 text-sm font-medium transition-colors
								{formData.category === cat.id
									? 'border-ekf-red bg-red-50 text-ekf-red'
									: 'border-gray-200 text-gray-600 hover:border-gray-300'}"
						>
							{cat.name}
						</button>
					{/each}
				</div>
			</div>

			<!-- Employee -->
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Сотрудник *</label>
				<select
					bind:value={formData.employee_id}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
				>
					<option value="">Выберите сотрудника</option>
					{#each employees as emp}
						<option value={emp.id}>{emp.name} - {emp.position}</option>
					{/each}
				</select>
			</div>

			<!-- Project (optional) -->
			{#if formData.category === 'project'}
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Проект</label>
					<select
						bind:value={formData.project_id}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
					>
						<option value="">Выберите проект</option>
						{#each projects as proj}
							<option value={proj.id}>{proj.name}</option>
						{/each}
					</select>
				</div>
			{/if}

			<!-- Title -->
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Название встречи</label>
				<input
					type="text"
					bind:value={formData.title}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
					placeholder="Например: Еженедельная встреча"
				/>
			</div>

			<!-- Date -->
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Дата</label>
				<input
					type="date"
					bind:value={formData.date}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
				/>
			</div>

			<!-- Audio Upload -->
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-2">Аудиозапись *</label>
				<div class="border-2 border-dashed border-gray-300 rounded-lg p-6 text-center hover:border-gray-400 transition-colors">
					{#if audioFile}
						<div class="space-y-3">
							<div class="text-green-600 font-medium">{audioFile.name}</div>
							<div class="text-sm text-gray-500">
								{(audioFile.size / 1024 / 1024).toFixed(2)} MB
							</div>
							<audio controls class="mx-auto">
								<source src={audioUrl} type={audioFile.type} />
							</audio>
							<button
								type="button"
								onclick={() => { audioFile = null; audioUrl = ''; }}
								class="text-red-600 hover:text-red-700 text-sm"
							>
								Удалить
							</button>
						</div>
					{:else}
						<input
							type="file"
							accept="audio/*"
							onchange={handleFileSelect}
							class="hidden"
							id="audio-input"
						/>
						<label for="audio-input" class="cursor-pointer">
							<div class="text-gray-400 mb-2">
								<svg class="w-12 h-12 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
								</svg>
							</div>
							<div class="text-gray-600">Нажмите для загрузки или перетащите файл</div>
							<div class="text-sm text-gray-400 mt-1">MP3, WAV, M4A до 100MB</div>
						</label>
					{/if}
				</div>
			</div>

			<!-- Submit -->
			<div class="pt-4">
				<button
					type="button"
					onclick={processMeeting}
					disabled={processing}
					class="w-full px-6 py-3 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed font-medium"
				>
					{#if processing}
						<span class="flex items-center justify-center gap-2">
							<svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
							</svg>
							{progress}
						</span>
					{:else}
						Обработать встречу
					{/if}
				</button>
			</div>
		</div>
	{/if}
</div>
