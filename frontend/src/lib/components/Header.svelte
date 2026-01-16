<script lang="ts">
	import { sidebarOpen, notifications } from '$lib/stores/app';
	import type { User } from '$lib/stores/auth';

	interface Props {
		user?: User | null;
		subordinates?: User[];
		title?: string;
		onProfileClick?: () => void;
		onLogout?: () => void;
	}

	let { user = null, subordinates = [], title = 'EKF Team Hub', onProfileClick, onLogout }: Props = $props();

	let showUserMenu = $state(false);

	function toggleSidebar() {
		sidebarOpen.update(v => !v);
	}

	function handleProfileClick() {
		if (onProfileClick) {
			onProfileClick();
		}
		showUserMenu = false;
	}

	function handleLogout() {
		if (onLogout) {
			onLogout();
		}
		showUserMenu = false;
	}
</script>

<header class="bg-white shadow-sm border-b border-gray-200 px-4 py-3">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-4">
			<button
				onclick={toggleSidebar}
				class="lg:hidden p-2 rounded-lg hover:bg-gray-100"
			>
				<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</button>
			<h1 class="text-xl font-semibold text-gray-800">{title}</h1>
		</div>

		<div class="flex items-center gap-4">
			<!-- Notifications -->
			<button class="p-2 rounded-lg hover:bg-gray-100 relative">
				<svg class="w-6 h-6 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
				</svg>
			</button>

			<!-- User Menu -->
			{#if user}
				<div class="relative">
					<button
						onclick={() => showUserMenu = !showUserMenu}
						class="flex items-center gap-2 hover:bg-gray-100 rounded-lg px-2 py-1 transition-colors cursor-pointer"
					>
						<div class="text-right hidden sm:block">
							<div class="text-sm text-gray-700 font-medium">{user.name}</div>
							{#if subordinates.length > 0}
								<div class="text-xs text-gray-400">{subordinates.length} подчинённых</div>
							{/if}
						</div>
						{#if user.photo_base64}
							<img src="data:image/jpeg;base64,{user.photo_base64}" alt="" class="w-8 h-8 rounded-full object-cover" />
						{:else}
							<div class="w-8 h-8 rounded-full bg-ekf-red text-white flex items-center justify-center text-sm font-medium">
								{user.name.charAt(0)}
							</div>
						{/if}
					</button>

					{#if showUserMenu}
						<div class="absolute right-0 top-full mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50">
							<button
								onclick={handleProfileClick}
								class="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 flex items-center gap-2"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
								</svg>
								Мой профиль
							</button>
							{#if onLogout}
								<button
									onclick={handleLogout}
									class="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 flex items-center gap-2"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
									</svg>
									Выйти
								</button>
							{/if}
						</div>
					{/if}
				</div>
			{:else}
				<a href="/login" class="text-sm text-ekf-red hover:underline">Войти</a>
			{/if}
		</div>
	</div>
</header>
