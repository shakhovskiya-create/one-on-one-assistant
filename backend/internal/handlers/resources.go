package handlers

import (
	"github.com/ekf/one-on-one-backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// ==================== Resource Allocations ====================

// ListResourceAllocations returns all allocations with optional filters
func (h *Handler) ListResourceAllocations(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	query := h.DB.From("resource_allocations").Select("*, employee:employees(id, name, position, photo_base64), task:tasks(id, title), project:projects(id, name)")

	// Filters
	if employeeID := c.Query("employee_id"); employeeID != "" {
		query = query.Eq("employee_id", employeeID)
	}
	if projectID := c.Query("project_id"); projectID != "" {
		query = query.Eq("project_id", projectID)
	}
	if taskID := c.Query("task_id"); taskID != "" {
		query = query.Eq("task_id", taskID)
	}

	var allocations []models.ResourceAllocation
	err := query.Order("period_start", true).Limit(100).Execute(&allocations)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(allocations)
}

// GetResourceAllocation returns a single allocation by ID
func (h *Handler) GetResourceAllocation(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var allocation models.ResourceAllocation
	err := h.DB.From("resource_allocations").
		Select("*, employee:employees(id, name, position, photo_base64), task:tasks(id, title), project:projects(id, name)").
		Eq("id", id).Single().Execute(&allocation)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Allocation not found"})
	}

	return c.JSON(allocation)
}

// CreateResourceAllocation creates a new allocation
func (h *Handler) CreateResourceAllocation(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var input struct {
		EmployeeID            string  `json:"employee_id"`
		TaskID                *string `json:"task_id"`
		ProjectID             *string `json:"project_id"`
		Role                  *string `json:"role"`
		AllocatedHoursPerWeek int     `json:"allocated_hours_per_week"`
		PeriodStart           string  `json:"period_start"`
		PeriodEnd             *string `json:"period_end"`
		Notes                 *string `json:"notes"`
		CreatedBy             *string `json:"created_by"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.EmployeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "employee_id is required"})
	}
	if input.PeriodStart == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "period_start is required"})
	}

	payload := map[string]interface{}{
		"employee_id":              input.EmployeeID,
		"allocated_hours_per_week": input.AllocatedHoursPerWeek,
		"period_start":             input.PeriodStart,
	}
	if input.TaskID != nil {
		payload["task_id"] = *input.TaskID
	}
	if input.ProjectID != nil {
		payload["project_id"] = *input.ProjectID
	}
	if input.Role != nil {
		payload["role"] = *input.Role
	}
	if input.PeriodEnd != nil {
		payload["period_end"] = *input.PeriodEnd
	}
	if input.Notes != nil {
		payload["notes"] = *input.Notes
	}
	if input.CreatedBy != nil {
		payload["created_by"] = *input.CreatedBy
	}

	result, err := h.DB.Insert("resource_allocations", payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdateResourceAllocation updates an existing allocation
func (h *Handler) UpdateResourceAllocation(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var input map[string]interface{}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Remove id from input if present
	delete(input, "id")

	result, err := h.DB.Update("resource_allocations", "id", id, input)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// DeleteResourceAllocation removes an allocation
func (h *Handler) DeleteResourceAllocation(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	if err := h.DB.Delete("resource_allocations", "id", id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Allocation deleted"})
}

// ==================== Employee Capacity ====================

// GetResourceCapacity returns capacity info for all employees
func (h *Handler) GetResourceCapacity(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	projectID := c.Query("project_id")

	// Get all employees with resource planning fields
	var employees []struct {
		ID                  string `json:"id"`
		Name                string `json:"name"`
		Position            string `json:"position"`
		WorkHoursPerWeek    *int   `json:"work_hours_per_week"`
		AvailabilityPercent *int   `json:"availability_percent"`
	}
	err := h.DB.From("employees").Select("id, name, position, work_hours_per_week, availability_percent").Order("name", false).Execute(&employees)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Get allocations
	allocQuery := h.DB.From("resource_allocations").Select("employee_id, allocated_hours_per_week")
	if projectID != "" {
		allocQuery = allocQuery.Eq("project_id", projectID)
	}

	var allocations []struct {
		EmployeeID            string `json:"employee_id"`
		AllocatedHoursPerWeek int    `json:"allocated_hours_per_week"`
	}
	allocQuery.Execute(&allocations)

	// Build allocation map by employee
	allocByEmployee := make(map[string]int)
	for _, alloc := range allocations {
		allocByEmployee[alloc.EmployeeID] += alloc.AllocatedHoursPerWeek
	}

	// Calculate capacity for each employee
	capacities := make([]models.ResourceCapacity, 0, len(employees))
	for _, emp := range employees {
		weeklyHours := 40
		if emp.WorkHoursPerWeek != nil {
			weeklyHours = *emp.WorkHoursPerWeek
		}
		availabilityPct := 100
		if emp.AvailabilityPercent != nil {
			availabilityPct = *emp.AvailabilityPercent
		}

		availableHours := float64(weeklyHours) * float64(availabilityPct) / 100.0
		allocatedHours := float64(allocByEmployee[emp.ID])
		freeHours := availableHours - allocatedHours

		var utilizationPct float64
		if availableHours > 0 {
			utilizationPct = (allocatedHours / availableHours) * 100
		}

		capacities = append(capacities, models.ResourceCapacity{
			EmployeeID:         emp.ID,
			EmployeeName:       emp.Name,
			Position:           emp.Position,
			WeeklyHours:        weeklyHours,
			AvailabilityPct:    availabilityPct,
			AvailableHours:     availableHours,
			AllocatedHours:     allocatedHours,
			FreeHours:          freeHours,
			UtilizationPercent: utilizationPct,
			Overloaded:         utilizationPct > 100,
		})
	}

	return c.JSON(capacities)
}

// GetResourceStats returns aggregated resource statistics
func (h *Handler) GetResourceStats(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	projectID := c.Query("project_id")

	// Get employees
	var employees []struct {
		ID                  string `json:"id"`
		WorkHoursPerWeek    *int   `json:"work_hours_per_week"`
		AvailabilityPercent *int   `json:"availability_percent"`
	}
	h.DB.From("employees").Select("id, work_hours_per_week, availability_percent").Execute(&employees)

	// Get allocations
	allocQuery := h.DB.From("resource_allocations").Select("employee_id, allocated_hours_per_week")
	if projectID != "" {
		allocQuery = allocQuery.Eq("project_id", projectID)
	}

	var allocations []struct {
		EmployeeID            string `json:"employee_id"`
		AllocatedHoursPerWeek int    `json:"allocated_hours_per_week"`
	}
	allocQuery.Execute(&allocations)

	// Build allocation map
	allocByEmployee := make(map[string]int)
	for _, alloc := range allocations {
		allocByEmployee[alloc.EmployeeID] += alloc.AllocatedHoursPerWeek
	}

	// Calculate stats
	stats := models.ResourceAllocationStats{
		TotalEmployees:   len(employees),
		TotalAllocations: len(allocations),
	}

	var totalUtilization float64
	employeesWithAlloc := 0

	for _, emp := range employees {
		weeklyHours := 40
		if emp.WorkHoursPerWeek != nil {
			weeklyHours = *emp.WorkHoursPerWeek
		}
		availabilityPct := 100
		if emp.AvailabilityPercent != nil {
			availabilityPct = *emp.AvailabilityPercent
		}

		availableHours := float64(weeklyHours) * float64(availabilityPct) / 100.0
		allocatedHours := float64(allocByEmployee[emp.ID])

		if allocatedHours > 0 {
			employeesWithAlloc++
			utilizationPct := (allocatedHours / availableHours) * 100
			totalUtilization += utilizationPct

			if utilizationPct > 100 {
				stats.OverloadedCount++
			} else if utilizationPct < 50 {
				stats.UnderutilizedCnt++
			}
		}
	}

	if employeesWithAlloc > 0 {
		stats.AvgUtilization = totalUtilization / float64(employeesWithAlloc)
	}

	return c.JSON(stats)
}

// ==================== Employee Absences ====================

// ListEmployeeAbsences returns absences for an employee
func (h *Handler) ListEmployeeAbsences(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	query := h.DB.From("employee_absences").Select("*, employee:employees(id, name, photo_base64)")

	if employeeID := c.Query("employee_id"); employeeID != "" {
		query = query.Eq("employee_id", employeeID)
	}

	var absences []models.EmployeeAbsence
	err := query.Order("start_date", true).Limit(100).Execute(&absences)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(absences)
}

// CreateEmployeeAbsence creates a new absence record
func (h *Handler) CreateEmployeeAbsence(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var input struct {
		EmployeeID  string  `json:"employee_id"`
		AbsenceType string  `json:"absence_type"`
		StartDate   string  `json:"start_date"`
		EndDate     string  `json:"end_date"`
		Description *string `json:"description"`
		Source      string  `json:"source"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.EmployeeID == "" || input.AbsenceType == "" || input.StartDate == "" || input.EndDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "employee_id, absence_type, start_date, end_date are required"})
	}

	if input.Source == "" {
		input.Source = "manual"
	}

	payload := map[string]interface{}{
		"employee_id":  input.EmployeeID,
		"absence_type": input.AbsenceType,
		"start_date":   input.StartDate,
		"end_date":     input.EndDate,
		"source":       input.Source,
	}
	if input.Description != nil {
		payload["description"] = *input.Description
	}

	result, err := h.DB.Insert("employee_absences", payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// DeleteEmployeeAbsence removes an absence record
func (h *Handler) DeleteEmployeeAbsence(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	if err := h.DB.Delete("employee_absences", "id", id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Absence deleted"})
}

// ==================== Employee Resource Settings ====================

// UpdateEmployeeResourceSettings updates resource planning fields for an employee
func (h *Handler) UpdateEmployeeResourceSettings(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var input struct {
		WorkHoursPerWeek    *int     `json:"work_hours_per_week"`
		AvailabilityPercent *int     `json:"availability_percent"`
		HourlyRate          *float64 `json:"hourly_rate"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	payload := make(map[string]interface{})
	if input.WorkHoursPerWeek != nil {
		payload["work_hours_per_week"] = *input.WorkHoursPerWeek
	}
	if input.AvailabilityPercent != nil {
		payload["availability_percent"] = *input.AvailabilityPercent
	}
	if input.HourlyRate != nil {
		payload["hourly_rate"] = *input.HourlyRate
	}

	if len(payload) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No fields to update"})
	}

	result, err := h.DB.Update("employees", "id", id, payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}
