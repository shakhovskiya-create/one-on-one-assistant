package handlers

import (
	"encoding/json"
	"time"

	"github.com/ekf/one-on-one-backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// ListSprints returns all sprints, optionally filtered by project or status
func (h *Handler) ListSprints(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	projectID := c.Query("project_id")
	status := c.Query("status")

	query := h.DB.From("sprints").Select("*").Order("start_date", false)

	if projectID != "" {
		query = query.Eq("project_id", projectID)
	}
	if status != "" {
		query = query.Eq("status", status)
	}

	var sprints []models.Sprint
	if err := query.Execute(&sprints); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Compute task counts and points for each sprint
	for i := range sprints {
		var counts []struct {
			Total           int `json:"total"`
			Done            int `json:"done"`
			TotalPoints     int `json:"total_points"`
			CompletedPoints int `json:"completed_points"`
		}
		h.DB.From("tasks").
			Select("count(*) as total, count(*) filter (where status = 'done') as done, coalesce(sum(story_points), 0) as total_points, coalesce(sum(story_points) filter (where status = 'done'), 0) as completed_points").
			Eq("sprint_id", sprints[i].ID).
			Execute(&counts)

		if len(counts) > 0 {
			sprints[i].TasksCount = counts[0].Total
			sprints[i].TasksDone = counts[0].Done
			sprints[i].TotalPoints = counts[0].TotalPoints
			sprints[i].CompletedPoints = counts[0].CompletedPoints
			if counts[0].Total > 0 {
				sprints[i].Progress = (counts[0].Done * 100) / counts[0].Total
			}
		}
	}

	return c.JSON(sprints)
}

// GetSprint returns a single sprint by ID with its tasks
func (h *Handler) GetSprint(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var sprint models.Sprint
	err := h.DB.From("sprints").Select("*").Eq("id", id).Single().Execute(&sprint)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Sprint not found"})
	}

	// Get task counts and points
	var counts []struct {
		Total           int `json:"total"`
		Done            int `json:"done"`
		TotalPoints     int `json:"total_points"`
		CompletedPoints int `json:"completed_points"`
	}
	h.DB.From("tasks").
		Select("count(*) as total, count(*) filter (where status = 'done') as done, coalesce(sum(story_points), 0) as total_points, coalesce(sum(story_points) filter (where status = 'done'), 0) as completed_points").
		Eq("sprint_id", id).
		Execute(&counts)

	if len(counts) > 0 {
		sprint.TasksCount = counts[0].Total
		sprint.TasksDone = counts[0].Done
		sprint.TotalPoints = counts[0].TotalPoints
		sprint.CompletedPoints = counts[0].CompletedPoints
		if counts[0].Total > 0 {
			sprint.Progress = (counts[0].Done * 100) / counts[0].Total
		}
	}

	// Get tasks in this sprint
	var tasks []models.Task
	h.DB.From("tasks").Select("*").Eq("sprint_id", id).Order("priority", false).Execute(&tasks)

	return c.JSON(fiber.Map{
		"sprint": sprint,
		"tasks":  tasks,
	})
}

// GetActiveSprint returns the currently active sprint for a project
func (h *Handler) GetActiveSprint(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	projectID := c.Query("project_id")

	query := h.DB.From("sprints").Select("*").Eq("status", "active")
	if projectID != "" {
		query = query.Eq("project_id", projectID)
	}

	var sprints []models.Sprint
	if err := query.Execute(&sprints); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if len(sprints) == 0 {
		return c.JSON(nil)
	}

	sprint := sprints[0]

	// Get task counts
	var counts []struct {
		Total           int `json:"total"`
		Done            int `json:"done"`
		TotalPoints     int `json:"total_points"`
		CompletedPoints int `json:"completed_points"`
	}
	h.DB.From("tasks").
		Select("count(*) as total, count(*) filter (where status = 'done') as done, coalesce(sum(story_points), 0) as total_points, coalesce(sum(story_points) filter (where status = 'done'), 0) as completed_points").
		Eq("sprint_id", sprint.ID).
		Execute(&counts)

	if len(counts) > 0 {
		sprint.TasksCount = counts[0].Total
		sprint.TasksDone = counts[0].Done
		sprint.TotalPoints = counts[0].TotalPoints
		sprint.CompletedPoints = counts[0].CompletedPoints
		if counts[0].Total > 0 {
			sprint.Progress = (counts[0].Done * 100) / counts[0].Total
		}
	}

	return c.JSON(sprint)
}

// CreateSprint creates a new sprint
func (h *Handler) CreateSprint(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID, _ := c.Locals("user_id").(string)

	var req struct {
		ProjectID *string `json:"project_id"`
		Name      string  `json:"name"`
		Goal      *string `json:"goal"`
		StartDate string  `json:"start_date"`
		EndDate   string  `json:"end_date"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name is required"})
	}
	if req.StartDate == "" || req.EndDate == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Start and end dates are required"})
	}

	data := map[string]interface{}{
		"name":       req.Name,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
		"status":     "planning",
	}

	if req.ProjectID != nil && *req.ProjectID != "" {
		data["project_id"] = *req.ProjectID
	}
	if req.Goal != nil {
		data["goal"] = *req.Goal
	}
	if userID != "" {
		data["created_by"] = userID
	}

	result, err := h.DB.Insert("sprints", data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var created []models.Sprint
	json.Unmarshal(result, &created)

	if len(created) > 0 {
		return c.Status(201).JSON(created[0])
	}

	return c.Status(201).JSON(fiber.Map{"success": true})
}

// UpdateSprint updates a sprint
func (h *Handler) UpdateSprint(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var req struct {
		Name      *string `json:"name"`
		Goal      *string `json:"goal"`
		StartDate *string `json:"start_date"`
		EndDate   *string `json:"end_date"`
		Status    *string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	data := map[string]interface{}{
		"updated_at": time.Now().Format(time.RFC3339),
	}

	if req.Name != nil {
		data["name"] = *req.Name
	}
	if req.Goal != nil {
		data["goal"] = *req.Goal
	}
	if req.StartDate != nil {
		data["start_date"] = *req.StartDate
	}
	if req.EndDate != nil {
		data["end_date"] = *req.EndDate
	}
	if req.Status != nil {
		// If activating, deactivate other sprints in same project first
		if *req.Status == "active" {
			// Get current sprint to find project_id
			var current models.Sprint
			h.DB.From("sprints").Select("project_id").Eq("id", id).Single().Execute(&current)

			// Deactivate other active sprints in same project
			if current.ProjectID != nil {
				h.DB.From("sprints").
					Update(map[string]interface{}{"status": "planning"}).
					Eq("project_id", *current.ProjectID).
					Eq("status", "active").
					Neq("id", id)
			}
		}
		data["status"] = *req.Status

		// If completing, calculate velocity
		if *req.Status == "completed" {
			var counts []struct {
				CompletedPoints int `json:"completed_points"`
			}
			h.DB.From("tasks").
				Select("coalesce(sum(story_points) filter (where status = 'done'), 0) as completed_points").
				Eq("sprint_id", id).
				Execute(&counts)

			if len(counts) > 0 {
				data["velocity"] = counts[0].CompletedPoints
			}
		}
	}

	result, err := h.DB.Update("sprints", "id", id, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var updated []models.Sprint
	json.Unmarshal(result, &updated)

	if len(updated) > 0 {
		return c.JSON(updated[0])
	}

	return c.JSON(fiber.Map{"success": true})
}

// DeleteSprint deletes a sprint
func (h *Handler) DeleteSprint(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	// First, remove sprint_id from all tasks
	h.DB.Update("tasks", "sprint_id", id, map[string]interface{}{
		"sprint_id": nil,
	})

	// Delete the sprint
	h.DB.Delete("sprints", "id", id)

	return c.JSON(fiber.Map{"success": true})
}

// StartSprint activates a sprint
func (h *Handler) StartSprint(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	// Get sprint to find project_id
	var sprint models.Sprint
	err := h.DB.From("sprints").Select("*").Eq("id", id).Single().Execute(&sprint)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Sprint not found"})
	}

	// Deactivate other active sprints in same project
	if sprint.ProjectID != nil {
		h.DB.From("sprints").
			Update(map[string]interface{}{"status": "planning", "updated_at": time.Now().Format(time.RFC3339)}).
			Eq("project_id", *sprint.ProjectID).
			Eq("status", "active")
	}

	// Activate this sprint
	data := map[string]interface{}{
		"status":     "active",
		"updated_at": time.Now().Format(time.RFC3339),
	}

	result, err := h.DB.Update("sprints", "id", id, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var updated []models.Sprint
	json.Unmarshal(result, &updated)

	if len(updated) > 0 {
		return c.JSON(updated[0])
	}

	return c.JSON(fiber.Map{"success": true})
}

// CompleteSprint completes a sprint and calculates velocity
func (h *Handler) CompleteSprint(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	// Calculate velocity (completed story points)
	var counts []struct {
		CompletedPoints int `json:"completed_points"`
	}
	h.DB.From("tasks").
		Select("coalesce(sum(story_points) filter (where status = 'done'), 0) as completed_points").
		Eq("sprint_id", id).
		Execute(&counts)

	velocity := 0
	if len(counts) > 0 {
		velocity = counts[0].CompletedPoints
	}

	data := map[string]interface{}{
		"status":     "completed",
		"velocity":   velocity,
		"updated_at": time.Now().Format(time.RFC3339),
	}

	result, err := h.DB.Update("sprints", "id", id, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var updated []models.Sprint
	json.Unmarshal(result, &updated)

	if len(updated) > 0 {
		// Add computed fields
		updated[0].Velocity = velocity
		return c.JSON(updated[0])
	}

	return c.JSON(fiber.Map{"success": true, "velocity": velocity})
}
