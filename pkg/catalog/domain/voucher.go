package domain

import "time"

// Voucher defines one redeemable voucher code entry.
type Voucher struct {
	// ID stores stable voucher identifier.
	ID int
	// Code stores the unique redeemable code string.
	Code string
	// RewardType stores the reward category (currency, badge, furniture).
	RewardType string
	// RewardCurrencyType stores the currency type identifier when RewardType is "currency".
	RewardCurrencyType *int
	// RewardData stores reward-specific configuration payload.
	RewardData string
	// MaxUses stores the total allowed redemptions.
	MaxUses int
	// CurrentUses stores the current redemption count.
	CurrentUses int
	// Enabled stores whether the voucher is currently redeemable.
	Enabled bool
	// CreatedAt stores voucher creation timestamp.
	CreatedAt time.Time
}

// IsExhausted reports whether the voucher has reached max uses.
func (v Voucher) IsExhausted() bool {
	return v.CurrentUses >= v.MaxUses
}

// VoucherRedemption defines one per-user redemption audit row.
type VoucherRedemption struct {
	// ID stores stable redemption row identifier.
	ID int
	// VoucherID stores the redeemed voucher identifier.
	VoucherID int
	// UserID stores the redeeming user identifier.
	UserID int
	// RedeemedAt stores the redemption timestamp.
	RedeemedAt time.Time
}
