package handlers

import (
	"encoding/json"
	"time"

	"github.com/ekf/one-on-one-backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// ListTimeEntries returns time entries for a task
func (h *Handler) ListTimeEntries(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	taskID := c.Params("id")
	if taskID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Task ID is required"})
	}

	var entries []models.TimeEntry
	err := h.DB.From("time_entries").Select("*, employee:employees(id, name)").
		Eq("task_id", taskID).Order("date", true).Execute(&entries)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(entries)
}

// CreateTimeEntry creates a new time entry for a task
func (h *Handler) CreateTimeEntry(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	taskID := c.Params("id")
	userID, _ := c.Locals("user_id").(string)

	var req struct {
		Hours       float64 `json:"hours"`
		Description string  `json:"description"`
		Date        string  `json:"date"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Hours <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Hours must be positive"})
	}

	// Default to today if no date provided
	if req.Date == "" {
		req.Date = time.Now().Format("2006-01-02")
	}

	data := map[string]interface{}{
		"task_id":     taskID,
		"employee_id": userID,
		"hours":       req.Hours,
		"description": req.Description,
		"date":        req.Date,
	}

	result, err := h.DB.Insert("time_entries", data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var entry models.TimeEntry
	json.Unmarshal(result, &entry)

	return c.Status(201).JSON(entry)
}

// UpdateTimeEntry updates a time entry
func (h *Handler) UpdateTimeEntry(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	entryID := c.Params("entry_id")
	userID, _ := c.Locals("user_id").(string)

	// Check ownership
	var existing struct {
		EmployeeID string `json:"employee_id"`
	}
	err := h.DB.From("time_entries").Select("employee_id").
		Eq("id", entryID).Single().Execute(&existing)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Time entry not found"})
	}

	if existing.EmployeeID != userID {
		return c.Status(403).JSON(fiber.Map{"error": "Can only edit your own time entries"})
	}

	var req struct {
		Hours       *float64 `json:"hours"`
		Description *string  `json:"description"`
		Date        *string  `json:"date"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	data := make(map[string]interface{})
	if req.Hours != nil {
		if *req.Hours <= 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Hours must be positive"})
		}
		data["hours"] = *req.Hours
	}
	if req.Description != nil {
		data["description"] = *req.Description
	}
	if req.Date != nil {
		data["date"] = *req.Date
	}

	if len(data) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No fields to update"})
	}

	_, err = h.DB.Update("time_entries", "id", entryID, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

// DeleteTimeEntry deletes a time entry
func (h *Handler) DeleteTimeEntry(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	entryID := c.Params("entry_id")
	userID, _ := c.Locals("user_id").(string)

	// Check ownership
	var existing struct {
		EmployeeID string `json:"employee_id"`
	}
	err := h.DB.From("time_entries").Select("employee_id").
		Eq("id", entryID).Single().Execute(&existing)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Time entry not found"})
	}

	if existing.EmployeeID != userID {
		return c.Status(403).JSON(fiber.Map{"error": "Can only delete your own time entries"})
	}

	if err := h.DB.Delete("time_entries", "id", entryID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

// GetMyTimeEntries returns time entries for the current user
func (h *Handler) GetMyTimeEntries(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID, _ := c.Locals("user_id").(string)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := h.DB.From("time_entries").Select("*, task:tasks(id, title, status)").
		Eq("employee_id", userID)

	if startDate != "" {
		query = query.Gte("date", startDate)
	}
	if endDate != "" {
		query = query.Lte("date", endDate)
	}

	var entries []struct {
		models.TimeEntry
		Task *struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"task,omitempty"`
	}
	err := query.Order("date", true).Execute(&entries)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Calculate totals
	var totalHours float64
	for _, e := range entries {
		totalHours += e.Hours
	}

	return c.JSON(fiber.Map{
		"entries":     entries,
		"total_hours": totalHours,
	})
}

// GetTaskResourceSummary returns resource summary for a task
func (h *Handler) GetTaskResourceSummary(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	taskID := c.Params("id")

	// Get task with resource fields
	var task struct {
		EstimatedHours *float64 `json:"estimated_hours"`
		ActualHours    *float64 `json:"actual_hours"`
		EstimatedCost  *float64 `json:"estimated_cost"`
		ActualCost     *float64 `json:"actual_cost"`
		AssigneeID     *string  `json:"assignee_id"`
	}
	err := h.DB.From("tasks").Select("estimated_hours, actual_hours, estimated_cost, actual_cost, assignee_id").
		Eq("id", taskID).Single().Execute(&task)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Task not found"})
	}

	// Get assignee hourly rate
	var hourlyRate float64
	if task.AssigneeID != nil {
		var assignee struct {
			HourlyRate *float64 `json:"hourly_rate"`
		}
		h.DB.From("employees").Select("hourly_rate").
			Eq("id", *task.AssigneeID).Single().Execute(&assignee)
		if assignee.HourlyRate != nil {
			hourlyRate = *assignee.HourlyRate
		}
	}

	// Get time entries summary
	var entries []struct {
		Hours float64 `json:"hours"`
	}
	h.DB.From("time_entries").Select("hours").Eq("task_id", taskID).Execute(&entries)

	var loggedHours float64
	for _, e := range entries {
		loggedHours += e.Hours
	}

	calculatedCost := loggedHours * hourlyRate

	return c.JSON(fiber.Map{
		"estimated_hours": task.EstimatedHours,
		"actual_hours":    task.ActualHours,
		"estimated_cost":  task.EstimatedCost,
		"actual_cost":     task.ActualCost,
		"logged_hours":    loggedHours,
		"hourly_rate":     hourlyRate,
		"calculated_cost": calculatedCost,
	})
}
