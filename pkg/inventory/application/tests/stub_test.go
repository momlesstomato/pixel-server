package tests

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
)

// repositoryStub defines deterministic inventory repository behavior.
type repositoryStub struct {
	// credits stores deterministic credits balance.
	credits int
	// badge stores deterministic badge return.
	badge domain.Badge
	// effect stores deterministic effect return.
	effect domain.Effect
	// findErr stores deterministic find error.
	findErr error
	// deleteErr stores deterministic delete error.
	deleteErr error
	// currencies stores optional deterministic currency list; nil uses the default.
	currencies []domain.Currency
}

// ListCurrencyTypes returns a deterministic list of standard currency types.
func (s repositoryStub) ListCurrencyTypes(_ context.Context) ([]domain.ActivityCurrencyType, error) {
	return []domain.ActivityCurrencyType{
		{ID: 0, Name: "duckets", DisplayName: "Duckets", Trackable: true, Enabled: true},
		{ID: 5, Name: "diamonds", DisplayName: "Diamonds", Trackable: true, Enabled: true},
	}, nil
}

// FindCurrencyTypeByID returns a deterministic currency type definition.
func (s repositoryStub) FindCurrencyTypeByID(_ context.Context, id int) (domain.ActivityCurrencyType, error) {
	return domain.ActivityCurrencyType{ID: id, Name: "duckets", DisplayName: "Duckets", Enabled: true}, s.findErr
}

// ListBadges returns deterministic badge list.
func (s repositoryStub) ListBadges(_ context.Context, _ int) ([]domain.Badge, error) {
	return []domain.Badge{s.badge}, s.findErr
}

// AwardBadge returns deterministic badge.
func (s repositoryStub) AwardBadge(_ context.Context, _ int, code string) (domain.Badge, error) {
	return domain.Badge{ID: 1, BadgeCode: code}, nil
}

// RevokeBadge returns deterministic error.
func (s repositoryStub) RevokeBadge(_ context.Context, _ int, _ string) error {
	return s.deleteErr
}

// UpdateBadgeSlots returns deterministic error.
func (s repositoryStub) UpdateBadgeSlots(_ context.Context, _ int, _ []domain.BadgeSlot) error {
	return nil
}

// GetEquippedBadges returns deterministic badge slots.
func (s repositoryStub) GetEquippedBadges(_ context.Context, _ int) ([]domain.BadgeSlot, error) {
	return []domain.BadgeSlot{{SlotID: 1, BadgeCode: "ACH1"}}, nil
}

// GetCredits returns deterministic credits.
func (s repositoryStub) GetCredits(_ context.Context, _ int) (int, error) {
	return s.credits, s.findErr
}

// SetCredits returns deterministic error.
func (s repositoryStub) SetCredits(_ context.Context, _ int, _ int) error {
	return nil
}

// AddCredits returns deterministic new balance.
func (s repositoryStub) AddCredits(_ context.Context, _ int, amount int) (int, error) {
	return s.credits + amount, nil
}

// GetCurrency returns deterministic currency balance.
func (s repositoryStub) GetCurrency(_ context.Context, _ int, _ domain.CurrencyType) (int, error) {
	return 100, s.findErr
}

// ListCurrencies returns deterministic currency list.
func (s repositoryStub) ListCurrencies(_ context.Context, _ int) ([]domain.Currency, error) {
	if s.currencies != nil {
		return s.currencies, nil
	}
	return []domain.Currency{{ID: 1, Type: domain.CurrencyDuckets, Amount: 100}}, nil
}

// SetCurrency returns deterministic error.
func (s repositoryStub) SetCurrency(_ context.Context, _ int, _ domain.CurrencyType, _ int) error {
	return nil
}

// AddCurrency returns deterministic new balance.
func (s repositoryStub) AddCurrency(_ context.Context, _ int, _ domain.CurrencyType, amount int) (int, error) {
	return 100 + amount, nil
}

// RecordTransaction returns deterministic error.
func (s repositoryStub) RecordTransaction(_ context.Context, _ domain.CurrencyTransaction) error {
	return nil
}

// ListTransactions returns deterministic transaction list.
func (s repositoryStub) ListTransactions(_ context.Context, _ int, _ domain.CurrencyType, _ int) ([]domain.CurrencyTransaction, error) {
	return []domain.CurrencyTransaction{}, nil
}

// ListEffects returns deterministic effect list.
func (s repositoryStub) ListEffects(_ context.Context, _ int) ([]domain.Effect, error) {
	return []domain.Effect{s.effect}, nil
}

// AwardEffect returns deterministic effect.
func (s repositoryStub) AwardEffect(_ context.Context, _ int, effectID int, _ int, _ bool) (domain.Effect, error) {
	return domain.Effect{ID: 1, EffectID: effectID}, nil
}

// ActivateEffect returns deterministic effect.
func (s repositoryStub) ActivateEffect(_ context.Context, _ int, effectID int) (domain.Effect, error) {
	return domain.Effect{ID: 1, EffectID: effectID}, nil
}

// RemoveExpiredEffects returns deterministic expired effects.
func (s repositoryStub) RemoveExpiredEffects(_ context.Context) ([]domain.ExpiredEffect, error) {
	return []domain.ExpiredEffect{}, nil
}
