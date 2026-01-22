<script lang="ts">
	import { onMount } from 'svelte';
	import { github, type GitHubRepository, type GitHubCommit, type GitHubBranch, type GitHubPullRequest } from '$lib/api/client';

	let configured = $state(false);
	let loading = $state(true);
	let error = $state('');

	// Repository state
	let repoUrl = $state('');
	let owner = $state('');
	let repo = $state('');
	let repository = $state<GitHubRepository | null>(null);

	// Data
	let commits = $state<GitHubCommit[]>([]);
	let branches = $state<GitHubBranch[]>([]);
	let pullRequests = $state<GitHubPullRequest[]>([]);

	// UI state
	let activeTab = $state<'commits' | 'branches' | 'pulls'>('commits');
	let loadingData = $state(false);
	let selectedBranch = $state('');

	onMount(async () => {
		try {
			const status = await github.status();
			configured = status.configured;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to check GitHub status';
		} finally {
			loading = false;
		}
	});

	async function parseAndLoadRepo() {
		if (!repoUrl.trim()) return;

		loadingData = true;
		error = '';

		try {
			const parsed = await github.parseUrl(repoUrl);
			owner = parsed.owner;
			repo = parsed.repo;
			await loadRepository();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to parse repository URL';
		} finally {
			loadingData = false;
		}
	}

	async function loadRepository() {
		if (!owner || !repo) return;

		loadingData = true;
		error = '';

		try {
			repository = await github.getRepository(owner, repo);
			selectedBranch = repository.default_branch;
			await Promise.all([
				loadCommits(),
				loadBranches(),
				loadPullRequests()
			]);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load repository';
		} finally {
			loadingData = false;
		}
	}

	async function loadCommits() {
		try {
			commits = await github.getCommits(owner, repo, selectedBranch, 30);
		} catch (e) {
			console.error('Failed to load commits:', e);
		}
	}

	async function loadBranches() {
		try {
			branches = await github.getBranches(owner, repo, 30);
		} catch (e) {
			console.error('Failed to load branches:', e);
		}
	}

	async function loadPullRequests() {
		try {
			pullRequests = await github.getPullRequests(owner, repo, 'all', 30);
		} catch (e) {
			console.error('Failed to load pull requests:', e);
		}
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleDateString('ru-RU', {
			day: 'numeric',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function truncateMessage(message: string, maxLength: number = 80): string {
		const firstLine = message.split('\n')[0];
		if (firstLine.length <= maxLength) return firstLine;
		return firstLine.substring(0, maxLength) + '...';
	}
</script>

<svelte:head>
	<title>GitHub - EKF Hub</title>
</svelte:head>

<div class="p-6">
	<h1 class="text-2xl font-bold text-gray-800 mb-6">GitHub</h1>

	{#if loading}
		<div class="flex justify-center items-center h-64">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-ekf-red"></div>
		</div>
	{:else if !configured}
		<div class="bg-yellow-50 border border-yellow-200 rounded-lg p-6 text-center">
			<svg class="w-16 h-16 text-yellow-500 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
			</svg>
			<h2 class="text-xl font-semibold text-yellow-800 mb-2">GitHub не настроен</h2>
			<p class="text-yellow-700">Для работы с GitHub необходимо добавить токен в переменную окружения GITHUB_TOKEN</p>
		</div>
	{:else}
		<!-- Repository Input -->
		<div class="bg-white rounded-lg shadow p-4 mb-6">
			<div class="flex gap-4">
				<input
					type="text"
					bind:value={repoUrl}
					placeholder="Введите URL репозитория (например, https://github.com/owner/repo)"
					class="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red"
					onkeydown={(e) => e.key === 'Enter' && parseAndLoadRepo()}
				/>
				<button
					onclick={parseAndLoadRepo}
					disabled={loadingData || !repoUrl.trim()}
					class="px-6 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed"
				>
					{loadingData ? 'Загрузка...' : 'Загрузить'}
				</button>
			</div>
		</div>

		{#if error}
			<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-6">
				{error}
			</div>
		{/if}

		{#if repository}
			<!-- Repository Info -->
			<div class="bg-white rounded-lg shadow p-6 mb-6">
				<div class="flex items-start justify-between">
					<div>
						<div class="flex items-center gap-3 mb-2">
							<svg class="w-8 h-8 text-gray-700" fill="currentColor" viewBox="0 0 24 24">
								<path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
							</svg>
							<a href={repository.html_url} target="_blank" rel="noopener noreferrer" class="text-2xl font-bold text-gray-800 hover:text-ekf-red">
								{repository.full_name}
							</a>
							{#if repository.private}
								<span class="px-2 py-1 text-xs bg-yellow-100 text-yellow-800 rounded">Private</span>
							{/if}
							{#if repository.fork}
								<span class="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded">Fork</span>
							{/if}
						</div>
						{#if repository.description}
							<p class="text-gray-600 mb-4">{repository.description}</p>
						{/if}
						<div class="flex flex-wrap gap-4 text-sm text-gray-500">
							{#if repository.language}
								<span class="flex items-center gap-1">
									<span class="w-3 h-3 rounded-full bg-blue-500"></span>
									{repository.language}
								</span>
							{/if}
							<span class="flex items-center gap-1">
								<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
									<path d="M12 .587l3.668 7.568 8.332 1.151-6.064 5.828 1.48 8.279-7.416-3.967-7.417 3.967 1.481-8.279-6.064-5.828 8.332-1.151z"/>
								</svg>
								{repository.stargazers_count}
							</span>
							<span class="flex items-center gap-1">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
								</svg>
								{repository.forks_count}
							</span>
							<span>Ветка: {repository.default_branch}</span>
							<span>Обновлён: {formatDate(repository.pushed_at)}</span>
						</div>
					</div>
				</div>
			</div>

			<!-- Tabs -->
			<div class="bg-white rounded-lg shadow">
				<div class="border-b border-gray-200">
					<nav class="flex -mb-px">
						<button
							onclick={() => activeTab = 'commits'}
							class="px-6 py-3 border-b-2 font-medium text-sm {activeTab === 'commits' ? 'border-ekf-red text-ekf-red' : 'border-transparent text-gray-500 hover:text-gray-700'}"
						>
							Коммиты ({commits.length})
						</button>
						<button
							onclick={() => activeTab = 'branches'}
							class="px-6 py-3 border-b-2 font-medium text-sm {activeTab === 'branches' ? 'border-ekf-red text-ekf-red' : 'border-transparent text-gray-500 hover:text-gray-700'}"
						>
							Ветки ({branches.length})
						</button>
						<button
							onclick={() => activeTab = 'pulls'}
							class="px-6 py-3 border-b-2 font-medium text-sm {activeTab === 'pulls' ? 'border-ekf-red text-ekf-red' : 'border-transparent text-gray-500 hover:text-gray-700'}"
						>
							Pull Requests ({pullRequests.length})
						</button>
					</nav>
				</div>

				<div class="p-4">
					{#if activeTab === 'commits'}
						<!-- Branch selector for commits -->
						<div class="mb-4">
							<select
								bind:value={selectedBranch}
								onchange={loadCommits}
								class="px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-ekf-red"
							>
								{#each branches as branch}
									<option value={branch.name}>{branch.name}</option>
								{/each}
							</select>
						</div>

						<!-- Commits list -->
						<div class="divide-y divide-gray-100">
							{#each commits as commit}
								<div class="py-3 flex items-start gap-4">
									{#if commit.author}
										<img
											src={commit.author.avatar_url}
											alt={commit.author.login}
											class="w-10 h-10 rounded-full"
										/>
									{:else}
										<div class="w-10 h-10 rounded-full bg-gray-300 flex items-center justify-center">
											<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
											</svg>
										</div>
									{/if}
									<div class="flex-1 min-w-0">
										<a href={commit.html_url} target="_blank" rel="noopener noreferrer" class="text-gray-800 hover:text-ekf-red font-medium">
											{truncateMessage(commit.commit.message)}
										</a>
										<div class="text-sm text-gray-500 mt-1">
											<span class="font-medium">{commit.author?.login || commit.commit.author.name}</span>
											<span class="mx-1">·</span>
											<span>{formatDate(commit.commit.author.date)}</span>
											<span class="mx-1">·</span>
											<code class="text-xs bg-gray-100 px-1.5 py-0.5 rounded">{commit.sha.substring(0, 7)}</code>
										</div>
									</div>
								</div>
							{/each}
						</div>
					{:else if activeTab === 'branches'}
						<div class="divide-y divide-gray-100">
							{#each branches as branch}
								<div class="py-3 flex items-center justify-between">
									<div class="flex items-center gap-3">
										<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
										</svg>
										<span class="font-medium">{branch.name}</span>
										{#if branch.name === repository?.default_branch}
											<span class="px-2 py-0.5 text-xs bg-green-100 text-green-800 rounded">default</span>
										{/if}
										{#if branch.protected}
											<span class="px-2 py-0.5 text-xs bg-yellow-100 text-yellow-800 rounded">protected</span>
										{/if}
									</div>
									<code class="text-xs text-gray-500 bg-gray-100 px-2 py-1 rounded">{branch.commit.sha.substring(0, 7)}</code>
								</div>
							{/each}
						</div>
					{:else if activeTab === 'pulls'}
						<div class="divide-y divide-gray-100">
							{#each pullRequests as pr}
								<div class="py-3">
									<div class="flex items-start gap-3">
										<span class="mt-0.5 px-2 py-0.5 text-xs font-medium rounded {pr.state === 'open' ? 'bg-green-100 text-green-800' : pr.merged_at ? 'bg-purple-100 text-purple-800' : 'bg-red-100 text-red-800'}">
											{pr.state === 'open' ? 'Open' : pr.merged_at ? 'Merged' : 'Closed'}
										</span>
										<div class="flex-1 min-w-0">
											<a href={pr.html_url} target="_blank" rel="noopener noreferrer" class="text-gray-800 hover:text-ekf-red font-medium">
												{pr.title}
											</a>
											<div class="text-sm text-gray-500 mt-1">
												<span>#{pr.number}</span>
												<span class="mx-1">·</span>
												<span>opened by {pr.user?.login || 'unknown'}</span>
												<span class="mx-1">·</span>
												<span>{formatDate(pr.created_at)}</span>
											</div>
											<div class="text-xs text-gray-400 mt-1">
												<code>{pr.head.ref}</code>
												<span class="mx-1">→</span>
												<code>{pr.base.ref}</code>
											</div>
										</div>
									</div>
								</div>
							{/each}
							{#if pullRequests.length === 0}
								<div class="py-8 text-center text-gray-500">
									Нет pull requests
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>
		{:else if !error}
			<div class="bg-white rounded-lg shadow p-12 text-center">
				<svg class="w-16 h-16 text-gray-300 mx-auto mb-4" fill="currentColor" viewBox="0 0 24 24">
					<path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
				</svg>
				<h2 class="text-xl font-semibold text-gray-600 mb-2">Введите URL репозитория</h2>
				<p class="text-gray-500">Поддерживаются форматы: https://github.com/owner/repo, git@github.com:owner/repo.git</p>
			</div>
		{/if}
	{/if}
</div>
