package protocol

// Command represents a command from backend to connector
type Command struct {
	Command   string                 `json:"command"`
	RequestID string                 `json:"request_id"`
	Params    map[string]interface{} `json:"params"`
}

// Response represents a response from connector to backend
type Response struct {
	Type      string      `json:"type"`
	RequestID string      `json:"request_id"`
	Command   string      `json:"command"`
	Success   bool        `json:"success"`
	Error     string      `json:"error,omitempty"`
	Result    interface{} `json:"result,omitempty"`
	Timestamp string      `json:"timestamp"`
}
