package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"Scholarweave/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LibraryService struct {
	pool *pgxpool.Pool
}

func NewLibraryService(pool *pgxpool.Pool) *LibraryService {
	return &LibraryService{pool: pool}
}

func (s *LibraryService) AddFavorite(userID string, input models.AddFavoriteRequest) (models.SavedPaper, error) {
	paperID := strings.TrimSpace(input.PaperID)
	if paperID == "" {
		return models.SavedPaper{}, errors.New("paper_id is required")
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "Untitled"
	}
	source := strings.TrimSpace(input.Source)
	if source == "" {
		source = "OpenAlex"
	}
	doi := strings.TrimSpace(input.DOI)
	now := time.Now().UTC()

	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO favorites (user_id, paper_id, title, doi, cited_by_count, source, saved_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (user_id, paper_id) DO UPDATE SET title=$3, doi=$4, cited_by_count=$5, source=$6, saved_at=$7`,
		userID, paperID, title, doi, input.CitedByCount, source, now,
	)
	if err != nil {
		slog.Error("failed to add favorite", "error", err)
		return models.SavedPaper{}, errors.New("failed to save favorite")
	}

	return models.SavedPaper{
		PaperID:      paperID,
		Title:        title,
		DOI:          doi,
		SavedAt:      now,
		CitedByCount: input.CitedByCount,
		Source:       source,
	}, nil
}

func (s *LibraryService) ListFavorites(userID string) []models.SavedPaper {
	rows, err := s.pool.Query(context.Background(),
		`SELECT paper_id, title, doi, cited_by_count, source, saved_at
		 FROM favorites WHERE user_id = $1 ORDER BY saved_at DESC`,
		userID,
	)
	if err != nil {
		slog.Error("failed to list favorites", "error", err)
		return []models.SavedPaper{}
	}
	defer rows.Close()

	result := []models.SavedPaper{}
	for rows.Next() {
		var fav models.SavedPaper
		if err := rows.Scan(&fav.PaperID, &fav.Title, &fav.DOI, &fav.CitedByCount, &fav.Source, &fav.SavedAt); err != nil {
			slog.Error("failed to scan favorite", "error", err)
			continue
		}
		result = append(result, fav)
	}
	if err := rows.Err(); err != nil {
		slog.Error("row iteration error in ListFavorites", "error", err)
	}
	return result
}

func (s *LibraryService) RemoveFavorite(userID, paperID string) error {
	paperID = strings.TrimSpace(paperID)
	tag, err := s.pool.Exec(context.Background(),
		`DELETE FROM favorites WHERE user_id = $1 AND paper_id = $2`,
		userID, paperID,
	)
	if err != nil {
		slog.Error("failed to remove favorite", "error", err)
		return errors.New("failed to remove favorite")
	}
	if tag.RowsAffected() == 0 {
		return errors.New("favorite not found")
	}
	return nil
}

func (s *LibraryService) CreateReadingList(userID string, name, description string) (models.ReadingList, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" {
		return models.ReadingList{}, errors.New("name is required")
	}

	id := uuid.NewString()
	now := time.Now().UTC()

	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO reading_lists (id, user_id, name, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		id, userID, name, description, now, now,
	)
	if err != nil {
		slog.Error("failed to create reading list", "error", err)
		return models.ReadingList{}, errors.New("failed to create reading list")
	}

	return models.ReadingList{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
		Papers:      []models.SavedPaper{},
	}, nil
}

func (s *LibraryService) ListReadingLists(userID string) []models.ReadingList {
	rows, err := s.pool.Query(context.Background(),
		`SELECT id, name, description, created_at, updated_at FROM reading_lists WHERE user_id = $1 ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		slog.Error("failed to list reading lists", "error", err)
		return []models.ReadingList{}
	}
	defer rows.Close()

	var lists []models.ReadingList
	for rows.Next() {
		var rl models.ReadingList
		if err := rows.Scan(&rl.ID, &rl.Name, &rl.Description, &rl.CreatedAt, &rl.UpdatedAt); err != nil {
			slog.Error("failed to scan reading list", "error", err)
			continue
		}
		rl.Papers = s.getReadingListItems(rl.ID)
		lists = append(lists, rl)
	}
	if err := rows.Err(); err != nil {
		slog.Error("row iteration error in ListReadingLists", "error", err)
	}
	if lists == nil {
		lists = []models.ReadingList{}
	}
	return lists
}

func (s *LibraryService) getReadingListItems(listID string) []models.SavedPaper {
	rows, err := s.pool.Query(context.Background(),
		`SELECT paper_id, title, doi, cited_by_count, source, saved_at
		 FROM reading_list_items WHERE list_id = $1 ORDER BY saved_at DESC`,
		listID,
	)
	if err != nil {
		slog.Error("failed to list reading list items", "error", err)
		return []models.SavedPaper{}
	}
	defer rows.Close()

	result := []models.SavedPaper{}
	for rows.Next() {
		var item models.SavedPaper
		if err := rows.Scan(&item.PaperID, &item.Title, &item.DOI, &item.CitedByCount, &item.Source, &item.SavedAt); err != nil {
			slog.Error("failed to scan reading list item", "error", err)
			continue
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		slog.Error("row iteration error in getReadingListItems", "error", err)
	}
	return result
}

func (s *LibraryService) AddItemToReadingList(userID, listID string, input models.AddReadingListItemRequest) (models.ReadingList, error) {
	listID = strings.TrimSpace(listID)
	paperID := strings.TrimSpace(input.PaperID)
	if paperID == "" {
		return models.ReadingList{}, errors.New("paper_id is required")
	}

	// Verify ownership
	var ownerID string
	err := s.pool.QueryRow(context.Background(),
		`SELECT user_id FROM reading_lists WHERE id = $1`, listID,
	).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ReadingList{}, errors.New("reading list not found")
		}
		slog.Error("failed to verify list ownership", "error", err)
		return models.ReadingList{}, errors.New("reading list not found")
	}
	if ownerID != userID {
		return models.ReadingList{}, errors.New("reading list not found")
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "Untitled"
	}
	source := strings.TrimSpace(input.Source)
	if source == "" {
		source = "OpenAlex"
	}
	doi := strings.TrimSpace(input.DOI)
	now := time.Now().UTC()

	_, err = s.pool.Exec(context.Background(),
		`INSERT INTO reading_list_items (list_id, paper_id, title, doi, cited_by_count, source, saved_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (list_id, paper_id) DO UPDATE SET title=$3, doi=$4, cited_by_count=$5, source=$6, saved_at=$7`,
		listID, paperID, title, doi, input.CitedByCount, source, now,
	)
	if err != nil {
		slog.Error("failed to add item to reading list", "error", err)
		return models.ReadingList{}, errors.New("failed to add paper to reading list")
	}

	// Update the reading list's updated_at
	if _, updateErr := s.pool.Exec(context.Background(), `UPDATE reading_lists SET updated_at = $1 WHERE id = $2`, now, listID); updateErr != nil {
		slog.Error("failed to update reading list timestamp", "error", updateErr)
	}

	return s.getReadingListByID(listID)
}

func (s *LibraryService) RemoveItemFromReadingList(userID, listID, paperID string) (models.ReadingList, error) {
	listID = strings.TrimSpace(listID)
	paperID = strings.TrimSpace(paperID)

	// Verify ownership
	var ownerID string
	err := s.pool.QueryRow(context.Background(),
		`SELECT user_id FROM reading_lists WHERE id = $1`, listID,
	).Scan(&ownerID)
	if err != nil {
		return models.ReadingList{}, errors.New("reading list not found")
	}
	if ownerID != userID {
		return models.ReadingList{}, errors.New("reading list not found")
	}

	tag, err := s.pool.Exec(context.Background(),
		`DELETE FROM reading_list_items WHERE list_id = $1 AND paper_id = $2`,
		listID, paperID,
	)
	if err != nil {
		slog.Error("failed to remove item from reading list", "error", err)
		return models.ReadingList{}, errors.New("failed to remove paper from reading list")
	}
	if tag.RowsAffected() == 0 {
		return models.ReadingList{}, errors.New("paper not found in reading list")
	}

	now := time.Now().UTC()
	if _, updateErr := s.pool.Exec(context.Background(), `UPDATE reading_lists SET updated_at = $1 WHERE id = $2`, now, listID); updateErr != nil {
		slog.Error("failed to update reading list timestamp", "error", updateErr)
	}

	return s.getReadingListByID(listID)
}

func (s *LibraryService) DeleteReadingList(userID, listID string) error {
	listID = strings.TrimSpace(listID)

	// Verify ownership
	var ownerID string
	err := s.pool.QueryRow(context.Background(),
		`SELECT user_id FROM reading_lists WHERE id = $1`, listID,
	).Scan(&ownerID)
	if err != nil {
		return errors.New("reading list not found")
	}
	if ownerID != userID {
		return errors.New("reading list not found")
	}

	_, err = s.pool.Exec(context.Background(), `DELETE FROM reading_lists WHERE id = $1`, listID)
	if err != nil {
		slog.Error("failed to delete reading list", "error", err)
		return errors.New("failed to delete reading list")
	}
	return nil
}

func (s *LibraryService) getReadingListByID(listID string) (models.ReadingList, error) {
	var rl models.ReadingList
	err := s.pool.QueryRow(context.Background(),
		`SELECT id, name, description, created_at, updated_at FROM reading_lists WHERE id = $1`, listID,
	).Scan(&rl.ID, &rl.Name, &rl.Description, &rl.CreatedAt, &rl.UpdatedAt)
	if err != nil {
		return models.ReadingList{}, errors.New("reading list not found")
	}
	rl.Papers = s.getReadingListItems(listID)
	return rl, nil
}
