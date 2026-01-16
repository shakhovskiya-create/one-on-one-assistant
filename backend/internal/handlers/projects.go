package handlers

import (
	"encoding/json"

	"github.com/ekf/one-on-one-backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// ListProjects returns all projects
func (h *Handler) ListProjects(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	query := h.DB.From("projects").Select("*")

	if status := c.Query("status"); status != "" {
		query = query.Eq("status", status)
	}

	var projects []models.Project
	err := query.Order("created_at", true).Execute(&projects)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(projects)
}

// GetProject returns a single project with stats
func (h *Handler) GetProject(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var project models.Project
	err := h.DB.From("projects").Select("*").Eq("id", id).Single().Execute(&project)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Project not found"})
	}

	// Get task stats
	var tasks []struct {
		Status string `json:"status"`
	}
	h.DB.From("tasks").Select("status").Eq("project_id", id).Execute(&tasks)

	var meetings []struct {
		ID string `json:"id"`
	}
	h.DB.From("meetings").Select("id").Eq("project_id", id).Execute(&meetings)

	taskCount := len(tasks)
	doneCount := 0
	for _, t := range tasks {
		if t.Status == "done" {
			doneCount++
		}
	}

	progress := 0
	if taskCount > 0 {
		progress = (doneCount * 100) / taskCount
	}

	return c.JSON(fiber.Map{
		"id":            project.ID,
		"name":          project.Name,
		"description":   project.Description,
		"status":        project.Status,
		"start_date":    project.StartDate,
		"end_date":      project.EndDate,
		"created_at":    project.CreatedAt,
		"task_count":    taskCount,
		"meeting_count": len(meetings),
		"progress":      progress,
	})
}

// CreateProject creates a new project
func (h *Handler) CreateProject(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var input struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Status      string  `json:"status"`
		StartDate   *string `json:"start_date"`
		EndDate     *string `json:"end_date"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name is required"})
	}

	// Build insert data without ID (let database generate it)
	data := map[string]interface{}{
		"name": input.Name,
	}
	if input.Description != nil {
		data["description"] = *input.Description
	}
	if input.Status != "" {
		data["status"] = input.Status
	} else {
		data["status"] = "active"
	}
	if input.StartDate != nil {
		data["start_date"] = *input.StartDate
	}
	if input.EndDate != nil {
		data["end_date"] = *input.EndDate
	}

	result, err := h.DB.Insert("projects", data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var created []models.Project
	json.Unmarshal(result, &created)

	if len(created) > 0 {
		return c.Status(201).JSON(created[0])
	}

	return c.Status(201).JSON(fiber.Map{"status": "created"})
}

// UpdateProject updates a project
func (h *Handler) UpdateProject(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	delete(updates, "id")
	delete(updates, "created_at")

	result, err := h.DB.Update("projects", "id", id, updates)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var updated []models.Project
	json.Unmarshal(result, &updated)

	if len(updated) > 0 {
		return c.JSON(updated[0])
	}

	return c.JSON(fiber.Map{"status": "updated"})
}

// DeleteProject deletes a project
func (h *Handler) DeleteProject(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	if err := h.DB.Delete("projects", "id", id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}
