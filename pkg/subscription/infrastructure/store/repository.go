package store

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
	"gorm.io/gorm"
)

// Store persists subscription data using PostgreSQL via GORM.
type Store struct {
	// database stores the ORM client reference.
	database *gorm.DB
}

// NewRepository creates one PostgreSQL subscription repository.
func NewRepository(database *gorm.DB) (*Store, error) {
	if database == nil {
		return nil, fmt.Errorf("postgres database is required")
	}
	return &Store{database: database}, nil
}

// compile-time interface assertion.
var _ domain.Repository = (*Store)(nil)
