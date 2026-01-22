<script lang="ts">
	import { onMount } from 'svelte';
	import { versions, projects, type Version, type Project, type Task } from '$lib/api/client';

	let loading = $state(true);
	let error = $state('');

	// Data
	let versionsList = $state<Version[]>([]);
	let projectsList = $state<Project[]>([]);

	// Filters
	let filterProject = $state('');
	let filterStatus = $state('');

	// UI state
	let showCreateModal = $state(false);
	let showDetailModal = $state(false);
	let showReleaseNotesModal = $state(false);
	let selectedVersion = $state<Version | null>(null);
	let selectedTasks = $state<Task[]>([]);
	let releaseNotes = $state<{ features: Task[]; fixes: Task[]; other: Task[]; total: number } | null>(null);

	// Create form
	let newVersion = $state({
		name: '',
		description: '',
		project_id: '',
		start_date: '',
		release_date: ''
	});
	let saving = $state(false);

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		loading = true;
		error = '';
		try {
			const [versionsData, projectsData] = await Promise.all([
				versions.list({ project_id: filterProject || undefined, status: filterStatus || undefined }),
				projects.list()
			]);
			versionsList = versionsData;
			projectsList = projectsData;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load data';
		} finally {
			loading = false;
		}
	}

	async function openVersionDetail(version: Version) {
		selectedVersion = version;
		try {
			const data = await versions.get(version.id);
			selectedVersion = data.version;
			selectedTasks = data.tasks;
			showDetailModal = true;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load version details';
		}
	}

	async function openReleaseNotes(version: Version) {
		selectedVersion = version;
		try {
			const data = await versions.getReleaseNotes(version.id);
			selectedVersion = data.version;
			releaseNotes = {
				features: data.features,
				fixes: data.fixes,
				other: data.other,
				total: data.total
			};
			showReleaseNotesModal = true;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load release notes';
		}
	}

	async function createVersion() {
		if (!newVersion.name.trim()) return;

		saving = true;
		error = '';
		try {
			await versions.create({
				name: newVersion.name,
				description: newVersion.description || undefined,
				project_id: newVersion.project_id || undefined,
				start_date: newVersion.start_date || undefined,
				release_date: newVersion.release_date || undefined
			});
			showCreateModal = false;
			newVersion = { name: '', description: '', project_id: '', start_date: '', release_date: '' };
			await loadData();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to create version';
		} finally {
			saving = false;
		}
	}

	async function releaseVersion(version: Version) {
		if (!confirm(`Выпустить версию "${version.name}"?`)) return;

		try {
			await versions.release(version.id);
			await loadData();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to release version';
		}
	}

	async function archiveVersion(version: Version) {
		if (!confirm(`Архивировать версию "${version.name}"?`)) return;

		try {
			await versions.update(version.id, { status: 'archived' });
			await loadData();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to archive version';
		}
	}

	async function deleteVersion(version: Version) {
		if (!confirm(`Удалить версию "${version.name}"? Задачи не будут удалены.`)) return;

		try {
			await versions.delete(version.id);
			await loadData();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to delete version';
		}
	}

	function formatDate(dateStr?: string): string {
		if (!dateStr) return '-';
		const date = new Date(dateStr);
		return date.toLocaleDateString('ru-RU', {
			day: 'numeric',
			month: 'short',
			year: 'numeric'
		});
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'unreleased': return 'bg-blue-100 text-blue-800';
			case 'released': return 'bg-green-100 text-green-800';
			case 'archived': return 'bg-gray-100 text-gray-800';
			default: return 'bg-gray-100 text-gray-800';
		}
	}

	function getStatusLabel(status: string): string {
		switch (status) {
			case 'unreleased': return 'В разработке';
			case 'released': return 'Выпущена';
			case 'archived': return 'Архив';
			default: return status;
		}
	}

	function getProjectName(projectId?: string): string {
		if (!projectId) return '-';
		const project = projectsList.find(p => p.id === projectId);
		return project?.name || '-';
	}

	$effect(() => {
		if (filterProject !== undefined || filterStatus !== undefined) {
			loadData();
		}
	});
</script>

<svelte:head>
	<title>Releases - EKF Hub</title>
</svelte:head>

<div class="p-6">
	<div class="flex justify-between items-center mb-6">
		<h1 class="text-2xl font-bold text-gray-800">Releases</h1>
		<button
			onclick={() => showCreateModal = true}
			class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 flex items-center gap-2"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
			</svg>
			Новая версия
		</button>
	</div>

	<!-- Filters -->
	<div class="bg-white rounded-lg shadow p-4 mb-6">
		<div class="flex gap-4">
			<div class="flex-1">
				<label class="block text-sm font-medium text-gray-700 mb-1">Проект</label>
				<select
					bind:value={filterProject}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red"
				>
					<option value="">Все проекты</option>
					{#each projectsList as project}
						<option value={project.id}>{project.name}</option>
					{/each}
				</select>
			</div>
			<div class="flex-1">
				<label class="block text-sm font-medium text-gray-700 mb-1">Статус</label>
				<select
					bind:value={filterStatus}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red"
				>
					<option value="">Все статусы</option>
					<option value="unreleased">В разработке</option>
					<option value="released">Выпущена</option>
					<option value="archived">Архив</option>
				</select>
			</div>
		</div>
	</div>

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-6">
			{error}
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center items-center h-64">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-ekf-red"></div>
		</div>
	{:else if versionsList.length === 0}
		<div class="bg-white rounded-lg shadow p-12 text-center">
			<svg class="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
			</svg>
			<h2 class="text-xl font-semibold text-gray-600 mb-2">Нет версий</h2>
			<p class="text-gray-500">Создайте первую версию для отслеживания релизов</p>
		</div>
	{:else}
		<div class="grid gap-4">
			{#each versionsList as version}
				<div class="bg-white rounded-lg shadow p-6">
					<div class="flex items-start justify-between">
						<div class="flex-1">
							<div class="flex items-center gap-3 mb-2">
								<button
									onclick={() => openVersionDetail(version)}
									class="text-xl font-bold text-gray-800 hover:text-ekf-red"
								>
									{version.name}
								</button>
								<span class="px-2 py-1 text-xs font-medium rounded {getStatusColor(version.status)}">
									{getStatusLabel(version.status)}
								</span>
							</div>
							{#if version.description}
								<p class="text-gray-600 mb-3">{version.description}</p>
							{/if}
							<div class="flex flex-wrap gap-4 text-sm text-gray-500">
								<span>Проект: {getProjectName(version.project_id)}</span>
								<span>Начало: {formatDate(version.start_date)}</span>
								<span>Релиз: {formatDate(version.release_date)}</span>
								{#if version.released_at}
									<span class="text-green-600">Выпущена: {formatDate(version.released_at)}</span>
								{/if}
							</div>

							<!-- Progress bar -->
							<div class="mt-4">
								<div class="flex justify-between text-sm text-gray-600 mb-1">
									<span>Прогресс: {version.tasks_done || 0} / {version.tasks_count || 0} задач</span>
									<span>{version.progress || 0}%</span>
								</div>
								<div class="w-full bg-gray-200 rounded-full h-2">
									<div
										class="bg-ekf-red h-2 rounded-full transition-all"
										style="width: {version.progress || 0}%"
									></div>
								</div>
							</div>
						</div>

						<!-- Actions -->
						<div class="flex gap-2 ml-4">
							{#if version.status === 'unreleased'}
								<button
									onclick={() => releaseVersion(version)}
									class="p-2 text-green-600 hover:bg-green-50 rounded-lg"
									title="Выпустить"
								>
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
								</button>
							{/if}
							{#if version.status === 'released'}
								<button
									onclick={() => openReleaseNotes(version)}
									class="p-2 text-blue-600 hover:bg-blue-50 rounded-lg"
									title="Release Notes"
								>
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
									</svg>
								</button>
								<button
									onclick={() => archiveVersion(version)}
									class="p-2 text-gray-600 hover:bg-gray-50 rounded-lg"
									title="Архивировать"
								>
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
									</svg>
								</button>
							{/if}
							<button
								onclick={() => deleteVersion(version)}
								class="p-2 text-red-600 hover:bg-red-50 rounded-lg"
								title="Удалить"
							>
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
								</svg>
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Create Version Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<div class="bg-white rounded-lg shadow-xl w-full max-w-md p-6">
			<div class="flex justify-between items-center mb-4">
				<h2 class="text-xl font-bold">Новая версия</h2>
				<button onclick={() => showCreateModal = false} class="text-gray-500 hover:text-gray-700">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Название *</label>
					<input
						type="text"
						bind:value={newVersion.name}
						placeholder="v1.0.0"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Проект</label>
					<select
						bind:value={newVersion.project_id}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red"
					>
						<option value="">Без проекта</option>
						{#each projectsList as project}
							<option value={project.id}>{project.name}</option>
						{/each}
					</select>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Описание</label>
					<textarea
						bind:value={newVersion.description}
						rows={3}
						placeholder="Описание версии..."
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red"
					></textarea>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Дата начала</label>
						<input
							type="date"
							bind:value={newVersion.start_date}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Плановый релиз</label>
						<input
							type="date"
							bind:value={newVersion.release_date}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red"
						/>
					</div>
				</div>
			</div>

			<div class="flex justify-end gap-3 mt-6">
				<button
					onclick={() => showCreateModal = false}
					class="px-4 py-2 text-gray-600 hover:text-gray-800"
				>
					Отмена
				</button>
				<button
					onclick={createVersion}
					disabled={saving || !newVersion.name.trim()}
					class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 disabled:opacity-50"
				>
					{saving ? 'Создание...' : 'Создать'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Version Detail Modal -->
{#if showDetailModal && selectedVersion}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<div class="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-[80vh] overflow-hidden">
			<div class="flex justify-between items-center p-6 border-b">
				<div>
					<h2 class="text-xl font-bold">{selectedVersion.name}</h2>
					<span class="px-2 py-1 text-xs font-medium rounded {getStatusColor(selectedVersion.status)}">
						{getStatusLabel(selectedVersion.status)}
					</span>
				</div>
				<button onclick={() => showDetailModal = false} class="text-gray-500 hover:text-gray-700">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="p-6 overflow-y-auto max-h-[60vh]">
				{#if selectedVersion.description}
					<p class="text-gray-600 mb-4">{selectedVersion.description}</p>
				{/if}

				<h3 class="font-semibold text-gray-800 mb-3">Задачи в этой версии ({selectedTasks.length})</h3>

				{#if selectedTasks.length === 0}
					<p class="text-gray-500">Нет задач привязанных к этой версии</p>
				{:else}
					<div class="divide-y">
						{#each selectedTasks as task}
							<div class="py-3 flex items-center justify-between">
								<div>
									<span class="font-medium">{task.title}</span>
									<div class="text-sm text-gray-500">
										<span class="px-2 py-0.5 text-xs rounded {task.status === 'done' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}">
											{task.status}
										</span>
									</div>
								</div>
								<a href="/tasks?id={task.id}" class="text-ekf-red hover:underline text-sm">
									Открыть
								</a>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}

<!-- Release Notes Modal -->
{#if showReleaseNotesModal && selectedVersion && releaseNotes}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<div class="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-[80vh] overflow-hidden">
			<div class="flex justify-between items-center p-6 border-b">
				<div>
					<h2 class="text-xl font-bold">Release Notes: {selectedVersion.name}</h2>
					<span class="text-sm text-gray-500">Всего задач: {releaseNotes.total}</span>
				</div>
				<button onclick={() => showReleaseNotesModal = false} class="text-gray-500 hover:text-gray-700">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="p-6 overflow-y-auto max-h-[60vh]">
				{#if releaseNotes.features.length > 0}
					<div class="mb-6">
						<h3 class="font-semibold text-green-700 mb-2">
							<svg class="w-5 h-5 inline mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
							</svg>
							Новые функции ({releaseNotes.features.length})
						</h3>
						<ul class="list-disc list-inside text-gray-700 space-y-1">
							{#each releaseNotes.features as task}
								<li>{task.title}</li>
							{/each}
						</ul>
					</div>
				{/if}

				{#if releaseNotes.fixes.length > 0}
					<div class="mb-6">
						<h3 class="font-semibold text-blue-700 mb-2">
							<svg class="w-5 h-5 inline mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
							</svg>
							Исправления ({releaseNotes.fixes.length})
						</h3>
						<ul class="list-disc list-inside text-gray-700 space-y-1">
							{#each releaseNotes.fixes as task}
								<li>{task.title}</li>
							{/each}
						</ul>
					</div>
				{/if}

				{#if releaseNotes.other.length > 0}
					<div class="mb-6">
						<h3 class="font-semibold text-gray-700 mb-2">
							<svg class="w-5 h-5 inline mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
							</svg>
							Прочее ({releaseNotes.other.length})
						</h3>
						<ul class="list-disc list-inside text-gray-700 space-y-1">
							{#each releaseNotes.other as task}
								<li>{task.title}</li>
							{/each}
						</ul>
					</div>
				{/if}

				{#if releaseNotes.total === 0}
					<p class="text-gray-500 text-center py-4">Нет завершённых задач в этой версии</p>
				{/if}
			</div>

			<div class="p-4 border-t bg-gray-50 flex justify-end">
				<button
					onclick={() => {
						// Copy to clipboard
						let text = `# Release Notes: ${selectedVersion?.name}\n\n`;
						if (releaseNotes?.features.length) {
							text += `## Новые функции\n`;
							releaseNotes.features.forEach(t => text += `- ${t.title}\n`);
							text += '\n';
						}
						if (releaseNotes?.fixes.length) {
							text += `## Исправления\n`;
							releaseNotes.fixes.forEach(t => text += `- ${t.title}\n`);
							text += '\n';
						}
						if (releaseNotes?.other.length) {
							text += `## Прочее\n`;
							releaseNotes.other.forEach(t => text += `- ${t.title}\n`);
						}
						navigator.clipboard.writeText(text);
						alert('Release notes скопированы!');
					}}
					class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700"
				>
					Копировать
				</button>
			</div>
		</div>
	</div>
{/if}
