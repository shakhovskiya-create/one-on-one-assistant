package handlers

import (
	"strconv"

	"github.com/ekf/one-on-one-backend/internal/services/github"
	"github.com/gofiber/fiber/v2"
)

// GetGitHubStatus returns whether GitHub is configured
func (h *Handler) GetGitHubStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"configured": h.GitHub.IsConfigured(),
	})
}

// GetRepository returns repository information
func (h *Handler) GetRepository(c *fiber.Ctx) error {
	owner := c.Params("owner")
	repo := c.Params("repo")

	if owner == "" || repo == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "owner and repo are required",
		})
	}

	repository, err := h.GitHub.GetRepository(owner, repo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(repository)
}

// GetCommits returns commits from a repository
func (h *Handler) GetCommits(c *fiber.Ctx) error {
	owner := c.Params("owner")
	repo := c.Params("repo")
	branch := c.Query("branch", "")
	limit, _ := strconv.Atoi(c.Query("limit", "30"))

	if owner == "" || repo == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "owner and repo are required",
		})
	}

	commits, err := h.GitHub.GetCommits(owner, repo, branch, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(commits)
}

// GetBranches returns branches from a repository
func (h *Handler) GetBranches(c *fiber.Ctx) error {
	owner := c.Params("owner")
	repo := c.Params("repo")
	limit, _ := strconv.Atoi(c.Query("limit", "30"))

	if owner == "" || repo == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "owner and repo are required",
		})
	}

	branches, err := h.GitHub.GetBranches(owner, repo, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(branches)
}

// GetPullRequests returns pull requests from a repository
func (h *Handler) GetPullRequests(c *fiber.Ctx) error {
	owner := c.Params("owner")
	repo := c.Params("repo")
	state := c.Query("state", "all")
	limit, _ := strconv.Atoi(c.Query("limit", "30"))

	if owner == "" || repo == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "owner and repo are required",
		})
	}

	prs, err := h.GitHub.GetPullRequests(owner, repo, state, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(prs)
}

// GetTaskCommits returns commits mentioning a task ID
func (h *Handler) GetTaskCommits(c *fiber.Ctx) error {
	taskID := c.Params("id")
	owner := c.Query("owner")
	repo := c.Query("repo")
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if taskID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "task ID is required",
		})
	}

	if owner == "" || repo == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "owner and repo query params are required",
		})
	}

	commits, err := h.GitHub.SearchCommitsForTaskID(owner, repo, taskID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(commits)
}

// ParseRepoURL parses a GitHub URL and returns owner/repo
func (h *Handler) ParseRepoURL(c *fiber.Ctx) error {
	var req struct {
		URL string `json:"url"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	owner, repo, err := github.ParseRepoURL(req.URL)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"owner": owner,
		"repo":  repo,
	})
}
