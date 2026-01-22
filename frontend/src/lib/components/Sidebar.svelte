<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { sidebarOpen } from '$lib/stores/app';
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

	const menuItems = [
		{ href: '/', label: 'Дашборд', icon: 'dashboard' },
		{ href: '/employees', label: 'Сотрудники', icon: 'users' },
		{ href: '/projects', label: 'Проекты', icon: 'folder' },
		{ href: '/meetings', label: 'Встречи', icon: 'calendar' },
		{ href: '/tasks', label: 'Задачи', icon: 'tasks' },
		{ href: '/messenger', label: 'Сообщения', icon: 'messenger' },
		{ href: '/mail', label: 'Почта', icon: 'mail' },
		{ href: '/calendar', label: 'Календарь', icon: 'schedule' },
		{ href: '/confluence', label: 'Confluence', icon: 'book' },
		{ href: '/releases', label: 'Releases', icon: 'release' },
		{ href: '/analytics', label: 'Аналитика', icon: 'chart' },
		{ href: '/settings', label: 'Настройки', icon: 'settings' },
	];

	const isAdmin = $derived(userRole === 'admin' || userRole === 'super_admin');
</script>

<aside class="bg-ekf-dark text-white w-64 h-full flex flex-col {$sidebarOpen ? '' : 'hidden lg:flex'}">
	<!-- Logo -->
	<div class="p-4 border-b border-gray-700">
		<div class="flex items-center gap-3">
			<div class="bg-ekf-red text-white font-bold text-xl px-3 py-1 rounded">
				EKF
			</div>
			<span class="text-lg font-semibold">Hub</span>
		</div>
	</div>

	<!-- Navigation -->
	<nav class="flex-1 p-4 overflow-y-auto">
		<ul class="space-y-1">
			{#each menuItems as item}
				<li>
					<a
						href={item.href}
						class="flex items-center gap-3 px-4 py-3 rounded-lg transition-colors
							{$page.url.pathname === item.href || ($page.url.pathname.startsWith(item.href) && item.href !== '/')
								? 'bg-ekf-red text-white'
								: 'text-gray-300 hover:bg-gray-700 hover:text-white'}"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							{#if item.icon === 'dashboard'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
							{:else if item.icon === 'users'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
							{:else if item.icon === 'folder'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
							{:else if item.icon === 'calendar'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
							{:else if item.icon === 'tasks'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
							{:else if item.icon === 'schedule'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
							{:else if item.icon === 'chart'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
							{:else if item.icon === 'messenger'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
							{:else if item.icon === 'mail'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
							{:else if item.icon === 'book'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
							{:else if item.icon === 'upload'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
							{:else if item.icon === 'github'}
								<path fill="currentColor" d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
							{:else if item.icon === 'release'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
							{:else if item.icon === 'script'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
							{:else if item.icon === 'settings'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
							{/if}
						</svg>
						<span>{item.label}</span>
					</a>
				</li>
			{/each}

			<!-- Admin link (only for admins) -->
			{#if isAdmin}
				<li class="mt-4 pt-4 border-t border-gray-700">
					<a
						href="/admin"
						class="flex items-center gap-3 px-4 py-3 rounded-lg transition-colors
							{$page.url.pathname.startsWith('/admin')
								? 'bg-ekf-red text-white'
								: 'text-gray-300 hover:bg-gray-700 hover:text-white'}"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
						</svg>
						<span>Админ-панель</span>
					</a>
				</li>
			{/if}
		</ul>
	</nav>

</aside>
