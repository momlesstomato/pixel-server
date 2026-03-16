package store

import (
	"context"
	"errors"
	"strings"

	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
	messengermodel "github.com/momlesstomato/pixel-server/pkg/messenger/infrastructure/model"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// CreateRequest persists one friend request row.
func (store *Store) CreateRequest(ctx context.Context, fromUserID, toUserID int) (domain.FriendRequest, error) {
	row := messengermodel.Request{FromUserID: fromUserID, ToUserID: toUserID}
	if err := store.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.FriendRequest{}, err
	}
	return domain.FriendRequest{ID: row.ID, FromUserID: row.FromUserID, ToUserID: row.ToUserID, CreatedAt: row.CreatedAt}, nil
}

// FindRequest returns one friend request by identifier.
func (store *Store) FindRequest(ctx context.Context, requestID int) (domain.FriendRequest, error) {
	var row messengermodel.Request
	if err := store.database.WithContext(ctx).First(&row, requestID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.FriendRequest{}, domain.ErrRequestNotFound
	} else if err != nil {
		return domain.FriendRequest{}, err
	}
	return domain.FriendRequest{ID: row.ID, FromUserID: row.FromUserID, ToUserID: row.ToUserID, CreatedAt: row.CreatedAt}, nil
}

// FindRequestByUsers returns a request row between two users if it exists.
func (store *Store) FindRequestByUsers(ctx context.Context, fromUserID, toUserID int) (domain.FriendRequest, bool, error) {
	var row messengermodel.Request
	query := store.database.WithContext(ctx).Where("from_user_id = ? AND to_user_id = ?", fromUserID, toUserID).Limit(1).Find(&row)
	if query.Error != nil {
		return domain.FriendRequest{}, false, query.Error
	}
	if query.RowsAffected == 0 {
		return domain.FriendRequest{}, false, nil
	}
	return domain.FriendRequest{ID: row.ID, FromUserID: row.FromUserID, ToUserID: row.ToUserID, CreatedAt: row.CreatedAt}, true, nil
}

// ListRequests returns all pending requests addressed to one user.
func (store *Store) ListRequests(ctx context.Context, toUserID int) ([]domain.FriendRequest, error) {
	var rows []messengermodel.Request
	if err := store.database.WithContext(ctx).Where("to_user_id = ?", toUserID).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.FriendRequest, len(rows))
	for i, row := range rows {
		result[i] = domain.FriendRequest{ID: row.ID, FromUserID: row.FromUserID, ToUserID: row.ToUserID, CreatedAt: row.CreatedAt}
	}
	return result, nil
}

// DeleteRequest removes one request row by identifier.
func (store *Store) DeleteRequest(ctx context.Context, requestID int) error {
	return store.database.WithContext(ctx).Delete(&messengermodel.Request{}, requestID).Error
}

// DeleteAllRequests removes all pending requests addressed to one user.
func (store *Store) DeleteAllRequests(ctx context.Context, toUserID int) error {
	return store.database.WithContext(ctx).Where("to_user_id = ?", toUserID).Delete(&messengermodel.Request{}).Error
}

// FindUserIDByUsername returns the identifier for one username if it exists.
func (store *Store) FindUserIDByUsername(ctx context.Context, username string) (int, bool, error) {
	var record usermodel.Record
	err := store.database.WithContext(ctx).Where("username = ?", username).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return int(record.ID), true, nil
}

// SearchUsers returns users whose username contains the query string.
func (store *Store) SearchUsers(ctx context.Context, query string, limit int) ([]domain.SearchResult, error) {
	var records []usermodel.Record
	like := "%" + strings.ToLower(query) + "%"
	if err := store.database.WithContext(ctx).Where("LOWER(username) LIKE ?", like).Limit(limit).Find(&records).Error; err != nil {
		return nil, err
	}
	result := make([]domain.SearchResult, len(records))
	for i, r := range records {
		result[i] = domain.SearchResult{
			ID: int(r.ID), Username: r.Username, Figure: r.Figure,
			Gender: r.Gender, Motto: r.Motto,
		}
	}
	return result, nil
}

// FindUsersByIDs returns profile records for a set of user identifiers.
func (store *Store) FindUsersByIDs(ctx context.Context, ids []int) ([]domain.SearchResult, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var records []usermodel.Record
	if err := store.database.WithContext(ctx).Where("id IN (?)", ids).Find(&records).Error; err != nil {
		return nil, err
	}
	result := make([]domain.SearchResult, len(records))
	for i, r := range records {
		result[i] = domain.SearchResult{
			ID: int(r.ID), Username: r.Username, Figure: r.Figure,
			Gender: r.Gender, Motto: r.Motto,
		}
	}
	return result, nil
}
