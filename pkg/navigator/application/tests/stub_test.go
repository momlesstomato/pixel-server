package tests

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
)

// repositoryStub defines deterministic navigator repository behavior.
type repositoryStub struct {
	// category stores deterministic category return.
	category domain.Category
	// room stores deterministic room return.
	room domain.Room
	// search stores deterministic saved search return.
	search domain.SavedSearch
	// favourites stores deterministic favourite count.
	favourites int
	// findErr stores deterministic find error.
	findErr error
	// deleteErr stores deterministic delete error.
	deleteErr error
}

// ListCategories returns deterministic category list.
func (s repositoryStub) ListCategories(_ context.Context) ([]domain.Category, error) {
	return []domain.Category{s.category}, nil
}

// FindCategoryByID returns deterministic category.
func (s repositoryStub) FindCategoryByID(_ context.Context, _ int) (domain.Category, error) {
	return s.category, s.findErr
}

// CreateCategory returns deterministic category.
func (s repositoryStub) CreateCategory(_ context.Context, c domain.Category) (domain.Category, error) {
	c.ID = 1
	return c, nil
}

// DeleteCategory returns deterministic error.
func (s repositoryStub) DeleteCategory(_ context.Context, _ int) error {
	return s.deleteErr
}

// ListRooms returns deterministic room list.
func (s repositoryStub) ListRooms(_ context.Context, _ domain.RoomFilter) ([]domain.Room, int, error) {
	return []domain.Room{s.room}, 1, nil
}

// FindRoomByID returns deterministic room.
func (s repositoryStub) FindRoomByID(_ context.Context, _ int) (domain.Room, error) {
	return s.room, s.findErr
}

// CreateRoom returns deterministic room.
func (s repositoryStub) CreateRoom(_ context.Context, r domain.Room) (domain.Room, error) {
	r.ID = 1
	return r, nil
}

// UpdateRoom returns deterministic room.
func (s repositoryStub) UpdateRoom(_ context.Context, _ int, _ domain.RoomPatch) (domain.Room, error) {
	return s.room, s.findErr
}

// DeleteRoom returns deterministic error.
func (s repositoryStub) DeleteRoom(_ context.Context, _ int) error {
	return s.deleteErr
}

// ListSavedSearches returns deterministic saved search list.
func (s repositoryStub) ListSavedSearches(_ context.Context, _ int) ([]domain.SavedSearch, error) {
	return []domain.SavedSearch{s.search}, nil
}

// CreateSavedSearch returns deterministic saved search.
func (s repositoryStub) CreateSavedSearch(_ context.Context, ss domain.SavedSearch) (domain.SavedSearch, error) {
	ss.ID = 1
	return ss, nil
}

// DeleteSavedSearch returns deterministic error.
func (s repositoryStub) DeleteSavedSearch(_ context.Context, _ int) error {
	return s.deleteErr
}

// ListFavourites returns deterministic favourite list.
func (s repositoryStub) ListFavourites(_ context.Context, _ int) ([]domain.Favourite, error) {
	return []domain.Favourite{{UserID: 1, RoomID: 1}}, nil
}

// AddFavourite returns nil.
func (s repositoryStub) AddFavourite(_ context.Context, _ int, _ int) error {
	return nil
}

// RemoveFavourite returns nil.
func (s repositoryStub) RemoveFavourite(_ context.Context, _ int, _ int) error {
	return nil
}

// CountFavourites returns deterministic count.
func (s repositoryStub) CountFavourites(_ context.Context, _ int) (int, error) {
	return s.favourites, nil
}
