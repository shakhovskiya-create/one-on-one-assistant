package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ekf/one-on-one-backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// ListServiceTickets returns tickets with filters
func (h *Handler) ListServiceTickets(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	query := h.DB.From("service_tickets").Select("*, requester:employees!service_tickets_requester_id_fkey(id, name, photo_base64), assignee:employees!service_tickets_assignee_id_fkey(id, name, photo_base64)")

	// Filters
	if requesterID := c.Query("requester_id"); requesterID != "" {
		query = query.Eq("requester_id", requesterID)
	}
	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		query = query.Eq("assignee_id", assigneeID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Eq("status", status)
	}
	if ticketType := c.Query("type"); ticketType != "" {
		query = query.Eq("type", ticketType)
	}
	if priority := c.Query("priority"); priority != "" {
		query = query.Eq("priority", priority)
	}
	if categoryID := c.Query("category_id"); categoryID != "" {
		query = query.Eq("category_id", categoryID)
	}

	var tickets []models.ServiceTicket
	err := query.Order("created_at", true).Execute(&tickets)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tickets)
}

// GetServiceTicket returns a single ticket with details
func (h *Handler) GetServiceTicket(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var ticket models.ServiceTicket
	err := h.DB.From("service_tickets").
		Select("*, requester:employees!service_tickets_requester_id_fkey(id, name, email, position, department, photo_base64), assignee:employees!service_tickets_assignee_id_fkey(id, name, email, position, department, photo_base64)").
		Eq("id", id).Single().Execute(&ticket)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Ticket not found"})
	}

	// Get comments
	var comments []models.ServiceTicketComment
	h.DB.From("service_ticket_comments").Select("*, author:employees(id, name, photo_base64)").Eq("ticket_id", id).Order("created_at", false).Execute(&comments)
	ticket.Comments = comments

	// Get activity
	var activity []models.ServiceTicketActivity
	h.DB.From("service_ticket_activity").Select("*, actor:employees(id, name)").Eq("ticket_id", id).Order("created_at", true).Limit(50).Execute(&activity)
	ticket.Activity = activity

	return c.JSON(ticket)
}

// CreateServiceTicket creates a new ticket
func (h *Handler) CreateServiceTicket(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var input struct {
		Type        string  `json:"type"`
		Title       string  `json:"title"`
		Description *string `json:"description"`
		CategoryID  *string `json:"category_id"`
		Priority    string  `json:"priority"`
		Impact      string  `json:"impact"`
		RequesterID string  `json:"requester_id"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Title == "" || input.RequesterID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Title and requester_id required"})
	}

	// Default values
	if input.Type == "" {
		input.Type = "service_request"
	}
	if input.Priority == "" {
		input.Priority = "medium"
	}
	if input.Impact == "" {
		input.Impact = "individual"
	}

	// Generate ticket number based on type
	prefix := "SD"
	switch input.Type {
	case "incident":
		prefix = "INC"
	case "service_request":
		prefix = "REQ"
	case "change":
		prefix = "CHG"
	case "problem":
		prefix = "PRB"
	}

	// Get next number
	var lastTicket []struct {
		Number string `json:"number"`
	}
	h.DB.From("service_tickets").Select("number").Ilike("number", prefix+"-%").Order("created_at", true).Limit(1).Execute(&lastTicket)

	nextNum := 1
	if len(lastTicket) > 0 {
		// Parse number like "INC-2026-0005" -> 5
		var num int
		fmt.Sscanf(lastTicket[0].Number, prefix+"-2026-%d", &num)
		nextNum = num + 1
	}

	ticketNumber := fmt.Sprintf("%s-2026-%04d", prefix, nextNum)

	// Calculate SLA deadline
	var slaDeadline *time.Time
	slaHours := 24 // Default 24 hours
	switch input.Priority {
	case "critical":
		slaHours = 4
	case "high":
		slaHours = 8
	case "medium":
		slaHours = 24
	case "low":
		slaHours = 72
	}
	deadline := time.Now().Add(time.Duration(slaHours) * time.Hour)
	slaDeadline = &deadline

	data := map[string]interface{}{
		"number":       ticketNumber,
		"type":         input.Type,
		"title":        input.Title,
		"description":  input.Description,
		"category_id":  input.CategoryID,
		"priority":     input.Priority,
		"impact":       input.Impact,
		"status":       "new",
		"requester_id": input.RequesterID,
		"sla_deadline": slaDeadline,
	}

	result, err := h.DB.Insert("service_tickets", data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var created []models.ServiceTicket
	json.Unmarshal(result, &created)
	if len(created) == 0 {
		return c.Status(201).JSON(fiber.Map{"status": "created", "number": ticketNumber})
	}

	// Log activity
	h.DB.Insert("service_ticket_activity", map[string]interface{}{
		"ticket_id": created[0].ID,
		"actor_id":  input.RequesterID,
		"action":    "created",
	})

	return c.Status(201).JSON(created[0])
}

// UpdateServiceTicket updates a ticket
func (h *Handler) UpdateServiceTicket(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var input struct {
		Status      *string `json:"status"`
		Priority    *string `json:"priority"`
		AssigneeID  *string `json:"assignee_id"`
		Resolution  *string `json:"resolution"`
		CategoryID  *string `json:"category_id"`
		Title       *string `json:"title"`
		Description *string `json:"description"`
		ActorID     *string `json:"actor_id"` // Who made the change
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Get current ticket for activity logging
	var currentTicket models.ServiceTicket
	h.DB.From("service_tickets").Eq("id", id).Single().Execute(&currentTicket)

	data := map[string]interface{}{
		"updated_at": time.Now(),
	}

	// Track changes for activity log
	var actorID *string = input.ActorID

	if input.Status != nil {
		// Log status change
		if currentTicket.Status != *input.Status {
			h.DB.Insert("service_ticket_activity", map[string]interface{}{
				"ticket_id": id,
				"actor_id":  actorID,
				"action":    "status_changed",
				"old_value": currentTicket.Status,
				"new_value": *input.Status,
			})
		}
		data["status"] = *input.Status

		// Set resolved/closed timestamps
		if *input.Status == "resolved" {
			now := time.Now()
			data["resolved_at"] = now
		}
		if *input.Status == "closed" {
			now := time.Now()
			data["closed_at"] = now
		}
	}
	if input.Priority != nil {
		data["priority"] = *input.Priority
	}
	if input.AssigneeID != nil {
		// Log assignment
		h.DB.Insert("service_ticket_activity", map[string]interface{}{
			"ticket_id": id,
			"actor_id":  actorID,
			"action":    "assigned",
			"new_value": *input.AssigneeID,
		})
		data["assignee_id"] = *input.AssigneeID
	}
	if input.Resolution != nil {
		data["resolution"] = *input.Resolution
	}
	if input.CategoryID != nil {
		data["category_id"] = *input.CategoryID
	}
	if input.Title != nil {
		data["title"] = *input.Title
	}
	if input.Description != nil {
		data["description"] = *input.Description
	}

	result, err := h.DB.Update("service_tickets", "id", id, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var updated []models.ServiceTicket
	json.Unmarshal(result, &updated)
	if len(updated) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Ticket not found"})
	}

	return c.JSON(updated[0])
}

// AddServiceTicketComment adds a comment to a ticket
func (h *Handler) AddServiceTicketComment(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	ticketID := c.Params("id")
	var input struct {
		AuthorID   string `json:"author_id"`
		Content    string `json:"content"`
		IsInternal bool   `json:"is_internal"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.AuthorID == "" || input.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Author ID and content required"})
	}

	result, err := h.DB.Insert("service_ticket_comments", map[string]interface{}{
		"ticket_id":   ticketID,
		"author_id":   input.AuthorID,
		"content":     input.Content,
		"is_internal": input.IsInternal,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Log activity
	action := "comment_added"
	if input.IsInternal {
		action = "internal_note_added"
	}
	h.DB.Insert("service_ticket_activity", map[string]interface{}{
		"ticket_id": ticketID,
		"actor_id":  input.AuthorID,
		"action":    action,
	})

	var created []models.ServiceTicketComment
	json.Unmarshal(result, &created)
	if len(created) == 0 {
		return c.Status(201).JSON(fiber.Map{"status": "created"})
	}

	// Fetch with author info
	var comment models.ServiceTicketComment
	err = h.DB.From("service_ticket_comments").
		Select("*, author:employees(id, name, photo_base64)").
		Eq("id", created[0].ID).
		Single().
		Execute(&comment)
	if err != nil {
		return c.Status(201).JSON(created[0])
	}

	return c.Status(201).JSON(comment)
}

// ListServiceTicketCategories returns all categories
func (h *Handler) ListServiceTicketCategories(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var categories []models.ServiceTicketCategory
	err := h.DB.From("service_ticket_categories").Select("*").Order("name", false).Execute(&categories)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(categories)
}

// GetServiceDeskStats returns dashboard statistics
func (h *Handler) GetServiceDeskStats(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	// Get open tickets count
	var openTickets []models.ServiceTicket
	h.DB.From("service_tickets").Select("id, status, sla_deadline").In("status", []string{"new", "in_progress", "pending"}).Execute(&openTickets)

	openCount := len(openTickets)
	slaBreached := 0
	slaWarning := 0
	now := time.Now()

	for _, t := range openTickets {
		if t.SLADeadline != nil {
			timeLeft := t.SLADeadline.Sub(now)
			if timeLeft < 0 {
				slaBreached++
			} else if timeLeft < 2*time.Hour {
				slaWarning++
			}
		}
	}

	// Get resolved today
	todayStart := time.Now().Truncate(24 * time.Hour).Format("2006-01-02")
	var resolvedToday []models.ServiceTicket
	h.DB.From("service_tickets").Select("id").Eq("status", "resolved").Gte("resolved_at", todayStart).Execute(&resolvedToday)

	// Calculate SLA compliance (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	var recentClosed []struct {
		ID          string     `json:"id"`
		SLADeadline *time.Time `json:"sla_deadline"`
		ResolvedAt  *time.Time `json:"resolved_at"`
	}
	h.DB.From("service_tickets").Select("id, sla_deadline, resolved_at").Eq("status", "resolved").Gte("resolved_at", thirtyDaysAgo).Execute(&recentClosed)

	slaCompliance := 100
	if len(recentClosed) > 0 {
		withinSLA := 0
		for _, t := range recentClosed {
			if t.SLADeadline != nil && t.ResolvedAt != nil {
				if t.ResolvedAt.Before(*t.SLADeadline) || t.ResolvedAt.Equal(*t.SLADeadline) {
					withinSLA++
				}
			}
		}
		slaCompliance = (withinSLA * 100) / len(recentClosed)
	}

	return c.JSON(fiber.Map{
		"open_tickets":   openCount,
		"sla_breached":   slaBreached,
		"sla_warning":    slaWarning,
		"resolved_today": len(resolvedToday),
		"sla_compliance": slaCompliance,
	})
}

// GetMyServiceTickets returns tickets for current user
func (h *Handler) GetMyServiceTickets(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "user_id required"})
	}

	var tickets []models.ServiceTicket
	err := h.DB.From("service_tickets").
		Select("*, assignee:employees!service_tickets_assignee_id_fkey(id, name, photo_base64)").
		Eq("requester_id", userID).
		Order("created_at", true).
		Limit(10).
		Execute(&tickets)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tickets)
}
