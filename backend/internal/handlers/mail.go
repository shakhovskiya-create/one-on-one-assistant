package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// GetMailFolders returns user's mail folders
func (h *Handler) GetMailFolders(c *fiber.Ctx) error {
	if h.EWS == nil {
		return c.Status(503).JSON(fiber.Map{"error": "EWS not configured"})
	}

	username := c.Query("username")
	password := c.Query("password")

	if username == "" || password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username and password required"})
	}

	folders, err := h.EWS.GetMailFolders("", username, password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(folders)
}

// GetEmails returns emails from a folder
func (h *Handler) GetEmails(c *fiber.Ctx) error {
	if h.EWS == nil {
		return c.Status(503).JSON(fiber.Map{"error": "EWS not configured"})
	}

	username := c.Query("username")
	password := c.Query("password")
	folderID := c.Query("folder_id")
	limit := c.QueryInt("limit", 50)

	if username == "" || password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username and password required"})
	}

	emails, err := h.EWS.GetEmails("", username, password, folderID, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(emails)
}

// SendMailRequest represents the request body for sending email
type SendMailRequest struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	To       []string `json:"to"`
	CC       []string `json:"cc"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
}

// SendEmail sends an email
func (h *Handler) SendEmail(c *fiber.Ctx) error {
	if h.EWS == nil {
		return c.Status(503).JSON(fiber.Map{"error": "EWS not configured"})
	}

	var req SendMailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username and password required"})
	}

	if len(req.To) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "at least one recipient required"})
	}

	if err := h.EWS.SendEmail(req.Username, req.Password, req.Subject, req.To, req.Body, req.CC); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Email sent"})
}

// MarkEmailReadRequest represents the request body for marking email as read
type MarkEmailReadRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	ItemID    string `json:"item_id"`
	ChangeKey string `json:"change_key"`
}

// MarkEmailAsRead marks an email as read
func (h *Handler) MarkEmailAsRead(c *fiber.Ctx) error {
	if h.EWS == nil {
		return c.Status(503).JSON(fiber.Map{"error": "EWS not configured"})
	}

	var req MarkEmailReadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Username == "" || req.Password == "" || req.ItemID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username, password, and item_id required"})
	}

	if err := h.EWS.MarkEmailAsRead(req.Username, req.Password, req.ItemID, req.ChangeKey); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

// DeleteEmailRequest represents the request body for deleting email
type DeleteEmailRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	ItemID    string `json:"item_id"`
	ChangeKey string `json:"change_key"`
}

// DeleteEmail moves an email to deleted items
func (h *Handler) DeleteEmail(c *fiber.Ctx) error {
	if h.EWS == nil {
		return c.Status(503).JSON(fiber.Map{"error": "EWS not configured"})
	}

	var req DeleteEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Username == "" || req.Password == "" || req.ItemID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username, password, and item_id required"})
	}

	if err := h.EWS.DeleteEmail(req.Username, req.Password, req.ItemID, req.ChangeKey); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

// GetEmailBodyRequest represents the request body for getting email body
type GetEmailBodyRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	ItemID    string `json:"item_id"`
	ChangeKey string `json:"change_key"`
}

// GetEmailBody returns the full body of an email
func (h *Handler) GetEmailBody(c *fiber.Ctx) error {
	if h.EWS == nil {
		return c.Status(503).JSON(fiber.Map{"error": "EWS not configured"})
	}

	var req GetEmailBodyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Username == "" || req.Password == "" || req.ItemID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username, password, and item_id required"})
	}

	body, err := h.EWS.GetEmailBody(req.Username, req.Password, req.ItemID, req.ChangeKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"body": body})
}
