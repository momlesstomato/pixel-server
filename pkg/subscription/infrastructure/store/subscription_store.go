package store

import (
	"context"
	"errors"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
	submodel "github.com/momlesstomato/pixel-server/pkg/subscription/infrastructure/model"
	"gorm.io/gorm"
)

// FindActiveSubscription resolves active subscription for one user.
func (store *Store) FindActiveSubscription(ctx context.Context, userID int) (domain.Subscription, error) {
	var row submodel.Subscription
	err := store.database.WithContext(ctx).Where("user_id = ? AND active = true", userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Subscription{}, domain.ErrSubscriptionNotFound
	}
	if err != nil {
		return domain.Subscription{}, err
	}
	return mapSubscription(row), nil
}

// CreateSubscription persists one subscription row.
func (store *Store) CreateSubscription(ctx context.Context, sub domain.Subscription) (domain.Subscription, error) {
	row := submodel.Subscription{
		UserID: uint(sub.UserID), SubscriptionType: string(sub.SubscriptionType),
		StartedAt: time.Now().UTC(), DurationDays: sub.DurationDays, Active: true,
	}
	if err := store.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Subscription{}, err
	}
	return mapSubscription(row), nil
}

// ExtendSubscription adds days to an existing active subscription.
func (store *Store) ExtendSubscription(ctx context.Context, userID int, days int) (domain.Subscription, error) {
	result := store.database.WithContext(ctx).Model(&submodel.Subscription{}).
		Where("user_id = ? AND active = true", userID).
		Update("duration_days", gorm.Expr("duration_days + ?", days))
	if result.Error != nil {
		return domain.Subscription{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Subscription{}, domain.ErrSubscriptionNotFound
	}
	return store.FindActiveSubscription(ctx, userID)
}

// DeactivateSubscription marks one subscription as inactive.
func (store *Store) DeactivateSubscription(ctx context.Context, subscriptionID int) error {
	result := store.database.WithContext(ctx).Model(&submodel.Subscription{}).
		Where("id = ?", subscriptionID).Update("active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrSubscriptionNotFound
	}
	return nil
}

// FindExpiredActive resolves all active but elapsed subscriptions.
func (store *Store) FindExpiredActive(ctx context.Context) ([]domain.Subscription, error) {
	var rows []submodel.Subscription
	err := store.database.WithContext(ctx).
		Where("active = true AND started_at + make_interval(days => duration_days) < NOW()").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]domain.Subscription, len(rows))
	for i, row := range rows {
		result[i] = mapSubscription(row)
	}
	return result, nil
}

// ListClubOffers resolves all enabled club membership offers.
func (store *Store) ListClubOffers(ctx context.Context) ([]domain.ClubOffer, error) {
	var rows []submodel.ClubOffer
	if err := store.database.WithContext(ctx).Where("enabled = true").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.ClubOffer, len(rows))
	for i, row := range rows {
		result[i] = mapClubOffer(row)
	}
	return result, nil
}

// FindClubOfferByID resolves one club offer by identifier.
func (store *Store) FindClubOfferByID(ctx context.Context, id int) (domain.ClubOffer, error) {
	var row submodel.ClubOffer
	if err := store.database.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ClubOffer{}, domain.ErrClubOfferNotFound
		}
		return domain.ClubOffer{}, err
	}
	return mapClubOffer(row), nil
}

// CreateClubOffer persists one club offer row.
func (store *Store) CreateClubOffer(ctx context.Context, offer domain.ClubOffer) (domain.ClubOffer, error) {
	row := submodel.ClubOffer{
		Name: offer.Name, Days: offer.Days, Credits: offer.Credits,
		Points: offer.Points, PointsType: offer.PointsType,
		OfferType: offer.OfferType, Giftable: offer.Giftable, Enabled: offer.Enabled,
	}
	if err := store.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.ClubOffer{}, err
	}
	return mapClubOffer(row), nil
}

// DeleteClubOffer removes one club offer by identifier.
func (store *Store) DeleteClubOffer(ctx context.Context, id int) error {
	result := store.database.WithContext(ctx).Delete(&submodel.ClubOffer{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrClubOfferNotFound
	}
	return nil
}

// mapSubscription converts one GORM model into domain subscription.
func mapSubscription(row submodel.Subscription) domain.Subscription {
	return domain.Subscription{
		ID: int(row.ID), UserID: int(row.UserID),
		SubscriptionType: domain.SubscriptionType(row.SubscriptionType),
		StartedAt: row.StartedAt, DurationDays: row.DurationDays,
		Active: row.Active, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

// mapClubOffer converts one GORM model into domain club offer.
func mapClubOffer(row submodel.ClubOffer) domain.ClubOffer {
	return domain.ClubOffer{
		ID: int(row.ID), Name: row.Name, Days: row.Days,
		Credits: row.Credits, Points: row.Points,
		PointsType: row.PointsType, OfferType: row.OfferType,
		Giftable: row.Giftable, Enabled: row.Enabled,
	}
}
