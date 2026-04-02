package tests

import (
	"context"
	"testing"

	navigatorapplication "github.com/momlesstomato/pixel-server/pkg/navigator/application"
	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
)

// TestServiceSavedSearchCRUD verifies saved search create, list, and delete behavior.
func TestServiceSavedSearchCRUD(t *testing.T) {
	stub := repositoryStub{search: domain.SavedSearch{ID: 1, UserID: 1, SearchCode: "hotel_view"}}
	service, _ := navigatorapplication.NewService(stub)
	if _, err := service.ListSavedSearches(context.Background(), 0); err == nil {
		t.Fatalf("expected list failure for invalid user id")
	}
	searches, err := service.ListSavedSearches(context.Background(), 1)
	if err != nil || len(searches) != 1 {
		t.Fatalf("unexpected list result %+v err=%v", searches, err)
	}
	if _, err := service.CreateSavedSearch(context.Background(), domain.SavedSearch{}); err == nil {
		t.Fatalf("expected create failure for invalid user id")
	}
	if _, err := service.CreateSavedSearch(context.Background(), domain.SavedSearch{UserID: 1}); err == nil {
		t.Fatalf("expected create failure for empty search code")
	}
	created, err := service.CreateSavedSearch(context.Background(), domain.SavedSearch{UserID: 1, SearchCode: "test"})
	if err != nil || created.ID != 1 {
		t.Fatalf("unexpected create result %+v err=%v", created, err)
	}
	if err := service.DeleteSavedSearch(context.Background(), 0); err == nil {
		t.Fatalf("expected delete failure for invalid id")
	}
}

// TestServiceFavouriteCRUD verifies favourite add, list, and remove behavior.
func TestServiceFavouriteCRUD(t *testing.T) {
	stub := repositoryStub{}
	service, _ := navigatorapplication.NewService(stub)
	if _, err := service.ListFavourites(context.Background(), 0); err == nil {
		t.Fatalf("expected list failure for invalid user id")
	}
	favs, err := service.ListFavourites(context.Background(), 1)
	if err != nil || len(favs) != 1 {
		t.Fatalf("unexpected list result %+v err=%v", favs, err)
	}
	if err := service.AddFavourite(context.Background(), 0, 1); err == nil {
		t.Fatalf("expected add failure for invalid user id")
	}
	if err := service.AddFavourite(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected add failure for invalid room id")
	}
	if err := service.AddFavourite(context.Background(), 1, 1); err != nil {
		t.Fatalf("unexpected add error %v", err)
	}
	if err := service.RemoveFavourite(context.Background(), 0, 1); err == nil {
		t.Fatalf("expected remove failure for invalid user id")
	}
	if err := service.RemoveFavourite(context.Background(), 1, 1); err != nil {
		t.Fatalf("unexpected remove error %v", err)
	}
}

// TestServiceFavouriteLimitReached verifies max favourite enforcement.
func TestServiceFavouriteLimitReached(t *testing.T) {
	stub := repositoryStub{favourites: navigatorapplication.MaxFavourites}
	service, _ := navigatorapplication.NewService(stub)
	if err := service.AddFavourite(context.Background(), 1, 99); err != domain.ErrFavouriteLimitReached {
		t.Fatalf("expected favourite limit error, got %v", err)
	}
}
