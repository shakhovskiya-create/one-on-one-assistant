<script lang="ts">
	import { onMount } from 'svelte';
	import { projects as projectsApi } from '$lib/api/client';
	import type { Project } from '$lib/api/client';

	let projects: Project[] = $state([]);
	let loading = $state(true);
	let showCreateModal = $state(false);
	let newProject = $state({ name: '', description: '' });

	onMount(async () => {
		try {
			projects = await projectsApi.list();
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	async function createProject() {
		if (!newProject.name) return;
		try {
			const created = await projectsApi.create(newProject);
			projects = [...projects, created];
			showCreateModal = false;
			newProject = { name: '', description: '' };
		} catch (e) {
			console.error(e);
		}
	}

	async function deleteProject(id: string) {
		if (!confirm('Удалить проект?')) return;
		try {
			await projectsApi.delete(id);
			projects = projects.filter(p => p.id !== id);
		} catch (e) {
			console.error(e);
		}
	}
</script>

<svelte:head>
	<title>Проекты - EKF Team Hub</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-gray-900">Проекты</h1>
		<button
			onclick={() => showCreateModal = true}
			class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors"
		>
			Новый проект
		</button>
	</div>

	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="text-gray-500">Загрузка...</div>
		</div>
	{:else if projects.length === 0}
		<div class="bg-white rounded-xl shadow-sm p-12 text-center">
			<div class="text-gray-400 text-lg mb-4">Проектов пока нет</div>
			<button
				onclick={() => showCreateModal = true}
				class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors"
			>
				Создать первый проект
			</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each projects as project}
				<div class="bg-white rounded-xl shadow-sm p-6 hover:shadow-md transition-shadow">
					<div class="flex items-start justify-between mb-4">
						<a href="/projects/{project.id}" class="text-lg font-semibold text-gray-900 hover:text-ekf-red">
							{project.name}
						</a>
						<button
							onclick={() => deleteProject(project.id)}
							class="text-gray-400 hover:text-red-500"
							title="Удалить"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
							</svg>
						</button>
					</div>
					{#if project.description}
						<p class="text-gray-600 text-sm line-clamp-3">{project.description}</p>
					{/if}
					<div class="mt-4 flex items-center justify-between">
						<span class="text-sm text-gray-500">{project.meetings_count || 0} встреч</span>
						<a href="/projects/{project.id}" class="text-sm text-ekf-red hover:underline flex items-center gap-1">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
							</svg>
							Gantt
						</a>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Create Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
		<div class="bg-white rounded-xl shadow-xl p-6 w-full max-w-md">
			<h2 class="text-xl font-bold text-gray-900 mb-4">Новый проект</h2>
			<div class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Название</label>
					<input
						type="text"
						bind:value={newProject.name}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						placeholder="Введите название проекта"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Описание</label>
					<textarea
						bind:value={newProject.description}
						rows="3"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						placeholder="Описание проекта"
					></textarea>
				</div>
			</div>
			<div class="flex gap-3 mt-6">
				<button
					onclick={() => showCreateModal = false}
					class="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
				>
					Отмена
				</button>
				<button
					onclick={createProject}
					class="flex-1 px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700"
				>
					Создать
				</button>
			</div>
		</div>
	</div>
{/if}
