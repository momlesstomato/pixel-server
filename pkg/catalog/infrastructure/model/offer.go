package model

import "time"

// Offer stores one catalog offer row in PostgreSQL.
type Offer struct {
	// ID stores stable offer identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// PageID stores the owning page identifier.
	PageID uint `gorm:"not null;index"`
	// ItemDefinitionID stores the furniture definition foreign key.
	ItemDefinitionID uint `gorm:"not null;index"`
	// CostCredits stores the credits price component.
	CostCredits int `gorm:"column:cost_credits;not null;default:0"`
	// CostActivityPoints stores the activity-point price component.
	CostActivityPoints int `gorm:"column:cost_activity_points;not null;default:0"`
	// ActivityPointType stores the activity-point currency type identifier.
	ActivityPointType int `gorm:"column:activity_point_type;not null;default:0"`
	// Amount stores number of items per single purchase.
	Amount int `gorm:"not null;default:1"`
	// LimitedTotal stores total limited edition print run.
	LimitedTotal int `gorm:"not null;default:0"`
	// LimitedSells stores current limited edition sold count.
	LimitedSells int `gorm:"not null;default:0"`
	// OfferActive stores whether the offer is currently purchasable.
	OfferActive bool `gorm:"not null;default:true"`
	// ExtraData stores item-specific custom data payload.
	ExtraData string `gorm:"size:255;not null;default:''"`
	// BadgeID stores optional badge code awarded with purchase.
	BadgeID string `gorm:"size:10;not null;default:''"`
	// ClubOnly stores whether club membership is required.
	ClubOnly bool `gorm:"not null;default:false"`
	// OrderNum stores display sort position.
	OrderNum int `gorm:"not null;default:0"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}

// TableName returns the PostgreSQL table name for Offer.
func (Offer) TableName() string { return "catalog_items" }
