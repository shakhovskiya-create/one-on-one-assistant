<script lang="ts">
	import { page } from '$app/stores';
	import { sidebarOpen, currentUser } from '$lib/stores/app';

	const menuItems = [
		{ href: '/', label: 'Дашборд', icon: 'dashboard' },
		{ href: '/employees', label: 'Сотрудники', icon: 'users' },
		{ href: '/projects', label: 'Проекты', icon: 'folder' },
		{ href: '/meetings', label: 'Встречи', icon: 'calendar' },
		{ href: '/tasks', label: 'Задачи', icon: 'tasks' },
		{ href: '/calendar', label: 'Календарь', icon: 'schedule' },
		{ href: '/analytics', label: 'Аналитика', icon: 'chart' },
		{ href: '/settings', label: 'Настройки', icon: 'settings' },
	];
</script>

<aside class="bg-ekf-dark text-white w-64 min-h-screen flex flex-col {$sidebarOpen ? '' : 'hidden lg:flex'}">
	<!-- Logo -->
	<div class="p-4 border-b border-gray-700">
		<div class="flex items-center gap-3">
			<div class="bg-ekf-red text-white font-bold text-xl px-3 py-1 rounded">
				EKF
			</div>
			<span class="text-lg font-semibold">Team Hub</span>
		</div>
	</div>

	<!-- Navigation -->
	<nav class="flex-1 p-4">
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
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
							{:else if item.icon === 'chart'}
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
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

	<!-- User Info -->
	{#if $currentUser}
		<div class="p-4 border-t border-gray-700">
			<div class="flex items-center gap-3">
				<div class="w-10 h-10 rounded-full bg-gray-600 flex items-center justify-center">
					{$currentUser.name.charAt(0)}
				</div>
				<div class="flex-1 min-w-0">
					<div class="text-sm font-medium truncate">{$currentUser.name}</div>
					<div class="text-xs text-gray-400 truncate">{$currentUser.position}</div>
				</div>
			</div>
		</div>
	{/if}
</aside>
