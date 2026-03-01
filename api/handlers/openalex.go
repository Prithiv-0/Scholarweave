package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	urlpkg "net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"Scholarweave/internal/models"
	"Scholarweave/internal/services"

	"github.com/gofiber/fiber/v3"
)

type OpenAlexHandler struct {
	baseURL       string
	client        *http.Client
	searchService *services.SearchService
}

type OpenAlexResponse struct {
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
	Results []OpenAlexWork `json:"results"`
}

type OpenAlexWork struct {
	ID                    string            `json:"id"`
	Title                 string            `json:"title"`
	Authors               []Author          `json:"authorships"`
	Concepts              []OpenAlexConcept `json:"concepts"`
	Abstract              string            `json:"abstract"`
	AbstractInvertedIndex map[string][]int  `json:"abstract_inverted_index"`
	DOI                   string            `json:"doi"`
	CitedByCount          int               `json:"cited_by_count"`
	PublicationDate       string            `json:"publication_date"`
	Type                  string            `json:"type"`
	ReferencedWorks       []string          `json:"referenced_works"`
	RelatedWorks          []string          `json:"related_works"`
}

type Author struct {
	Author struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
		ORCID       string `json:"orcid"`
	} `json:"author"`
	Institutions []struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"institutions"`
}

type OpenAlexConcept struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	Level       int     `json:"level"`
	Score       float64 `json:"score"`
	Wikidata    string  `json:"wikidata"`
}

type NormalizedSearchResponse struct {
	Meta    SearchMeta     `json:"meta"`
	Results []models.Paper `json:"results"`
}

type SearchMeta struct {
	Count      int    `json:"count"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	TotalPages int    `json:"total_pages"`
	Sort       string `json:"sort"`
}

func NewOpenAlexHandler(baseURL string, timeout time.Duration) *OpenAlexHandler {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.openalex.org"
	}
	if timeout <= 0 {
		timeout = 12 * time.Second
	}

	return &OpenAlexHandler{
		baseURL:       strings.TrimRight(baseURL, "/"),
		client:        &http.Client{Timeout: timeout},
		searchService: services.NewSearchService(),
	}
}

// PingUpstream checks if OpenAlex is reachable.
func (h *OpenAlexHandler) PingUpstream() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.baseURL, nil)
	if err != nil {
		return false
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func parseIntQuery(c fiber.Ctx, key string, defaultValue int, minValue int, maxValue int) (int, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", key)
	}
	if value < minValue || value > maxValue {
		return 0, fmt.Errorf("%s must be between %d and %d", key, minValue, maxValue)
	}

	return value, nil
}

func mapSort(sort string) (string, string) {
	normalizedSort := strings.ToLower(strings.TrimSpace(sort))
	switch normalizedSort {
	case "", "relevance":
		return "", "relevance"
	case "citations":
		return "cited_by_count:desc", "citations"
	case "date":
		return "publication_date:desc", "date"
	default:
		return "", "relevance"
	}
}

func (h *OpenAlexHandler) fetchJSON(ctx context.Context, apiURL string, target any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return resp.StatusCode, fmt.Errorf("openalex returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return resp.StatusCode, err
	}

	return resp.StatusCode, nil
}

var validOpenAlexID = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func normalizeOpenAlexID(id string) string {
	if unescaped, err := urlpkg.PathUnescape(id); err == nil {
		id = unescaped
	}

	if strings.HasPrefix(id, "http://") || strings.HasPrefix(id, "https://") {
		if parsed, err := urlpkg.Parse(id); err == nil {
			id = path.Base(parsed.Path)
		}
	}

	result := strings.TrimSpace(id)
	if result != "" && !validOpenAlexID.MatchString(result) {
		return ""
	}
	return result
}

func xmlEscape(value string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(value))
	return buf.String()
}

func (h *OpenAlexHandler) toNormalizedPaper(work OpenAlexWork) (*models.Paper, error) {
	normalizedPaper := &models.Paper{
		ID:              work.ID,
		Title:           work.Title,
		Abstract:        reconstructAbstract(work.Abstract, work.AbstractInvertedIndex),
		DOI:             work.DOI,
		CitedByCount:    work.CitedByCount,
		Source:          "openalex",
		PublicationDate: work.PublicationDate,
		Type:            work.Type,
		Authors:         make([]models.Author, len(work.Authors)),
		Concepts:        make([]models.Concept, 0, len(work.Concepts)),
		ReferencedWorks: make([]string, 0, len(work.ReferencedWorks)),
		RelatedWorks:    make([]string, 0, len(work.RelatedWorks)),
	}

	institutions := make(map[string]models.Institution)

	for i, author := range work.Authors {
		normalizedPaper.Authors[i] = models.Author{
			Name:  author.Author.DisplayName,
			ID:    author.Author.ID,
			ORCID: author.Author.ORCID,
		}

		for _, institution := range author.Institutions {
			if strings.TrimSpace(institution.DisplayName) == "" {
				continue
			}

			instID := normalizeOpenAlexID(institution.ID)
			key := instID
			if key == "" {
				key = institution.DisplayName
			}

			institutions[key] = models.Institution{
				ID:          instID,
				DisplayName: strings.TrimSpace(institution.DisplayName),
			}
		}
	}

	for _, concept := range work.Concepts {
		normalizedPaper.Concepts = append(normalizedPaper.Concepts, models.Concept{
			ID:          normalizeOpenAlexID(concept.ID),
			DisplayName: strings.TrimSpace(concept.DisplayName),
			Level:       concept.Level,
			Score:       concept.Score,
			Wikidata:    strings.TrimSpace(concept.Wikidata),
		})
	}

	for _, ref := range work.ReferencedWorks {
		if id := normalizeOpenAlexID(ref); id != "" {
			normalizedPaper.ReferencedWorks = append(normalizedPaper.ReferencedWorks, id)
		}
	}

	for _, related := range work.RelatedWorks {
		if id := normalizeOpenAlexID(related); id != "" {
			normalizedPaper.RelatedWorks = append(normalizedPaper.RelatedWorks, id)
		}
	}

	for _, institution := range institutions {
		normalizedPaper.Institutions = append(normalizedPaper.Institutions, institution)
	}

	if err := h.searchService.NormalizePaperData(normalizedPaper); err != nil {
		return nil, err
	}

	return normalizedPaper, nil
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
	id := normalizeOpenAlexID(c.Params("id"))
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Paper ID is required")
	}

	apiURL := fmt.Sprintf("%s/works/%s", h.baseURL, id)
	var paper OpenAlexWork
	statusCode, err := h.fetchJSON(c, apiURL, &paper)
	if err != nil {
		if statusCode == http.StatusNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Paper not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch paper from OpenAlex")
	}

	normalizedPaper, err := h.toNormalizedPaper(paper)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to normalize paper data")
	}

	return c.JSON(normalizedPaper)
}

// SearchPapers handles multi-word queries with filters, pagination, sorting.
func (h *OpenAlexHandler) SearchPapers(c fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Query parameter 'q' is required")
	}

	page, pageErr := parseIntQuery(c, "page", 1, 1, 1000)
	if pageErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, pageErr.Error())
	}

	perPage, perPageErr := parseIntQuery(c, "per_page", 10, 1, 50)
	if perPageErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, perPageErr.Error())
	}

	fromYear, fromYearErr := parseIntQuery(c, "from_year", 0, 0, 3000)
	if fromYearErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, fromYearErr.Error())
	}

	toYear, toYearErr := parseIntQuery(c, "to_year", 0, 0, 3000)
	if toYearErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, toYearErr.Error())
	}

	if fromYear > 0 && toYear > 0 && fromYear > toYear {
		return fiber.NewError(fiber.StatusBadRequest, "from_year must be less than or equal to to_year")
	}

	publicationType := strings.TrimSpace(c.Query("type"))
	subject := strings.TrimSpace(c.Query("subject"))

	openAlexSort, appliedSort := mapSort(c.Query("sort"))

	searchQuery := strings.TrimSpace(query)
	if subject != "" {
		searchQuery = strings.TrimSpace(searchQuery + " " + subject)
	}

	params := urlpkg.Values{}
	params.Set("search", searchQuery)
	params.Set("page", strconv.Itoa(page))
	params.Set("per-page", strconv.Itoa(perPage))

	if openAlexSort != "" {
		params.Set("sort", openAlexSort)
	}

	filters := make([]string, 0, 3)
	if fromYear > 0 && toYear > 0 {
		filters = append(filters, fmt.Sprintf("publication_year:%d-%d", fromYear, toYear))
	} else if fromYear > 0 {
		filters = append(filters, fmt.Sprintf("from_publication_date:%d-01-01", fromYear))
	} else if toYear > 0 {
		filters = append(filters, fmt.Sprintf("to_publication_date:%d-12-31", toYear))
	}

	if publicationType != "" {
		filters = append(filters, fmt.Sprintf("type:%s", publicationType))
	}

	if len(filters) > 0 {
		params.Set("filter", strings.Join(filters, ","))
	}

	apiURL := fmt.Sprintf("%s/works?%s", h.baseURL, params.Encode())
	var result OpenAlexResponse
	_, err := h.fetchJSON(c, apiURL, &result)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch from OpenAlex")
	}

	response := NormalizedSearchResponse{}
	totalPages := int(math.Ceil(float64(result.Meta.Count) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
	}
	response.Meta = SearchMeta{
		Count:      result.Meta.Count,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		Sort:       appliedSort,
	}
	response.Results = make([]models.Paper, 0, len(result.Results))

	for _, work := range result.Results {
		normalizedPaper, normalizeErr := h.toNormalizedPaper(work)
		if normalizeErr != nil {
			// Skip bad papers instead of failing the whole search
			slog.Warn("skipping paper due to normalization error", "paper_id", work.ID, "error", normalizeErr)
			continue
		}
		response.Results = append(response.Results, *normalizedPaper)
	}

	return c.JSON(response)
}

// GetPaperRDF serves paper metadata in RDF/XML format
func (h *OpenAlexHandler) GetPaperRDF(c fiber.Ctx) error {
	id := normalizeOpenAlexID(c.Params("id"))
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Paper ID is required")
	}

	apiURL := fmt.Sprintf("%s/works/%s", h.baseURL, id)
	var paper OpenAlexWork
	statusCode, err := h.fetchJSON(c, apiURL, &paper)
	if err != nil {
		if statusCode == http.StatusNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Paper not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch from OpenAlex")
	}

	// Use actual publication date, fall back to current year
	pubDate := strings.TrimSpace(paper.PublicationDate)
	if pubDate == "" {
		pubDate = time.Now().Format("2006")
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
    <dc:date>%s</dc:date>
`, xmlEscape(paper.ID), xmlEscape(paper.Title), xmlEscape(paper.DOI), xmlEscape(pubDate))

	for _, author := range paper.Authors {
		rdf += fmt.Sprintf(`    <dc:creator>
      <foaf:Person>
        <foaf:name>%s</foaf:name>
      </foaf:Person>
    </dc:creator>
`, xmlEscape(author.Author.DisplayName))
	}

	rdf += `  </rdf:Description>
</rdf:RDF>`

	c.Set("Content-Type", "application/rdf+xml")
	return c.SendString(rdf)
}
