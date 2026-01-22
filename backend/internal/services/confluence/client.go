package confluence

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents a Confluence REST API client
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient creates a new Confluence client
func NewClient(baseURL, username, password string) *Client {
	return &Client{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsConfigured returns true if Confluence is configured
func (c *Client) IsConfigured() bool {
	return c.baseURL != "" && c.username != "" && c.password != ""
}

// Space represents a Confluence space
type Space struct {
	ID          int64  `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description struct {
		Plain struct {
			Value string `json:"value"`
		} `json:"plain"`
	} `json:"description,omitempty"`
	Links struct {
		WebUI string `json:"webui"`
	} `json:"_links"`
}

// Content represents a Confluence page or blog post
type Content struct {
	ID      string `json:"id"`
	Type    string `json:"type"` // page, blogpost
	Status  string `json:"status"`
	Title   string `json:"title"`
	Space   *Space `json:"space,omitempty"`
	Version struct {
		Number int `json:"number"`
	} `json:"version,omitempty"`
	Body struct {
		Storage struct {
			Value string `json:"value"`
		} `json:"storage,omitempty"`
		View struct {
			Value string `json:"value"`
		} `json:"view,omitempty"`
	} `json:"body,omitempty"`
	Links struct {
		WebUI  string `json:"webui"`
		TinyUI string `json:"tinyui"`
	} `json:"_links"`
	Children struct {
		Page struct {
			Results []Content `json:"results"`
		} `json:"page,omitempty"`
	} `json:"children,omitempty"`
	Ancestors []Content `json:"ancestors,omitempty"`
}

// SearchResult represents a search result
type SearchResult struct {
	Content               Content `json:"content"`
	Title                 string  `json:"title"`
	Excerpt               string  `json:"excerpt"`
	URL                   string  `json:"url"`
	LastModified          string  `json:"lastModified"`
	FriendlyDate          string  `json:"friendlyLastModified"`
	ResultGlobalContainer struct {
		Title      string `json:"title"`
		DisplayURL string `json:"displayUrl"`
	} `json:"resultGlobalContainer,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Results []json.RawMessage `json:"results"`
	Start   int               `json:"start"`
	Limit   int               `json:"limit"`
	Size    int               `json:"size"`
	Links   struct {
		Next string `json:"next"`
	} `json:"_links"`
}

// SearchResponse represents search API response
type SearchResponse struct {
	Results   []SearchResult `json:"results"`
	Start     int            `json:"start"`
	Limit     int            `json:"limit"`
	Size      int            `json:"size"`
	TotalSize int            `json:"totalSize"`
	CQLQuery  string         `json:"cqlQuery"`
}

// doRequest performs an HTTP request with basic auth
func (c *Client) doRequest(method, endpoint string, body io.Reader) ([]byte, error) {
	reqURL := c.baseURL + endpoint
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth
	auth := base64.StdEncoding.EncodeToString([]byte(c.username + ":" + c.password))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetSpaces returns a list of spaces
func (c *Client) GetSpaces(limit int) ([]Space, error) {
	if limit == 0 {
		limit = 25
	}

	endpoint := fmt.Sprintf("/rest/api/space?limit=%d&expand=description.plain", limit)
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Results []Space `json:"results"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Results, nil
}

// GetSpace returns a specific space by key
func (c *Client) GetSpace(spaceKey string) (*Space, error) {
	endpoint := fmt.Sprintf("/rest/api/space/%s?expand=description.plain", spaceKey)
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var space Space
	if err := json.Unmarshal(body, &space); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &space, nil
}

// GetContent returns content by ID
func (c *Client) GetContent(contentID string, expandBody bool) (*Content, error) {
	expand := "version,space,ancestors"
	if expandBody {
		expand += ",body.view,body.storage"
	}

	endpoint := fmt.Sprintf("/rest/api/content/%s?expand=%s", contentID, expand)
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var content Content
	if err := json.Unmarshal(body, &content); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &content, nil
}

// GetSpaceContent returns pages in a space
func (c *Client) GetSpaceContent(spaceKey string, contentType string, limit int) ([]Content, error) {
	if contentType == "" {
		contentType = "page"
	}
	if limit == 0 {
		limit = 25
	}

	endpoint := fmt.Sprintf("/rest/api/content?spaceKey=%s&type=%s&limit=%d&expand=version,space",
		spaceKey, contentType, limit)
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Results []Content `json:"results"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Results, nil
}

// GetChildPages returns child pages of a content
func (c *Client) GetChildPages(contentID string, limit int) ([]Content, error) {
	if limit == 0 {
		limit = 25
	}

	endpoint := fmt.Sprintf("/rest/api/content/%s/child/page?limit=%d&expand=version", contentID, limit)
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Results []Content `json:"results"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Results, nil
}

// Search performs a CQL search
func (c *Client) Search(query string, spaceKey string, limit int) (*SearchResponse, error) {
	if limit == 0 {
		limit = 20
	}

	// Build CQL query
	cql := fmt.Sprintf("text~\"%s\"", query)
	if spaceKey != "" {
		cql = fmt.Sprintf("space=%s AND %s", spaceKey, cql)
	}

	endpoint := fmt.Sprintf("/rest/api/search?cql=%s&limit=%d&expand=content.space,content.version",
		url.QueryEscape(cql), limit)
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response SearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetRecentlyViewed returns recently viewed content for the authenticated user
func (c *Client) GetRecentlyViewed(limit int) ([]Content, error) {
	if limit == 0 {
		limit = 10
	}

	// Use CQL to get recent content
	endpoint := fmt.Sprintf("/rest/api/content?orderby=history.lastUpdated desc&limit=%d&expand=version,space", limit)
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Results []Content `json:"results"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Results, nil
}

// GetWebUIURL returns the full web UI URL for content
func (c *Client) GetWebUIURL(webUIPath string) string {
	if webUIPath == "" {
		return ""
	}
	return c.baseURL + webUIPath
}
