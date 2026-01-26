<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { improvements, auth, type ImprovementRequestType } from '$lib/api/client';

	let loading = $state(false);
	let error = $state<string | null>(null);
	let currentUserId = $state<string | null>(null);
	let types = $state<ImprovementRequestType[]>([]);

	// Form data
	let title = $state('');
	let description = $state('');
	let businessValue = $state('');
	let expectedEffect = $state('');
	let typeId = $state<string | null>(null);
	let priority = $state('medium');
	let estimatedBudget = $state<number | null>(null);
	let estimatedStart = $state('');
	let estimatedEnd = $state('');

	onMount(async () => {
		try {
			const [user, typesData] = await Promise.all([
				auth.getMe(),
				improvements.getTypes()
			]);
			currentUserId = user.id;
			types = typesData;
		} catch (e) {
			error = 'Не удалось получить данные';
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
			const request = await improvements.create({
				title: title.trim(),
				description: description.trim() || undefined,
				business_value: businessValue.trim() || undefined,
				expected_effect: expectedEffect.trim() || undefined,
				initiator_id: currentUserId,
				type_id: typeId || undefined,
				priority,
				estimated_budget: estimatedBudget || undefined,
				estimated_start: estimatedStart || undefined,
				estimated_end: estimatedEnd || undefined,
			});

			goto(`/improvements/${request.id}`);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Не удалось создать заявку';
		} finally {
			loading = false;
		}
	}

	const priorities = [
		{ value: 'low', label: 'Низкий', description: 'Не срочно', color: 'bg-green-500' },
		{ value: 'medium', label: 'Средний', description: 'Обычный', color: 'bg-yellow-500' },
		{ value: 'high', label: 'Высокий', description: 'Важно', color: 'bg-orange-500' },
		{ value: 'critical', label: 'Критический', description: 'Срочно', color: 'bg-red-500' },
	];
</script>

<svelte:head>
	<title>Новая заявка на улучшение - EKF Hub</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 py-8">
	<div class="max-w-2xl mx-auto px-6">
		<!-- Header -->
		<div class="mb-6">
			<a href="/improvements" class="text-sm text-gray-500 hover:text-ekf-red flex items-center gap-1 mb-2">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
				</svg>
				Назад к заявкам
			</a>
			<h1 class="text-2xl font-bold text-gray-900">Новая заявка на улучшение</h1>
			<p class="text-gray-500 mt-1">Опишите вашу идею или инициативу</p>
		</div>

		{#if error}
			<div class="bg-red-50 text-red-700 p-4 rounded-xl mb-6">{error}</div>
		{/if}

		<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-6">
			<!-- Title & Description -->
			<div class="bg-white rounded-xl p-5 border border-gray-200 space-y-4">
				<div>
					<label for="title" class="block text-sm font-medium text-gray-700 mb-1">
						Название <span class="text-red-500">*</span>
					</label>
					<input
						type="text"
						id="title"
						bind:value={title}
						placeholder="Краткое название улучшения"
						class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
						required
					/>
				</div>

				<div>
					<label for="description" class="block text-sm font-medium text-gray-700 mb-1">Описание проблемы/потребности</label>
					<textarea
						id="description"
						bind:value={description}
						placeholder="Подробное описание: что не устраивает, что хотелось бы изменить"
						rows="4"
						class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
					></textarea>
				</div>
			</div>

			<!-- Business Value -->
			<div class="bg-white rounded-xl p-5 border border-gray-200 space-y-4">
				<h3 class="font-medium text-gray-900">Бизнес-обоснование</h3>

				<div>
					<label for="businessValue" class="block text-sm font-medium text-gray-700 mb-1">Бизнес-ценность</label>
					<textarea
						id="businessValue"
						bind:value={businessValue}
						placeholder="Какую проблему бизнеса решает? Кому это нужно?"
						rows="3"
						class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
					></textarea>
				</div>

				<div>
					<label for="expectedEffect" class="block text-sm font-medium text-gray-700 mb-1">Ожидаемый эффект</label>
					<textarea
						id="expectedEffect"
						bind:value={expectedEffect}
						placeholder="Какие KPI улучшатся? Какая экономия/выгода?"
						rows="3"
						class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
					></textarea>
				</div>
			</div>

			<!-- Type & Priority -->
			<div class="bg-white rounded-xl p-5 border border-gray-200 space-y-4">
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label for="type" class="block text-sm font-medium text-gray-700 mb-1">Тип улучшения</label>
						<select
							id="type"
							bind:value={typeId}
							class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
						>
							<option value={null}>Не выбран</option>
							{#each types as type}
								<option value={type.id}>{type.name}</option>
							{/each}
						</select>
					</div>

					<div>
						<label for="budget" class="block text-sm font-medium text-gray-700 mb-1">Предварительный бюджет (руб)</label>
						<input
							type="number"
							id="budget"
							bind:value={estimatedBudget}
							placeholder="0"
							min="0"
							step="1000"
							class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
						/>
					</div>
				</div>

				<div>
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
			</div>

			<!-- Timeline -->
			<div class="bg-white rounded-xl p-5 border border-gray-200">
				<h3 class="font-medium text-gray-900 mb-4">Сроки реализации (ориентировочно)</h3>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label for="startDate" class="block text-sm font-medium text-gray-700 mb-1">Планируемое начало</label>
						<input
							type="date"
							id="startDate"
							bind:value={estimatedStart}
							class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
						/>
					</div>
					<div>
						<label for="endDate" class="block text-sm font-medium text-gray-700 mb-1">Планируемое завершение</label>
						<input
							type="date"
							id="endDate"
							bind:value={estimatedEnd}
							class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
						/>
					</div>
				</div>
			</div>

			<!-- Submit -->
			<div class="flex gap-3">
				<a
					href="/improvements"
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
						Создать черновик
					{/if}
				</button>
			</div>

			<p class="text-sm text-gray-500 text-center">
				Заявка будет создана как черновик. После заполнения всех полей вы сможете отправить её на рассмотрение.
			</p>
		</form>
	</div>
</div>
