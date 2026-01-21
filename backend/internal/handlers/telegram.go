package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ekf/one-on-one-backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// TelegramWebhookUpdate represents an incoming Telegram webhook update
type TelegramWebhookUpdate struct {
	UpdateID int64                   `json:"update_id"`
	Message  *TelegramWebhookMessage `json:"message"`
}

// TelegramWebhookMessage represents a Telegram message in webhook
type TelegramWebhookMessage struct {
	MessageID int64                `json:"message_id"`
	Chat      *TelegramWebhookChat `json:"chat"`
	From      *TelegramWebhookUser `json:"from"`
	Text      string               `json:"text"`
	Date      int64                `json:"date"`
}

// TelegramWebhookChat represents chat info
type TelegramWebhookChat struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

// TelegramWebhookUser represents user info
type TelegramWebhookUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// TelegramWebhook handles incoming Telegram webhook calls
func (h *Handler) TelegramWebhook(c *fiber.Ctx) error {
	channelID := c.Params("channel_id")
	if channelID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "channel_id required"})
	}

	// Parse webhook update
	var update TelegramWebhookUpdate
	if err := c.BodyParser(&update); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid update"})
	}

	if update.Message == nil || update.Message.Text == "" {
		return c.SendStatus(200) // Ignore non-text messages
	}

	// Verify channel exists and has Telegram enabled
	var conv struct {
		ID              string `json:"id"`
		TelegramChatID  int64  `json:"telegram_chat_id"`
		TelegramEnabled bool   `json:"telegram_enabled"`
	}
	err := h.DB.From("conversations").Select("id, telegram_chat_id, telegram_enabled").
		Eq("id", channelID).Single().Execute(&conv)
	if err != nil || !conv.TelegramEnabled {
		return c.Status(404).JSON(fiber.Map{"error": "channel not found or telegram disabled"})
	}

	// Verify chat ID matches (security check)
	if conv.TelegramChatID != update.Message.Chat.ID {
		return c.Status(403).JSON(fiber.Map{"error": "chat ID mismatch"})
	}

	// Format sender name
	senderName := update.Message.From.FirstName
	if update.Message.From.LastName != "" {
		senderName += " " + update.Message.From.LastName
	}
	if update.Message.From.Username != "" {
		senderName = fmt.Sprintf("%s (@%s)", senderName, update.Message.From.Username)
	}

	// Create message in channel (as system message from Telegram)
	content := fmt.Sprintf("[TG] %s: %s", senderName, update.Message.Text)

	msgData := map[string]interface{}{
		"conversation_id": channelID,
		"content":         content,
		"message_type":    "system",
	}

	result, err := h.DB.Insert("messages", msgData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create message"})
	}

	var created []models.Message
	json.Unmarshal(result, &created)

	// Update conversation timestamp
	h.DB.Update("conversations", "id", channelID, map[string]interface{}{
		"updated_at": time.Now().Format(time.RFC3339),
	})

	// Broadcast to WebSocket
	if len(created) > 0 {
		var participants []struct {
			EmployeeID string `json:"employee_id"`
		}
		h.DB.From("conversation_participants").Select("employee_id").
			Eq("conversation_id", channelID).Execute(&participants)

		recipients := make([]string, len(participants))
		for i, p := range participants {
			recipients[i] = p.EmployeeID
		}

		hub.broadcast <- WSMessage{
			Type:           "new_message",
			ConversationID: channelID,
			Message:        created[0],
			Recipients:     recipients,
		}
	}

	return c.SendStatus(200)
}

// ConfigureTelegramBot configures Telegram bot for a channel
func (h *Handler) ConfigureTelegramBot(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Not authenticated"})
	}

	channelID := c.Params("channel_id")

	var req struct {
		BotToken string `json:"bot_token"`
		ChatID   int64  `json:"chat_id"`
		Enabled  bool   `json:"enabled"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Verify user is channel creator/admin
	var conv struct {
		ID        string  `json:"id"`
		Type      string  `json:"type"`
		CreatedBy *string `json:"created_by"`
	}
	err := h.DB.From("conversations").Select("id, type, created_by").
		Eq("id", channelID).Single().Execute(&conv)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Channel not found"})
	}

	if conv.Type != "channel" {
		return c.Status(400).JSON(fiber.Map{"error": "Telegram integration only available for channels"})
	}

	if conv.CreatedBy == nil || *conv.CreatedBy != userID {
		return c.Status(403).JSON(fiber.Map{"error": "Only channel creator can configure Telegram"})
	}

	// Update channel with Telegram settings
	updateData := map[string]interface{}{
		"telegram_enabled": req.Enabled,
	}
	if req.BotToken != "" {
		updateData["telegram_bot_token"] = req.BotToken
	}
	if req.ChatID != 0 {
		updateData["telegram_chat_id"] = req.ChatID
	}

	_, err = h.DB.Update("conversations", "id", channelID, updateData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update channel"})
	}

	// Generate webhook URL
	webhookURL := fmt.Sprintf("%s/api/v1/telegram/webhook/%s", c.BaseURL(), channelID)

	return c.JSON(fiber.Map{
		"success":     true,
		"webhook_url": webhookURL,
		"message":     "Configure this webhook URL in your Telegram bot settings",
	})
}

// GetTelegramConfig returns Telegram configuration for a channel
func (h *Handler) GetTelegramConfig(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Not authenticated"})
	}

	channelID := c.Params("channel_id")

	var conv struct {
		ID              string  `json:"id"`
		Type            string  `json:"type"`
		CreatedBy       *string `json:"created_by"`
		TelegramEnabled bool    `json:"telegram_enabled"`
		TelegramChatID  *int64  `json:"telegram_chat_id"`
	}
	err := h.DB.From("conversations").Select("id, type, created_by, telegram_enabled, telegram_chat_id").
		Eq("id", channelID).Single().Execute(&conv)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Channel not found"})
	}

	// Only creator can see config
	if conv.CreatedBy == nil || *conv.CreatedBy != userID {
		return c.Status(403).JSON(fiber.Map{"error": "Only channel creator can view Telegram config"})
	}

	webhookURL := fmt.Sprintf("%s/api/v1/telegram/webhook/%s", c.BaseURL(), channelID)

	return c.JSON(fiber.Map{
		"enabled":     conv.TelegramEnabled,
		"chat_id":     conv.TelegramChatID,
		"webhook_url": webhookURL,
	})
}

// ForwardToTelegram forwards a message from EKF Hub to a linked Telegram chat
func (h *Handler) ForwardToTelegram(conversationID, senderName, content, messageType string) error {
	if h.DB == nil {
		return fmt.Errorf("database not configured")
	}

	// Get conversation Telegram settings
	var conv struct {
		ID               string  `json:"id"`
		Type             string  `json:"type"`
		TelegramEnabled  bool    `json:"telegram_enabled"`
		TelegramBotToken *string `json:"telegram_bot_token"`
		TelegramChatID   *int64  `json:"telegram_chat_id"`
	}
	err := h.DB.From("conversations").
		Select("id, type, telegram_enabled, telegram_bot_token, telegram_chat_id").
		Eq("id", conversationID).Single().Execute(&conv)
	if err != nil {
		return nil // Conversation not found, not an error for forwarding
	}

	// Check if Telegram is enabled and configured
	if !conv.TelegramEnabled || conv.TelegramBotToken == nil || conv.TelegramChatID == nil {
		return nil // Not configured, silently skip
	}

	if *conv.TelegramBotToken == "" || *conv.TelegramChatID == 0 {
		return nil // Not properly configured
	}

	// Format message based on type
	var text string
	switch messageType {
	case "voice":
		text = fmt.Sprintf("🎤 *%s* отправил голосовое сообщение", senderName)
	case "video":
		text = fmt.Sprintf("📹 *%s* отправил видеосообщение", senderName)
	case "file":
		text = fmt.Sprintf("📎 *%s* отправил файл", senderName)
	case "gif":
		text = fmt.Sprintf("🎬 *%s* отправил GIF", senderName)
	default:
		text = fmt.Sprintf("*%s*: %s", senderName, content)
	}

	// Send to Telegram
	telegramURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", *conv.TelegramBotToken)

	payload := map[string]interface{}{
		"chat_id":    *conv.TelegramChatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	jsonPayload, _ := json.Marshal(payload)
	resp, err := http.Post(telegramURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send to Telegram: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Telegram API returned status %d", resp.StatusCode)
	}

	return nil
}

// SendTelegramNotification sends a notification to a user via Telegram bot
func (h *Handler) SendTelegramNotification(userID, message string) error {
	if h.DB == nil || h.Config.TelegramBotToken == "" {
		return nil // Not configured
	}

	// Get user's Telegram chat ID (if linked)
	var employee struct {
		TelegramChatID *int64 `json:"telegram_chat_id"`
	}
	err := h.DB.From("employees").Select("telegram_chat_id").
		Eq("id", userID).Single().Execute(&employee)
	if err != nil || employee.TelegramChatID == nil {
		return nil // User doesn't have Telegram linked
	}

	// Send notification
	telegramURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", h.Config.TelegramBotToken)

	payload := map[string]interface{}{
		"chat_id":    *employee.TelegramChatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	jsonPayload, _ := json.Marshal(payload)
	resp, err := http.Post(telegramURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
