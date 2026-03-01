package models

import "time"

type SavedPaper struct {
	PaperID      string    `json:"paper_id"`
	Title        string    `json:"title"`
	DOI          string    `json:"doi"`
	SavedAt      time.Time `json:"saved_at"`
	CitedByCount int       `json:"cited_by_count"`
	Source       string    `json:"source"`
}

type ReadingList struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Papers      []SavedPaper `json:"papers"`
}

type AddFavoriteRequest struct {
	PaperID      string `json:"paper_id"`
	Title        string `json:"title"`
	DOI          string `json:"doi"`
	CitedByCount int    `json:"cited_by_count"`
	Source       string `json:"source"`
}

type CreateReadingListRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AddReadingListItemRequest struct {
	PaperID      string `json:"paper_id"`
	Title        string `json:"title"`
	DOI          string `json:"doi"`
	CitedByCount int    `json:"cited_by_count"`
	Source       string `json:"source"`
}
