package handlers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/ekf/one-on-one-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"golang.org/x/crypto/bcrypt"
)

// ConnectorStatus returns connector status
func (h *Handler) ConnectorStatus(c *fiber.Ctx) error {
	adStatus := "unknown"
	employeeCount := 0

	if h.DB != nil {
		var employees []struct {
			ID string `json:"id"`
		}
		h.DB.From("employees").Select("id").Execute(&employees)
		employeeCount = len(employees)
		if employeeCount > 0 {
			if h.Connector.IsConnected() {
				adStatus = "connected"
			} else {
				adStatus = "disconnected"
			}
		}
	}

	return c.JSON(fiber.Map{
		"connected":            h.Connector.IsConnected(),
		"pending_requests":     0,
		"ad_status":            adStatus,
		"ad_sync_enabled":      false,
		"last_sync":            "",
		"employee_count":       employeeCount,
		"calendar_integration": "ews",
		"ews_url":              h.Config.EWSURL,
		"ews_configured":       h.Config.EWSURL != "" && h.Config.EWSUsername != "" && h.Config.EWSPassword != "",
		"yandex_configured":    h.Config.YandexAPIKey != "" && h.Config.YandexFolderID != "",
		"openai_configured":    h.Config.OpenAIKey != "",
		"anthropic_configured": h.Config.AnthropicKey != "",
	})
}

// ConnectorWebSocket handles WebSocket connection from on-prem connector
func (h *Handler) ConnectorWebSocket(conn *websocket.Conn) {
	// Get API key from query
	apiKey := conn.Query("token")

	if err := h.Connector.Connect(conn, apiKey); err != nil {
		conn.Close()
		return
	}

	defer h.Connector.Disconnect()

	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		h.Connector.HandleMessage(messageType, data)
	}
}

// SyncADUsers syncs users from Active Directory
func (h *Handler) SyncADUsers(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	mode := c.Query("mode", "full")
	includePhotos := c.Query("include_photos", "true") == "true"
	requireDepartment := c.Query("require_department", "true") == "true"

	// Stats tracking
	stats := fiber.Map{
		"mode":               mode,
		"total_in_ad":        0,
		"with_department":    0,
		"without_department": 0,
		"filtered_out":       0,
		"new_users":          0,
		"updated_users":      0,
		"skipped_existing":   0,
		"managers_updated":   0,
		"errors":             []string{},
	}

	// Get existing emails
	existingEmails := make(map[string]bool)
	if mode == "new_only" || mode == "changes" {
		var existing []struct {
			Email string `json:"email"`
		}
		h.DB.From("employees").Select("email").Execute(&existing)
		for _, e := range existing {
			if e.Email != "" {
				existingEmails[strings.ToLower(e.Email)] = true
			}
		}
	}

	// Fetch users from connector
	result, err := h.Connector.SendCommand("sync_users", map[string]interface{}{
		"include_photo":      includePhotos,
		"require_department": requireDepartment,
		"require_email":      true,
		"mode":               mode,
	}, 300*time.Second)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	users, _ := result["users"].([]interface{})
	if connStats, ok := result["stats"].(map[string]interface{}); ok {
		stats["total_in_ad"] = connStats["total_in_ad"]
		stats["with_department"] = connStats["with_department"]
		stats["without_department"] = connStats["without_department"]
		stats["filtered_out"] = connStats["filtered_out"]
	}

	// Process users
	var batch []map[string]interface{}
	newCount := 0
	updatedCount := 0
	skippedCount := 0

	for _, u := range users {
		user, ok := u.(map[string]interface{})
		if !ok {
			continue
		}

		email, _ := user["email"].(string)
		if email == "" {
			continue
		}

		isExisting := existingEmails[strings.ToLower(email)]

		if mode == "new_only" && isExisting {
			skippedCount++
			continue
		}

		userData := map[string]interface{}{
			"name":       user["name"],
			"email":      email,
			"position":   user["title"],
			"department": user["department"],
			"ad_dn":      user["dn"],
			"manager_dn": user["manager_dn"],
			"ad_login":   user["login"],
			"phone":      user["phone"],
			"mobile":     user["mobile"],
		}

		if includePhotos {
			if photo, ok := user["photo_base64"].(string); ok && photo != "" {
				userData["photo_base64"] = photo
			}
		}

		batch = append(batch, userData)

		if isExisting {
			updatedCount++
		} else {
			newCount++
		}
	}

	// Upsert batch
	if len(batch) > 0 {
		_, err := h.DB.Upsert("employees", batch, "email")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	// Update manager relationships
	var employees []struct {
		ID        string  `json:"id"`
		ADDN      *string `json:"ad_dn"`
		ManagerDN *string `json:"manager_dn"`
	}
	h.DB.From("employees").Select("id, ad_dn, manager_dn").Execute(&employees)

	dnToID := make(map[string]string)
	for _, e := range employees {
		if e.ADDN != nil {
			dnToID[*e.ADDN] = e.ID
		}
	}

	managersUpdated := 0
	for _, emp := range employees {
		if emp.ManagerDN != nil {
			if managerID, ok := dnToID[*emp.ManagerDN]; ok {
				h.DB.Update("employees", "id", emp.ID, map[string]interface{}{
					"manager_id": managerID,
				})
				managersUpdated++
			}
		}
	}

	stats["new_users"] = newCount
	stats["updated_users"] = updatedCount
	stats["skipped_existing"] = skippedCount
	stats["managers_updated"] = managersUpdated

	return c.JSON(fiber.Map{
		"success":  true,
		"imported": newCount + updatedCount,
		"stats":    stats,
	})
}

// AuthenticateAD authenticates a user against AD or local database
func (h *Handler) AuthenticateAD(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Username and password required"})
	}

	// First, try local authentication (works without connector)
	if h.DB != nil {
		employee, authErr := h.authenticateLocal(username, password)
		if authErr == nil && employee != nil {
			forceChange, _ := employee["force_password_change"].(bool)
			delete(employee, "force_password_change")

			// Generate JWT token
			userID, _ := employee["id"].(string)
			email, _ := employee["email"].(string)
			name, _ := employee["name"].(string)
			department, _ := employee["department"].(string)

			token, err := h.JWT.GenerateToken(userID, email, name, department)
			if err != nil {
				token = generateToken() // Fallback to simple token
			}

			return c.JSON(fiber.Map{
				"authenticated":         true,
				"employee":              employee,
				"token":                 token,
				"force_password_change": forceChange,
			})
		}
	}

	// Fall back to connector-based AD authentication
	if h.Connector.IsConnected() {
		result, err := h.Connector.SendCommand("authenticate", map[string]interface{}{
			"username": username,
			"password": password,
		}, 30*time.Second)

		if err == nil {
			authenticated, _ := result["authenticated"].(bool)
			if authenticated {
				user, _ := result["user"].(map[string]interface{})
				email, _ := user["email"].(string)

				var employee map[string]interface{}

				if h.DB != nil && email != "" {
					var existing []map[string]interface{}
					h.DB.From("employees").Select("*").Eq("email", email).Execute(&existing)

					// Encrypt and store password for EWS access
					encryptedPassword, encErr := utils.EncryptPassword(password, h.Config.JWTSecret)
					if encErr != nil {
						encryptedPassword = "" // If encryption fails, don't store
					}

					if len(existing) > 0 {
						employee = existing[0]
						employeeID, _ := employee["id"].(string)

						// Update encrypted password
						if encryptedPassword != "" {
							updateData := map[string]interface{}{
								"encrypted_password": encryptedPassword,
							}
							h.DB.Update("employees", "id", employeeID, updateData)
						}
					} else {
						// Auto-create from AD
						empData := map[string]interface{}{
							"name":               user["name"],
							"email":              email,
							"position":           user["title"],
							"ad_dn":              user["dn"],
							"ad_login":           user["login"],
							"encrypted_password": encryptedPassword,
						}
						insertResult, _ := h.DB.Insert("employees", empData)
						var created []map[string]interface{}
						json.Unmarshal(insertResult, &created)
						if len(created) > 0 {
							employee = created[0]
						}
					}
				}

				// Generate JWT token
				userID, _ := employee["id"].(string)
				name, _ := employee["name"].(string)
				department, _ := employee["department"].(string)

				token, tokenErr := h.JWT.GenerateToken(userID, email, name, department)
				if tokenErr != nil {
					token = generateToken() // Fallback to simple token
				}

				return c.JSON(fiber.Map{
					"authenticated": true,
					"employee":      employee,
					"token":         token,
				})
			}
			errMsg, _ := result["error"].(string)
			return c.JSON(fiber.Map{"authenticated": false, "error": errMsg})
		}
	}

	// Both local and AD auth failed
	return c.JSON(fiber.Map{
		"authenticated": false,
		"error":         "Неверные учётные данные",
	})
}

// authenticateLocal checks local password hash in database
func (h *Handler) authenticateLocal(username, password string) (map[string]interface{}, error) {
	// Normalize username - handle domain\user format
	loginName := username
	if strings.Contains(username, "\\") {
		parts := strings.SplitN(username, "\\", 2)
		loginName = parts[1]
	}

	// Try to find user by email or ad_login
	var employees []map[string]interface{}

	// First try by email
	h.DB.From("employees").Select("*").Ilike("email", loginName+"%").Execute(&employees)

	if len(employees) == 0 {
		// Try by ad_login
		h.DB.From("employees").Select("*").Ilike("ad_login", loginName).Execute(&employees)
	}

	if len(employees) == 0 {
		// Try exact email match
		h.DB.From("employees").Select("*").Eq("email", loginName).Execute(&employees)
	}

	if len(employees) == 0 {
		return nil, fiber.NewError(401, "User not found")
	}

	employee := employees[0]
	passwordHash, _ := employee["password_hash"].(string)

	if passwordHash == "" {
		return nil, fiber.NewError(401, "No local password set")
	}

	// Compare password with hash
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, fiber.NewError(401, "Invalid password")
	}

	// Remove password_hash from response
	delete(employee, "password_hash")

	return employee, nil
}

// SetLocalPassword sets a local password for a user (admin endpoint)
func (h *Handler) SetLocalPassword(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email and password required"})
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Update employee with password hash
	_, err = h.DB.Update("employees", "email", req.Email, map[string]interface{}{
		"password_hash": string(hash),
		"is_local_user": true,
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update password"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Password set successfully",
	})
}

// ChangePassword allows user to change their own password
func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var req struct {
		UserID      string `json:"user_id"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.UserID == "" || req.NewPassword == "" {
		return c.Status(400).JSON(fiber.Map{"error": "User ID and new password required"})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 6 characters"})
	}

	// Verify old password if provided
	if req.OldPassword != "" {
		var employees []map[string]interface{}
		h.DB.From("employees").Select("password_hash").Eq("id", req.UserID).Execute(&employees)

		if len(employees) == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "User not found"})
		}

		oldHash, _ := employees[0]["password_hash"].(string)
		if oldHash != "" {
			if err := bcrypt.CompareHashAndPassword([]byte(oldHash), []byte(req.OldPassword)); err != nil {
				return c.Status(401).JSON(fiber.Map{"error": "Неверный текущий пароль"})
			}
		}
	}

	// Hash the new password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Update password and clear force_password_change flag
	_, err = h.DB.Update("employees", "id", req.UserID, map[string]interface{}{
		"password_hash":         string(hash),
		"force_password_change": false,
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update password"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Пароль успешно изменён",
	})
}

// CreateLocalUser creates a new local user (not from AD)
func (h *Handler) CreateLocalUser(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var req struct {
		Name       string `json:"name"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		Position   string `json:"position"`
		Department string `json:"department"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name, email and password required"})
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Create employee
	empData := map[string]interface{}{
		"name":          req.Name,
		"email":         req.Email,
		"password_hash": string(hash),
		"position":      req.Position,
		"department":    req.Department,
		"is_local_user": true,
	}

	result, err := h.DB.Insert("employees", empData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create user: " + err.Error()})
	}

	var created []map[string]interface{}
	json.Unmarshal(result, &created)

	if len(created) > 0 {
		delete(created[0], "password_hash")
		return c.JSON(fiber.Map{
			"success":  true,
			"employee": created[0],
		})
	}

	return c.JSON(fiber.Map{"success": true})
}

// RefreshToken refreshes a JWT token
func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Missing authorization header"})
	}

	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	newToken, err := h.JWT.RefreshToken(tokenString)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	return c.JSON(fiber.Map{
		"token": newToken,
	})
}

// GetMe returns current user info from JWT token
func (h *Handler) GetMe(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil || userID.(string) == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Not authenticated"})
	}

	if h.DB == nil {
		return c.JSON(fiber.Map{
			"id":         userID,
			"email":      c.Locals("email"),
			"name":       c.Locals("name"),
			"department": c.Locals("department"),
		})
	}

	var employees []map[string]interface{}
	h.DB.From("employees").Select("id, name, email, position, department, manager_id, photo_base64").Eq("id", userID.(string)).Execute(&employees)

	if len(employees) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(employees[0])
}

// GetSubordinates returns subordinates for a manager
func (h *Handler) GetSubordinates(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	employeeID := c.Params("id")
	fromDB := c.Query("from_db", "true") == "true"

	if fromDB {
		var subordinates []map[string]interface{}
		h.DB.From("employees").Select("id, name, email, position, department, manager_id").Eq("manager_id", employeeID).Execute(&subordinates)
		return c.JSON(subordinates)
	}

	// Get from AD via connector
	var employee struct {
		ADDN *string `json:"ad_dn"`
	}
	h.DB.From("employees").Select("ad_dn").Eq("id", employeeID).Single().Execute(&employee)

	if employee.ADDN == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Employee not found or not synced from AD"})
	}

	result, err := h.Connector.SendCommand("get_subordinates", map[string]interface{}{
		"manager_dn": *employee.ADDN,
	}, 30*time.Second)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

func generateToken() string {
	// Simple UUID-like token
	return time.Now().Format("20060102150405") + "-" + randomString(16)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
