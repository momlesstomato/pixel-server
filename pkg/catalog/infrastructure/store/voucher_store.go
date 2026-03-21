package store

import (
	"context"
	"errors"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	catalogmodel "github.com/momlesstomato/pixel-server/pkg/catalog/infrastructure/model"
	"gorm.io/gorm"
)

// FindVoucherByCode resolves one voucher by unique code.
func (store *Store) FindVoucherByCode(ctx context.Context, code string) (domain.Voucher, error) {
	var row catalogmodel.Voucher
	if err := store.database.WithContext(ctx).Where("code = ?", code).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Voucher{}, domain.ErrVoucherNotFound
		}
		return domain.Voucher{}, err
	}
	return mapVoucher(row), nil
}

// CreateVoucher persists one voucher row.
func (store *Store) CreateVoucher(ctx context.Context, v domain.Voucher) (domain.Voucher, error) {
	row := catalogmodel.Voucher{
		Code: v.Code, RewardType: v.RewardType, RewardData: v.RewardData,
		MaxUses: v.MaxUses, Enabled: v.Enabled,
	}
	if err := store.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Voucher{}, err
	}
	return mapVoucher(row), nil
}

// DeleteVoucher removes one voucher by identifier.
func (store *Store) DeleteVoucher(ctx context.Context, id int) error {
	result := store.database.WithContext(ctx).Delete(&catalogmodel.Voucher{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrVoucherNotFound
	}
	return nil
}

// ListVouchers resolves all voucher rows.
func (store *Store) ListVouchers(ctx context.Context) ([]domain.Voucher, error) {
	var rows []catalogmodel.Voucher
	if err := store.database.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Voucher, len(rows))
	for i, row := range rows {
		result[i] = mapVoucher(row)
	}
	return result, nil
}

// RedeemVoucher atomically increments use count and records redemption.
func (store *Store) RedeemVoucher(ctx context.Context, voucherID int, userID int) error {
	return store.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&catalogmodel.Voucher{}).
			Where("id = ? AND current_uses < max_uses AND enabled = true", voucherID).
			Update("current_uses", gorm.Expr("current_uses + 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrVoucherExhausted
		}
		redemption := catalogmodel.VoucherRedemption{
			VoucherID: uint(voucherID), UserID: uint(userID),
			RedeemedAt: time.Now().UTC(),
		}
		return tx.Create(&redemption).Error
	})
}

// HasUserRedeemedVoucher checks per-user voucher redemption.
func (store *Store) HasUserRedeemedVoucher(ctx context.Context, voucherID int, userID int) (bool, error) {
	var count int64
	err := store.database.WithContext(ctx).Model(&catalogmodel.VoucherRedemption{}).
		Where("voucher_id = ? AND user_id = ?", voucherID, userID).
		Count(&count).Error
	return count > 0, err
}
