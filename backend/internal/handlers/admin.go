package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// AdminStats represents dashboard statistics
type AdminStats struct {
	TotalUsers       int `json:"total_users"`
	ActiveUsers      int `json:"active_users"`
	TotalTasks       int `json:"total_tasks"`
	CompletedTasks   int `json:"completed_tasks"`
	TotalMeetings    int `json:"total_meetings"`
	TotalMessages    int `json:"total_messages"`
	AdminCount       int `json:"admin_count"`
	DepartmentsCount int `json:"departments_count"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         string                 `json:"id"`
	UserID     *string                `json:"user_id"`
	Action     string                 `json:"action"`
	EntityType *string                `json:"entity_type"`
	EntityID   *string                `json:"entity_id"`
	OldValue   map[string]interface{} `json:"old_value"`
	NewValue   map[string]interface{} `json:"new_value"`
	IPAddress  *string                `json:"ip_address"`
	UserAgent  *string                `json:"user_agent"`
	CreatedAt  time.Time              `json:"created_at"`
	User       *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"user,omitempty"`
}

// SystemSetting represents a system setting
type SystemSetting struct {
	ID          string                 `json:"id"`
	Key         string                 `json:"key"`
	Value       interface{}            `json:"value"`
	Description *string                `json:"description"`
	UpdatedBy   *string                `json:"updated_by"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// GetAdminStats returns dashboard statistics for admins
func (h *Handler) GetAdminStats(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	stats := AdminStats{}

	// Count total users
	var users []struct{ ID string }
	h.DB.From("employees").Select("id").Execute(&users)
	stats.TotalUsers = len(users)

	// Count admins
	var admins []struct{ ID string }
	h.DB.From("employees").Select("id").Eq("role", "admin").Execute(&admins)
	var superAdmins []struct{ ID string }
	h.DB.From("employees").Select("id").Eq("role", "super_admin").Execute(&superAdmins)
	stats.AdminCount = len(admins) + len(superAdmins)

	// Count tasks
	var tasks []struct{ ID string }
	h.DB.From("tasks").Select("id").Execute(&tasks)
	stats.TotalTasks = len(tasks)

	var completedTasks []struct{ ID string }
	h.DB.From("tasks").Select("id").Eq("status", "done").Execute(&completedTasks)
	stats.CompletedTasks = len(completedTasks)

	// Count meetings
	var meetings []struct{ ID string }
	h.DB.From("meetings").Select("id").Execute(&meetings)
	stats.TotalMeetings = len(meetings)

	// Count messages
	var messages []struct{ ID string }
	h.DB.From("messages").Select("id").Execute(&messages)
	stats.TotalMessages = len(messages)

	// Count unique departments
	var depts []struct{ Department string }
	h.DB.From("employees").Select("department").Execute(&depts)
	deptMap := make(map[string]bool)
	for _, d := range depts {
		if d.Department != "" {
			deptMap[d.Department] = true
		}
	}
	stats.DepartmentsCount = len(deptMap)

	return c.JSON(stats)
}

// ListUsers returns all users with their roles
func (h *Handler) ListUsers(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var users []struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		Email      string  `json:"email"`
		Department *string `json:"department"`
		Position   string  `json:"position"`
		Role       *string `json:"role"`
		CreatedAt  *string `json:"created_at"`
	}

	err := h.DB.From("employees").Select("id, name, email, department, position, role, created_at").
		Order("name", false).Execute(&users)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Set default role for users without role
	for i := range users {
		if users[i].Role == nil {
			role := "user"
			users[i].Role = &role
		}
	}

	return c.JSON(users)
}

// UpdateUserRole updates a user's role
func (h *Handler) UpdateUserRole(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID := c.Params("id")
	currentUserID, _ := c.Locals("user_id").(string)
	currentUserRole, _ := c.Locals("user_role").(string)

	var req struct {
		Role string `json:"role"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Validate role
	validRoles := map[string]bool{"user": true, "admin": true, "super_admin": true}
	if !validRoles[req.Role] {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role. Must be: user, admin, or super_admin"})
	}

	// Only super_admin can create other super_admins
	if req.Role == "super_admin" && currentUserRole != "super_admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Only super_admin can create other super_admins"})
	}

	// Get old value for audit
	var oldUser struct {
		Role *string `json:"role"`
	}
	h.DB.From("employees").Select("role").Eq("id", userID).Single().Execute(&oldUser)

	// Update role
	_, err := h.DB.Update("employees", "id", userID, map[string]interface{}{
		"role": req.Role,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Create audit log
	oldRole := "user"
	if oldUser.Role != nil {
		oldRole = *oldUser.Role
	}
	h.createAuditLog(c, currentUserID, "user.role.update", "employee", userID,
		map[string]interface{}{"role": oldRole},
		map[string]interface{}{"role": req.Role})

	return c.JSON(fiber.Map{"success": true})
}

// GetSystemSettings returns all system settings
func (h *Handler) GetSystemSettings(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var settings []SystemSetting
	err := h.DB.From("system_settings").Select("*").Execute(&settings)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(settings)
}

// UpdateSystemSetting updates a system setting
func (h *Handler) UpdateSystemSetting(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID, _ := c.Locals("user_id").(string)

	var req struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Key == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Key is required"})
	}

	// Get old value for audit
	var oldSetting struct {
		Value interface{} `json:"value"`
	}
	h.DB.From("system_settings").Select("value").Eq("key", req.Key).Single().Execute(&oldSetting)

	// Update setting
	_, err := h.DB.Update("system_settings", "key", req.Key, map[string]interface{}{
		"value":      req.Value,
		"updated_by": userID,
		"updated_at": time.Now(),
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Create audit log
	h.createAuditLog(c, userID, "setting.update", "system_setting", req.Key,
		map[string]interface{}{"value": oldSetting.Value},
		map[string]interface{}{"value": req.Value})

	return c.JSON(fiber.Map{"success": true})
}

// GetAuditLogs returns audit logs with pagination
func (h *Handler) GetAuditLogs(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	action := c.Query("action")
	entityType := c.Query("entity_type")

	query := h.DB.From("audit_logs").Select("*, user:employees(id, name)")

	if action != "" {
		query = query.Ilike("action", action+"%")
	}
	if entityType != "" {
		query = query.Eq("entity_type", entityType)
	}

	var logs []AuditLog
	err := query.Order("created_at", true).Limit(limit).Offset(offset).Execute(&logs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(logs)
}

// GetDepartments returns unique departments list
func (h *Handler) GetDepartments(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var employees []struct {
		Department string `json:"department"`
	}
	err := h.DB.From("employees").Select("department").Execute(&employees)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Get unique departments with counts
	deptCounts := make(map[string]int)
	for _, e := range employees {
		if e.Department != "" {
			deptCounts[e.Department]++
		}
	}

	// Get workflow assignments
	var workflows []struct {
		Department     string `json:"department"`
		WorkflowModeID string `json:"workflow_mode_id"`
	}
	h.DB.From("department_workflows").Select("department, workflow_mode_id").Execute(&workflows)

	workflowMap := make(map[string]string)
	for _, w := range workflows {
		workflowMap[w.Department] = w.WorkflowModeID
	}

	// Build result
	type DeptInfo struct {
		Name           string  `json:"name"`
		EmployeeCount  int     `json:"employee_count"`
		WorkflowModeID *string `json:"workflow_mode_id"`
	}

	var result []DeptInfo
	for dept, count := range deptCounts {
		info := DeptInfo{
			Name:          dept,
			EmployeeCount: count,
		}
		if wfID, ok := workflowMap[dept]; ok {
			info.WorkflowModeID = &wfID
		}
		result = append(result, info)
	}

	return c.JSON(result)
}

// createAuditLog helper to create audit log entry
func (h *Handler) createAuditLog(c *fiber.Ctx, userID, action, entityType, entityID string, oldValue, newValue map[string]interface{}) {
	if h.DB == nil {
		return
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	h.DB.Insert("audit_logs", map[string]interface{}{
		"user_id":     userID,
		"action":      action,
		"entity_type": entityType,
		"entity_id":   entityID,
		"old_value":   oldValue,
		"new_value":   newValue,
		"ip_address":  ipAddress,
		"user_agent":  userAgent,
	})
}

// GetCurrentUserRole returns the role of the current user
func (h *Handler) GetCurrentUserRole(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var employee struct {
		Role *string `json:"role"`
	}
	err := h.DB.From("employees").Select("role").Eq("id", userID).Single().Execute(&employee)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	role := "user"
	if employee.Role != nil {
		role = *employee.Role
	}

	return c.JSON(fiber.Map{"role": role})
}
