<script lang="ts">
	import '../app.css';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Header from '$lib/components/Header.svelte';
	import { notifications } from '$lib/stores/app';

	let { children } = $props();
</script>

<div class="flex min-h-screen">
	<Sidebar />

	<div class="flex-1 flex flex-col">
		<Header />

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
