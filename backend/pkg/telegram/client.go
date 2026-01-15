package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client handles Telegram Bot API
type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient creates a new Telegram client
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{},
	}
}

// SendMessage sends a message to a Telegram chat
func (c *Client) SendMessage(chatID int64, text string) error {
	if c.token == "" {
		return nil
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.token)

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram error: %d", resp.StatusCode)
	}

	return nil
}

// WebhookUpdate represents an incoming Telegram webhook update
type WebhookUpdate struct {
	Message *Message `json:"message"`
}

// Message represents a Telegram message
type Message struct {
	Chat *Chat  `json:"chat"`
	Text string `json:"text"`
	From *User  `json:"from"`
}

// Chat represents a Telegram chat
type Chat struct {
	ID int64 `json:"id"`
}

// User represents a Telegram user
type User struct {
	Username string `json:"username"`
}
