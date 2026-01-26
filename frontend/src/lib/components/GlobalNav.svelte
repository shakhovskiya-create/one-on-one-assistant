<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import type { User } from '$lib/stores/auth';

	interface Props {
		user?: User | null;
		onLogout?: () => void;
	}

	let { user = null, onLogout }: Props = $props();
	let showUserMenu = $state(false);

	const modules = [
		{ href: '/', label: 'Дашборд', exact: true },
		{ href: '/tasks', label: 'Задачи' },
		{ href: '/meetings', label: 'Встречи' },
		{ href: '/employees', label: 'Сотрудники' },
		{ href: '/analytics', label: 'Аналитика' },
	];

	function isActive(href: string, exact: boolean = false): boolean {
		if (exact) {
			return $page.url.pathname === href;
		}
		return $page.url.pathname === href || $page.url.pathname.startsWith(href + '/');
	}

	function handleLogout() {
		if (onLogout) {
			onLogout();
		}
		showUserMenu = false;
	}
</script>

<div class="fixed top-0 left-0 right-0 z-50 bg-ekf-dark text-white h-12 flex items-center px-4">
	<!-- Logo -->
	<a href="/" class="flex items-center gap-3 mr-8">
		<div class="w-8 h-8 bg-ekf-red rounded flex items-center justify-center font-bold text-sm">E</div>
		<span class="font-semibold">EKF Hub</span>
	</a>

	<!-- Module Navigation -->
	<nav class="flex items-center gap-1">
		{#each modules as mod}
			<a
				href={mod.href}
				class="px-3 py-1.5 rounded text-sm transition-colors
					{isActive(mod.href, mod.exact)
						? 'bg-ekf-red text-white'
						: 'text-gray-300 hover:text-white hover:bg-white/10'}"
			>
				{mod.label}
			</a>
		{/each}
	</nav>

	<!-- Right Side -->
	<div class="ml-auto flex items-center gap-4">
		<!-- Notifications -->
		<button class="p-2 hover:bg-white/10 rounded-lg relative">
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
			</svg>
		</button>

		<!-- User Menu -->
		{#if user}
			<div class="relative">
				<button
					onclick={() => showUserMenu = !showUserMenu}
					class="flex items-center gap-2 hover:bg-white/10 rounded-lg px-2 py-1 transition-colors cursor-pointer"
				>
					{#if user.photo_base64}
						<img src="data:image/jpeg;base64,{user.photo_base64}" alt="" class="w-8 h-8 rounded-full object-cover" />
					{:else}
						<div class="w-8 h-8 rounded-full bg-ekf-red flex items-center justify-center text-sm font-medium">
							{user.name.charAt(0)}
						</div>
					{/if}
					<span class="text-sm hidden sm:block">{user.name.split(' ')[0]}</span>
				</button>

				{#if showUserMenu}
					<div class="absolute right-0 top-full mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50">
						<a
							href="/profile"
							onclick={() => showUserMenu = false}
							class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 flex items-center gap-2"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
							</svg>
							Мой профиль
						</a>
						<a
							href="/settings"
							onclick={() => showUserMenu = false}
							class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 flex items-center gap-2"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
							</svg>
							Настройки
						</a>
						<div class="border-t border-gray-200 my-1"></div>
						<button
							onclick={handleLogout}
							class="w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-gray-100 flex items-center gap-2"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
							</svg>
							Выйти
						</button>
					</div>
				{/if}
			</div>
		{:else}
			<a href="/login" class="text-sm text-gray-300 hover:text-white">Войти</a>
		{/if}
	</div>
</div>
