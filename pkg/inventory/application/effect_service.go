package application

import (
	"context"
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
)

// ListEffects resolves all effect rows for one user.
func (service *Service) ListEffects(ctx context.Context, userID int) ([]domain.Effect, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.ListEffects(ctx, userID)
}

// AwardEffect persists or increments one effect for one user.
func (service *Service) AwardEffect(ctx context.Context, userID int, effectID int, duration int, permanent bool) (domain.Effect, error) {
	if userID <= 0 {
		return domain.Effect{}, fmt.Errorf("user id must be positive")
	}
	if effectID <= 0 {
		return domain.Effect{}, fmt.Errorf("effect id must be positive")
	}
	return service.repository.AwardEffect(ctx, userID, effectID, duration, permanent)
}

// ActivateEffect sets activation timestamp for one effect.
func (service *Service) ActivateEffect(ctx context.Context, userID int, effectID int) (domain.Effect, error) {
	if userID <= 0 {
		return domain.Effect{}, fmt.Errorf("user id must be positive")
	}
	if effectID <= 0 {
		return domain.Effect{}, fmt.Errorf("effect id must be positive")
	}
	return service.repository.ActivateEffect(ctx, userID, effectID)
}

// RemoveExpiredEffects deletes all expired effects globally.
func (service *Service) RemoveExpiredEffects(ctx context.Context) ([]domain.ExpiredEffect, error) {
	return service.repository.RemoveExpiredEffects(ctx)
}
