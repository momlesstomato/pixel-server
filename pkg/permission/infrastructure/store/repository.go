package store

import (
	"fmt"

	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	"gorm.io/gorm"
)

// Repository persists permission groups and grants using PostgreSQL via GORM.
type Repository struct {
	// database stores ORM client reference.
	database *gorm.DB
}

// NewRepository creates one PostgreSQL permission repository.
func NewRepository(database *gorm.DB) (*Repository, error) {
	if database == nil {
		return nil, fmt.Errorf("postgres database is required")
	}
	return &Repository{database: database}, nil
}

// mapGroup converts one persisted group model into domain payload.
func mapGroup(value permissionmodel.Group) permissiondomain.Group {
	return permissiondomain.Group{
		ID: int(value.ID), Name: value.Name, DisplayName: value.DisplayName, Priority: value.Priority,
		ClubLevel: value.ClubLevel, SecurityLevel: value.SecurityLevel, IsAmbassador: value.IsAmbassador, IsDefault: value.IsDefault,
	}
}
