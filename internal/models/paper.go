package models

// Paper represents a normalized academic paper across different sources
type Paper struct {
	ID              string        `json:"id"`
	Title           string        `json:"title"`
	Abstract        string        `json:"abstract"`
	DOI             string        `json:"doi"`
	Authors         []Author      `json:"authors"`
	Concepts        []Concept     `json:"concepts,omitempty"`
	Institutions    []Institution `json:"institutions,omitempty"`
	ReferencedWorks []string      `json:"referenced_works,omitempty"`
	RelatedWorks    []string      `json:"related_works,omitempty"`
	CitedByCount    int           `json:"cited_by_count"`
	Source          string        `json:"source"`           // e.g., "openalex", "orkg"
	PublicationDate string        `json:"publication_date"` // e.g., "2024-01-15"
	Type            string        `json:"type"`             // e.g., "journal-article"
}

type Author struct {
	Name  string `json:"name"`
	ID    string `json:"id,omitempty"`
	ORCID string `json:"orcid,omitempty"`
}

type Concept struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	Level       int     `json:"level"`
	Score       float64 `json:"score"`
	Wikidata    string  `json:"wikidata,omitempty"`
}

type Institution struct {
	ID          string `json:"id,omitempty"`
	DisplayName string `json:"display_name"`
}
