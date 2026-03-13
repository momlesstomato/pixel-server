package application

import (
	"context"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// repositoryStub defines deterministic repository behavior for tests.
type repositoryStub struct {
	// saved stores deterministic user payload.
	saved domain.User
	// findErr stores deterministic find failure.
	findErr error
	// deleteErr stores deterministic delete failure.
	deleteErr error
	// firstLogin stores deterministic login stamp output.
	firstLogin bool
	// loginErr stores deterministic login stamp failure.
	loginErr error
}

// Create returns deterministic user payload.
func (stub repositoryStub) Create(_ context.Context, _ string) (domain.User, error) {
	return stub.saved, nil
}

// FindByID returns deterministic find result.
func (stub repositoryStub) FindByID(_ context.Context, _ int) (domain.User, error) {
	if stub.findErr != nil {
		return domain.User{}, stub.findErr
	}
	return stub.saved, nil
}

// DeleteByID returns deterministic delete result.
func (stub repositoryStub) DeleteByID(_ context.Context, _ int) error { return stub.deleteErr }

// RecordLogin returns deterministic login stamp output.
func (stub repositoryStub) RecordLogin(_ context.Context, _ int, _ string, _ time.Time) (bool, error) {
	if stub.loginErr != nil {
		return false, stub.loginErr
	}
	return stub.firstLogin, nil
}

// UpdateProfile returns deterministic user payload.
func (stub repositoryStub) UpdateProfile(_ context.Context, _ int, _ domain.ProfilePatch) (domain.User, error) {
	return stub.saved, nil
}

// LoadSettings returns deterministic settings payload.
func (stub repositoryStub) LoadSettings(_ context.Context, userID int) (domain.Settings, error) {
	return domain.Settings{UserID: userID}, nil
}

// SaveSettings returns deterministic settings payload.
func (stub repositoryStub) SaveSettings(_ context.Context, userID int, _ domain.SettingsPatch) (domain.Settings, error) {
	return domain.Settings{UserID: userID}, nil
}

// RecordRespect returns deterministic respect total.
func (stub repositoryStub) RecordRespect(_ context.Context, _ int, _ int, _ domain.RespectTargetType, _ time.Time) (int, error) {
	return 1, nil
}

// RemainingRespects returns deterministic remaining respects count.
func (stub repositoryStub) RemainingRespects(_ context.Context, _ int, _ domain.RespectTargetType, _ time.Time) (int, error) {
	return 2, nil
}

// longUsername returns a username bigger than allowed max length.
func longUsername() string {
	value := ""
	for index := 0; index < 65; index++ {
		value += "a"
	}
	return value
}
