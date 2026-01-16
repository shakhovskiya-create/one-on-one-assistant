package connector

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ekf/one-on-one-connector/internal/ews"
	"github.com/ekf/one-on-one-connector/pkg/protocol"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Backend struct {
		URL                string `yaml:"url"`
		APIKey             string `yaml:"api_key"`
		ReconnectInterval  int    `yaml:"reconnect_interval"`
		HeartbeatInterval  int    `yaml:"heartbeat_interval"`
		InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
	} `yaml:"backend"`

	EWS struct {
		URL        string `yaml:"url"`
		Domain     string `yaml:"domain"`
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
		SkipVerify bool   `yaml:"skip_verify"`
	} `yaml:"ews"`
}

type Connector struct {
	config    *Config
	ws        *websocket.Conn
	ewsClient *ews.Client
	running   bool
	stopChan  chan struct{}
}

func New(configPath string) (*Connector, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ewsClient := ews.NewClient(config.EWS.URL, config.EWS.Domain, config.EWS.SkipVerify)

	return &Connector{
		config:    config,
		ewsClient: ewsClient,
		stopChan:  make(chan struct{}),
	}, nil
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Expand environment variables
	content := os.ExpandEnv(string(data))

	var config Config
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Connector) connect() error {
	u, err := url.Parse(c.config.Backend.URL)
	if err != nil {
		return fmt.Errorf("invalid backend URL: %w", err)
	}

	// Add API key as query parameter
	q := u.Query()
	q.Set("token", c.config.Backend.APIKey)
	u.RawQuery = q.Encode()

	dialer := websocket.DefaultDialer
	if c.config.Backend.InsecureSkipVerify {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	ws, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.ws = ws
	log.Printf("Connected to backend: %s", strings.Split(c.config.Backend.URL, "?")[0])
	return nil
}

func (c *Connector) disconnect() {
	if c.ws != nil {
		c.ws.Close()
		c.ws = nil
		log.Println("Disconnected from backend")
	}
}

func (c *Connector) sendMessage(msg interface{}) error {
	if c.ws == nil {
		return fmt.Errorf("not connected")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return c.ws.WriteMessage(websocket.TextMessage, data)
}

func (c *Connector) handleMessage(data []byte) {
	var cmd protocol.Command
	if err := json.Unmarshal(data, &cmd); err != nil {
		log.Printf("Failed to parse command: %v", err)
		return
	}

	log.Printf("Received command: %s (request_id: %s)", cmd.Command, cmd.RequestID)

	var result interface{}
	var cmdErr error

	switch cmd.Command {
	case "ping":
		result = map[string]interface{}{
			"pong":      true,
			"timestamp": time.Now().Format(time.RFC3339),
		}

	case "get_calendar":
		result, cmdErr = c.handleGetCalendar(cmd.Params)

	case "sync_calendar":
		result, cmdErr = c.handleSyncCalendar(cmd.Params)

	case "find_free_slots":
		result, cmdErr = c.handleFindFreeSlots(cmd.Params)

	default:
		cmdErr = fmt.Errorf("unknown command: %s", cmd.Command)
	}

	// Send response
	resp := protocol.Response{
		Type:      "response",
		RequestID: cmd.RequestID,
		Command:   cmd.Command,
		Success:   cmdErr == nil,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if cmdErr != nil {
		resp.Error = cmdErr.Error()
	} else {
		resp.Result = result
	}

	if err := c.sendMessage(resp); err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}

func (c *Connector) handleGetCalendar(params map[string]interface{}) (interface{}, error) {
	email, ok := params["email"].(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("email is required")
	}

	username, _ := params["username"].(string)
	password, _ := params["password"].(string)
	daysBack := int(getFloat64(params, "days_back", 7))
	daysForward := int(getFloat64(params, "days_forward", 30))

	// Use connector credentials if not provided
	if username == "" {
		username = c.config.EWS.Username
	}
	if password == "" {
		password = c.config.EWS.Password
	}

	events, err := c.ewsClient.GetCalendarEvents(email, username, password, daysBack, daysForward)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar: %w", err)
	}

	return events, nil
}

func (c *Connector) handleSyncCalendar(params map[string]interface{}) (interface{}, error) {
	// Same as get_calendar for now
	return c.handleGetCalendar(params)
}

func (c *Connector) handleFindFreeSlots(params map[string]interface{}) (interface{}, error) {
	// TODO: Implement free/busy time slots
	return map[string]interface{}{
		"slots": []interface{}{},
	}, nil
}

func (c *Connector) heartbeatLoop() {
	interval := time.Duration(c.config.Backend.HeartbeatInterval) * time.Second
	if interval == 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.sendMessage(map[string]interface{}{
				"type":      "heartbeat",
				"timestamp": time.Now().Format(time.RFC3339),
				"status":    "online",
			}); err != nil {
				log.Printf("Heartbeat error: %v", err)
				return
			}

		case <-c.stopChan:
			return
		}
	}
}

func (c *Connector) messageLoop() {
	for c.running {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			return
		}

		c.handleMessage(message)
	}
}

func (c *Connector) Run() error {
	c.running = true
	reconnectInterval := time.Duration(c.config.Backend.ReconnectInterval) * time.Second
	if reconnectInterval == 0 {
		reconnectInterval = 5 * time.Second
	}

	log.Println("Starting connector...")
	log.Printf("EWS URL: %s", c.config.EWS.URL)
	log.Printf("Backend URL: %s", strings.Split(c.config.Backend.URL, "?")[0])

	for c.running {
		if err := c.connect(); err != nil {
			log.Printf("Connection failed: %v", err)
			if c.running {
				log.Printf("Reconnecting in %v...", reconnectInterval)
				time.Sleep(reconnectInterval)
			}
			continue
		}

		// Run loops
		go c.heartbeatLoop()
		c.messageLoop()

		c.disconnect()

		if c.running {
			log.Printf("Reconnecting in %v...", reconnectInterval)
			time.Sleep(reconnectInterval)
		}
	}

	return nil
}

func (c *Connector) Stop() {
	log.Println("Stopping connector...")
	c.running = false
	close(c.stopChan)
	c.disconnect()
}

func getFloat64(m map[string]interface{}, key string, defaultValue float64) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return defaultValue
}
