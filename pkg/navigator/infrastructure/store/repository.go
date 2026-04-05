package store

import (
	"fmt"
	"strings"

	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
	model "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/model"
	"gorm.io/gorm"
)

// Store persists navigator data using PostgreSQL via GORM.
type Store struct {
	// database stores the ORM client reference.
	database *gorm.DB
}

// NewRepository creates one PostgreSQL navigator repository.
func NewRepository(database *gorm.DB) (*Store, error) {
	if database == nil {
		return nil, fmt.Errorf("postgres database is required")
	}
	return &Store{database: database}, nil
}

// compile-time interface assertion.
var _ domain.Repository = (*Store)(nil)

// mapCategory converts persistence model to domain type.
func mapCategory(row model.Category) domain.Category {
	return domain.Category{
		ID: int(row.ID), Caption: row.Caption, Visible: row.Visible,
		OrderNum: row.OrderNum, IconImage: row.IconImage,
		CategoryType: row.CategoryType, CreatedAt: row.CreatedAt,
	}
}

// mapRoom converts persistence model to domain type.
func mapRoom(row model.Room) domain.Room {
	tags := splitTags(row.Tags)
	return domain.Room{
		ID: int(row.ID), OwnerID: int(row.OwnerID), OwnerName: row.OwnerName,
		Name: row.Name, Description: row.Description, State: row.State,
		CategoryID: int(row.CategoryID), MaxUsers: row.MaxUsers,
		Score: row.Score, Tags: tags, TradeMode: row.TradeMode,
		PromotedUntil: row.PromotedUntil, PromotionName: row.PromotionName,
		StaffPick: row.StaffPick,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

// splitTags splits comma-separated tag string into slice.
func splitTags(raw string) []string {
	if raw == "" {
		return nil
	}
	return strings.Split(raw, ",")
}

// joinTags joins tag slice into comma-separated string.
func joinTags(tags []string) string {
	return strings.Join(tags, ",")
}
