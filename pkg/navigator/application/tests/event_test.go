package tests

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdknavigator "github.com/momlesstomato/pixel-sdk/events/navigator"
	navigatorapplication "github.com/momlesstomato/pixel-server/pkg/navigator/application"
	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
)

// TestRoomCreateFiresEvents verifies event dispatch on room creation.
func TestRoomCreateFiresEvents(t *testing.T) {
	stub := repositoryStub{room: domain.Room{ID: 1, Name: "Lobby", OwnerID: 1}}
	service, _ := navigatorapplication.NewService(stub)
	var fired []string
	service.SetEventFirer(func(ev sdk.Event) {
		switch ev.(type) {
		case *sdknavigator.RoomCreating:
			fired = append(fired, "creating")
		case *sdknavigator.RoomCreated:
			fired = append(fired, "created")
		}
	})
	if _, err := service.CreateRoom(context.Background(), domain.Room{Name: "X", OwnerID: 1}); err != nil {
		t.Fatalf("unexpected create error %v", err)
	}
	if len(fired) != 2 || fired[0] != "creating" || fired[1] != "created" {
		t.Fatalf("expected [creating, created], got %v", fired)
	}
}

// TestRoomCreateCancelledByPlugin verifies plugin cancellation on room creation.
func TestRoomCreateCancelledByPlugin(t *testing.T) {
	stub := repositoryStub{}
	service, _ := navigatorapplication.NewService(stub)
	service.SetEventFirer(func(ev sdk.Event) {
		if c, ok := ev.(sdk.Cancellable); ok {
			c.Cancel()
		}
	})
	if _, err := service.CreateRoom(context.Background(), domain.Room{Name: "X", OwnerID: 1}); err == nil {
		t.Fatalf("expected cancellation error")
	}
}

// TestRoomDeleteFiresEvents verifies event dispatch on room deletion.
func TestRoomDeleteFiresEvents(t *testing.T) {
	stub := repositoryStub{}
	service, _ := navigatorapplication.NewService(stub)
	var fired []string
	service.SetEventFirer(func(ev sdk.Event) {
		switch ev.(type) {
		case *sdknavigator.RoomDeleting:
			fired = append(fired, "deleting")
		case *sdknavigator.RoomDeleted:
			fired = append(fired, "deleted")
		}
	})
	if err := service.DeleteRoom(context.Background(), 1); err != nil {
		t.Fatalf("unexpected delete error %v", err)
	}
	if len(fired) != 2 || fired[0] != "deleting" || fired[1] != "deleted" {
		t.Fatalf("expected [deleting, deleted], got %v", fired)
	}
}

// TestRoomDeleteCancelledByPlugin verifies plugin cancellation on room deletion.
func TestRoomDeleteCancelledByPlugin(t *testing.T) {
	stub := repositoryStub{}
	service, _ := navigatorapplication.NewService(stub)
	service.SetEventFirer(func(ev sdk.Event) {
		if c, ok := ev.(sdk.Cancellable); ok {
			c.Cancel()
		}
	})
	if err := service.DeleteRoom(context.Background(), 1); err == nil {
		t.Fatalf("expected cancellation error")
	}
}

// TestFavouriteAddFiresEvents verifies event dispatch on favourite addition.
func TestFavouriteAddFiresEvents(t *testing.T) {
	stub := repositoryStub{}
	service, _ := navigatorapplication.NewService(stub)
	var fired []string
	service.SetEventFirer(func(ev sdk.Event) {
		switch ev.(type) {
		case *sdknavigator.FavouriteAdding:
			fired = append(fired, "adding")
		case *sdknavigator.FavouriteAdded:
			fired = append(fired, "added")
		}
	})
	if err := service.AddFavourite(context.Background(), 1, 1); err != nil {
		t.Fatalf("unexpected add error %v", err)
	}
	if len(fired) != 2 || fired[0] != "adding" || fired[1] != "added" {
		t.Fatalf("expected [adding, added], got %v", fired)
	}
}

// TestFavouriteAddCancelledByPlugin verifies plugin cancellation on favourite add.
func TestFavouriteAddCancelledByPlugin(t *testing.T) {
	stub := repositoryStub{}
	service, _ := navigatorapplication.NewService(stub)
	service.SetEventFirer(func(ev sdk.Event) {
		if c, ok := ev.(sdk.Cancellable); ok {
			c.Cancel()
		}
	})
	if err := service.AddFavourite(context.Background(), 1, 1); err == nil {
		t.Fatalf("expected cancellation error")
	}
}
