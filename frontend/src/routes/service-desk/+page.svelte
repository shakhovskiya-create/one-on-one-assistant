<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { serviceDesk, auth } from '$lib/api/client';
	import type { ServiceTicket, ServiceDeskStats, ServiceTicketCategory } from '$lib/api/client';

	let tickets = $state<ServiceTicket[]>([]);
	let stats = $state<ServiceDeskStats | null>(null);
	let categories = $state<ServiceTicketCategory[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let currentUserId = $state<string | null>(null);
	let searchQuery = $state('');

	const serviceCategories = [
		{ id: 'equipment', name: 'Оборудование', description: 'Ноутбуки, мониторы, периферия', icon: '💻', color: 'from-ekf-red to-red-600', services: 12 },
		{ id: 'software', name: 'ПО', description: 'Лицензии, установка, обновления', icon: '🔧', color: 'from-purple-500 to-violet-600', services: 18 },
		{ id: 'access', name: 'Доступы', description: 'Системы, VPN, папки, учётки', icon: '🔑', color: 'from-green-500 to-emerald-600', services: 9 },
		{ id: 'network', name: 'Сеть и VPN', description: 'Подключение, WiFi, удалённый доступ', icon: '🌐', color: 'from-orange-500 to-amber-600', services: 6 },
		{ id: 'email', name: 'Почта и Календарь', description: 'Outlook, общие ящики, календари', icon: '📧', color: 'from-blue-500 to-cyan-600', services: 8 },
		{ id: 'hr', name: 'HR сервисы', description: 'Приём, увольнение, справки', icon: '👥', color: 'from-pink-500 to-rose-600', services: 15 },
		{ id: 'facilities', name: 'Хозяйственные', description: 'Офис, переговорные, ключи', icon: '🏢', color: 'from-teal-500 to-cyan-600', services: 7 },
		{ id: 'other', name: 'Другие запросы', description: 'Общие вопросы и консультации', icon: '❓', color: 'from-gray-500 to-gray-600', services: 0 },
	];

	onMount(async () => {
		// Set a timeout to prevent infinite loading
		const timeoutId = setTimeout(() => {
			if (loading) {
				loading = false;
				error = 'Превышено время ожидания. Попробуйте обновить страницу.';
			}
		}, 10000);

		try {
			// Try to get user, but don't block if it fails
			let userId: string | null = null;
			try {
				const user = await auth.getMe();
				userId = user.id;
				currentUserId = userId;
			} catch {
				// User not authenticated, show portal without personal tickets
				console.log('User not authenticated, showing public portal');
			}

			// Load stats regardless of auth status
			const statsData = await serviceDesk.getStats().catch(() => null);
			stats = statsData;

			// Only load tickets if user is authenticated
			if (userId) {
				const ticketList = await serviceDesk.getMyTickets(userId).catch(() => []);
				tickets = ticketList || [];
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Не удалось загрузить данные';
		} finally {
			clearTimeout(timeoutId);
			loading = false;
		}
	});

	function getStatusColor(status: string): string {
		switch (status) {
			case 'new': return 'bg-blue-100 text-blue-700';
			case 'in_progress': return 'bg-yellow-100 text-yellow-700';
			case 'pending': return 'bg-purple-100 text-purple-700';
			case 'resolved': return 'bg-green-100 text-green-700';
			case 'closed': return 'bg-gray-100 text-gray-700';
			default: return 'bg-gray-100 text-gray-700';
		}
	}

	function getStatusLabel(status: string): string {
		switch (status) {
			case 'new': return 'Новый';
			case 'in_progress': return 'В работе';
			case 'pending': return 'Ожидает';
			case 'resolved': return 'Решён';
			case 'closed': return 'Закрыт';
			default: return status;
		}
	}

	function getPriorityBorderColor(priority: string): string {
		switch (priority) {
			case 'critical': return 'border-l-red-500';
			case 'high': return 'border-l-orange-500';
			case 'medium': return 'border-l-yellow-500';
			case 'low': return 'border-l-green-500';
			default: return 'border-l-gray-500';
		}
	}

	function getSLAStatus(ticket: ServiceTicket): { label: string; class: string } {
		if (!ticket.sla_deadline) return { label: '', class: '' };

		const deadline = new Date(ticket.sla_deadline);
		const now = new Date();
		const diff = deadline.getTime() - now.getTime();
		const hours = Math.floor(diff / (1000 * 60 * 60));

		if (diff < 0) {
			return { label: 'Просрочено', class: 'bg-red-100 text-red-700' };
		} else if (hours < 2) {
			return { label: `${hours}ч`, class: 'bg-yellow-100 text-yellow-700' };
		} else if (hours < 24) {
			return { label: `${hours}ч`, class: 'bg-green-100 text-green-700' };
		} else {
			const days = Math.floor(hours / 24);
			return { label: `${days}д`, class: 'bg-green-100 text-green-700' };
		}
	}

	const openTickets = $derived(tickets.filter(t => !['resolved', 'closed'].includes(t.status)));
</script>

<svelte:head>
	<title>Service Desk - EKF Hub</title>
</svelte:head>

<div class="min-h-screen bg-gray-50">
	<!-- Agent Console Link -->
	<div class="fixed top-12 right-4 z-40">
		<a
			href="/service-desk/agent"
			class="inline-flex items-center gap-2 px-3 py-1.5 bg-ekf-dark/90 text-gray-300 hover:text-white rounded-lg text-sm transition-colors border border-gray-700"
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
			</svg>
			Консоль агента
		</a>
	</div>

	<!-- Hero Section -->
	<section class="bg-gradient-to-br from-ekf-dark via-gray-900 to-ekf-dark text-white overflow-hidden relative">
		<div class="absolute inset-0 bg-[radial-gradient(circle_at_30%_50%,rgba(229,57,53,0.15),transparent_50%)]"></div>
		<div class="max-w-6xl mx-auto px-6 py-12 relative">
			<div class="flex items-center gap-12">
				<div class="flex-1">
					<div class="inline-flex items-center gap-2 bg-ekf-red/20 text-ekf-red px-3 py-1 rounded-full text-sm mb-4">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z"></path>
						</svg>
						IT Service Desk
					</div>
					<h1 class="text-3xl font-bold mb-3">Чем мы можем помочь?</h1>
					<p class="text-lg text-gray-400 mb-6">Получите помощь или найдите ответ в базе знаний</p>

					<!-- Search -->
					<div class="relative mb-6">
						<input
							type="text"
							bind:value={searchQuery}
							placeholder="Поиск по базе знаний..."
							class="w-full pl-12 pr-6 py-3 bg-white/10 border border-white/20 rounded-xl text-white placeholder-gray-400 focus:ring-2 focus:ring-ekf-red focus:border-transparent"
						>
						<svg class="w-5 h-5 text-gray-400 absolute left-4 top-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
						</svg>
					</div>

					<div class="flex gap-3">
						<a href="/service-desk/create" class="bg-ekf-red text-white px-5 py-2.5 rounded-xl font-medium hover:bg-red-700 flex items-center gap-2 shadow-lg shadow-red-500/30">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
							</svg>
							Новый запрос
						</a>
						<a href="/service-desk/create?type=incident" class="bg-white/10 text-white px-5 py-2.5 rounded-xl font-medium hover:bg-white/20 flex items-center gap-2 border border-white/20">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
							</svg>
							Сообщить о проблеме
						</a>
					</div>
				</div>

				<!-- Stats -->
				{#if stats}
					<div class="grid grid-cols-2 gap-3 w-72">
						<div class="bg-white/10 backdrop-blur rounded-xl p-4 border border-white/10">
							<div class="text-2xl font-bold text-green-400">{stats.sla_compliance}%</div>
							<div class="text-xs text-gray-400">SLA соблюдено</div>
						</div>
						<div class="bg-white/10 backdrop-blur rounded-xl p-4 border border-white/10">
							<div class="text-2xl font-bold text-blue-400">{stats.resolved_today}</div>
							<div class="text-xs text-gray-400">Решено сегодня</div>
						</div>
						<div class="bg-white/10 backdrop-blur rounded-xl p-4 border border-white/10">
							<div class="text-2xl font-bold text-yellow-400">{stats.open_tickets}</div>
							<div class="text-xs text-gray-400">Открытых заявок</div>
						</div>
						<div class="bg-white/10 backdrop-blur rounded-xl p-4 border border-white/10">
							<div class="text-2xl font-bold text-purple-400">24/7</div>
							<div class="text-xs text-gray-400">Техподдержка</div>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</section>

	<div class="max-w-6xl mx-auto px-6 py-8">
		<!-- My Tickets -->
		{#if loading}
			<div class="flex items-center justify-center py-12">
				<div class="animate-spin w-8 h-8 border-4 border-ekf-red border-t-transparent rounded-full"></div>
			</div>
		{:else if error}
			<div class="bg-red-50 text-red-700 p-4 rounded-xl mb-6">{error}</div>
		{:else if openTickets.length > 0}
			<section class="mb-8">
				<div class="flex items-center justify-between mb-4">
					<div>
						<h2 class="text-lg font-semibold text-gray-900">Мои заявки</h2>
						<p class="text-sm text-gray-500">{openTickets.length} открытых заявок</p>
					</div>
					<a href="/service-desk/my-tickets" class="text-ekf-red text-sm font-medium hover:underline flex items-center gap-1">
						Смотреть все
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
						</svg>
					</a>
				</div>
				<div class="grid grid-cols-3 gap-4">
					{#each openTickets.slice(0, 3) as ticket}
						<a
							href="/service-desk/tickets/{ticket.id}"
							class="bg-white rounded-xl p-5 border-l-4 {getPriorityBorderColor(ticket.priority)} shadow-sm hover:shadow-md transition-all cursor-pointer"
						>
							<div class="flex items-start justify-between mb-3">
								<span class="text-xs font-medium text-gray-400">{ticket.number}</span>
								<span class="text-xs font-medium px-2 py-1 rounded-full {getStatusColor(ticket.status)}">
									{getStatusLabel(ticket.status)}
								</span>
							</div>
							<h3 class="font-semibold mb-2 text-gray-900 line-clamp-1">{ticket.title}</h3>
							<p class="text-sm text-gray-500 mb-3 line-clamp-2">{ticket.description || 'Нет описания'}</p>
							<div class="flex items-center justify-between text-sm">
								<div class="flex items-center gap-2">
									{#if ticket.assignee}
										{#if ticket.assignee.photo_base64}
											<img src="data:image/jpeg;base64,{ticket.assignee.photo_base64}" alt="" class="w-6 h-6 rounded-full object-cover">
										{:else}
											<div class="w-6 h-6 rounded-full bg-ekf-red flex items-center justify-center text-white text-xs">
												{ticket.assignee.name.charAt(0)}
											</div>
										{/if}
										<span class="text-gray-500">{ticket.assignee.name.split(' ')[0]}</span>
									{:else}
										<span class="text-gray-400">Не назначен</span>
									{/if}
								</div>
								{#if ticket.sla_deadline && !['resolved', 'closed'].includes(ticket.status)}
									{@const sla = getSLAStatus(ticket)}
									<span class="text-xs px-2 py-1 rounded-full {sla.class} font-medium">
										SLA: {sla.label}
									</span>
								{/if}
							</div>
						</a>
					{/each}
				</div>
			</section>
		{/if}

		<!-- Service Catalog -->
		<section class="mb-8">
			<div class="mb-4">
				<h2 class="text-lg font-semibold text-gray-900">Каталог услуг</h2>
				<p class="text-sm text-gray-500">Выберите категорию или воспользуйтесь поиском</p>
			</div>
			<div class="grid grid-cols-4 gap-4">
				{#each serviceCategories as category}
					<a
						href="/service-desk/create?category={category.id}"
						class="bg-white rounded-xl p-5 border border-gray-100 cursor-pointer hover:shadow-lg hover:border-ekf-red transition-all group"
					>
						<div class="w-12 h-12 bg-gradient-to-br {category.color} rounded-xl flex items-center justify-center mb-3 shadow-lg text-2xl">
							{category.icon}
						</div>
						<h3 class="font-semibold mb-1 text-gray-900">{category.name}</h3>
						<p class="text-sm text-gray-500 mb-2">{category.description}</p>
						{#if category.services > 0}
							<span class="text-xs text-ekf-red font-medium">{category.services} услуг →</span>
						{:else}
							<span class="text-xs text-gray-400 font-medium">Свободная форма →</span>
						{/if}
					</a>
				{/each}
			</div>
		</section>

		<!-- Quick Incident Reports -->
		<section class="mb-8">
			<div class="bg-gradient-to-r from-red-50 to-orange-50 rounded-2xl p-5 border border-red-100">
				<div class="flex items-start gap-4">
					<div class="w-12 h-12 bg-gradient-to-br from-ekf-red to-red-600 rounded-xl flex items-center justify-center flex-shrink-0 shadow-lg">
						<svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
						</svg>
					</div>
					<div class="flex-1">
						<h3 class="font-bold text-lg text-gray-900 mb-2">Сообщить о проблеме</h3>
						<p class="text-gray-600 mb-4">Что-то сломалось или не работает? Опишите проблему для немедленного реагирования.</p>
						<div class="grid grid-cols-3 gap-3">
							<a
								href="/service-desk/create?type=incident&template=login"
								class="bg-white px-4 py-3 rounded-xl border border-red-100 hover:border-ekf-red hover:shadow-md text-left transition-all group"
							>
								<div class="w-8 h-8 bg-red-100 rounded-lg flex items-center justify-center mb-2 group-hover:bg-ekf-red transition-colors">
									<svg class="w-4 h-4 text-ekf-red group-hover:text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path>
									</svg>
								</div>
								<div class="font-semibold text-sm text-gray-900 mb-1">Не могу войти</div>
								<div class="text-xs text-gray-500">Логин, пароль, блокировка</div>
							</a>
							<a
								href="/service-desk/create?type=incident&template=app_error"
								class="bg-white px-4 py-3 rounded-xl border border-red-100 hover:border-ekf-red hover:shadow-md text-left transition-all group"
							>
								<div class="w-8 h-8 bg-red-100 rounded-lg flex items-center justify-center mb-2 group-hover:bg-ekf-red transition-colors">
									<svg class="w-4 h-4 text-ekf-red group-hover:text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
									</svg>
								</div>
								<div class="font-semibold text-sm text-gray-900 mb-1">Ошибка в приложении</div>
								<div class="text-xs text-gray-500">Сбои, баги, неожиданное поведение</div>
							</a>
							<a
								href="/service-desk/create?type=incident"
								class="bg-white px-4 py-3 rounded-xl border border-red-100 hover:border-ekf-red hover:shadow-md text-left transition-all group"
							>
								<div class="w-8 h-8 bg-red-100 rounded-lg flex items-center justify-center mb-2 group-hover:bg-ekf-red transition-colors">
									<svg class="w-4 h-4 text-ekf-red group-hover:text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
									</svg>
								</div>
								<div class="font-semibold text-sm text-gray-900 mb-1">Другая проблема</div>
								<div class="text-xs text-gray-500">Свободная форма</div>
							</a>
						</div>
					</div>
				</div>
			</div>
		</section>
	</div>
</div>
