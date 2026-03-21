package model

import "time"

// MarketplaceOffer stores one marketplace listing row in PostgreSQL.
type MarketplaceOffer struct {
	// ID stores stable marketplace offer identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// SellerID stores the listing owner identifier.
	SellerID uint `gorm:"not null;index:idx_mp_offers_seller"`
	// ItemID stores the offered item identifier.
	ItemID uint `gorm:"not null"`
	// DefinitionID stores the item definition reference.
	DefinitionID uint `gorm:"not null;index:idx_mp_offers_definition"`
	// AskingPrice stores the seller requested price.
	AskingPrice int `gorm:"not null"`
	// State stores the current offer lifecycle state.
	State string `gorm:"size:20;not null;default:open;index:idx_mp_offers_state"`
	// BuyerID stores the purchaser identifier when sold.
	BuyerID *uint `gorm:"index:idx_mp_offers_buyer"`
	// SoldAt stores the sale completion timestamp.
	SoldAt *time.Time
	// ExpireAt stores the listing expiration timestamp.
	ExpireAt time.Time `gorm:"not null;index:idx_mp_offers_expire"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
}

// TableName returns the PostgreSQL table name for MarketplaceOffer.
func (MarketplaceOffer) TableName() string { return "marketplace_offers" }

// PriceHistory stores one aggregated price history row in PostgreSQL.
type PriceHistory struct {
	// ID stores stable price history entry identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// SpriteID stores the item sprite being tracked.
	SpriteID int `gorm:"not null;index:idx_price_history_sprite"`
	// DayOffset stores how many days ago this entry represents.
	DayOffset int `gorm:"not null"`
	// AvgPrice stores average sale price for that day.
	AvgPrice int `gorm:"not null"`
	// SoldCount stores total units sold for that day.
	SoldCount int `gorm:"not null"`
	// RecordedAt stores when this aggregate was computed.
	RecordedAt time.Time `gorm:"not null"`
}

// TableName returns the PostgreSQL table name for PriceHistory.
func (PriceHistory) TableName() string { return "marketplace_price_history" }
