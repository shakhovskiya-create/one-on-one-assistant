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

// GetAttachmentsRequest represents the request body for getting attachments
type GetAttachmentsRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	ItemID    string `json:"item_id"`
	ChangeKey string `json:"change_key"`
}

// GetAttachments returns the list of attachments for an email
func (h *Handler) GetAttachments(c *fiber.Ctx) error {
	if h.EWS == nil {
		return c.Status(503).JSON(fiber.Map{"error": "EWS not configured"})
	}

	var req GetAttachmentsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Username == "" || req.Password == "" || req.ItemID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username, password, and item_id required"})
	}

	attachments, err := h.EWS.GetAttachments(req.Username, req.Password, req.ItemID, req.ChangeKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"attachments": attachments})
}

// GetAttachmentContentRequest represents the request for downloading attachment
type GetAttachmentContentRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	AttachmentID string `json:"attachment_id"`
}

// GetAttachmentContent returns the content of a specific attachment
func (h *Handler) GetAttachmentContent(c *fiber.Ctx) error {
	if h.EWS == nil {
		return c.Status(503).JSON(fiber.Map{"error": "EWS not configured"})
	}

	var req GetAttachmentContentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Username == "" || req.Password == "" || req.AttachmentID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username, password, and attachment_id required"})
	}

	name, contentType, content, err := h.EWS.GetAttachmentContent(req.Username, req.Password, req.AttachmentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"name":         name,
		"content_type": contentType,
		"content":      string(content), // Base64 encoded content
	})
}

// RespondToMeetingRequest represents the request for responding to meeting invitation
type RespondToMeetingRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	ItemID    string `json:"item_id"`
	ChangeKey string `json:"change_key"`
	Response  string `json:"response"` // Accept, Decline, Tentative
}

// RespondToMeeting handles Accept/Decline/Tentative for meeting invitations
func (h *Handler) RespondToMeeting(c *fiber.Ctx) error {
	if h.EWS == nil {
		return c.Status(503).JSON(fiber.Map{"error": "EWS not configured"})
	}

	var req RespondToMeetingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Username == "" || req.Password == "" || req.ItemID == "" || req.Response == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username, password, item_id, and response required"})
	}

	// Validate response type
	if req.Response != "Accept" && req.Response != "Decline" && req.Response != "Tentative" {
		return c.Status(400).JSON(fiber.Map{"error": "response must be Accept, Decline, or Tentative"})
	}

	err := h.EWS.RespondToMeetingRequest(req.Username, req.Password, req.ItemID, req.ChangeKey, req.Response)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "response": req.Response})
}
