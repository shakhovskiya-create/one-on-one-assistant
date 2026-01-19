import { browser } from '$app/environment';

// All browser requests go through SvelteKit proxy (/api/v1/...)
// Server-side requests go directly to backend
const BASE_URL = browser ? '/api/v1' : 'http://backend:8080/api/v1';
const API_URL = BASE_URL;

interface RequestOptions {
	method?: string;
	body?: unknown;
	headers?: Record<string, string>;
}

function getAuthHeaders(): Record<string, string> {
	const headers: Record<string, string> = {};
	if (browser) {
		const token = localStorage.getItem('auth_token');
		if (token && token !== 'authenticated') {
			headers['Authorization'] = `Bearer ${token}`;
		}
	}
	return headers;
}

async function request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
	const { method = 'GET', body, headers = {} } = options;

	const config: RequestInit = {
		method,
		headers: {
			'Content-Type': 'application/json',
			...getAuthHeaders(),
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
	create: (data: Partial<Meeting>) => request<Meeting>('/meetings', { method: 'POST', body: data }),
	getCategories: () => request<MeetingCategory[]>('/meeting-categories'),
	process: async (formData: FormData) => {
		const response = await fetch(`${API_URL}/process-meeting`, {
			method: 'POST',
			headers: getAuthHeaders(),
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
	addComment: (id: string, data: { author_id: string; content: string }) =>
		request(`/tasks/${id}/comments`, { method: 'POST', body: data }),
	getKanban: (params?: { assignee_id?: string; project_id?: string }) => {
		const query = new URLSearchParams(params as Record<string, string>).toString();
		return request<KanbanBoard>(`/kanban${query ? `?${query}` : ''}`);
	},
	moveKanban: (taskId: string, newStatus: string) =>
		request(`/kanban/move?task_id=${taskId}&new_status=${newStatus}`, { method: 'PUT' }),
};

// Analytics
export const analytics = {
	getDashboard: (period?: string, managerId?: string) => {
		const params = new URLSearchParams();
		if (period) params.append('period', period);
		if (managerId) params.append('manager_id', managerId);
		const query = params.toString();
		return request<DashboardData>(`/analytics/dashboard${query ? `?${query}` : ''}`);
	},
	getEmployee: (id: string, period?: string) =>
		request<EmployeeAnalytics>(`/analytics/employee/${id}${period ? `?period=${period}` : ''}`),
	getEmployeeByCategory: (id: string) => request(`/analytics/employee/${id}/by-category`),
	getTeamStats: (managerId: string) => request<TeamMemberStats[]>(`/analytics/team/${managerId}`),
};

// Calendar (EWS)
export const calendar = {
	get: (employeeId: string) => request(`/calendar/${employeeId}`),
	getSimple: (employeeId: string) => request<CalendarEvent[]>(`/calendar/${employeeId}/simple`),
	findFreeSlots: (data: FreeSlotsRequest) =>
		request('/calendar/free-slots', { method: 'POST', body: data }),
	sync: (data: CalendarSyncRequest) =>
		request('/calendar/sync', { method: 'POST', body: data }),
};

// Messenger
export const messenger = {
	listConversations: (userId: string) => request<Conversation[]>(`/conversations?user_id=${userId}`),
	getConversation: (id: string, userId: string, limit?: number, offset?: number) => {
		const params = new URLSearchParams();
		params.append('user_id', userId); // Required for access control
		if (limit) params.append('limit', limit.toString());
		if (offset) params.append('offset', offset.toString());
		return request<{ conversation: Conversation; participants: Employee[]; messages: Message[] }>(`/conversations/${id}?${params}`);
	},
	createConversation: (data: { type?: string; name?: string; participants: string[] }) =>
		request<Conversation>('/conversations', { method: 'POST', body: data }),
	sendMessage: (data: { conversation_id: string; sender_id: string; content: string; reply_to_id?: string }) =>
		request<Message>('/messages', { method: 'POST', body: data }),
	getWebSocketUrl: (userId: string) => {
		// WebSocket connects directly to the backend, not through the API proxy
		const wsProtocol = browser && window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const wsHost = browser ? window.location.host : 'backend:8080';
		const token = browser ? localStorage.getItem('auth_token') : null;
		const params = new URLSearchParams({ user_id: userId });
		if (token && token !== 'authenticated') {
			params.append('token', token);
		}
		return `${wsProtocol}//${wsHost}/ws/messenger?${params.toString()}`;
	},
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

// Files
export const files = {
	upload: async (file: File, options?: { entityType?: string; entityId?: string; uploadedBy?: string }) => {
		const formData = new FormData();
		formData.append('file', file);
		if (options?.entityType) formData.append('entity_type', options.entityType);
		if (options?.entityId) formData.append('entity_id', options.entityId);
		if (options?.uploadedBy) formData.append('uploaded_by', options.uploadedBy);

		const response = await fetch(`${API_URL}/files`, {
			method: 'POST',
			headers: getAuthHeaders(),
			body: formData
		});
		if (!response.ok) {
			const error = await response.json().catch(() => ({ error: response.statusText }));
			throw new Error(error.error || 'Upload failed');
		}
		return response.json() as Promise<FileUploadResult>;
	},
	list: (params?: { entityType?: string; entityId?: string; uploadedBy?: string }) => {
		const query = new URLSearchParams();
		if (params?.entityType) query.append('entity_type', params.entityType);
		if (params?.entityId) query.append('entity_id', params.entityId);
		if (params?.uploadedBy) query.append('uploaded_by', params.uploadedBy);
		return request<FileMetadata[]>(`/files${query.toString() ? `?${query}` : ''}`);
	},
	get: (id: string) => request<FileMetadata>(`/files/${id}`),
	getUrl: (id: string) => request<{ id: string; name: string; url: string; content_type: string }>(`/files/${id}/url`),
	delete: (id: string) => request(`/files/${id}`, { method: 'DELETE' }),
	attach: (fileId: string, entityType: string, entityId: string) =>
		request('/files/attach', { method: 'POST', body: { file_id: fileId, entity_type: entityType, entity_id: entityId } }),
};

// BPMN / Camunda
export const bpmn = {
	status: () => request<BPMNStatus>('/bpmn/status'),
	getDefinitions: () => request<ProcessDefinition[]>('/bpmn/definitions'),
	getDefinition: (key: string) => request<ProcessDefinition>(`/bpmn/definitions/${key}`),
	startProcess: (processKey: string, businessKey?: string, variables?: Record<string, unknown>) =>
		request<ProcessInstance>('/bpmn/processes', {
			method: 'POST',
			body: { process_key: processKey, business_key: businessKey, variables }
		}),
	getProcesses: (processKey?: string, active?: boolean) => {
		const params = new URLSearchParams();
		if (processKey) params.append('process_key', processKey);
		if (active !== undefined) params.append('active', String(active));
		return request<ProcessInstance[]>(`/bpmn/processes${params.toString() ? `?${params}` : ''}`);
	},
	getProcess: (id: string) => request<{ instance: ProcessInstance; variables: Record<string, unknown> }>(`/bpmn/processes/${id}`),
	cancelProcess: (id: string) => request(`/bpmn/processes/${id}`, { method: 'DELETE' }),
	getTasks: (assignee?: string, processInstanceId?: string) => {
		const params = new URLSearchParams();
		if (assignee) params.append('assignee', assignee);
		if (processInstanceId) params.append('process_instance_id', processInstanceId);
		return request<BPMNTask[]>(`/bpmn/tasks${params.toString() ? `?${params}` : ''}`);
	},
	getTask: (id: string) => request<BPMNTask>(`/bpmn/tasks/${id}`),
	completeTask: (id: string, variables?: Record<string, unknown>) =>
		request(`/bpmn/tasks/${id}/complete`, { method: 'POST', body: { variables } }),
	claimTask: (id: string, userId: string) =>
		request(`/bpmn/tasks/${id}/claim`, { method: 'POST', body: { user_id: userId } }),
	unclaimTask: (id: string) =>
		request(`/bpmn/tasks/${id}/unclaim`, { method: 'POST' }),
};

// Types
export interface BPMNStatus {
	configured: boolean;
	status: string;
	message?: string;
	url?: string;
	error?: string;
}

export interface ProcessDefinition {
	id: string;
	key: string;
	name: string;
	description?: string;
	version: number;
	suspended: boolean;
}

export interface ProcessInstance {
	id: string;
	definitionId: string;
	businessKey?: string;
	suspended: boolean;
	ended?: boolean;
}

export interface BPMNTask {
	id: string;
	name: string;
	assignee?: string;
	created: string;
	due?: string;
	description?: string;
	priority: number;
	processDefinitionId?: string;
	processInstanceId?: string;
	taskDefinitionKey?: string;
	formKey?: string;
}

export interface FileUploadResult {
	id: string;
	name: string;
	url: string;
	content_type: string;
	size_bytes: number;
	storage_path: string;
}

export interface FileMetadata {
	id: string;
	name: string;
	original_name: string;
	storage_path: string;
	bucket: string;
	content_type: string;
	size_bytes: number;
	uploaded_by?: string;
	entity_type?: string;
	entity_id?: string;
	url?: string;
	created_at: string;
}

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
	action_items?: Array<string | { task?: string; improvement?: string; responsible?: string; deadline?: string | null }>;
	agreements?: Array<string | { task?: string; responsible?: string; deadline?: string | null }>;
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
	priority?: number;
	flag_color?: string;
	assignee_id?: string;
	assignee_name?: string;
	project_id?: string;
	parent_id?: string;
	start_date?: string;
	due_date?: string;
	progress?: number;
	is_epic?: boolean;
	assignee?: Employee;
	project?: Project;
	tags?: { name: string; color: string }[];
	dependencies?: TaskDependency[];
}

export interface TaskDependency {
	id: string;
	task_id: string;
	depends_on_task_id: string;
	dependency_type: string;
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

export interface TeamMemberStats {
	id: string;
	name: string;
	position: string;
	photo_base64?: string;
	subordinates: number;
	open_tasks: number;
	overdue_tasks: number;
	meetings: number;
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
	yandex_configured?: boolean;
	openai_configured?: boolean;
	anthropic_configured?: boolean;
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
	days_back?: number;
	days_forward?: number;
}

// Messenger types
export interface Conversation {
	id: string;
	type: 'direct' | 'group';
	name?: string;
	created_at?: string;
	updated_at?: string;
	participants?: Employee[];
	last_message?: Message;
	unread_count?: number;
}

export interface Message {
	id: string;
	conversation_id: string;
	sender_id: string;
	content: string;
	message_type: 'text' | 'file' | 'system';
	reply_to_id?: string;
	edited_at?: string;
	created_at?: string;
	sender?: Employee;
	reply_to?: Message;
}

// Mail types
export interface MailFolder {
	id: string;
	display_name: string;
	unread_count: number;
	total_count: number;
}

export interface EmailPerson {
	name: string;
	email: string;
}

export interface EmailMessage {
	id: string;
	subject: string;
	from?: EmailPerson;
	to?: EmailPerson[];
	cc?: EmailPerson[];
	body: string;
	body_preview?: string;
	received_at: string;
	is_read: boolean;
	has_attachments: boolean;
	folder_id?: string;
}

// Mail API
export const mail = {
	getFolders: (username: string, password: string) =>
		request<MailFolder[]>(`/mail/folders?username=${encodeURIComponent(username)}&password=${encodeURIComponent(password)}`),
	getEmails: (username: string, password: string, folderId?: string, limit?: number) => {
		const params = new URLSearchParams({ username, password });
		if (folderId) params.append('folder_id', folderId);
		if (limit) params.append('limit', limit.toString());
		return request<EmailMessage[]>(`/mail/emails?${params}`);
	},
	getEmailBody: (data: { username: string; password: string; item_id: string; change_key?: string }) =>
		request<{ body: string }>('/mail/body', { method: 'POST', body: data }),
	sendEmail: (data: { username: string; password: string; to: string[]; cc?: string[]; subject: string; body: string }) =>
		request('/mail/send', { method: 'POST', body: data }),
	markAsRead: (data: { username: string; password: string; item_id: string; change_key?: string }) =>
		request('/mail/mark-read', { method: 'POST', body: data }),
	deleteEmail: (data: { username: string; password: string; item_id: string; change_key?: string }) =>
		request('/mail/email', { method: 'DELETE', body: data }),
};

// Combined API object for convenience
export const api = {
	employees,
	projects,
	meetings,
	tasks,
	analytics,
	calendar,
	messenger,
	connector,
	files,
	bpmn,
	mail
};
