package middleware

import (
	"github.com/ekf/one-on-one-backend/internal/database"
	"github.com/gofiber/fiber/v2"
)

// RBACContext holds user authorization context
type RBACContext struct {
	UserID    string
	Role      string
	ManagerID *string
}

// GetRBACContext extracts and enriches RBAC context from request
func GetRBACContext(c *fiber.Ctx, db database.DBClient) (*RBACContext, error) {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return nil, fiber.NewError(401, "Not authenticated")
	}

	// Get user's role and manager_id from database
	var employee struct {
		Role      *string `json:"role"`
		ManagerID *string `json:"manager_id"`
	}
	err := db.From("employees").Select("role, manager_id").Eq("id", userID).Single().Execute(&employee)
	if err != nil {
		return nil, fiber.NewError(403, "User not found")
	}

	role := "user"
	if employee.Role != nil && *employee.Role != "" {
		role = *employee.Role
	}

	return &RBACContext{
		UserID:    userID,
		Role:      role,
		ManagerID: employee.ManagerID,
	}, nil
}

// IsAdmin checks if user has admin or super_admin role
func (ctx *RBACContext) IsAdmin() bool {
	return ctx.Role == "admin" || ctx.Role == "super_admin"
}

// IsHR checks if user has HR role
func (ctx *RBACContext) IsHR() bool {
	return ctx.Role == "hr" || ctx.Role == "hr_admin" || ctx.IsAdmin()
}

// IsSelf checks if user is accessing their own data
func (ctx *RBACContext) IsSelf(targetID string) bool {
	return ctx.UserID == targetID
}

// IsManager checks if user is a manager (has direct reports)
func (ctx *RBACContext) IsManager() bool {
	return ctx.Role == "manager" || ctx.Role == "team_lead" || ctx.IsAdmin()
}

// CanAccessEmployeeData checks if user can view basic employee info
// Anyone can view basic employee data (name, position, department)
func (ctx *RBACContext) CanAccessEmployeeData(targetID string) bool {
	return true // Basic info is available to all authenticated users
}

// CanAccessSensitiveEmployeeData checks if user can view sensitive employee data
// Sensitive data includes: mood scores, red flags, HR analytics, etc.
func (ctx *RBACContext) CanAccessSensitiveEmployeeData(targetID string, db database.DBClient) bool {
	// Admins and HR can access all data
	if ctx.IsHR() || ctx.IsAdmin() {
		return true
	}

	// User can access their own data
	if ctx.IsSelf(targetID) {
		return true
	}

	// Check if user is the target's manager (direct or indirect)
	return ctx.isManagerOf(targetID, db)
}

// CanAccessCalendar checks if user can view someone's calendar
func (ctx *RBACContext) CanAccessCalendar(targetID string, db database.DBClient) bool {
	// Admins can access all calendars
	if ctx.IsAdmin() {
		return true
	}

	// User can access their own calendar
	if ctx.IsSelf(targetID) {
		return true
	}

	// Check if user is the target's manager
	return ctx.isManagerOf(targetID, db)
}

// CanModifyEmployee checks if user can modify employee data
func (ctx *RBACContext) CanModifyEmployee(targetID string) bool {
	// Only admins and HR can modify employee data
	return ctx.IsHR() || ctx.IsAdmin()
}

// CanDeleteEmployee checks if user can delete employee
func (ctx *RBACContext) CanDeleteEmployee() bool {
	// Only admins can delete employees
	return ctx.IsAdmin()
}

// isManagerOf checks if the user is a manager of the target employee
// This includes both direct and indirect management (manager of manager)
func (ctx *RBACContext) isManagerOf(targetID string, db database.DBClient) bool {
	// Check up to 5 levels of management hierarchy
	currentID := targetID
	for i := 0; i < 5; i++ {
		var employee struct {
			ManagerID *string `json:"manager_id"`
		}
		err := db.From("employees").Select("manager_id").Eq("id", currentID).Single().Execute(&employee)
		if err != nil || employee.ManagerID == nil {
			return false
		}

		if *employee.ManagerID == ctx.UserID {
			return true
		}

		currentID = *employee.ManagerID
	}
	return false
}

// RequireEmployeeAccess is a middleware that checks if user can access employee data
func RequireEmployeeAccess(db database.DBClient, checkSensitive bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, err := GetRBACContext(c, db)
		if err != nil {
			return err
		}

		targetID := c.Params("id")
		if targetID == "" {
			targetID = c.Query("employee_id")
		}
		if targetID == "" {
			// No target ID, let the handler deal with it
			c.Locals("rbac_context", ctx)
			return c.Next()
		}

		if checkSensitive {
			if !ctx.CanAccessSensitiveEmployeeData(targetID, db) {
				return c.Status(403).JSON(fiber.Map{
					"error": "Access denied: you don't have permission to view this employee's sensitive data",
				})
			}
		} else {
			if !ctx.CanAccessEmployeeData(targetID) {
				return c.Status(403).JSON(fiber.Map{
					"error": "Access denied: you don't have permission to view this employee",
				})
			}
		}

		c.Locals("rbac_context", ctx)
		return c.Next()
	}
}

// RequireCalendarAccess is a middleware that checks if user can access calendar
func RequireCalendarAccess(db database.DBClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, err := GetRBACContext(c, db)
		if err != nil {
			return err
		}

		targetID := c.Params("id")
		if targetID == "" {
			targetID = c.Query("employee_id")
		}
		if targetID == "" {
			// No target ID, let the handler deal with it
			c.Locals("rbac_context", ctx)
			return c.Next()
		}

		if !ctx.CanAccessCalendar(targetID, db) {
			return c.Status(403).JSON(fiber.Map{
				"error": "Access denied: you don't have permission to view this calendar",
			})
		}

		c.Locals("rbac_context", ctx)
		return c.Next()
	}
}

// RequireAdminOrHR is a middleware that allows only admin or HR roles
func RequireAdminOrHR(db database.DBClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, err := GetRBACContext(c, db)
		if err != nil {
			return err
		}

		if !ctx.IsHR() && !ctx.IsAdmin() {
			return c.Status(403).JSON(fiber.Map{
				"error": "Access denied: admin or HR role required",
			})
		}

		c.Locals("rbac_context", ctx)
		return c.Next()
	}
}
