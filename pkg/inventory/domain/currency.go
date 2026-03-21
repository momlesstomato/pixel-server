package domain

import "time"

// CurrencyType identifies an activity-point currency variant.
type CurrencyType int

const (
	// CurrencyDuckets identifies the seasonal duckets currency.
	CurrencyDuckets CurrencyType = 0
	// CurrencyDiamonds identifies the premium diamonds currency.
	CurrencyDiamonds CurrencyType = 5
	// CurrencySeasonal identifies limited-time seasonal points.
	CurrencySeasonal CurrencyType = 105
)

// Currency holds one user balance for a specific currency type.
type Currency struct {
	// ID stores stable currency row identifier.
	ID int
	// UserID stores the currency owner identifier.
	UserID int
	// Type stores the currency variant.
	Type CurrencyType
	// Amount stores the current balance.
	Amount int
	// UpdatedAt stores the last balance change timestamp.
	UpdatedAt time.Time
}

// TransactionSource categorizes the origin of a currency transaction.
type TransactionSource string

const (
	// SourcePurchase identifies a catalog purchase transaction.
	SourcePurchase TransactionSource = "purchase"
	// SourceSale identifies a marketplace sale transaction.
	SourceSale TransactionSource = "sale"
	// SourceVoucher identifies a voucher redemption transaction.
	SourceVoucher TransactionSource = "voucher"
	// SourceTrade identifies a player-to-player trade.
	SourceTrade TransactionSource = "trade"
	// SourceAdmin identifies an administrative adjustment.
	SourceAdmin TransactionSource = "admin"
	// SourceMarketplace identifies a marketplace fee/refund.
	SourceMarketplace TransactionSource = "marketplace"
	// SourceReward identifies a reward or achievement payout.
	SourceReward TransactionSource = "reward"
)

// CurrencyTransaction records one trackable currency movement.
type CurrencyTransaction struct {
	// ID stores stable transaction identifier.
	ID int
	// UserID stores the affected user identifier.
	UserID int
	// CurrencyType stores the currency variant.
	CurrencyType CurrencyType
	// Amount stores the signed change value.
	Amount int
	// BalanceAfter stores the resulting balance.
	BalanceAfter int
	// Source stores the transaction origin category.
	Source TransactionSource
	// ReferenceType stores the related entity kind.
	ReferenceType string
	// ReferenceID stores the related entity identifier.
	ReferenceID string
	// CreatedAt stores the transaction timestamp.
	CreatedAt time.Time
}
