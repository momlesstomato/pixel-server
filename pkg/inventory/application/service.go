package application

import (
	"context"
	"fmt"

	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
)

// Service defines inventory application use-cases.
type Service struct {
	// repository stores inventory persistence contract implementation.
	repository domain.Repository
	// fire stores optional plugin event dispatch behavior.
	fire func(sdk.Event)
}

// NewService creates one inventory service.
func NewService(repository domain.Repository) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf("inventory repository is required")
	}
	return &Service{repository: repository}, nil
}

// SetEventFirer configures optional plugin event dispatch behavior.
func (service *Service) SetEventFirer(fire func(sdk.Event)) {
	service.fire = fire
}

// GetCredits resolves credit balance for one user.
func (service *Service) GetCredits(ctx context.Context, userID int) (int, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("user id must be positive")
	}
	return service.repository.GetCredits(ctx, userID)
}

// SetCredits updates credit balance for one user.
func (service *Service) SetCredits(ctx context.Context, userID int, credits int) error {
	if userID <= 0 {
		return fmt.Errorf("user id must be positive")
	}
	return service.repository.SetCredits(ctx, userID, credits)
}

// AddCredits atomically adds a signed credit amount and returns new balance.
func (service *Service) AddCredits(ctx context.Context, userID int, amount int) (int, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("user id must be positive")
	}
	return service.repository.AddCredits(ctx, userID, amount)
}

// GetCurrency resolves one activity-point balance.
func (service *Service) GetCurrency(ctx context.Context, userID int, ct domain.CurrencyType) (int, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("user id must be positive")
	}
	return service.repository.GetCurrency(ctx, userID, ct)
}

// ListCurrencies resolves all activity-point balances for one user.
func (service *Service) ListCurrencies(ctx context.Context, userID int) ([]domain.Currency, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.ListCurrencies(ctx, userID)
}

// AddCurrencyTracked atomically adds signed amount with transaction audit.
func (service *Service) AddCurrencyTracked(ctx context.Context, userID int, ct domain.CurrencyType, amount int, source domain.TransactionSource, refType string, refID string) (int, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("user id must be positive")
	}
	newBalance, err := service.repository.AddCurrency(ctx, userID, ct, amount)
	if err != nil {
		return 0, err
	}
	tx := domain.CurrencyTransaction{
		UserID: userID, CurrencyType: ct, Amount: amount,
		BalanceAfter: newBalance, Source: source,
		ReferenceType: refType, ReferenceID: refID,
	}
	if txErr := service.repository.RecordTransaction(ctx, tx); txErr != nil {
		return newBalance, txErr
	}
	return newBalance, nil
}

// ListTransactions resolves recent currency transactions.
func (service *Service) ListTransactions(ctx context.Context, userID int, ct domain.CurrencyType, limit int) ([]domain.CurrencyTransaction, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	if limit <= 0 {
		limit = 50
	}
	return service.repository.ListTransactions(ctx, userID, ct, limit)
}
