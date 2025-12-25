package services

import (
	"Scholarweave/internal/models"
	"strings"
)

type SearchService struct {
	// Will add cache and other dependencies later
}

func NewSearchService() *SearchService {
	return &SearchService{}
}

// Update NormalizePaperData to handle different sources
func (s *SearchService) NormalizePaperData(paper *models.Paper) error {
	paper.Title = strings.TrimSpace(paper.Title)
	paper.Abstract = strings.TrimSpace(paper.Abstract)

	// Normalize source names
	switch paper.Source {
	case "openalex":
		paper.Source = "OpenAlex"
	default:
		paper.Source = "Unknown"
	}

	return nil
}

// EnrichMetadata adds additional metadata from other sources
func (s *SearchService) EnrichMetadata(paper *models.Paper) error {
	// TODO: Implement metadata enrichment
	return nil
}
