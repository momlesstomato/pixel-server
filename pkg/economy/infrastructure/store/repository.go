package store

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/economy/domain"
	"gorm.io/gorm"
)

// Store persists economy data using PostgreSQL via GORM.
type Store struct {
	// database stores the ORM client reference.
	database *gorm.DB
}

// NewRepository creates one PostgreSQL economy repository.
func NewRepository(database *gorm.DB) (*Store, error) {
	if database == nil {
		return nil, fmt.Errorf("postgres database is required")
	}
	return &Store{database: database}, nil
}

// compile-time interface assertion.
var _ domain.Repository = (*Store)(nil)
