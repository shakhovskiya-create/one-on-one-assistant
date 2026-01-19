<script lang="ts">
	import { page } from '$app/stores';
	import { sidebarOpen } from '$lib/stores/app';
	import type { User } from '$lib/stores/auth';

	interface Props {
		user?: User | null;
		subordinates?: User[];
		onLogout?: () => void;
	}

	let { user = null, subordinates = [], onLogout }: Props = $props();

	const menuItems = [
		{ href: '/', label: 'Дашборд', icon: 'dashboard' },
		{ href: '/employees', label: 'Сотрудники', icon: 'users' },
		{ href: '/projects', label: 'Проекты', icon: 'folder' },
		{ href: '/meetings', label: 'Встречи', icon: 'calendar' },
		{ href: '/tasks', label: 'Задачи', icon: 'tasks' },
		{ href: '/messenger', label: 'Сообщения', icon: 'messenger' },
		{ href: '/calendar', label: 'Календарь', icon: 'schedule' },
		{ href: '/analytics', label: 'Аналитика', icon: 'chart' },
	];
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
							{:else if item.icon === 'upload'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
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
		</ul>
	</nav>

</aside>
