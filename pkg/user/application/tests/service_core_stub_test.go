package tests

import (
	"context"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// coreRepositoryStub defines deterministic repository behavior for core service tests.
type coreRepositoryStub struct {
	saved      domain.User
	findErr    error
	deleteErr  error
	firstLogin bool
	loginErr   error
}

// Create returns deterministic user payload.
func (stub coreRepositoryStub) Create(_ context.Context, _ string) (domain.User, error) {
	return stub.saved, nil
}

// FindByID returns deterministic find result.
func (stub coreRepositoryStub) FindByID(_ context.Context, _ int) (domain.User, error) {
	if stub.findErr != nil {
		return domain.User{}, stub.findErr
	}
	return stub.saved, nil
}

// DeleteByID returns deterministic delete result.
func (stub coreRepositoryStub) DeleteByID(_ context.Context, _ int) error { return stub.deleteErr }

// UpdateProfile returns deterministic profile update payload.
func (stub coreRepositoryStub) UpdateProfile(_ context.Context, _ int, patch domain.ProfilePatch) (domain.User, error) {
	updated := stub.saved
	if patch.Motto != nil {
		updated.Motto = *patch.Motto
	}
	if patch.Figure != nil {
		updated.Figure = *patch.Figure
	}
	if patch.Gender != nil {
		updated.Gender = *patch.Gender
	}
	return updated, nil
}

// LoadSettings returns deterministic settings payload.
func (stub coreRepositoryStub) LoadSettings(_ context.Context, userID int) (domain.Settings, error) {
	return domain.Settings{UserID: userID}, nil
}

// SaveSettings returns deterministic settings payload.
func (stub coreRepositoryStub) SaveSettings(_ context.Context, userID int, _ domain.SettingsPatch) (domain.Settings, error) {
	return domain.Settings{UserID: userID}, nil
}

// RecordRespect returns deterministic respect value.
func (stub coreRepositoryStub) RecordRespect(_ context.Context, _ int, _ int, _ domain.RespectTargetType, _ time.Time) (int, error) {
	return 1, nil
}

// RemainingRespects returns deterministic remaining respects value.
func (stub coreRepositoryStub) RemainingRespects(_ context.Context, _ int, _ domain.RespectTargetType, _ time.Time) (int, error) {
	return 2, nil
}

// RecordLogin returns deterministic login marker.
func (stub coreRepositoryStub) RecordLogin(_ context.Context, _ int, _ string, _ time.Time) (bool, error) {
	if stub.loginErr != nil {
		return false, stub.loginErr
	}
	return stub.firstLogin, nil
}

// LoadWardrobe returns deterministic wardrobe payload.
func (stub coreRepositoryStub) LoadWardrobe(_ context.Context, _ int) ([]domain.WardrobeSlot, error) {
	return []domain.WardrobeSlot{}, nil
}

// SaveWardrobeSlot returns deterministic wardrobe save result.
func (stub coreRepositoryStub) SaveWardrobeSlot(_ context.Context, _ int, _ domain.WardrobeSlot) error {
	return nil
}

// ListIgnoredUsernames returns deterministic ignored usernames payload.
func (stub coreRepositoryStub) ListIgnoredUsernames(_ context.Context, _ int) ([]string, error) {
	return []string{}, nil
}

// ListIgnoredUsers returns deterministic ignored user entries.
func (stub coreRepositoryStub) ListIgnoredUsers(_ context.Context, _ int) ([]domain.IgnoreEntry, error) {
	return []domain.IgnoreEntry{}, nil
}

// IgnoreUserByUsername returns deterministic ignored user identifier.
func (stub coreRepositoryStub) IgnoreUserByUsername(_ context.Context, _ int, _ string) (int, error) {
	return 2, nil
}

// IgnoreUserByID returns deterministic ignore result.
func (stub coreRepositoryStub) IgnoreUserByID(_ context.Context, _ int, _ int) error { return nil }

// UnignoreUserByUsername returns deterministic unignore result.
func (stub coreRepositoryStub) UnignoreUserByUsername(_ context.Context, _ int, _ string) (int, error) {
	return 2, nil
}

// UnignoreUserByID returns deterministic unignore result.
func (stub coreRepositoryStub) UnignoreUserByID(_ context.Context, _ int, _ int) error {
	return nil
}

// FindByUsername returns deterministic user payload by username.
func (stub coreRepositoryStub) FindByUsername(_ context.Context, _ string) (domain.User, error) {
	return stub.saved, nil
}

// LoadProfile returns deterministic profile payload.
func (stub coreRepositoryStub) LoadProfile(_ context.Context, _ int, userID int, openProfileWindow bool) (domain.Profile, error) {
	return domain.Profile{UserID: userID, OpenProfileWindow: openProfileWindow}, nil
}

// ListRespects returns deterministic respect records.
func (stub coreRepositoryStub) ListRespects(_ context.Context, _ int, _ int, _ int) ([]domain.RespectRecord, error) {
	return []domain.RespectRecord{}, nil
}

// IsUsernameAvailable returns deterministic availability marker.
func (stub coreRepositoryStub) IsUsernameAvailable(_ context.Context, _ string, _ int) (bool, error) {
	return true, nil
}

// ChangeUsername returns deterministic user payload with changed username.
func (stub coreRepositoryStub) ChangeUsername(_ context.Context, _ int, username string, _ bool) (domain.User, error) {
	updated := stub.saved
	updated.Username = username
	return updated, nil
}
