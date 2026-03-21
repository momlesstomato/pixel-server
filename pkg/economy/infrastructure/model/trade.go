package model

import "time"

// TradeLog stores one completed trade audit row in PostgreSQL.
type TradeLog struct {
	// ID stores stable trade log identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// UserOneID stores the first trading user identifier.
	UserOneID uint `gorm:"not null;index"`
	// UserTwoID stores the second trading user identifier.
	UserTwoID uint `gorm:"not null;index"`
	// TradedAt stores the trade completion timestamp.
	TradedAt time.Time `gorm:"not null"`
}

// TableName returns the PostgreSQL table name for TradeLog.
func (TradeLog) TableName() string { return "trade_logs" }

// TradeLogItem stores one item exchanged in a trade in PostgreSQL.
type TradeLogItem struct {
	// ID stores stable trade log item identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// TradeID stores the owning trade log identifier.
	TradeID uint `gorm:"not null;index"`
	// ItemID stores the traded item instance identifier.
	ItemID uint `gorm:"not null"`
	// UserID stores the item source owner identifier.
	UserID uint `gorm:"not null"`
	// DefinitionID stores the item definition identifier.
	DefinitionID uint `gorm:"not null"`
}

// TableName returns the PostgreSQL table name for TradeLogItem.
func (TradeLogItem) TableName() string { return "trade_log_items" }
