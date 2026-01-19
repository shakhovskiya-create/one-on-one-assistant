<script lang="ts">
	import { onMount } from 'svelte';
	import { subordinates, user } from '$lib/stores/auth';
	import { analytics as analyticsApi, meetings } from '$lib/api/client';
	import type { DashboardData, EmployeeAnalytics, MeetingCategory } from '$lib/api/client';

	let dashboard: DashboardData | null = $state(null);
	let employeeAnalytics: EmployeeAnalytics | null = $state(null);
	let categories: MeetingCategory[] = $state([]);
	let loading = $state(true);
	let selectedPeriod = $state('month');
	let selectedEmployee = $state('');
	let selectedCategory = $state('all');
	let viewScope = $state<'my' | 'all'>('my'); // Default to "my team"

	onMount(async () => {
		try {
			// Get manager_id for filtering when viewScope is 'my'
			const managerId = viewScope === 'my' && $user ? $user.id : undefined;
			const [dashboardData, categoriesData] = await Promise.all([
				analyticsApi.getDashboard(selectedPeriod, managerId).catch(() => null),
				meetings.getCategories().catch(() => [])
			]);
			dashboard = dashboardData;
			categories = categoriesData || [];
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});

	async function loadEmployeeAnalytics(employeeId: string, period: string) {
		if (!employeeId) {
			employeeAnalytics = null;
			return;
		}
		try {
			employeeAnalytics = await analyticsApi.getEmployee(employeeId, period);
		} catch (e) {
			console.error(e);
			employeeAnalytics = null;
		}
	}

	$effect(() => {
		const period = selectedPeriod;
		const scope = viewScope;
		const currentUser = $user;
		if (selectedEmployee) {
			loadEmployeeAnalytics(selectedEmployee, period);
		} else {
			// Get manager_id for filtering when viewScope is 'my'
			const managerId = scope === 'my' && currentUser ? currentUser.id : undefined;
			analyticsApi.getDashboard(period, managerId).then((data) => {
				dashboard = data;
			}).catch((e) => {
				console.error(e);
			});
		}
	});

	function getMoodTrend(): 'up' | 'down' | 'stable' | null {
		if (!employeeAnalytics || !employeeAnalytics.mood_history || employeeAnalytics.mood_history.length < 2) return null;
		const recent = employeeAnalytics.mood_history.slice(-3);
		if (recent.length < 2) return null;

		const avg = recent.reduce((sum, m) => sum + m.score, 0) / recent.length;
		const prevData = employeeAnalytics.mood_history.slice(-6, -3);
		const prevAvg = prevData.length > 0
			? prevData.reduce((sum, m) => sum + m.score, 0) / prevData.length
			: avg;

		const diff = avg - prevAvg;
		if (diff > 0.5) return 'up';
		if (diff < -0.5) return 'down';
		return 'stable';
	}

	const moodTrend = $derived(getMoodTrend());

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
	}

	// Chart helpers
	function getChartPath(data: { score: number }[], width: number, height: number): string {
		if (!data || data.length === 0) return '';
		const max = 10;
		const min = 0;
		const padding = 10;
		const chartWidth = width - padding * 2;
		const chartHeight = height - padding * 2;

		const points = data.map((d, i) => {
			const x = padding + (i / (data.length - 1)) * chartWidth;
			const y = padding + chartHeight - ((d.score - min) / (max - min)) * chartHeight;
			return `${x},${y}`;
		});

		return `M ${points.join(' L ')}`;
	}

	function getChartArea(data: { score: number }[], width: number, height: number): string {
		if (!data || data.length === 0) return '';
		const max = 10;
		const min = 0;
		const padding = 10;
		const chartWidth = width - padding * 2;
		const chartHeight = height - padding * 2;

		const points = data.map((d, i) => {
			const x = padding + (i / (data.length - 1)) * chartWidth;
			const y = padding + chartHeight - ((d.score - min) / (max - min)) * chartHeight;
			return `${x},${y}`;
		});

		const firstX = padding;
		const lastX = padding + chartWidth;
		const bottomY = padding + chartHeight;

		return `M ${firstX},${bottomY} L ${points.join(' L ')} L ${lastX},${bottomY} Z`;
	}
</script>

<svelte:head>
	<title>Аналитика - EKF Hub</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex items-center justify-between flex-wrap gap-3">
		<div>
			<h1 class="text-xl font-bold text-gray-900">Аналитика</h1>
			<p class="text-sm text-gray-500">
				{viewScope === 'my' ? 'Мои подчинённые' : 'Все сотрудники'}
			</p>
		</div>
		<div class="flex flex-wrap items-center gap-3">
			<!-- Scope Toggle -->
			<div class="flex rounded-lg border border-gray-200 overflow-hidden bg-white">
				<button
					onclick={() => { viewScope = 'my'; selectedEmployee = ''; }}
					class="px-3 py-1.5 text-sm transition-colors {viewScope === 'my' ? 'bg-ekf-red text-white' : 'text-gray-600 hover:bg-gray-50'}"
				>
					Мои
				</button>
				<button
					onclick={() => { viewScope = 'all'; selectedEmployee = ''; }}
					class="px-3 py-1.5 text-sm transition-colors {viewScope === 'all' ? 'bg-ekf-red text-white' : 'text-gray-600 hover:bg-gray-50'}"
				>
					Все
				</button>
			</div>

			<!-- Employee Selector -->
			<select
				bind:value={selectedEmployee}
				class="border border-gray-200 rounded-lg px-3 py-1.5 text-sm focus:ring-1 focus:ring-ekf-red focus:border-ekf-red"
			>
				<option value="">Общая аналитика</option>
				{#if viewScope === 'my'}
					{#each $subordinates as emp}
						<option value={emp.id}>{emp.name}</option>
					{/each}
				{:else}
					{#each $subordinates as emp}
						<option value={emp.id}>{emp.name}</option>
					{/each}
				{/if}
			</select>

			<!-- Period Selector -->
			<div class="flex rounded-lg border border-gray-200 overflow-hidden bg-white">
				{#each [
					{ key: 'week', label: 'Неделя' },
					{ key: 'month', label: 'Месяц' },
					{ key: 'quarter', label: 'Квартал' },
					{ key: 'year', label: 'Год' }
				] as period}
					<button
						onclick={() => selectedPeriod = period.key}
						class="px-3 py-1.5 text-sm transition-colors
							{selectedPeriod === period.key
								? 'bg-ekf-red text-white'
								: 'text-gray-600 hover:bg-gray-50'}"
					>
						{period.label}
					</button>
				{/each}
			</div>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center h-64">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
		</div>
	{:else if selectedEmployee && employeeAnalytics}
		<!-- Employee-specific analytics -->
		<div class="bg-white rounded-xl shadow-sm p-6">
			<div class="flex items-center justify-between">
				<div>
					<h2 class="text-xl font-semibold text-gray-900">
						{$subordinates.find(s => s.id === selectedEmployee)?.name}
					</h2>
					<p class="text-gray-500">
						{$subordinates.find(s => s.id === selectedEmployee)?.position}
					</p>
				</div>
				<div class="flex items-center gap-6">
					<div class="text-center">
						<p class="text-sm text-gray-500">Всего встреч</p>
						<p class="text-2xl font-bold text-gray-900">{employeeAnalytics.total_meetings}</p>
					</div>
					<div class="text-center">
						<p class="text-sm text-gray-500">Тренд настроения</p>
						<div class="flex items-center gap-2 justify-center">
							{#if moodTrend === 'up'}
								<svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
								</svg>
								<span class="text-green-600 font-medium">Растет</span>
							{:else if moodTrend === 'down'}
								<svg class="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6" />
								</svg>
								<span class="text-red-600 font-medium">Падает</span>
							{:else if moodTrend === 'stable'}
								<svg class="w-6 h-6 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14" />
								</svg>
								<span class="text-gray-600 font-medium">Стабильно</span>
							{:else}
								<span class="text-gray-400">Нет данных</span>
							{/if}
						</div>
					</div>
				</div>
			</div>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Mood Chart -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-4">Динамика настроения</h3>
				{#if employeeAnalytics.mood_history && employeeAnalytics.mood_history.length > 0}
					<div class="relative h-64">
						<svg class="w-full h-full" viewBox="0 0 400 200">
							<!-- Grid lines -->
							{#each [0, 2.5, 5, 7.5, 10] as val}
								<line
									x1="40" y1={180 - val * 16}
									x2="390" y2={180 - val * 16}
									stroke="#e5e7eb" stroke-width="1" />
								<text x="35" y={185 - val * 16} text-anchor="end" class="text-xs fill-gray-400">{val}</text>
							{/each}

							<!-- Area fill -->
							<path
								d={getChartArea(employeeAnalytics.mood_history, 350, 160).replace(/10,/g, '40,').replace(/340,/g, '390,')}
								fill="url(#moodGradient)" />

							<!-- Line -->
							<path
								d={getChartPath(employeeAnalytics.mood_history, 350, 160).replace(/10,/g, '40,').replace(/340,/g, '390,')}
								fill="none" stroke="#ef4444" stroke-width="2" />

							<!-- Points -->
							{#each employeeAnalytics.mood_history as point, i}
								{@const x = 40 + (i / (employeeAnalytics.mood_history.length - 1)) * 350}
								{@const y = 180 - (point.score / 10) * 160}
								<circle cx={x} cy={y} r="4" fill="#ef4444" />
							{/each}

							<defs>
								<linearGradient id="moodGradient" x1="0" x2="0" y1="0" y2="1">
									<stop offset="0%" stop-color="#ef4444" stop-opacity="0.3" />
									<stop offset="100%" stop-color="#ef4444" stop-opacity="0" />
								</linearGradient>
							</defs>
						</svg>

						<!-- X-axis labels -->
						<div class="absolute bottom-0 left-10 right-0 flex justify-between text-xs text-gray-400 px-2">
							{#each employeeAnalytics.mood_history.filter((_, i) => i % Math.ceil(employeeAnalytics.mood_history.length / 5) === 0) as point}
								<span>{formatDate(point.date)}</span>
							{/each}
						</div>
					</div>
				{:else}
					<div class="h-64 flex items-center justify-center text-gray-400">
						Недостаточно данных для отображения графика
					</div>
				{/if}
			</div>

			<!-- Agreements Pie Chart -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-4">Статус договорённостей</h3>
				{#if employeeAnalytics.agreement_stats && employeeAnalytics.agreement_stats.total > 0}
					{@const total = employeeAnalytics.agreement_stats.total}
					{@const completed = employeeAnalytics.agreement_stats.completed}
					{@const pending = employeeAnalytics.agreement_stats.pending}
					{@const overdue = employeeAnalytics.agreement_stats.overdue}

					{@const completedDash = (completed / total) * 440}
					{@const pendingDash = (pending / total) * 440}
					{@const overdueDash = (overdue / total) * 440}
					<div class="flex items-center justify-center">
						<svg class="w-48 h-48" viewBox="0 0 200 200">
							<!-- Completed -->
							{#if completed > 0}
								<circle
									cx="100" cy="100" r="70"
									fill="transparent"
									stroke="#22c55e" stroke-width="30"
									stroke-dasharray="{completedDash} 440"
									transform="rotate(-90 100 100)" />
							{/if}

							<!-- Pending -->
							{#if pending > 0}
								<circle
									cx="100" cy="100" r="70"
									fill="transparent"
									stroke="#f59e0b" stroke-width="30"
									stroke-dasharray="{pendingDash} 440"
									stroke-dashoffset="{-completedDash}"
									transform="rotate(-90 100 100)" />
							{/if}

							<!-- Overdue -->
							{#if overdue > 0}
								<circle
									cx="100" cy="100" r="70"
									fill="transparent"
									stroke="#ef4444" stroke-width="30"
									stroke-dasharray="{overdueDash} 440"
									stroke-dashoffset="{-(completedDash + pendingDash)}"
									transform="rotate(-90 100 100)" />
							{/if}

							<!-- Center text -->
							<text x="100" y="95" text-anchor="middle" class="text-2xl font-bold fill-gray-900">{total}</text>
							<text x="100" y="115" text-anchor="middle" class="text-sm fill-gray-500">всего</text>
						</svg>
					</div>

					<div class="flex justify-center gap-6 mt-4">
						<div class="flex items-center gap-2">
							<div class="w-3 h-3 rounded-full bg-green-500"></div>
							<span class="text-sm text-gray-600">Выполнено: {completed}</span>
						</div>
						<div class="flex items-center gap-2">
							<div class="w-3 h-3 rounded-full bg-yellow-500"></div>
							<span class="text-sm text-gray-600">В работе: {pending}</span>
						</div>
						<div class="flex items-center gap-2">
							<div class="w-3 h-3 rounded-full bg-red-500"></div>
							<span class="text-sm text-gray-600">Просрочено: {overdue}</span>
						</div>
					</div>
				{:else}
					<div class="h-48 flex items-center justify-center text-gray-400">
						Нет договорённостей
					</div>
				{/if}
			</div>

			<!-- Red Flags History -->
			<div class="bg-white rounded-xl shadow-sm p-6 lg:col-span-2">
				<h3 class="font-semibold text-gray-900 mb-4 flex items-center gap-2">
					<svg class="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
					Красные флаги
				</h3>
				{#if employeeAnalytics.red_flags_history && employeeAnalytics.red_flags_history.length > 0}
					<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
						{#each employeeAnalytics.red_flags_history as flag}
							<div class="p-4 bg-red-50 border border-red-200 rounded-lg">
								<p class="text-sm text-gray-600 mb-2">{formatDate(flag.date)}</p>
								{#if flag.flags}
									{#if flag.flags.burnout_signs}
										<p class="text-red-700 text-sm">Признаки выгорания</p>
									{/if}
									{#if flag.flags.turnover_risk && flag.flags.turnover_risk !== 'low'}
										<p class="text-red-700 text-sm">
											Риск ухода: {flag.flags.turnover_risk === 'high' ? 'Высокий' : 'Средний'}
										</p>
									{/if}
								{/if}
							</div>
						{/each}
					</div>
				{:else}
					<div class="flex items-center justify-center py-8 text-green-600">
						<svg class="w-6 h-6 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						Красных флагов не обнаружено
					</div>
				{/if}
			</div>
		</div>

	{:else if dashboard}
		<!-- Dashboard analytics (no employee selected) -->

		<!-- Overview Stats -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
			<a href="/employees" class="bg-white rounded-xl shadow-sm p-6 hover:shadow-md transition-shadow cursor-pointer">
				<div class="flex items-center gap-3">
					<div class="w-12 h-12 bg-orange-50 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-ekf-red" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
						</svg>
					</div>
					<div>
						<p class="text-sm font-medium text-gray-500">Всего сотрудников</p>
						<p class="text-3xl font-bold text-gray-900">{dashboard.total_employees}</p>
					</div>
				</div>
			</a>
			<a href="/meetings" class="bg-white rounded-xl shadow-sm p-6 hover:shadow-md transition-shadow cursor-pointer">
				<div class="flex items-center gap-3">
					<div class="w-12 h-12 bg-green-50 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
					</div>
					<div>
						<p class="text-sm font-medium text-gray-500">Встречи за период</p>
						<p class="text-3xl font-bold text-ekf-red">{dashboard.meetings_this_month}</p>
					</div>
				</div>
			</a>
			<a href="/meetings" class="bg-white rounded-xl shadow-sm p-6 hover:shadow-md transition-shadow cursor-pointer">
				<div class="flex items-center gap-3">
					<div class="w-12 h-12 rounded-lg flex items-center justify-center
						{dashboard.average_mood >= 7 ? 'bg-green-50' : dashboard.average_mood >= 5 ? 'bg-yellow-50' : 'bg-red-50'}">
						<svg class="w-6 h-6 {dashboard.average_mood >= 7 ? 'text-green-600' : dashboard.average_mood >= 5 ? 'text-yellow-600' : 'text-red-600'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<div>
						<p class="text-sm font-medium text-gray-500">Среднее настроение</p>
						<p class="text-3xl font-bold {dashboard.average_mood >= 7 ? 'text-green-600' : dashboard.average_mood >= 5 ? 'text-yellow-600' : 'text-red-600'}">
							{dashboard.average_mood?.toFixed(1) || '-'}/10
						</p>
					</div>
				</div>
			</a>
			<a href="/tasks?status=done" class="bg-white rounded-xl shadow-sm p-6 hover:shadow-md transition-shadow cursor-pointer">
				<div class="flex items-center gap-3">
					<div class="w-12 h-12 bg-blue-50 rounded-lg flex items-center justify-center">
						<svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<div>
						<p class="text-sm font-medium text-gray-500">Задач выполнено</p>
						<p class="text-3xl font-bold text-green-600">{dashboard.tasks_completed}</p>
					</div>
				</div>
			</a>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Mood Trend Chart -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-4">Динамика настроения команды</h3>
				{#if dashboard.mood_trend && dashboard.mood_trend.length > 0}
					<div class="h-64 flex items-end gap-2">
						{#each dashboard.mood_trend as point, i}
							{@const height = (point.score / 10) * 100}
							<div class="flex-1 flex flex-col items-center gap-1">
								<span class="text-xs text-gray-500">{point.score.toFixed(1)}</span>
								<div
									class="w-full rounded-t transition-all
										{point.score >= 7 ? 'bg-green-500' : point.score >= 5 ? 'bg-yellow-500' : 'bg-red-500'}"
									style="height: {height}%"
								></div>
								<span class="text-xs text-gray-400">{formatDate(point.date)}</span>
							</div>
						{/each}
					</div>
				{:else}
					<div class="h-64 flex items-center justify-center text-gray-400">
						Недостаточно данных для отображения тренда
					</div>
				{/if}
			</div>

			<!-- Employees Needing Attention -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-4 flex items-center gap-2">
					<svg class="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
					Требуют внимания
				</h3>
				{#if dashboard.employees_needing_attention && dashboard.employees_needing_attention.length > 0}
					<div class="space-y-3 max-h-64 overflow-y-auto">
						{#each dashboard.employees_needing_attention as emp}
							<a href="/employees/{emp.id}" class="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 border border-red-100">
								<div class="flex items-center gap-3">
									<div class="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center text-red-600 font-medium">
										{emp.name.charAt(0)}
									</div>
									<div>
										<p class="font-medium text-gray-900">{emp.name}</p>
										<p class="text-sm text-gray-500">{emp.reason}</p>
									</div>
								</div>
								<span class="text-sm font-medium text-red-600">
									{emp.days_since_meeting} дней
								</span>
							</a>
						{/each}
					</div>
				{:else}
					<div class="h-64 flex items-center justify-center text-green-600">
						<svg class="w-6 h-6 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						Все сотрудники в порядке
					</div>
				{/if}
			</div>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Meeting Categories -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-4">Типы встреч</h3>
				{#if dashboard.meetings_by_category && Object.keys(dashboard.meetings_by_category).length > 0}
					{@const total = Object.values(dashboard.meetings_by_category).reduce((a, b) => a + b, 0)}
					<div class="space-y-3">
						{#each Object.entries(dashboard.meetings_by_category) as [catCode, count]}
							{@const cat = categories.find(c => c.code === catCode)}
							{@const percentage = total > 0 ? (count / total) * 100 : 0}
							<div>
								<div class="flex justify-between text-sm mb-1">
									<span class="text-gray-600">{cat?.name || catCode}</span>
									<span class="font-medium">{count}</span>
								</div>
								<div class="h-2 bg-gray-100 rounded-full overflow-hidden">
									<div class="h-full bg-ekf-red rounded-full" style="width: {percentage}%"></div>
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<div class="py-8 text-center text-gray-400">Нет данных</div>
				{/if}
			</div>

			<!-- Task Stats -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-4">Статистика задач</h3>
				<div class="space-y-4">
					<a href="/tasks?status=backlog,todo" class="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors cursor-pointer">
						<span class="text-gray-600">К выполнению</span>
						<span class="font-bold text-gray-900">{dashboard.tasks_todo || 0}</span>
					</a>
					<a href="/tasks?status=in_progress" class="flex items-center justify-between p-3 bg-blue-50 rounded-lg hover:bg-blue-100 transition-colors cursor-pointer">
						<span class="text-gray-600">В работе</span>
						<span class="font-bold text-blue-600">{dashboard.tasks_in_progress || 0}</span>
					</a>
					<a href="/tasks?status=done" class="flex items-center justify-between p-3 bg-green-50 rounded-lg hover:bg-green-100 transition-colors cursor-pointer">
						<span class="text-gray-600">Выполнено</span>
						<span class="font-bold text-green-600">{dashboard.tasks_completed || 0}</span>
					</a>
				</div>
			</div>

			<!-- Agreements Stats -->
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-4">Договорённости</h3>
				<div class="space-y-4">
					<div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
						<span class="text-gray-600">Всего</span>
						<span class="font-bold text-gray-900">{dashboard.agreements_total || 0}</span>
					</div>
					<div class="flex items-center justify-between p-3 bg-green-50 rounded-lg">
						<span class="text-gray-600">Выполнено</span>
						<span class="font-bold text-green-600">{dashboard.agreements_completed || 0}</span>
					</div>
					<div class="flex items-center justify-between p-3 bg-red-50 rounded-lg">
						<span class="text-gray-600">Просрочено</span>
						<span class="font-bold text-red-600">{dashboard.agreements_overdue || 0}</span>
					</div>
				</div>
			</div>
		</div>

		<!-- Recent Meetings -->
		{#if dashboard.recent_meetings && dashboard.recent_meetings.length > 0}
			<div class="bg-white rounded-xl shadow-sm p-6">
				<h3 class="font-semibold text-gray-900 mb-4">Последние встречи</h3>
				<div class="space-y-3">
					{#each dashboard.recent_meetings.slice(0, 5) as meeting}
						<a href="/meetings/{meeting.id}" class="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 border">
							<div>
								<p class="font-medium text-gray-900">{meeting.title || 'Без названия'}</p>
								<p class="text-sm text-gray-500">{formatDate(meeting.date)} {meeting.employee_name ? `• ${meeting.employee_name}` : ''}</p>
							</div>
							{#if meeting.mood_score}
								<span class="text-sm font-medium px-2 py-1 rounded
									{meeting.mood_score >= 7 ? 'bg-green-100 text-green-700' :
									meeting.mood_score >= 5 ? 'bg-yellow-100 text-yellow-700' : 'bg-red-100 text-red-700'}">
									{meeting.mood_score}/10
								</span>
							{/if}
						</a>
					{/each}
				</div>
			</div>
		{/if}
	{:else}
		<div class="text-center py-12">
			<div class="text-gray-400 text-lg">Не удалось загрузить данные аналитики</div>
		</div>
	{/if}
</div>
