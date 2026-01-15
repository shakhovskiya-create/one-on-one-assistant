<script lang="ts">
	import { onMount } from 'svelte';
	import { connector as connectorApi } from '$lib/api/client';
	import type { ConnectorStatus } from '$lib/api/client';

	let status: ConnectorStatus | null = $state(null);
	let loading = $state(true);
	let syncing = $state(false);
	let credentials = $state({ username: '', password: '' });

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

	async function authenticate() {
		if (!credentials.username || !credentials.password) {
			alert('Введите учётные данные');
			return;
		}
		try {
			const result = await connectorApi.authenticate(credentials.username, credentials.password);
			if (result.success) {
				alert('Аутентификация успешна');
			} else {
				alert('Ошибка аутентификации');
			}
		} catch (e) {
			console.error(e);
			alert('Ошибка подключения');
		}
	}
</script>

<svelte:head>
	<title>Настройки - EKF Team Hub</title>
</svelte:head>

<div class="max-w-4xl mx-auto space-y-6">
	<h1 class="text-2xl font-bold text-gray-900">Настройки</h1>

	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="text-gray-500">Загрузка...</div>
		</div>
	{:else}
		<!-- Connector Status -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Статус коннектора</h2>

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

		<!-- AD Authentication -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Аутентификация AD</h2>
			<p class="text-gray-600 text-sm mb-4">
				Введите учётные данные Active Directory для проверки подключения.
			</p>

			<div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Логин</label>
					<input
						type="text"
						bind:value={credentials.username}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						placeholder="domain\\username"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Пароль</label>
					<input
						type="password"
						bind:value={credentials.password}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
					/>
				</div>
			</div>

			<button
				onclick={authenticate}
				class="px-4 py-2 bg-gray-800 text-white rounded-lg hover:bg-gray-900 transition-colors"
			>
				Проверить подключение
			</button>
		</div>

		<!-- API Settings -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4">API настройки</h2>

			<div class="space-y-4">
				<div class="flex items-center justify-between p-4 rounded-lg bg-gray-50">
					<div>
						<div class="font-medium text-gray-900">OpenAI API</div>
						<div class="text-sm text-gray-500">Для транскрибации Whisper</div>
					</div>
					<div class="flex items-center gap-2">
						<div class="w-3 h-3 rounded-full bg-green-500"></div>
						<span class="text-sm text-gray-600">Настроено</span>
					</div>
				</div>

				<div class="flex items-center justify-between p-4 rounded-lg bg-gray-50">
					<div>
						<div class="font-medium text-gray-900">Anthropic API</div>
						<div class="text-sm text-gray-500">Для анализа встреч (Claude)</div>
					</div>
					<div class="flex items-center gap-2">
						<div class="w-3 h-3 rounded-full bg-green-500"></div>
						<span class="text-sm text-gray-600">Настроено</span>
					</div>
				</div>

				<div class="flex items-center justify-between p-4 rounded-lg bg-gray-50">
					<div>
						<div class="font-medium text-gray-900">Yandex SpeechKit</div>
						<div class="text-sm text-gray-500">Альтернативная транскрибация</div>
					</div>
					<div class="flex items-center gap-2">
						<div class="w-3 h-3 rounded-full bg-gray-400"></div>
						<span class="text-sm text-gray-600">Не настроено</span>
					</div>
				</div>

				<div class="flex items-center justify-between p-4 rounded-lg bg-gray-50">
					<div>
						<div class="font-medium text-gray-900">Exchange EWS</div>
						<div class="text-sm text-gray-500">Интеграция с календарём</div>
					</div>
					<div class="flex items-center gap-2">
						<div class="w-3 h-3 rounded-full {status?.ews_configured ? 'bg-green-500' : 'bg-gray-400'}"></div>
						<span class="text-sm text-gray-600">{status?.ews_configured ? 'Настроено' : 'Не настроено'}</span>
					</div>
				</div>
			</div>
		</div>

		<!-- About -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4">О приложении</h2>
			<div class="space-y-2 text-gray-600">
				<p><strong>EKF Team Hub</strong> - система управления встречами 1-на-1</p>
				<p>Версия: 2.0.0 (SvelteKit + Go)</p>
				<p>© 2024 EKF</p>
			</div>
		</div>
	{/if}
</div>
