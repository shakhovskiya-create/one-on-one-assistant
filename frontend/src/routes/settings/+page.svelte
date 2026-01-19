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

<div class="max-w-4xl mx-auto space-y-6">
	<h1 class="text-2xl font-bold text-gray-900">Настройки системы</h1>

	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
		</div>
	{:else}
		<!-- System Stats -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Статистика системы</h2>

			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				<div class="p-4 rounded-lg bg-blue-50 border border-blue-100">
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-lg bg-blue-500 flex items-center justify-center">
							<svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
							</svg>
						</div>
						<div>
							<div class="text-2xl font-bold text-blue-700">{status?.employee_count || 0}</div>
							<div class="text-sm text-blue-600">Сотрудников</div>
						</div>
					</div>
				</div>

				<div class="p-4 rounded-lg bg-purple-50 border border-purple-100">
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-lg bg-purple-500 flex items-center justify-center">
							<svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
							</svg>
						</div>
						<div>
							<div class="text-2xl font-bold text-purple-700">{status?.meetings_count || 0}</div>
							<div class="text-sm text-purple-600">Встреч</div>
						</div>
					</div>
				</div>

				<div class="p-4 rounded-lg bg-green-50 border border-green-100">
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-lg bg-green-500 flex items-center justify-center">
							<svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
							</svg>
						</div>
						<div>
							<div class="text-2xl font-bold text-green-700">{status?.tasks_count || 0}</div>
							<div class="text-sm text-green-600">Задач</div>
						</div>
					</div>
				</div>

				<div class="p-4 rounded-lg bg-amber-50 border border-amber-100">
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-lg bg-amber-500 flex items-center justify-center">
							<svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
							</svg>
						</div>
						<div>
							<div class="text-2xl font-bold text-amber-700">{status?.projects_count || 0}</div>
							<div class="text-sm text-amber-600">Проектов</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Task breakdown -->
			<div class="mt-4 grid grid-cols-2 gap-4">
				<div class="p-3 rounded-lg bg-gray-50 flex items-center justify-between">
					<span class="text-sm text-gray-600">Открытых задач</span>
					<span class="font-semibold text-gray-900">{status?.open_tasks_count || 0}</span>
				</div>
				<div class="p-3 rounded-lg bg-red-50 flex items-center justify-between">
					<span class="text-sm text-red-600">Просроченных задач</span>
					<span class="font-semibold text-red-700">{status?.overdue_tasks_count || 0}</span>
				</div>
			</div>
		</div>

		<!-- System Status -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Состояние системы</h2>

			<div class="space-y-3">
				<div class="flex items-center justify-between p-4 rounded-lg border border-gray-100">
					<div class="flex items-center gap-4">
						<div class="w-10 h-10 rounded-lg bg-gray-50 flex items-center justify-center">
							<svg class="w-5 h-5 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
							</svg>
						</div>
						<div>
							<div class="font-medium text-gray-900">База данных</div>
							<div class="text-sm text-gray-500">PostgreSQL</div>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<div class="w-2.5 h-2.5 rounded-full {status?.db_status === 'connected' ? 'bg-green-500' : 'bg-red-500'}"></div>
						<span class="text-sm {status?.db_status === 'connected' ? 'text-green-600' : 'text-red-600'}">
							{status?.db_status === 'connected' ? 'Подключена' : 'Отключена'}
						</span>
					</div>
				</div>

				<div class="flex items-center justify-between p-4 rounded-lg border border-gray-100">
					<div class="flex items-center gap-4">
						<div class="w-10 h-10 rounded-lg bg-gray-50 flex items-center justify-center">
							<svg class="w-5 h-5 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
							</svg>
						</div>
						<div>
							<div class="font-medium text-gray-900">Active Directory</div>
							<div class="text-sm text-gray-500">Корпоративный каталог</div>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<div class="w-2.5 h-2.5 rounded-full {status?.ad_status === 'connected' ? 'bg-green-500' : 'bg-gray-300'}"></div>
						<span class="text-sm text-gray-600">
							{status?.ad_status === 'connected' ? 'Подключен' : status?.ad_status === 'disconnected' ? 'Отключен' : 'Не настроен'}
						</span>
					</div>
				</div>
			</div>
		</div>

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
							{#if status?.ews_url}
								<div class="text-xs text-gray-400 mt-1 font-mono">{status.ews_url}</div>
							{/if}
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

		<!-- Version Info -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Информация о версии</h2>

			<div class="grid grid-cols-2 gap-4">
				<div class="p-3 rounded-lg bg-gray-50">
					<div class="text-sm text-gray-500">Версия</div>
					<div class="font-medium text-gray-900">1.0.0</div>
				</div>
				<div class="p-3 rounded-lg bg-gray-50">
					<div class="text-sm text-gray-500">Тип календаря</div>
					<div class="font-medium text-gray-900">{status?.calendar_integration === 'ews' ? 'Exchange EWS' : 'Локальный'}</div>
				</div>
			</div>
		</div>
	{/if}
</div>
