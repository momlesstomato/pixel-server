package domain

import "context"

// Repository defines inventory persistence behavior.
type Repository interface {
	// ListBadges resolves all badge rows for one user.
	ListBadges(ctx context.Context, userID int) ([]Badge, error)
	// AwardBadge persists one badge for one user.
	AwardBadge(ctx context.Context, userID int, badgeCode string) (Badge, error)
	// RevokeBadge removes one badge by user and code.
	RevokeBadge(ctx context.Context, userID int, badgeCode string) error
	// UpdateBadgeSlots replaces equipped badge slot assignments for one user.
	UpdateBadgeSlots(ctx context.Context, userID int, slots []BadgeSlot) error
	// GetEquippedBadges resolves currently equipped badge slots for one user.
	GetEquippedBadges(ctx context.Context, userID int) ([]BadgeSlot, error)
	// GetCredits resolves credit balance for one user.
	GetCredits(ctx context.Context, userID int) (int, error)
	// SetCredits updates credit balance for one user.
	SetCredits(ctx context.Context, userID int, credits int) error
	// AddCredits atomically adds a signed credit amount and returns new balance.
	AddCredits(ctx context.Context, userID int, amount int) (int, error)
	// GetCurrency resolves one activity-point balance for one user and type.
	GetCurrency(ctx context.Context, userID int, currencyType CurrencyType) (int, error)
	// ListCurrencies resolves all activity-point balances for one user.
	ListCurrencies(ctx context.Context, userID int) ([]Currency, error)
	// SetCurrency updates one activity-point balance for one user and type.
	SetCurrency(ctx context.Context, userID int, currencyType CurrencyType, amount int) error
	// AddCurrency atomically adds signed amount to one currency and returns new balance.
	AddCurrency(ctx context.Context, userID int, currencyType CurrencyType, amount int) (int, error)
	// RecordTransaction persists one currency transaction audit row.
	RecordTransaction(ctx context.Context, tx CurrencyTransaction) error
	// ListTransactions resolves recent transactions for one user and type.
	ListTransactions(ctx context.Context, userID int, currencyType CurrencyType, limit int) ([]CurrencyTransaction, error)
	// ListEffects resolves all effect rows for one user.
	ListEffects(ctx context.Context, userID int) ([]Effect, error)
	// AwardEffect persists or increments one effect for one user.
	AwardEffect(ctx context.Context, userID int, effectID int, duration int, permanent bool) (Effect, error)
	// ActivateEffect sets activation timestamp for one effect.
	ActivateEffect(ctx context.Context, userID int, effectID int) (Effect, error)
	// RemoveExpiredEffects deletes all expired effects and returns removed IDs.
	RemoveExpiredEffects(ctx context.Context) ([]ExpiredEffect, error)
}
