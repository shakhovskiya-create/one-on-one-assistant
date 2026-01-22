package models

import "time"

// Employee represents a team member
type Employee struct {
	ID                    string     `json:"id"`
	Name                  string     `json:"name"`
	Email                 string     `json:"email,omitempty"`
	Position              string     `json:"position"`
	Department            string     `json:"department,omitempty"`
	ManagerID             *string    `json:"manager_id,omitempty"`
	MeetingFrequency      string     `json:"meeting_frequency,omitempty"`
	MeetingDay            *string    `json:"meeting_day,omitempty"`
	DevelopmentPriorities *string    `json:"development_priorities,omitempty"`
	PhotoBase64           *string    `json:"photo_base64,omitempty"`
	ADDN                  *string    `json:"ad_dn,omitempty"`
	ADLogin               *string    `json:"ad_login,omitempty"`
	Phone                 *string    `json:"phone,omitempty"`
	Mobile                *string    `json:"mobile,omitempty"`
	TelegramUsername      *string    `json:"telegram_username,omitempty"`
	TelegramChatID        *int64     `json:"telegram_chat_id,omitempty"`
	HourlyRate            *float64   `json:"hourly_rate,omitempty"`
	EncryptedPassword     *string    `json:"-"` // Never expose in JSON - for EWS access only
	CreatedAt             *time.Time `json:"created_at,omitempty"`
}

// Project represents a team project
type Project struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Status      string     `json:"status"`
	StartDate   *string    `json:"start_date,omitempty"`
	EndDate     *string    `json:"end_date,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

// MeetingCategory represents a type of meeting
type MeetingCategory struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// Meeting represents a recorded meeting
type Meeting struct {
	ID                string                 `json:"id"`
	Title             *string                `json:"title,omitempty"`
	EmployeeID        *string                `json:"employee_id,omitempty"`
	ProjectID         *string                `json:"project_id,omitempty"`
	CategoryID        *string                `json:"category_id,omitempty"`
	OrganizerID       *string                `json:"organizer_id,omitempty"`
	Date              string                 `json:"date"`
	StartTime         *string                `json:"start_time,omitempty"`
	EndTime           *string                `json:"end_time,omitempty"`
	Location          *string                `json:"location,omitempty"`
	ExchangeID        *string                `json:"exchange_id,omitempty"`
	ExchangeData      map[string]interface{} `json:"exchange_data,omitempty"`
	Transcript        *string                `json:"transcript,omitempty"`
	TranscriptWhisper *string                `json:"transcript_whisper,omitempty"`
	TranscriptYandex  *string                `json:"transcript_yandex,omitempty"`
	TranscriptMerged  *string                `json:"transcript_merged,omitempty"`
	Summary           *string                `json:"summary,omitempty"`
	MoodScore         *int                   `json:"mood_score,omitempty"`
	Analysis          map[string]interface{} `json:"analysis,omitempty"`
	CreatedAt         *time.Time             `json:"created_at,omitempty"`

	// Joined fields - tags must match PostgreSQL relation names
	Employee *Employee        `json:"employees,omitempty"`
	Project  *Project         `json:"projects,omitempty"`
	Category *MeetingCategory `json:"meeting_categories,omitempty"`
}

// MeetingParticipant links employees to meetings
type MeetingParticipant struct {
	ID         string    `json:"id"`
	MeetingID  string    `json:"meeting_id"`
	EmployeeID string    `json:"employee_id"`
	Employee   *Employee `json:"employees,omitempty"`
}

// Agreement represents a commitment from a meeting
type Agreement struct {
	ID          string   `json:"id"`
	MeetingID   string   `json:"meeting_id"`
	Task        string   `json:"task"`
	Responsible string   `json:"responsible"`
	Deadline    *string  `json:"deadline,omitempty"`
	Status      string   `json:"status"`
	Meeting     *Meeting `json:"meetings,omitempty"`
}

// Task represents a work item
type Task struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	Status          string     `json:"status"`
	Priority        int        `json:"priority"`
	StoryPoints     *int       `json:"story_points,omitempty"`
	Sprint          *string    `json:"sprint,omitempty"`
	FlagColor       *string    `json:"flag_color,omitempty"`
	AssigneeID      *string    `json:"assignee_id,omitempty"`
	CoAssigneeID    *string    `json:"co_assignee_id,omitempty"`
	CreatorID       *string    `json:"creator_id,omitempty"`
	MeetingID       *string    `json:"meeting_id,omitempty"`
	ProjectID       *string    `json:"project_id,omitempty"`
	ParentID        *string    `json:"parent_id,omitempty"`
	IsEpic          bool       `json:"is_epic"`
	DueDate         *string    `json:"due_date,omitempty"`
	OriginalDueDate *string    `json:"original_due_date,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	// Resource planning fields
	EstimatedHours *float64 `json:"estimated_hours,omitempty"`
	ActualHours    *float64 `json:"actual_hours,omitempty"`
	EstimatedCost  *float64 `json:"estimated_cost,omitempty"`
	ActualCost     *float64 `json:"actual_cost,omitempty"`
	// Release version
	FixVersionID *string `json:"fix_version_id,omitempty"`
	// Sprint
	SprintID *string `json:"sprint_id,omitempty"`
	// Relations
	Tags       []Tag         `json:"tags,omitempty"`
	FixVersion *Version      `json:"fix_version,omitempty"`
	SprintRef  *Sprint       `json:"sprint_ref,omitempty"`
	Assignee   *Employee     `json:"assignee,omitempty"`
	CoAssignee *Employee     `json:"co_assignee,omitempty"`
	Creator    *Employee     `json:"creator,omitempty"`
	Project    *Project      `json:"project,omitempty"`
	Subtasks   []Task        `json:"subtasks,omitempty"`
	Comments   []TaskComment `json:"comments,omitempty"`
	History    []TaskHistory `json:"history,omitempty"`
	Progress   int           `json:"progress,omitempty"`
}

// Tag represents a label for tasks
type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// TaskComment represents a comment on a task
type TaskComment struct {
	ID        string     `json:"id"`
	TaskID    string     `json:"task_id"`
	AuthorID  string     `json:"author_id"`
	Content   string     `json:"content"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	Author    *Employee  `json:"author,omitempty"`
}

// TaskHistory represents a change to a task
type TaskHistory struct {
	ID        string     `json:"id"`
	TaskID    string     `json:"task_id"`
	FieldName string     `json:"field_name"`
	OldValue  *string    `json:"old_value,omitempty"`
	NewValue  *string    `json:"new_value,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// TimeEntry represents time logged on a task
type TimeEntry struct {
	ID          string     `json:"id"`
	TaskID      string     `json:"task_id"`
	EmployeeID  string     `json:"employee_id"`
	Hours       float64    `json:"hours"`
	Description *string    `json:"description,omitempty"`
	Date        string     `json:"date"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	Employee    *Employee  `json:"employee,omitempty"`
}

// TelegramUser links employees to Telegram
type TelegramUser struct {
	ID                   string `json:"id"`
	EmployeeID           string `json:"employee_id"`
	TelegramUsername     string `json:"telegram_username"`
	TelegramChatID       int64  `json:"telegram_chat_id"`
	NotificationsEnabled bool   `json:"notifications_enabled"`
}

// Conversation represents a chat between users
type Conversation struct {
	ID              string     `json:"id"`
	Type            string     `json:"type"` // "direct", "group", or "channel"
	Name            *string    `json:"name,omitempty"`
	Description     *string    `json:"description,omitempty"` // For channels
	CreatedBy       *string    `json:"created_by,omitempty"`  // Channel owner
	TelegramEnabled bool       `json:"telegram_enabled"`      // Telegram bot linked
	TelegramChatID  *int64     `json:"telegram_chat_id,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
	Participants    []Employee `json:"participants,omitempty"`
	LastMessage     *Message   `json:"last_message,omitempty"`
	UnreadCount     int        `json:"unread_count,omitempty"`
}

// ConversationParticipant links users to conversations
type ConversationParticipant struct {
	ID             string     `json:"id"`
	ConversationID string     `json:"conversation_id"`
	EmployeeID     string     `json:"employee_id"`
	JoinedAt       *time.Time `json:"joined_at,omitempty"`
	LastReadAt     *time.Time `json:"last_read_at,omitempty"`
	Employee       *Employee  `json:"employee,omitempty"`
}

// Message represents a chat message
type Message struct {
	ID              string     `json:"id"`
	ConversationID  string     `json:"conversation_id"`
	SenderID        string     `json:"sender_id"`
	Content         string     `json:"content"`
	MessageType     string     `json:"message_type"` // "text", "file", "voice", "video", "sticker", "gif", "system"
	FileID          *string    `json:"file_id,omitempty"`
	FileURL         string     `json:"file_url,omitempty"`         // Populated at runtime from Storage
	DurationSeconds *int       `json:"duration_seconds,omitempty"` // For voice/video messages
	ThumbnailURL    *string    `json:"thumbnail_url,omitempty"`    // For video messages
	ReplyToID       *string    `json:"reply_to_id,omitempty"`
	EditedAt        *time.Time `json:"edited_at,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	Sender          *Employee  `json:"sender,omitempty"`
	ReplyTo         *Message   `json:"reply_to,omitempty"`
}

// Version represents a release version for a project (JIRA-like)
type Version struct {
	ID          string     `json:"id"`
	ProjectID   *string    `json:"project_id,omitempty"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Status      string     `json:"status"` // "unreleased", "released", "archived"
	StartDate   *string    `json:"start_date,omitempty"`
	ReleaseDate *string    `json:"release_date,omitempty"`
	ReleasedAt  *time.Time `json:"released_at,omitempty"`
	CreatedBy   *string    `json:"created_by,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	// Computed fields
	Project    *Project `json:"project,omitempty"`
	TasksCount int      `json:"tasks_count,omitempty"`
	TasksDone  int      `json:"tasks_done,omitempty"`
	Progress   int      `json:"progress,omitempty"` // percentage
}

// Sprint represents a Scrum sprint for task grouping
type Sprint struct {
	ID        string     `json:"id"`
	ProjectID *string    `json:"project_id,omitempty"`
	Name      string     `json:"name"`
	Goal      *string    `json:"goal,omitempty"`
	StartDate string     `json:"start_date"`
	EndDate   string     `json:"end_date"`
	Status    string     `json:"status"` // "planning", "active", "completed"
	Velocity  int        `json:"velocity,omitempty"`
	CreatedBy *string    `json:"created_by,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	// Computed fields
	Project         *Project `json:"project,omitempty"`
	TasksCount      int      `json:"tasks_count,omitempty"`
	TasksDone       int      `json:"tasks_done,omitempty"`
	TotalPoints     int      `json:"total_points,omitempty"`
	CompletedPoints int      `json:"completed_points,omitempty"`
	Progress        int      `json:"progress,omitempty"` // percentage
}
