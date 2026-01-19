package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ekf/one-on-one-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
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

// SyncADUsersDirect syncs users directly from AD (not via connector)
func (h *Handler) SyncADUsersDirect(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	if h.AD == nil {
		return c.Status(500).JSON(fiber.Map{"error": "AD not configured"})
	}

	// Get credentials from request body or query params
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Try to parse body, fallback to query params
	c.BodyParser(&req)
	if req.Username == "" {
		req.Username = c.Query("username")
		req.Password = c.Query("password")
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Username and password required for AD sync"})
	}

	// Always include photos as per user requirement
	includePhotos := true

	// Get all users from AD using user credentials
	// This will automatically filter users without departments
	users, err := h.AD.GetAllUsersWithCredentials(req.Username, req.Password, includePhotos)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Stats tracking
	stats := fiber.Map{
		"total_in_ad":   len(users),
		"new_users":     0,
		"updated_users": 0,
	}

	// Get existing emails
	var existing []struct {
		Email string `json:"email"`
	}
	h.DB.From("employees").Select("email").Execute(&existing)
	existingEmails := make(map[string]bool)
	for _, e := range existing {
		if e.Email != "" {
			existingEmails[strings.ToLower(e.Email)] = true
		}
	}

	// Process users
	var batch []map[string]interface{}
	newCount := 0
	updatedCount := 0

	for _, user := range users {
		if user.Email == "" {
			continue
		}

		isExisting := existingEmails[strings.ToLower(user.Email)]

		userData := map[string]interface{}{
			"name":       user.Name,
			"email":      user.Email,
			"position":   user.Title,
			"department": user.Department,
			"ad_dn":      user.DN,
			"manager_dn": user.ManagerDN,
			"ad_login":   user.Login,
			"phone":      user.Phone,
			"mobile":     user.Mobile,
		}

		if includePhotos && user.PhotoBase64 != "" {
			userData["photo_base64"] = user.PhotoBase64
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
	stats["managers_updated"] = managersUpdated

	return c.JSON(fiber.Map{
		"success":  true,
		"imported": newCount + updatedCount,
		"stats":    stats,
	})
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

// AuthenticateAD authenticates a user against AD only
func (h *Handler) AuthenticateAD(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}

	// Try to parse body (JSON or form data)
	if err := c.BodyParser(&req); err != nil {
		// Fallback to manual form parsing
		req.Username = c.FormValue("username")
		req.Password = c.FormValue("password")
	}

	// Also check query params as fallback
	if req.Username == "" {
		req.Username = c.Query("username")
	}
	if req.Password == "" {
		req.Password = c.Query("password")
	}

	username := req.Username
	password := req.Password

	if username == "" || password == "" {
		return c.Status(400).JSON(fiber.Map{
			"error":        "Username and password required",
			"content_type": c.Get("Content-Type"),
		})
	}

	// Try direct AD authentication (no connector needed)
	if h.AD != nil {
		adUser, authErr := h.AD.Authenticate(username, password)
		if authErr != nil {
			// Log AD authentication error for debugging
			c.Locals("ad_error", authErr.Error())
		}
		if authErr == nil && adUser != nil {
			email := adUser.Email
			name := adUser.Name

			// Debug: log what AD returned
			c.Locals("ad_user_info", fiber.Map{
				"email":    email,
				"name":     name,
				"username": adUser.Username,
			})

			var employee map[string]interface{}

			if h.DB != nil && email != "" {
				var existing []map[string]interface{}
				err := h.DB.From("employees").Select("*").Eq("email", email).Limit(1).Execute(&existing)

				// Debug: log database query result
				c.Locals("db_query_error", err)
				c.Locals("db_result_count", len(existing))

				// Encrypt and store password for EWS access
				encryptedPassword, encErr := utils.EncryptPassword(password, h.Config.JWTSecret)
				if encErr != nil {
					encryptedPassword = ""
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
				}
			}

			if employee != nil {
				employeeID, _ := employee["id"].(string)
				department, _ := employee["department"].(string)

				token, err := h.JWT.GenerateToken(employeeID, email, name, department)
				if err != nil {
					token = generateToken()
				}

				return c.JSON(fiber.Map{
					"authenticated":         true,
					"employee":              employee,
					"token":                 token,
					"force_password_change": false,
				})
			}

			// User authenticated in AD but not in database - return error with debug info
			response := fiber.Map{
				"error":         "User not found in system. Please contact administrator.",
				"authenticated": false,
			}
			if adUserInfo := c.Locals("ad_user_info"); adUserInfo != nil {
				response["ad_user"] = adUserInfo
			}
			if dbErr := c.Locals("db_query_error"); dbErr != nil {
				response["db_error"] = fmt.Sprint(dbErr)
			}
			if dbCount := c.Locals("db_result_count"); dbCount != nil {
				response["db_result_count"] = dbCount
			}
			return c.Status(403).JSON(response)
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

	// AD auth failed - include debug info if available
	response := fiber.Map{
		"authenticated": false,
		"error":         "Неверные учётные данные",
	}

	// Add debug error in development
	if adError := c.Locals("ad_error"); adError != nil {
		response["debug_error"] = adError
	}

	return c.JSON(response)
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
