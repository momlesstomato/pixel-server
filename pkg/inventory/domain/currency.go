package domain

import "time"

// CurrencyType identifies an activity-point currency variant by its registered integer ID.
// These IDs match the wire-protocol activityPointType field and are stored in the
// currency_types table. The constants below are the three Habbo-standard seeds;
// operators may register additional types directly in the database.
type CurrencyType int

const (
	// CurrencyCredits is the reserved type ID for Habbo gold credits stored in user_currencies.
	// Type -1 is used to distinguish credits from all activity-point types, which start at 0.
	CurrencyCredits CurrencyType = -1
	// CurrencyDuckets is the ID for duckets (pixel coins); seeded as the default scroll currency.
	CurrencyDuckets CurrencyType = 0
	// CurrencyDiamonds is the ID for premium diamonds; seeded as the premium activity currency.
	CurrencyDiamonds CurrencyType = 5
	// CurrencySeasonal is the ID for seasonal / event points; seeded as the limited-time currency.
	CurrencySeasonal CurrencyType = 105
)

// ActivityCurrencyType represents one registered activity-point currency definition
// as stored in the currency_types table.
type ActivityCurrencyType struct {
	// ID stores the wire-protocol activity-point type identifier.
	ID int
	// Name stores the unique internal name for this currency.
	Name string
	// DisplayName stores the player-visible label for this currency.
	DisplayName string
	// Trackable reports whether balance changes are recorded in currency_transactions.
	Trackable bool
	// Enabled reports whether this currency is currently active.
	Enabled bool
}

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
