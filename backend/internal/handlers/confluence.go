package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// ConfluenceStatus returns the Confluence configuration status
func (h *Handler) ConfluenceStatus(c *fiber.Ctx) error {
	configured := h.Confluence != nil && h.Confluence.IsConfigured()
	return c.JSON(fiber.Map{
		"configured": configured,
		"url":        h.Config.ConfluenceURL,
	})
}

// GetConfluenceSpaces returns a list of Confluence spaces
func (h *Handler) GetConfluenceSpaces(c *fiber.Ctx) error {
	if h.Confluence == nil || !h.Confluence.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Confluence not configured"})
	}

	limit := c.QueryInt("limit", 25)
	spaces, err := h.Confluence.GetSpaces(limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get spaces: " + err.Error()})
	}

	// Add full web URLs
	for i := range spaces {
		spaces[i].Links.WebUI = h.Confluence.GetWebUIURL(spaces[i].Links.WebUI)
	}

	return c.JSON(fiber.Map{
		"spaces": spaces,
	})
}

// GetConfluenceSpace returns a specific space
func (h *Handler) GetConfluenceSpace(c *fiber.Ctx) error {
	if h.Confluence == nil || !h.Confluence.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Confluence not configured"})
	}

	spaceKey := c.Params("key")
	if spaceKey == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Space key is required"})
	}

	space, err := h.Confluence.GetSpace(spaceKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get space: " + err.Error()})
	}

	space.Links.WebUI = h.Confluence.GetWebUIURL(space.Links.WebUI)

	return c.JSON(space)
}

// GetConfluenceSpaceContent returns pages in a space
func (h *Handler) GetConfluenceSpaceContent(c *fiber.Ctx) error {
	if h.Confluence == nil || !h.Confluence.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Confluence not configured"})
	}

	spaceKey := c.Params("key")
	if spaceKey == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Space key is required"})
	}

	contentType := c.Query("type", "page")
	limit := c.QueryInt("limit", 25)

	content, err := h.Confluence.GetSpaceContent(spaceKey, contentType, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get content: " + err.Error()})
	}

	// Add full web URLs
	for i := range content {
		content[i].Links.WebUI = h.Confluence.GetWebUIURL(content[i].Links.WebUI)
	}

	return c.JSON(fiber.Map{
		"pages": content,
	})
}

// GetConfluencePage returns a specific page
func (h *Handler) GetConfluencePage(c *fiber.Ctx) error {
	if h.Confluence == nil || !h.Confluence.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Confluence not configured"})
	}

	pageID := c.Params("id")
	if pageID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Page ID is required"})
	}

	expandBody := c.QueryBool("expand_body", false)

	page, err := h.Confluence.GetContent(pageID, expandBody)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get page: " + err.Error()})
	}

	page.Links.WebUI = h.Confluence.GetWebUIURL(page.Links.WebUI)

	return c.JSON(page)
}

// GetConfluenceChildPages returns child pages
func (h *Handler) GetConfluenceChildPages(c *fiber.Ctx) error {
	if h.Confluence == nil || !h.Confluence.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Confluence not configured"})
	}

	pageID := c.Params("id")
	if pageID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Page ID is required"})
	}

	limit := c.QueryInt("limit", 25)

	children, err := h.Confluence.GetChildPages(pageID, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get child pages: " + err.Error()})
	}

	// Add full web URLs
	for i := range children {
		children[i].Links.WebUI = h.Confluence.GetWebUIURL(children[i].Links.WebUI)
	}

	return c.JSON(fiber.Map{
		"pages": children,
	})
}

// SearchConfluence searches Confluence content
func (h *Handler) SearchConfluence(c *fiber.Ctx) error {
	if h.Confluence == nil || !h.Confluence.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Confluence not configured"})
	}

	query := c.Query("q")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Query is required"})
	}

	spaceKey := c.Query("space")
	limit := c.QueryInt("limit", 20)

	results, err := h.Confluence.Search(query, spaceKey, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Search failed: " + err.Error()})
	}

	// Add full web URLs
	for i := range results.Results {
		results.Results[i].Content.Links.WebUI = h.Confluence.GetWebUIURL(results.Results[i].Content.Links.WebUI)
	}

	return c.JSON(results)
}

// GetRecentConfluencePages returns recently updated pages
func (h *Handler) GetRecentConfluencePages(c *fiber.Ctx) error {
	if h.Confluence == nil || !h.Confluence.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{"error": "Confluence not configured"})
	}

	limit := c.QueryInt("limit", 10)

	pages, err := h.Confluence.GetRecentlyViewed(limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get recent pages: " + err.Error()})
	}

	// Add full web URLs
	for i := range pages {
		pages[i].Links.WebUI = h.Confluence.GetWebUIURL(pages[i].Links.WebUI)
	}

	return c.JSON(fiber.Map{
		"pages": pages,
	})
}
