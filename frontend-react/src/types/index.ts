// User & Auth Types
export interface User {
  id: string;
  name: string;
  email: string;
  position: string;
  department: string | null;
  manager_id: string | null;
  photo_base64?: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  subordinates: User[];
}

// Task Types
export interface Task {
  id: string;
  title: string;
  description: string;
  status: 'backlog' | 'todo' | 'in_progress' | 'review' | 'done';
  priority: 'critical' | 'high' | 'medium' | 'low';
  task_type: 'feature' | 'bug' | 'techdebt' | 'improvement';
  assignee_id?: string;
  assignee?: User;
  reporter_id?: string;
  reporter?: User;
  project_id?: string;
  sprint_id?: string;
  fix_version_id?: string;
  story_points?: number;
  due_date?: string;
  created_at: string;
  updated_at: string;
}

// Meeting Types
export interface Meeting {
  id: string;
  title: string;
  date: string;
  start_time: string;
  end_time: string;
  location?: string;
  type: 'one_on_one' | 'team' | 'project' | 'standup' | 'retrospective';
  employee_id?: string;
  manager_id?: string;
  participants?: User[];
  status: 'scheduled' | 'in_progress' | 'completed' | 'cancelled';
  notes?: string;
  mood_score?: number;
  analysis?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

// Calendar Event Types
export interface CalendarEvent {
  id: string;
  subject: string;
  start: string;
  end: string;
  location?: string;
  organizer?: string;
  attendees?: string[];
  is_recurring?: boolean;
  body?: string;
}

// Email Types
export interface Email {
  id: string;
  subject: string;
  from: string;
  to: string[];
  cc?: string[];
  date: string;
  body: string;
  is_read: boolean;
  has_attachments: boolean;
  folder: string;
}

export interface EmailFolder {
  id: string;
  name: string;
  unread_count: number;
  total_count: number;
}

// Messenger Types
export interface Channel {
  id: string;
  name: string;
  type: 'direct' | 'group' | 'channel';
  members: User[];
  last_message?: Message;
  unread_count: number;
  created_at: string;
}

export interface Message {
  id: string;
  channel_id: string;
  sender_id: string;
  sender?: User;
  content: string;
  type: 'text' | 'file' | 'image' | 'gif';
  file_url?: string;
  reactions?: Record<string, string[]>;
  created_at: string;
  updated_at?: string;
}

// Service Desk Types
export interface ServiceTicket {
  id: string;
  ticket_number: string;
  title: string;
  description: string;
  type: 'incident' | 'service_request' | 'change' | 'problem';
  status: 'new' | 'open' | 'in_progress' | 'pending' | 'resolved' | 'closed';
  priority: 'critical' | 'high' | 'medium' | 'low';
  category_id?: string;
  requester_id: string;
  requester?: User;
  assignee_id?: string;
  assignee?: User;
  sla_due_date?: string;
  resolution?: string;
  created_at: string;
  updated_at: string;
}

// Sprint Types
export interface Sprint {
  id: string;
  name: string;
  goal?: string;
  start_date: string;
  end_date: string;
  status: 'planning' | 'active' | 'completed';
  velocity?: number;
  created_at: string;
}

// Version/Release Types
export interface Version {
  id: string;
  name: string;
  description?: string;
  release_date?: string;
  status: 'unreleased' | 'released' | 'archived';
  created_at: string;
}

// Project Types
export interface Project {
  id: string;
  name: string;
  key: string;
  description?: string;
  lead_id?: string;
  lead?: User;
  status: 'active' | 'archived';
  created_at: string;
}

// API Response Types
export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}
