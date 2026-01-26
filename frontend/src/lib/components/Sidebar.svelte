<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { auth as authApi } from '$lib/api/client';
	import type { User } from '$lib/stores/auth';

	interface Props {
		user?: User | null;
		subordinates?: User[];
		onLogout?: () => void;
	}

	let { user = null, subordinates = [], onLogout }: Props = $props();
	let userRole = $state<string>('user');

	onMount(async () => {
		try {
			const { role } = await authApi.getRole();
			userRole = role;
		} catch {
			// Ignore - user is not logged in or doesn't have role
		}
	});

	// Context Navigation — строго по макетам (2026-01-26)
	// ЗАПРЕЩЕНО дублировать Global Navigation (Почта, Сообщения, SD, Аналитика)

	// Раздел: Задачи (по макету 01-tasks.html)
	const tasksSidebarItems = [
		{ href: '/tasks', label: 'Доска задач', icon: 'board', group: 'Планирование' },
		{ href: '/tasks/backlog', label: 'Бэклог', icon: 'backlog', group: 'Планирование' },
		{ href: '/projects', label: 'Проекты', icon: 'folder', group: 'Планирование' },
		{ href: '/sprints', label: 'Спринты', icon: 'sprint', group: 'Спринты' },
		{ href: '/releases', label: 'Релизы', icon: 'release', group: 'Релизы' },
		{ href: '/github', label: 'GitHub', icon: 'github', group: 'Интеграции' },
		{ href: '/confluence', label: 'Confluence', icon: 'book', group: 'Интеграции' },
	];

	// Раздел: Встречи (по макету 04-meetings.html)
	const meetingsSidebarItems = [
		{ href: '/meetings', label: 'Календарь', icon: 'calendar', group: 'Календарь' },
		{ href: '/calendar', label: 'Расписание', icon: 'schedule', group: 'Календарь' },
		{ href: '/meetings/upcoming', label: 'Предстоящие', icon: 'upcoming', group: 'Встречи' },
		{ href: '/meetings/past', label: 'Прошедшие', icon: 'past', group: 'Встречи' },
	];

	// Раздел: Service Desk (по макетам 02, 03)
	const serviceDeskSidebarItems = [
		{ href: '/service-desk', label: 'Мои заявки', icon: 'tickets', group: 'Обращения' },
		{ href: '/service-desk/create', label: 'Создать заявку', icon: 'plus', group: 'Обращения' },
	];

	// Раздел: Главная / Сотрудники / Аналитика — минимальный sidebar
	const defaultSidebarItems = [
		{ href: '/settings', label: 'Настройки', icon: 'settings', group: 'Система' },
	];

	// Determine which sidebar to show based on current path
	function getSidebarItems() {
		const path = $page.url.pathname;

		// Tasks module
		if (path.startsWith('/tasks') || path.startsWith('/projects') || path.startsWith('/sprints') || path.startsWith('/releases') || path.startsWith('/github') || path.startsWith('/confluence')) {
			return { items: tasksSidebarItems, title: 'Управление проектом' };
		}

		// Meetings module
		if (path.startsWith('/meetings') || path.startsWith('/calendar')) {
			return { items: meetingsSidebarItems, title: 'Встречи' };
		}

		// Service Desk module
		if (path.startsWith('/service-desk')) {
			return { items: serviceDeskSidebarItems, title: 'Service Desk' };
		}

		// Default (Главная, Сотрудники, Почта, Сообщения, Аналитика)
		return { items: defaultSidebarItems, title: 'Настройки' };
	}

	const sidebarConfig = $derived(getSidebarItems());
	const isAdmin = $derived(userRole === 'admin' || userRole === 'super_admin');

	// Group items by group
	const groupedItems = $derived(() => {
		const groups: Record<string, typeof sidebarConfig.items> = {};
		for (const item of sidebarConfig.items) {
			if (!groups[item.group]) {
				groups[item.group] = [];
			}
			groups[item.group].push(item);
		}
		return groups;
	});
</script>

<div class="h-full flex flex-col">
	<!-- Sidebar Title -->
	<div class="p-4 border-b border-white/10">
		<h2 class="text-sm font-medium text-gray-400 uppercase tracking-wide">{sidebarConfig.title}</h2>
	</div>

	<!-- Navigation with groups -->
	<nav class="flex-1 p-2 overflow-y-auto">
		{#each Object.entries(groupedItems()) as [groupName, items]}
			<div class="mb-4">
				<div class="px-3 py-1 text-xs font-medium text-gray-500 uppercase tracking-wider">{groupName}</div>
				<ul class="space-y-1 mt-1">
					{#each items as item}
						<li>
							<a
								href={item.href}
								class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors
									{$page.url.pathname === item.href || ($page.url.pathname.startsWith(item.href) && item.href !== '/')
										? 'bg-ekf-red text-white'
										: 'text-gray-300 hover:bg-white/10 hover:text-white'}"
							>
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									{#if item.icon === 'board'}
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
									{:else if item.icon === 'backlog'}
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
									{:else if item.icon === 'folder'}
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
									{:else if item.icon === 'sprint'}
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
									{:else if item.icon === 'release'}
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
									{:else if item.icon === 'calendar'}
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
									{:else if item.icon === 'schedule'}
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
									{:else if item.icon === 'upcoming'}
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
									{:else if item.icon === 'past'}
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
									{:else if item.icon === 'tickets'}
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z" />
									{:else if item.icon === 'plus'}
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
									{:else if item.icon === 'book'}
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
									{:else if item.icon === 'github'}
										<path fill="currentColor" d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
									{:else if item.icon === 'settings'}
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
									{/if}
								</svg>
								<span>{item.label}</span>
							</a>
						</li>
					{/each}
				</ul>
			</div>
		{/each}

		<!-- Admin link (only for admins) -->
		{#if isAdmin}
			<div class="mt-4 pt-4 border-t border-white/10">
				<a
					href="/admin"
					class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors
						{$page.url.pathname.startsWith('/admin')
							? 'bg-ekf-red text-white'
							: 'text-gray-300 hover:bg-white/10 hover:text-white'}"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
					</svg>
					<span>Админ-панель</span>
				</a>
			</div>
		{/if}
	</nav>

	<!-- Убрано дублирование профиля пользователя (GAP-012) -->
	<!-- Профиль теперь только в GlobalNav (top-bar) -->
</div>
