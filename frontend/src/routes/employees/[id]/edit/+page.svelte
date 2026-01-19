<script lang="ts">
	import { page } from '$app/stores';

	const id = $page.params.id;
</script>

<svelte:head>
	<title>Редактировать сотрудника - EKF Team Hub</title>
</svelte:head>

<div class="max-w-2xl mx-auto bg-white rounded-xl shadow-sm p-8">
	<h1 class="text-2xl font-bold text-gray-900 mb-2">Редактировать сотрудника</h1>
	<p class="text-gray-600 mb-6">
		Редактирование профиля доступно через корпоративные системы. В приложении можно посмотреть карточку сотрудника.
	</p>
	<div class="flex gap-4">
		<a href="/employees/{id}" class="text-ekf-red hover:underline">Вернуться к карточке</a>
		<a href="/employees" class="text-gray-600 hover:underline">К списку сотрудников</a>
	</div>
</div>
