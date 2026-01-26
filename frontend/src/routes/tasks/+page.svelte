<script lang="ts">
	import { onMount } from 'svelte';
	import { api, workflows, tasks as tasksApi, github, versions as versionsApi, sprints as sprintsApi } from '$lib/api/client';
	import { user, subordinates } from '$lib/stores/auth';
	import type { TaskDependency, StatusColumn, TimeEntry, ResourceSummary, GitHubCommit, GitHubPullRequest, Version, Sprint } from '$lib/api/client';

	// Types
	interface Task {
		id: string;
		title: string;
		description?: string;
		status: 'backlog' | 'todo' | 'in_progress' | 'review' | 'done';
		priority?: number;
		story_points?: number;
		assignee_id?: string;
		assignee?: any;
		project_id?: string;
		project?: any;
		tags?: { name: string; color: string }[];
		due_date?: string;
		parent_id?: string;
		sprint?: string;
		sprint_id?: string;
		fix_version_id?: string;
		created_at: string;
		// Task type
		task_type?: 'feature' | 'bug' | 'tech_debt' | 'improvement' | 'task';
		// Resource planning
		estimated_hours?: number;
		actual_hours?: number;
		estimated_cost?: number;
		actual_cost?: number;
	}

	interface Project {
		id: string;
		name: string;
	}

	// State
	let tasks: Task[] = $state([]);
	let projects: Project[] = $state([]);
	let employees: any[] = $state([]);
	let sprints: Sprint[] = $state([]);
	let versions: Version[] = $state([]);
	let activeSprint: Sprint | null = $state(null);
	let loading = $state(true);

	// View mode
	let viewMode = $state<'list' | 'kanban'>('list');

	// Filters
	let filterProject = $state('');
	let filterAssignee = $state('');
	let filterStatus = $state('');
	let filterSprint = $state('');
	let searchQuery = $state('');

	// Modal state
	let showTaskModal = $state(false);
	let editingTask: Partial<Task> | null = $state(null);
	let selectedTask: Task | null = $state(null);

	// Dependencies state
	let taskDependencies: TaskDependency[] = $state([]);
	let taskDependents: TaskDependency[] = $state([]);
	let taskBlockers: Task[] = $state([]);
	let isTaskBlocked = $state(false);
	let showDependencyPicker = $state(false);
	let dependencySearch = $state('');

	// Time tracking state
	let timeEntries: TimeEntry[] = $state([]);
	let resourceSummary: ResourceSummary | null = $state(null);
	let showTimeEntryForm = $state(false);
	let newTimeEntry = $state({ hours: 0, description: '', date: new Date().toISOString().split('T')[0] });

	// Loading states for task modal
	let loadingTaskDetails = $state(false);

	// GitHub commits and PRs state
	let taskCommits: GitHubCommit[] = $state([]);
	let taskPullRequests: GitHubPullRequest[] = $state([]);
	let loadingCommits = $state(false);
	let githubConfigured = $state(false);
	// Default repo for EKF - can be made configurable per project later
	const defaultGitHubOwner = 'ekf-electrotechnika';
	const defaultGitHubRepo = 'one-on-one-assistant';

	// Status columns for Kanban (dynamically loaded from workflow)
	let statusColumns: StatusColumn[] = $state([
		{ id: 'backlog', label: 'Backlog', color: 'bg-gray-100', wipLimit: 0 },
		{ id: 'todo', label: 'К выполнению', color: 'bg-blue-50', wipLimit: 10 },
		{ id: 'in_progress', label: 'В работе', color: 'bg-yellow-50', wipLimit: 5 },
		{ id: 'done', label: 'Готово', color: 'bg-green-50', wipLimit: 0 }
	]);
	let workflowName = $state('simple');
	let userDepartment = $state('');

	const priorityLabels: Record<number, { label: string; color: string }> = {
		1: { label: 'Критический', color: 'text-red-600 bg-red-50' },
		2: { label: 'Высокий', color: 'text-orange-600 bg-orange-50' },
		3: { label: 'Средний', color: 'text-yellow-600 bg-yellow-50' },
		4: { label: 'Низкий', color: 'text-blue-600 bg-blue-50' },
		5: { label: 'Минимальный', color: 'text-gray-600 bg-gray-50' }
	};

	// Task type labels for badges
	const taskTypeLabels: Record<string, { label: string; color: string; icon: string }> = {
		feature: { label: 'Фича', color: 'bg-blue-100 text-blue-700', icon: '✨' },
		bug: { label: 'Баг', color: 'bg-red-100 text-red-700', icon: '🐛' },
		tech_debt: { label: 'Техдолг', color: 'bg-purple-100 text-purple-700', icon: '🔧' },
		improvement: { label: 'Улучшение', color: 'bg-green-100 text-green-700', icon: '📈' },
		task: { label: 'Задача', color: 'bg-gray-100 text-gray-700', icon: '📋' }
	};

	// Priority border colors for Kanban cards
	const priorityBorderColors: Record<number, string> = {
		1: 'border-l-red-500',
		2: 'border-l-orange-500',
		3: 'border-l-yellow-400',
		4: 'border-l-blue-300',
		5: 'border-l-gray-300'
	};

	onMount(async () => {
		await loadData();
		// Check GitHub configuration
		try {
			const status = await github.status();
			githubConfigured = status.configured;
		} catch {
			githubConfigured = false;
		}
	});

	async function loadData() {
		loading = true;
		try {
			const [tasksRes, projectsRes, employeesRes, workflowRes, sprintsRes, versionsRes, activeSprintRes] = await Promise.all([
				api.tasks.list(),
				api.projects.list().catch(() => []),
				api.employees.list().catch(() => []),
				workflows.getMyWorkflow().catch(() => null),
				sprintsApi.list().catch(() => []),
				versionsApi.list().catch(() => []),
				sprintsApi.getActive().catch(() => null)
			]);
			tasks = tasksRes || [];
			projects = projectsRes || [];
			employees = employeesRes || [];
			sprints = sprintsRes || [];
			versions = versionsRes || [];
			activeSprint = activeSprintRes;

			// Apply workflow statuses if loaded
			if (workflowRes?.workflow?.statuses) {
				statusColumns = workflowRes.workflow.statuses;
				workflowName = workflowRes.workflow.name;
				userDepartment = workflowRes.department || '';
			}
		} catch (e) {
			console.error('Error loading data:', e);
		}
		loading = false;
	}

	function openNewTask() {
		editingTask = {
			title: '',
			description: '',
			status: 'todo',
			priority: 3,
			story_points: undefined,
			assignee_id: $user?.id || '',
			project_id: filterProject || '',
			due_date: '',
			task_type: 'task',
			estimated_hours: undefined,
			estimated_cost: undefined,
			sprint_id: activeSprint?.id || '',
			fix_version_id: ''
		};
		// Reset dependencies for new task
		taskDependencies = [];
		taskDependents = [];
		taskBlockers = [];
		isTaskBlocked = false;
		// Reset time entries for new task
		timeEntries = [];
		resourceSummary = null;
		showTimeEntryForm = false;
		newTimeEntry = { hours: 0, description: '', date: new Date().toISOString().split('T')[0] };
		// Reset commits and PRs
		taskCommits = [];
		taskPullRequests = [];
		showTaskModal = true;
	}

	// Story points options (Fibonacci-like)
	const storyPointsOptions = [1, 2, 3, 5, 8, 13, 21];

	async function openEditTask(task: Task) {
		editingTask = { ...task };
		showTaskModal = true;
		// Reset state
		taskDependencies = [];
		taskDependents = [];
		taskBlockers = [];
		isTaskBlocked = false;
		timeEntries = [];
		resourceSummary = null;
		taskCommits = [];
		taskPullRequests = [];
		// Load dependencies, time entries, and commits for existing task
		if (task.id) {
			loadingTaskDetails = true;
			try {
				await Promise.all([
					loadTaskDependencies(task.id),
					loadTimeEntries(task.id),
					loadResourceSummary(task.id),
					loadTaskCommits(task.id),
					loadTaskPullRequests(task.id)
				]);
			} finally {
				loadingTaskDetails = false;
			}
		}
	}

	async function saveTask() {
		if (!editingTask?.title?.trim()) return;
		try {
			// Clean up empty string values to undefined
			const taskData = {
				...editingTask,
				assignee_id: editingTask.assignee_id || undefined,
				project_id: editingTask.project_id || undefined,
				due_date: editingTask.due_date || undefined,
				priority: Number(editingTask.priority) || 3,
				story_points: editingTask.story_points ? Number(editingTask.story_points) : undefined,
				// Task type
				task_type: editingTask.task_type || 'task',
				// Resource planning fields
				estimated_hours: editingTask.estimated_hours ? Number(editingTask.estimated_hours) : undefined,
				estimated_cost: editingTask.estimated_cost ? Number(editingTask.estimated_cost) : undefined,
				// Sprint and Version
				sprint_id: editingTask.sprint_id || undefined,
				fix_version_id: editingTask.fix_version_id || undefined
			};

			if (editingTask.id) {
				const updated = await api.tasks.update(editingTask.id, taskData);
				tasks = tasks.map(t => t.id === editingTask!.id ? updated : t);
			} else {
				const created = await api.tasks.create(taskData);
				tasks = [created, ...tasks];
			}
			showTaskModal = false;
			editingTask = null;
		} catch (e) {
			console.error('Error saving task:', e);
		}
	}

	// Dependencies functions
	async function loadTaskDependencies(taskId: string) {
		try {
			// Load dependencies and blocked status in parallel
			const [result, blockedResult] = await Promise.all([
				api.tasks.getDependencies(taskId),
				api.tasks.isBlocked(taskId)
			]);
			taskDependencies = result.dependencies || [];
			taskDependents = result.dependents || [];
			isTaskBlocked = blockedResult.blocked;
			taskBlockers = blockedResult.blockers || [];
		} catch (e) {
			console.error('Error loading dependencies:', e);
			taskDependencies = [];
			taskDependents = [];
			isTaskBlocked = false;
			taskBlockers = [];
		}
	}

	async function addDependency(dependsOnTaskId: string) {
		if (!editingTask?.id) return;
		try {
			const dep = await api.tasks.addDependency(editingTask.id, dependsOnTaskId);
			taskDependencies = [...taskDependencies, dep];
			showDependencyPicker = false;
			dependencySearch = '';
			// Reload blocked status
			const blockedResult = await api.tasks.isBlocked(editingTask.id);
			isTaskBlocked = blockedResult.blocked;
			taskBlockers = blockedResult.blockers || [];
		} catch (e: any) {
			alert(e.message || 'Не удалось добавить зависимость');
		}
	}

	async function removeDependency(depId: string) {
		if (!editingTask?.id) return;
		try {
			await api.tasks.removeDependency(editingTask.id, depId);
			taskDependencies = taskDependencies.filter(d => d.id !== depId);
			// Reload blocked status
			const blockedResult = await api.tasks.isBlocked(editingTask.id);
			isTaskBlocked = blockedResult.blocked;
			taskBlockers = blockedResult.blockers || [];
		} catch (e) {
			console.error('Error removing dependency:', e);
		}
	}

	// Time tracking functions
	async function loadTimeEntries(taskId: string) {
		try {
			timeEntries = await tasksApi.getTimeEntries(taskId);
		} catch (e) {
			console.error('Error loading time entries:', e);
			timeEntries = [];
		}
	}

	async function loadResourceSummary(taskId: string) {
		try {
			resourceSummary = await tasksApi.getResourceSummary(taskId);
		} catch (e) {
			console.error('Error loading resource summary:', e);
			resourceSummary = null;
		}
	}

	// GitHub commits and PRs functions
	async function loadTaskCommits(taskId: string) {
		if (!githubConfigured) return;
		loadingCommits = true;
		try {
			const commits = await github.getTaskCommits(taskId, defaultGitHubOwner, defaultGitHubRepo, 10);
			taskCommits = commits || [];
		} catch (e) {
			console.error('Error loading commits:', e);
			taskCommits = [];
		} finally {
			loadingCommits = false;
		}
	}

	async function loadTaskPullRequests(taskId: string) {
		if (!githubConfigured) return;
		try {
			const prs = await github.getTaskPullRequests(taskId, defaultGitHubOwner, defaultGitHubRepo, 10);
			taskPullRequests = prs || [];
		} catch (e) {
			console.error('Error loading pull requests:', e);
			taskPullRequests = [];
		}
	}

	function formatCommitDate(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' });
	}

	function shortenSha(sha: string): string {
		return sha.substring(0, 7);
	}

	async function addTimeEntry() {
		if (!editingTask?.id || newTimeEntry.hours <= 0) return;
		try {
			const entry = await tasksApi.addTimeEntry(editingTask.id, newTimeEntry);
			timeEntries = [entry, ...timeEntries];
			newTimeEntry = { hours: 0, description: '', date: new Date().toISOString().split('T')[0] };
			showTimeEntryForm = false;
			// Reload resource summary
			await loadResourceSummary(editingTask.id);
		} catch (e: any) {
			alert(e.message || 'Не удалось добавить запись времени');
		}
	}

	async function deleteTimeEntry(entryId: string) {
		if (!editingTask?.id) return;
		try {
			await tasksApi.deleteTimeEntry(editingTask.id, entryId);
			timeEntries = timeEntries.filter(e => e.id !== entryId);
			// Reload resource summary
			await loadResourceSummary(editingTask.id);
		} catch (e) {
			console.error('Error deleting time entry:', e);
		}
	}

	function formatHours(hours: number | undefined): string {
		if (!hours) return '-';
		return `${hours.toFixed(1)} ч`;
	}

	function formatCost(cost: number | undefined): string {
		if (!cost) return '-';
		return `${cost.toLocaleString('ru-RU')} ₽`;
	}

	// Filter tasks for dependency picker (exclude current task and already added)
	$effect(() => {
		// This will reactively filter when dependencySearch changes
	});

	function getAvailableDependencies(): Task[] {
		const currentId = editingTask?.id;
		const existingIds = new Set(taskDependencies.map(d => d.depends_on_task_id));
		return tasks.filter(t =>
			t.id !== currentId &&
			!existingIds.has(t.id) &&
			(dependencySearch === '' || t.title.toLowerCase().includes(dependencySearch.toLowerCase()))
		);
	}

	async function updateTaskStatus(task: Task, newStatus: string) {
		try {
			const updated = await api.tasks.update(task.id, { status: newStatus });
			tasks = tasks.map(t => t.id === task.id ? updated : t);
		} catch (e) {
			console.error('Error updating task:', e);
		}
	}

	async function deleteTask(id: string) {
		if (!confirm('Удалить задачу?')) return;
		try {
			await api.tasks.delete(id);
			tasks = tasks.filter(t => t.id !== id);
		} catch (e) {
			console.error('Error deleting task:', e);
		}
	}

	// Filtered tasks
	let filteredTasks = $derived.by(() => {
		let result = tasks;

		if (filterProject) {
			result = result.filter(t => t.project_id === filterProject);
		}
		if (filterAssignee) {
			result = result.filter(t => t.assignee_id === filterAssignee);
		}
		if (filterStatus) {
			result = result.filter(t => t.status === filterStatus);
		}
		if (filterSprint) {
			result = result.filter(t => t.sprint_id === filterSprint);
		}
		if (searchQuery) {
			const q = searchQuery.toLowerCase();
			result = result.filter(t =>
				t.title.toLowerCase().includes(q) ||
				t.description?.toLowerCase().includes(q)
			);
		}

		// Use spread to create a copy before sorting (Svelte 5 prohibits mutating state in $derived)
		return [...result].sort((a, b) => {
			// Sort by priority first, then by created_at
			if ((a.priority || 3) !== (b.priority || 3)) {
				return (a.priority || 3) - (b.priority || 3);
			}
			return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
		});
	});

	function getTasksByStatus(status: string) {
		return filteredTasks.filter(t => t.status === status);
	}

	function getStoryPointsByStatus(status: string): number {
		return getTasksByStatus(status).reduce((sum, t) => sum + (t.story_points || 0), 0);
	}

	function getTotalStoryPoints(): number {
		return filteredTasks.reduce((sum, t) => sum + (t.story_points || 0), 0);
	}

	function isWipLimitExceeded(status: string): boolean {
		const column = statusColumns.find(c => c.id === status);
		if (!column || column.wipLimit === 0) return false;
		return getTasksByStatus(status).length > column.wipLimit;
	}

	function getCompletedTasksCount(): number {
		return filteredTasks.filter(t => t.status === 'done').length;
	}

	function getInProgressTasksCount(): number {
		return filteredTasks.filter(t => t.status === 'in_progress').length;
	}

	// Quick add task in column
	let quickAddColumn: string | null = $state(null);
	let quickAddTitle = $state('');

	async function quickAddTask(status: string) {
		if (!quickAddTitle.trim()) {
			quickAddColumn = null;
			return;
		}
		try {
			const taskData = {
				title: quickAddTitle,
				status,
				priority: 3,
				assignee_id: $user?.id || undefined,
				project_id: filterProject || undefined
			};
			const created = await api.tasks.create(taskData);
			tasks = [created, ...tasks];
			quickAddTitle = '';
			quickAddColumn = null;
		} catch (e) {
			console.error('Error creating task:', e);
		}
	}

	function getEmployeeName(id: string): string {
		if (!id) return 'Не назначен';
		const emp = employees.find(e => e.id === id);
		return emp?.name || 'Неизвестный';
	}

	function getProjectName(id: string): string {
		if (!id) return '';
		const proj = projects.find(p => p.id === id);
		return proj?.name || '';
	}

	function getSprintName(id: string | undefined): string {
		if (!id) return '';
		const sprint = sprints.find(s => s.id === id);
		return sprint?.name || '';
	}

	function getVersionName(id: string | undefined): string {
		if (!id) return '';
		const version = versions.find(v => v.id === id);
		return version?.name || '';
	}

	function getEmployeePhoto(id: string): string | null {
		if (!id) return null;
		const emp = employees.find(e => e.id === id);
		return emp?.photo_base64 || null;
	}

	function formatDate(dateStr: string | undefined): string {
		if (!dateStr) return '';
		const date = new Date(dateStr);
		const today = new Date();
		const tomorrow = new Date(today);
		tomorrow.setDate(tomorrow.getDate() + 1);

		if (date.toDateString() === today.toDateString()) return 'Сегодня';
		if (date.toDateString() === tomorrow.toDateString()) return 'Завтра';
		return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
	}

	function isOverdue(dateStr: string | undefined): boolean {
		if (!dateStr) return false;
		return new Date(dateStr) < new Date();
	}

	// Drag and drop for Kanban
	let draggedTask: Task | null = $state(null);

	function handleDragStart(e: DragEvent, task: Task) {
		draggedTask = task;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
		}
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		if (e.dataTransfer) {
			e.dataTransfer.dropEffect = 'move';
		}
	}

	function handleDrop(e: DragEvent, newStatus: string) {
		e.preventDefault();
		if (draggedTask && draggedTask.status !== newStatus) {
			updateTaskStatus(draggedTask, newStatus);
		}
		draggedTask = null;
	}
</script>

<svelte:head>
	<title>Задачи - EKF Hub</title>
</svelte:head>

<div class="flex h-[calc(100vh-4rem)] -m-4">
	<!-- Project Sidebar -->
	<aside class="w-60 bg-ekf-dark text-white flex flex-col flex-shrink-0">
		<!-- Project Selector -->
		<div class="p-4 border-b border-gray-700">
			<button class="flex items-center gap-2 w-full hover:bg-white/5 rounded-lg p-1 -m-1 transition-colors">
				<div class="w-8 h-8 bg-ekf-red rounded flex items-center justify-center text-white text-xs font-bold">EH</div>
				<div class="flex-1 min-w-0 text-left">
					<div class="font-semibold text-sm truncate">EKF Hub</div>
					<div class="text-xs text-gray-400">Активный проект</div>
				</div>
				<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
			</button>
		</div>

		<nav class="flex-1 p-3 space-y-1 overflow-y-auto">
			<!-- Планирование -->
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-2 pb-1">Планирование</div>
			<a href="/tasks" class="flex items-center gap-3 px-3 py-2 rounded-lg bg-ekf-red text-white">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2"></path></svg>
				<span>Доска задач</span>
			</a>
			<a href="/tasks?view=backlog" class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-300 hover:bg-gray-700">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16"></path></svg>
				<span>Бэклог</span>
				<span class="ml-auto text-xs text-gray-400">{tasks.filter(t => t.status === 'backlog').length}</span>
			</a>
			<a href="/tasks?view=roadmap" class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-300 hover:bg-gray-700">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path></svg>
				<span>Roadmap</span>
			</a>

			<!-- Спринты -->
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-4 pb-1">Спринты</div>
			{#if activeSprint}
				<button
					onclick={() => filterSprint = activeSprint?.id || ''}
					class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-300 hover:bg-gray-700 w-full text-left {filterSprint === activeSprint.id ? 'bg-gray-700' : ''}"
				>
					<div class="w-5 h-5 flex items-center justify-center">
						<div class="w-2 h-2 bg-green-400 rounded-full"></div>
					</div>
					<span class="truncate">{activeSprint.name}</span>
					<span class="ml-auto text-xs bg-green-500/20 text-green-400 px-1.5 py-0.5 rounded">Active</span>
				</button>
			{/if}
			{#each sprints.filter(s => s.id !== activeSprint?.id && s.status === 'planning').slice(0, 2) as sprint}
				<button
					onclick={() => filterSprint = sprint.id}
					class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-300 hover:bg-gray-700 w-full text-left {filterSprint === sprint.id ? 'bg-gray-700' : ''}"
				>
					<div class="w-5 h-5 flex items-center justify-center">
						<div class="w-2 h-2 bg-gray-500 rounded-full"></div>
					</div>
					<span class="truncate">{sprint.name}</span>
					<span class="ml-auto text-xs text-gray-500">Planned</span>
				</button>
			{/each}
			<a href="/sprints" class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-300 hover:bg-gray-700">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"></path></svg>
				<span>Архив спринтов</span>
			</a>

			<!-- Релизы -->
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-4 pb-1">Релизы</div>
			{#each versions.filter(v => v.status === 'unreleased').slice(0, 2) as version}
				<button
					onclick={() => { /* filter by version */ }}
					class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-300 hover:bg-gray-700 w-full text-left"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"></path></svg>
					<span class="truncate">{version.name}</span>
					<span class="ml-auto text-xs bg-yellow-500/20 text-yellow-400 px-1.5 py-0.5 rounded">Dev</span>
				</button>
			{/each}
			{#each versions.filter(v => v.status === 'released').slice(0, 1) as version}
				<button class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-300 hover:bg-gray-700 w-full text-left">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
					<span class="truncate">{version.name}</span>
					<span class="ml-auto text-xs text-gray-500">Released</span>
				</button>
			{/each}

			<!-- Тестирование -->
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-4 pb-1">Тестирование</div>
			<a href="/test-plans" class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-300 hover:bg-gray-700">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"></path></svg>
				<span>Тест-планы</span>
			</a>
			<a href="/test-cases" class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-300 hover:bg-gray-700">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"></path></svg>
				<span>Тест-кейсы</span>
			</a>
			<a href="/test-runs" class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-300 hover:bg-gray-700">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
				<span>Прогоны</span>
			</a>

			<!-- Документация -->
			<div class="text-xs text-gray-500 uppercase tracking-wider px-3 pt-4 pb-1">Документация</div>
			<a href="/wiki" class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-300 hover:bg-gray-700">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"></path></svg>
				<span>Wiki</span>
			</a>
			<a href="/requirements" class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-300 hover:bg-gray-700">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
				<span>Требования</span>
			</a>

			<div class="pt-4 mt-4 border-t border-gray-700">
				<a href="/project-settings" class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-300 hover:bg-gray-700">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>
					<span>Настройки проекта</span>
				</a>
			</div>
		</nav>

		<!-- User -->
		{#if $user}
			<div class="p-4 border-t border-gray-700">
				<div class="flex items-center gap-3">
					{#if $user.photo_base64}
						<img src="data:image/jpeg;base64,{$user.photo_base64}" alt={$user.name} class="w-10 h-10 rounded-full object-cover" />
					{:else}
						<div class="w-10 h-10 rounded-full bg-gradient-to-br from-ekf-red to-red-600 flex items-center justify-center text-white font-bold">
							{$user.name?.charAt(0) || '?'}
						</div>
					{/if}
					<div class="flex-1 min-w-0">
						<div class="font-medium text-sm truncate">{$user.name}</div>
						<div class="text-xs text-gray-400">{$user.position || 'Сотрудник'}</div>
					</div>
					<div class="w-2 h-2 bg-green-400 rounded-full"></div>
				</div>
			</div>
		{/if}
	</aside>

	<!-- Main Content -->
	<main class="flex-1 flex flex-col overflow-hidden bg-gray-100">
		<!-- Sprint Header -->
		{#if activeSprint && (filterSprint === activeSprint.id || filterSprint === '')}
			<div class="bg-white border-b px-6 py-4">
				<div class="flex items-center justify-between">
					<div>
						<div class="flex items-center gap-3">
							<h1 class="text-xl font-semibold">{activeSprint.name} — Доска задач</h1>
							<span class="text-xs bg-green-100 text-green-700 px-2 py-1 rounded-full">Активный</span>
						</div>
						<div class="text-sm text-gray-500 mt-1">
							{#if activeSprint.start_date && activeSprint.end_date}
								{new Date(activeSprint.start_date).toLocaleDateString('ru-RU', {day: 'numeric', month: 'short'})} — {new Date(activeSprint.end_date).toLocaleDateString('ru-RU', {day: 'numeric', month: 'short'})} •
							{/if}
							{getCompletedTasksCount()} из {filteredTasks.length} задач выполнено
						</div>
					</div>
					<div class="flex items-center gap-3">
						<div class="flex items-center gap-2">
							<span class="text-sm text-gray-500">Прогресс:</span>
							<div class="w-32 h-2 bg-gray-200 rounded-full overflow-hidden">
								<div class="h-full bg-ekf-red rounded-full transition-all" style="width: {filteredTasks.length > 0 ? Math.round((getCompletedTasksCount() / filteredTasks.length) * 100) : 0}%"></div>
							</div>
							<span class="text-sm font-medium">{filteredTasks.length > 0 ? Math.round((getCompletedTasksCount() / filteredTasks.length) * 100) : 0}%</span>
						</div>
						<button
							onclick={openNewTask}
							class="px-4 py-2 bg-ekf-red text-white rounded-lg text-sm font-medium hover:bg-red-700 flex items-center gap-2"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
							Новая задача
						</button>
					</div>
				</div>
			</div>
		{:else}
			<div class="bg-white border-b px-6 py-4">
				<div class="flex items-center justify-between">
					<div>
						<h1 class="text-xl font-semibold">Задачи</h1>
						<div class="flex items-center gap-3 text-sm text-gray-500">
							<span>{filteredTasks.length} задач</span>
							{#if getTotalStoryPoints() > 0}
								<span class="text-indigo-600 font-medium">{getTotalStoryPoints()} SP</span>
							{/if}
							<span class="flex items-center gap-1">
								<span class="w-2 h-2 rounded-full bg-yellow-400"></span>
								{getInProgressTasksCount()} в работе
							</span>
							<span class="flex items-center gap-1">
								<span class="w-2 h-2 rounded-full bg-green-500"></span>
								{getCompletedTasksCount()} выполнено
							</span>
						</div>
					</div>
					<button
						onclick={openNewTask}
						class="px-4 py-2 bg-ekf-red text-white rounded-lg text-sm font-medium hover:bg-red-700 flex items-center gap-2"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
						Новая задача
					</button>
				</div>
			</div>
		{/if}

		<!-- Filters & View Toggle -->
		<div class="bg-white border-b px-6 py-3 flex flex-wrap items-center gap-4">
			<!-- Search -->
			<div class="relative">
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Поиск задач..."
					class="pl-9 pr-4 py-1.5 border rounded-lg text-sm w-64 focus:ring-2 focus:ring-ekf-red focus:border-ekf-red"
				/>
				<svg class="w-4 h-4 text-gray-400 absolute left-3 top-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
			</div>

			<!-- Filters -->
			<select bind:value={filterAssignee} class="border rounded-lg px-3 py-1.5 text-sm">
				<option value="">Все исполнители</option>
				<option value={$user?.id}>Мои задачи</option>
				{#each $subordinates as sub}
					<option value={sub.id}>{sub.name}</option>
				{/each}
			</select>

			<select bind:value={filterStatus} class="border rounded-lg px-3 py-1.5 text-sm">
				<option value="">Все статусы</option>
				{#each statusColumns as col}
					<option value={col.id}>{col.label}</option>
				{/each}
			</select>

			<select bind:value={filterProject} class="border rounded-lg px-3 py-1.5 text-sm">
				<option value="">Все типы</option>
				{#each projects as project}
					<option value={project.id}>{project.name}</option>
				{/each}
			</select>

			<!-- View Toggle -->
			<div class="flex rounded-lg border border-gray-200 overflow-hidden ml-auto">
				<button
					onclick={() => viewMode = 'list'}
					class="px-3 py-1.5 text-sm transition-colors {viewMode === 'list' ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}"
				>
					Список
				</button>
				<button
					onclick={() => viewMode = 'kanban'}
					class="px-3 py-1.5 text-sm transition-colors {viewMode === 'kanban' ? 'bg-ekf-red text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}"
				>
					Kanban
				</button>
			</div>
		</div>

		<!-- Content Area -->
		<div class="flex-1 overflow-auto p-6">
			{#if loading}
				<div class="flex items-center justify-center h-48">
					<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-red"></div>
				</div>
			{:else if viewMode === 'list'}
		<!-- List View -->
		<div class="bg-white rounded-lg shadow-sm overflow-hidden">
			{#if filteredTasks.length === 0}
				<div class="text-center py-12">
					<svg class="w-12 h-12 mx-auto text-gray-300 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
					</svg>
					<p class="text-gray-500 mb-3">Нет задач</p>
					<button onclick={openNewTask} class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 text-sm">
						Создать первую задачу
					</button>
				</div>
			{:else}
				<table class="w-full text-sm">
					<thead class="bg-gray-50 border-b">
						<tr>
							<th class="px-4 py-2 text-left font-medium text-gray-500">Задача</th>
							<th class="px-4 py-2 text-left font-medium text-gray-500 w-32">Статус</th>
							<th class="px-4 py-2 text-left font-medium text-gray-500 w-20">Приор.</th>
							<th class="px-4 py-2 text-center font-medium text-gray-500 w-12" title="Story Points">SP</th>
							<th class="px-4 py-2 text-left font-medium text-gray-500 w-36">Исполнитель</th>
							<th class="px-4 py-2 text-left font-medium text-gray-500 w-24">Срок</th>
							<th class="px-4 py-2 w-16"></th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100">
						{#each filteredTasks as task (task.id)}
							<tr class="hover:bg-gray-50 transition-colors">
								<td class="px-4 py-2">
									<button onclick={() => openEditTask(task)} class="text-left group">
										<div class="font-medium text-gray-900 group-hover:text-ekf-red">{task.title}</div>
										{#if task.project_id}
											<div class="text-xs text-gray-400">{getProjectName(task.project_id)}</div>
										{/if}
									</button>
								</td>
								<td class="px-4 py-2">
									<select
										value={task.status}
										onchange={(e) => updateTaskStatus(task, (e.target as HTMLSelectElement).value)}
										class="w-full px-2 py-1 text-xs border border-gray-200 rounded focus:outline-none focus:ring-1 focus:ring-ekf-red"
									>
										{#each statusColumns as col}
											<option value={col.id}>{col.label}</option>
										{/each}
									</select>
								</td>
								<td class="px-4 py-2">
									<span class="px-2 py-0.5 text-xs rounded {priorityLabels[task.priority || 3].color}">
										P{task.priority || 3}
									</span>
								</td>
								<td class="px-4 py-2 text-center">
									{#if task.story_points}
										<span class="px-1.5 py-0.5 text-xs rounded bg-indigo-50 text-indigo-600 font-medium">
											{task.story_points}
										</span>
									{:else}
										<span class="text-gray-300">—</span>
									{/if}
								</td>
								<td class="px-4 py-2 text-gray-600 text-xs">{getEmployeeName(task.assignee_id || '')}</td>
								<td class="px-4 py-2">
									{#if task.due_date}
										<span class="{isOverdue(task.due_date) && task.status !== 'done' ? 'text-red-600 font-medium' : 'text-gray-600'}">
											{formatDate(task.due_date)}
										</span>
									{:else}
										<span class="text-gray-300">—</span>
									{/if}
								</td>
								<td class="px-4 py-2">
									<button
										onclick={() => deleteTask(task.id)}
										class="p-1 text-gray-400 hover:text-red-600 rounded transition-colors"
										title="Удалить"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
										</svg>
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>
	{:else}
		<!-- Kanban View -->
		<div class="flex gap-3 overflow-x-auto pb-4" style="min-height: calc(100vh - 280px);">
			{#each statusColumns as column}
				<div
					class="flex-shrink-0 w-72 flex flex-col rounded-lg {column.color} {isWipLimitExceeded(column.id) ? 'ring-2 ring-red-400' : ''}"
					ondragover={handleDragOver}
					ondrop={(e) => handleDrop(e, column.id)}
				>
					<div class="p-3 font-medium text-gray-700 border-b border-gray-200/50">
						<div class="flex items-center justify-between">
							<span>{column.label}</span>
							<div class="flex items-center gap-1.5">
								{#if getStoryPointsByStatus(column.id) > 0}
									<span class="text-xs bg-indigo-100 text-indigo-600 px-1.5 py-0.5 rounded font-medium" title="Story Points">
										{getStoryPointsByStatus(column.id)} SP
									</span>
								{/if}
								<span class="text-xs px-1.5 py-0.5 rounded {isWipLimitExceeded(column.id) ? 'bg-red-100 text-red-600 font-bold' : 'bg-white/80'}">
									{getTasksByStatus(column.id).length}{#if column.wipLimit > 0}/{column.wipLimit}{/if}
								</span>
							</div>
						</div>
						{#if isWipLimitExceeded(column.id)}
							<div class="mt-1 text-xs text-red-600 flex items-center gap-1">
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
								</svg>
								WIP лимит превышен!
							</div>
						{/if}
						<!-- Quick Add Button -->
						<button
							onclick={() => quickAddColumn = quickAddColumn === column.id ? null : column.id}
							class="mt-2 w-full py-1.5 text-xs text-gray-500 hover:text-gray-700 hover:bg-white/50 rounded border border-dashed border-gray-300 flex items-center justify-center gap-1 transition-colors"
						>
							<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
							</svg>
							Добавить
						</button>
						{#if quickAddColumn === column.id}
							<div class="mt-2">
								<input
									type="text"
									bind:value={quickAddTitle}
									placeholder="Название задачи"
									class="w-full px-2 py-1.5 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-ekf-red"
									onkeydown={(e) => e.key === 'Enter' && quickAddTask(column.id)}
								/>
								<div class="flex gap-1 mt-1">
									<button
										onclick={() => quickAddTask(column.id)}
										class="flex-1 py-1 bg-ekf-red text-white text-xs rounded hover:bg-red-700"
									>
										Создать
									</button>
									<button
										onclick={() => { quickAddColumn = null; quickAddTitle = ''; }}
										class="px-2 py-1 text-gray-500 text-xs hover:bg-gray-100 rounded"
									>
										Отмена
									</button>
								</div>
							</div>
						{/if}
					</div>
					<div class="flex-1 p-2 space-y-2 overflow-y-auto">
						{#each getTasksByStatus(column.id) as task (task.id)}
							<div
								class="task-card bg-white rounded-lg p-4 shadow-sm border border-gray-100 cursor-pointer transition-all duration-150 hover:-translate-y-0.5 hover:shadow-lg border-l-[3px] {priorityBorderColors[task.priority || 3]} {task.status === 'done' ? 'opacity-75' : ''}"
								draggable="true"
								ondragstart={(e) => handleDragStart(e, task)}
								onclick={() => openEditTask(task)}
							>
								<!-- Header: Task ID + Type Badge -->
								<div class="flex items-start justify-between mb-2">
									<span class="text-xs text-gray-400 font-mono">EKF-{task.id.substring(0, 4).toUpperCase()}</span>
									{#if task.task_type && taskTypeLabels[task.task_type]}
										<span class="text-xs px-2 py-0.5 rounded-full font-medium {taskTypeLabels[task.task_type].color}">
											{taskTypeLabels[task.task_type].icon} {taskTypeLabels[task.task_type].label}
										</span>
									{/if}
								</div>

								<!-- Title -->
								<h4 class="font-medium text-gray-900 text-sm mb-2 line-clamp-2">{task.title}</h4>

								<!-- Sprint badge -->
								{#if task.sprint_id}
									<div class="mb-2">
										<span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] bg-blue-50 text-blue-600 font-medium">
											⚡ {getSprintName(task.sprint_id)}
										</span>
									</div>
								{/if}

								<!-- Footer: Assignee + Story Points -->
								<div class="flex items-center justify-between pt-2 border-t border-gray-50">
									<div class="flex items-center gap-2">
										{#if task.assignee_id}
											{@const photo = getEmployeePhoto(task.assignee_id)}
											{#if photo}
												<img
													src="data:image/jpeg;base64,{photo}"
													alt={getEmployeeName(task.assignee_id)}
													class="w-6 h-6 rounded-full object-cover ring-2 ring-white shadow-sm"
												/>
											{:else}
												<div class="w-6 h-6 rounded-full bg-gradient-to-br from-ekf-red to-red-600 text-white flex items-center justify-center text-[10px] font-bold ring-2 ring-white shadow-sm">
													{getEmployeeName(task.assignee_id).charAt(0)}
												</div>
											{/if}
											<span class="text-xs text-gray-500 truncate max-w-[100px]">{getEmployeeName(task.assignee_id).split(' ')[0]}</span>
										{:else}
											<div class="w-6 h-6 rounded-full bg-gray-200 flex items-center justify-center">
												<svg class="w-3 h-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
												</svg>
											</div>
										{/if}
									</div>
									<div class="flex items-center gap-1.5">
										{#if task.due_date}
											<span class="text-[10px] {isOverdue(task.due_date) && task.status !== 'done' ? 'text-red-600 font-medium' : 'text-gray-400'}">
												{formatDate(task.due_date)}
											</span>
										{/if}
										{#if task.story_points}
											<span class="px-1.5 py-0.5 rounded bg-indigo-100 text-indigo-700 text-[10px] font-bold">
												{task.story_points} SP
											</span>
										{/if}
									</div>
								</div>
							</div>
						{/each}
						{#if getTasksByStatus(column.id).length === 0}
							<div class="text-center py-8 text-gray-400 text-sm">
								Перетащите задачу сюда
							</div>
						{/if}
					</div>
				</div>
			{/each}
			</div>
		{/if}
		</div>
	</main>
</div>

<!-- Task Modal -->
{#if showTaskModal && editingTask}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onclick={() => showTaskModal = false}>
		<div class="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()}>
			<div class="p-4 border-b flex items-center justify-between">
				<h2 class="font-bold text-gray-900">{editingTask.id ? 'Редактировать задачу' : 'Новая задача'}</h2>
				<button onclick={() => showTaskModal = false} class="p-1 hover:bg-gray-100 rounded">
					<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>
			<form onsubmit={(e) => { e.preventDefault(); saveTask(); }} class="p-4 space-y-3">
				<div>
					<label class="block text-xs font-medium text-gray-500 mb-1">Название *</label>
					<input
						type="text"
						bind:value={editingTask.title}
						required
						class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-1 focus:ring-ekf-red"
						placeholder="Что нужно сделать?"
					/>
				</div>
				<div>
					<label class="block text-xs font-medium text-gray-500 mb-1">Описание</label>
					<textarea
						bind:value={editingTask.description}
						rows="2"
						class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm resize-none focus:outline-none focus:ring-1 focus:ring-ekf-red"
						placeholder="Подробности..."
					></textarea>
				</div>
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">Проект</label>
						<select bind:value={editingTask.project_id} class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm">
							<option value="">Без проекта</option>
							{#each projects as project}
								<option value={project.id}>{project.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">Исполнитель</label>
						<select bind:value={editingTask.assignee_id} class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm">
							<option value="">Не назначен</option>
							<option value={$user?.id}>Я ({$user?.name})</option>
							{#each $subordinates as sub}
								{#if sub.id !== $user?.id}
									<option value={sub.id}>{sub.name}</option>
								{/if}
							{/each}
						</select>
					</div>
				</div>
				<div class="grid grid-cols-5 gap-3">
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">Тип</label>
						<select bind:value={editingTask.task_type} class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm">
							<option value="task">📋 Задача</option>
							<option value="feature">✨ Фича</option>
							<option value="bug">🐛 Баг</option>
							<option value="tech_debt">🔧 Техдолг</option>
							<option value="improvement">📈 Улучшение</option>
						</select>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">Статус</label>
						<select bind:value={editingTask.status} class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm">
							{#each statusColumns as col}
								<option value={col.id}>{col.label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">Приоритет</label>
						<select bind:value={editingTask.priority} class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm">
							<option value={1}>P1 - Критический</option>
							<option value={2}>P2 - Высокий</option>
							<option value={3}>P3 - Средний</option>
							<option value={4}>P4 - Низкий</option>
							<option value={5}>P5 - Минимальный</option>
						</select>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">SP</label>
						<select bind:value={editingTask.story_points} class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm">
							<option value={undefined}>—</option>
							{#each storyPointsOptions as sp}
								<option value={sp}>{sp}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">Срок</label>
						<input
							type="date"
							bind:value={editingTask.due_date}
							class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm"
						/>
					</div>
				</div>

				<!-- Sprint and Version -->
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">
							<span class="flex items-center gap-1">
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
								</svg>
								Спринт
							</span>
						</label>
						<select bind:value={editingTask.sprint_id} class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm">
							<option value="">Без спринта</option>
							{#each sprints.filter(s => s.status !== 'completed') as sprint}
								<option value={sprint.id}>
									{sprint.name}
									{#if sprint.status === 'active'}
										(активный)
									{/if}
								</option>
							{/each}
							{#if sprints.filter(s => s.status === 'completed').length > 0}
								<optgroup label="Завершённые">
									{#each sprints.filter(s => s.status === 'completed') as sprint}
										<option value={sprint.id}>{sprint.name}</option>
									{/each}
								</optgroup>
							{/if}
						</select>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-500 mb-1">
							<span class="flex items-center gap-1">
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
								</svg>
								Версия (Fix Version)
							</span>
						</label>
						<select bind:value={editingTask.fix_version_id} class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm">
							<option value="">Без версии</option>
							{#each versions.filter(v => v.status === 'unreleased') as version}
								<option value={version.id}>{version.name}</option>
							{/each}
							{#if versions.filter(v => v.status === 'released').length > 0}
								<optgroup label="Выпущенные">
									{#each versions.filter(v => v.status === 'released') as version}
										<option value={version.id}>{version.name}</option>
									{/each}
								</optgroup>
							{/if}
						</select>
					</div>
				</div>

				<!-- Dependencies Section (only for existing tasks) -->
				{#if editingTask.id}
					<div class="border-t pt-3 mt-3">
						<div class="flex items-center justify-between mb-2">
							<label class="block text-xs font-medium text-gray-500">Зависимости (блокирующие задачи)</label>
							{#if loadingTaskDetails}
								<span class="text-xs text-gray-400">Загрузка...</span>
							{:else}
								<button
									type="button"
									onclick={() => showDependencyPicker = !showDependencyPicker}
									class="text-xs text-ekf-red hover:text-red-700"
								>
									+ Добавить
								</button>
							{/if}
						</div>

						{#if loadingTaskDetails}
							<div class="flex items-center justify-center py-4">
								<svg class="w-5 h-5 animate-spin text-gray-400" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
							</div>
						{:else}
						<!-- Blocked warning -->
						{#if isTaskBlocked}
							<div class="mb-2 p-2 bg-yellow-50 border border-yellow-200 rounded-lg text-xs text-yellow-800">
								<span class="font-medium">⚠️ Задача заблокирована.</span> Ожидает завершения:
								<ul class="mt-1 list-disc list-inside">
									{#each taskBlockers as blocker}
										<li>{blocker.title}</li>
									{/each}
								</ul>
							</div>
						{/if}

						<!-- Current dependencies -->
						{#if taskDependencies.length > 0}
							<div class="space-y-1 mb-2">
								{#each taskDependencies as dep}
									<div class="flex items-center justify-between p-2 bg-gray-50 rounded text-xs">
										<div class="flex items-center gap-2">
											<span class={dep.depends_on_task?.status === 'done' ? 'text-green-600' : 'text-orange-600'}>
												{dep.depends_on_task?.status === 'done' ? '✓' : '○'}
											</span>
											<span class="truncate">{dep.depends_on_task?.title || 'Неизвестно'}</span>
										</div>
										<button
											type="button"
											onclick={() => removeDependency(dep.id)}
											class="text-gray-400 hover:text-red-500"
										>
											×
										</button>
									</div>
								{/each}
							</div>
						{:else}
							<p class="text-xs text-gray-400 mb-2">Нет зависимостей</p>
						{/if}

						<!-- Dependency picker -->
						{#if showDependencyPicker}
							<div class="border rounded-lg p-2 bg-white">
								<input
									type="text"
									bind:value={dependencySearch}
									placeholder="Поиск задачи..."
									class="w-full px-2 py-1 border border-gray-200 rounded text-xs mb-2"
								/>
								<div class="max-h-32 overflow-y-auto space-y-1">
									{#each getAvailableDependencies().slice(0, 10) as task}
										<button
											type="button"
											onclick={() => addDependency(task.id)}
											class="w-full text-left px-2 py-1 hover:bg-gray-100 rounded text-xs truncate"
										>
											{task.title}
										</button>
									{:else}
										<p class="text-xs text-gray-400 p-2">Нет доступных задач</p>
									{/each}
								</div>
							</div>
						{/if}

						<!-- Dependents info -->
						{#if taskDependents.length > 0}
							<div class="mt-3 pt-2 border-t">
								<p class="text-xs text-gray-500 mb-1">Блокирует другие задачи ({taskDependents.length}):</p>
								<div class="text-xs text-gray-400">
									{#each taskDependents.slice(0, 3) as dep}
										<span class="mr-2">• {dep.depends_on_task?.title}</span>
									{/each}
									{#if taskDependents.length > 3}
										<span>и ещё {taskDependents.length - 3}...</span>
									{/if}
								</div>
							</div>
						{/if}
						{/if}
					</div>

					<!-- Resource Planning Section -->
					<div class="border-t pt-3 mt-3">
						<label class="block text-xs font-medium text-gray-500 mb-2">Планирование ресурсов</label>

						<div class="grid grid-cols-2 gap-2 mb-3">
							<div>
								<label class="block text-xs text-gray-400 mb-1">Оценка (часы)</label>
								<input
									type="number"
									min="0"
									step="0.5"
									bind:value={editingTask.estimated_hours}
									class="w-full px-2 py-1 border border-gray-200 rounded text-sm"
									placeholder="0"
								/>
							</div>
							<div>
								<label class="block text-xs text-gray-400 mb-1">Оценка (руб.)</label>
								<input
									type="number"
									min="0"
									bind:value={editingTask.estimated_cost}
									class="w-full px-2 py-1 border border-gray-200 rounded text-sm"
									placeholder="0"
								/>
							</div>
						</div>

						{#if resourceSummary}
							<div class="p-2 bg-gray-50 rounded-lg text-xs space-y-1">
								<div class="flex justify-between">
									<span class="text-gray-500">Залогировано:</span>
									<span class="font-medium">{formatHours(resourceSummary.logged_hours)}</span>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-500">Фактические часы:</span>
									<span class="font-medium">{formatHours(resourceSummary.actual_hours)}</span>
								</div>
								{#if resourceSummary.hourly_rate > 0}
									<div class="flex justify-between">
										<span class="text-gray-500">Расчётная стоимость:</span>
										<span class="font-medium">{formatCost(resourceSummary.calculated_cost)}</span>
									</div>
								{/if}
								{#if editingTask.estimated_hours && resourceSummary.logged_hours}
									<div class="mt-2 pt-2 border-t">
										<div class="flex justify-between items-center">
											<span class="text-gray-500">Прогресс:</span>
											<span class="font-medium {resourceSummary.logged_hours > (editingTask.estimated_hours || 0) ? 'text-red-600' : 'text-green-600'}">
												{Math.round((resourceSummary.logged_hours / (editingTask.estimated_hours || 1)) * 100)}%
											</span>
										</div>
										<div class="w-full bg-gray-200 rounded-full h-1.5 mt-1">
											<div
												class="h-1.5 rounded-full transition-all {resourceSummary.logged_hours > (editingTask.estimated_hours || 0) ? 'bg-red-500' : 'bg-green-500'}"
												style="width: {Math.min(100, (resourceSummary.logged_hours / (editingTask.estimated_hours || 1)) * 100)}%"
											></div>
										</div>
									</div>
								{/if}
							</div>
						{/if}
					</div>

					<!-- Time Tracking Section -->
					<div class="border-t pt-3 mt-3">
						<div class="flex items-center justify-between mb-2">
							<label class="block text-xs font-medium text-gray-500">Учёт времени</label>
							<button
								type="button"
								onclick={() => showTimeEntryForm = !showTimeEntryForm}
								class="text-xs text-ekf-red hover:text-red-700"
							>
								+ Добавить время
							</button>
						</div>

						{#if showTimeEntryForm}
							<div class="border rounded-lg p-2 bg-white mb-2">
								<div class="grid grid-cols-3 gap-2 mb-2">
									<div>
										<label class="block text-xs text-gray-400 mb-1">Часы</label>
										<input
											type="number"
											min="0.25"
											step="0.25"
											bind:value={newTimeEntry.hours}
											class="w-full px-2 py-1 border border-gray-200 rounded text-xs"
											placeholder="0"
										/>
									</div>
									<div class="col-span-2">
										<label class="block text-xs text-gray-400 mb-1">Дата</label>
										<input
											type="date"
											bind:value={newTimeEntry.date}
											class="w-full px-2 py-1 border border-gray-200 rounded text-xs"
										/>
									</div>
								</div>
								<input
									type="text"
									bind:value={newTimeEntry.description}
									placeholder="Описание работы..."
									class="w-full px-2 py-1 border border-gray-200 rounded text-xs mb-2"
								/>
								<div class="flex justify-end gap-2">
									<button
										type="button"
										onclick={() => showTimeEntryForm = false}
										class="px-2 py-1 text-gray-500 text-xs"
									>
										Отмена
									</button>
									<button
										type="button"
										onclick={addTimeEntry}
										disabled={newTimeEntry.hours <= 0}
										class="px-2 py-1 bg-ekf-red text-white rounded text-xs disabled:opacity-50"
									>
										Добавить
									</button>
								</div>
							</div>
						{/if}

						{#if timeEntries.length > 0}
							<div class="space-y-1 max-h-32 overflow-y-auto">
								{#each timeEntries as entry}
									<div class="flex items-center justify-between p-2 bg-gray-50 rounded text-xs">
										<div class="flex-1">
											<div class="flex items-center gap-2">
												<span class="font-medium">{entry.hours} ч</span>
												<span class="text-gray-400">{new Date(entry.date).toLocaleDateString('ru-RU')}</span>
											</div>
											{#if entry.description}
												<p class="text-gray-500 truncate">{entry.description}</p>
											{/if}
										</div>
										<button
											type="button"
											onclick={() => deleteTimeEntry(entry.id)}
											class="text-gray-400 hover:text-red-500 ml-2"
										>
											×
										</button>
									</div>
								{/each}
							</div>
						{:else}
							<p class="text-xs text-gray-400">Нет записей времени</p>
						{/if}
					</div>

					<!-- GitHub Commits Section -->
					{#if githubConfigured}
						<div class="border-t pt-3 mt-3">
							<div class="flex items-center justify-between mb-2">
								<label class="block text-xs font-medium text-gray-500 flex items-center gap-1.5">
									<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
										<path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
									</svg>
									Связанные коммиты
								</label>
								{#if loadingCommits}
									<span class="text-xs text-gray-400">Загрузка...</span>
								{/if}
							</div>

							{#if loadingCommits}
								<div class="flex items-center justify-center py-4">
									<svg class="w-5 h-5 animate-spin text-gray-400" fill="none" viewBox="0 0 24 24">
										<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
										<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
									</svg>
								</div>
							{:else if taskCommits.length > 0}
								<div class="space-y-1.5 max-h-40 overflow-y-auto">
									{#each taskCommits as commit}
										<a
											href={commit.html_url}
											target="_blank"
											rel="noopener noreferrer"
											class="block p-2 bg-gray-50 hover:bg-gray-100 rounded text-xs transition-colors"
										>
											<div class="flex items-start gap-2">
												{#if commit.author?.avatar_url}
													<img
														src={commit.author.avatar_url}
														alt={commit.author.login}
														class="w-5 h-5 rounded-full flex-shrink-0"
													/>
												{:else}
													<div class="w-5 h-5 rounded-full bg-gray-300 flex-shrink-0"></div>
												{/if}
												<div class="flex-1 min-w-0">
													<div class="flex items-center gap-1.5">
														<code class="text-xs bg-gray-200 px-1 py-0.5 rounded font-mono text-gray-600">{shortenSha(commit.sha)}</code>
														<span class="text-gray-400">{formatCommitDate(commit.commit.author.date)}</span>
													</div>
													<p class="text-gray-700 truncate mt-0.5">{commit.commit.message.split('\n')[0]}</p>
													<span class="text-gray-500">{commit.commit.author.name}</span>
												</div>
											</div>
										</a>
									{/each}
								</div>
							{:else}
								<div class="text-xs text-gray-400 py-2">
									<p>Нет коммитов, упоминающих задачу #{editingTask.id?.substring(0, 8)}</p>
									<p class="mt-1 text-gray-300">Используйте ID задачи в сообщениях коммитов</p>
								</div>
							{/if}

							<!-- Pull Requests Section -->
							<div class="mt-3 pt-3 border-t border-gray-100">
								<label class="block text-xs font-medium text-gray-500 flex items-center gap-1.5 mb-2">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"/>
									</svg>
									Связанные Pull Requests
								</label>

								{#if taskPullRequests.length > 0}
									<div class="space-y-1.5 max-h-40 overflow-y-auto">
										{#each taskPullRequests as pr}
											<a
												href={pr.html_url}
												target="_blank"
												rel="noopener noreferrer"
												class="block p-2 bg-gray-50 hover:bg-gray-100 rounded text-xs transition-colors"
											>
												<div class="flex items-start gap-2">
													{#if pr.user?.avatar_url}
														<img
															src={pr.user.avatar_url}
															alt={pr.user.login}
															class="w-5 h-5 rounded-full flex-shrink-0"
														/>
													{:else}
														<div class="w-5 h-5 rounded-full bg-gray-300 flex-shrink-0"></div>
													{/if}
													<div class="flex-1 min-w-0">
														<div class="flex items-center gap-1.5">
															<span class="font-medium text-gray-700">#{pr.number}</span>
															<span class={`px-1.5 py-0.5 rounded text-[10px] font-medium ${
																pr.state === 'open' ? 'bg-green-100 text-green-700' :
																pr.merged_at ? 'bg-purple-100 text-purple-700' :
																'bg-red-100 text-red-700'
															}`}>
																{pr.state === 'open' ? 'Open' : pr.merged_at ? 'Merged' : 'Closed'}
															</span>
															<span class="text-gray-400 text-[10px]">{formatCommitDate(pr.created_at)}</span>
														</div>
														<p class="text-gray-700 truncate mt-0.5">{pr.title}</p>
														<span class="text-gray-500">{pr.user?.login}</span>
													</div>
												</div>
											</a>
										{/each}
									</div>
								{:else}
									<div class="text-xs text-gray-400 py-2">
										<p>Нет PR, упоминающих задачу #{editingTask.id?.substring(0, 8)}</p>
									</div>
								{/if}
							</div>
						</div>
					{/if}
				{/if}

				<div class="flex justify-end gap-2 pt-2">
					<button
						type="button"
						onclick={() => showTaskModal = false}
						class="px-4 py-2 text-gray-600 hover:text-gray-900 text-sm"
					>
						Отмена
					</button>
					<button
						type="submit"
						class="px-4 py-2 bg-ekf-red text-white rounded-lg hover:bg-red-700 text-sm"
					>
						{editingTask.id ? 'Сохранить' : 'Создать'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
