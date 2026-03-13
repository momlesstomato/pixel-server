package store

import (
	"context"
	"errors"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// LoadSettings resolves user settings and lazily creates defaults when missing.
func (repository *Repository) LoadSettings(ctx context.Context, userID int) (domain.Settings, error) {
	record, err := repository.loadSettings(ctx, userID)
	if err != nil {
		return domain.Settings{}, err
	}
	return mapSettings(record), nil
}

// SaveSettings applies partial settings update and returns updated settings payload.
func (repository *Repository) SaveSettings(ctx context.Context, userID int, patch domain.SettingsPatch) (domain.Settings, error) {
	record, err := repository.loadSettings(ctx, userID)
	if err != nil {
		return domain.Settings{}, err
	}
	updates := map[string]any{}
	if patch.VolumeSystem != nil {
		updates["volume_system"] = *patch.VolumeSystem
	}
	if patch.VolumeFurni != nil {
		updates["volume_furni"] = *patch.VolumeFurni
	}
	if patch.VolumeTrax != nil {
		updates["volume_trax"] = *patch.VolumeTrax
	}
	if patch.OldChat != nil {
		updates["old_chat"] = *patch.OldChat
	}
	if patch.RoomInvites != nil {
		updates["room_invites"] = *patch.RoomInvites
	}
	if patch.CameraFollow != nil {
		updates["camera_follow"] = *patch.CameraFollow
	}
	if patch.Flags != nil {
		updates["flags"] = *patch.Flags
	}
	if patch.ChatType != nil {
		updates["chat_type"] = *patch.ChatType
	}
	if len(updates) > 0 {
		if err := repository.database.WithContext(ctx).Model(&usermodel.Settings{}).Where("id = ?", record.ID).Updates(updates).Error; err != nil {
			return domain.Settings{}, err
		}
	}
	return repository.LoadSettings(ctx, userID)
}

// loadSettings resolves settings row and creates defaults when missing.
func (repository *Repository) loadSettings(ctx context.Context, userID int) (usermodel.Settings, error) {
	if _, err := repository.loadRecord(ctx, userID); err != nil {
		return usermodel.Settings{}, err
	}
	var settings usermodel.Settings
	err := repository.database.WithContext(ctx).Where("user_id = ?", userID).First(&settings).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		settings = usermodel.Settings{UserID: uint(userID)}
		if createErr := repository.database.WithContext(ctx).Create(&settings).Error; createErr != nil {
			return usermodel.Settings{}, createErr
		}
		return settings, nil
	}
	if err != nil {
		return usermodel.Settings{}, err
	}
	return settings, nil
}

// mapSettings converts one settings model into domain payload.
func mapSettings(settings usermodel.Settings) domain.Settings {
	return domain.Settings{
		UserID: int(settings.UserID), VolumeSystem: settings.VolumeSystem,
		VolumeFurni: settings.VolumeFurni, VolumeTrax: settings.VolumeTrax,
		OldChat: settings.OldChat, RoomInvites: settings.RoomInvites,
		CameraFollow: settings.CameraFollow, Flags: settings.Flags, ChatType: settings.ChatType,
	}
}
