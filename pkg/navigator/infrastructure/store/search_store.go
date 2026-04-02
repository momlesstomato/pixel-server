package store

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
	model "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/model"
)

// ListSavedSearches resolves all saved searches for one user.
func (s *Store) ListSavedSearches(ctx context.Context, userID int) ([]domain.SavedSearch, error) {
	var rows []model.SavedSearch
	if err := s.database.WithContext(ctx).Where("user_id = ?", userID).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.SavedSearch, len(rows))
	for i, row := range rows {
		result[i] = mapSavedSearch(row)
	}
	return result, nil
}

// CreateSavedSearch persists one saved search row.
func (s *Store) CreateSavedSearch(ctx context.Context, search domain.SavedSearch) (domain.SavedSearch, error) {
	row := model.SavedSearch{
		UserID: uint(search.UserID), SearchCode: search.SearchCode, Filter: search.Filter,
	}
	if err := s.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.SavedSearch{}, err
	}
	return mapSavedSearch(row), nil
}

// DeleteSavedSearch removes one saved search by identifier.
func (s *Store) DeleteSavedSearch(ctx context.Context, id int) error {
	result := s.database.WithContext(ctx).Delete(&model.SavedSearch{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrSearchNotFound
	}
	return nil
}

// ListFavourites resolves all favourite entries for one user.
func (s *Store) ListFavourites(ctx context.Context, userID int) ([]domain.Favourite, error) {
	var rows []model.Favourite
	if err := s.database.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Favourite, len(rows))
	for i, row := range rows {
		result[i] = domain.Favourite{UserID: int(row.UserID), RoomID: int(row.RoomID), CreatedAt: row.CreatedAt}
	}
	return result, nil
}

// AddFavourite creates one favourite entry for a user-room pair.
func (s *Store) AddFavourite(ctx context.Context, userID int, roomID int) error {
	row := model.Favourite{UserID: uint(userID), RoomID: uint(roomID)}
	if err := s.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.ErrFavouriteAlreadyExists
	}
	return nil
}

// RemoveFavourite deletes one favourite entry by user-room pair.
func (s *Store) RemoveFavourite(ctx context.Context, userID int, roomID int) error {
	result := s.database.WithContext(ctx).Where("user_id = ? AND room_id = ?", userID, roomID).Delete(&model.Favourite{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrFavouriteNotFound
	}
	return nil
}

// CountFavourites returns favourite count for one user.
func (s *Store) CountFavourites(ctx context.Context, userID int) (int, error) {
	var count int64
	if err := s.database.WithContext(ctx).Model(&model.Favourite{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// mapSavedSearch converts persistence model to domain type.
func mapSavedSearch(row model.SavedSearch) domain.SavedSearch {
	return domain.SavedSearch{
		ID: int(row.ID), UserID: int(row.UserID),
		SearchCode: row.SearchCode, Filter: row.Filter, CreatedAt: row.CreatedAt,
	}
}
