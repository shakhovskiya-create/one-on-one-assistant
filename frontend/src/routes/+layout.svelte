<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth, user, isAuthenticated, isLoading, subordinates } from '$lib/stores/auth';
	import { notifications, currentUser } from '$lib/stores/app';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Header from '$lib/components/Header.svelte';

	let { children } = $props();

	// Pages that don't require auth
	const publicPages = ['/login'];

	// Sync auth user to app store for backwards compatibility
	$effect(() => {
		if ($user) {
			currentUser.set($user);
		}
	});

	$effect(() => {
		// Wait for auth to finish loading
		if ($isLoading) return;

		const isPublicPage = publicPages.includes($page.url.pathname);

		if (!$isAuthenticated && !isPublicPage) {
			goto('/login');
		}
	});

	function handleLogout() {
		auth.logout();
		goto('/login');
	}
</script>

{#if $isLoading}
	<div class="min-h-screen flex items-center justify-center bg-gray-100">
		<div class="text-center">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-ekf-red mx-auto mb-4"></div>
			<p class="text-gray-500">Загрузка...</p>
		</div>
	</div>
{:else if $page.url.pathname === '/login'}
	{@render children()}
{:else if $isAuthenticated}
	<div class="flex min-h-screen">
		<Sidebar user={$user} subordinates={$subordinates} onLogout={handleLogout} />

		<div class="flex-1 flex flex-col">
			<Header user={$user} />

			<main class="flex-1 p-6 bg-gray-50">
				{@render children()}
			</main>
		</div>
	</div>

	<!-- Notifications -->
	<div class="fixed bottom-4 right-4 space-y-2 z-50">
		{#each $notifications as notification (notification.id)}
			<div
				class="px-4 py-3 rounded-lg shadow-lg flex items-center gap-3 min-w-[300px]
					{notification.type === 'success' ? 'bg-green-500 text-white' : ''}
					{notification.type === 'error' ? 'bg-red-500 text-white' : ''}
					{notification.type === 'info' ? 'bg-blue-500 text-white' : ''}
					{notification.type === 'warning' ? 'bg-yellow-500 text-white' : ''}"
			>
				<span class="flex-1">{notification.message}</span>
				<button
					onclick={() => notifications.remove(notification.id)}
					class="opacity-70 hover:opacity-100"
				>
					&times;
				</button>
			</div>
		{/each}
	</div>
{:else}
	<!-- Redirecting to login -->
	<div class="min-h-screen flex items-center justify-center bg-gray-100">
		<div class="text-gray-500">Перенаправление...</div>
	</div>
{/if}
