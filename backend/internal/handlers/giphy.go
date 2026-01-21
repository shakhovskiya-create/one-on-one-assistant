package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

// GiphyResponse represents the GIPHY API response
type GiphyResponse struct {
	Data []GiphyGif `json:"data"`
}

// GiphyGif represents a single GIF from GIPHY
type GiphyGif struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Images struct {
		Original struct {
			URL    string `json:"url"`
			Width  string `json:"width"`
			Height string `json:"height"`
		} `json:"original"`
		FixedHeight struct {
			URL    string `json:"url"`
			Width  string `json:"width"`
			Height string `json:"height"`
		} `json:"fixed_height"`
		FixedHeightSmall struct {
			URL    string `json:"url"`
			Width  string `json:"width"`
			Height string `json:"height"`
		} `json:"fixed_height_small"`
		Preview struct {
			URL string `json:"url"`
		} `json:"preview"`
	} `json:"images"`
}

// GifResult is the simplified response for the frontend
type GifResult struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	PreviewURL  string `json:"preview_url"`
	Width       string `json:"width"`
	Height      string `json:"height"`
}

// SearchGifs searches GIPHY for GIFs
func (h *Handler) SearchGifs(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Query parameter 'q' is required"})
	}

	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	// GIPHY API key - using public beta key for demo, should be configured
	apiKey := h.Config.GiphyAPIKey
	if apiKey == "" {
		apiKey = "dc6zaTOxFJmzC" // GIPHY public beta key
	}

	apiURL := fmt.Sprintf(
		"https://api.giphy.com/v1/gifs/search?api_key=%s&q=%s&limit=%d&offset=%d&rating=g&lang=ru",
		apiKey,
		url.QueryEscape(query),
		limit,
		offset,
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch GIFs"})
	}
	defer resp.Body.Close()

	var giphyResp GiphyResponse
	if err := json.NewDecoder(resp.Body).Decode(&giphyResp); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse GIPHY response"})
	}

	// Transform to simplified format
	results := make([]GifResult, len(giphyResp.Data))
	for i, gif := range giphyResp.Data {
		results[i] = GifResult{
			ID:         gif.ID,
			Title:      gif.Title,
			URL:        gif.Images.FixedHeight.URL,
			PreviewURL: gif.Images.FixedHeightSmall.URL,
			Width:      gif.Images.FixedHeight.Width,
			Height:     gif.Images.FixedHeight.Height,
		}
	}

	return c.JSON(fiber.Map{"gifs": results})
}

// TrendingGifs returns trending GIFs from GIPHY
func (h *Handler) TrendingGifs(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	apiKey := h.Config.GiphyAPIKey
	if apiKey == "" {
		apiKey = "dc6zaTOxFJmzC"
	}

	apiURL := fmt.Sprintf(
		"https://api.giphy.com/v1/gifs/trending?api_key=%s&limit=%d&offset=%d&rating=g",
		apiKey,
		limit,
		offset,
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch trending GIFs"})
	}
	defer resp.Body.Close()

	var giphyResp GiphyResponse
	if err := json.NewDecoder(resp.Body).Decode(&giphyResp); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse GIPHY response"})
	}

	results := make([]GifResult, len(giphyResp.Data))
	for i, gif := range giphyResp.Data {
		results[i] = GifResult{
			ID:         gif.ID,
			Title:      gif.Title,
			URL:        gif.Images.FixedHeight.URL,
			PreviewURL: gif.Images.FixedHeightSmall.URL,
			Width:      gif.Images.FixedHeight.Width,
			Height:     gif.Images.FixedHeight.Height,
		}
	}

	return c.JSON(fiber.Map{"gifs": results})
}
