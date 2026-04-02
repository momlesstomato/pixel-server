package httpapi

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
)

// Service defines inventory API behavior required by HTTP routes.
type Service interface {
	// GetCredits resolves credit balance for one user.
	GetCredits(context.Context, int) (int, error)
	// AddCredits atomically adds a signed credit amount.
	AddCredits(context.Context, int, int) (int, error)
	// ListCurrencies resolves all activity-point balances for one user.
	ListCurrencies(context.Context, int) ([]domain.Currency, error)
	// AddCurrencyTracked atomically adds signed amount with transaction audit.
	AddCurrencyTracked(context.Context, int, domain.CurrencyType, int, domain.TransactionSource, string, string) (int, error)
	// ListBadges resolves all badge rows for one user.
	ListBadges(context.Context, int) ([]domain.Badge, error)
	// AwardBadge grants one badge to a user.
	AwardBadge(context.Context, int, string) (domain.Badge, error)
	// RevokeBadge removes one badge from a user.
	RevokeBadge(context.Context, int, string) error
	// UpdateBadgeSlots replaces equipped badge slot assignments.
	UpdateBadgeSlots(context.Context, int, []domain.BadgeSlot) error
	// GetEquippedBadges resolves currently equipped badge slots.
	GetEquippedBadges(context.Context, int) ([]domain.BadgeSlot, error)
	// ListEffects resolves all effect rows for one user.
	ListEffects(context.Context, int) ([]domain.Effect, error)
}
