package middleware

import (
	"github.com/ekf/one-on-one-backend/internal/database"
	"github.com/gofiber/fiber/v2"
)

// AdminAuth middleware checks if user has admin role
func AdminAuth(db database.DBClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Not authenticated"})
		}

		// Get user role from database
		var employee struct {
			Role *string `json:"role"`
		}
		err := db.From("employees").Select("role").Eq("id", userID).Single().Execute(&employee)
		if err != nil {
			return c.Status(403).JSON(fiber.Map{"error": "User not found"})
		}

		role := "user"
		if employee.Role != nil {
			role = *employee.Role
		}

		if role != "admin" && role != "super_admin" {
			return c.Status(403).JSON(fiber.Map{"error": "Admin access required"})
		}

		// Store role in context for further checks
		c.Locals("user_role", role)

		return c.Next()
	}
}

// SuperAdminAuth middleware checks if user has super_admin role
func SuperAdminAuth(db database.DBClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Not authenticated"})
		}

		// Get user role from database
		var employee struct {
			Role *string `json:"role"`
		}
		err := db.From("employees").Select("role").Eq("id", userID).Single().Execute(&employee)
		if err != nil {
			return c.Status(403).JSON(fiber.Map{"error": "User not found"})
		}

		role := "user"
		if employee.Role != nil {
			role = *employee.Role
		}

		if role != "super_admin" {
			return c.Status(403).JSON(fiber.Map{"error": "Super admin access required"})
		}

		c.Locals("user_role", role)

		return c.Next()
	}
}
