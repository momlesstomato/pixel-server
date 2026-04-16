package migration

import (
	"testing"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	furnituremodel "github.com/momlesstomato/pixel-server/pkg/furniture/infrastructure/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestStep08BackfillZeroSpriteIDRepairsZeroRows verifies zero sprite identifiers are restored to the definition id.
func TestStep08BackfillZeroSpriteIDRepairsZeroRows(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&furnituremodel.Definition{}); err != nil {
		t.Fatalf("expected baseline migration success, got %v", err)
	}
	rows := []furnituremodel.Definition{{ID: 2081, ItemName: "hc_rllr", PublicName: "HC Roller", SpriteID: 0}, {ID: 26, ItemName: "chair_silo", PublicName: "Gray Dining Chair", SpriteID: 26}}
	if err := database.Create(&rows).Error; err != nil {
		t.Fatalf("expected seed row insert success, got %v", err)
	}
	migrator := gormigrate.New(database, nil, []*gormigrate.Migration{Step08BackfillZeroSpriteID()})
	if err := migrator.Migrate(); err != nil {
		t.Fatalf("expected migration success, got %v", err)
	}
	var repaired furnituremodel.Definition
	if err := database.First(&repaired, 2081).Error; err != nil {
		t.Fatalf("expected repaired definition query success, got %v", err)
	}
	if repaired.SpriteID != 2081 {
		t.Fatalf("expected sprite_id 2081, got %d", repaired.SpriteID)
	}
	var preserved furnituremodel.Definition
	if err := database.First(&preserved, 26).Error; err != nil {
		t.Fatalf("expected preserved definition query success, got %v", err)
	}
	if preserved.SpriteID != 26 {
		t.Fatalf("expected sprite_id 26 to remain unchanged, got %d", preserved.SpriteID)
	}
}