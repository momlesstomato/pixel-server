package application

import (
	"context"
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
)

// FindVoucherByCode resolves one voucher by unique code.
func (service *Service) FindVoucherByCode(ctx context.Context, code string) (domain.Voucher, error) {
	if code == "" {
		return domain.Voucher{}, fmt.Errorf("voucher code is required")
	}
	return service.repository.FindVoucherByCode(ctx, code)
}

// CreateVoucher persists one validated voucher.
func (service *Service) CreateVoucher(ctx context.Context, v domain.Voucher) (domain.Voucher, error) {
	if v.Code == "" {
		return domain.Voucher{}, fmt.Errorf("voucher code is required")
	}
	return service.repository.CreateVoucher(ctx, v)
}

// DeleteVoucher removes one voucher by identifier.
func (service *Service) DeleteVoucher(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("voucher id must be positive")
	}
	return service.repository.DeleteVoucher(ctx, id)
}

// ListVouchers resolves all voucher rows.
func (service *Service) ListVouchers(ctx context.Context) ([]domain.Voucher, error) {
	return service.repository.ListVouchers(ctx)
}

// RedeemVoucher atomically validates and redeems one voucher for one user.
func (service *Service) RedeemVoucher(ctx context.Context, code string, userID int) (domain.Voucher, error) {
	if code == "" {
		return domain.Voucher{}, fmt.Errorf("voucher code is required")
	}
	if userID <= 0 {
		return domain.Voucher{}, fmt.Errorf("user id must be positive")
	}
	v, err := service.repository.FindVoucherByCode(ctx, code)
	if err != nil {
		return domain.Voucher{}, err
	}
	if !v.Enabled {
		return domain.Voucher{}, domain.ErrVoucherDisabled
	}
	if v.IsExhausted() {
		return domain.Voucher{}, domain.ErrVoucherExhausted
	}
	redeemed, err := service.repository.HasUserRedeemedVoucher(ctx, v.ID, userID)
	if err != nil {
		return domain.Voucher{}, err
	}
	if redeemed {
		return domain.Voucher{}, domain.ErrVoucherAlreadyRedeemed
	}
	if err := service.repository.RedeemVoucher(ctx, v.ID, userID); err != nil {
		return domain.Voucher{}, err
	}
	v.CurrentUses++
	return v, nil
}
