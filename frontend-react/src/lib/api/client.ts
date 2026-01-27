import type {
  User,
  Task,
  Meeting,
  CalendarEvent,
  Email,
  EmailFolder,
  Channel,
  Message,
  ServiceTicket,
  Sprint,
  Version,
  Project,
} from '@/types';

const API_URL = '/api';

interface RequestConfig {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  body?: unknown;
  headers?: Record<string, string>;
}

/**
 * Base API client with HttpOnly cookie authentication
 */
async function request<T>(endpoint: string, config: RequestConfig = {}): Promise<T> {
  const { method = 'GET', body, headers = {} } = config;

  const requestConfig: RequestInit = {
    method,
    credentials: 'include', // Send HttpOnly cookies
    headers: {
      'Content-Type': 'application/json',
      ...headers,
    },
  };

  if (body && method !== 'GET') {
    requestConfig.body = JSON.stringify(body);
  }

  const response = await fetch(`${API_URL}${endpoint}`, requestConfig);

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Request failed' }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return response.json();
}

// Auth API
export const auth = {
  login: async (username: string, password: string) => {
    const formData = new FormData();
    formData.append('username', username);
    formData.append('password', password);

    const response = await fetch(`${API_URL}/v1/ad/authenticate`, {
      method: 'POST',
      credentials: 'include',
      body: formData,
    });

    return response.json();
  },

  logout: async () => {
    await fetch(`${API_URL}/v1/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    });
  },

  refresh: async () => {
    const response = await fetch(`${API_URL}/v1/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
    });
    return response.json();
  },

  getSubordinates: (userId: string) =>
    request<User[]>(`/v1/ad/subordinates/${userId}`),

  changePassword: (userId: string, oldPassword: string, newPassword: string) =>
    request<{ success: boolean }>('/v1/users/change-password', {
      method: 'POST',
      body: { user_id: userId, old_password: oldPassword, new_password: newPassword },
    }),
};

// Employees API
export const employees = {
  list: () => request<User[]>('/v1/employees'),
  get: (id: string) => request<User>(`/v1/employees/${id}`),
  getDossier: (id: string) => request<{
    employee: User;
    one_on_one_count: number;
    tasks: { total: number; done: number; in_progress: number };
    mood_history: Array<{ date: string; score: number }>;
    red_flags_history: Array<{ date: string; flags: Record<string, string> }>;
    recent_meetings: Meeting[];
  }>(`/v1/employees/${id}/dossier`),
  getTeam: (managerId?: string) =>
    request<User[]>(`/v1/employees/team${managerId ? `?manager_id=${managerId}` : ''}`),
  create: (data: Partial<User>) =>
    request<User>('/v1/employees', { method: 'POST', body: data }),
  update: (id: string, data: Partial<User>) =>
    request<User>(`/v1/employees/${id}`, { method: 'PUT', body: data }),
  delete: (id: string) =>
    request<void>(`/v1/employees/${id}`, { method: 'DELETE' }),
};

// Tasks API
export const tasks = {
  list: (filters?: { status?: string; assignee_id?: string; sprint_id?: string }) => {
    const params = new URLSearchParams();
    if (filters?.status) params.append('status', filters.status);
    if (filters?.assignee_id) params.append('assignee_id', filters.assignee_id);
    if (filters?.sprint_id) params.append('sprint_id', filters.sprint_id);
    return request<Task[]>(`/v1/tasks?${params}`);
  },
  get: (id: string) => request<Task>(`/v1/tasks/${id}`),
  create: (data: Partial<Task>) =>
    request<Task>('/v1/tasks', { method: 'POST', body: data }),
  update: (id: string, data: Partial<Task>) =>
    request<Task>(`/v1/tasks/${id}`, { method: 'PUT', body: data }),
  delete: (id: string) =>
    request<void>(`/v1/tasks/${id}`, { method: 'DELETE' }),
  updateStatus: (id: string, status: Task['status']) =>
    request<Task>(`/v1/tasks/${id}/status`, { method: 'PATCH', body: { status } }),
  getDependencies: (id: string) =>
    request<Task[]>(`/v1/tasks/${id}/dependencies`),
  getBlocked: (id: string) =>
    request<{ blocked: boolean; blocked_by: Task[] }>(`/v1/tasks/${id}/blocked`),
};

// Meetings API
export const meetings = {
  list: (filters?: { employee_id?: string; type?: string }) => {
    const params = new URLSearchParams();
    if (filters?.employee_id) params.append('employee_id', filters.employee_id);
    if (filters?.type) params.append('type', filters.type);
    return request<Meeting[]>(`/v1/meetings?${params}`);
  },
  get: (id: string) => request<Meeting>(`/v1/meetings/${id}`),
  create: (data: Partial<Meeting>) =>
    request<Meeting>('/v1/meetings', { method: 'POST', body: data }),
  update: (id: string, data: Partial<Meeting>) =>
    request<Meeting>(`/v1/meetings/${id}`, { method: 'PUT', body: data }),
  delete: (id: string) =>
    request<void>(`/v1/meetings/${id}`, { method: 'DELETE' }),
};

// Calendar API
export const calendar = {
  getEvents: (userId: string, daysBack = 7, daysForward = 30) =>
    request<CalendarEvent[]>(`/v1/calendar?user_id=${userId}&days_back=${daysBack}&days_forward=${daysForward}`),
  getMeetingRooms: () => request<CalendarEvent[]>('/v1/calendar/rooms'),
};

// Mail API
export const mail = {
  getFolders: (email: string, password: string) =>
    request<EmailFolder[]>('/v1/mail/folders', {
      method: 'POST',
      body: { email, password },
    }),
  getEmails: (email: string, password: string, folderId: string, limit = 50) =>
    request<Email[]>('/v1/mail/emails', {
      method: 'POST',
      body: { email, password, folder_id: folderId, limit },
    }),
  getEmail: (email: string, password: string, messageId: string) =>
    request<Email>('/v1/mail/email', {
      method: 'POST',
      body: { email, password, message_id: messageId },
    }),
  sendEmail: (email: string, password: string, to: string[], subject: string, body: string) =>
    request<{ success: boolean }>('/v1/mail/send', {
      method: 'POST',
      body: { email, password, to, subject, body },
    }),
};

// Messenger API
export const messenger = {
  getChannels: () => request<Channel[]>('/v1/messenger/channels'),
  getChannel: (id: string) => request<Channel>(`/v1/messenger/channels/${id}`),
  createChannel: (data: { name: string; type: Channel['type']; member_ids: string[] }) =>
    request<Channel>('/v1/messenger/channels', { method: 'POST', body: data }),
  getMessages: (channelId: string, limit = 50, before?: string) => {
    const params = new URLSearchParams({ limit: String(limit) });
    if (before) params.append('before', before);
    return request<Message[]>(`/v1/messenger/channels/${channelId}/messages?${params}`);
  },
  sendMessage: (channelId: string, content: string, type: Message['type'] = 'text') =>
    request<Message>(`/v1/messenger/channels/${channelId}/messages`, {
      method: 'POST',
      body: { content, type },
    }),
};

// Service Desk API
export const serviceDesk = {
  getTickets: (filters?: { status?: string; type?: string; assignee_id?: string }) => {
    const params = new URLSearchParams();
    if (filters?.status) params.append('status', filters.status);
    if (filters?.type) params.append('type', filters.type);
    if (filters?.assignee_id) params.append('assignee_id', filters.assignee_id);
    return request<ServiceTicket[]>(`/v1/service-desk/tickets?${params}`);
  },
  getTicket: (id: string) => request<ServiceTicket>(`/v1/service-desk/tickets/${id}`),
  createTicket: (data: Partial<ServiceTicket>) =>
    request<ServiceTicket>('/v1/service-desk/tickets', { method: 'POST', body: data }),
  updateTicket: (id: string, data: Partial<ServiceTicket>) =>
    request<ServiceTicket>(`/v1/service-desk/tickets/${id}`, { method: 'PUT', body: data }),
  getStats: () =>
    request<{
      total: number;
      open: number;
      in_progress: number;
      resolved: number;
      overdue: number;
    }>('/v1/service-desk/stats'),
};

// Sprints API
export const sprints = {
  list: () => request<Sprint[]>('/v1/sprints'),
  get: (id: string) => request<Sprint>(`/v1/sprints/${id}`),
  getActive: () => request<Sprint | null>('/v1/sprints/active'),
  create: (data: Partial<Sprint>) =>
    request<Sprint>('/v1/sprints', { method: 'POST', body: data }),
  update: (id: string, data: Partial<Sprint>) =>
    request<Sprint>(`/v1/sprints/${id}`, { method: 'PUT', body: data }),
  start: (id: string) =>
    request<Sprint>(`/v1/sprints/${id}/start`, { method: 'POST' }),
  complete: (id: string) =>
    request<Sprint>(`/v1/sprints/${id}/complete`, { method: 'POST' }),
};

// Versions API
export const versions = {
  list: () => request<Version[]>('/v1/versions'),
  get: (id: string) => request<Version>(`/v1/versions/${id}`),
  create: (data: Partial<Version>) =>
    request<Version>('/v1/versions', { method: 'POST', body: data }),
  update: (id: string, data: Partial<Version>) =>
    request<Version>(`/v1/versions/${id}`, { method: 'PUT', body: data }),
  release: (id: string) =>
    request<Version>(`/v1/versions/${id}/release`, { method: 'POST' }),
};

// Projects API
export const projects = {
  list: () => request<Project[]>('/v1/projects'),
  get: (id: string) => request<Project>(`/v1/projects/${id}`),
  create: (data: Partial<Project>) =>
    request<Project>('/v1/projects', { method: 'POST', body: data }),
  update: (id: string, data: Partial<Project>) =>
    request<Project>(`/v1/projects/${id}`, { method: 'PUT', body: data }),
};

// Analytics API
export const analytics = {
  getOverview: () =>
    request<{
      tasks_completed: number;
      meetings_held: number;
      active_users: number;
      tickets_resolved: number;
    }>('/v1/analytics/overview'),
  getTaskStats: (period: 'week' | 'month' | 'quarter') =>
    request<Array<{ date: string; completed: number; created: number }>>(`/v1/analytics/tasks?period=${period}`),
};

// WebSocket URL helper
export const getWebSocketUrl = (userId: string): string => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  // Token is now sent via HttpOnly cookie, not in URL
  return `${protocol}//${host}/ws/messenger?user_id=${userId}`;
};
