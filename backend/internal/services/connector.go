package services

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

// ConnectorManager manages WebSocket connection to on-prem connector
type ConnectorManager struct {
	connector       *websocket.Conn
	pendingRequests map[string]chan ConnectorResponse
	connected       bool
	mutex           sync.RWMutex
	apiKey          string
}

// ConnectorCommand represents a command to send to the connector
type ConnectorCommand struct {
	Command   string                 `json:"command"`
	RequestID string                 `json:"request_id"`
	Params    map[string]interface{} `json:"params"`
}

// ConnectorResponse represents a response from the connector
type ConnectorResponse struct {
	Type      string                 `json:"type"`
	RequestID string                 `json:"request_id"`
	Command   string                 `json:"command"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// NewConnectorManager creates a new connector manager
func NewConnectorManager(apiKey string) *ConnectorManager {
	return &ConnectorManager{
		pendingRequests: make(map[string]chan ConnectorResponse),
		apiKey:          apiKey,
	}
}

// Connect handles a new connector connection
func (m *ConnectorManager) Connect(conn *websocket.Conn, apiKey string) error {
	if m.apiKey != "" && apiKey != m.apiKey {
		return errors.New("invalid API key")
	}

	m.mutex.Lock()
	m.connector = conn
	m.connected = true
	m.mutex.Unlock()

	return nil
}

// Disconnect handles connector disconnect
func (m *ConnectorManager) Disconnect() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.connector = nil
	m.connected = false

	// Fail all pending requests
	for _, ch := range m.pendingRequests {
		ch <- ConnectorResponse{
			Success: false,
			Error:   "Connector disconnected",
		}
		close(ch)
	}
	m.pendingRequests = make(map[string]chan ConnectorResponse)
}

// IsConnected returns connector status
func (m *ConnectorManager) IsConnected() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.connected
}

// SendCommand sends a command to the connector and waits for response
func (m *ConnectorManager) SendCommand(command string, params map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	m.mutex.RLock()
	if !m.connected || m.connector == nil {
		m.mutex.RUnlock()
		return nil, errors.New("connector not connected")
	}
	conn := m.connector
	m.mutex.RUnlock()

	requestID := uuid.New().String()
	responseChan := make(chan ConnectorResponse, 1)

	m.mutex.Lock()
	m.pendingRequests[requestID] = responseChan
	m.mutex.Unlock()

	defer func() {
		m.mutex.Lock()
		delete(m.pendingRequests, requestID)
		m.mutex.Unlock()
	}()

	cmd := ConnectorCommand{
		Command:   command,
		RequestID: requestID,
		Params:    params,
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return nil, err
	}

	select {
	case resp := <-responseChan:
		if !resp.Success {
			errMsg := resp.Error
			if errMsg == "" {
				if resp.Result != nil {
					if e, ok := resp.Result["error"].(string); ok {
						errMsg = e
					}
				}
			}
			if errMsg == "" {
				errMsg = "Unknown error"
			}
			return nil, errors.New(errMsg)
		}
		return resp.Result, nil
	case <-time.After(timeout):
		return nil, errors.New("request timeout")
	}
}

// HandleResponse processes a response from the connector
func (m *ConnectorManager) HandleResponse(data []byte) {
	var resp ConnectorResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}

	m.mutex.RLock()
	ch, exists := m.pendingRequests[resp.RequestID]
	m.mutex.RUnlock()

	if exists {
		ch <- resp
	}
}

// HandleMessage processes incoming messages from connector
func (m *ConnectorManager) HandleMessage(messageType int, data []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	msgType, _ := msg["type"].(string)

	switch msgType {
	case "heartbeat":
		// Just acknowledge, no action needed
	case "response":
		m.HandleResponse(data)
	}
}
