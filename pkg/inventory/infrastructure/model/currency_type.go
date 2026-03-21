package model

import "time"

// CurrencyType stores one registered currency type row in PostgreSQL.
type CurrencyType struct {
	// ID stores operator-assigned integer identifier.
	ID int `gorm:"primaryKey;not null"`
	// Name stores unique internal name for this currency.
	Name string `gorm:"size:50;uniqueIndex;not null"`
	// DisplayName stores player-visible currency label.
	DisplayName string `gorm:"size:100;not null;default:''"`
	// Trackable stores whether balance changes are recorded in currency_transactions.
	Trackable bool `gorm:"not null;default:false"`
	// Enabled stores whether this currency is active.
	Enabled bool `gorm:"not null;default:true"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
}

// TableName returns the PostgreSQL table name for CurrencyType.
func (CurrencyType) TableName() string { return "currency_types" }
