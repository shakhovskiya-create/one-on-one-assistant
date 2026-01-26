<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { serviceDesk, auth } from '$lib/api/client';

	let loading = $state(false);
	let error = $state<string | null>(null);
	let currentUserId = $state<string | null>(null);

	// Form data
	let ticketType = $state('service_request');
	let title = $state('');
	let description = $state('');
	let priority = $state('medium');
	let impact = $state('individual');
	let categoryId = $state<string | null>(null);

	// Get URL params
	const urlType = $page.url.searchParams.get('type');
	const urlCategory = $page.url.searchParams.get('category');
	const urlTemplate = $page.url.searchParams.get('template');

	// Apply URL params
	if (urlType) {
		ticketType = urlType;
	}
	if (urlCategory) {
		categoryId = urlCategory;
	}
	if (urlTemplate) {
		switch (urlTemplate) {
			case 'login':
				title = 'Не могу войти в систему';
				priority = 'high';
				break;
			case 'app_error':
				title = 'Ошибка в приложении';
				priority = 'high';
				break;
		}
	}

	onMount(async () => {
		try {
			const user = await auth.getMe();
			currentUserId = user.id;
		} catch (e) {
			error = 'Не удалось получить данные пользователя';
		}
	});

	async function handleSubmit() {
		if (!title.trim()) {
			error = 'Введите название заявки';
			return;
		}
		if (!currentUserId) {
			error = 'Не удалось определить пользователя';
			return;
		}

		loading = true;
		error = null;

		try {
			const ticket = await serviceDesk.createTicket({
				type: ticketType,
				title: title.trim(),
				description: description.trim() || undefined,
				priority,
				impact,
				category_id: categoryId || undefined,
				requester_id: currentUserId,
			});

			goto(`/service-desk/tickets/${ticket.id}`);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Не удалось создать заявку';
		} finally {
			loading = false;
		}
	}

	const ticketTypes = [
		{ value: 'incident', label: 'Инцидент', description: 'Что-то не работает' },
		{ value: 'service_request', label: 'Запрос на услугу', description: 'Нужен доступ, ПО и т.д.' },
		{ value: 'change', label: 'Запрос на изменение', description: 'Изменить конфигурацию' },
	];

	const priorities = [
		{ value: 'low', label: 'Низкий', description: '72ч SLA', color: 'bg-green-500' },
		{ value: 'medium', label: 'Средний', description: '24ч SLA', color: 'bg-yellow-500' },
		{ value: 'high', label: 'Высокий', description: '8ч SLA', color: 'bg-orange-500' },
		{ value: 'critical', label: 'Критический', description: '4ч SLA', color: 'bg-red-500' },
	];

	const impacts = [
		{ value: 'individual', label: 'Один сотрудник', description: 'Проблема затрагивает только меня' },
		{ value: 'department', label: 'Отдел', description: 'Проблема затрагивает весь отдел' },
		{ value: 'organization', label: 'Вся организация', description: 'Проблема затрагивает всю компанию' },
	];
</script>

<svelte:head>
	<title>Новая заявка - Service Desk - EKF Hub</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 py-8">
	<div class="max-w-2xl mx-auto px-6">
		<!-- Header -->
		<div class="mb-6">
			<a href="/service-desk" class="text-sm text-gray-500 hover:text-ekf-red flex items-center gap-1 mb-2">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
				</svg>
				Назад к Service Desk
			</a>
			<h1 class="text-2xl font-bold text-gray-900">Новая заявка</h1>
			<p class="text-gray-500 mt-1">Заполните форму для создания обращения</p>
		</div>

		{#if error}
			<div class="bg-red-50 text-red-700 p-4 rounded-xl mb-6">{error}</div>
		{/if}

		<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-6">
			<!-- Ticket Type -->
			<div class="bg-white rounded-xl p-5 border border-gray-200">
				<label class="block text-sm font-medium text-gray-700 mb-3">Тип обращения</label>
				<div class="grid grid-cols-3 gap-3">
					{#each ticketTypes as type}
						<button
							type="button"
							onclick={() => ticketType = type.value}
							class="p-3 rounded-lg border-2 text-left transition-all {ticketType === type.value ? 'border-ekf-red bg-red-50' : 'border-gray-200 hover:border-gray-300'}"
						>
							<div class="font-medium text-sm {ticketType === type.value ? 'text-ekf-red' : 'text-gray-900'}">{type.label}</div>
							<div class="text-xs text-gray-500">{type.description}</div>
						</button>
					{/each}
				</div>
			</div>

			<!-- Title & Description -->
			<div class="bg-white rounded-xl p-5 border border-gray-200 space-y-4">
				<div>
					<label for="title" class="block text-sm font-medium text-gray-700 mb-1">
						Тема <span class="text-red-500">*</span>
					</label>
					<input
						type="text"
						id="title"
						bind:value={title}
						placeholder="Кратко опишите проблему или запрос"
						class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
						required
					>
				</div>

				<div>
					<label for="description" class="block text-sm font-medium text-gray-700 mb-1">Описание</label>
					<textarea
						id="description"
						bind:value={description}
						placeholder="Подробное описание (что произошло, когда, что пробовали сделать)"
						rows="4"
						class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
					></textarea>
				</div>
			</div>

			<!-- Priority -->
			<div class="bg-white rounded-xl p-5 border border-gray-200">
				<label class="block text-sm font-medium text-gray-700 mb-3">Приоритет</label>
				<div class="grid grid-cols-4 gap-2">
					{#each priorities as p}
						<button
							type="button"
							onclick={() => priority = p.value}
							class="p-3 rounded-lg border-2 text-center transition-all {priority === p.value ? 'border-ekf-red bg-red-50' : 'border-gray-200 hover:border-gray-300'}"
						>
							<div class="flex items-center justify-center gap-2 mb-1">
								<div class="w-2 h-2 rounded-full {p.color}"></div>
								<span class="font-medium text-sm {priority === p.value ? 'text-ekf-red' : 'text-gray-900'}">{p.label}</span>
							</div>
							<div class="text-xs text-gray-500">{p.description}</div>
						</button>
					{/each}
				</div>
			</div>

			<!-- Impact (for incidents) -->
			{#if ticketType === 'incident'}
				<div class="bg-white rounded-xl p-5 border border-gray-200">
					<label class="block text-sm font-medium text-gray-700 mb-3">Масштаб влияния</label>
					<div class="space-y-2">
						{#each impacts as imp}
							<button
								type="button"
								onclick={() => impact = imp.value}
								class="w-full p-3 rounded-lg border-2 text-left transition-all flex items-center gap-3 {impact === imp.value ? 'border-ekf-red bg-red-50' : 'border-gray-200 hover:border-gray-300'}"
							>
								<div class="w-4 h-4 rounded-full border-2 flex items-center justify-center {impact === imp.value ? 'border-ekf-red' : 'border-gray-300'}">
									{#if impact === imp.value}
										<div class="w-2 h-2 rounded-full bg-ekf-red"></div>
									{/if}
								</div>
								<div>
									<div class="font-medium text-sm {impact === imp.value ? 'text-ekf-red' : 'text-gray-900'}">{imp.label}</div>
									<div class="text-xs text-gray-500">{imp.description}</div>
								</div>
							</button>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Submit -->
			<div class="flex gap-3">
				<a
					href="/service-desk"
					class="flex-1 py-3 px-6 border border-gray-300 rounded-xl text-center font-medium text-gray-700 hover:bg-gray-50 transition-colors"
				>
					Отмена
				</a>
				<button
					type="submit"
					disabled={loading || !title.trim()}
					class="flex-1 py-3 px-6 bg-ekf-red text-white rounded-xl font-medium hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
				>
					{#if loading}
						<div class="animate-spin w-5 h-5 border-2 border-white border-t-transparent rounded-full"></div>
						Создание...
					{:else}
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
						</svg>
						Создать заявку
					{/if}
				</button>
			</div>
		</form>
	</div>
</div>
