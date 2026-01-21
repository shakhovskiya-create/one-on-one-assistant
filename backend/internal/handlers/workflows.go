package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// WorkflowMode represents a workflow configuration
type WorkflowMode struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Statuses    interface{} `json:"statuses"`
	IsDefault   bool        `json:"is_default"`
}

// StatusColumn represents a single status in workflow
type StatusColumn struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Color    string `json:"color"`
	WIPLimit int    `json:"wipLimit"`
}

// GetWorkflowForUser returns workflow based on user's department
func (h *Handler) GetWorkflowForUser(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Not authenticated"})
	}

	// Get user's department
	var employee struct {
		Department string `json:"department"`
	}
	err := h.DB.From("employees").Select("department").Eq("id", userID).Single().Execute(&employee)
	if err != nil {
		// Return default workflow
		return h.getDefaultWorkflow(c)
	}

	// Check if department has custom workflow
	var deptWorkflow struct {
		WorkflowModeID string `json:"workflow_mode_id"`
	}
	err = h.DB.From("department_workflows").Select("workflow_mode_id").
		Eq("department", employee.Department).Single().Execute(&deptWorkflow)

	if err != nil || deptWorkflow.WorkflowModeID == "" {
		// Return default workflow
		return h.getDefaultWorkflow(c)
	}

	// Get the workflow mode
	var workflow WorkflowMode
	err = h.DB.From("workflow_modes").Select("*").
		Eq("id", deptWorkflow.WorkflowModeID).Single().Execute(&workflow)

	if err != nil {
		return h.getDefaultWorkflow(c)
	}

	return c.JSON(fiber.Map{
		"workflow":   workflow,
		"department": employee.Department,
	})
}

// GetWorkflowByDepartment returns workflow for specific department (admin use)
func (h *Handler) GetWorkflowByDepartment(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	department := c.Query("department")
	if department == "" {
		return h.getDefaultWorkflow(c)
	}

	// Check if department has custom workflow
	var deptWorkflow struct {
		WorkflowModeID string `json:"workflow_mode_id"`
	}
	err := h.DB.From("department_workflows").Select("workflow_mode_id").
		Eq("department", department).Single().Execute(&deptWorkflow)

	if err != nil || deptWorkflow.WorkflowModeID == "" {
		return h.getDefaultWorkflow(c)
	}

	var workflow WorkflowMode
	err = h.DB.From("workflow_modes").Select("*").
		Eq("id", deptWorkflow.WorkflowModeID).Single().Execute(&workflow)

	if err != nil {
		return h.getDefaultWorkflow(c)
	}

	return c.JSON(fiber.Map{
		"workflow":   workflow,
		"department": department,
	})
}

// ListWorkflowModes returns all available workflow modes
func (h *Handler) ListWorkflowModes(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var modes []WorkflowMode
	err := h.DB.From("workflow_modes").Select("*").Execute(&modes)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(modes)
}

// SetDepartmentWorkflow sets workflow mode for a department (admin only)
func (h *Handler) SetDepartmentWorkflow(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var req struct {
		Department     string `json:"department"`
		WorkflowModeID string `json:"workflow_mode_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Department == "" || req.WorkflowModeID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Department and workflow_mode_id required"})
	}

	// Check if exists, update or insert
	var existing []struct{ ID string }
	h.DB.From("department_workflows").Select("id").
		Eq("department", req.Department).Execute(&existing)

	if len(existing) > 0 {
		// Update
		_, err := h.DB.Update("department_workflows", "id", existing[0].ID, map[string]interface{}{
			"workflow_mode_id": req.WorkflowModeID,
		})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	} else {
		// Insert
		_, err := h.DB.Insert("department_workflows", map[string]interface{}{
			"department":       req.Department,
			"workflow_mode_id": req.WorkflowModeID,
		})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(fiber.Map{"success": true})
}

// getDefaultWorkflow returns the default simple workflow
func (h *Handler) getDefaultWorkflow(c *fiber.Ctx) error {
	var workflow WorkflowMode
	err := h.DB.From("workflow_modes").Select("*").
		Eq("is_default", "true").Single().Execute(&workflow)

	if err != nil {
		// Fallback hardcoded
		return c.JSON(fiber.Map{
			"workflow": fiber.Map{
				"name":        "simple",
				"description": "Простой Kanban",
				"statuses": []StatusColumn{
					{ID: "backlog", Label: "Backlog", Color: "bg-gray-100", WIPLimit: 0},
					{ID: "todo", Label: "К выполнению", Color: "bg-blue-50", WIPLimit: 10},
					{ID: "in_progress", Label: "В работе", Color: "bg-yellow-50", WIPLimit: 5},
					{ID: "done", Label: "Готово", Color: "bg-green-50", WIPLimit: 0},
				},
			},
			"department": "",
		})
	}

	return c.JSON(fiber.Map{
		"workflow":   workflow,
		"department": "",
	})
}

// ListDepartmentWorkflows returns all department workflow configurations
func (h *Handler) ListDepartmentWorkflows(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var configs []struct {
		ID             string        `json:"id"`
		Department     string        `json:"department"`
		WorkflowModeID string        `json:"workflow_mode_id"`
		WorkflowMode   *WorkflowMode `json:"workflow_mode"`
	}

	err := h.DB.From("department_workflows").Select("*, workflow_mode:workflow_modes(*)").Execute(&configs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(configs)
}
