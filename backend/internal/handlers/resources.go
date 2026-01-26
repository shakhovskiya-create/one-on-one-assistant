package handlers

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/ekf/one-on-one-backend/internal/models"
)

// ==================== Resource Allocations ====================

// ListResourceAllocations returns all allocations with optional filters
func (h *Handler) ListResourceAllocations(c *fiber.Ctx) error {
	qb := NewQueryBuilder("resource_allocations")

	// Filters
	if employeeID := c.Query("employee_id"); employeeID != "" {
		qb.Filter("employee_id", "eq", employeeID)
	}
	if projectID := c.Query("project_id"); projectID != "" {
		qb.Filter("project_id", "eq", projectID)
	}
	if taskID := c.Query("task_id"); taskID != "" {
		qb.Filter("task_id", "eq", taskID)
	}

	// Date filters
	if startFrom := c.Query("start_from"); startFrom != "" {
		qb.Filter("period_start", "gte", startFrom)
	}
	if endTo := c.Query("end_to"); endTo != "" {
		qb.Filter("period_end", "lte", endTo)
	}

	qb.Order("period_start", false)
	qb.Limit(c.QueryInt("limit", 100))
	qb.Offset(c.QueryInt("offset", 0))

	var allocations []models.ResourceAllocation
	statusCode, err := h.PostgrestRequest("GET", qb.Build(), nil, &allocations)
	if err != nil || statusCode >= 400 {
		return c.Status(statusCode).JSON(fiber.Map{"error": "Failed to fetch allocations"})
	}

	return c.JSON(allocations)
}

// GetResourceAllocation returns a single allocation by ID
func (h *Handler) GetResourceAllocation(c *fiber.Ctx) error {
	id := c.Params("id")
	qb := NewQueryBuilder("resource_allocations")
	qb.Filter("id", "eq", id)

	var allocations []models.ResourceAllocation
	statusCode, err := h.PostgrestRequest("GET", qb.Build(), nil, &allocations)
	if err != nil || statusCode >= 400 || len(allocations) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Allocation not found"})
	}

	return c.JSON(allocations[0])
}

// CreateResourceAllocation creates a new allocation
func (h *Handler) CreateResourceAllocation(c *fiber.Ctx) error {
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

	body, _ := json.Marshal(payload)
	var result []models.ResourceAllocation
	statusCode, err := h.PostgrestRequest("POST", "resource_allocations", body, &result)
	if err != nil || statusCode >= 400 {
		log.Printf("Error creating allocation: %v, status: %d", err, statusCode)
		return c.Status(statusCode).JSON(fiber.Map{"error": "Failed to create allocation"})
	}

	if len(result) > 0 {
		return c.Status(fiber.StatusCreated).JSON(result[0])
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Allocation created"})
}

// UpdateResourceAllocation updates an existing allocation
func (h *Handler) UpdateResourceAllocation(c *fiber.Ctx) error {
	id := c.Params("id")

	var input map[string]interface{}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Remove id from input if present
	delete(input, "id")

	body, _ := json.Marshal(input)
	qb := NewQueryBuilder("resource_allocations")
	qb.Filter("id", "eq", id)

	var result []models.ResourceAllocation
	statusCode, err := h.PostgrestRequest("PATCH", qb.Build(), body, &result)
	if err != nil || statusCode >= 400 {
		return c.Status(statusCode).JSON(fiber.Map{"error": "Failed to update allocation"})
	}

	if len(result) > 0 {
		return c.JSON(result[0])
	}
	return c.JSON(fiber.Map{"message": "Allocation updated"})
}

// DeleteResourceAllocation removes an allocation
func (h *Handler) DeleteResourceAllocation(c *fiber.Ctx) error {
	id := c.Params("id")
	qb := NewQueryBuilder("resource_allocations")
	qb.Filter("id", "eq", id)

	statusCode, err := h.PostgrestRequest("DELETE", qb.Build(), nil, nil)
	if err != nil || statusCode >= 400 {
		return c.Status(statusCode).JSON(fiber.Map{"error": "Failed to delete allocation"})
	}

	return c.JSON(fiber.Map{"message": "Allocation deleted"})
}

// ==================== Employee Capacity ====================

// GetResourceCapacity returns capacity info for all employees
func (h *Handler) GetResourceCapacity(c *fiber.Ctx) error {
	// Get optional filters
	projectID := c.Query("project_id")
	periodStart := c.Query("period_start")
	periodEnd := c.Query("period_end")

	// Get all employees with resource planning fields
	qb := NewQueryBuilder("employees")
	qb.Select("id,name,position,work_hours_per_week,availability_percent")
	qb.Order("name", true)

	var employees []struct {
		ID                  string `json:"id"`
		Name                string `json:"name"`
		Position            string `json:"position"`
		WorkHoursPerWeek    *int   `json:"work_hours_per_week"`
		AvailabilityPercent *int   `json:"availability_percent"`
	}
	statusCode, err := h.PostgrestRequest("GET", qb.Build(), nil, &employees)
	if err != nil || statusCode >= 400 {
		return c.Status(statusCode).JSON(fiber.Map{"error": "Failed to fetch employees"})
	}

	// Get allocations for these employees
	allocQb := NewQueryBuilder("resource_allocations")
	if projectID != "" {
		allocQb.Filter("project_id", "eq", projectID)
	}
	if periodStart != "" {
		allocQb.Filter("period_start", "gte", periodStart)
	}
	if periodEnd != "" {
		allocQb.Filter("period_end", "lte", periodEnd)
	}

	var allocations []models.ResourceAllocation
	statusCode, err = h.PostgrestRequest("GET", allocQb.Build(), nil, &allocations)
	if err != nil || statusCode >= 400 {
		allocations = []models.ResourceAllocation{} // Continue without allocations
	}

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
	// Get capacity data first
	projectID := c.Query("project_id")

	// Get employees
	qb := NewQueryBuilder("employees")
	qb.Select("id,work_hours_per_week,availability_percent")
	var employees []struct {
		ID                  string `json:"id"`
		WorkHoursPerWeek    *int   `json:"work_hours_per_week"`
		AvailabilityPercent *int   `json:"availability_percent"`
	}
	h.PostgrestRequest("GET", qb.Build(), nil, &employees)

	// Get allocations
	allocQb := NewQueryBuilder("resource_allocations")
	if projectID != "" {
		allocQb.Filter("project_id", "eq", projectID)
	}
	var allocations []models.ResourceAllocation
	h.PostgrestRequest("GET", allocQb.Build(), nil, &allocations)

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
	employeeID := c.Query("employee_id")

	qb := NewQueryBuilder("employee_absences")
	if employeeID != "" {
		qb.Filter("employee_id", "eq", employeeID)
	}

	// Date filters
	if startFrom := c.Query("start_from"); startFrom != "" {
		qb.Filter("start_date", "gte", startFrom)
	}
	if endTo := c.Query("end_to"); endTo != "" {
		qb.Filter("end_date", "lte", endTo)
	}

	qb.Order("start_date", false)
	qb.Limit(c.QueryInt("limit", 100))

	var absences []models.EmployeeAbsence
	statusCode, err := h.PostgrestRequest("GET", qb.Build(), nil, &absences)
	if err != nil || statusCode >= 400 {
		return c.Status(statusCode).JSON(fiber.Map{"error": "Failed to fetch absences"})
	}

	return c.JSON(absences)
}

// CreateEmployeeAbsence creates a new absence record
func (h *Handler) CreateEmployeeAbsence(c *fiber.Ctx) error {
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

	body, _ := json.Marshal(payload)
	var result []models.EmployeeAbsence
	statusCode, err := h.PostgrestRequest("POST", "employee_absences", body, &result)
	if err != nil || statusCode >= 400 {
		return c.Status(statusCode).JSON(fiber.Map{"error": "Failed to create absence"})
	}

	if len(result) > 0 {
		return c.Status(fiber.StatusCreated).JSON(result[0])
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Absence created"})
}

// DeleteEmployeeAbsence removes an absence record
func (h *Handler) DeleteEmployeeAbsence(c *fiber.Ctx) error {
	id := c.Params("id")
	qb := NewQueryBuilder("employee_absences")
	qb.Filter("id", "eq", id)

	statusCode, err := h.PostgrestRequest("DELETE", qb.Build(), nil, nil)
	if err != nil || statusCode >= 400 {
		return c.Status(statusCode).JSON(fiber.Map{"error": "Failed to delete absence"})
	}

	return c.JSON(fiber.Map{"message": "Absence deleted"})
}

// ==================== Employee Resource Settings ====================

// UpdateEmployeeResourceSettings updates resource planning fields for an employee
func (h *Handler) UpdateEmployeeResourceSettings(c *fiber.Ctx) error {
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

	body, _ := json.Marshal(payload)
	qb := NewQueryBuilder("employees")
	qb.Filter("id", "eq", id)

	var result []models.Employee
	statusCode, err := h.PostgrestRequest("PATCH", qb.Build(), body, &result)
	if err != nil || statusCode >= 400 {
		return c.Status(statusCode).JSON(fiber.Map{"error": "Failed to update employee"})
	}

	if len(result) > 0 {
		return c.JSON(result[0])
	}
	return c.JSON(fiber.Map{"message": "Employee resource settings updated"})
}
