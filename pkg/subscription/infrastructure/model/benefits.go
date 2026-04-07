package model

import "time"

// PaydayConfig stores the active HC payday configuration.
type PaydayConfig struct {
	// ID stores the singleton row identifier.
	ID uint `gorm:"primaryKey"`
	// IntervalDays stores the number of days between paydays.
	IntervalDays int `gorm:"not null;default:31"`
	// KickbackPercentage stores the percentage of cycle spend returned on payday.
	KickbackPercentage float64 `gorm:"not null;default:0"`
	// FlatCredits stores the fixed payday reward.
	FlatCredits int `gorm:"not null;default:0"`
	// MinimumCreditsSpent stores the minimum cycle spend required for kickback rewards.
	MinimumCreditsSpent int `gorm:"not null;default:0"`
	// StreakBonusCredits stores the extra reward for consecutive paydays.
	StreakBonusCredits int `gorm:"not null;default:0"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}

// TableName returns the PostgreSQL table name for PaydayConfig.
func (PaydayConfig) TableName() string { return "subscription_payday_config" }

// BenefitsState stores per-user HC payday and club-gift progress.
type BenefitsState struct {
	// UserID stores the owning user identifier.
	UserID uint `gorm:"primaryKey"`
	// FirstSubscriptionAt stores the first known HC activation timestamp.
	FirstSubscriptionAt time.Time `gorm:"not null"`
	// NextPaydayAt stores the next payday eligibility timestamp.
	NextPaydayAt time.Time `gorm:"not null"`
	// CycleCreditsSpent stores the current cycle credit spend.
	CycleCreditsSpent int `gorm:"not null;default:0"`
	// RewardStreak stores the count of consecutive processed paydays.
	RewardStreak int `gorm:"not null;default:0"`
	// TotalCreditsRewarded stores lifetime payday credits granted.
	TotalCreditsRewarded int `gorm:"not null;default:0"`
	// TotalCreditsMissed stores lifetime cycle spend that yielded no reward.
	TotalCreditsMissed int `gorm:"not null;default:0"`
	// ClubGiftsClaimed stores the number of claimed HC monthly gifts.
	ClubGiftsClaimed int `gorm:"not null;default:0"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}

// TableName returns the PostgreSQL table name for BenefitsState.
func (BenefitsState) TableName() string { return "subscription_benefits" }

// ClubGift stores one redeemable HC gift option.
type ClubGift struct {
	// ID stores the stable club gift identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// Name stores the client-visible gift name and selection key.
	Name string `gorm:"size:100;not null;uniqueIndex"`
	// ItemDefinitionID stores the delivered furniture definition identifier.
	ItemDefinitionID uint `gorm:"not null;index"`
	// ExtraData stores delivered item custom data.
	ExtraData string `gorm:"size:255;not null;default:''"`
	// DaysRequired stores the required HC age in days.
	DaysRequired int `gorm:"not null;default:31"`
	// VIPOnly stores whether the gift is VIP-only.
	VIPOnly bool `gorm:"not null;default:false"`
	// Enabled stores whether the gift is claimable.
	Enabled bool `gorm:"not null;default:true"`
	// OrderNum stores display ordering.
	OrderNum int `gorm:"not null;default:0"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}

// TableName returns the PostgreSQL table name for ClubGift.
func (ClubGift) TableName() string { return "subscription_club_gifts" }