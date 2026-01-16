<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';

	let username = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

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
			goto('/');
		} else {
			error = result.error || 'Ошибка авторизации';
		}

		loading = false;
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
					placeholder="domain\\username или email"
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
