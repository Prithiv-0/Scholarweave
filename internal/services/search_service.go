package services

import (
	"strings"

	"Scholarweave/internal/models"
)

type SearchService struct{}

func NewSearchService() *SearchService {
	return &SearchService{}
}

// NormalizePaperData cleans and normalizes paper fields.
func (s *SearchService) NormalizePaperData(paper *models.Paper) error {
	paper.Title = strings.TrimSpace(paper.Title)
	paper.Abstract = strings.TrimSpace(paper.Abstract)
	paper.PublicationDate = strings.TrimSpace(paper.PublicationDate)
	paper.Type = strings.TrimSpace(paper.Type)

	// Normalize source names
	switch paper.Source {
	case "openalex":
		paper.Source = "OpenAlex"
	default:
		if paper.Source == "" {
			paper.Source = "Unknown"
		}
	}

	return nil
}
