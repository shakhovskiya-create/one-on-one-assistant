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
			// Include all error details in the thrown error
			const err = new Error(error.error || 'Upload failed') as any;
			err.details = error.details;
			err.hint = error.hint;
			err.errors = error.errors;
			throw err;
		}
		return response.json();
	},
};

// Tasks
// Task dependency type
export interface TaskDependency {
	id: string;
	task_id: string;
	depends_on_task_id: string;
	dependency_type: string;
	created_at: string;
	depends_on_task?: Task;
}

// Workflow types
export interface StatusColumn {
	id: string;
	label: string;
	color: string;
	wipLimit: number;
}

export interface WorkflowMode {
	id: string;
	name: string;
	description: string;
	statuses: StatusColumn[];
	is_default: boolean;
}

// Workflows API
export const workflows = {
	getMyWorkflow: () =>
		request<{ workflow: WorkflowMode; department: string }>('/workflows/me'),
	list: () =>
		request<WorkflowMode[]>('/workflows'),
	listDepartments: () =>
		request<{ id: string; department: string; workflow_mode_id: string; workflow_mode: WorkflowMode }[]>('/workflows/departments'),
	setDepartmentWorkflow: (department: string, workflowModeId: string) =>
		request('/workflows/departments', { method: 'POST', body: { department, workflow_mode_id: workflowModeId } }),
};

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
	// Dependencies
	getDependencies: (id: string) =>
		request<{ dependencies: TaskDependency[]; dependents: TaskDependency[] }>(`/tasks/${id}/dependencies`),
	addDependency: (id: string, dependsOnTaskId: string, type?: string) =>
		request<TaskDependency>(`/tasks/${id}/dependencies`, {
			method: 'POST',
			body: { depends_on_task_id: dependsOnTaskId, dependency_type: type || 'finish_to_start' }
		}),
	removeDependency: (id: string, depId: string) =>
		request(`/tasks/${id}/dependencies/${depId}`, { method: 'DELETE' }),
	isBlocked: (id: string) =>
		request<{ blocked: boolean; blockers: Task[] }>(`/tasks/${id}/blocked`),
	// Time entries
	getTimeEntries: (id: string) =>
		request<TimeEntry[]>(`/tasks/${id}/time-entries`),
	addTimeEntry: (id: string, data: { hours: number; description?: string; date?: string }) =>
		request<TimeEntry>(`/tasks/${id}/time-entries`, { method: 'POST', body: data }),
	updateTimeEntry: (taskId: string, entryId: string, data: { hours?: number; description?: string; date?: string }) =>
		request(`/tasks/${taskId}/time-entries/${entryId}`, { method: 'PUT', body: data }),
	deleteTimeEntry: (taskId: string, entryId: string) =>
		request(`/tasks/${taskId}/time-entries/${entryId}`, { method: 'DELETE' }),
	getResourceSummary: (id: string) =>
		request<ResourceSummary>(`/tasks/${id}/resources`),
};

// Time entries for current user
export const timeEntries = {
	getMy: (params?: { start_date?: string; end_date?: string }) => {
		const query = params ? '?' + new URLSearchParams(params).toString() : '';
		return request<{ entries: TimeEntry[]; total_hours: number }>(`/time-entries/me${query}`);
	},
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
export interface CreateMeetingRequest {
	subject: string;
	body?: string;
	start: string; // ISO 8601 format
	end: string; // ISO 8601 format
	location?: string;
	required_attendees?: string[]; // Employee IDs or emails
	optional_attendees?: string[]; // Employee IDs or emails
	is_online_meeting?: boolean;
}

export interface UpdateMeetingRequest {
	item_id: string;
	change_key?: string;
	subject?: string;
	start?: string;
	end?: string;
	location?: string;
}

export interface DeleteMeetingRequest {
	item_id: string;
	change_key?: string;
	send_cancellations?: boolean;
}

export const calendar = {
	get: (employeeId: string) => request(`/calendar/${employeeId}`),
	getSimple: (employeeId: string) => request<CalendarEvent[]>(`/calendar/${employeeId}/simple`),
	findFreeSlots: (data: FreeSlotsRequest) =>
		request('/calendar/free-slots', { method: 'POST', body: data }),
	sync: (data: CalendarSyncRequest) =>
		request('/calendar/sync', { method: 'POST', body: data }),
	getRooms: (employeeId: string) =>
		request<{ rooms: MeetingRoom[] }>(`/calendar/rooms?employee_id=${employeeId}`),
	createMeeting: (data: CreateMeetingRequest) =>
		request<{ success: boolean; exchange_id: string; message: string }>('/calendar/create', { method: 'POST', body: data }),
	updateMeeting: (data: UpdateMeetingRequest) =>
		request<{ success: boolean; message: string }>('/calendar/update', { method: 'PUT', body: data }),
	deleteMeeting: (data: DeleteMeetingRequest) =>
		request<{ success: boolean; message: string }>('/calendar/delete', { method: 'DELETE', body: data }),
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
	createConversation: (data: { type?: string; name?: string; description?: string; participants: string[] }) =>
		request<Conversation>('/conversations', { method: 'POST', body: data }),
	sendMessage: (data: {
		conversation_id: string;
		sender_id: string;
		content: string;
		message_type?: 'text' | 'voice' | 'video' | 'file' | 'sticker' | 'gif';
		file_id?: string;
		duration_seconds?: number;
		reply_to_id?: string
	}) => request<Message>('/messages', { method: 'POST', body: data }),
	updateMessage: (messageId: string, content: string) =>
		request<Message>(`/messages/${messageId}`, { method: 'PUT', body: { content } }),
	deleteMessage: (messageId: string) =>
		request<{ success: boolean }>(`/messages/${messageId}`, { method: 'DELETE' }),
	addReaction: (messageId: string, emoji: string) =>
		request<{ success: boolean; reactions: { emoji: string; users: string[] }[] }>(`/messages/${messageId}/reactions`, { method: 'POST', body: { emoji } }),
	getReactions: (messageId: string) =>
		request<{ reactions: { emoji: string; users: string[] }[] }>(`/messages/${messageId}/reactions`),
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
	// Telegram integration
	getTelegramConfig: (channelId: string) =>
		request<{ enabled: boolean; chat_id?: number; webhook_url: string }>(`/channels/${channelId}/telegram`),
	configureTelegram: (channelId: string, data: { bot_token?: string; chat_id?: number; enabled: boolean }) =>
		request<{ success: boolean; webhook_url: string; message: string }>(`/channels/${channelId}/telegram`, { method: 'POST', body: data }),
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
	uploadBlob: async (blob: Blob, filename: string, options?: { entityType?: string; entityId?: string; uploadedBy?: string }) => {
		const formData = new FormData();
		formData.append('file', blob, filename);
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

// Speech-to-Text
export interface TranscribeResult {
	transcript: string;
	whisper?: string;
	yandex?: string;
}

export const speech = {
	transcribe: async (audioBlob: Blob, service: 'auto' | 'whisper' | 'yandex' | 'both' = 'auto'): Promise<TranscribeResult> => {
		const formData = new FormData();
		// Determine file extension based on blob type
		const ext = audioBlob.type.includes('webm') ? '.webm' : audioBlob.type.includes('mp3') ? '.mp3' : '.ogg';
		formData.append('audio', audioBlob, `voice${ext}`);

		const response = await fetch(`${API_URL}/speech/transcribe?service=${service}`, {
			method: 'POST',
			headers: getAuthHeaders(),
			body: formData
		});

		if (!response.ok) {
			const error = await response.json().catch(() => ({ error: response.statusText }));
			throw new Error(error.error || error.details || 'Transcription failed');
		}

		return response.json();
	}
};

// Confluence
export interface ConfluenceSpace {
	id: number;
	key: string;
	name: string;
	type: string;
	description?: { plain?: { value: string } };
	_links: { webui: string };
}

export interface ConfluenceContent {
	id: string;
	type: string;
	status: string;
	title: string;
	space?: ConfluenceSpace;
	version?: { number: number };
	body?: {
		storage?: { value: string };
		view?: { value: string };
	};
	_links: { webui: string; tinyui?: string };
	ancestors?: ConfluenceContent[];
}

export interface ConfluenceSearchResult {
	content: ConfluenceContent;
	title: string;
	excerpt: string;
	url: string;
	lastModified: string;
	friendlyLastModified: string;
}

export const confluence = {
	status: () => request<{ configured: boolean; url: string }>('/confluence/status'),
	getSpaces: (limit?: number) =>
		request<{ spaces: ConfluenceSpace[] }>(`/confluence/spaces${limit ? `?limit=${limit}` : ''}`),
	getSpace: (key: string) => request<ConfluenceSpace>(`/confluence/spaces/${key}`),
	getSpaceContent: (spaceKey: string, type?: string, limit?: number) => {
		const params = new URLSearchParams();
		if (type) params.append('type', type);
		if (limit) params.append('limit', String(limit));
		return request<{ pages: ConfluenceContent[] }>(`/confluence/spaces/${spaceKey}/content${params.toString() ? `?${params}` : ''}`);
	},
	getPage: (id: string, expandBody?: boolean) =>
		request<ConfluenceContent>(`/confluence/pages/${id}${expandBody ? '?expand_body=true' : ''}`),
	getChildPages: (id: string, limit?: number) =>
		request<{ pages: ConfluenceContent[] }>(`/confluence/pages/${id}/children${limit ? `?limit=${limit}` : ''}`),
	search: (query: string, spaceKey?: string, limit?: number) => {
		const params = new URLSearchParams();
		params.append('q', query);
		if (spaceKey) params.append('space', spaceKey);
		if (limit) params.append('limit', String(limit));
		return request<{ results: ConfluenceSearchResult[]; totalSize: number }>(`/confluence/search?${params}`);
	},
	getRecent: (limit?: number) =>
		request<{ pages: ConfluenceContent[] }>(`/confluence/recent${limit ? `?limit=${limit}` : ''}`),
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
	hourly_rate?: number;
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
	// Resource planning fields
	estimated_hours?: number;
	actual_hours?: number;
	estimated_cost?: number;
	actual_cost?: number;
	// Relations
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

export interface TimeEntry {
	id: string;
	task_id: string;
	employee_id: string;
	hours: number;
	description?: string;
	date: string;
	created_at?: string;
	employee?: Employee;
}

export interface ResourceSummary {
	estimated_hours?: number;
	actual_hours?: number;
	estimated_cost?: number;
	actual_cost?: number;
	logged_hours: number;
	hourly_rate: number;
	calculated_cost: number;
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
	db_status?: string;
	employee_count?: number;
	meetings_count?: number;
	tasks_count?: number;
	open_tasks_count?: number;
	overdue_tasks_count?: number;
	projects_count?: number;
	calendar_integration?: string;
	ews_url?: string;
	ews_configured?: boolean;
	yandex_configured?: boolean;
	openai_configured?: boolean;
	anthropic_configured?: boolean;
}

export interface CalendarPerson {
	name: string;
	email: string;
}

export interface CalendarAttendee {
	name: string;
	email: string;
	response?: string; // "Accept", "Decline", "Tentative", "Unknown"
	optional?: boolean; // Required (false) vs Optional (true) attendee
}

export interface CalendarEvent {
	id: string;
	subject: string;
	title?: string;
	start: string;
	start_time?: string;
	end: string;
	end_time?: string;
	date?: string;
	location?: string;
	organizer?: CalendarPerson;
	attendees?: CalendarAttendee[];
	is_recurring?: boolean;
	is_cancelled?: boolean;
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

export interface MeetingRoom {
	name: string;
	email: string;
	capacity?: number;
}

// Messenger types
export interface Conversation {
	id: string;
	type: 'direct' | 'group' | 'channel';
	name?: string;
	description?: string;
	created_by?: string;
	telegram_enabled?: boolean;
	telegram_chat_id?: number;
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
	message_type: 'text' | 'file' | 'voice' | 'video' | 'sticker' | 'gif' | 'system';
	file_id?: string;
	file_url?: string;
	duration_seconds?: number;
	thumbnail_url?: string;
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
	change_key?: string;
	conversation_id?: string;
	item_class?: string;
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

export interface EmailAttachment {
	id: string;
	name: string;
	content_type: string;
	size: number;
	is_inline: boolean;
	content_id?: string;
}


// GIF types
export interface GifResult {
	id: string;
	title: string;
	url: string;
	preview_url: string;
	width: string;
	height: string;
}

// Mail API - credentials sent in POST body for security
export const mail = {
	getFolders: (username: string, password: string) =>
		request<MailFolder[]>('/mail/folders', { method: 'POST', body: { username, password } }),
	getEmails: (username: string, password: string, folderId?: string, limit?: number) =>
		request<EmailMessage[]>('/mail/emails', {
			method: 'POST',
			body: { username, password, folder_id: folderId, limit: limit || 50 }
		}),
	getEmailBody: (data: { username: string; password: string; item_id: string; change_key?: string }) =>
		request<{ body: string }>('/mail/body', { method: 'POST', body: data }),
	sendEmail: (data: { username: string; password: string; to: string[]; cc?: string[]; subject: string; body: string; attachments?: { name: string; content: string }[] }) =>
		request('/mail/send', { method: 'POST', body: data }),
	markAsRead: (data: { username: string; password: string; item_id: string; change_key?: string }) =>
		request('/mail/mark-read', { method: 'POST', body: data }),
	deleteEmail: (data: { username: string; password: string; item_id: string; change_key?: string }) =>
		request('/mail/email', { method: 'DELETE', body: data }),
	getAttachments: (data: { username: string; password: string; item_id: string; change_key?: string }) =>
		request<{ attachments: EmailAttachment[] }>('/mail/attachments', { method: 'POST', body: data }),
	getAttachmentContent: (data: { username: string; password: string; attachment_id: string }) =>
		request<{ name: string; content_type: string; content: string }>('/mail/attachment/content', { method: 'POST', body: data }),
	respondToMeeting: (data: { username: string; password: string; item_id: string; change_key?: string; response: 'Accept' | 'Decline' | 'Tentative' }) =>
		request<{ success: boolean; response: string }>('/mail/meeting/respond', { method: 'POST', body: data }),
};

// GIF API (GIPHY)
export const giphy = {
	search: async (query: string, limit: number = 20, offset: number = 0): Promise<GifResult[]> => {
		const params = new URLSearchParams({ q: query, limit: limit.toString(), offset: offset.toString() });
		const response = await fetch(`${API_URL}/gifs/search?${params}`);
		const data = await response.json();
		return data.gifs || [];
	},
	trending: async (limit: number = 20, offset: number = 0): Promise<GifResult[]> => {
		const params = new URLSearchParams({ limit: limit.toString(), offset: offset.toString() });
		const response = await fetch(`${API_URL}/gifs/trending?${params}`);
		const data = await response.json();
		return data.gifs || [];
	}
};

// GitHub types
export interface GitHubRepository {
	id: number;
	name: string;
	full_name: string;
	description: string;
	html_url: string;
	clone_url: string;
	default_branch: string;
	private: boolean;
	fork: boolean;
	created_at: string;
	updated_at: string;
	pushed_at: string;
	language: string;
	stargazers_count: number;
	forks_count: number;
}

export interface GitHubCommitAuthor {
	name: string;
	email: string;
	date: string;
}

export interface GitHubCommit {
	sha: string;
	html_url: string;
	commit: {
		message: string;
		author: GitHubCommitAuthor;
		committer: GitHubCommitAuthor;
	};
	author?: {
		login: string;
		avatar_url: string;
		html_url: string;
	};
	committer?: {
		login: string;
		avatar_url: string;
		html_url: string;
	};
	stats?: {
		additions: number;
		deletions: number;
		total: number;
	};
}

export interface GitHubBranch {
	name: string;
	commit: {
		sha: string;
		url: string;
	};
	protected: boolean;
}

export interface GitHubPullRequest {
	id: number;
	number: number;
	state: string;
	title: string;
	body: string;
	html_url: string;
	created_at: string;
	updated_at: string;
	closed_at?: string;
	merged_at?: string;
	user?: {
		login: string;
		avatar_url: string;
		html_url: string;
	};
	head: {
		ref: string;
		sha: string;
	};
	base: {
		ref: string;
		sha: string;
	};
	mergeable?: boolean;
}

// GitHub API
export const github = {
	status: () => request<{ configured: boolean }>('/github/status'),
	parseUrl: (url: string) =>
		request<{ owner: string; repo: string }>('/github/parse-url', {
			method: 'POST',
			body: { url }
		}),
	getRepository: (owner: string, repo: string) =>
		request<GitHubRepository>(`/github/repos/${owner}/${repo}`),
	getCommits: (owner: string, repo: string, branch?: string, limit?: number) => {
		const params = new URLSearchParams();
		if (branch) params.append('branch', branch);
		if (limit) params.append('limit', String(limit));
		return request<GitHubCommit[]>(`/github/repos/${owner}/${repo}/commits${params.toString() ? `?${params}` : ''}`);
	},
	getBranches: (owner: string, repo: string, limit?: number) =>
		request<GitHubBranch[]>(`/github/repos/${owner}/${repo}/branches${limit ? `?limit=${limit}` : ''}`),
	getPullRequests: (owner: string, repo: string, state?: string, limit?: number) => {
		const params = new URLSearchParams();
		if (state) params.append('state', state);
		if (limit) params.append('limit', String(limit));
		return request<GitHubPullRequest[]>(`/github/repos/${owner}/${repo}/pulls${params.toString() ? `?${params}` : ''}`);
	},
	getTaskCommits: (taskId: string, owner: string, repo: string, limit?: number) => {
		const params = new URLSearchParams({ owner, repo });
		if (limit) params.append('limit', String(limit));
		return request<GitHubCommit[]>(`/github/tasks/${taskId}/commits?${params}`);
	},
	getTaskPullRequests: (taskId: string, owner: string, repo: string, limit?: number) => {
		const params = new URLSearchParams({ owner, repo });
		if (limit) params.append('limit', String(limit));
		return request<GitHubPullRequest[]>(`/github/tasks/${taskId}/pulls?${params}`);
	}
};

// Admin types
export interface AdminStats {
	total_users: number;
	active_users: number;
	total_tasks: number;
	completed_tasks: number;
	total_meetings: number;
	total_messages: number;
	admin_count: number;
	departments_count: number;
}

export interface AdminUser {
	id: string;
	name: string;
	email: string;
	department?: string;
	position: string;
	role?: string;
	created_at?: string;
}

export interface AuditLog {
	id: string;
	user_id?: string;
	action: string;
	entity_type?: string;
	entity_id?: string;
	old_value?: Record<string, unknown>;
	new_value?: Record<string, unknown>;
	ip_address?: string;
	user_agent?: string;
	created_at: string;
	user?: { id: string; name: string };
}

export interface SystemSetting {
	id: string;
	key: string;
	value: unknown;
	description?: string;
	updated_by?: string;
	updated_at: string;
}

export interface DepartmentInfo {
	name: string;
	employee_count: number;
	workflow_mode_id?: string;
}

// Versions/Releases (JIRA-like)
export interface Version {
	id: string;
	project_id?: string;
	name: string;
	description?: string;
	status: 'unreleased' | 'released' | 'archived';
	start_date?: string;
	release_date?: string;
	released_at?: string;
	created_by?: string;
	created_at?: string;
	updated_at?: string;
	project?: Project;
	tasks_count?: number;
	tasks_done?: number;
	progress?: number;
}

export interface VersionWithTasks {
	version: Version;
	tasks: Task[];
}

export interface ReleaseNotes {
	version: Version;
	features: Task[];
	fixes: Task[];
	other: Task[];
	total: number;
}

// Versions API
export const versions = {
	list: (params?: { project_id?: string; status?: string }) => {
		if (!params) return request<Version[]>('/versions');
		// Filter out undefined values to prevent "undefined" string in URL
		const filtered: Record<string, string> = {};
		if (params.project_id) filtered.project_id = params.project_id;
		if (params.status) filtered.status = params.status;
		const query = Object.keys(filtered).length > 0 ? '?' + new URLSearchParams(filtered).toString() : '';
		return request<Version[]>(`/versions${query}`);
	},
	get: (id: string) => request<VersionWithTasks>(`/versions/${id}`),
	create: (data: Partial<Version>) => request<Version>('/versions', { method: 'POST', body: data }),
	update: (id: string, data: Partial<Version>) => request<Version>(`/versions/${id}`, { method: 'PUT', body: data }),
	delete: (id: string) => request(`/versions/${id}`, { method: 'DELETE' }),
	release: (id: string) => request<Version>(`/versions/${id}/release`, { method: 'POST' }),
	getReleaseNotes: (id: string) => request<ReleaseNotes>(`/versions/${id}/release-notes`),
};

// Sprint type
export interface Sprint {
	id: string;
	project_id?: string;
	name: string;
	goal?: string;
	start_date: string;
	end_date: string;
	status: 'planning' | 'active' | 'completed';
	velocity?: number;
	created_by?: string;
	created_at?: string;
	updated_at?: string;
	project?: Project;
	tasks_count?: number;
	tasks_done?: number;
	total_points?: number;
	completed_points?: number;
	progress?: number;
}

export interface SprintWithTasks {
	sprint: Sprint;
	tasks: Task[];
}

// Sprints API
export const sprints = {
	list: (params?: { project_id?: string; status?: string }) => {
		if (!params) return request<Sprint[]>('/sprints');
		const filtered: Record<string, string> = {};
		if (params.project_id) filtered.project_id = params.project_id;
		if (params.status) filtered.status = params.status;
		const query = Object.keys(filtered).length > 0 ? '?' + new URLSearchParams(filtered).toString() : '';
		return request<Sprint[]>(`/sprints${query}`);
	},
	get: (id: string) => request<SprintWithTasks>(`/sprints/${id}`),
	getActive: (projectId?: string) => {
		const query = projectId ? `?project_id=${projectId}` : '';
		return request<Sprint | null>(`/sprints/active${query}`);
	},
	create: (data: Partial<Sprint>) => request<Sprint>('/sprints', { method: 'POST', body: data }),
	update: (id: string, data: Partial<Sprint>) => request<Sprint>(`/sprints/${id}`, { method: 'PUT', body: data }),
	delete: (id: string) => request(`/sprints/${id}`, { method: 'DELETE' }),
	start: (id: string) => request<Sprint>(`/sprints/${id}/start`, { method: 'POST' }),
	complete: (id: string) => request<Sprint>(`/sprints/${id}/complete`, { method: 'POST' }),
};

// Admin API (requires admin role)
const ADMIN_URL = browser ? '/api/admin' : 'http://backend:8080/api/admin';

async function adminRequest<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
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

	const response = await fetch(`${ADMIN_URL}${endpoint}`, config);

	if (!response.ok) {
		const error = await response.json().catch(() => ({ error: response.statusText }));
		throw new Error(error.error || error.message || 'Request failed');
	}

	return response.json();
}

export const admin = {
	getStats: () => adminRequest<AdminStats>('/stats'),
	listUsers: () => adminRequest<AdminUser[]>('/users'),
	updateUserRole: (userId: string, role: string) =>
		adminRequest(`/users/${userId}/role`, { method: 'PUT', body: { role } }),
	getSettings: () => adminRequest<SystemSetting[]>('/settings'),
	updateSetting: (key: string, value: unknown) =>
		adminRequest('/settings', { method: 'PUT', body: { key, value } }),
	getAuditLogs: (params?: { limit?: number; offset?: number; action?: string; entity_type?: string }) => {
		const query = params ? '?' + new URLSearchParams(params as Record<string, string>).toString() : '';
		return adminRequest<AuditLog[]>(`/audit-logs${query}`);
	},
	getDepartments: () => adminRequest<DepartmentInfo[]>('/departments'),
};

// Auth - get current user role
export const auth = {
	getRole: () => request<{ role: string }>('/auth/role'),
	getMe: () => request<{ id: string; name: string; email: string; department?: string }>('/auth/me'),
	refresh: () => request<{ token: string }>('/auth/refresh', { method: 'POST' }),
};

// Service Desk types
export interface ServiceTicket {
	id: string;
	number: string;
	type: 'incident' | 'service_request' | 'change' | 'problem';
	title: string;
	description?: string;
	category_id?: string;
	priority: 'low' | 'medium' | 'high' | 'critical';
	impact?: 'individual' | 'department' | 'organization';
	status: 'new' | 'in_progress' | 'pending' | 'resolved' | 'closed';
	requester_id: string;
	assignee_id?: string;
	sla_deadline?: string;
	resolution?: string;
	resolved_at?: string;
	closed_at?: string;
	created_at?: string;
	updated_at?: string;
	requester?: Employee;
	assignee?: Employee;
	category?: ServiceTicketCategory;
	comments?: ServiceTicketComment[];
	activity?: ServiceTicketActivity[];
}

export interface ServiceTicketCategory {
	id: string;
	name: string;
	description?: string;
	icon?: string;
	color?: string;
	sla_hours?: number;
	parent_id?: string;
}

export interface ServiceTicketComment {
	id: string;
	ticket_id: string;
	author_id: string;
	content: string;
	is_internal: boolean;
	created_at?: string;
	author?: Employee;
}

export interface ServiceTicketActivity {
	id: string;
	ticket_id: string;
	actor_id?: string;
	action: string;
	old_value?: string;
	new_value?: string;
	created_at?: string;
	actor?: Employee;
}

export interface ServiceDeskStats {
	open_tickets: number;
	sla_breached: number;
	sla_warning: number;
	resolved_today: number;
	sla_compliance: number;
}

// Service Desk API
export const serviceDesk = {
	listTickets: (params?: { requester_id?: string; assignee_id?: string; status?: string; type?: string; priority?: string }) => {
		if (!params) return request<ServiceTicket[]>('/service-desk/tickets');
		const filtered: Record<string, string> = {};
		Object.entries(params).forEach(([k, v]) => { if (v) filtered[k] = v; });
		const query = Object.keys(filtered).length > 0 ? '?' + new URLSearchParams(filtered).toString() : '';
		return request<ServiceTicket[]>(`/service-desk/tickets${query}`);
	},
	getMyTickets: (userId: string) => request<ServiceTicket[]>(`/service-desk/tickets/my?user_id=${userId}`),
	getTicket: (id: string) => request<ServiceTicket>(`/service-desk/tickets/${id}`),
	createTicket: (data: {
		type?: string;
		title: string;
		description?: string;
		category_id?: string;
		priority?: string;
		impact?: string;
		requester_id: string;
	}) => request<ServiceTicket>('/service-desk/tickets', { method: 'POST', body: data }),
	updateTicket: (id: string, data: {
		status?: string;
		priority?: string;
		assignee_id?: string;
		resolution?: string;
		category_id?: string;
		title?: string;
		description?: string;
		actor_id?: string;
	}) => request<ServiceTicket>(`/service-desk/tickets/${id}`, { method: 'PUT', body: data }),
	addComment: (ticketId: string, data: { author_id: string; content: string; is_internal?: boolean }) =>
		request<ServiceTicketComment>(`/service-desk/tickets/${ticketId}/comments`, { method: 'POST', body: data }),
	getCategories: () => request<ServiceTicketCategory[]>('/service-desk/categories'),
	getStats: () => request<ServiceDeskStats>('/service-desk/stats'),
};

// Improvement Request types
export interface ImprovementRequest {
	id: string;
	number: string;  // IR-2026-0001
	title: string;
	description?: string;
	business_value?: string;
	expected_effect?: string;
	initiator_id: string;
	department_id?: string;
	sponsor_id?: string;
	estimated_budget?: number;
	approved_budget?: number;
	estimated_start?: string;
	estimated_end?: string;
	status: 'draft' | 'submitted' | 'screening' | 'evaluation' | 'manager_approval' | 'committee_review' | 'budgeting' | 'project_created' | 'in_progress' | 'completed' | 'rejected';
	committee_date?: string;
	committee_decision?: string;
	project_id?: string;
	rejection_reason?: string;
	rejected_by?: string;
	rejected_at?: string;
	type_id?: string;
	priority?: 'low' | 'medium' | 'high' | 'critical';
	created_at?: string;
	updated_at?: string;
	submitted_at?: string;
	approved_at?: string;
	initiator?: Employee;
	sponsor?: Employee;
	project?: Project;
	type?: ImprovementRequestType;
	comments?: ImprovementRequestComment[];
	approvals?: ImprovementRequestApproval[];
	activity?: ImprovementRequestActivity[];
}

export interface ImprovementRequestType {
	id: string;
	name: string;
	description?: string;
	icon?: string;
	color?: string;
}

export interface ImprovementRequestComment {
	id: string;
	request_id: string;
	author_id: string;
	content: string;
	is_internal: boolean;
	created_at?: string;
	author?: Employee;
}

export interface ImprovementRequestApproval {
	id: string;
	request_id: string;
	approver_id: string;
	stage: string;
	decision: 'approved' | 'rejected' | 'pending';
	comment?: string;
	created_at?: string;
	approver?: Employee;
}

export interface ImprovementRequestActivity {
	id: string;
	request_id: string;
	actor_id?: string;
	action: string;
	old_value?: string;
	new_value?: string;
	created_at?: string;
	actor?: Employee;
}

export interface ImprovementRequestStats {
	total: number;
	draft: number;
	pending: number;
	approved: number;
	rejected: number;
	by_status: Record<string, number>;
}

// Improvement Requests API
export const improvements = {
	list: (params?: { initiator_id?: string; sponsor_id?: string; status?: string; department_id?: string; type_id?: string; priority?: string }) => {
		if (!params) return request<ImprovementRequest[]>('/improvements');
		const filtered: Record<string, string> = {};
		Object.entries(params).forEach(([k, v]) => { if (v) filtered[k] = v; });
		const query = Object.keys(filtered).length > 0 ? '?' + new URLSearchParams(filtered).toString() : '';
		return request<ImprovementRequest[]>(`/improvements${query}`);
	},
	getMy: (userId: string) => request<ImprovementRequest[]>(`/improvements/my?user_id=${userId}`),
	get: (id: string) => request<ImprovementRequest>(`/improvements/${id}`),
	getTypes: () => request<ImprovementRequestType[]>('/improvements/types'),
	getStats: () => request<ImprovementRequestStats>('/improvements/stats'),
	create: (data: {
		title: string;
		description?: string;
		business_value?: string;
		expected_effect?: string;
		initiator_id: string;
		department_id?: string;
		sponsor_id?: string;
		estimated_budget?: number;
		estimated_start?: string;
		estimated_end?: string;
		type_id?: string;
		priority?: string;
	}) => request<ImprovementRequest>('/improvements', { method: 'POST', body: data }),
	update: (id: string, data: {
		title?: string;
		description?: string;
		business_value?: string;
		expected_effect?: string;
		sponsor_id?: string;
		estimated_budget?: number;
		approved_budget?: number;
		estimated_start?: string;
		estimated_end?: string;
		type_id?: string;
		priority?: string;
		committee_date?: string;
		committee_decision?: string;
		actor_id?: string;
	}) => request<ImprovementRequest>(`/improvements/${id}`, { method: 'PUT', body: data }),
	submit: (id: string, actorId: string) =>
		request<ImprovementRequest>(`/improvements/${id}/submit`, { method: 'POST', body: { actor_id: actorId } }),
	approve: (id: string, data: { approver_id: string; comment?: string; approved_budget?: number }) =>
		request<ImprovementRequest>(`/improvements/${id}/approve`, { method: 'POST', body: data }),
	reject: (id: string, data: { rejector_id: string; reason: string }) =>
		request<ImprovementRequest>(`/improvements/${id}/reject`, { method: 'POST', body: data }),
	createProject: (id: string, actorId: string) =>
		request<ImprovementRequest>(`/improvements/${id}/create-project`, { method: 'POST', body: { actor_id: actorId } }),
	addComment: (requestId: string, data: { author_id: string; content: string; is_internal?: boolean }) =>
		request<ImprovementRequestComment>(`/improvements/${requestId}/comments`, { method: 'POST', body: data }),
};

// ======== Resource Planning (GAP-006) ========

export interface ResourceAllocation {
	id: string;
	employee_id: string;
	task_id?: string;
	project_id?: string;
	role?: string;
	allocated_hours_per_week: number;
	period_start: string;
	period_end?: string;
	notes?: string;
	created_at?: string;
	updated_at?: string;
	created_by?: string;
	employee?: Employee;
	task?: Task;
	project?: Project;
}

export interface EmployeeAbsence {
	id: string;
	employee_id: string;
	absence_type: 'vacation' | 'sick_leave' | 'holiday' | 'out_of_office';
	start_date: string;
	end_date: string;
	description?: string;
	source: 'manual' | 'exchange' | 'hr_system';
	created_at?: string;
	employee?: Employee;
}

export interface ResourceCapacity {
	employee_id: string;
	employee_name: string;
	position?: string;
	weekly_hours: number;
	availability_pct: number;
	available_hours: number;
	allocated_hours: number;
	free_hours: number;
	utilization_percent: number;
	overloaded: boolean;
}

export interface ResourceStats {
	total_employees: number;
	total_allocations: number;
	overloaded_count: number;
	underutilized_count: number;
	avg_utilization: number;
}

// Resource Planning API
export const resources = {
	// Allocations
	listAllocations: (params?: { employee_id?: string; project_id?: string; task_id?: string; start_from?: string; end_to?: string }) => {
		if (!params) return request<ResourceAllocation[]>('/resources/allocations');
		const filtered: Record<string, string> = {};
		Object.entries(params).forEach(([k, v]) => { if (v) filtered[k] = v; });
		const query = Object.keys(filtered).length > 0 ? '?' + new URLSearchParams(filtered).toString() : '';
		return request<ResourceAllocation[]>(`/resources/allocations${query}`);
	},
	getAllocation: (id: string) => request<ResourceAllocation>(`/resources/allocations/${id}`),
	createAllocation: (data: {
		employee_id: string;
		task_id?: string;
		project_id?: string;
		role?: string;
		allocated_hours_per_week: number;
		period_start: string;
		period_end?: string;
		notes?: string;
		created_by?: string;
	}) => request<ResourceAllocation>('/resources/allocations', { method: 'POST', body: data }),
	updateAllocation: (id: string, data: Partial<ResourceAllocation>) =>
		request<ResourceAllocation>(`/resources/allocations/${id}`, { method: 'PUT', body: data }),
	deleteAllocation: (id: string) =>
		request<{ message: string }>(`/resources/allocations/${id}`, { method: 'DELETE' }),

	// Capacity
	getCapacity: (params?: { project_id?: string; period_start?: string; period_end?: string }) => {
		if (!params) return request<ResourceCapacity[]>('/resources/capacity');
		const filtered: Record<string, string> = {};
		Object.entries(params).forEach(([k, v]) => { if (v) filtered[k] = v; });
		const query = Object.keys(filtered).length > 0 ? '?' + new URLSearchParams(filtered).toString() : '';
		return request<ResourceCapacity[]>(`/resources/capacity${query}`);
	},
	getStats: (params?: { project_id?: string }) => {
		const query = params?.project_id ? `?project_id=${params.project_id}` : '';
		return request<ResourceStats>(`/resources/stats${query}`);
	},

	// Absences
	listAbsences: (params?: { employee_id?: string; start_from?: string; end_to?: string }) => {
		if (!params) return request<EmployeeAbsence[]>('/resources/absences');
		const filtered: Record<string, string> = {};
		Object.entries(params).forEach(([k, v]) => { if (v) filtered[k] = v; });
		const query = Object.keys(filtered).length > 0 ? '?' + new URLSearchParams(filtered).toString() : '';
		return request<EmployeeAbsence[]>(`/resources/absences${query}`);
	},
	createAbsence: (data: {
		employee_id: string;
		absence_type: string;
		start_date: string;
		end_date: string;
		description?: string;
		source?: string;
	}) => request<EmployeeAbsence>('/resources/absences', { method: 'POST', body: data }),
	deleteAbsence: (id: string) =>
		request<{ message: string }>(`/resources/absences/${id}`, { method: 'DELETE' }),

	// Employee resource settings
	updateEmployeeResourceSettings: (employeeId: string, data: {
		work_hours_per_week?: number;
		availability_percent?: number;
		hourly_rate?: number;
	}) => request<Employee>(`/employees/${employeeId}/resource-settings`, { method: 'PUT', body: data }),
};

// Combined API object for convenience
export const api = {
	employees,
	projects,
	meetings,
	tasks,
	timeEntries,
	analytics,
	calendar,
	messenger,
	connector,
	files,
	bpmn,
	mail,
	giphy,
	github,
	versions,
	sprints,
	admin,
	auth,
	serviceDesk,
	improvements,
	resources
};
