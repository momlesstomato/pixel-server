package tests

import (
	"context"
	"testing"

	furnitureapplication "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
)

// TestToggleMultistateCyclesModes verifies multistate items advance and wrap across configured modes.
func TestToggleMultistateCyclesModes(t *testing.T) {
	service, _ := furnitureapplication.NewService(repositoryStub{
		item:       domain.Item{ID: 5, UserID: 1, RoomID: 9, DefinitionID: 3, ExtraData: "2"},
		definition: domain.Definition{ID: 3, InteractionModesCount: 3},
	})
	item, _, err := service.ToggleMultistate(context.Background(), 5, 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ExtraData != "0" {
		t.Fatalf("expected wrapped state 0, got %s", item.ExtraData)
	}
}

// TestStartDiceRollMarksRolling verifies dice activation writes the rolling sentinel state.
func TestStartDiceRollMarksRolling(t *testing.T) {
	service, _ := furnitureapplication.NewService(repositoryStub{
		item:       domain.Item{ID: 5, UserID: 1, RoomID: 9, DefinitionID: 3, ExtraData: "0"},
		definition: domain.Definition{ID: 3, InteractionType: domain.InteractionDice},
	})
	item, _, started, err := service.StartDiceRoll(context.Background(), 5, 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !started {
		t.Fatalf("expected dice roll to start")
	}
	if item.ExtraData != "-1" {
		t.Fatalf("expected rolling state -1, got %s", item.ExtraData)
	}
}

// TestSetStackHeightClampsBounds verifies stack-helper height overrides are bounded and encoded with two decimals.
func TestSetStackHeightClampsBounds(t *testing.T) {
	service, _ := furnitureapplication.NewService(repositoryStub{
		item:       domain.Item{ID: 5, UserID: 1, RoomID: 9, DefinitionID: 3},
		definition: domain.Definition{ID: 3, InteractionType: domain.InteractionStackHelper},
	})
	item, _, err := service.SetStackHeight(context.Background(), 5, 9, 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ExtraData != "40.00" {
		t.Fatalf("expected clamped height 40.00, got %s", item.ExtraData)
	}
}
