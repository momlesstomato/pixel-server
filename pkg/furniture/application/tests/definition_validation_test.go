package tests

import (
	"context"
	"testing"

	furnitureapplication "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
)

// TestCreateDefinitionRejectsNonPositiveSpriteID verifies sprite identifier validation.
func TestCreateDefinitionRejectsNonPositiveSpriteID(t *testing.T) {
	service, _ := furnitureapplication.NewService(repositoryStub{})
	_, err := service.CreateDefinition(context.Background(), domain.Definition{ItemName: "hc_rllr", SpriteID: 0})
	if err == nil {
		t.Fatalf("expected sprite id validation failure")
	}
}