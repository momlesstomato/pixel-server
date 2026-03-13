package store

import (
	"context"
	"strings"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
)

// Create inserts one user row and returns domain representation.
func (repository *Repository) Create(ctx context.Context, username string) (domain.User, error) {
	record := usermodel.Record{Username: username}
	if err := repository.database.WithContext(ctx).Create(&record).Error; err != nil {
		return domain.User{}, err
	}
	return mapUser(record), nil
}

// FindByID loads one user row by identifier.
func (repository *Repository) FindByID(ctx context.Context, id int) (domain.User, error) {
	record, err := repository.loadRecord(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return mapUser(record), nil
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

// UpdateProfile applies partial identity updates and returns updated user payload.
func (repository *Repository) UpdateProfile(ctx context.Context, userID int, patch domain.ProfilePatch) (domain.User, error) {
	updates := map[string]any{}
	if patch.Figure != nil {
		updates["figure"] = strings.TrimSpace(*patch.Figure)
	}
	if patch.Gender != nil {
		updates["gender"] = strings.ToUpper(strings.TrimSpace(*patch.Gender))
	}
	if patch.Motto != nil {
		updates["motto"] = strings.TrimSpace(*patch.Motto)
	}
	if patch.HomeRoomID != nil {
		updates["home_room_id"] = *patch.HomeRoomID
	}
	if len(updates) > 0 {
		result := repository.database.WithContext(ctx).Model(&usermodel.Record{}).Where("id = ?", userID).Updates(updates)
		if result.Error != nil {
			return domain.User{}, result.Error
		}
		if result.RowsAffected == 0 {
			return domain.User{}, domain.ErrUserNotFound
		}
	}
	return repository.FindByID(ctx, userID)
}
