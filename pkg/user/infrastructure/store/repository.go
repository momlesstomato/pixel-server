package store

import (
	"context"
	"errors"
	"fmt"
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

// loadRecord resolves one user record by identifier.
func (repository *Repository) loadRecord(ctx context.Context, id int) (usermodel.Record, error) {
	var record usermodel.Record
	err := repository.database.WithContext(ctx).First(&record, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return usermodel.Record{}, domain.ErrUserNotFound
	}
	if err != nil {
		return usermodel.Record{}, err
	}
	return record, nil
}

// mapUser converts one model record into domain user payload.
func mapUser(record usermodel.Record) domain.User {
	return domain.User{
		ID: int(record.ID), Username: record.Username, Figure: record.Figure,
		Gender: record.Gender, Motto: record.Motto, RealName: record.RealName,
		RespectsReceived: record.RespectsReceived, HomeRoomID: record.HomeRoomID,
		CanChangeName: record.CanChangeName, NoobnessLevel: record.NoobnessLevel,
		SafetyLocked: record.SafetyLocked, GroupID: int(record.GroupID),
	}
}

// utcDayStart returns UTC day start for one timestamp.
func utcDayStart(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
