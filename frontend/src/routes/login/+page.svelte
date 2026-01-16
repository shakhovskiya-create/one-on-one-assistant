<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';

	let username = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	// Password change state
	let showPasswordChange = $state(false);
	let pendingUserId = $state('');
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordChangeError = $state('');
	let changingPassword = $state(false);

	// Redirect if already authenticated
	$effect(() => {
		if ($isAuthenticated) {
			goto('/');
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!username || !password) {
			error = 'Введите логин и пароль';
			return;
		}

		loading = true;
		error = '';

		const result = await auth.login(username, password);

		if (result.success) {
			if (result.forcePasswordChange && result.userId) {
				// Show password change dialog
				pendingUserId = result.userId;
				currentPassword = password;
				showPasswordChange = true;
				loading = false;
			} else {
				goto('/');
			}
		} else {
			error = result.error || 'Ошибка авторизации';
			loading = false;
		}
	}

	async function handlePasswordChange(e: Event) {
		e.preventDefault();
		passwordChangeError = '';

		if (!newPassword || !confirmPassword) {
			passwordChangeError = 'Заполните все поля';
			return;
		}

		if (newPassword.length < 6) {
			passwordChangeError = 'Пароль должен быть не менее 6 символов';
			return;
		}

		if (newPassword !== confirmPassword) {
			passwordChangeError = 'Пароли не совпадают';
			return;
		}

		changingPassword = true;

		const result = await auth.changePassword(pendingUserId, currentPassword, newPassword);

		if (result.success) {
			// Now login again with new password
			const loginResult = await auth.login(username, newPassword);
			if (loginResult.success && !loginResult.forcePasswordChange) {
				goto('/');
			} else {
				passwordChangeError = 'Ошибка входа после смены пароля';
			}
		} else {
			passwordChangeError = result.error || 'Не удалось изменить пароль';
		}

		changingPassword = false;
	}
</script>

<svelte:head>
	<title>Вход - EKF Team Hub</title>
</svelte:head>

<div class="min-h-screen bg-gray-100 flex items-center justify-center p-4">
	<div class="bg-white rounded-xl shadow-lg p-8 w-full max-w-md">
		<div class="text-center mb-8">
			<div class="inline-flex items-center gap-3 mb-4">
				<div class="bg-ekf-red text-white font-bold text-2xl px-4 py-2 rounded">
					EKF
				</div>
				<span class="text-2xl font-semibold text-gray-900">Team Hub</span>
			</div>
			<p class="text-gray-500">Войдите с учётными данными Active Directory</p>
		</div>

		<form onsubmit={handleSubmit} class="space-y-6">
			{#if error}
				<div class="p-4 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
					{error}
				</div>
			{/if}

			<div>
				<label class="block text-sm font-medium text-gray-700 mb-2">Логин</label>
				<input
					type="text"
					bind:value={username}
					class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
					placeholder="email или логин"
					disabled={loading}
				/>
			</div>

			<div>
				<label class="block text-sm font-medium text-gray-700 mb-2">Пароль</label>
				<input
					type="password"
					bind:value={password}
					class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
					placeholder="Пароль"
					disabled={loading}
				/>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="w-full py-3 bg-ekf-red text-white rounded-lg font-medium hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{#if loading}
					<span class="inline-flex items-center gap-2">
						<svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
						</svg>
						Вход...
					</span>
				{:else}
					Войти
				{/if}
			</button>
		</form>

		<div class="mt-6 text-center text-sm text-gray-500">
			<p>Используйте учётные данные корпоративной сети</p>
		</div>
	</div>
</div>

{#if showPasswordChange}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
		<div class="bg-white rounded-xl shadow-xl p-8 w-full max-w-md">
			<div class="text-center mb-6">
				<div class="w-16 h-16 bg-yellow-100 rounded-full flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
					</svg>
				</div>
				<h2 class="text-xl font-bold text-gray-900">Смена пароля</h2>
				<p class="text-gray-500 mt-2">Для продолжения работы необходимо сменить пароль</p>
			</div>

			<form onsubmit={handlePasswordChange} class="space-y-4">
				{#if passwordChangeError}
					<div class="p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
						{passwordChangeError}
					</div>
				{/if}

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Новый пароль</label>
					<input
						type="password"
						bind:value={newPassword}
						class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						placeholder="Минимум 6 символов"
						disabled={changingPassword}
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Подтвердите пароль</label>
					<input
						type="password"
						bind:value={confirmPassword}
						class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						placeholder="Повторите пароль"
						disabled={changingPassword}
					/>
				</div>

				<button
					type="submit"
					disabled={changingPassword}
					class="w-full py-3 bg-ekf-red text-white rounded-lg font-medium hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
				>
					{#if changingPassword}
						<span class="inline-flex items-center gap-2">
							<svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
							</svg>
							Сохранение...
						</span>
					{:else}
						Сменить пароль
					{/if}
				</button>
			</form>
		</div>
	</div>
{/if}
