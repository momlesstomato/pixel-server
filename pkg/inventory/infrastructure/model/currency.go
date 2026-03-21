package model

import "time"

// Currency stores one user activity-point balance row in PostgreSQL.
type Currency struct {
	// UserID stores the balance owner identifier.
	UserID uint `gorm:"primaryKey;not null"`
	// CurrencyType stores the currency type identifier.
	CurrencyType int `gorm:"primaryKey;not null"`
	// Amount stores current balance value.
	Amount int `gorm:"not null;default:0"`
}

// TableName returns the PostgreSQL table name for Currency.
func (Currency) TableName() string { return "user_currencies" }

// CurrencyTransaction stores one trackable currency audit row in PostgreSQL.
type CurrencyTransaction struct {
	// ID stores stable transaction identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// UserID stores the transaction owner identifier.
	UserID uint `gorm:"not null;index:idx_currency_tx_user_type_created"`
	// CurrencyType stores the affected currency type.
	CurrencyType int `gorm:"not null;index:idx_currency_tx_user_type_created"`
	// Amount stores signed change value.
	Amount int `gorm:"not null"`
	// BalanceAfter stores resulting balance after the transaction.
	BalanceAfter int `gorm:"not null"`
	// Reason stores the transaction reason code.
	Reason string `gorm:"size:50;not null"`
	// ReferenceType stores the related entity type.
	ReferenceType string `gorm:"size:50;not null;default:''"`
	// ReferenceID stores the related entity identifier.
	ReferenceID uint `gorm:"not null;default:0"`
	// CreatedAt stores transaction timestamp.
	CreatedAt time.Time `gorm:"not null;index:idx_currency_tx_user_type_created"`
}

// TableName returns the PostgreSQL table name for CurrencyTransaction.
func (CurrencyTransaction) TableName() string { return "currency_transactions" }
