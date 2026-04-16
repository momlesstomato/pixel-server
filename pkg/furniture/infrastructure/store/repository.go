package store

import (
	"fmt"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
	furnituremodel "github.com/momlesstomato/pixel-server/pkg/furniture/infrastructure/model"
	"gorm.io/gorm"
)

// Store persists furniture data using PostgreSQL via GORM.
type Store struct {
	// database stores the ORM client reference.
	database *gorm.DB
}

// NewRepository creates one PostgreSQL furniture repository.
func NewRepository(database *gorm.DB) (*Store, error) {
	if database == nil {
		return nil, fmt.Errorf("postgres database is required")
	}
	return &Store{database: database}, nil
}

// compile-time interface assertion.
var _ domain.Repository = (*Store)(nil)

// mapDefinition converts one GORM model into domain definition.
func mapDefinition(row furnituremodel.Definition) domain.Definition {
	return domain.Definition{
		ID: int(row.ID), ItemName: row.ItemName, PublicName: row.PublicName,
		ItemType: domain.ItemType(row.ItemType), Width: int(row.Width), Length: int(row.Length),
		StackHeight: row.StackHeight, CanStack: row.CanStack, CanSit: row.CanSit, CanLay: row.CanLay,
		IsWalkable: row.IsWalkable, SpriteID: row.SpriteID,
		AllowRecycle: row.AllowRecycle, AllowTrade: row.AllowTrade,
		AllowMarketplaceSell: row.AllowMarketplaceSell, AllowGift: row.AllowGift,
		AllowInventoryStack:   row.AllowInventoryStack,
		InteractionType:       domain.InteractionType(row.InteractionType),
		InteractionModesCount: int(row.InteractionModesCount),
		EffectID:              row.EffectID,
	}
}

// mapItem converts one GORM model into domain item.
func mapItem(row furnituremodel.Item) domain.Item {
	return domain.Item{
		ID: int(row.ID), UserID: int(row.UserID), RoomID: int(row.RoomID),
		DefinitionID: int(row.DefinitionID), ExtraData: row.ExtraData,
		InteractionData: row.InteractionData,
		LimitedNumber:   row.LimitedNumber, LimitedTotal: row.LimitedTotal,
		X: row.X, Y: row.Y, Z: row.Z, Dir: row.Dir,
		WallPosition: row.WallPosition,
		CreatedAt:    row.CreatedAt,
	}
}

// toItemRecord converts domain item into GORM model for persistence.
func toItemRecord(item domain.Item) furnituremodel.Item {
	return furnituremodel.Item{
		UserID: uint(item.UserID), DefinitionID: uint(item.DefinitionID),
		ExtraData: item.ExtraData, InteractionData: item.InteractionData,
		LimitedNumber: item.LimitedNumber,
		LimitedTotal:  item.LimitedTotal, CreatedAt: time.Now().UTC(),
	}
}
