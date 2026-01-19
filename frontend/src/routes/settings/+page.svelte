<script lang="ts">
	import { onMount } from 'svelte';
	import { connector as connectorApi } from '$lib/api/client';
	import type { ConnectorStatus } from '$lib/api/client';

	let status: ConnectorStatus | null = $state(null);
	let loading = $state(true);

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
</script>

<svelte:head>
	<title>Настройки - EKF Hub</title>
</svelte:head>

<div class="max-w-4xl mx-auto space-y-4">
	<h1 class="text-xl font-bold text-gray-900">Настройки</h1>

	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
		</div>
	{:else}
		<!-- Integrations -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Интеграции</h2>

			<div class="space-y-3">
				<div class="flex items-center justify-between p-4 rounded-lg border border-gray-100 hover:border-gray-200 transition-colors">
					<div class="flex items-center gap-4">
						<div class="w-10 h-10 rounded-lg bg-blue-50 flex items-center justify-center">
							<svg class="w-5 h-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
							</svg>
						</div>
						<div>
							<div class="font-medium text-gray-900">Exchange Calendar</div>
							<div class="text-sm text-gray-500">Синхронизация с корпоративным календарём</div>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<div class="w-2.5 h-2.5 rounded-full {status?.ews_configured ? 'bg-green-500' : 'bg-gray-300'}"></div>
						<span class="text-sm text-gray-600">{status?.ews_configured ? 'Активно' : 'Не настроено'}</span>
					</div>
				</div>

				<div class="flex items-center justify-between p-4 rounded-lg border border-gray-100 hover:border-gray-200 transition-colors">
					<div class="flex items-center gap-4">
						<div class="w-10 h-10 rounded-lg bg-green-50 flex items-center justify-center">
							<svg class="w-5 h-5 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
							</svg>
						</div>
						<div>
							<div class="font-medium text-gray-900">OpenAI Whisper</div>
							<div class="text-sm text-gray-500">Транскрибация аудио встреч</div>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<div class="w-2.5 h-2.5 rounded-full {status?.openai_configured ? 'bg-green-500' : 'bg-gray-300'}"></div>
						<span class="text-sm text-gray-600">{status?.openai_configured ? 'Активно' : 'Не настроено'}</span>
					</div>
				</div>

				<div class="flex items-center justify-between p-4 rounded-lg border border-gray-100 hover:border-gray-200 transition-colors">
					<div class="flex items-center gap-4">
						<div class="w-10 h-10 rounded-lg bg-orange-50 flex items-center justify-center">
							<svg class="w-5 h-5 text-orange-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
							</svg>
						</div>
						<div>
							<div class="font-medium text-gray-900">Claude AI</div>
							<div class="text-sm text-gray-500">Анализ и резюмирование встреч</div>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<div class="w-2.5 h-2.5 rounded-full {status?.anthropic_configured ? 'bg-green-500' : 'bg-gray-300'}"></div>
						<span class="text-sm text-gray-600">{status?.anthropic_configured ? 'Активно' : 'Не настроено'}</span>
					</div>
				</div>

				<div class="flex items-center justify-between p-4 rounded-lg border border-gray-100 hover:border-gray-200 transition-colors">
					<div class="flex items-center gap-4">
						<div class="w-10 h-10 rounded-lg bg-yellow-50 flex items-center justify-center">
							<svg class="w-5 h-5 text-yellow-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129" />
							</svg>
						</div>
						<div>
							<div class="font-medium text-gray-900">Yandex SpeechKit</div>
							<div class="text-sm text-gray-500">Альтернативная транскрибация (русский язык)</div>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<div class="w-2.5 h-2.5 rounded-full {status?.yandex_configured ? 'bg-green-500' : 'bg-gray-300'}"></div>
						<span class="text-sm text-gray-600">{status?.yandex_configured ? 'Активно' : 'Не настроено'}</span>
					</div>
				</div>
			</div>
		</div>

		<!-- System Info -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Информация о системе</h2>

			<div class="grid grid-cols-2 gap-4">
				<div class="p-4 rounded-lg bg-gray-50">
					<div class="text-sm text-gray-500 mb-1">Сотрудников в системе</div>
					<div class="text-2xl font-bold text-gray-900">{status?.employee_count || 0}</div>
				</div>
				<div class="p-4 rounded-lg bg-gray-50">
					<div class="text-sm text-gray-500 mb-1">Календарь</div>
					<div class="text-lg font-medium text-gray-900">{status?.calendar_integration === 'ews' ? 'Exchange EWS' : 'Локальный'}</div>
				</div>
			</div>
		</div>
	{/if}
</div>
