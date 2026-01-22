package handlers

import (
	"encoding/json"
	"time"

	"github.com/ekf/one-on-one-backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// ListVersions returns all versions, optionally filtered by project
func (h *Handler) ListVersions(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	projectID := c.Query("project_id")
	status := c.Query("status")

	query := h.DB.From("versions").Select("*").Order("release_date", false)

	if projectID != "" {
		query = query.Eq("project_id", projectID)
	}
	if status != "" {
		query = query.Eq("status", status)
	}

	var versions []models.Version
	if err := query.Execute(&versions); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Compute task counts for each version
	for i := range versions {
		var counts []struct {
			Total int `json:"total"`
			Done  int `json:"done"`
		}
		h.DB.From("tasks").
			Select("count(*) as total, count(*) filter (where status = 'done') as done").
			Eq("fix_version_id", versions[i].ID).
			Execute(&counts)

		if len(counts) > 0 {
			versions[i].TasksCount = counts[0].Total
			versions[i].TasksDone = counts[0].Done
			if counts[0].Total > 0 {
				versions[i].Progress = (counts[0].Done * 100) / counts[0].Total
			}
		}
	}

	return c.JSON(versions)
}

// GetVersion returns a single version by ID
func (h *Handler) GetVersion(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var version models.Version
	err := h.DB.From("versions").Select("*").Eq("id", id).Single().Execute(&version)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Version not found"})
	}

	// Get task counts
	var counts []struct {
		Total int `json:"total"`
		Done  int `json:"done"`
	}
	h.DB.From("tasks").
		Select("count(*) as total, count(*) filter (where status = 'done') as done").
		Eq("fix_version_id", id).
		Execute(&counts)

	if len(counts) > 0 {
		version.TasksCount = counts[0].Total
		version.TasksDone = counts[0].Done
		if counts[0].Total > 0 {
			version.Progress = (counts[0].Done * 100) / counts[0].Total
		}
	}

	// Get tasks in this version
	var tasks []models.Task
	h.DB.From("tasks").Select("*").Eq("fix_version_id", id).Order("priority", false).Execute(&tasks)

	return c.JSON(fiber.Map{
		"version": version,
		"tasks":   tasks,
	})
}

// CreateVersion creates a new version
func (h *Handler) CreateVersion(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID, _ := c.Locals("user_id").(string)

	var req struct {
		ProjectID   string  `json:"project_id"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
		StartDate   *string `json:"start_date"`
		ReleaseDate *string `json:"release_date"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name is required"})
	}

	data := map[string]interface{}{
		"name":   req.Name,
		"status": "unreleased",
	}

	if req.ProjectID != "" {
		data["project_id"] = req.ProjectID
	}
	if req.Description != nil {
		data["description"] = *req.Description
	}
	if req.StartDate != nil {
		data["start_date"] = *req.StartDate
	}
	if req.ReleaseDate != nil {
		data["release_date"] = *req.ReleaseDate
	}
	if userID != "" {
		data["created_by"] = userID
	}

	result, err := h.DB.Insert("versions", data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var created []models.Version
	json.Unmarshal(result, &created)

	if len(created) > 0 {
		return c.Status(201).JSON(created[0])
	}

	return c.Status(201).JSON(fiber.Map{"success": true})
}

// UpdateVersion updates a version
func (h *Handler) UpdateVersion(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
		StartDate   *string `json:"start_date"`
		ReleaseDate *string `json:"release_date"`
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
	if req.Description != nil {
		data["description"] = *req.Description
	}
	if req.Status != nil {
		data["status"] = *req.Status
		// If releasing, set released_at
		if *req.Status == "released" {
			data["released_at"] = time.Now().Format(time.RFC3339)
		}
	}
	if req.StartDate != nil {
		data["start_date"] = *req.StartDate
	}
	if req.ReleaseDate != nil {
		data["release_date"] = *req.ReleaseDate
	}

	result, err := h.DB.Update("versions", "id", id, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var updated []models.Version
	json.Unmarshal(result, &updated)

	if len(updated) > 0 {
		return c.JSON(updated[0])
	}

	return c.JSON(fiber.Map{"success": true})
}

// DeleteVersion deletes a version
func (h *Handler) DeleteVersion(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	// First, remove fix_version_id from all tasks with this version
	h.DB.Update("tasks", "fix_version_id", id, map[string]interface{}{
		"fix_version_id": nil,
	})

	// Delete the version
	h.DB.Delete("versions", "id", id)

	return c.JSON(fiber.Map{"success": true})
}

// ReleaseVersion marks a version as released
func (h *Handler) ReleaseVersion(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	data := map[string]interface{}{
		"status":      "released",
		"released_at": time.Now().Format(time.RFC3339),
		"updated_at":  time.Now().Format(time.RFC3339),
	}

	result, err := h.DB.Update("versions", "id", id, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var updated []models.Version
	json.Unmarshal(result, &updated)

	if len(updated) > 0 {
		return c.JSON(updated[0])
	}

	return c.JSON(fiber.Map{"success": true})
}

// GetVersionReleaseNotes generates release notes for a version
func (h *Handler) GetVersionReleaseNotes(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	// Get version
	var version models.Version
	err := h.DB.From("versions").Select("*").Eq("id", id).Single().Execute(&version)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Version not found"})
	}

	// Get all completed tasks in this version
	var tasks []models.Task
	h.DB.From("tasks").Select("*").
		Eq("fix_version_id", id).
		Eq("status", "done").
		Order("priority", false).
		Execute(&tasks)

	// Group by type/category
	features := []models.Task{}
	fixes := []models.Task{}
	other := []models.Task{}

	for _, t := range tasks {
		title := t.Title
		if len(title) > 3 {
			prefix := title[:3]
			switch prefix {
			case "fix", "Fix", "FIX":
				fixes = append(fixes, t)
			case "fea", "Fea", "add", "Add", "new", "New":
				features = append(features, t)
			default:
				other = append(other, t)
			}
		} else {
			other = append(other, t)
		}
	}

	return c.JSON(fiber.Map{
		"version":  version,
		"features": features,
		"fixes":    fixes,
		"other":    other,
		"total":    len(tasks),
	})
}
