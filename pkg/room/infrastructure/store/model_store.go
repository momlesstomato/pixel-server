package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/infrastructure/model"
	"gorm.io/gorm"
)

// ModelStore persists room model data using PostgreSQL via GORM.
type ModelStore struct {
	// database stores the ORM client reference.
	database *gorm.DB
}

// NewModelStore creates one room model repository.
func NewModelStore(database *gorm.DB) (*ModelStore, error) {
	if database == nil {
		return nil, fmt.Errorf("postgres database is required")
	}
	return &ModelStore{database: database}, nil
}

// compile-time interface assertion.
var _ domain.ModelRepository = (*ModelStore)(nil)

// FindModelBySlug resolves one room model by slug identifier.
func (s *ModelStore) FindModelBySlug(ctx context.Context, slug string) (domain.RoomModel, error) {
	var row model.RoomModel
	if err := s.database.WithContext(ctx).Where("slug = ?", slug).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.RoomModel{}, domain.ErrRoomModelNotFound
		}
		return domain.RoomModel{}, err
	}
	return mapModel(row), nil
}

// ListModels returns all available room model templates.
func (s *ModelStore) ListModels(ctx context.Context) ([]domain.RoomModel, error) {
	var rows []model.RoomModel
	if err := s.database.WithContext(ctx).Order("slug ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.RoomModel, len(rows))
	for i, row := range rows {
		result[i] = mapModel(row)
	}
	return result, nil
}

// mapModel converts persistence model to domain type.
func mapModel(row model.RoomModel) domain.RoomModel {
	return domain.RoomModel{
		ID: int(row.ID), Slug: row.Slug, Heightmap: row.Heightmap,
		DoorX: row.DoorX, DoorY: row.DoorY, DoorZ: row.DoorZ,
		DoorDir: row.DoorDir, WallHeight: row.WallHeight,
	}
}
