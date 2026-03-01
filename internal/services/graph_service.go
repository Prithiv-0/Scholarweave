package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"Scholarweave/internal/models"
)

type GraphService struct {
	baseURL string
	client  *http.Client
}

type openAlexWorkResponse struct {
	ID              string               `json:"id"`
	Title           string               `json:"title"`
	DOI             string               `json:"doi"`
	CitedByCount    int                  `json:"cited_by_count"`
	PublicationDate string               `json:"publication_date"`
	Type            string               `json:"type"`
	ReferencedWorks []string             `json:"referenced_works"`
	RelatedWorks    []string             `json:"related_works"`
	Concepts        []openAlexConcept    `json:"concepts"`
	Authorships     []openAlexAuthorship `json:"authorships"`
}

type openAlexWorksListResponse struct {
	Results []openAlexWorkResponse `json:"results"`
}

type openAlexConcept struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	Level       int     `json:"level"`
	Score       float64 `json:"score"`
	Wikidata    string  `json:"wikidata"`
}

type openAlexAuthorship struct {
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

func NewGraphService(baseURL string, timeout time.Duration) *GraphService {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.openalex.org"
	}
	if timeout <= 0 {
		timeout = 12 * time.Second
	}

	return &GraphService{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: timeout},
	}
}

func (s *GraphService) BuildPaperGraph(ctx context.Context, paperID string, depth int) (*models.KnowledgeGraph, error) {
	normalizedID := normalizeOpenAlexID(paperID)
	if normalizedID == "" {
		return nil, fmt.Errorf("paper ID is required")
	}
	if depth < 1 {
		depth = 1
	}
	if depth > 2 {
		depth = 2
	}

	rootWork, err := s.fetchWork(ctx, normalizedID)
	if err != nil {
		return nil, err
	}

	nodes := make(map[string]models.GraphNode)
	edges := make(map[string]models.GraphEdge)

	addWorkToGraph(nodes, edges, rootWork)

	if depth > 1 {
		candidateIDs := make([]string, 0, len(rootWork.ReferencedWorks)+len(rootWork.RelatedWorks))
		candidateIDs = append(candidateIDs, rootWork.ReferencedWorks...)
		candidateIDs = append(candidateIDs, rootWork.RelatedWorks...)

		seen := make(map[string]struct{})
		maxNeighbors := 15
		processed := 0
		for _, candidate := range candidateIDs {
			nID := normalizeOpenAlexID(candidate)
			if nID == "" {
				continue
			}
			if _, ok := seen[nID]; ok {
				continue
			}
			seen[nID] = struct{}{}
			if processed >= maxNeighbors {
				break
			}

			neighbor, fetchErr := s.fetchWork(ctx, nID)
			if fetchErr != nil {
				continue
			}
			addWorkToGraph(nodes, edges, neighbor)
			processed++
		}
	}

	s.hydratePaperNodeLabels(ctx, nodes, 12)
	graph := buildSortedGraph(nodes, edges, normalizeOpenAlexID(rootWork.ID))

	return graph, nil
}

func (s *GraphService) BuildAuthorGraph(ctx context.Context, authorID string) (*models.KnowledgeGraph, error) {
	normalizedAuthorID := normalizeOpenAlexID(authorID)
	if normalizedAuthorID == "" {
		return nil, fmt.Errorf("author ID is required")
	}

	works, err := s.fetchWorksByFilter(ctx, "author.id:"+normalizedAuthorID, 25)
	if err != nil {
		return nil, err
	}
	if len(works) == 0 {
		return nil, fmt.Errorf("no works found for author")
	}

	nodes := make(map[string]models.GraphNode)
	edges := make(map[string]models.GraphEdge)
	centerLabel := normalizedAuthorID

	for i := range works {
		work := works[i]
		addWorkToGraph(nodes, edges, &work)

		for _, authorship := range work.Authorships {
			candidateID := normalizeOpenAlexID(authorship.Author.ID)
			if candidateID == normalizedAuthorID {
				if name := strings.TrimSpace(authorship.Author.DisplayName); name != "" {
					centerLabel = name
				}
				break
			}
		}
	}

	nodes[normalizedAuthorID] = models.GraphNode{
		ID:    normalizedAuthorID,
		Label: centerLabel,
		Type:  models.NodeTypeAuthor,
	}

	s.hydratePaperNodeLabels(ctx, nodes, 12)
	graph := buildSortedGraph(nodes, edges, normalizedAuthorID)

	return graph, nil
}

func (s *GraphService) BuildConceptGraph(ctx context.Context, conceptID string) (*models.KnowledgeGraph, error) {
	normalizedConceptID := normalizeOpenAlexID(conceptID)
	if normalizedConceptID == "" {
		return nil, fmt.Errorf("concept ID is required")
	}

	works, err := s.fetchWorksByFilter(ctx, "concepts.id:"+normalizedConceptID, 25)
	if err != nil {
		return nil, err
	}
	if len(works) == 0 {
		return nil, fmt.Errorf("no works found for concept")
	}

	nodes := make(map[string]models.GraphNode)
	edges := make(map[string]models.GraphEdge)
	centerLabel := normalizedConceptID

	for i := range works {
		work := works[i]
		addWorkToGraph(nodes, edges, &work)

		for _, concept := range work.Concepts {
			candidateID := normalizeOpenAlexID(concept.ID)
			if candidateID == normalizedConceptID {
				if name := strings.TrimSpace(concept.DisplayName); name != "" {
					centerLabel = name
				}
				break
			}
		}
	}

	nodes[normalizedConceptID] = models.GraphNode{
		ID:    normalizedConceptID,
		Label: centerLabel,
		Type:  models.NodeTypeConcept,
	}

	s.hydratePaperNodeLabels(ctx, nodes, 12)
	graph := buildSortedGraph(nodes, edges, normalizedConceptID)

	return graph, nil
}

func (s *GraphService) BuildSearchGraph(ctx context.Context, paperIDs []string) (*models.KnowledgeGraph, error) {
	if len(paperIDs) == 0 {
		return nil, fmt.Errorf("paper IDs are required")
	}

	nodes := make(map[string]models.GraphNode)
	edges := make(map[string]models.GraphEdge)
	seen := make(map[string]struct{})
	maxPapers := 25
	processed := 0
	centerNodeID := ""

	for _, rawID := range paperIDs {
		normalizedID := normalizeOpenAlexID(rawID)
		if normalizedID == "" {
			continue
		}
		if _, exists := seen[normalizedID]; exists {
			continue
		}
		seen[normalizedID] = struct{}{}

		if centerNodeID == "" {
			centerNodeID = normalizedID
		}
		if processed >= maxPapers {
			break
		}

		work, err := s.fetchWork(ctx, normalizedID)
		if err != nil {
			continue
		}
		addWorkToGraph(nodes, edges, work)
		processed++
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("failed to build graph from provided paper IDs")
	}

	if centerNodeID == "" {
		for nodeID, node := range nodes {
			if node.Type == models.NodeTypePaper {
				centerNodeID = nodeID
				break
			}
		}
	}

	s.hydratePaperNodeLabels(ctx, nodes, 12)
	graph := buildSortedGraph(nodes, edges, centerNodeID)

	return graph, nil
}

func (s *GraphService) hydratePaperNodeLabels(ctx context.Context, nodes map[string]models.GraphNode, maxHydrations int) {
	if maxHydrations < 1 {
		maxHydrations = 1
	}

	// Collect candidates in deterministic order
	var candidates []string
	for nodeID, node := range nodes {
		if node.Type != models.NodeTypePaper {
			continue
		}
		if label := strings.TrimSpace(node.Label); label != "" && label != nodeID {
			continue
		}
		candidates = append(candidates, nodeID)
	}
	sort.Strings(candidates)

	hydrated := 0
	for _, nodeID := range candidates {
		if hydrated >= maxHydrations {
			break
		}
		node := nodes[nodeID]

		work, err := s.fetchWork(ctx, nodeID)
		if err != nil {
			continue
		}

		if title := strings.TrimSpace(work.Title); title != "" {
			node.Label = title
			nodes[nodeID] = node
			hydrated++
		}
	}
}

func (s *GraphService) fetchWork(ctx context.Context, id string) (*openAlexWorkResponse, error) {
	apiURL := fmt.Sprintf("%s/works/%s", s.baseURL, normalizeOpenAlexID(id))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch work: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("openalex returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var work openAlexWorkResponse
	if err := json.NewDecoder(resp.Body).Decode(&work); err != nil {
		return nil, fmt.Errorf("decode work: %w", err)
	}

	return &work, nil
}

func (s *GraphService) fetchWorksByFilter(ctx context.Context, filter string, perPage int) ([]openAlexWorkResponse, error) {
	if perPage < 1 {
		perPage = 25
	}

	query := urlpkg.Values{}
	query.Set("filter", strings.TrimSpace(filter))
	query.Set("per-page", fmt.Sprintf("%d", perPage))
	query.Set("sort", "cited_by_count:desc")

	apiURL := fmt.Sprintf("%s/works?%s", s.baseURL, query.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch works by filter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("openalex returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result openAlexWorksListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode works list: %w", err)
	}

	return result.Results, nil
}

// buildSortedGraph creates a KnowledgeGraph with deterministic node/edge ordering.
func buildSortedGraph(nodes map[string]models.GraphNode, edges map[string]models.GraphEdge, centerNodeID string) *models.KnowledgeGraph {
	sortedNodes := make([]models.GraphNode, 0, len(nodes))
	for _, node := range nodes {
		sortedNodes = append(sortedNodes, node)
	}
	sort.Slice(sortedNodes, func(i, j int) bool {
		return sortedNodes[i].ID < sortedNodes[j].ID
	})

	sortedEdges := make([]models.GraphEdge, 0, len(edges))
	for _, edge := range edges {
		sortedEdges = append(sortedEdges, edge)
	}
	sort.Slice(sortedEdges, func(i, j int) bool {
		if sortedEdges[i].Source != sortedEdges[j].Source {
			return sortedEdges[i].Source < sortedEdges[j].Source
		}
		if sortedEdges[i].Target != sortedEdges[j].Target {
			return sortedEdges[i].Target < sortedEdges[j].Target
		}
		if sortedEdges[i].Relationship != sortedEdges[j].Relationship {
			return sortedEdges[i].Relationship < sortedEdges[j].Relationship
		}
		return sortedEdges[i].Weight < sortedEdges[j].Weight
	})

	return &models.KnowledgeGraph{
		Nodes:        sortedNodes,
		Edges:        sortedEdges,
		CenterNodeID: centerNodeID,
	}
}

func addWorkToGraph(nodes map[string]models.GraphNode, edges map[string]models.GraphEdge, work *openAlexWorkResponse) {
	paperID := normalizeOpenAlexID(work.ID)
	if paperID == "" {
		return
	}

	nodes[paperID] = models.GraphNode{
		ID:    paperID,
		Label: strings.TrimSpace(work.Title),
		Type:  models.NodeTypePaper,
		Properties: map[string]any{
			"doi":              strings.TrimSpace(work.DOI),
			"cited_by_count":   work.CitedByCount,
			"publication_date": strings.TrimSpace(work.PublicationDate),
			"paper_type":       strings.TrimSpace(work.Type),
		},
	}

	for _, authorship := range work.Authorships {
		authorID := normalizeOpenAlexID(authorship.Author.ID)
		authorName := strings.TrimSpace(authorship.Author.DisplayName)
		if authorID == "" || authorName == "" {
			continue
		}

		nodes[authorID] = models.GraphNode{
			ID:    authorID,
			Label: authorName,
			Type:  models.NodeTypeAuthor,
			Properties: map[string]any{
				"orcid": strings.TrimSpace(authorship.Author.ORCID),
			},
		}
		addEdge(edges, paperID, authorID, models.EdgeTypeAuthored, 1.0)

		for _, inst := range authorship.Institutions {
			instID := normalizeOpenAlexID(inst.ID)
			instName := strings.TrimSpace(inst.DisplayName)
			if instID == "" || instName == "" {
				continue
			}

			nodes[instID] = models.GraphNode{
				ID:    instID,
				Label: instName,
				Type:  models.NodeTypeInstitution,
			}
			addEdge(edges, authorID, instID, models.EdgeTypeAffiliated, 1.0)
		}
	}

	for _, concept := range work.Concepts {
		conceptID := normalizeOpenAlexID(concept.ID)
		label := strings.TrimSpace(concept.DisplayName)
		if conceptID == "" || label == "" {
			continue
		}

		nodes[conceptID] = models.GraphNode{
			ID:    conceptID,
			Label: label,
			Type:  models.NodeTypeConcept,
			Properties: map[string]any{
				"level":    concept.Level,
				"score":    concept.Score,
				"wikidata": strings.TrimSpace(concept.Wikidata),
			},
		}
		addEdge(edges, paperID, conceptID, models.EdgeTypeHasConcept, concept.Score)
	}

	for _, referenced := range work.ReferencedWorks {
		referenceID := normalizeOpenAlexID(referenced)
		if referenceID == "" {
			continue
		}
		if _, exists := nodes[referenceID]; !exists {
			nodes[referenceID] = models.GraphNode{ID: referenceID, Label: referenceID, Type: models.NodeTypePaper}
		}
		addEdge(edges, paperID, referenceID, models.EdgeTypeCites, 1.0)
	}

	for _, related := range work.RelatedWorks {
		relatedID := normalizeOpenAlexID(related)
		if relatedID == "" {
			continue
		}
		if _, exists := nodes[relatedID]; !exists {
			nodes[relatedID] = models.GraphNode{ID: relatedID, Label: relatedID, Type: models.NodeTypePaper}
		}
		addEdge(edges, paperID, relatedID, models.EdgeTypeRelated, 1.0)
	}
}

func addEdge(edges map[string]models.GraphEdge, source string, target string, relationship models.GraphEdgeType, weight float64) {
	if source == "" || target == "" {
		return
	}
	key := source + "|" + target + "|" + string(relationship)
	edges[key] = models.GraphEdge{
		Source:       source,
		Target:       target,
		Relationship: relationship,
		Weight:       weight,
	}
}

var validOpenAlexIDRe = regexp.MustCompile(`^[A-Za-z0-9]+$`)

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
	if result != "" && !validOpenAlexIDRe.MatchString(result) {
		return ""
	}
	return result
}
