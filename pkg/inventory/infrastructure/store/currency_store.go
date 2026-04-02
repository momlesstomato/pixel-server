package store

import (
	"context"
	"errors"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
	inventorymodel "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/model"
	"gorm.io/gorm"
)

// GetCredits resolves credit balance for one user.
func (store *Store) GetCredits(ctx context.Context, userID int) (int, error) {
	return store.GetCurrency(ctx, userID, domain.CurrencyCredits)
}

// SetCredits updates credit balance for one user.
func (store *Store) SetCredits(ctx context.Context, userID int, credits int) error {
	return store.SetCurrency(ctx, userID, domain.CurrencyCredits, credits)
}

// AddCredits atomically adds a signed credit amount and returns new balance.
func (store *Store) AddCredits(ctx context.Context, userID int, amount int) (int, error) {
	return store.AddCurrency(ctx, userID, domain.CurrencyCredits, amount)
}

// GetCurrency resolves one activity-point balance for one user and type.
func (store *Store) GetCurrency(ctx context.Context, userID int, currencyType domain.CurrencyType) (int, error) {
	var row inventorymodel.Currency
	err := store.database.WithContext(ctx).Where("user_id = ? AND currency_type = ?", userID, int(currencyType)).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return row.Amount, nil
}

// ListCurrencies resolves all activity-point balances for one user.
func (store *Store) ListCurrencies(ctx context.Context, userID int) ([]domain.Currency, error) {
	var rows []inventorymodel.Currency
	if err := store.database.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Currency, len(rows))
	for i, row := range rows {
		result[i] = domain.Currency{UserID: int(row.UserID), Type: domain.CurrencyType(row.CurrencyType), Amount: row.Amount}
	}
	return result, nil
}

// SetCurrency updates one activity-point balance for one user and type.
func (store *Store) SetCurrency(ctx context.Context, userID int, currencyType domain.CurrencyType, amount int) error {
	row := inventorymodel.Currency{UserID: uint(userID), CurrencyType: int(currencyType), Amount: amount}
	return store.database.WithContext(ctx).
		Where("user_id = ? AND currency_type = ?", userID, int(currencyType)).
		Assign(inventorymodel.Currency{Amount: amount}).
		FirstOrCreate(&row).Error
}

// AddCurrency atomically adds signed amount to one currency and returns new balance.
func (store *Store) AddCurrency(ctx context.Context, userID int, currencyType domain.CurrencyType, amount int) (int, error) {
	var newBalance int
	err := store.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row inventorymodel.Currency
		resErr := tx.Where("user_id = ? AND currency_type = ?", userID, int(currencyType)).First(&row).Error
		if errors.Is(resErr, gorm.ErrRecordNotFound) {
			row = inventorymodel.Currency{UserID: uint(userID), CurrencyType: int(currencyType), Amount: 0}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		} else if resErr != nil {
			return resErr
		}
		newBalance = row.Amount + amount
		if newBalance < 0 {
			return domain.ErrInsufficientCurrency
		}
		return tx.Model(&row).Update("amount", newBalance).Error
	})
	return newBalance, err
}

// RecordTransaction persists one currency transaction audit row.
func (store *Store) RecordTransaction(ctx context.Context, tx domain.CurrencyTransaction) error {
	row := inventorymodel.CurrencyTransaction{
		UserID: uint(tx.UserID), CurrencyType: int(tx.CurrencyType),
		Amount: tx.Amount, BalanceAfter: tx.BalanceAfter,
		Reason: string(tx.Source), ReferenceType: tx.ReferenceType,
		CreatedAt: time.Now().UTC(),
	}
	return store.database.WithContext(ctx).Create(&row).Error
}

// ListTransactions resolves recent transactions for one user and type.
func (store *Store) ListTransactions(ctx context.Context, userID int, currencyType domain.CurrencyType, limit int) ([]domain.CurrencyTransaction, error) {
	var rows []inventorymodel.CurrencyTransaction
	err := store.database.WithContext(ctx).
		Where("user_id = ? AND currency_type = ?", userID, int(currencyType)).
		Order("created_at DESC").Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]domain.CurrencyTransaction, len(rows))
	for i, row := range rows {
		result[i] = domain.CurrencyTransaction{
			ID: int(row.ID), UserID: int(row.UserID),
			CurrencyType: domain.CurrencyType(row.CurrencyType),
			Amount: row.Amount, BalanceAfter: row.BalanceAfter,
			Source: domain.TransactionSource(row.Reason),
			ReferenceType: row.ReferenceType, CreatedAt: row.CreatedAt,
		}
	}
	return result, nil
}
