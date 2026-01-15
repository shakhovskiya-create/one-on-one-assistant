package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// SupabaseClient handles database operations
type SupabaseClient struct {
	URL    string
	Key    string
	client *http.Client
}

// NewSupabaseClient creates a new Supabase client
func NewSupabaseClient(url, key string) *SupabaseClient {
	return &SupabaseClient{
		URL:    strings.TrimSuffix(url, "/"),
		Key:    key,
		client: &http.Client{},
	}
}

// Query represents a database query
type Query struct {
	client    *SupabaseClient
	table     string
	selectStr string
	filters   []string
	orderBy   string
	limitVal  int
	single    bool
}

// From starts a query on a table
func (c *SupabaseClient) From(table string) *Query {
	return &Query{
		client:    c,
		table:     table,
		selectStr: "*",
		filters:   []string{},
	}
}

// Select specifies columns to return
func (q *Query) Select(columns string) *Query {
	q.selectStr = columns
	return q
}

// Eq adds an equality filter
func (q *Query) Eq(column, value string) *Query {
	q.filters = append(q.filters, fmt.Sprintf("%s=eq.%s", column, value))
	return q
}

// Neq adds a not-equal filter
func (q *Query) Neq(column, value string) *Query {
	q.filters = append(q.filters, fmt.Sprintf("%s=neq.%s", column, value))
	return q
}

// In adds an IN filter
func (q *Query) In(column string, values []string) *Query {
	q.filters = append(q.filters, fmt.Sprintf("%s=in.(%s)", column, strings.Join(values, ",")))
	return q
}

// Gte adds a >= filter
func (q *Query) Gte(column, value string) *Query {
	q.filters = append(q.filters, fmt.Sprintf("%s=gte.%s", column, value))
	return q
}

// Lte adds a <= filter
func (q *Query) Lte(column, value string) *Query {
	q.filters = append(q.filters, fmt.Sprintf("%s=lte.%s", column, value))
	return q
}

// Lt adds a < filter
func (q *Query) Lt(column, value string) *Query {
	q.filters = append(q.filters, fmt.Sprintf("%s=lt.%s", column, value))
	return q
}

// IsNull adds IS NULL filter
func (q *Query) IsNull(column string) *Query {
	q.filters = append(q.filters, fmt.Sprintf("%s=is.null", column))
	return q
}

// Order adds ordering
func (q *Query) Order(column string, desc bool) *Query {
	order := "asc"
	if desc {
		order = "desc"
	}
	q.orderBy = fmt.Sprintf("%s.%s", column, order)
	return q
}

// Limit limits results
func (q *Query) Limit(n int) *Query {
	q.limitVal = n
	return q
}

// Single returns single result
func (q *Query) Single() *Query {
	q.single = true
	q.limitVal = 1
	return q
}

// Execute runs a SELECT query
func (q *Query) Execute(result interface{}) error {
	reqURL := fmt.Sprintf("%s/rest/v1/%s?select=%s", q.client.URL, q.table, url.QueryEscape(q.selectStr))

	for _, filter := range q.filters {
		reqURL += "&" + filter
	}

	if q.orderBy != "" {
		reqURL += "&order=" + q.orderBy
	}

	if q.limitVal > 0 {
		reqURL += fmt.Sprintf("&limit=%d", q.limitVal)
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", q.client.Key)
	req.Header.Set("Authorization", "Bearer "+q.client.Key)

	if q.single {
		req.Header.Set("Accept", "application/vnd.pgrst.object+json")
	}

	resp, err := q.client.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("supabase error %d: %s", resp.StatusCode, string(body))
	}

	return json.Unmarshal(body, result)
}

// Insert inserts data into a table
func (c *SupabaseClient) Insert(table string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/rest/v1/%s", c.URL, table)
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.Key)
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("supabase insert error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Update updates data in a table
func (c *SupabaseClient) Update(table, column, value string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/rest/v1/%s?%s=eq.%s", c.URL, table, column, value)
	req, err := http.NewRequest("PATCH", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.Key)
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("supabase update error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Upsert upserts data into a table
func (c *SupabaseClient) Upsert(table string, data interface{}, onConflict string) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/rest/v1/%s", c.URL, table)
	if onConflict != "" {
		reqURL += "?on_conflict=" + onConflict
	}

	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.Key)
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation,resolution=merge-duplicates")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("supabase upsert error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Delete deletes data from a table
func (c *SupabaseClient) Delete(table, column, value string) error {
	reqURL := fmt.Sprintf("%s/rest/v1/%s?%s=eq.%s", c.URL, table, column, value)
	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", c.Key)
	req.Header.Set("Authorization", "Bearer "+c.Key)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase delete error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
