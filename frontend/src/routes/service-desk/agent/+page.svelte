<script lang="ts">
	import { onMount } from 'svelte';
	import { serviceDesk, auth } from '$lib/api/client';
	import type { ServiceTicket, ServiceDeskStats } from '$lib/api/client';

	let tickets = $state<ServiceTicket[]>([]);
	let stats = $state<ServiceDeskStats | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let selectedTicket = $state<ServiceTicket | null>(null);

	// Filters
	let filterStatus = $state('all');
	let filterType = $state('all');
	let filterPriority = $state('all');
	let searchQuery = $state('');

	// Sidebar navigation
	let activeNav = $state('queue');

	onMount(async () => {
		try {
			const [ticketList, statsData] = await Promise.all([
				serviceDesk.list(),
				serviceDesk.getStats().catch(() => null),
			]);
			tickets = ticketList || [];
			stats = statsData;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load data';
		} finally {
			loading = false;
		}
	});

	const filteredTickets = $derived.by(() => {
		let result = tickets;

		if (activeNav === 'incidents') {
			result = result.filter(t => t.type === 'incident');
		} else if (activeNav === 'requests') {
			result = result.filter(t => t.type === 'service_request');
		} else if (activeNav === 'my') {
			// My assigned tickets would filter by assignee_id
		}

		if (filterStatus !== 'all') {
			result = result.filter(t => t.status === filterStatus);
		}
		if (filterPriority !== 'all') {
			result = result.filter(t => t.priority === filterPriority);
		}
		if (searchQuery) {
			const q = searchQuery.toLowerCase();
			result = result.filter(t =>
				t.title.toLowerCase().includes(q) ||
				t.number.toLowerCase().includes(q)
			);
		}

		return result;
	});

	const queueCount = $derived(tickets.filter(t => !['resolved', 'closed'].includes(t.status)).length);
	const incidentCount = $derived(tickets.filter(t => t.type === 'incident').length);
	const requestCount = $derived(tickets.filter(t => t.type === 'service_request').length);

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

	function getPriorityClass(priority: string): string {
		switch (priority) {
			case 'critical': return 'text-red-600';
			case 'high': return 'text-orange-600';
			case 'medium': return 'text-yellow-600';
			case 'low': return 'text-green-600';
			default: return 'text-gray-600';
		}
	}

	function getPriorityLabel(priority: string): string {
		switch (priority) {
			case 'critical': return 'Критический';
			case 'high': return 'Высокий';
			case 'medium': return 'Средний';
			case 'low': return 'Низкий';
			default: return priority;
		}
	}

	function getSLAStatus(ticket: ServiceTicket): { label: string; class: string } {
		if (!ticket.sla_deadline) return { label: '-', class: 'text-gray-400' };

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

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' });
	}
</script>

<svelte:head>
	<title>Service Desk Agent Console - EKF Hub</title>
</svelte:head>

<div class="flex h-[calc(100vh-48px)]">
	<!-- Sidebar -->
	<aside class="w-60 bg-ekf-dark text-white flex flex-col flex-shrink-0">
		<div class="p-4 border-b border-gray-700">
			<div class="flex items-center gap-2">
				<svg class="w-6 h-6 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z"></path>
				</svg>
				<div>
					<div class="font-semibold text-sm">Service Desk</div>
					<div class="text-xs text-gray-400">Agent Console</div>
				</div>
			</div>
		</div>

		<nav class="flex-1 p-3 space-y-1">
			<button
				onclick={() => activeNav = 'queue'}
				class="w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors {activeNav === 'queue' ? 'bg-ekf-red text-white' : 'text-gray-300 hover:bg-gray-700'}"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"></path>
				</svg>
				<span>Queue</span>
				<span class="ml-auto bg-white/20 text-white text-xs px-2 py-0.5 rounded-full">{queueCount}</span>
			</button>

			<button
				onclick={() => activeNav = 'incidents'}
				class="w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors {activeNav === 'incidents' ? 'bg-ekf-red text-white' : 'text-gray-300 hover:bg-gray-700'}"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
				</svg>
				<span>Incidents</span>
				<span class="ml-auto text-xs text-gray-400">{incidentCount}</span>
			</button>

			<button
				onclick={() => activeNav = 'requests'}
				class="w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors {activeNav === 'requests' ? 'bg-ekf-red text-white' : 'text-gray-300 hover:bg-gray-700'}"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>
				</svg>
				<span>Service Requests</span>
				<span class="ml-auto text-xs text-gray-400">{requestCount}</span>
			</button>

			<div class="border-t border-gray-700 my-3"></div>

			<button
				onclick={() => activeNav = 'my'}
				class="w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors {activeNav === 'my' ? 'bg-ekf-red text-white' : 'text-gray-300 hover:bg-gray-700'}"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
				</svg>
				<span>My Tickets</span>
			</button>
		</nav>

		<!-- Portal Link -->
		<div class="p-3 border-t border-gray-700">
			<a href="/service-desk" class="flex items-center gap-2 px-3 py-2 text-gray-400 hover:text-white text-sm">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path>
				</svg>
				Портал пользователя
			</a>
		</div>
	</aside>

	<!-- Main Content -->
	<div class="flex-1 flex flex-col overflow-hidden">
		<!-- Stats Bar -->
		{#if stats}
			<div class="bg-white border-b p-4 flex gap-4">
				<div class="flex items-center gap-3 px-4 py-2 bg-gray-50 rounded-lg">
					<div class="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
						<svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
						</svg>
					</div>
					<div>
						<div class="text-2xl font-bold text-green-600">{stats.sla_compliance}%</div>
						<div class="text-xs text-gray-500">SLA</div>
					</div>
				</div>

				<div class="flex items-center gap-3 px-4 py-2 bg-gray-50 rounded-lg">
					<div class="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
						<svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
						</svg>
					</div>
					<div>
						<div class="text-2xl font-bold text-blue-600">{stats.open_tickets}</div>
						<div class="text-xs text-gray-500">Открыто</div>
					</div>
				</div>

				<div class="flex items-center gap-3 px-4 py-2 bg-gray-50 rounded-lg">
					<div class="w-10 h-10 bg-yellow-100 rounded-lg flex items-center justify-center">
						<svg class="w-5 h-5 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
						</svg>
					</div>
					<div>
						<div class="text-2xl font-bold text-yellow-600">{stats.overdue_tickets || 0}</div>
						<div class="text-xs text-gray-500">Просрочено</div>
					</div>
				</div>

				<div class="flex items-center gap-3 px-4 py-2 bg-gray-50 rounded-lg">
					<div class="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center">
						<svg class="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
						</svg>
					</div>
					<div>
						<div class="text-2xl font-bold text-purple-600">{stats.resolved_today}</div>
						<div class="text-xs text-gray-500">Решено сегодня</div>
					</div>
				</div>
			</div>
		{/if}

		<!-- Filters -->
		<div class="bg-white border-b p-4 flex items-center gap-4">
			<div class="relative flex-1 max-w-md">
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Поиск по номеру или названию..."
					class="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-ekf-red focus:border-transparent"
				/>
				<svg class="w-5 h-5 text-gray-400 absolute left-3 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
				</svg>
			</div>

			<select bind:value={filterStatus} class="px-3 py-2 border border-gray-200 rounded-lg text-sm">
				<option value="all">Все статусы</option>
				<option value="new">Новые</option>
				<option value="in_progress">В работе</option>
				<option value="pending">Ожидают</option>
				<option value="resolved">Решённые</option>
			</select>

			<select bind:value={filterPriority} class="px-3 py-2 border border-gray-200 rounded-lg text-sm">
				<option value="all">Все приоритеты</option>
				<option value="critical">Критический</option>
				<option value="high">Высокий</option>
				<option value="medium">Средний</option>
				<option value="low">Низкий</option>
			</select>
		</div>

		<!-- Tickets Table -->
		<div class="flex-1 overflow-auto">
			{#if loading}
				<div class="flex items-center justify-center h-64">
					<div class="animate-spin w-8 h-8 border-4 border-ekf-red border-t-transparent rounded-full"></div>
				</div>
			{:else if error}
				<div class="m-4 bg-red-50 text-red-700 p-4 rounded-lg">{error}</div>
			{:else if filteredTickets.length === 0}
				<div class="flex flex-col items-center justify-center h-64 text-gray-400">
					<svg class="w-16 h-16 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>
					</svg>
					<p class="text-lg">Нет заявок</p>
				</div>
			{:else}
				<table class="w-full">
					<thead class="bg-gray-50 sticky top-0">
						<tr class="text-left text-xs text-gray-500 uppercase">
							<th class="px-4 py-3 font-medium">Номер</th>
							<th class="px-4 py-3 font-medium">Тип</th>
							<th class="px-4 py-3 font-medium">Приоритет</th>
							<th class="px-4 py-3 font-medium w-1/3">Тема</th>
							<th class="px-4 py-3 font-medium">Заявитель</th>
							<th class="px-4 py-3 font-medium">Статус</th>
							<th class="px-4 py-3 font-medium">SLA</th>
							<th class="px-4 py-3 font-medium">Создано</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100">
						{#each filteredTickets as ticket (ticket.id)}
							<tr
								class="hover:bg-gray-50 cursor-pointer transition-colors {selectedTicket?.id === ticket.id ? 'bg-red-50 border-l-4 border-l-ekf-red' : ''}"
								onclick={() => selectedTicket = ticket}
							>
								<td class="px-4 py-3">
									<a href="/service-desk/tickets/{ticket.id}" class="text-ekf-red font-medium hover:underline">
										{ticket.number}
									</a>
								</td>
								<td class="px-4 py-3">
									{#if ticket.type === 'incident'}
										<span class="inline-flex items-center gap-1 text-red-600 text-xs">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
											</svg>
											Incident
										</span>
									{:else}
										<span class="inline-flex items-center gap-1 text-blue-600 text-xs">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>
											</svg>
											Request
										</span>
									{/if}
								</td>
								<td class="px-4 py-3">
									<span class="text-sm font-medium {getPriorityClass(ticket.priority)}">
										{getPriorityLabel(ticket.priority)}
									</span>
								</td>
								<td class="px-4 py-3">
									<div class="font-medium text-gray-900 line-clamp-1">{ticket.title}</div>
								</td>
								<td class="px-4 py-3">
									{#if ticket.requester}
										<div class="flex items-center gap-2">
											{#if ticket.requester.photo_base64}
												<img src="data:image/jpeg;base64,{ticket.requester.photo_base64}" alt="" class="w-6 h-6 rounded-full object-cover" />
											{:else}
												<div class="w-6 h-6 rounded-full bg-gray-200 flex items-center justify-center text-xs text-gray-500">
													{ticket.requester.name.charAt(0)}
												</div>
											{/if}
											<span class="text-sm text-gray-600">{ticket.requester.name.split(' ')[0]}</span>
										</div>
									{:else}
										<span class="text-gray-400 text-sm">-</span>
									{/if}
								</td>
								<td class="px-4 py-3">
									<span class="px-2 py-1 rounded-full text-xs font-medium {getStatusColor(ticket.status)}">
										{getStatusLabel(ticket.status)}
									</span>
								</td>
								<td class="px-4 py-3">
									{@const sla = getSLAStatus(ticket)}
									<span class="px-2 py-1 rounded text-xs font-medium {sla.class}">
										{sla.label}
									</span>
								</td>
								<td class="px-4 py-3 text-sm text-gray-500">
									{formatDate(ticket.created_at)}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>
	</div>
</div>
