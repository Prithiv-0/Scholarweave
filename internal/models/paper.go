package models

// Paper represents a normalized academic paper across different sources
type Paper struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Abstract     string   `json:"abstract"`
	DOI          string   `json:"doi"`
	Authors      []Author `json:"authors"`
	CitedByCount int      `json:"cited_by_count"`
	Source       string   `json:"source"` // e.g., "openalex", "orkg"
}

type Author struct {
	Name  string `json:"name"`
	ID    string `json:"id,omitempty"`
	ORCID string `json:"orcid,omitempty"`
}
