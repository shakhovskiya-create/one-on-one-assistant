<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { auth as authApi, sprints as sprintsApi, versions as versionsApi } from '$lib/api/client';
	import type { User } from '$lib/stores/auth';
	import type { Sprint, Version } from '$lib/api/client';

	interface Props {
		user?: User | null;
		subordinates?: User[];
		onLogout?: () => void;
	}

	let { user = null, subordinates = [], onLogout }: Props = $props();
	let userRole = $state<string>('user');
	let sprints = $state<Sprint[]>([]);
	let versions = $state<Version[]>([]);
	let activeSprint = $state<Sprint | null>(null);

	onMount(async () => {
		try {
			const { role } = await authApi.getRole();
			userRole = role;
		} catch {
			// Ignore - user is not logged in or doesn't have role
		}

		// Load sprints and versions for Tasks sidebar
		try {
			const [sprintList, activeSprintData, versionList] = await Promise.all([
				sprintsApi.list(),
				sprintsApi.getActive().catch(() => null),
				versionsApi.list()
			]);
			sprints = sprintList || [];
			activeSprint = activeSprintData;
			versions = versionList || [];
		} catch {
			// Ignore - data will load when needed
		}
	});

	// Context Navigation — строго по макетам (2026-01-26)
	// ЗАПРЕЩЕНО дублировать Global Navigation (Почта, Сообщения, SD, Аналитика)

	const isAdmin = $derived(userRole === 'admin' || userRole === 'super_admin');
	const path = $derived($page.url.pathname);

	// Determine current module
	const currentModule = $derived(() => {
		if (path.startsWith('/tasks') || path.startsWith('/projects') || path.startsWith('/sprints') || path.startsWith('/releases') || path.startsWith('/github') || path.startsWith('/confluence') || path.startsWith('/tests') || path.startsWith('/docs')) {
			return 'tasks';
		}
		if (path.startsWith('/meetings') || path.startsWith('/calendar')) {
			return 'meetings';
		}
		if (path.startsWith('/service-desk')) {
			return 'service-desk';
		}
		return 'default';
	});

	function isActive(href: string, exact: boolean = false): string {
		if (exact) {
			return path === href ? 'bg-ekf-red text-white' : 'text-gray-300 hover:bg-white/10 hover:text-white';
		}
		return path === href || path.startsWith(href + '/')
			? 'bg-ekf-red text-white'
			: 'text-gray-300 hover:bg-white/10 hover:text-white';
	}

	// Get recent sprints (active + planned)
	const recentSprints = $derived(() => {
		const sorted = [...sprints].sort((a, b) => {
			if (a.status === 'active') return -1;
			if (b.status === 'active') return 1;
			if (a.status === 'planned' && b.status !== 'planned') return -1;
			if (b.status === 'planned' && a.status !== 'planned') return 1;
			return new Date(b.start_date || 0).getTime() - new Date(a.start_date || 0).getTime();
		});
		return sorted.slice(0, 2);
	});

	// Get recent versions (unreleased + last released)
	const recentVersions = $derived(() => {
		const unreleased = versions.filter(v => !v.released).slice(0, 1);
		const released = versions.filter(v => v.released).slice(0, 1);
		return [...unreleased, ...released];
	});
</script>

<div class="h-full flex flex-col">
	{#if currentModule() === 'tasks'}
		<!-- TASKS MODULE - по макету 01-tasks.html -->

		<!-- Project Selector -->
		<div class="p-4 border-b border-gray-700">
			<div class="flex items-center gap-2">
				<div class="w-8 h-8 bg-ekf-red rounded flex items-center justify-center text-white text-xs font-bold">EH</div>
				<div class="flex-1 min-w-0">
					<div class="font-semibold text-sm truncate text-white">EKF Hub</div>
					<div class="text-xs text-gray-400">Активный проект</div>
				</div>
				<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
				</svg>
			</div>
		</div>

		<nav class="flex-1 p-3 space-y-1 overflow-y-auto">
			<!-- Планирование -->
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-2 pb-1">Планирование</div>
			<a href="/tasks" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/tasks', true)}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2"></path>
				</svg>
				<span>Доска задач</span>
			</a>
			<a href="/tasks/backlog" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/tasks/backlog')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16"></path>
				</svg>
				<span>Бэклог</span>
			</a>
			<a href="/tasks/roadmap" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/tasks/roadmap')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
				</svg>
				<span>Roadmap</span>
			</a>

			<!-- Спринты -->
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-4 pb-1">Спринты</div>
			{#each recentSprints() as sprint}
				<a href="/sprints/{sprint.id}" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive(`/sprints/${sprint.id}`)}">
					<div class="w-5 h-5 flex items-center justify-center">
						<div class="w-2 h-2 rounded-full {sprint.status === 'active' ? 'bg-green-400' : 'bg-gray-500'}"></div>
					</div>
					<span>{sprint.name}</span>
					{#if sprint.status === 'active'}
						<span class="ml-auto text-xs bg-green-500/20 text-green-400 px-1.5 py-0.5 rounded">Active</span>
					{:else if sprint.status === 'planned'}
						<span class="ml-auto text-xs text-gray-500">Planned</span>
					{/if}
				</a>
			{/each}
			<a href="/sprints" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/sprints', true)}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"></path>
				</svg>
				<span>Все спринты</span>
			</a>

			<!-- Релизы -->
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-4 pb-1">Релизы</div>
			{#each recentVersions() as version}
				<a href="/releases/{version.id}" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive(`/releases/${version.id}`)}">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						{#if version.released}
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
						{:else}
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"></path>
						{/if}
					</svg>
					<span>{version.name}</span>
					{#if version.released}
						<span class="ml-auto text-xs text-gray-500">Released</span>
					{:else}
						<span class="ml-auto text-xs bg-yellow-500/20 text-yellow-400 px-1.5 py-0.5 rounded">Dev</span>
					{/if}
				</a>
			{/each}
			<a href="/releases" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/releases', true)}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"></path>
				</svg>
				<span>Все релизы</span>
			</a>

			<!-- Тестирование -->
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-4 pb-1">Тестирование</div>
			<a href="/tests/plans" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/tests/plans')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"></path>
				</svg>
				<span>Тест-планы</span>
			</a>
			<a href="/tests/cases" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/tests/cases')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"></path>
				</svg>
				<span>Тест-кейсы</span>
			</a>
			<a href="/tests/runs" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/tests/runs')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
				</svg>
				<span>Прогоны</span>
			</a>

			<!-- Документация -->
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-4 pb-1">Документация</div>
			<a href="/docs/wiki" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/docs/wiki')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"></path>
				</svg>
				<span>Wiki</span>
			</a>
			<a href="/docs/requirements" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/docs/requirements')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
				</svg>
				<span>Требования</span>
			</a>

			<!-- Интеграции -->
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-4 pb-1">Интеграции</div>
			<a href="/github" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/github')}">
				<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
					<path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
				</svg>
				<span>GitHub</span>
			</a>
			<a href="/confluence" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/confluence')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"></path>
				</svg>
				<span>Confluence</span>
			</a>
		</nav>

	{:else if currentModule() === 'meetings'}
		<!-- MEETINGS MODULE - по макету 04-meetings.html -->
		<div class="p-4 border-b border-white/10">
			<h2 class="text-sm font-medium text-gray-400 uppercase tracking-wide">Встречи</h2>
		</div>

		<nav class="flex-1 p-3 space-y-1 overflow-y-auto">
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-2 pb-1">Календарь</div>
			<a href="/meetings" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/meetings', true)}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
				</svg>
				<span>Мой календарь</span>
			</a>
			<a href="/meetings/team" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/meetings/team')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
				</svg>
				<span>Команда</span>
			</a>
			<a href="/meetings/rooms" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/meetings/rooms')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path>
				</svg>
				<span>Переговорные</span>
			</a>

			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-4 pb-1">Встречи</div>
			<a href="/meetings/upcoming" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/meetings/upcoming')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"></path>
				</svg>
				<span>Предстоящие</span>
			</a>
			<a href="/meetings/past" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/meetings/past')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
				</svg>
				<span>Прошедшие</span>
			</a>
			<a href="/meetings/no-notes" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/meetings/no-notes')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
				</svg>
				<span>Без протокола</span>
			</a>

			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-4 pb-1">Записи</div>
			<a href="/meetings/transcriptions" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/meetings/transcriptions')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
				</svg>
				<span>Транскрипции</span>
			</a>
			<a href="/meetings/analysis" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/meetings/analysis')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"></path>
				</svg>
				<span>AI Анализ</span>
			</a>
		</nav>

	{:else if currentModule() === 'service-desk'}
		<!-- SERVICE DESK MODULE - по макетам 02, 03 -->
		<div class="p-4 border-b border-white/10">
			<h2 class="text-sm font-medium text-gray-400 uppercase tracking-wide">Service Desk</h2>
		</div>

		<nav class="flex-1 p-3 space-y-1 overflow-y-auto">
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-2 pb-1">Обращения</div>
			<a href="/service-desk" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/service-desk', true)}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z"></path>
				</svg>
				<span>Мои заявки</span>
			</a>
			<a href="/service-desk/create" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/service-desk/create')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
				</svg>
				<span>Создать заявку</span>
			</a>

			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-4 pb-1">Статистика</div>
			<a href="/service-desk/stats" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/service-desk/stats')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
				</svg>
				<span>Статистика</span>
			</a>
		</nav>

	{:else}
		<!-- DEFAULT MODULE (Главная, Сотрудники, Почта, Сообщения, Аналитика) -->
		<div class="p-4 border-b border-white/10">
			<h2 class="text-sm font-medium text-gray-400 uppercase tracking-wide">Навигация</h2>
		</div>

		<nav class="flex-1 p-3 space-y-1 overflow-y-auto">
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-2 pb-1">Система</div>
			<a href="/settings" class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm {isActive('/settings')}">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
				</svg>
				<span>Настройки</span>
			</a>
		</nav>
	{/if}

	<!-- Admin link (only for admins) -->
	{#if isAdmin}
		<div class="p-3 border-t border-white/10">
			<a
				href="/admin"
				class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors {isActive('/admin')}"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path>
				</svg>
				<span>Админ-панель</span>
			</a>
		</div>
	{/if}
</div>
