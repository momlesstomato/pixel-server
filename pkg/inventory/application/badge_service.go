package application

import (
	"context"
	"fmt"

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
	return service.repository.AwardBadge(ctx, userID, badgeCode)
}

// RevokeBadge removes one badge from a user.
func (service *Service) RevokeBadge(ctx context.Context, userID int, badgeCode string) error {
	if userID <= 0 {
		return fmt.Errorf("user id must be positive")
	}
	if badgeCode == "" {
		return fmt.Errorf("badge code is required")
	}
	return service.repository.RevokeBadge(ctx, userID, badgeCode)
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
