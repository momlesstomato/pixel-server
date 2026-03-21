package model

import "time"

// Voucher stores one redeemable voucher row in PostgreSQL.
type Voucher struct {
	// ID stores stable voucher identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// Code stores the unique redeemable code string.
	Code string `gorm:"size:128;uniqueIndex;not null"`
	// RewardType stores the reward category (currency, badge, furniture).
	RewardType string `gorm:"size:20;not null"`
	// RewardCurrencyType stores the currency type identifier when RewardType is "currency".
	RewardCurrencyType *int `gorm:"column:reward_currency_type"`
	// RewardData stores reward-specific configuration payload.
	RewardData string `gorm:"type:text;not null;default:''"`
	// MaxUses stores the total allowed redemptions.
	MaxUses int `gorm:"not null;default:1"`
	// CurrentUses stores the current redemption count.
	CurrentUses int `gorm:"not null;default:0"`
	// Enabled stores whether the voucher is currently redeemable.
	Enabled bool `gorm:"not null;default:true"`
	// CreatedAt stores voucher creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}

// TableName returns the PostgreSQL table name for Voucher.
func (Voucher) TableName() string { return "vouchers" }

// VoucherRedemption stores one per-user redemption audit row in PostgreSQL.
type VoucherRedemption struct {
	// ID stores stable redemption row identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// VoucherID stores the redeemed voucher identifier.
	VoucherID uint `gorm:"not null;index:idx_voucher_user,unique"`
	// UserID stores the redeeming user identifier.
	UserID uint `gorm:"not null;index:idx_voucher_user,unique"`
	// RedeemedAt stores the redemption timestamp.
	RedeemedAt time.Time `gorm:"not null"`
}

// TableName returns the PostgreSQL table name for VoucherRedemption.
func (VoucherRedemption) TableName() string { return "voucher_redemptions" }
