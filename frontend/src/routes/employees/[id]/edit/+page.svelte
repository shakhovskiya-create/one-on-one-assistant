<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { employees as employeesApi } from '$lib/api/client';
	import type { Employee } from '$lib/api/client';

	let employee: Employee | null = $state(null);
	let loading = $state(true);
	let saving = $state(false);

	const id = $page.params.id;

	onMount(async () => {
		try {
			employee = await employeesApi.get(id);
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	async function saveEmployee() {
		if (!employee) return;
		saving = true;
		try {
			await employeesApi.update(id, {
				name: employee.name,
				position: employee.position,
				department: employee.department,
				email: employee.email,
				phone: employee.phone,
				telegram_username: employee.telegram_username
			});
			await goto(`/employees/${id}`);
		} catch (e) {
			console.error(e);
			alert('Ошибка сохранения');
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Редактирование - {employee?.name || 'Сотрудник'} - EKF Team Hub</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center h-64">
		<div class="text-gray-500">Загрузка...</div>
	</div>
{:else if employee}
	<div class="max-w-2xl mx-auto space-y-6">
		<div class="flex items-center justify-between">
			<h1 class="text-2xl font-bold text-gray-900">Редактирование сотрудника</h1>
			<a href="/employees/{id}" class="text-gray-500 hover:text-gray-700">
				Отмена
			</a>
		</div>

		<div class="bg-white rounded-xl shadow-sm p-6 space-y-6">
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">ФИО *</label>
				<input
					type="text"
					bind:value={employee.name}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
				/>
			</div>

			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Должность *</label>
				<input
					type="text"
					bind:value={employee.position}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
				/>
			</div>

			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Отдел</label>
				<input
					type="text"
					bind:value={employee.department}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
				/>
			</div>

			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
				<input
					type="email"
					bind:value={employee.email}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
				/>
			</div>

			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Телефон</label>
				<input
					type="tel"
					bind:value={employee.phone}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
				/>
			</div>

			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Telegram</label>
				<input
					type="text"
					bind:value={employee.telegram_username}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
					placeholder="@username"
				/>
			</div>

			<div class="pt-4 flex gap-3">
				<a
					href="/employees/{id}"
					class="flex-1 px-4 py-2 text-center border border-gray-300 rounded-lg hover:bg-gray-50"
				>
					Отмена
				</a>
				<button
					onclick={saveEmployee}
					disabled={saving}
					class="flex-1 px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50"
				>
					{#if saving}
						Сохранение...
					{:else}
						Сохранить
					{/if}
				</button>
			</div>
		</div>
	</div>
{:else}
	<div class="text-center py-12">
		<div class="text-gray-400 text-lg">Сотрудник не найден</div>
	</div>
{/if}
