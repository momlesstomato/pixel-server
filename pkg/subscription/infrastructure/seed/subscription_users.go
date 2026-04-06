package seed

import (
	"time"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	submodel "github.com/momlesstomato/pixel-server/pkg/subscription/infrastructure/model"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// subscribedUsernames lists demo and test user accounts that should have active HC.
var subscribedUsernames = []string{
	"test_vip", "test_admin",
	"demo_vip", "demo_admin",
}

// Step02SubscriptionUsers returns seed step that gives demo VIP and admin users an active HC subscription.
func Step02SubscriptionUsers() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260405_S03_subscription_users",
		Migrate: func(database *gorm.DB) error {
			return ensureSubscriptionUsers(database)
		},
		Rollback: func(database *gorm.DB) error {
			return rollbackSubscriptionUsers(database)
		},
	}
}

// ensureSubscriptionUsers creates a one-year HC subscription for each listed user when missing.
func ensureSubscriptionUsers(database *gorm.DB) error {
	var users []usermodel.Record
	if err := database.Where("username IN ?", subscribedUsernames).Find(&users).Error; err != nil {
		return err
	}
	now := time.Now()
	subs := make([]submodel.Subscription, 0, len(users))
	for _, u := range users {
		var count int64
		if err := database.Model(&submodel.Subscription{}).
			Where("user_id = ? AND subscription_type = ?", u.ID, "habbo_club").
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		subs = append(subs, submodel.Subscription{
			UserID: u.ID, SubscriptionType: "habbo_club",
			StartedAt: now, DurationDays: 365, Active: true,
		})
	}
	if len(subs) == 0 {
		return nil
	}
	return database.Clauses(clause.OnConflict{DoNothing: true}).Create(&subs).Error
}

// rollbackSubscriptionUsers removes the seeded HC subscriptions for listed users.
func rollbackSubscriptionUsers(database *gorm.DB) error {
	var users []usermodel.Record
	if err := database.Where("username IN ?", subscribedUsernames).Find(&users).Error; err != nil {
		return err
	}
	ids := make([]uint, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	return database.Where("user_id IN ? AND subscription_type = ?", ids, "habbo_club").
		Delete(&submodel.Subscription{}).Error
}
