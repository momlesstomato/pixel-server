package postgresstore

import (
	"context"
	"errors"
	"fmt"

	usermodel "github.com/momlesstomato/pixel-server/core/postgres/model/user"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
	"gorm.io/gorm"
)

// Repository persists users using PostgreSQL via GORM.
type Repository struct {
	// database stores ORM client reference.
	database *gorm.DB
}

// NewRepository creates one PostgreSQL user repository.
func NewRepository(database *gorm.DB) (*Repository, error) {
	if database == nil {
		return nil, fmt.Errorf("postgres database is required")
	}
	return &Repository{database: database}, nil
}

// Create inserts one user row and returns domain representation.
func (repository *Repository) Create(ctx context.Context, username string) (domain.User, error) {
	record := usermodel.Record{Username: username}
	if err := repository.database.WithContext(ctx).Create(&record).Error; err != nil {
		return domain.User{}, err
	}
	return domain.User{ID: int(record.ID), Username: record.Username}, nil
}

// FindByID loads one user row by identifier.
func (repository *Repository) FindByID(ctx context.Context, id int) (domain.User, error) {
	var record usermodel.Record
	err := repository.database.WithContext(ctx).First(&record, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{ID: int(record.ID), Username: record.Username}, nil
}

// DeleteByID soft-deletes one user row by identifier.
func (repository *Repository) DeleteByID(ctx context.Context, id int) error {
	result := repository.database.WithContext(ctx).Delete(&usermodel.Record{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
