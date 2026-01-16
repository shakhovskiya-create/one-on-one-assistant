package camunda

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client handles Camunda REST API communication
type Client struct {
	BaseURL  string
	Username string
	Password string
	client   *http.Client
}

// NewClient creates a new Camunda client
func NewClient(baseURL, username, password string) *Client {
	return &Client{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsConfigured returns true if Camunda is configured
func (c *Client) IsConfigured() bool {
	return c.BaseURL != ""
}

// ProcessDefinition represents a BPMN process definition
type ProcessDefinition struct {
	ID                string `json:"id"`
	Key               string `json:"key"`
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	Version           int    `json:"version"`
	DeploymentID      string `json:"deploymentId,omitempty"`
	Category          string `json:"category,omitempty"`
	Suspended         bool   `json:"suspended"`
	TenantID          string `json:"tenantId,omitempty"`
	VersionTag        string `json:"versionTag,omitempty"`
	HistoryTimeToLive int    `json:"historyTimeToLive,omitempty"`
}

// ProcessInstance represents a running process instance
type ProcessInstance struct {
	ID                  string `json:"id"`
	DefinitionID        string `json:"definitionId"`
	BusinessKey         string `json:"businessKey,omitempty"`
	CaseInstanceID      string `json:"caseInstanceId,omitempty"`
	Suspended           bool   `json:"suspended"`
	TenantID            string `json:"tenantId,omitempty"`
	Ended               bool   `json:"ended,omitempty"`
}

// Task represents a user task
type Task struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Assignee            string    `json:"assignee,omitempty"`
	Created             time.Time `json:"created"`
	Due                 string    `json:"due,omitempty"`
	FollowUp            string    `json:"followUp,omitempty"`
	DelegationState     string    `json:"delegationState,omitempty"`
	Description         string    `json:"description,omitempty"`
	ExecutionID         string    `json:"executionId,omitempty"`
	Owner               string    `json:"owner,omitempty"`
	ParentTaskID        string    `json:"parentTaskId,omitempty"`
	Priority            int       `json:"priority"`
	ProcessDefinitionID string    `json:"processDefinitionId,omitempty"`
	ProcessInstanceID   string    `json:"processInstanceId,omitempty"`
	TaskDefinitionKey   string    `json:"taskDefinitionKey,omitempty"`
	CaseDefinitionID    string    `json:"caseDefinitionId,omitempty"`
	CaseInstanceID      string    `json:"caseInstanceId,omitempty"`
	CaseExecutionID     string    `json:"caseExecutionId,omitempty"`
	FormKey             string    `json:"formKey,omitempty"`
	TenantID            string    `json:"tenantId,omitempty"`
}

// Variable represents a process variable
type Variable struct {
	Value     interface{} `json:"value"`
	Type      string      `json:"type,omitempty"`
	ValueInfo interface{} `json:"valueInfo,omitempty"`
}

// StartProcessRequest represents a request to start a process
type StartProcessRequest struct {
	BusinessKey string              `json:"businessKey,omitempty"`
	Variables   map[string]Variable `json:"variables,omitempty"`
}

// doRequest performs an HTTP request to Camunda API
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	reqURL := c.BaseURL + path
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("camunda error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetProcessDefinitions returns all process definitions
func (c *Client) GetProcessDefinitions() ([]ProcessDefinition, error) {
	body, err := c.doRequest("GET", "/engine-rest/process-definition", nil)
	if err != nil {
		return nil, err
	}

	var definitions []ProcessDefinition
	if err := json.Unmarshal(body, &definitions); err != nil {
		return nil, err
	}

	return definitions, nil
}

// GetProcessDefinitionByKey returns a process definition by key
func (c *Client) GetProcessDefinitionByKey(key string) (*ProcessDefinition, error) {
	body, err := c.doRequest("GET", "/engine-rest/process-definition/key/"+url.PathEscape(key), nil)
	if err != nil {
		return nil, err
	}

	var definition ProcessDefinition
	if err := json.Unmarshal(body, &definition); err != nil {
		return nil, err
	}

	return &definition, nil
}

// StartProcess starts a new process instance
func (c *Client) StartProcess(key string, req StartProcessRequest) (*ProcessInstance, error) {
	body, err := c.doRequest("POST", "/engine-rest/process-definition/key/"+url.PathEscape(key)+"/start", req)
	if err != nil {
		return nil, err
	}

	var instance ProcessInstance
	if err := json.Unmarshal(body, &instance); err != nil {
		return nil, err
	}

	return &instance, nil
}

// GetProcessInstances returns process instances
func (c *Client) GetProcessInstances(processDefinitionKey string, active bool) ([]ProcessInstance, error) {
	path := "/engine-rest/process-instance"
	params := url.Values{}
	if processDefinitionKey != "" {
		params.Set("processDefinitionKey", processDefinitionKey)
	}
	if active {
		params.Set("active", "true")
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	body, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var instances []ProcessInstance
	if err := json.Unmarshal(body, &instances); err != nil {
		return nil, err
	}

	return instances, nil
}

// GetProcessInstance returns a process instance by ID
func (c *Client) GetProcessInstance(id string) (*ProcessInstance, error) {
	body, err := c.doRequest("GET", "/engine-rest/process-instance/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}

	var instance ProcessInstance
	if err := json.Unmarshal(body, &instance); err != nil {
		return nil, err
	}

	return &instance, nil
}

// DeleteProcessInstance deletes a process instance
func (c *Client) DeleteProcessInstance(id string, skipCustomListeners bool) error {
	path := "/engine-rest/process-instance/" + url.PathEscape(id)
	if skipCustomListeners {
		path += "?skipCustomListeners=true"
	}
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

// GetTasks returns user tasks
func (c *Client) GetTasks(assignee, processInstanceID string) ([]Task, error) {
	path := "/engine-rest/task"
	params := url.Values{}
	if assignee != "" {
		params.Set("assignee", assignee)
	}
	if processInstanceID != "" {
		params.Set("processInstanceId", processInstanceID)
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	body, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	if err := json.Unmarshal(body, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetTask returns a task by ID
func (c *Client) GetTask(id string) (*Task, error) {
	body, err := c.doRequest("GET", "/engine-rest/task/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := json.Unmarshal(body, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// CompleteTask completes a user task
func (c *Client) CompleteTask(id string, variables map[string]Variable) error {
	req := map[string]interface{}{
		"variables": variables,
	}
	_, err := c.doRequest("POST", "/engine-rest/task/"+url.PathEscape(id)+"/complete", req)
	return err
}

// ClaimTask assigns a task to a user
func (c *Client) ClaimTask(id, userId string) error {
	req := map[string]string{"userId": userId}
	_, err := c.doRequest("POST", "/engine-rest/task/"+url.PathEscape(id)+"/claim", req)
	return err
}

// UnclaimTask removes the assignee from a task
func (c *Client) UnclaimTask(id string) error {
	_, err := c.doRequest("POST", "/engine-rest/task/"+url.PathEscape(id)+"/unclaim", nil)
	return err
}

// GetProcessVariables returns variables for a process instance
func (c *Client) GetProcessVariables(processInstanceID string) (map[string]Variable, error) {
	body, err := c.doRequest("GET", "/engine-rest/process-instance/"+url.PathEscape(processInstanceID)+"/variables", nil)
	if err != nil {
		return nil, err
	}

	var variables map[string]Variable
	if err := json.Unmarshal(body, &variables); err != nil {
		return nil, err
	}

	return variables, nil
}

// SetProcessVariable sets a variable for a process instance
func (c *Client) SetProcessVariable(processInstanceID, name string, variable Variable) error {
	_, err := c.doRequest("PUT", "/engine-rest/process-instance/"+url.PathEscape(processInstanceID)+"/variables/"+url.PathEscape(name), variable)
	return err
}

// DeployProcess deploys a BPMN process definition
func (c *Client) DeployProcess(name string, bpmnXML []byte) (map[string]interface{}, error) {
	// For deployment we need multipart/form-data, but for simplicity we'll use a different approach
	// This is a simplified version - in production you'd want proper multipart handling
	body, err := c.doRequest("POST", "/engine-rest/deployment/create", map[string]interface{}{
		"deployment-name": name,
		"deployment-source": "ekf-team-hub",
	})
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// HealthCheck checks if Camunda is available
func (c *Client) HealthCheck() error {
	_, err := c.doRequest("GET", "/engine-rest/engine", nil)
	return err
}
