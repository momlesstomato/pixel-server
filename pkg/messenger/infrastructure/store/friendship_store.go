package store

import (
	"context"
	"errors"

	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
	messengermodel "github.com/momlesstomato/pixel-server/pkg/messenger/infrastructure/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ListFriendships returns all friendship rows for one user.
func (store *Store) ListFriendships(ctx context.Context, userID int) ([]domain.Friendship, error) {
	var rows []messengermodel.Friendship
	if err := store.database.WithContext(ctx).Where("user_one_id = ? OR user_two_id = ?", userID, userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Friendship, len(rows))
	for i, row := range rows {
		friendID := row.UserTwoID
		if row.UserOneID != userID {
			friendID = row.UserOneID
		}
		result[i] = domain.Friendship{
			UserOneID: userID, UserTwoID: friendID,
			Relationship: domain.RelationshipType(row.Relationship), CreatedAt: row.CreatedAt,
		}
	}
	return result, nil
}

// AreFriends reports whether two users share a friendship row.
func (store *Store) AreFriends(ctx context.Context, userOneID, userTwoID int) (bool, error) {
	left, right := canonicalPair(userOneID, userTwoID)
	var count int64
	err := store.database.WithContext(ctx).Model(&messengermodel.Friendship{}).
		Where("user_one_id = ? AND user_two_id = ?", left, right).Count(&count).Error
	return count > 0, err
}

// CountFriends returns the number of friends for one user.
func (store *Store) CountFriends(ctx context.Context, userID int) (int, error) {
	var count int64
	err := store.database.WithContext(ctx).Model(&messengermodel.Friendship{}).
		Where("user_one_id = ? OR user_two_id = ?", userID, userID).Count(&count).Error
	return int(count), err
}

// AddFriendship persists one canonical row for one friendship.
func (store *Store) AddFriendship(ctx context.Context, userOneID, userTwoID int) error {
	left, right := canonicalPair(userOneID, userTwoID)
	row := messengermodel.Friendship{UserOneID: left, UserTwoID: right}
	return store.database.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

// RemoveFriendship deletes one canonical row for one friendship.
func (store *Store) RemoveFriendship(ctx context.Context, userOneID, userTwoID int) error {
	left, right := canonicalPair(userOneID, userTwoID)
	return store.database.WithContext(ctx).Where("user_one_id = ? AND user_two_id = ?", left, right).Delete(&messengermodel.Friendship{}).Error
}

// SetRelationship updates the relationship type for one directional row.
func (store *Store) SetRelationship(ctx context.Context, userID, friendID int, rel domain.RelationshipType) error {
	left, right := canonicalPair(userID, friendID)
	return store.database.WithContext(ctx).Model(&messengermodel.Friendship{}).
		Where("user_one_id = ? AND user_two_id = ?", left, right).
		Update("relationship", int16(rel)).Error
}

// GetRelationship returns the relationship type for one directional row.
func (store *Store) GetRelationship(ctx context.Context, userID, friendID int) (domain.RelationshipType, error) {
	left, right := canonicalPair(userID, friendID)
	var row messengermodel.Friendship
	err := store.database.WithContext(ctx).Where("user_one_id = ? AND user_two_id = ?", left, right).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.RelationshipNone, domain.ErrNotFriends
	}
	return domain.RelationshipType(row.Relationship), err
}

// GetRelationshipCounts returns grouped relationship counts for one user profile.
func (store *Store) GetRelationshipCounts(ctx context.Context, userID int) ([]domain.RelationshipCount, error) {
	type row struct {
		Relationship int16
		Count        int
		SampleIDs    []int
	}
	var friendships []messengermodel.Friendship
	if err := store.database.WithContext(ctx).Where("(user_one_id = ? OR user_two_id = ?) AND relationship > 0", userID, userID).
		Find(&friendships).Error; err != nil {
		return nil, err
	}
	grouped := map[domain.RelationshipType]*domain.RelationshipCount{}
	for _, f := range friendships {
		t := domain.RelationshipType(f.Relationship)
		entry, ok := grouped[t]
		if !ok {
			entry = &domain.RelationshipCount{Type: t}
			grouped[t] = entry
		}
		entry.Count++
		if len(entry.SampleUserIDs) < 3 {
			sampleID := f.UserTwoID
			if f.UserOneID != userID {
				sampleID = f.UserOneID
			}
			entry.SampleUserIDs = append(entry.SampleUserIDs, sampleID)
		}
	}
	result := make([]domain.RelationshipCount, 0, len(grouped))
	for _, v := range grouped {
		result = append(result, *v)
	}
	return result, nil
}
