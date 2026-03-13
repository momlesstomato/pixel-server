package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
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

// RecordLogin stores one successful login event and returns first-login-of-day status.
func (repository *Repository) RecordLogin(ctx context.Context, userID int, holder string, loggedAt time.Time) (bool, error) {
	dayStart := time.Date(loggedAt.UTC().Year(), loggedAt.UTC().Month(), loggedAt.UTC().Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)
	var count int64
	query := repository.database.WithContext(ctx).Model(&usermodel.LoginEvent{}).Where("user_id = ? AND logged_at >= ? AND logged_at < ?", userID, dayStart, dayEnd)
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	event := usermodel.LoginEvent{UserID: userID, Holder: strings.TrimSpace(holder), LoggedAt: loggedAt.UTC()}
	if err := repository.database.WithContext(ctx).Create(&event).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}
