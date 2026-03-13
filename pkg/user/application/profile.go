package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// RespectResult defines user respect operation output.
type RespectResult struct {
	// RespectsReceived stores updated target respects received total.
	RespectsReceived int
	// Remaining stores actor remaining daily respects for user target type.
	Remaining int
}

// UpdateProfile validates and applies partial user identity updates.
func (service *Service) UpdateProfile(ctx context.Context, userID int, patch domain.ProfilePatch) (domain.User, error) {
	if userID <= 0 {
		return domain.User{}, fmt.Errorf("user id must be positive")
	}
	if patch.Figure != nil {
		trimmed := strings.TrimSpace(*patch.Figure)
		if trimmed == "" || len(trimmed) > 255 {
			return domain.User{}, fmt.Errorf("figure must be between 1 and 255 characters")
		}
		patch.Figure = &trimmed
	}
	if patch.Gender != nil {
		value := strings.ToUpper(strings.TrimSpace(*patch.Gender))
		if value != "M" && value != "F" {
			return domain.User{}, fmt.Errorf("gender must be M or F")
		}
		patch.Gender = &value
	}
	if patch.Motto != nil {
		value := strings.TrimSpace(*patch.Motto)
		if len(value) > 127 {
			return domain.User{}, fmt.Errorf("motto must be <= 127 characters")
		}
		patch.Motto = &value
	}
	if patch.HomeRoomID != nil && *patch.HomeRoomID < -1 {
		return domain.User{}, fmt.Errorf("home room id must be >= -1")
	}
	return service.repository.UpdateProfile(ctx, userID, patch)
}

// LoadSettings resolves one user settings payload.
func (service *Service) LoadSettings(ctx context.Context, userID int) (domain.Settings, error) {
	if userID <= 0 {
		return domain.Settings{}, fmt.Errorf("user id must be positive")
	}
	return service.repository.LoadSettings(ctx, userID)
}

// SaveSettings validates and applies one partial settings payload.
func (service *Service) SaveSettings(ctx context.Context, userID int, patch domain.SettingsPatch) (domain.Settings, error) {
	if userID <= 0 {
		return domain.Settings{}, fmt.Errorf("user id must be positive")
	}
	if err := validateSettingsPatch(patch); err != nil {
		return domain.Settings{}, err
	}
	return service.repository.SaveSettings(ctx, userID, patch)
}

// RecordUserRespect validates and stores one user-to-user respect event.
func (service *Service) RecordUserRespect(ctx context.Context, actorUserID int, targetUserID int, at time.Time) (RespectResult, error) {
	if actorUserID <= 0 || targetUserID <= 0 {
		return RespectResult{}, fmt.Errorf("actor and target user id must be positive")
	}
	if actorUserID == targetUserID {
		return RespectResult{}, fmt.Errorf("actor cannot respect itself")
	}
	if at.IsZero() {
		return RespectResult{}, fmt.Errorf("respect timestamp is required")
	}
	received, err := service.repository.RecordRespect(ctx, actorUserID, targetUserID, domain.RespectTargetUser, at.UTC())
	if err != nil {
		return RespectResult{}, err
	}
	remaining, err := service.repository.RemainingRespects(ctx, actorUserID, domain.RespectTargetUser, at.UTC())
	if err != nil {
		return RespectResult{}, err
	}
	return RespectResult{RespectsReceived: received, Remaining: remaining}, nil
}

// RemainingRespects resolves remaining respects for one user and target type.
func (service *Service) RemainingRespects(ctx context.Context, userID int, targetType domain.RespectTargetType, at time.Time) (int, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("user id must be positive")
	}
	if at.IsZero() {
		return 0, fmt.Errorf("timestamp is required")
	}
	return service.repository.RemainingRespects(ctx, userID, targetType, at.UTC())
}

// validateSettingsPatch validates settings patch field ranges.
func validateSettingsPatch(patch domain.SettingsPatch) error {
	for _, value := range []*int{patch.VolumeSystem, patch.VolumeFurni, patch.VolumeTrax} {
		if value != nil && (*value < 0 || *value > 100) {
			return fmt.Errorf("volume values must be between 0 and 100")
		}
	}
	if patch.ChatType != nil && (*patch.ChatType < 0 || *patch.ChatType > 1) {
		return fmt.Errorf("chat type must be between 0 and 1")
	}
	if patch.Flags != nil && *patch.Flags < 0 {
		return fmt.Errorf("flags must be >= 0")
	}
	return nil
}
