package store

import (
	"context"
	"errors"
	"strings"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LoadWardrobe resolves saved wardrobe slots for one user.
func (repository *Repository) LoadWardrobe(ctx context.Context, userID int) ([]domain.WardrobeSlot, error) {
	if _, err := repository.loadRecord(ctx, userID); err != nil {
		return nil, err
	}
	var rows []usermodel.WardrobeSlot
	if err := repository.database.WithContext(ctx).Where("user_id = ?", userID).Order("slot_id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	slots := make([]domain.WardrobeSlot, 0, len(rows))
	for _, row := range rows {
		slots = append(slots, domain.WardrobeSlot{SlotID: row.SlotID, Figure: row.Figure, Gender: row.Gender})
	}
	return slots, nil
}

// SaveWardrobeSlot upserts one wardrobe slot payload for one user.
func (repository *Repository) SaveWardrobeSlot(ctx context.Context, userID int, slot domain.WardrobeSlot) error {
	if _, err := repository.loadRecord(ctx, userID); err != nil {
		return err
	}
	row := usermodel.WardrobeSlot{UserID: uint(userID), SlotID: slot.SlotID, Figure: slot.Figure, Gender: slot.Gender}
	return repository.database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "slot_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"figure", "gender"}),
	}).Create(&row).Error
}

// ListIgnoredUsernames resolves ignored usernames for one user.
func (repository *Repository) ListIgnoredUsernames(ctx context.Context, userID int) ([]string, error) {
	if _, err := repository.loadRecord(ctx, userID); err != nil {
		return nil, err
	}
	usernames := []string{}
	query := repository.database.WithContext(ctx).Table("ignores").Select("users.username").Joins("JOIN users ON users.id = ignores.ignored_user_id")
	if err := query.Where("ignores.user_id = ? AND users.deleted_at IS NULL", userID).Order("users.username ASC").Scan(&usernames).Error; err != nil {
		return nil, err
	}
	return usernames, nil
}

// ListIgnoredUsers resolves ignored user entries for one user.
func (repository *Repository) ListIgnoredUsers(ctx context.Context, userID int) ([]domain.IgnoreEntry, error) {
	if _, err := repository.loadRecord(ctx, userID); err != nil {
		return nil, err
	}
	var rows []struct {
		UserID   int
		Username string
	}
	query := repository.database.WithContext(ctx).Table("ignores").Select("users.id AS user_id, users.username").Joins("JOIN users ON users.id = ignores.ignored_user_id")
	if err := query.Where("ignores.user_id = ? AND users.deleted_at IS NULL", userID).Order("users.username ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	entries := make([]domain.IgnoreEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, domain.IgnoreEntry{UserID: row.UserID, Username: row.Username})
	}
	return entries, nil
}

// IgnoreUserByUsername stores one ignore relation by target username.
func (repository *Repository) IgnoreUserByUsername(ctx context.Context, userID int, username string) (int, error) {
	target, err := repository.resolveTarget(ctx, strings.TrimSpace(username))
	if err != nil {
		return 0, err
	}
	return target.ID, repository.IgnoreUserByID(ctx, userID, target.ID)
}

// IgnoreUserByID stores one ignore relation by target user identifier.
func (repository *Repository) IgnoreUserByID(ctx context.Context, userID int, targetUserID int) error {
	if userID == targetUserID {
		return domain.ErrInvalidName
	}
	if _, err := repository.loadRecord(ctx, userID); err != nil {
		return err
	}
	if _, err := repository.loadRecord(ctx, targetUserID); err != nil {
		return err
	}
	row := usermodel.Ignore{UserID: uint(userID), IgnoredUserID: uint(targetUserID)}
	return repository.database.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

// UnignoreUserByUsername removes one ignore relation by target username.
func (repository *Repository) UnignoreUserByUsername(ctx context.Context, userID int, username string) (int, error) {
	target, err := repository.resolveTarget(ctx, strings.TrimSpace(username))
	if err != nil {
		return 0, err
	}
	result := repository.database.WithContext(ctx).Where("user_id = ? AND ignored_user_id = ?", userID, target.ID).Delete(&usermodel.Ignore{})
	if result.Error != nil {
		return 0, result.Error
	}
	return target.ID, nil
}

// UnignoreUserByID removes one ignore relation by target user identifier.
func (repository *Repository) UnignoreUserByID(ctx context.Context, userID int, targetUserID int) error {
	result := repository.database.WithContext(ctx).Where("user_id = ? AND ignored_user_id = ?", userID, targetUserID).Delete(&usermodel.Ignore{})
	return result.Error
}

// LoadProfile resolves one partial public profile payload.
func (repository *Repository) LoadProfile(ctx context.Context, userID int, openProfileWindow bool) (domain.Profile, error) {
	record, err := repository.loadRecord(ctx, userID)
	if err != nil {
		return domain.Profile{}, err
	}
	return domain.Profile{
		UserID: int(record.ID), Username: record.Username, Figure: record.Figure,
		Motto: record.Motto, IsOnline: true, OpenProfileWindow: openProfileWindow,
	}, nil
}

// resolveTarget resolves one target user by username.
func (repository *Repository) resolveTarget(ctx context.Context, username string) (domain.User, error) {
	if username == "" {
		return domain.User{}, domain.ErrUserNotFound
	}
	var record usermodel.Record
	query := repository.database.WithContext(ctx).Where("LOWER(username) = LOWER(?)", username).First(&record)
	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if query.Error != nil {
		return domain.User{}, query.Error
	}
	return mapUser(record), nil
}
