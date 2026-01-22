<script lang="ts">
	import { onMount } from 'svelte';
	import { confluence } from '$lib/api/client';
	import type { ConfluenceSpace, ConfluenceContent, ConfluenceSearchResult } from '$lib/api/client';

	let configured = $state(false);
	let confluenceUrl = $state('');
	let loading = $state(true);
	let spaces: ConfluenceSpace[] = $state([]);
	let selectedSpace: ConfluenceSpace | null = $state(null);
	let spaceContent: ConfluenceContent[] = $state([]);
	let loadingContent = $state(false);

	// Search
	let searchQuery = $state('');
	let searchResults: ConfluenceSearchResult[] = $state([]);
	let searching = $state(false);
	let searchTotalSize = $state(0);

	// Page view
	let selectedPage: ConfluenceContent | null = $state(null);
	let loadingPage = $state(false);

	// Recent pages
	let recentPages: ConfluenceContent[] = $state([]);

	onMount(async () => {
		try {
			const status = await confluence.status();
			configured = status.configured;
			confluenceUrl = status.url;

			if (configured) {
				const [spacesRes, recentRes] = await Promise.all([
					confluence.getSpaces(50),
					confluence.getRecent(10)
				]);
				spaces = spacesRes.spaces || [];
				recentPages = recentRes.pages || [];
			}
		} catch (e) {
			console.error('Failed to load Confluence status:', e);
		} finally {
			loading = false;
		}
	});

	async function selectSpace(space: ConfluenceSpace) {
		selectedSpace = space;
		selectedPage = null;
		loadingContent = true;
		try {
			const res = await confluence.getSpaceContent(space.key, 'page', 50);
			spaceContent = res.pages || [];
		} catch (e) {
			console.error('Failed to load space content:', e);
			spaceContent = [];
		} finally {
			loadingContent = false;
		}
	}

	async function viewPage(page: ConfluenceContent) {
		loadingPage = true;
		try {
			const fullPage = await confluence.getPage(page.id, true);
			selectedPage = fullPage;
		} catch (e) {
			console.error('Failed to load page:', e);
		} finally {
			loadingPage = false;
		}
	}

	async function search() {
		if (!searchQuery.trim()) return;
		searching = true;
		try {
			const res = await confluence.search(searchQuery, selectedSpace?.key, 20);
			searchResults = res.results || [];
			searchTotalSize = res.totalSize || 0;
		} catch (e) {
			console.error('Search failed:', e);
			searchResults = [];
		} finally {
			searching = false;
		}
	}

	function clearSearch() {
		searchQuery = '';
		searchResults = [];
		searchTotalSize = 0;
	}

	function openInConfluence(url: string) {
		window.open(url, '_blank');
	}
</script>

<svelte:head>
	<title>Confluence - EKF Hub</title>
</svelte:head>

<div class="h-full flex flex-col bg-gray-50">
	<!-- Header -->
	<div class="bg-white border-b border-gray-200 px-6 py-4">
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-3">
				<div class="w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center">
					<svg class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="currentColor">
						<path d="M12.714 9.669a3.619 3.619 0 01-2.764-1.23c-.54-.61-.876-1.23-.876-1.83 0-.42.12-.78.336-1.02.24-.27.576-.39.936-.39h4.776c.36 0 .696.12.936.39.216.24.336.6.336 1.02 0 .6-.336 1.22-.876 1.83a3.619 3.619 0 01-2.804 1.23zm0 4.662a3.619 3.619 0 01-2.764-1.23c-.54-.61-.876-1.23-.876-1.83 0-.42.12-.78.336-1.02.24-.27.576-.39.936-.39h4.776c.36 0 .696.12.936.39.216.24.336.6.336 1.02 0 .6-.336 1.22-.876 1.83a3.619 3.619 0 01-2.804 1.23zm0 4.662a3.619 3.619 0 01-2.764-1.23c-.54-.61-.876-1.23-.876-1.83 0-.42.12-.78.336-1.02.24-.27.576-.39.936-.39h4.776c.36 0 .696.12.936.39.216.24.336.6.336 1.02 0 .6-.336 1.22-.876 1.83a3.619 3.619 0 01-2.804 1.23z"/>
					</svg>
				</div>
				<div>
					<h1 class="text-xl font-semibold text-gray-900">Confluence</h1>
					<p class="text-sm text-gray-500">База знаний компании</p>
				</div>
			</div>
			{#if configured}
				<a
					href={confluenceUrl}
					target="_blank"
					class="px-4 py-2 text-sm text-blue-600 hover:bg-blue-50 rounded-lg flex items-center gap-2"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
					</svg>
					Открыть в Confluence
				</a>
			{/if}
		</div>

		{#if configured}
			<!-- Search -->
			<div class="mt-4">
				<div class="flex gap-2">
					<div class="flex-1 relative">
						<input
							type="text"
							bind:value={searchQuery}
							onkeydown={(e) => e.key === 'Enter' && search()}
							placeholder="Поиск по Confluence..."
							class="w-full px-4 py-2 pl-10 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						/>
						<svg class="w-5 h-5 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
						</svg>
					</div>
					<button
						onclick={search}
						disabled={searching || !searchQuery.trim()}
						class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						{searching ? 'Поиск...' : 'Найти'}
					</button>
					{#if searchResults.length > 0}
						<button
							onclick={clearSearch}
							class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
						>
							Очистить
						</button>
					{/if}
				</div>
			</div>
		{/if}
	</div>

	{#if loading}
		<div class="flex-1 flex items-center justify-center">
			<div class="text-center">
				<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
				<p class="mt-4 text-gray-500">Загрузка...</p>
			</div>
		</div>
	{:else if !configured}
		<div class="flex-1 flex items-center justify-center">
			<div class="text-center max-w-md">
				<div class="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
				</div>
				<h2 class="text-lg font-medium text-gray-900 mb-2">Confluence не настроен</h2>
				<p class="text-gray-500">Обратитесь к администратору для настройки интеграции с Confluence.</p>
			</div>
		</div>
	{:else}
		<div class="flex-1 flex overflow-hidden">
			<!-- Sidebar: Spaces -->
			<div class="w-64 bg-white border-r border-gray-200 flex flex-col">
				<div class="p-4 border-b border-gray-200">
					<h2 class="font-medium text-gray-900">Пространства</h2>
				</div>
				<div class="flex-1 overflow-y-auto">
					{#each spaces as space}
						<button
							onclick={() => selectSpace(space)}
							class="w-full px-4 py-3 text-left hover:bg-gray-50 flex items-center gap-3 border-b border-gray-100"
							class:bg-blue-50={selectedSpace?.key === space.key}
							class:text-blue-700={selectedSpace?.key === space.key}
						>
							<div class="w-8 h-8 bg-blue-100 text-blue-600 rounded flex items-center justify-center font-medium text-sm">
								{space.key.substring(0, 2).toUpperCase()}
							</div>
							<div class="flex-1 min-w-0">
								<div class="font-medium truncate">{space.name}</div>
								<div class="text-xs text-gray-500">{space.key}</div>
							</div>
						</button>
					{/each}
				</div>
			</div>

			<!-- Main Content -->
			<div class="flex-1 flex flex-col overflow-hidden">
				{#if searchResults.length > 0}
					<!-- Search Results -->
					<div class="flex-1 overflow-y-auto p-6">
						<div class="flex items-center justify-between mb-4">
							<h2 class="text-lg font-medium text-gray-900">
								Результаты поиска
								<span class="text-sm font-normal text-gray-500">({searchTotalSize} найдено)</span>
							</h2>
						</div>
						<div class="space-y-3">
							{#each searchResults as result}
								<div
									class="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-shadow cursor-pointer"
									onclick={() => openInConfluence(result.content._links.webui)}
								>
									<h3 class="font-medium text-blue-600 hover:text-blue-800">{result.title}</h3>
									<p class="text-sm text-gray-600 mt-1 line-clamp-2">{@html result.excerpt}</p>
									<div class="flex items-center gap-4 mt-2 text-xs text-gray-500">
										<span>{result.content.space?.name || 'Unknown space'}</span>
										<span>{result.friendlyLastModified}</span>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{:else if selectedPage}
					<!-- Page View -->
					<div class="flex-1 overflow-y-auto">
						<div class="bg-white border-b border-gray-200 px-6 py-4 sticky top-0">
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<button
										onclick={() => selectedPage = null}
										class="p-1 hover:bg-gray-100 rounded"
									>
										<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
										</svg>
									</button>
									<h2 class="text-lg font-medium text-gray-900">{selectedPage.title}</h2>
								</div>
								<button
									onclick={() => openInConfluence(selectedPage?._links.webui || '')}
									class="px-3 py-1.5 text-sm text-blue-600 hover:bg-blue-50 rounded-lg flex items-center gap-1"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
									</svg>
									Открыть
								</button>
							</div>
							{#if selectedPage.ancestors && selectedPage.ancestors.length > 0}
								<div class="flex items-center gap-1 mt-2 text-sm text-gray-500">
									{#each selectedPage.ancestors as ancestor, i}
										<span class="hover:text-blue-600 cursor-pointer" onclick={() => viewPage(ancestor)}>
											{ancestor.title}
										</span>
										{#if i < selectedPage.ancestors.length - 1}
											<span>/</span>
										{/if}
									{/each}
								</div>
							{/if}
						</div>
						<div class="p-6">
							{#if loadingPage}
								<div class="flex items-center justify-center py-12">
									<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
								</div>
							{:else if selectedPage.body?.view?.value}
								<div class="prose max-w-none confluence-content">
									{@html selectedPage.body.view.value}
								</div>
							{:else}
								<p class="text-gray-500 italic">Содержимое страницы недоступно</p>
							{/if}
						</div>
					</div>
				{:else if selectedSpace}
					<!-- Space Content -->
					<div class="flex-1 overflow-y-auto p-6">
						<div class="flex items-center justify-between mb-4">
							<div>
								<h2 class="text-lg font-medium text-gray-900">{selectedSpace.name}</h2>
								{#if selectedSpace.description?.plain?.value}
									<p class="text-sm text-gray-500 mt-1">{selectedSpace.description.plain.value}</p>
								{/if}
							</div>
							<button
								onclick={() => openInConfluence(selectedSpace?._links.webui || '')}
								class="px-3 py-1.5 text-sm text-blue-600 hover:bg-blue-50 rounded-lg flex items-center gap-1"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
								</svg>
								Открыть
							</button>
						</div>

						{#if loadingContent}
							<div class="flex items-center justify-center py-12">
								<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
							</div>
						{:else if spaceContent.length === 0}
							<p class="text-gray-500 text-center py-12">Нет страниц в этом пространстве</p>
						{:else}
							<div class="space-y-2">
								{#each spaceContent as page}
									<button
										onclick={() => viewPage(page)}
										class="w-full text-left bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-shadow flex items-center gap-3"
									>
										<svg class="w-5 h-5 text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
										</svg>
										<div class="flex-1 min-w-0">
											<div class="font-medium text-gray-900 truncate">{page.title}</div>
											{#if page.version}
												<div class="text-xs text-gray-500">Версия {page.version.number}</div>
											{/if}
										</div>
										<svg class="w-5 h-5 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
										</svg>
									</button>
								{/each}
							</div>
						{/if}
					</div>
				{:else}
					<!-- Recent Pages -->
					<div class="flex-1 overflow-y-auto p-6">
						<h2 class="text-lg font-medium text-gray-900 mb-4">Недавние страницы</h2>
						{#if recentPages.length === 0}
							<p class="text-gray-500 text-center py-12">Выберите пространство слева или воспользуйтесь поиском</p>
						{:else}
							<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
								{#each recentPages as page}
									<button
										onclick={() => viewPage(page)}
										class="text-left bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-shadow"
									>
										<div class="font-medium text-gray-900 truncate">{page.title}</div>
										<div class="text-sm text-gray-500 mt-1">{page.space?.name || 'Unknown space'}</div>
										{#if page.version}
											<div class="text-xs text-gray-400 mt-2">Версия {page.version.number}</div>
										{/if}
									</button>
								{/each}
							</div>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	:global(.confluence-content) {
		line-height: 1.6;
	}

	:global(.confluence-content img) {
		max-width: 100%;
		height: auto;
	}

	:global(.confluence-content table) {
		border-collapse: collapse;
		width: 100%;
	}

	:global(.confluence-content td),
	:global(.confluence-content th) {
		padding: 0.5rem;
		border: 1px solid #e5e7eb;
	}

	:global(.confluence-content a) {
		color: #2563eb;
	}

	:global(.confluence-content a:hover) {
		text-decoration: underline;
	}

	.line-clamp-2 {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>
