package services

import (
	"os"
	"testing"

	"Scholarweave/internal/models"
)

func TestLibraryServiceFavoritesAndLists(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	svc := NewLibraryService(nil) // We ideally need to inject a real connection pool here
	userID := "user-1"

	favorite, err := svc.AddFavorite(userID, models.AddFavoriteRequest{
		PaperID:      "paper-123",
		Title:        "Interesting Paper",
		DOI:          "10.1000/test",
		CitedByCount: 10,
		Source:       "OpenAlex",
	})
	if err != nil {
		t.Fatalf("expected add favorite to succeed, got error: %v", err)
	}
	if favorite.PaperID != "paper-123" {
		t.Fatalf("expected paper id paper-123, got %s", favorite.PaperID)
	}

	favorites := svc.ListFavorites(userID)
	if len(favorites) != 1 {
		t.Fatalf("expected 1 favorite, got %d", len(favorites))
	}

	list, err := svc.CreateReadingList(userID, "ML Papers", "Machine learning backlog")
	if err != nil {
		t.Fatalf("expected create reading list to succeed, got error: %v", err)
	}

	updatedList, err := svc.AddItemToReadingList(userID, list.ID, models.AddReadingListItemRequest{
		PaperID:      favorite.PaperID,
		Title:        favorite.Title,
		DOI:          favorite.DOI,
		CitedByCount: favorite.CitedByCount,
		Source:       favorite.Source,
	})
	if err != nil {
		t.Fatalf("expected add item to reading list to succeed, got error: %v", err)
	}
	if len(updatedList.Papers) != 1 {
		t.Fatalf("expected 1 paper in reading list, got %d", len(updatedList.Papers))
	}

	lists := svc.ListReadingLists(userID)
	if len(lists) != 1 {
		t.Fatalf("expected 1 reading list, got %d", len(lists))
	}

	_, err = svc.RemoveItemFromReadingList(userID, list.ID, favorite.PaperID)
	if err != nil {
		t.Fatalf("expected remove item from list to succeed, got error: %v", err)
	}

	if err := svc.RemoveFavorite(userID, favorite.PaperID); err != nil {
		t.Fatalf("expected remove favorite to succeed, got error: %v", err)
	}

	if len(svc.ListFavorites(userID)) != 0 {
		t.Fatalf("expected no favorites after removal")
	}
}
