package navigator

import (
	"context"
	"fmt"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	navigatorapplication "github.com/momlesstomato/pixel-server/pkg/navigator/application"
	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
	navigatormodel "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/model"
	navigatorstore "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupE2E opens an in-memory SQLite database and returns a navigator service.
func setupE2E(t *testing.T) *navigatorapplication.Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err = db.AutoMigrate(&navigatormodel.Category{}, &navigatormodel.Room{}, &navigatormodel.SavedSearch{}, &navigatormodel.Favourite{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo, err := navigatorstore.NewRepository(db)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	svc, err := navigatorapplication.NewService(repo)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	var noop func(sdk.Event)
	svc.SetEventFirer(noop)
	return svc
}

// Test11NavigatorCategoryFlow verifies end-to-end category create, list, and delete.
func Test11NavigatorCategoryFlow(t *testing.T) {
	svc := setupE2E(t)
	ctx := context.Background()
	created, err := svc.CreateCategory(ctx, domain.Category{Caption: "Public Rooms", CategoryType: "public"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected non-zero category id")
	}
	cats, err := svc.ListCategories(ctx)
	if err != nil || len(cats) == 0 {
		t.Fatalf("list categories err=%v len=%d", err, len(cats))
	}
	found, err := svc.FindCategoryByID(ctx, created.ID)
	if err != nil || found.Caption != "Public Rooms" {
		t.Fatalf("find category err=%v caption=%s", err, found.Caption)
	}
	if err := svc.DeleteCategory(ctx, created.ID); err != nil {
		t.Fatalf("delete category: %v", err)
	}
	cats, _ = svc.ListCategories(ctx)
	if len(cats) != 0 {
		t.Fatalf("expected empty categories after delete, got %d", len(cats))
	}
}

// Test11NavigatorRoomFlow verifies end-to-end room create, list, find, and delete.
func Test11NavigatorRoomFlow(t *testing.T) {
	svc := setupE2E(t)
	ctx := context.Background()
	room, err := svc.CreateRoom(ctx, domain.Room{Name: "My Room", OwnerID: 1, OwnerName: "alice", State: "open"})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if room.ID == 0 {
		t.Fatalf("expected non-zero room id")
	}
	rooms, total, err := svc.ListRooms(ctx, domain.RoomFilter{})
	if err != nil || total != 1 {
		t.Fatalf("list rooms err=%v total=%d", err, total)
	}
	if rooms[0].Name != "My Room" {
		t.Fatalf("unexpected room name: %s", rooms[0].Name)
	}
	found, err := svc.FindRoomByID(ctx, room.ID)
	if err != nil || found.OwnerName != "alice" {
		t.Fatalf("find room err=%v owner=%s", err, found.OwnerName)
	}
	newName := "Updated Room"
	updated, err := svc.UpdateRoom(ctx, room.ID, domain.RoomPatch{Name: &newName})
	if err != nil || updated.Name != "Updated Room" {
		t.Fatalf("update room err=%v name=%s", err, updated.Name)
	}
	if err := svc.DeleteRoom(ctx, room.ID); err != nil {
		t.Fatalf("delete room: %v", err)
	}
}

// Test11NavigatorFavouriteFlow verifies end-to-end favourite add, list, and remove.
func Test11NavigatorFavouriteFlow(t *testing.T) {
	svc := setupE2E(t)
	ctx := context.Background()
	room, _ := svc.CreateRoom(ctx, domain.Room{Name: "Fav Room", OwnerID: 1, OwnerName: "alice"})
	if err := svc.AddFavourite(ctx, 1, room.ID); err != nil {
		t.Fatalf("add favourite: %v", err)
	}
	favs, err := svc.ListFavourites(ctx, 1)
	if err != nil || len(favs) != 1 {
		t.Fatalf("list favourites err=%v len=%d", err, len(favs))
	}
	if err := svc.RemoveFavourite(ctx, 1, room.ID); err != nil {
		t.Fatalf("remove favourite: %v", err)
	}
	favs, _ = svc.ListFavourites(ctx, 1)
	if len(favs) != 0 {
		t.Fatalf("expected empty favourites after remove, got %d", len(favs))
	}
}

// Test11NavigatorSavedSearchFlow verifies end-to-end saved search create, list, and delete.
func Test11NavigatorSavedSearchFlow(t *testing.T) {
	svc := setupE2E(t)
	ctx := context.Background()
	ss, err := svc.CreateSavedSearch(ctx, domain.SavedSearch{UserID: 1, SearchCode: "hotel_view", Filter: ""})
	if err != nil {
		t.Fatalf("create saved search: %v", err)
	}
	if ss.ID == 0 {
		t.Fatalf("expected non-zero saved search id")
	}
	searches, err := svc.ListSavedSearches(ctx, 1)
	if err != nil || len(searches) != 1 {
		t.Fatalf("list searches err=%v len=%d", err, len(searches))
	}
	if err := svc.DeleteSavedSearch(ctx, ss.ID); err != nil {
		t.Fatalf("delete saved search: %v", err)
	}
	searches, _ = svc.ListSavedSearches(ctx, 1)
	if len(searches) != 0 {
		t.Fatalf("expected empty searches after delete, got %d", len(searches))
	}
}

// Test11NavigatorRoomSearchFilter verifies room search with text query filter.
func Test11NavigatorRoomSearchFilter(t *testing.T) {
	t.Skip("ILIKE not supported on SQLite; tested against PostgreSQL in CI")
}
