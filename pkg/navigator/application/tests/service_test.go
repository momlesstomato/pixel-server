package tests

import (
	"context"
	"testing"

	navigatorapplication "github.com/momlesstomato/pixel-server/pkg/navigator/application"
	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
)

// TestNewServiceRejectsNilRepository verifies constructor precondition validation.
func TestNewServiceRejectsNilRepository(t *testing.T) {
	if _, err := navigatorapplication.NewService(nil); err == nil {
		t.Fatalf("expected nil repository validation failure")
	}
}

// TestServiceCategoryCRUD verifies category create, find, list, and delete behavior.
func TestServiceCategoryCRUD(t *testing.T) {
	stub := repositoryStub{category: domain.Category{ID: 1, Caption: "Public"}}
	service, _ := navigatorapplication.NewService(stub)
	cats, err := service.ListCategories(context.Background())
	if err != nil || len(cats) != 1 || cats[0].Caption != "Public" {
		t.Fatalf("unexpected list result %+v err=%v", cats, err)
	}
	if _, err := service.FindCategoryByID(context.Background(), 0); err == nil {
		t.Fatalf("expected find failure for invalid id")
	}
	cat, err := service.FindCategoryByID(context.Background(), 1)
	if err != nil || cat.Caption != "Public" {
		t.Fatalf("unexpected find result %+v err=%v", cat, err)
	}
	if _, err := service.CreateCategory(context.Background(), domain.Category{}); err == nil {
		t.Fatalf("expected create failure for empty caption")
	}
	created, err := service.CreateCategory(context.Background(), domain.Category{Caption: "Test"})
	if err != nil || created.ID != 1 {
		t.Fatalf("unexpected create result %+v err=%v", created, err)
	}
	if err := service.DeleteCategory(context.Background(), 0); err == nil {
		t.Fatalf("expected delete failure for invalid id")
	}
}

// TestServiceRoomCRUD verifies room create, find, list, update, and delete behavior.
func TestServiceRoomCRUD(t *testing.T) {
	stub := repositoryStub{room: domain.Room{ID: 1, Name: "Lobby", OwnerID: 1}}
	service, _ := navigatorapplication.NewService(stub)
	rooms, total, err := service.ListRooms(context.Background(), domain.RoomFilter{})
	if err != nil || total != 1 || rooms[0].Name != "Lobby" {
		t.Fatalf("unexpected list result %+v total=%d err=%v", rooms, total, err)
	}
	if _, err := service.FindRoomByID(context.Background(), 0); err == nil {
		t.Fatalf("expected find failure for invalid id")
	}
	room, err := service.FindRoomByID(context.Background(), 1)
	if err != nil || room.Name != "Lobby" {
		t.Fatalf("unexpected find result %+v err=%v", room, err)
	}
	if _, err := service.CreateRoom(context.Background(), domain.Room{}); err == nil {
		t.Fatalf("expected create failure for empty name")
	}
	if _, err := service.CreateRoom(context.Background(), domain.Room{Name: "X"}); err == nil {
		t.Fatalf("expected create failure for missing owner")
	}
	created, err := service.CreateRoom(context.Background(), domain.Room{Name: "Test", OwnerID: 1})
	if err != nil || created.ID != 1 {
		t.Fatalf("unexpected create result %+v err=%v", created, err)
	}
	if _, err := service.UpdateRoom(context.Background(), 0, domain.RoomPatch{}); err == nil {
		t.Fatalf("expected update failure for invalid id")
	}
	if err := service.DeleteRoom(context.Background(), 0); err == nil {
		t.Fatalf("expected delete failure for invalid id")
	}
}
