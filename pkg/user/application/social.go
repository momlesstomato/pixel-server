package application

import (
	"context"
	"fmt"
	"strings"

	sdkuser "github.com/momlesstomato/pixel-sdk/events/user"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// LoadWardrobe resolves saved wardrobe slots for one user.
func (service *Service) LoadWardrobe(ctx context.Context, userID int) ([]domain.WardrobeSlot, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.LoadWardrobe(ctx, userID)
}

// SaveWardrobeSlot validates and stores one wardrobe slot payload.
func (service *Service) SaveWardrobeSlot(ctx context.Context, userID int, slot domain.WardrobeSlot) error {
	if userID <= 0 {
		return fmt.Errorf("user id must be positive")
	}
	if slot.SlotID <= 0 || slot.SlotID > 50 {
		return fmt.Errorf("slot id must be between 1 and 50")
	}
	slot.Figure = strings.TrimSpace(slot.Figure)
	slot.Gender = strings.ToUpper(strings.TrimSpace(slot.Gender))
	if slot.Figure == "" || len(slot.Figure) > 255 {
		return fmt.Errorf("figure must be between 1 and 255 characters")
	}
	if slot.Gender != "M" && slot.Gender != "F" {
		return fmt.Errorf("gender must be M or F")
	}
	return service.repository.SaveWardrobeSlot(ctx, userID, slot)
}

// ListIgnoredUsernames resolves ignored usernames for one user.
func (service *Service) ListIgnoredUsernames(ctx context.Context, userID int) ([]string, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.ListIgnoredUsernames(ctx, userID)
}

// ListIgnoredUsers resolves ignored user entries for one user.
func (service *Service) ListIgnoredUsers(ctx context.Context, userID int) ([]domain.IgnoreEntry, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.ListIgnoredUsers(ctx, userID)
}

// AdminIgnoreUser stores one admin-initiated ignore relation.
func (service *Service) AdminIgnoreUser(ctx context.Context, userID int, targetUserID int) error {
	if userID <= 0 || targetUserID <= 0 {
		return fmt.Errorf("user ids must be positive")
	}
	return service.repository.IgnoreUserByID(ctx, userID, targetUserID)
}

// AdminUnignoreUser removes one admin-initiated ignore relation.
func (service *Service) AdminUnignoreUser(ctx context.Context, userID int, targetUserID int) error {
	if userID <= 0 || targetUserID <= 0 {
		return fmt.Errorf("user ids must be positive")
	}
	return service.repository.UnignoreUserByID(ctx, userID, targetUserID)
}

// IgnoreUserByUsername validates and stores one ignore relation by target username.
func (service *Service) IgnoreUserByUsername(ctx context.Context, connID string, userID int, username string) (int, error) {
	targetID, err := service.repository.IgnoreUserByUsername(ctx, userID, strings.TrimSpace(username))
	if err != nil {
		return 0, err
	}
	if service.fire != nil {
		event := &sdkuser.Ignored{ConnID: connID, UserID: userID, IgnoredUserID: targetID}
		service.fire(event)
		if event.Cancelled() {
			_, _ = service.repository.UnignoreUserByUsername(ctx, userID, strings.TrimSpace(username))
			return 0, fmt.Errorf("ignore cancelled by plugin")
		}
	}
	return targetID, nil
}

// IgnoreUserByID validates and stores one ignore relation by target user identifier.
func (service *Service) IgnoreUserByID(ctx context.Context, connID string, userID int, targetUserID int) error {
	if userID <= 0 || targetUserID <= 0 {
		return fmt.Errorf("user ids must be positive")
	}
	if service.fire != nil {
		event := &sdkuser.Ignored{ConnID: connID, UserID: userID, IgnoredUserID: targetUserID}
		service.fire(event)
		if event.Cancelled() {
			return fmt.Errorf("ignore cancelled by plugin")
		}
	}
	return service.repository.IgnoreUserByID(ctx, userID, targetUserID)
}

// UnignoreUserByUsername validates and removes one ignore relation by target username.
func (service *Service) UnignoreUserByUsername(ctx context.Context, connID string, userID int, username string) (int, error) {
	targetID, err := service.repository.UnignoreUserByUsername(ctx, userID, strings.TrimSpace(username))
	if err != nil {
		return 0, err
	}
	if service.fire != nil {
		event := &sdkuser.Unignored{ConnID: connID, UserID: userID, IgnoredUserID: targetID}
		service.fire(event)
		if event.Cancelled() {
			_ = service.repository.IgnoreUserByID(ctx, userID, targetID)
			return 0, fmt.Errorf("unignore cancelled by plugin")
		}
	}
	return targetID, nil
}

// LoadProfile resolves one user public profile payload for one viewer.
func (service *Service) LoadProfile(ctx context.Context, viewerUserID int, userID int, openProfileWindow bool) (domain.Profile, error) {
	if userID <= 0 {
		return domain.Profile{}, fmt.Errorf("user id must be positive")
	}
	if viewerUserID < 0 {
		return domain.Profile{}, fmt.Errorf("viewer user id must be zero or positive")
	}
	return service.repository.LoadProfile(ctx, viewerUserID, userID, openProfileWindow)
}

// ListRespects resolves respect audit rows for one target user.
func (service *Service) ListRespects(ctx context.Context, targetUserID int, limit int, offset int) ([]domain.RespectRecord, error) {
	if targetUserID <= 0 {
		return nil, fmt.Errorf("target user id must be positive")
	}
	return service.repository.ListRespects(ctx, targetUserID, limit, offset)
}
