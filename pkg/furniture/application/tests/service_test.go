package tests

import (
	"context"
	"errors"
	"testing"

	furnitureapplication "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
)

// TestNewServiceRejectsNilRepository verifies constructor precondition validation.
func TestNewServiceRejectsNilRepository(t *testing.T) {
	if _, err := furnitureapplication.NewService(nil); err == nil {
		t.Fatalf("expected nil repository validation failure")
	}
}

// TestServiceDefinitionCRUD verifies definition create and find behavior.
func TestServiceDefinitionCRUD(t *testing.T) {
	stub := repositoryStub{definition: domain.Definition{ID: 1, ItemName: "chair"}}
	service, _ := furnitureapplication.NewService(stub)
	if _, err := service.FindDefinitionByID(context.Background(), 0); err == nil {
		t.Fatalf("expected find failure for invalid id")
	}
	def, err := service.FindDefinitionByID(context.Background(), 1)
	if err != nil || def.ID != 1 {
		t.Fatalf("unexpected find result %+v err=%v", def, err)
	}
	if _, err := service.CreateDefinition(context.Background(), domain.Definition{}); err == nil {
		t.Fatalf("expected create failure for empty name")
	}
	created, err := service.CreateDefinition(context.Background(), domain.Definition{ItemName: "table"})
	if err != nil || created.ID != 1 {
		t.Fatalf("unexpected create result %+v err=%v", created, err)
	}
	defs, err := service.ListDefinitions(context.Background())
	if err != nil || len(defs) != 1 {
		t.Fatalf("unexpected list result len=%d err=%v", len(defs), err)
	}
}

// TestServiceItemCRUD verifies item create, find, transfer, and delete behavior.
func TestServiceItemCRUD(t *testing.T) {
	stub := repositoryStub{item: domain.Item{ID: 1, UserID: 1, DefinitionID: 1}}
	service, _ := furnitureapplication.NewService(stub)
	if _, err := service.FindItemByID(context.Background(), 0); err == nil {
		t.Fatalf("expected find failure for invalid id")
	}
	item, err := service.FindItemByID(context.Background(), 1)
	if err != nil || item.ID != 1 {
		t.Fatalf("unexpected find result %+v err=%v", item, err)
	}
	if _, err := service.CreateItem(context.Background(), domain.Item{}); err == nil {
		t.Fatalf("expected create failure for missing user id")
	}
	if _, err := service.CreateItem(context.Background(), domain.Item{UserID: 1}); err == nil {
		t.Fatalf("expected create failure for missing definition id")
	}
	if err := service.TransferItem(context.Background(), 0, 1); err == nil {
		t.Fatalf("expected transfer failure for invalid item id")
	}
	if err := service.TransferItem(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected transfer failure for invalid user id")
	}
	if err := service.DeleteItem(context.Background(), 0); err == nil {
		t.Fatalf("expected delete failure for invalid id")
	}
	if _, err := service.ListItemsByUserID(context.Background(), 0); err == nil {
		t.Fatalf("expected list failure for invalid user id")
	}
}

// TestServicePropagatesErrors verifies repository error propagation.
func TestServicePropagatesErrors(t *testing.T) {
	service, _ := furnitureapplication.NewService(repositoryStub{findErr: errors.New("boom"), deleteErr: errors.New("boom")})
	if _, err := service.FindDefinitionByID(context.Background(), 1); err == nil {
		t.Fatalf("expected find failure")
	}
	if err := service.DeleteDefinition(context.Background(), 1); err == nil {
		t.Fatalf("expected delete failure")
	}
}
