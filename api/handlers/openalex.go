package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"path"
	"strings"

	"Scholarweave/internal/models"
	"Scholarweave/internal/services"

	"github.com/gofiber/fiber/v3"
)

type OpenAlexHandler struct {
	baseURL       string
	searchService *services.SearchService
}

type OpenAlexResponse struct {
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
	Results []OpenAlexWork `json:"results"`
}

type OpenAlexWork struct {
	ID                      string                 `json:"id"`
	Title                   string                 `json:"title"`
	Authors                 []Author               `json:"authorships"`
	Abstract                string                 `json:"abstract"`
	AbstractInvertedIndex   map[string][]int       `json:"abstract_inverted_index"`
	DOI                     string                 `json:"doi"`
	CitedByCount            int                    `json:"cited_by_count"`
}

type Author struct {
	Author struct {
		DisplayName string `json:"display_name"`
	} `json:"author"`
}

func NewOpenAlexHandler() *OpenAlexHandler {
	return &OpenAlexHandler{
		baseURL:       "https://api.openalex.org",
		searchService: services.NewSearchService(),
	}
}

// reconstructAbstract builds abstract text from the inverted_index if abstract is empty
func reconstructAbstract(abstract string, invertedIndex map[string][]int) string {
	if abstract != "" {
		return abstract
	}
	if len(invertedIndex) == 0 {
		return ""
	}

	// Find max index to create array of words
	maxIdx := 0
	for _, indices := range invertedIndex {
		for _, idx := range indices {
			if idx > maxIdx {
				maxIdx = idx
			}
		}
	}

	// Build words array with placeholders
	words := make([]string, maxIdx+1)
	for word, indices := range invertedIndex {
		for _, idx := range indices {
			if idx >= 0 && idx < len(words) {
				words[idx] = word
			}
		}
	}

	// Join non-empty words
	result := strings.Builder{}
	for i, word := range words {
		if word != "" {
			if i > 0 && result.Len() > 0 {
				result.WriteString(" ")
			}
			result.WriteString(word)
		}
	}
	return result.String()
}

func (h *OpenAlexHandler) GetPaperByID(c fiber.Ctx) error {
	id := c.Params("id")
	// URL params may be percent-encoded (frontend encodes full OpenAlex URLs). Unescape first.
	if un, err := urlpkg.PathUnescape(id); err == nil {
		id = un
	}
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Paper ID is required",
		})
	}

	// If client passed a full URL (e.g. https://openalex.org/W12345), extract the last path segment
	if strings.HasPrefix(id, "http://") || strings.HasPrefix(id, "https://") {
		if u, err := urlpkg.Parse(id); err == nil {
			id = path.Base(u.Path)
		}
	}

	// Format URL for single paper lookup
	apiURL := fmt.Sprintf("%s/works/%s", h.baseURL, id)
	resp, err := http.Get(apiURL)
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

	// Map OpenAlexWork to internal model
	normalizedPaper := &models.Paper{
		ID:           paper.ID,
		Title:        paper.Title,
		Abstract:     reconstructAbstract(paper.Abstract, paper.AbstractInvertedIndex),
		DOI:          paper.DOI,
		CitedByCount: paper.CitedByCount,
		Authors:      make([]models.Author, len(paper.Authors)),
	}

	// Convert authors
	for i, author := range paper.Authors {
		normalizedPaper.Authors[i] = models.Author{
			Name: author.Author.DisplayName,
		}
	}

	// Use search service to normalize data
	if err := h.searchService.NormalizePaperData(normalizedPaper); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to normalize paper data",
		})
	}

	return c.JSON(normalizedPaper)

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
	apiURL := fmt.Sprintf("%s/works?search=%s&per_page=10", h.baseURL, urlpkg.QueryEscape(query))
	resp, err := http.Get(apiURL)
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

// GetPaperRDF serves paper metadata in RDF/XML format
func (h *OpenAlexHandler) GetPaperRDF(c fiber.Ctx) error {
	id := c.Params("id")
	// Reuse existing logic to decode ID
	if un, err := urlpkg.PathUnescape(id); err == nil {
		id = un
	}
	if strings.HasPrefix(id, "http://") || strings.HasPrefix(id, "https://") {
		if u, err := urlpkg.Parse(id); err == nil {
			id = path.Base(u.Path)
		}
	}

	apiURL := fmt.Sprintf("%s/works/%s", h.baseURL, id)
	resp, err := http.Get(apiURL)
	if err != nil {
		return c.Status(500).SendString("Failed to fetch from OpenAlex")
	}
	defer resp.Body.Close()

	var paper OpenAlexWork
	if err := json.NewDecoder(resp.Body).Decode(&paper); err != nil {
		return c.Status(500).SendString("Failed to decode response")
	}

	// Generate simple RDF/XML
	rdf := fmt.Sprintf(`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:dc="http://purl.org/dc/elements/1.1/"
         xmlns:foaf="http://xmlns.com/foaf/0.1/">
  <rdf:Description rdf:about="%s">
    <dc:title>%s</dc:title>
    <dc:identifier>%s</dc:identifier>
    <dc:source>OpenAlex</dc:source>
    <dc:date>2025</dc:date>
`, paper.ID, paper.Title, paper.DOI)

	for _, author := range paper.Authors {
		rdf += fmt.Sprintf(`    <dc:creator>
      <foaf:Person>
        <foaf:name>%s</foaf:name>
      </foaf:Person>
    </dc:creator>
`, author.Author.DisplayName)
	}

	rdf += `  </rdf:Description>
</rdf:RDF>`

	c.Set("Content-Type", "application/rdf+xml")
	return c.SendString(rdf)
}
