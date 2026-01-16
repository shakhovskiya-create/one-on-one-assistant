<script lang="ts">
	import { onMount } from 'svelte';
	import { connector as connectorApi } from '$lib/api/client';
	import type { ConnectorStatus } from '$lib/api/client';

	let status: ConnectorStatus | null = $state(null);
	let loading = $state(true);
	let syncing = $state(false);

	onMount(async () => {
		await loadStatus();
	});

	async function loadStatus() {
		try {
			status = await connectorApi.status();
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function syncAD() {
		syncing = true;
		try {
			await connectorApi.syncAD();
			alert('Синхронизация завершена');
			await loadStatus();
		} catch (e) {
			console.error(e);
			alert('Ошибка синхронизации');
		} finally {
			syncing = false;
		}
	}
</script>

<svelte:head>
	<title>Статус служб - EKF Team Hub</title>
</svelte:head>

<div class="max-w-4xl mx-auto space-y-6">
	<h1 class="text-2xl font-bold text-gray-900">Статус служб</h1>

	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="text-gray-500">Загрузка...</div>
		</div>
	{:else}
		<!-- Service Status -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
				<div class="p-4 rounded-lg bg-gray-50">
					<div class="text-sm text-gray-500 mb-1">Коннектор</div>
					<div class="flex items-center gap-2">
						<div class="w-3 h-3 rounded-full {status?.connected ? 'bg-green-500' : 'bg-red-500'}"></div>
						<span class="font-medium">{status?.connected ? 'Подключен' : 'Отключен'}</span>
					</div>
				</div>
				<div class="p-4 rounded-lg bg-gray-50">
					<div class="text-sm text-gray-500 mb-1">AD Sync</div>
					<div class="flex items-center gap-2">
						<div class="w-3 h-3 rounded-full {status?.ad_sync_enabled ? 'bg-green-500' : 'bg-gray-400'}"></div>
						<span class="font-medium">{status?.ad_sync_enabled ? 'Включено' : 'Выключено'}</span>
					</div>
				</div>
				<div class="p-4 rounded-lg bg-gray-50">
					<div class="text-sm text-gray-500 mb-1">Последняя синхронизация</div>
					<span class="font-medium">{status?.last_sync || 'Никогда'}</span>
				</div>
			</div>

			<div class="space-y-3 mb-6">
				<div class="flex items-center justify-between p-3 rounded-lg bg-gray-50">
					<div class="flex items-center gap-3">
						<div class="w-3 h-3 rounded-full bg-green-500"></div>
						<div>
							<div class="font-medium text-gray-900">OpenAI API</div>
							<div class="text-sm text-gray-500">Транскрибация Whisper</div>
						</div>
					</div>
				</div>

				<div class="flex items-center justify-between p-3 rounded-lg bg-gray-50">
					<div class="flex items-center gap-3">
						<div class="w-3 h-3 rounded-full bg-green-500"></div>
						<div>
							<div class="font-medium text-gray-900">Anthropic API</div>
							<div class="text-sm text-gray-500">Анализ встреч (Claude)</div>
						</div>
					</div>
				</div>

				<div class="flex items-center justify-between p-3 rounded-lg bg-gray-50">
					<div class="flex items-center gap-3">
						<div class="w-3 h-3 rounded-full {status?.ews_configured ? 'bg-green-500' : 'bg-gray-400'}"></div>
						<div>
							<div class="font-medium text-gray-900">Exchange EWS</div>
							<div class="text-sm text-gray-500">Интеграция с календарём</div>
						</div>
					</div>
				</div>

				<div class="flex items-center justify-between p-3 rounded-lg bg-gray-50">
					<div class="flex items-center gap-3">
						<div class="w-3 h-3 rounded-full bg-gray-400"></div>
						<div>
							<div class="font-medium text-gray-900">Yandex SpeechKit</div>
							<div class="text-sm text-gray-500">Альтернативная транскрибация</div>
						</div>
					</div>
				</div>
			</div>

			<button
				onclick={syncAD}
				disabled={syncing || !status?.connected}
				class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{#if syncing}
					Синхронизация...
				{:else}
					Синхронизировать AD
				{/if}
			</button>
		</div>
	{/if}
</div>
