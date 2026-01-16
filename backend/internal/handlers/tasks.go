package handlers

import (
	"encoding/json"
	"time"

	"github.com/ekf/one-on-one-backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// ListTasks returns tasks with filters
func (h *Handler) ListTasks(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	query := h.DB.From("tasks").Select("*, assignee:employees!tasks_assignee_id_fkey(id, name), project:projects(id, name)")

	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		query = query.Eq("assignee_id", assigneeID)
	}
	if projectID := c.Query("project_id"); projectID != "" {
		query = query.Eq("project_id", projectID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Eq("status", status)
	}
	if parentID := c.Query("parent_id"); parentID != "" {
		query = query.Eq("parent_id", parentID)
	} else if c.Query("include_subtasks") != "true" {
		query = query.IsNull("parent_id")
	}
	if isEpic := c.Query("is_epic"); isEpic != "" {
		query = query.Eq("is_epic", isEpic)
	}

	var tasks []models.Task
	err := query.Order("created_at", true).Execute(&tasks)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tasks)
}

// GetTask returns a single task with details
func (h *Handler) GetTask(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var task models.Task
	err := h.DB.From("tasks").
		Select("*, assignee:employees!tasks_assignee_id_fkey(id, name), project:projects(id, name)").
		Eq("id", id).Single().Execute(&task)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Task not found"})
	}

	// Get subtasks if epic
	if task.IsEpic {
		var subtasks []models.Task
		h.DB.From("tasks").Select("*").Eq("parent_id", id).Execute(&subtasks)
		task.Subtasks = subtasks

		if len(subtasks) > 0 {
			doneCount := 0
			for _, s := range subtasks {
				if s.Status == "done" {
					doneCount++
				}
			}
			task.Progress = (doneCount * 100) / len(subtasks)
		}
	}

	// Get comments
	var comments []models.TaskComment
	h.DB.From("task_comments").Select("*, author:employees(name)").Eq("task_id", id).Order("created_at", false).Execute(&comments)
	task.Comments = comments

	// Get history
	var history []models.TaskHistory
	h.DB.From("task_history").Select("*").Eq("task_id", id).Order("created_at", true).Limit(20).Execute(&history)
	task.History = history

	return c.JSON(task)
}

// AddTaskComment adds a comment to a task
func (h *Handler) AddTaskComment(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	taskID := c.Params("id")
	var input struct {
		AuthorID string `json:"author_id"`
		Content  string `json:"content"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.AuthorID == "" || input.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Author ID and content required"})
	}

	result, err := h.DB.Insert("task_comments", map[string]interface{}{
		"task_id":   taskID,
		"author_id": input.AuthorID,
		"content":   input.Content,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var created []models.TaskComment
	json.Unmarshal(result, &created)
	if len(created) == 0 {
		return c.Status(201).JSON(fiber.Map{"status": "created"})
	}

	var comment models.TaskComment
	err = h.DB.From("task_comments").
		Select("*, author:employees(name, photo_base64)").
		Eq("id", created[0].ID).
		Single().
		Execute(&comment)
	if err != nil {
		return c.Status(201).JSON(created[0])
	}

	return c.Status(201).JSON(comment)
}

// CreateTask creates a new task
func (h *Handler) CreateTask(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var input struct {
		Title        string   `json:"title"`
		Description  *string  `json:"description"`
		Status       string   `json:"status"`
		Priority     int      `json:"priority"`
		FlagColor    *string  `json:"flag_color"`
		AssigneeID   *string  `json:"assignee_id"`
		CoAssigneeID *string  `json:"co_assignee_id"`
		CreatorID    *string  `json:"creator_id"`
		MeetingID    *string  `json:"meeting_id"`
		ProjectID    *string  `json:"project_id"`
		ParentID     *string  `json:"parent_id"`
		IsEpic       bool     `json:"is_epic"`
		DueDate      *string  `json:"due_date"`
		Tags         []string `json:"tags"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Status == "" {
		input.Status = "backlog"
	}
	if input.Priority == 0 {
		input.Priority = 3
	}

	taskData := map[string]interface{}{
		"title":          input.Title,
		"description":    input.Description,
		"status":         input.Status,
		"priority":       input.Priority,
		"flag_color":     input.FlagColor,
		"assignee_id":    input.AssigneeID,
		"co_assignee_id": input.CoAssigneeID,
		"creator_id":     input.CreatorID,
		"meeting_id":     input.MeetingID,
		"project_id":     input.ProjectID,
		"parent_id":      input.ParentID,
		"is_epic":        input.IsEpic,
		"due_date":       input.DueDate,
	}

	if input.DueDate != nil {
		taskData["original_due_date"] = input.DueDate
	}

	result, err := h.DB.Insert("tasks", taskData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var created []models.Task
	json.Unmarshal(result, &created)

	if len(created) > 0 {
		// Add tags
		for _, tagName := range input.Tags {
			var tag models.Tag
			h.DB.From("tags").Select("id").Eq("name", tagName).Single().Execute(&tag)
			if tag.ID != "" {
				h.DB.Insert("task_tags", map[string]string{
					"task_id": created[0].ID,
					"tag_id":  tag.ID,
				})
			}
		}

		// Notify assignee
		if input.AssigneeID != nil && h.Telegram != nil {
			go h.notifyNewTask(created[0].ID, *input.AssigneeID, input.Title)
		}

		return c.Status(201).JSON(created[0])
	}

	return c.Status(201).JSON(fiber.Map{"status": "created"})
}

// UpdateTask updates a task
func (h *Handler) UpdateTask(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	// Get current task
	var current models.Task
	h.DB.From("tasks").Select("*").Eq("id", id).Single().Execute(&current)

	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	delete(updates, "id")
	delete(updates, "created_at")

	// Track history
	for field, newValue := range updates {
		var oldValue interface{}
		switch field {
		case "status":
			oldValue = current.Status
		case "priority":
			oldValue = current.Priority
		case "assignee_id":
			oldValue = current.AssigneeID
		case "due_date":
			oldValue = current.DueDate
		}

		if oldValue != newValue {
			h.DB.Insert("task_history", map[string]interface{}{
				"task_id":    id,
				"field_name": field,
				"old_value":  oldValue,
				"new_value":  newValue,
			})
		}
	}

	// Mark completed
	if newStatus, ok := updates["status"].(string); ok && newStatus == "done" && current.Status != "done" {
		updates["completed_at"] = time.Now().Format(time.RFC3339)
	}

	result, err := h.DB.Update("tasks", "id", id, updates)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var updated []models.Task
	json.Unmarshal(result, &updated)

	if len(updated) > 0 {
		return c.JSON(updated[0])
	}

	return c.JSON(fiber.Map{"status": "updated"})
}

// DeleteTask deletes a task
func (h *Handler) DeleteTask(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	if err := h.DB.Delete("tasks", "id", id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

// GetKanban returns tasks organized by status
func (h *Handler) GetKanban(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	query := h.DB.From("tasks").Select("*, assignee:employees!tasks_assignee_id_fkey(id, name)").IsNull("parent_id")

	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		query = query.Eq("assignee_id", assigneeID)
	}
	if projectID := c.Query("project_id"); projectID != "" {
		query = query.Eq("project_id", projectID)
	}

	var tasks []models.Task
	query.Execute(&tasks)

	kanban := map[string][]models.Task{
		"backlog":     {},
		"todo":        {},
		"in_progress": {},
		"review":      {},
		"done":        {},
	}

	for _, task := range tasks {
		if _, ok := kanban[task.Status]; ok {
			kanban[task.Status] = append(kanban[task.Status], task)
		}
	}

	return c.JSON(kanban)
}

// MoveTaskKanban moves a task to a new status
func (h *Handler) MoveTaskKanban(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	taskID := c.Query("task_id")
	newStatus := c.Query("new_status")

	if taskID == "" || newStatus == "" {
		return c.Status(400).JSON(fiber.Map{"error": "task_id and new_status required"})
	}

	// Get current status
	var current models.Task
	h.DB.From("tasks").Select("status").Eq("id", taskID).Single().Execute(&current)

	updates := map[string]interface{}{"status": newStatus}
	if newStatus == "done" && current.Status != "done" {
		updates["completed_at"] = time.Now().Format(time.RFC3339)
	}

	result, err := h.DB.Update("tasks", "id", taskID, updates)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Track history
	if current.Status != newStatus {
		h.DB.Insert("task_history", map[string]interface{}{
			"task_id":    taskID,
			"field_name": "status",
			"old_value":  current.Status,
			"new_value":  newStatus,
		})
	}

	var updated []models.Task
	json.Unmarshal(result, &updated)

	if len(updated) > 0 {
		return c.JSON(updated[0])
	}

	return c.JSON(fiber.Map{"status": "updated"})
}

func (h *Handler) notifyNewTask(taskID, assigneeID, title string) {
	if h.DB == nil || h.Telegram == nil {
		return
	}

	var user models.TelegramUser
	h.DB.From("telegram_users").Select("telegram_chat_id").Eq("employee_id", assigneeID).Eq("notifications_enabled", "true").Single().Execute(&user)

	if user.TelegramChatID > 0 {
		h.Telegram.SendMessage(user.TelegramChatID, "📌 <b>Новая задача:</b>\n\n"+title)
	}
}
