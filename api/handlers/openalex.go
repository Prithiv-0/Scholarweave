package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v3"
)

type OpenAlexHandler struct {
	baseURL string
}

type OpenAlexResponse struct {
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
	Results []OpenAlexWork `json:"results"`
}

type OpenAlexWork struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Authors      []Author `json:"authorships"`
	Abstract     string   `json:"abstract"`
	DOI          string   `json:"doi"`
	CitedByCount int      `json:"cited_by_count"`
}

type Author struct {
	Author struct {
		DisplayName string `json:"display_name"`
	} `json:"author"`
}

func NewOpenAlexHandler() *OpenAlexHandler {
	return &OpenAlexHandler{
		baseURL: "https://api.openalex.org",
	}
}

func (h *OpenAlexHandler) GetPaperByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Paper ID is required",
		})
	}

	// Format URL for single paper lookup
	url := fmt.Sprintf("%s/works/%s", h.baseURL, id)
	resp, err := http.Get(url)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch from OpenAlex",
		})
	}
	defer resp.Body.Close()

	var paper OpenAlexWork
	if err := json.NewDecoder(resp.Body).Decode(&paper); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to decode response",
		})
	}

	return c.JSON(paper)
}

// Update SearchPapers to handle multi-word queries
func (h *OpenAlexHandler) SearchPapers(c fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}

	// Update URL to properly encode the search query
	url := fmt.Sprintf("%s/works?search=%s&per_page=10", h.baseURL, url.QueryEscape(query))
	resp, err := http.Get(url)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch from OpenAlex",
		})
	}
	defer resp.Body.Close()

	var result OpenAlexResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to decode response",
		})
	}

	return c.JSON(result)
}
