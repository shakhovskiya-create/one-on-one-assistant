package handlers

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/ekf/one-on-one-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// CalendarAuthRequest contains EWS auth credentials
type CalendarAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// GetCalendar returns calendar events from Exchange
func (h *Handler) GetCalendar(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	employeeID := c.Params("id")

	// Get employee info
	var employees []struct {
		Email             string  `json:"email"`
		ADLogin           string  `json:"ad_login"`
		EncryptedPassword *string `json:"encrypted_password"`
	}
	err := h.DB.From("employees").Select("email, ad_login, encrypted_password").Eq("id", employeeID).Limit(1).Execute(&employees)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка базы данных: " + err.Error()})
	}
	if len(employees) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Сотрудник не найден"})
	}
	employee := employees[0]

	// Determine email to use
	ewsEmail := employee.Email
	if ewsEmail == "" && strings.Contains(employee.ADLogin, "@") {
		ewsEmail = employee.ADLogin
	}
	// If still no email, construct from ad_login
	if ewsEmail == "" && employee.ADLogin != "" {
		ewsEmail = employee.ADLogin + "@ekf.su"
	}
	if ewsEmail == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email сотрудника не найден. Обратитесь к администратору."})
	}

	daysBack := c.QueryInt("days_back", 7)
	daysForward := c.QueryInt("days_forward", 30)

	var events interface{}
	var source string
	var getErr error

	// Try connector first - use user's own credentials if available
	if h.Connector.IsConnected() {
		// Get user's encrypted password for EWS authentication
		var username, password string
		if employee.EncryptedPassword != nil && *employee.EncryptedPassword != "" {
			// Decrypt user's password
			decrypted, err := utils.DecryptPassword(*employee.EncryptedPassword, h.Config.JWTSecret)
			if err == nil {
				username = "ekfgroup\\" + employee.ADLogin
				password = decrypted
				log.Printf("Using user's own credentials for calendar: %s", username)
			} else {
				log.Printf("ERROR: Failed to decrypt password for %s: %v", employee.Email, err)
			}
		} else {
			log.Printf("WARNING: No encrypted password for %s", employee.Email)
		}

		result, err := h.Connector.SendCommand("get_calendar", map[string]interface{}{
			"email":        ewsEmail,
			"username":     username, // User's own credentials
			"password":     password,
			"days_back":    daysBack,
			"days_forward": daysForward,
		}, 30*time.Second)

		if err == nil {
			events = result
			source = "connector"
		} else {
			getErr = err
		}
	}

	// Fallback to direct EWS if connector failed (uses config credentials)
	if events == nil && h.EWS != nil {
		ewsEvents, err := h.EWS.GetCalendarEvents(ewsEmail, h.Config.EWSUsername, h.Config.EWSPassword, daysBack, daysForward)
		if err != nil {
			if getErr != nil {
				return c.Status(500).JSON(fiber.Map{
					"error":           "Ошибка подключения к Exchange",
					"connector_error": getErr.Error(),
					"ews_error":       err.Error(),
				})
			}
			return c.Status(500).JSON(fiber.Map{"error": "EWS error: " + err.Error()})
		}
		events = ewsEvents
		source = "direct_ews"
	}

	if events == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Не удалось получить календарь: коннектор недоступен и прямое подключение не настроено",
		})
	}

	return c.JSON(fiber.Map{
		"employee_id": employeeID,
		"email":       ewsEmail,
		"events":      events,
		"source":      source,
	})
}

// GetCalendarSimple returns calendar from database only
func (h *Handler) GetCalendarSimple(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	employeeID := c.Params("id")

	var meetings []map[string]interface{}
	h.DB.From("meetings").Select("*").Eq("employee_id", employeeID).Order("date", true).Execute(&meetings)

	return c.JSON(fiber.Map{
		"employee_id": employeeID,
		"events":      meetings,
		"source":      "database",
	})
}

// FreeSlotsRequest contains free slots query params
type FreeSlotsRequest struct {
	AttendeeIDs     []string `json:"attendee_ids"`
	Username        string   `json:"username"`
	Password        string   `json:"password"`
	StartDate       string   `json:"start_date"`
	EndDate         string   `json:"end_date"`
	DurationMinutes int      `json:"duration_minutes"`
}

// FindFreeSlots returns free slots using EWS
func (h *Handler) FindFreeSlots(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var req FreeSlotsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Get emails
	var employees []struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	h.DB.From("employees").Select("id, email").In("id", req.AttendeeIDs).Execute(&employees)

	var emails []string
	for _, e := range employees {
		if e.Email != "" {
			emails = append(emails, e.Email)
		}
	}

	if len(emails) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No valid employee emails found"})
	}

	freeBusy, err := h.EWS.GetFreeBusy(emails, req.Username, req.Password, req.StartDate, req.EndDate)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "EWS error: " + err.Error()})
	}

	// Collect all busy times
	var allBusy []map[string]string
	for _, times := range freeBusy {
		for _, t := range times {
			allBusy = append(allBusy, map[string]string{
				"start":  t.Start,
				"end":    t.End,
				"status": t.Status,
			})
		}
	}

	return c.JSON(fiber.Map{
		"attendees":      emails,
		"free_busy":      freeBusy,
		"all_busy_times": allBusy,
		"source":         "exchange_ews",
	})
}

// FreeSlotsSimple returns free slots from database only
func (h *Handler) FreeSlotsSimple(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	attendeeIDs := strings.Split(c.Query("attendee_ids"), ",")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var meetings []struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}
	h.DB.From("meetings").Select("start_time, end_time").In("employee_id", attendeeIDs).Gte("date", startDate).Lte("date", endDate).Execute(&meetings)

	var busyTimes []map[string]string
	for _, m := range meetings {
		if m.StartTime != "" && m.EndTime != "" {
			busyTimes = append(busyTimes, map[string]string{
				"start": m.StartTime,
				"end":   m.EndTime,
			})
		}
	}

	return c.JSON(fiber.Map{
		"busy_times": busyTimes,
		"source":     "database",
	})
}

// CalendarSyncRequest contains sync params
type CalendarSyncRequest struct {
	EmployeeID  string `json:"employee_id"`
	DaysBack    int    `json:"days_back"`
	DaysForward int    `json:"days_forward"`
}

// SyncCalendar syncs meetings from Exchange to database
func (h *Handler) SyncCalendar(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var req CalendarSyncRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.DaysBack == 0 {
		req.DaysBack = 7
	}
	if req.DaysForward == 0 {
		req.DaysForward = 30
	}

	// Get employee info
	var employees []struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		ADLogin string `json:"ad_login"`
		Name    string `json:"name"`
	}
	err := h.DB.From("employees").Select("id, email, ad_login, name").Eq("id", req.EmployeeID).Limit(1).Execute(&employees)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка базы данных: " + err.Error(), "employee_id": req.EmployeeID})
	}
	if len(employees) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Сотрудник не найден", "employee_id": req.EmployeeID})
	}
	employee := employees[0]

	// Determine the email to use for EWS
	ewsEmail := employee.Email
	if ewsEmail == "" && strings.Contains(employee.ADLogin, "@") {
		ewsEmail = employee.ADLogin
	}
	// If still no email, construct from ad_login
	if ewsEmail == "" && employee.ADLogin != "" {
		ewsEmail = employee.ADLogin + "@ekf.su"
	}

	if ewsEmail == "" {
		return c.Status(400).JSON(fiber.Map{
			"error":       "Не удалось определить email для синхронизации. Обратитесь к администратору.",
			"employee_id": employee.ID,
			"name":        employee.Name,
		})
	}

	var events interface{}
	var getErr error

	// Try connector first - uses service account credentials
	if h.Connector.IsConnected() {
		result, err := h.Connector.SendCommand("sync_calendar", map[string]interface{}{
			"email":        ewsEmail,
			"days_back":    req.DaysBack,
			"days_forward": req.DaysForward,
		}, 120*time.Second)

		if err == nil {
			events = result
		} else {
			getErr = err
		}
	}

	// Fallback to direct EWS if connector failed (uses config credentials)
	if events == nil && h.EWS != nil {
		ewsEvents, err := h.EWS.GetCalendarEvents(ewsEmail, h.Config.EWSUsername, h.Config.EWSPassword, req.DaysBack, req.DaysForward)
		if err != nil {
			if getErr != nil {
				return c.Status(500).JSON(fiber.Map{
					"error":           "Ошибка подключения к Exchange",
					"connector_error": getErr.Error(),
					"ews_error":       err.Error(),
					"ews_email":       ewsEmail,
					"ews_url":         h.Config.EWSURL,
				})
			}
			return c.Status(500).JSON(fiber.Map{
				"error":     "Ошибка подключения к Exchange: " + err.Error(),
				"ews_email": ewsEmail,
				"ews_url":   h.Config.EWSURL,
			})
		}
		events = ewsEvents
	}

	if events == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Не удалось получить календарь: коннектор недоступен и прямое подключение не настроено",
		})
	}

	// Convert events to slice for processing
	eventsList, ok := events.([]interface{})
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid events format"})
	}

	// If we successfully connected and employee had no email, save it
	if employee.Email == "" && ewsEmail != "" {
		h.DB.Update("employees", "id", employee.ID, map[string]string{"email": ewsEmail})
	}

	if len(eventsList) == 0 {
		return c.JSON(fiber.Map{"synced": 0, "message": "No events found", "ews_email": ewsEmail})
	}

	// Build email lookup
	var allEmployees []struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	h.DB.From("employees").Select("id, email, name").Execute(&allEmployees)

	emailToEmp := make(map[string]struct {
		ID   string
		Name string
	})
	for _, e := range allEmployees {
		if e.Email != "" {
			emailToEmp[strings.ToLower(e.Email)] = struct {
				ID   string
				Name string
			}{e.ID, e.Name}
		}
	}

	synced := 0
	for _, ev := range eventsList {
		eventMap, ok := ev.(map[string]interface{})
		if !ok {
			continue
		}

		eventID, _ := eventMap["id"].(string)
		if eventID == "" {
			continue
		}

		// Check existing
		var existing []struct {
			ID string `json:"id"`
		}
		h.DB.From("meetings").Select("id").Eq("exchange_id", eventID).Execute(&existing)

		subject, _ := eventMap["subject"].(string)
		start, _ := eventMap["start"].(string)
		end, _ := eventMap["end"].(string)
		location, _ := eventMap["location"].(string)

		meetingData := map[string]interface{}{
			"exchange_id":   eventID,
			"title":         subject,
			"date":          start[:10],
			"start_time":    start,
			"end_time":      end,
			"location":      location,
			"exchange_data": eventMap,
		}

		if organizer, ok := eventMap["organizer"].(map[string]interface{}); ok {
			if orgEmail, ok := organizer["email"].(string); ok && orgEmail != "" {
				if emp, ok := emailToEmp[strings.ToLower(orgEmail)]; ok {
					meetingData["organizer_id"] = emp.ID
				}
			}
		}

		var meetingID string
		if len(existing) > 0 {
			h.DB.Update("meetings", "id", existing[0].ID, meetingData)
			meetingID = existing[0].ID
		} else {
			result, _ := h.DB.Insert("meetings", meetingData)
			var created []map[string]interface{}
			json.Unmarshal(result, &created)
			if len(created) > 0 {
				meetingID, _ = created[0]["id"].(string)
			}
		}

		// Sync participants
		if attendees, ok := eventMap["attendees"].([]interface{}); ok {
			for _, att := range attendees {
				if attendeeMap, ok := att.(map[string]interface{}); ok {
					if attEmail, ok := attendeeMap["email"].(string); ok && attEmail != "" {
						if emp, ok := emailToEmp[strings.ToLower(attEmail)]; ok {
							h.DB.Insert("meeting_participants", map[string]string{
								"meeting_id":  meetingID,
								"employee_id": emp.ID,
							})
						}
					}
				}
			}
		}

		synced++
	}

	return c.JSON(fiber.Map{
		"synced":         synced,
		"total_events":   len(eventsList),
		"employee_email": employee.Email,
		"source":         "connector_ews",
	})
}
