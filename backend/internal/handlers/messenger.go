package handlers

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/ekf/one-on-one-backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// MessengerHub manages WebSocket connections
type MessengerHub struct {
	clients    map[string]map[*websocket.Conn]bool // userID -> connections
	broadcast  chan WSMessage
	register   chan *WSClient
	unregister chan *WSClient
	mu         sync.RWMutex
}

type WSClient struct {
	Conn   *websocket.Conn
	UserID string
}

type WSMessage struct {
	Type           string      `json:"type"`
	ConversationID string      `json:"conversation_id,omitempty"`
	Message        interface{} `json:"message,omitempty"`
	Data           interface{} `json:"data,omitempty"`
	Recipients     []string    `json:"-"`
}

var hub = &MessengerHub{
	clients:    make(map[string]map[*websocket.Conn]bool),
	broadcast:  make(chan WSMessage, 256),
	register:   make(chan *WSClient),
	unregister: make(chan *WSClient),
}

func init() {
	go hub.run()
}

func (h *MessengerHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*websocket.Conn]bool)
			}
			h.clients[client.UserID][client.Conn] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if conns, ok := h.clients[client.UserID]; ok {
				delete(conns, client.Conn)
				if len(conns) == 0 {
					delete(h.clients, client.UserID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			data, _ := json.Marshal(msg)
			for _, userID := range msg.Recipients {
				if conns, ok := h.clients[userID]; ok {
					for conn := range conns {
						conn.WriteMessage(websocket.TextMessage, data)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// MessengerWebSocket handles WebSocket connections for messenger
func (h *Handler) MessengerWebSocket(conn *websocket.Conn) {
	userID := conn.Query("user_id")
	tokenString := conn.Query("token")
	if userID == "" {
		conn.Close()
		return
	}
	if tokenString == "" {
		conn.Close()
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	claims, err := h.JWT.ValidateToken(tokenString)
	if err != nil || claims.UserID != userID {
		conn.Close()
		return
	}

	client := &WSClient{Conn: conn, UserID: userID}
	hub.register <- client

	defer func() {
		hub.unregister <- client
		conn.Close()
	}()

	// Send connection confirmation
	conn.WriteJSON(WSMessage{Type: "connected", Data: map[string]string{"user_id": userID}})

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg struct {
			Type           string `json:"type"`
			ConversationID string `json:"conversation_id"`
			Content        string `json:"content"`
			ReplyToID      string `json:"reply_to_id"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "send_message":
			h.handleSendMessage(userID, msg.ConversationID, msg.Content, msg.ReplyToID)
		case "typing":
			h.handleTyping(userID, msg.ConversationID)
		case "read":
			h.handleMarkRead(userID, msg.ConversationID)
		}
	}
}

func (h *Handler) handleSendMessage(senderID, conversationID, content, replyToID string) {
	if h.DB == nil || content == "" || conversationID == "" {
		return
	}

	// Create message
	msgData := map[string]interface{}{
		"conversation_id": conversationID,
		"sender_id":       senderID,
		"content":         content,
		"message_type":    "text",
	}
	if replyToID != "" {
		msgData["reply_to_id"] = replyToID
	}

	result, err := h.DB.Insert("messages", msgData)
	if err != nil {
		return
	}

	var created []models.Message
	json.Unmarshal(result, &created)
	if len(created) == 0 {
		return
	}

	newMsg := created[0]

	// Get sender info
	var sender models.Employee
	h.DB.From("employees").Select("id, name, photo_base64").Eq("id", senderID).Single().Execute(&sender)
	newMsg.Sender = &sender

	// Update conversation
	h.DB.Update("conversations", "id", conversationID, map[string]interface{}{
		"updated_at": time.Now().Format(time.RFC3339),
	})

	// Get participants
	var participants []struct {
		EmployeeID string `json:"employee_id"`
	}
	h.DB.From("conversation_participants").Select("employee_id").Eq("conversation_id", conversationID).Execute(&participants)

	recipients := make([]string, 0, len(participants))
	for _, p := range participants {
		recipients = append(recipients, p.EmployeeID)
	}

	// Broadcast message
	hub.broadcast <- WSMessage{
		Type:           "new_message",
		ConversationID: conversationID,
		Message:        newMsg,
		Recipients:     recipients,
	}
}

func (h *Handler) handleTyping(userID, conversationID string) {
	if conversationID == "" {
		return
	}

	var participants []struct {
		EmployeeID string `json:"employee_id"`
	}
	h.DB.From("conversation_participants").Select("employee_id").Eq("conversation_id", conversationID).Execute(&participants)

	recipients := make([]string, 0)
	for _, p := range participants {
		if p.EmployeeID != userID {
			recipients = append(recipients, p.EmployeeID)
		}
	}

	hub.broadcast <- WSMessage{
		Type:           "typing",
		ConversationID: conversationID,
		Data:           map[string]string{"user_id": userID},
		Recipients:     recipients,
	}
}

func (h *Handler) handleMarkRead(userID, conversationID string) {
	if h.DB == nil || conversationID == "" {
		return
	}

	h.DB.Update("conversation_participants", "conversation_id", conversationID,
		map[string]interface{}{
			"last_read_at": time.Now().Format(time.RFC3339),
		})
}

// ListConversations returns all conversations for a user
func (h *Handler) ListConversations(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Not authenticated"})
	}

	// Get conversation IDs for user
	var participantRecords []struct {
		ConversationID string `json:"conversation_id"`
	}
	h.DB.From("conversation_participants").Select("conversation_id").Eq("employee_id", userID).Execute(&participantRecords)

	if len(participantRecords) == 0 {
		return c.JSON([]models.Conversation{})
	}

	convIDs := make([]string, len(participantRecords))
	for i, p := range participantRecords {
		convIDs[i] = p.ConversationID
	}

	// Get conversations
	var conversations []models.Conversation
	h.DB.From("conversations").Select("*").In("id", convIDs).Order("updated_at", true).Execute(&conversations)

	// Enrich with participants and last message
	for i := range conversations {
		var participants []struct {
			EmployeeID string          `json:"employee_id"`
			Employee   models.Employee `json:"employees"`
		}
		h.DB.From("conversation_participants").
			Select("employee_id, employees(id, name, photo_base64, position)").
			Eq("conversation_id", conversations[i].ID).Execute(&participants)

		conversations[i].Participants = make([]models.Employee, len(participants))
		for j, p := range participants {
			conversations[i].Participants[j] = p.Employee
		}

		// Get last message
		var messages []models.Message
		h.DB.From("messages").Select("*, employees:sender_id(id, name, photo_base64)").
			Eq("conversation_id", conversations[i].ID).
			Order("created_at", true).
			Limit(1).Execute(&messages)
		if len(messages) > 0 {
			// Populate file URL for media message
			if messages[0].FileID != nil && *messages[0].FileID != "" && h.Storage != nil {
				var files []FileMetadata
				h.DB.From("files").Select("storage_path").Eq("id", *messages[0].FileID).Execute(&files)
				if len(files) > 0 {
					messages[0].FileURL = h.Storage.GetPublicURL(files[0].StoragePath)
				}
			}
			conversations[i].LastMessage = &messages[0]
		}
	}

	return c.JSON(conversations)
}

// GetConversation returns a single conversation with messages
func (h *Handler) GetConversation(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")
	userID, _ := c.Locals("user_id").(string)
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var conversation models.Conversation
	err := h.DB.From("conversations").Select("*").Eq("id", id).Single().Execute(&conversation)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Conversation not found"})
	}

	// Get participants
	var participantsData []struct {
		EmployeeID string          `json:"employee_id"`
		Employee   models.Employee `json:"employees"`
	}
	h.DB.From("conversation_participants").
		Select("employee_id, employees(id, name, photo_base64, position)").
		Eq("conversation_id", id).Execute(&participantsData)

	// Access control: verify user is a participant
	isParticipant := false
	for _, p := range participantsData {
		if p.EmployeeID == userID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied: not a participant"})
	}

	participants := make([]models.Employee, len(participantsData))
	for i, p := range participantsData {
		participants[i] = p.Employee
	}

	// Get messages
	var messages []models.Message
	h.DB.From("messages").
		Select("*, employees:sender_id(id, name, photo_base64)").
		Eq("conversation_id", id).
		Order("created_at", true).
		Limit(limit).
		Offset(offset).
		Execute(&messages)

	// Reverse for chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	// Populate file URLs for media messages
	if h.Storage != nil {
		for i := range messages {
			if messages[i].FileID != nil && *messages[i].FileID != "" {
				var files []FileMetadata
				h.DB.From("files").Select("storage_path").Eq("id", *messages[i].FileID).Execute(&files)
				if len(files) > 0 {
					messages[i].FileURL = h.Storage.GetPublicURL(files[0].StoragePath)
				}
			}
		}
	}

	return c.JSON(fiber.Map{
		"conversation": conversation,
		"participants": participants,
		"messages":     messages,
	})
}

// CreateConversation creates a new conversation
func (h *Handler) CreateConversation(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var req struct {
		Type         string   `json:"type"`
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Participants []string `json:"participants"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Channels can have just the creator
	minParticipants := 2
	if req.Type == "channel" {
		minParticipants = 1
	}
	if len(req.Participants) < minParticipants {
		return c.Status(400).JSON(fiber.Map{"error": "At least 2 participants required for chat"})
	}
	participantMap := make(map[string]bool, len(req.Participants))
	for _, id := range req.Participants {
		participantMap[id] = true
	}
	if !participantMap[userID] {
		return c.Status(403).JSON(fiber.Map{"error": "Creator must be a participant"})
	}

	convType := req.Type
	if convType == "" {
		if len(req.Participants) == 2 {
			convType = "direct"
		} else {
			convType = "group"
		}
	}

	// For direct conversations, check if one already exists
	if convType == "direct" && len(req.Participants) == 2 {
		var existing []struct {
			ConversationID string `json:"conversation_id"`
		}
		// This is a simplified check - in production you'd want a proper query
		h.DB.From("conversation_participants").
			Select("conversation_id").
			Eq("employee_id", req.Participants[0]).
			Execute(&existing)

		for _, e := range existing {
			var otherParticipants []struct {
				EmployeeID string `json:"employee_id"`
			}
			h.DB.From("conversation_participants").
				Select("employee_id").
				Eq("conversation_id", e.ConversationID).
				Execute(&otherParticipants)

			if len(otherParticipants) == 2 {
				hasOther := false
				for _, op := range otherParticipants {
					if op.EmployeeID == req.Participants[1] {
						hasOther = true
						break
					}
				}
				if hasOther {
					// Return existing conversation
					var conv models.Conversation
					h.DB.From("conversations").Select("*").Eq("id", e.ConversationID).Eq("type", "direct").Single().Execute(&conv)
					if conv.ID != "" {
						return c.JSON(conv)
					}
				}
			}
		}
	}

	// Create conversation
	convData := map[string]interface{}{
		"type": convType,
	}
	if req.Name != "" {
		convData["name"] = req.Name
	}
	if req.Description != "" {
		convData["description"] = req.Description
	}
	if convType == "channel" {
		convData["created_by"] = userID
	}

	result, err := h.DB.Insert("conversations", convData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var created []models.Conversation
	json.Unmarshal(result, &created)
	if len(created) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create conversation"})
	}

	conv := created[0]

	// Add participants
	for _, pID := range req.Participants {
		participantData := map[string]interface{}{
			"conversation_id": conv.ID,
			"employee_id":     pID,
		}
		// Channel creator is admin
		if convType == "channel" && pID == userID {
			participantData["is_admin"] = true
		}
		h.DB.Insert("conversation_participants", participantData)
	}

	return c.Status(201).JSON(conv)
}

// SendMessage sends a message via HTTP (alternative to WebSocket)
func (h *Handler) SendMessage(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Not authenticated"})
	}

	var req struct {
		ConversationID  string `json:"conversation_id"`
		SenderID        string `json:"sender_id"`
		Content         string `json:"content"`
		MessageType     string `json:"message_type"`
		FileID          string `json:"file_id"`
		DurationSeconds int    `json:"duration_seconds"`
		ReplyToID       string `json:"reply_to_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Voice and video messages don't require content (the file IS the content)
	isMediaMessage := req.MessageType == "voice" || req.MessageType == "video"
	if req.ConversationID == "" || req.SenderID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "conversation_id and sender_id required"})
	}
	if !isMediaMessage && req.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "content required for text messages"})
	}
	if isMediaMessage && req.FileID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "file_id required for media messages"})
	}
	if req.SenderID != userID {
		return c.Status(403).JSON(fiber.Map{"error": "Sender mismatch"})
	}

	// Access control: verify sender is a participant
	var participantCheck []struct {
		EmployeeID string `json:"employee_id"`
	}
	h.DB.From("conversation_participants").
		Select("employee_id").
		Eq("conversation_id", req.ConversationID).
		Eq("employee_id", req.SenderID).
		Execute(&participantCheck)

	if len(participantCheck) == 0 {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied: not a participant"})
	}

	// Determine message type
	msgType := req.MessageType
	if msgType == "" {
		msgType = "text"
	}

	// Create message
	msgData := map[string]interface{}{
		"conversation_id": req.ConversationID,
		"sender_id":       req.SenderID,
		"content":         req.Content,
		"message_type":    msgType,
	}
	if req.FileID != "" {
		msgData["file_id"] = req.FileID
	}
	if req.DurationSeconds > 0 {
		msgData["duration_seconds"] = req.DurationSeconds
	}
	if req.ReplyToID != "" {
		msgData["reply_to_id"] = req.ReplyToID
	}

	result, err := h.DB.Insert("messages", msgData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var created []models.Message
	json.Unmarshal(result, &created)
	if len(created) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create message"})
	}

	newMsg := created[0]

	// Get sender info
	var sender models.Employee
	h.DB.From("employees").Select("id, name, photo_base64").Eq("id", req.SenderID).Single().Execute(&sender)
	newMsg.Sender = &sender

	// Populate file URL for media messages
	if newMsg.FileID != nil && *newMsg.FileID != "" && h.Storage != nil {
		var files []FileMetadata
		h.DB.From("files").Select("storage_path").Eq("id", *newMsg.FileID).Execute(&files)
		if len(files) > 0 {
			newMsg.FileURL = h.Storage.GetPublicURL(files[0].StoragePath)
		}
	}

	// Update conversation
	h.DB.Update("conversations", "id", req.ConversationID, map[string]interface{}{
		"updated_at": time.Now().Format(time.RFC3339),
	})

	// Broadcast via WebSocket to OTHER participants (not the sender - they get the response)
	var participants []struct {
		EmployeeID string `json:"employee_id"`
	}
	h.DB.From("conversation_participants").Select("employee_id").Eq("conversation_id", req.ConversationID).Execute(&participants)

	recipients := make([]string, 0, len(participants))
	for _, p := range participants {
		// Exclude sender - they already get the message from the HTTP response
		if p.EmployeeID != req.SenderID {
			recipients = append(recipients, p.EmployeeID)
		}
	}

	if len(recipients) > 0 {
		hub.broadcast <- WSMessage{
			Type:           "new_message",
			ConversationID: req.ConversationID,
			Message:        newMsg,
			Recipients:     recipients,
		}
	}

	return c.Status(201).JSON(newMsg)
}
