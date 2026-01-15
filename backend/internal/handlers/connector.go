package handlers

import (
	"encoding/json"
	"strings"
	"time"

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
		"employee_count":       employeeCount,
		"calendar_integration": "ews",
		"ews_url":              h.Config.EWSURL,
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
		"mode":             mode,
		"total_in_ad":      0,
		"with_department":  0,
		"without_department": 0,
		"filtered_out":     0,
		"new_users":        0,
		"updated_users":    0,
		"skipped_existing": 0,
		"managers_updated": 0,
		"errors":           []string{},
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

// AuthenticateAD authenticates a user against AD
func (h *Handler) AuthenticateAD(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Username and password required"})
	}

	result, err := h.Connector.SendCommand("authenticate", map[string]interface{}{
		"username": username,
		"password": password,
	}, 30*time.Second)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"authenticated": false, "error": err.Error()})
	}

	authenticated, _ := result["authenticated"].(bool)
	if !authenticated {
		errMsg, _ := result["error"].(string)
		return c.JSON(fiber.Map{"authenticated": false, "error": errMsg})
	}

	user, _ := result["user"].(map[string]interface{})
	email, _ := user["email"].(string)

	var employee map[string]interface{}

	if h.DB != nil && email != "" {
		var existing []map[string]interface{}
		h.DB.From("employees").Select("*").Eq("email", email).Execute(&existing)

		if len(existing) > 0 {
			employee = existing[0]
		} else {
			// Auto-create from AD
			empData := map[string]interface{}{
				"name":     user["name"],
				"email":    email,
				"position": user["title"],
				"ad_dn":    user["dn"],
				"ad_login": user["login"],
			}
			result, _ := h.DB.Insert("employees", empData)
			var created []map[string]interface{}
			json.Unmarshal(result, &created)
			if len(created) > 0 {
				employee = created[0]
			}
		}
	}

	// Generate simple session token
	sessionToken := generateToken()

	return c.JSON(fiber.Map{
		"authenticated": true,
		"employee":      employee,
		"token":         sessionToken,
	})
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
