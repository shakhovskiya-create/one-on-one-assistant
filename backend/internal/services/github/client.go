package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Client represents a GitHub API client
type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient creates a new GitHub client
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsConfigured returns true if GitHub token is set
func (c *Client) IsConfigured() bool {
	return c.token != ""
}

// Repository represents a GitHub repository
type Repository struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	Fork          bool   `json:"fork"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	PushedAt      string `json:"pushed_at"`
	Language      string `json:"language"`
	StarCount     int    `json:"stargazers_count"`
	ForksCount    int    `json:"forks_count"`
}

// Commit represents a GitHub commit
type Commit struct {
	SHA     string `json:"sha"`
	HTMLURL string `json:"html_url"`
	Commit  struct {
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Date  string `json:"date"`
		} `json:"author"`
		Committer struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Date  string `json:"date"`
		} `json:"committer"`
	} `json:"commit"`
	Author *struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
		HTMLURL   string `json:"html_url"`
	} `json:"author"`
	Committer *struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
		HTMLURL   string `json:"html_url"`
	} `json:"committer"`
	Stats *struct {
		Additions int `json:"additions"`
		Deletions int `json:"deletions"`
		Total     int `json:"total"`
	} `json:"stats"`
}

// Branch represents a GitHub branch
type Branch struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	Protected bool `json:"protected"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	ID        int64  `json:"id"`
	Number    int    `json:"number"`
	State     string `json:"state"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	ClosedAt  string `json:"closed_at"`
	MergedAt  string `json:"merged_at"`
	User      *struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
		HTMLURL   string `json:"html_url"`
	} `json:"user"`
	Head struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"base"`
	Mergeable *bool `json:"mergeable"`
}

// doRequest performs an HTTP request with auth token
func (c *Client) doRequest(method, url string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// ParseRepoURL parses a GitHub URL and extracts owner and repo
func ParseRepoURL(repoURL string) (owner, repo string, err error) {
	// Support formats:
	// https://github.com/owner/repo
	// https://github.com/owner/repo.git
	// git@github.com:owner/repo.git
	// owner/repo

	repoURL = strings.TrimSuffix(repoURL, ".git")

	// SSH format
	if strings.HasPrefix(repoURL, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(repoURL, "git@github.com:"), "/")
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
	}

	// HTTPS format
	if strings.Contains(repoURL, "github.com") {
		u, err := url.Parse(repoURL)
		if err == nil {
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(parts) >= 2 {
				return parts[0], parts[1], nil
			}
		}
	}

	// Simple owner/repo format
	parts := strings.Split(repoURL, "/")
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("invalid repository URL format: %s", repoURL)
}

// GetRepository returns repository information
func (c *Client) GetRepository(owner, repo string) (*Repository, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	body, err := c.doRequest("GET", url)
	if err != nil {
		return nil, err
	}

	var repository Repository
	if err := json.Unmarshal(body, &repository); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &repository, nil
}

// GetCommits returns commits from a repository
func (c *Client) GetCommits(owner, repo, branch string, limit int) ([]Commit, error) {
	if limit == 0 {
		limit = 30
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?per_page=%d", owner, repo, limit)
	if branch != "" {
		url += "&sha=" + branch
	}

	body, err := c.doRequest("GET", url)
	if err != nil {
		return nil, err
	}

	var commits []Commit
	if err := json.Unmarshal(body, &commits); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return commits, nil
}

// GetBranches returns branches from a repository
func (c *Client) GetBranches(owner, repo string, limit int) ([]Branch, error) {
	if limit == 0 {
		limit = 30
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches?per_page=%d", owner, repo, limit)
	body, err := c.doRequest("GET", url)
	if err != nil {
		return nil, err
	}

	var branches []Branch
	if err := json.Unmarshal(body, &branches); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return branches, nil
}

// GetPullRequests returns pull requests from a repository
func (c *Client) GetPullRequests(owner, repo, state string, limit int) ([]PullRequest, error) {
	if limit == 0 {
		limit = 30
	}
	if state == "" {
		state = "all"
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=%s&per_page=%d", owner, repo, state, limit)
	body, err := c.doRequest("GET", url)
	if err != nil {
		return nil, err
	}

	var prs []PullRequest
	if err := json.Unmarshal(body, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return prs, nil
}

// SearchCommitsForTaskID finds commits mentioning a task ID
func (c *Client) SearchCommitsForTaskID(owner, repo, taskID string, limit int) ([]Commit, error) {
	commits, err := c.GetCommits(owner, repo, "", limit*2) // Get more commits to filter
	if err != nil {
		return nil, err
	}

	// Filter commits that mention the task ID
	pattern := regexp.MustCompile(`(?i)(task|#|EKF-)` + regexp.QuoteMeta(taskID))
	var matching []Commit
	for _, commit := range commits {
		if pattern.MatchString(commit.Commit.Message) {
			matching = append(matching, commit)
			if len(matching) >= limit {
				break
			}
		}
	}

	return matching, nil
}

// SearchPullRequestsForTaskID finds PRs mentioning a task ID
func (c *Client) SearchPullRequestsForTaskID(owner, repo, taskID string, limit int) ([]PullRequest, error) {
	prs, err := c.GetPullRequests(owner, repo, "all", limit*2) // Get more to filter
	if err != nil {
		return nil, err
	}

	// Filter PRs that mention the task ID in title or body
	pattern := regexp.MustCompile(`(?i)(task|#|EKF-)` + regexp.QuoteMeta(taskID))
	var matching []PullRequest
	for _, pr := range prs {
		if pattern.MatchString(pr.Title) || pattern.MatchString(pr.Body) {
			matching = append(matching, pr)
			if len(matching) >= limit {
				break
			}
		}
	}

	return matching, nil
}
