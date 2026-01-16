package handlers

import (
	"encoding/json"

	"github.com/ekf/one-on-one-backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// ListEmployees returns all employees
func (h *Handler) ListEmployees(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var employees []models.Employee
	err := h.DB.From("employees").Select("*").Order("name", false).Execute(&employees)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(employees)
}

// GetEmployee returns a single employee
func (h *Handler) GetEmployee(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var employee models.Employee
	err := h.DB.From("employees").Select("*").Eq("id", id).Single().Execute(&employee)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Employee not found"})
	}

	return c.JSON(employee)
}

// CreateEmployee creates a new employee
func (h *Handler) CreateEmployee(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var employee models.Employee
	if err := c.BodyParser(&employee); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.DB.Insert("employees", employee)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var created []models.Employee
	json.Unmarshal(result, &created)

	if len(created) > 0 {
		return c.Status(201).JSON(created[0])
	}

	return c.Status(201).JSON(fiber.Map{"status": "created"})
}

// UpdateEmployee updates an employee
func (h *Handler) UpdateEmployee(c *fiber.Ctx) error {
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

	result, err := h.DB.Update("employees", "id", id, updates)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var updated []models.Employee
	json.Unmarshal(result, &updated)

	if len(updated) > 0 {
		return c.JSON(updated[0])
	}

	return c.JSON(fiber.Map{"status": "updated"})
}

// DeleteEmployee deletes an employee
func (h *Handler) DeleteEmployee(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	if err := h.DB.Delete("employees", "id", id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

// GetEmployeeDossier returns full employee profile
func (h *Handler) GetEmployeeDossier(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	// Get employee
	var employee models.Employee
	err := h.DB.From("employees").Select("*").Eq("id", id).Single().Execute(&employee)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Employee not found"})
	}

	// Get 1-on-1 meetings
	var meetings []models.Meeting
	h.DB.From("meetings").Select("*").Eq("employee_id", id).Order("date", true).Execute(&meetings)

	// Get task stats
	var tasks []models.Task
	h.DB.From("tasks").Select("status").Eq("assignee_id", id).Execute(&tasks)

	tasksDone := 0
	tasksInProgress := 0
	for _, t := range tasks {
		if t.Status == "done" {
			tasksDone++
		} else if t.Status == "in_progress" {
			tasksInProgress++
		}
	}

	// Mood history
	var moodHistory []map[string]interface{}
	for _, m := range meetings {
		if m.MoodScore != nil {
			moodHistory = append(moodHistory, map[string]interface{}{
				"date":  m.Date,
				"score": *m.MoodScore,
			})
		}
	}

	// Red flags history
	var redFlags []map[string]interface{}
	for _, m := range meetings {
		if m.Analysis != nil {
			if flags, ok := m.Analysis["red_flags"].(map[string]interface{}); ok {
				burnout, _ := flags["burnout_signs"].(string)
				turnover, _ := flags["turnover_risk"].(string)
				if burnout != "" || (turnover != "low" && turnover != "") {
					redFlags = append(redFlags, map[string]interface{}{
						"date":  m.Date,
						"flags": flags,
					})
				}
			}
		}
	}

	// Get recent meetings safely
	var recentMeetings []models.Meeting
	if len(meetings) > 0 {
		recentMeetings = meetings[:min(5, len(meetings))]
	} else {
		recentMeetings = []models.Meeting{}
	}

	// Ensure arrays are not nil
	if moodHistory == nil {
		moodHistory = []map[string]interface{}{}
	}
	if redFlags == nil {
		redFlags = []map[string]interface{}{}
	}

	return c.JSON(fiber.Map{
		"employee":               employee,
		"one_on_one_count":       len(meetings),
		"project_meetings_count": 0,
		"tasks": fiber.Map{
			"total":       len(tasks),
			"done":        tasksDone,
			"in_progress": tasksInProgress,
		},
		"mood_history":      moodHistory,
		"red_flags_history": redFlags,
		"recent_meetings":   recentMeetings,
	})
}

// GetMyTeam returns subordinates of a manager
func (h *Handler) GetMyTeam(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	managerID := c.Query("manager_id")
	if managerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "manager_id required"})
	}

	includeIndirect := c.Query("include_indirect", "true") == "true"

	if includeIndirect {
		// Get all subordinates recursively (simplified - just direct for now)
		var subordinates []models.Employee
		h.DB.From("employees").Select("*").Eq("manager_id", managerID).Execute(&subordinates)
		return c.JSON(subordinates)
	}

	var subordinates []models.Employee
	h.DB.From("employees").Select("*").Eq("manager_id", managerID).Execute(&subordinates)
	return c.JSON(subordinates)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
