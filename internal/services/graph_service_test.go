package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"time"
)

func TestBuildPaperGraphDepthOne(t *testing.T) {
	works := map[string]map[string]any{
		"W1": {
			"id":               "https://openalex.org/W1",
			"title":            "Root Paper",
			"doi":              "10.1000/root",
			"cited_by_count":   42,
			"publication_date": "2024-01-01",
			"type":             "journal-article",
			"referenced_works": []string{"https://openalex.org/W2"},
			"related_works":    []string{"https://openalex.org/W3"},
			"concepts": []map[string]any{{
				"id":           "https://openalex.org/C1",
				"display_name": "Machine Learning",
				"level":        1,
				"score":        0.92,
			}},
			"authorships": []map[string]any{{
				"author": map[string]any{
					"id":           "https://openalex.org/A1",
					"display_name": "Jane Doe",
					"orcid":        "0000-0001",
				},
				"institutions": []map[string]any{{
					"id":           "https://openalex.org/I1",
					"display_name": "Scholar University",
				}},
			}},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := path.Base(r.URL.Path)
		payload, ok := works[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer ts.Close()

	svc := NewGraphService(ts.URL, 2*time.Second)
	graph, err := svc.BuildPaperGraph(context.Background(), "W1", 1)
	if err != nil {
		t.Fatalf("expected graph build to succeed, got error: %v", err)
	}

	if graph.CenterNodeID != "W1" {
		t.Fatalf("expected center node W1, got %s", graph.CenterNodeID)
	}
	if len(graph.Nodes) < 5 {
		t.Fatalf("expected at least 5 nodes, got %d", len(graph.Nodes))
	}

	containsRelation := func(rel string) bool {
		for _, edge := range graph.Edges {
			if string(edge.Relationship) == rel {
				return true
			}
		}
		return false
	}

	if !containsRelation("authored") {
		t.Fatalf("expected authored edge")
	}
	if !containsRelation("has_concept") {
		t.Fatalf("expected has_concept edge")
	}
	if !containsRelation("cites") {
		t.Fatalf("expected cites edge")
	}
	if !containsRelation("related") {
		t.Fatalf("expected related edge")
	}
}

func TestBuildPaperGraphDepthValidation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/works/") {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":               "https://openalex.org/W1",
			"title":            "Only Paper",
			"authorships":      []any{},
			"concepts":         []any{},
			"related_works":    []any{},
			"referenced_works": []any{},
		})
	}))
	defer ts.Close()

	svc := NewGraphService(ts.URL, 2*time.Second)
	if _, err := svc.BuildPaperGraph(context.Background(), "", 1); err == nil {
		t.Fatalf("expected empty paper ID to fail")
	}

	if _, err := svc.BuildPaperGraph(context.Background(), "W1", 99); err != nil {
		t.Fatalf("expected out-of-range depth to be clamped, got error: %v", err)
	}
}

func TestBuildAuthorGraph(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/works") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{
						"id":    "https://openalex.org/W10",
						"title": "Author Graph Paper",
						"authorships": []map[string]any{{
							"author": map[string]any{
								"id":           "https://openalex.org/A10",
								"display_name": "Alice Author",
							},
						}},
						"concepts":         []any{},
						"referenced_works": []any{},
						"related_works":    []any{},
					},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	svc := NewGraphService(ts.URL, 2*time.Second)
	graph, err := svc.BuildAuthorGraph(context.Background(), "A10")
	if err != nil {
		t.Fatalf("expected author graph build to succeed, got error: %v", err)
	}

	if graph.CenterNodeID != "A10" {
		t.Fatalf("expected center node A10, got %s", graph.CenterNodeID)
	}
}

func TestBuildConceptGraph(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/works") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{
						"id":    "https://openalex.org/W20",
						"title": "Concept Graph Paper",
						"concepts": []map[string]any{{
							"id":           "https://openalex.org/C20",
							"display_name": "Graph Concept",
							"score":        0.77,
						}},
						"authorships":      []any{},
						"referenced_works": []any{},
						"related_works":    []any{},
					},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	svc := NewGraphService(ts.URL, 2*time.Second)
	graph, err := svc.BuildConceptGraph(context.Background(), "C20")
	if err != nil {
		t.Fatalf("expected concept graph build to succeed, got error: %v", err)
	}

	if graph.CenterNodeID != "C20" {
		t.Fatalf("expected center node C20, got %s", graph.CenterNodeID)
	}
}

func TestBuildSearchGraph(t *testing.T) {
	works := map[string]map[string]any{
		"W30": {
			"id":    "https://openalex.org/W30",
			"title": "Search Graph Paper 1",
			"authorships": []map[string]any{{
				"author": map[string]any{
					"id":           "https://openalex.org/A30",
					"display_name": "Author 30",
				},
			}},
			"concepts":         []any{},
			"referenced_works": []any{},
			"related_works":    []any{},
		},
		"W31": {
			"id":               "https://openalex.org/W31",
			"title":            "Search Graph Paper 2",
			"authorships":      []any{},
			"concepts":         []any{},
			"referenced_works": []any{},
			"related_works":    []any{},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := path.Base(r.URL.Path)
		payload, ok := works[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer ts.Close()

	svc := NewGraphService(ts.URL, 2*time.Second)
	graph, err := svc.BuildSearchGraph(context.Background(), []string{"W30", "W31"})
	if err != nil {
		t.Fatalf("expected search graph build to succeed, got error: %v", err)
	}

	if graph.CenterNodeID != "W30" {
		t.Fatalf("expected center node W30, got %s", graph.CenterNodeID)
	}
	if len(graph.Nodes) < 2 {
		t.Fatalf("expected at least 2 nodes, got %d", len(graph.Nodes))
	}
}
