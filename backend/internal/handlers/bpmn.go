package handlers

import (
	"github.com/ekf/one-on-one-backend/pkg/camunda"
	"github.com/gofiber/fiber/v2"
)

// BPMNStatus returns BPMN/Camunda integration status
func (h *Handler) BPMNStatus(c *fiber.Ctx) error {
	if h.Camunda == nil || !h.Camunda.IsConfigured() {
		return c.JSON(fiber.Map{
			"configured": false,
			"status":     "not_configured",
			"message":    "Camunda BPMN не настроен. Укажите CAMUNDA_URL в переменных окружения.",
		})
	}

	err := h.Camunda.HealthCheck()
	if err != nil {
		return c.JSON(fiber.Map{
			"configured": true,
			"status":     "unavailable",
			"error":      err.Error(),
			"url":        h.Config.CamundaURL,
		})
	}

	return c.JSON(fiber.Map{
		"configured": true,
		"status":     "connected",
		"url":        h.Config.CamundaURL,
	})
}

// ListProcessDefinitions returns all available process definitions
func (h *Handler) ListProcessDefinitions(c *fiber.Ctx) error {
	if h.Camunda == nil || !h.Camunda.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Camunda не настроен"})
	}

	definitions, err := h.Camunda.GetProcessDefinitions()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(definitions)
}

// GetProcessDefinition returns a process definition by key
func (h *Handler) GetProcessDefinition(c *fiber.Ctx) error {
	if h.Camunda == nil || !h.Camunda.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Camunda не настроен"})
	}

	key := c.Params("key")
	definition, err := h.Camunda.GetProcessDefinitionByKey(key)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(definition)
}

// StartProcessRequest represents a request to start a process
type StartProcessRequest struct {
	ProcessKey  string                 `json:"process_key"`
	BusinessKey string                 `json:"business_key,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
}

// StartProcess starts a new process instance
func (h *Handler) StartProcess(c *fiber.Ctx) error {
	if h.Camunda == nil || !h.Camunda.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Camunda не настроен"})
	}

	var req StartProcessRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Неверный формат запроса"})
	}

	if req.ProcessKey == "" {
		return c.Status(400).JSON(fiber.Map{"error": "process_key обязателен"})
	}

	// Convert variables to Camunda format
	camundaVars := make(map[string]camunda.Variable)
	for k, v := range req.Variables {
		camundaVars[k] = camunda.Variable{Value: v}
	}

	instance, err := h.Camunda.StartProcess(req.ProcessKey, camunda.StartProcessRequest{
		BusinessKey: req.BusinessKey,
		Variables:   camundaVars,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Save process instance to database for tracking
	if h.DB != nil {
		h.DB.Insert("bpmn_instances", map[string]interface{}{
			"camunda_id":     instance.ID,
			"definition_id":  instance.DefinitionID,
			"business_key":   req.BusinessKey,
			"process_key":    req.ProcessKey,
			"status":         "active",
			"variables":      req.Variables,
		})
	}

	return c.JSON(fiber.Map{
		"id":            instance.ID,
		"definition_id": instance.DefinitionID,
		"business_key":  instance.BusinessKey,
		"status":        "started",
	})
}

// ListProcessInstances returns active process instances
func (h *Handler) ListProcessInstances(c *fiber.Ctx) error {
	if h.Camunda == nil || !h.Camunda.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Camunda не настроен"})
	}

	processKey := c.Query("process_key")
	activeOnly := c.QueryBool("active", true)

	instances, err := h.Camunda.GetProcessInstances(processKey, activeOnly)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(instances)
}

// GetProcessInstance returns a process instance by ID
func (h *Handler) GetProcessInstance(c *fiber.Ctx) error {
	if h.Camunda == nil || !h.Camunda.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Camunda не настроен"})
	}

	id := c.Params("id")
	instance, err := h.Camunda.GetProcessInstance(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Get variables
	variables, _ := h.Camunda.GetProcessVariables(id)

	return c.JSON(fiber.Map{
		"instance":  instance,
		"variables": variables,
	})
}

// DeleteProcessInstance cancels a process instance
func (h *Handler) DeleteProcessInstance(c *fiber.Ctx) error {
	if h.Camunda == nil || !h.Camunda.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Camunda не настроен"})
	}

	id := c.Params("id")
	err := h.Camunda.DeleteProcessInstance(id, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Update status in database
	if h.DB != nil {
		h.DB.Update("bpmn_instances", "camunda_id", id, map[string]string{"status": "cancelled"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Процесс отменён"})
}

// ListTasks returns user tasks
func (h *Handler) ListBPMNTasks(c *fiber.Ctx) error {
	if h.Camunda == nil || !h.Camunda.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Camunda не настроен"})
	}

	assignee := c.Query("assignee")
	processInstanceID := c.Query("process_instance_id")

	tasks, err := h.Camunda.GetTasks(assignee, processInstanceID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tasks)
}

// GetBPMNTask returns a task by ID
func (h *Handler) GetBPMNTask(c *fiber.Ctx) error {
	if h.Camunda == nil || !h.Camunda.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Camunda не настроен"})
	}

	id := c.Params("id")
	task, err := h.Camunda.GetTask(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(task)
}

// CompleteTaskRequest represents a request to complete a task
type CompleteTaskRequest struct {
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// CompleteBPMNTask completes a user task
func (h *Handler) CompleteBPMNTask(c *fiber.Ctx) error {
	if h.Camunda == nil || !h.Camunda.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Camunda не настроен"})
	}

	id := c.Params("id")

	var req CompleteTaskRequest
	c.BodyParser(&req)

	// Convert variables
	camundaVars := make(map[string]camunda.Variable)
	for k, v := range req.Variables {
		camundaVars[k] = camunda.Variable{Value: v}
	}

	err := h.Camunda.CompleteTask(id, camundaVars)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Задача выполнена"})
}

// ClaimTaskRequest represents a request to claim a task
type ClaimTaskRequest struct {
	UserID string `json:"user_id"`
}

// ClaimBPMNTask assigns a task to a user
func (h *Handler) ClaimBPMNTask(c *fiber.Ctx) error {
	if h.Camunda == nil || !h.Camunda.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Camunda не настроен"})
	}

	id := c.Params("id")

	var req ClaimTaskRequest
	if err := c.BodyParser(&req); err != nil || req.UserID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "user_id обязателен"})
	}

	err := h.Camunda.ClaimTask(id, req.UserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Задача назначена"})
}

// UnclaimBPMNTask removes assignee from a task
func (h *Handler) UnclaimBPMNTask(c *fiber.Ctx) error {
	if h.Camunda == nil || !h.Camunda.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Camunda не настроен"})
	}

	id := c.Params("id")
	err := h.Camunda.UnclaimTask(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Назначение снято"})
}
