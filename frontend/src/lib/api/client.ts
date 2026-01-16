import { browser } from '$app/environment';

const BASE_URL = browser ? (import.meta.env.VITE_API_URL || 'http://localhost:8080') : 'http://localhost:8080';
const API_URL = `${BASE_URL}/api/v1`;

interface RequestOptions {
	method?: string;
	body?: unknown;
	headers?: Record<string, string>;
}

async function request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
	const { method = 'GET', body, headers = {} } = options;

	const config: RequestInit = {
		method,
		headers: {
			'Content-Type': 'application/json',
			...headers
		}
	};

	if (body) {
		config.body = JSON.stringify(body);
	}

	const response = await fetch(`${API_URL}${endpoint}`, config);

	if (!response.ok) {
		const error = await response.json().catch(() => ({ error: response.statusText }));
		throw new Error(error.error || error.message || 'Request failed');
	}

	return response.json();
}

// Employees
export const employees = {
	list: () => request<Employee[]>('/employees'),
	get: (id: string) => request<Employee>(`/employees/${id}`),
	create: (data: Partial<Employee>) => request<Employee>('/employees', { method: 'POST', body: data }),
	update: (id: string, data: Partial<Employee>) => request<Employee>(`/employees/${id}`, { method: 'PUT', body: data }),
	delete: (id: string) => request(`/employees/${id}`, { method: 'DELETE' }),
	getDossier: (id: string) => request<EmployeeDossier>(`/employees/${id}/dossier`),
};

// Projects
export const projects = {
	list: (status?: string) => request<Project[]>(`/projects${status ? `?status=${status}` : ''}`),
	get: (id: string) => request<Project>(`/projects/${id}`),
	create: (data: Partial<Project>) => request<Project>('/projects', { method: 'POST', body: data }),
	update: (id: string, data: Partial<Project>) => request<Project>(`/projects/${id}`, { method: 'PUT', body: data }),
	delete: (id: string) => request(`/projects/${id}`, { method: 'DELETE' }),
};

// Meetings
export const meetings = {
	list: (params?: { employee_id?: string; project_id?: string }) => {
		const query = new URLSearchParams(params as Record<string, string>).toString();
		return request<Meeting[]>(`/meetings${query ? `?${query}` : ''}`);
	},
	get: (id: string) => request<Meeting>(`/meetings/${id}`),
	getCategories: () => request<MeetingCategory[]>('/meeting-categories'),
	process: async (formData: FormData) => {
		const response = await fetch(`${API_URL}/process-meeting`, {
			method: 'POST',
			body: formData
		});
		if (!response.ok) {
			const error = await response.json().catch(() => ({ error: response.statusText }));
			throw new Error(error.error || 'Upload failed');
		}
		return response.json();
	},
};

// Tasks
export const tasks = {
	list: (params?: { assignee_id?: string; project_id?: string; status?: string }) => {
		const query = new URLSearchParams(params as Record<string, string>).toString();
		return request<Task[]>(`/tasks${query ? `?${query}` : ''}`);
	},
	get: (id: string) => request<Task>(`/tasks/${id}`),
	create: (data: Partial<Task>) => request<Task>('/tasks', { method: 'POST', body: data }),
	update: (id: string, data: Partial<Task>) => request<Task>(`/tasks/${id}`, { method: 'PUT', body: data }),
	delete: (id: string) => request(`/tasks/${id}`, { method: 'DELETE' }),
	getKanban: (params?: { assignee_id?: string; project_id?: string }) => {
		const query = new URLSearchParams(params as Record<string, string>).toString();
		return request<KanbanBoard>(`/kanban${query ? `?${query}` : ''}`);
	},
	moveKanban: (taskId: string, newStatus: string) =>
		request(`/kanban/move?task_id=${taskId}&new_status=${newStatus}`, { method: 'PUT' }),
};

// Analytics
export const analytics = {
	getDashboard: () => request<DashboardData>('/analytics/dashboard'),
	getEmployee: (id: string) => request<EmployeeAnalytics>(`/analytics/employee/${id}`),
	getEmployeeByCategory: (id: string) => request(`/analytics/employee/${id}/by-category`),
};

// Calendar (EWS)
export const calendar = {
	get: (employeeId: string, auth: { username: string; password: string }) =>
		request(`/calendar/${employeeId}`, { method: 'POST', body: auth }),
	getSimple: (employeeId: string) => request<CalendarEvent[]>(`/calendar/${employeeId}/simple`),
	findFreeSlots: (data: FreeSlotsRequest) =>
		request('/calendar/free-slots', { method: 'POST', body: data }),
	sync: () => request('/calendar/sync', { method: 'POST' }),
};

// Connector
export const connector = {
	status: () => request<ConnectorStatus>('/connector/status'),
	syncAD: (params?: { mode?: string; include_photos?: boolean }) => {
		const query = new URLSearchParams(params as Record<string, string>).toString();
		return request(`/ad/sync${query ? `?${query}` : ''}`, { method: 'POST' });
	},
	authenticate: async (username: string, password: string) => {
		const formData = new FormData();
		formData.append('username', username);
		formData.append('password', password);
		const response = await fetch(`${API_URL}/ad/authenticate`, {
			method: 'POST',
			body: formData
		});
		return response.json();
	},
};

// Types
export interface Employee {
	id: string;
	name: string;
	email?: string;
	position: string;
	department?: string;
	manager_id?: string;
	photo_base64?: string;
	phone?: string;
	telegram_username?: string;
	created_at?: string;
}

export interface EmployeeDossier {
	employee: Employee;
	one_on_one_count: number;
	project_meetings_count: number;
	tasks: { total: number; done: number; in_progress: number };
	mood_history: { date: string; score: number }[];
	red_flags_history: { date: string; flags: unknown }[];
	recent_meetings: Meeting[];
}

export interface Project {
	id: string;
	name: string;
	description?: string;
	status?: string;
	start_date?: string;
	end_date?: string;
	task_count?: number;
	meetings_count?: number;
	progress?: number;
}

export interface Meeting {
	id: string;
	title?: string;
	employee_id?: string;
	employee_name?: string;
	project_id?: string;
	category_id?: string;
	category?: string;
	date: string;
	start_time?: string;
	end_time?: string;
	duration_minutes?: number;
	location?: string;
	transcript?: string;
	summary?: string;
	mood_score?: number;
	analysis?: MeetingAnalysis;
	employees?: Employee;
	projects?: Project;
	meeting_categories?: MeetingCategory;
	participants?: Employee[];
}

export interface MeetingAnalysis {
	key_topics?: string[];
	action_items?: string[];
	agreements?: string[];
	red_flags?: {
		burnout_signs?: string;
		turnover_risk?: string;
	};
}

export interface MeetingCategory {
	id: string;
	code: string;
	name: string;
}

export interface Task {
	id: string;
	title: string;
	description?: string;
	status: string;
	priority?: string;
	flag_color?: string;
	assignee_id?: string;
	assignee_name?: string;
	project_id?: string;
	due_date?: string;
	is_epic?: boolean;
	assignee?: Employee;
	project?: Project;
	tags?: { name: string; color: string }[];
}

export interface KanbanBoard {
	backlog: Task[];
	todo: Task[];
	in_progress: Task[];
	review: Task[];
	done: Task[];
}

export interface DashboardData {
	total_employees: number;
	meetings_this_month: number;
	average_mood: number;
	tasks_completed: number;
	tasks_todo: number;
	tasks_in_progress: number;
	mood_trend: { date: string; score: number }[];
	employees_needing_attention: { id: string; name: string; reason: string; days_since_meeting: number }[];
	meetings_by_category: Record<string, number>;
	agreements_total: number;
	agreements_completed: number;
	agreements_overdue: number;
	recent_meetings: Meeting[];
	employees?: Employee[];
	projects?: Project[];
}

export interface EmployeeAnalytics {
	mood_history: { date: string; score: number }[];
	red_flags_history: { date: string; flags: unknown }[];
	task_stats: { total: number; done: number; in_progress: number };
	agreement_stats: { total: number; completed: number; pending: number; overdue: number };
	total_meetings: number;
}

export interface ConnectorStatus {
	connected: boolean;
	ad_status?: string;
	ad_sync_enabled?: boolean;
	last_sync?: string;
	employee_count?: number;
	calendar_integration?: string;
	ews_url?: string;
	ews_configured?: boolean;
}

export interface CalendarEvent {
	id: string;
	subject: string;
	start: string;
	end: string;
	location?: string;
	organizer?: string;
}

export interface FreeSlotsRequest {
	attendee_ids: string[];
	username: string;
	password: string;
	start_date: string;
	end_date: string;
	duration_minutes?: number;
}

export interface CalendarSyncRequest {
	employee_id: string;
	username: string;
	password: string;
	days_back?: number;
	days_forward?: number;
}
