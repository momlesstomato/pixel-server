package application

import (
	"context"
	"fmt"

	sdkinventory "github.com/momlesstomato/pixel-sdk/events/inventory"
	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
)

// ListBadges resolves all badge rows for one user.
func (service *Service) ListBadges(ctx context.Context, userID int) ([]domain.Badge, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.ListBadges(ctx, userID)
}

// AwardBadge grants one badge to a user.
func (service *Service) AwardBadge(ctx context.Context, userID int, badgeCode string) (domain.Badge, error) {
	if userID <= 0 {
		return domain.Badge{}, fmt.Errorf("user id must be positive")
	}
	if badgeCode == "" {
		return domain.Badge{}, fmt.Errorf("badge code is required")
	}
	if service.fire != nil {
		event := &sdkinventory.BadgeAwarding{UserID: userID, BadgeCode: badgeCode}
		service.fire(event)
		if event.Cancelled() {
			return domain.Badge{}, fmt.Errorf("badge award cancelled by plugin")
		}
	}
	badge, err := service.repository.AwardBadge(ctx, userID, badgeCode)
	if err == nil && service.fire != nil {
		service.fire(&sdkinventory.BadgeAwarded{UserID: userID, BadgeCode: badgeCode})
	}
	return badge, err
}

// RevokeBadge removes one badge from a user.
func (service *Service) RevokeBadge(ctx context.Context, userID int, badgeCode string) error {
	if userID <= 0 {
		return fmt.Errorf("user id must be positive")
	}
	if badgeCode == "" {
		return fmt.Errorf("badge code is required")
	}
	if service.fire != nil {
		event := &sdkinventory.BadgeRevoking{UserID: userID, BadgeCode: badgeCode}
		service.fire(event)
		if event.Cancelled() {
			return fmt.Errorf("badge revoke cancelled by plugin")
		}
	}
	err := service.repository.RevokeBadge(ctx, userID, badgeCode)
	if err == nil && service.fire != nil {
		service.fire(&sdkinventory.BadgeRevoked{UserID: userID, BadgeCode: badgeCode})
	}
	return err
}

// UpdateBadgeSlots replaces equipped badge slot assignments for one user.
func (service *Service) UpdateBadgeSlots(ctx context.Context, userID int, slots []domain.BadgeSlot) error {
	if userID <= 0 {
		return fmt.Errorf("user id must be positive")
	}
	if len(slots) > domain.MaxBadgeSlots {
		return fmt.Errorf("cannot equip more than %d badges", domain.MaxBadgeSlots)
	}
	return service.repository.UpdateBadgeSlots(ctx, userID, slots)
}

// GetEquippedBadges resolves currently equipped badge slots.
func (service *Service) GetEquippedBadges(ctx context.Context, userID int) ([]domain.BadgeSlot, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.GetEquippedBadges(ctx, userID)
}
