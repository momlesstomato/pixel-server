package application

import (
	"context"
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
	sdknavigator "github.com/momlesstomato/pixel-sdk/events/navigator"
)

// ListSavedSearches resolves all saved searches for one user.
func (service *Service) ListSavedSearches(ctx context.Context, userID int) ([]domain.SavedSearch, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.ListSavedSearches(ctx, userID)
}

// CreateSavedSearch persists one saved search for one user.
func (service *Service) CreateSavedSearch(ctx context.Context, ss domain.SavedSearch) (domain.SavedSearch, error) {
	if ss.UserID <= 0 {
		return domain.SavedSearch{}, fmt.Errorf("user id must be positive")
	}
	if ss.SearchCode == "" {
		return domain.SavedSearch{}, fmt.Errorf("search code is required")
	}
	return service.repository.CreateSavedSearch(ctx, ss)
}

// DeleteSavedSearch removes one saved search by identifier.
func (service *Service) DeleteSavedSearch(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("search id must be positive")
	}
	return service.repository.DeleteSavedSearch(ctx, id)
}

// ListFavourites resolves all favourite room IDs for one user.
func (service *Service) ListFavourites(ctx context.Context, userID int) ([]domain.Favourite, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.ListFavourites(ctx, userID)
}

// AddFavourite creates one favourite entry, enforcing the per-user limit.
func (service *Service) AddFavourite(ctx context.Context, userID int, roomID int) error {
	if userID <= 0 || roomID <= 0 {
		return fmt.Errorf("user id and room id must be positive")
	}
	count, err := service.repository.CountFavourites(ctx, userID)
	if err != nil {
		return err
	}
	if count >= MaxFavourites {
		return domain.ErrFavouriteLimitReached
	}
	if service.fire != nil {
		ev := &sdknavigator.FavouriteAdding{UserID: userID, RoomID: roomID}
		service.fire(ev)
		if ev.Cancelled() {
			return fmt.Errorf("favourite addition cancelled by plugin")
		}
	}
	if err := service.repository.AddFavourite(ctx, userID, roomID); err != nil {
		return err
	}
	if service.fire != nil {
		service.fire(&sdknavigator.FavouriteAdded{UserID: userID, RoomID: roomID})
	}
	return nil
}

// RemoveFavourite deletes one favourite room entry.
func (service *Service) RemoveFavourite(ctx context.Context, userID int, roomID int) error {
	if userID <= 0 || roomID <= 0 {
		return fmt.Errorf("user id and room id must be positive")
	}
	return service.repository.RemoveFavourite(ctx, userID, roomID)
}
