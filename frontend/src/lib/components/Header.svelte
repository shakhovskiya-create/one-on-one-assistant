<script lang="ts">
	import { sidebarOpen, currentUser, notifications } from '$lib/stores/app';

	export let title = 'EKF Team Hub';

	function toggleSidebar() {
		sidebarOpen.update(v => !v);
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
			{#if $currentUser}
				<div class="flex items-center gap-2">
					<span class="text-sm text-gray-600">{$currentUser.name}</span>
					<div class="w-8 h-8 rounded-full bg-ekf-red text-white flex items-center justify-center text-sm font-medium">
						{$currentUser.name.charAt(0)}
					</div>
				</div>
			{:else}
				<a href="/login" class="text-sm text-ekf-red hover:underline">Войти</a>
			{/if}
		</div>
	</div>
</header>
