package tests

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkfurniture "github.com/momlesstomato/pixel-sdk/events/furniture"
	furnitureapplication "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
)

// TestDefinitionCreatingEventCancelsCreation verifies DefinitionCreating cancellation aborts creation.
func TestDefinitionCreatingEventCancelsCreation(t *testing.T) {
	service, _ := furnitureapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if value, ok := event.(*sdkfurniture.DefinitionCreating); ok {
			value.Cancel()
		}
	})
	if _, err := service.CreateDefinition(context.Background(), domain.Definition{ItemName: "chair", SpriteID: 1}); err == nil {
		t.Fatalf("expected definition creation to be cancelled")
	}
}

// TestDefinitionCreatingEventAllowsCreation verifies DefinitionCreating passes and fires after event.
func TestDefinitionCreatingEventAllowsCreation(t *testing.T) {
	var afterFired bool
	service, _ := furnitureapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if _, ok := event.(*sdkfurniture.DefinitionCreated); ok {
			afterFired = true
		}
	})
	def, err := service.CreateDefinition(context.Background(), domain.Definition{ItemName: "table", SpriteID: 2})
	if err != nil || def.ID != 1 {
		t.Fatalf("unexpected create result %+v err=%v", def, err)
	}
	if !afterFired {
		t.Fatalf("expected DefinitionCreated event to fire")
	}
}

// TestDefinitionDeletingEventCancelsDeletion verifies DefinitionDeleting cancellation aborts deletion.
func TestDefinitionDeletingEventCancelsDeletion(t *testing.T) {
	service, _ := furnitureapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if value, ok := event.(*sdkfurniture.DefinitionDeleting); ok {
			value.Cancel()
		}
	})
	if err := service.DeleteDefinition(context.Background(), 1); err == nil {
		t.Fatalf("expected definition deletion to be cancelled")
	}
}

// TestDefinitionDeletingEventAllowsDeletion verifies DefinitionDeleting passes and fires after event.
func TestDefinitionDeletingEventAllowsDeletion(t *testing.T) {
	var afterFired bool
	service, _ := furnitureapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if _, ok := event.(*sdkfurniture.DefinitionDeleted); ok {
			afterFired = true
		}
	})
	if err := service.DeleteDefinition(context.Background(), 1); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if !afterFired {
		t.Fatalf("expected DefinitionDeleted event to fire")
	}
}
