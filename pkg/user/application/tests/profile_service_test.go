package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// TestServiceUpdateProfileAndSettingsValidations verifies profile and settings validation behavior.
func TestServiceUpdateProfileAndSettingsValidations(t *testing.T) {
	service, _ := userapplication.NewService(repositoryStub{})
	if _, err := service.UpdateProfile(context.Background(), 0, domain.ProfilePatch{}); err == nil {
		t.Fatalf("expected invalid user id error")
	}
	gender := "x"
	if _, err := service.UpdateProfile(context.Background(), 1, domain.ProfilePatch{Gender: &gender}); err == nil {
		t.Fatalf("expected invalid gender error")
	}
	volume := 200
	if _, err := service.SaveSettings(context.Background(), 1, domain.SettingsPatch{VolumeSystem: &volume}); err == nil {
		t.Fatalf("expected invalid volume error")
	}
}

// TestServiceRecordUserRespectValidatesAndReturnsResult verifies respect operation behavior.
func TestServiceRecordUserRespectValidatesAndReturnsResult(t *testing.T) {
	service, _ := userapplication.NewService(repositoryStub{recordRespectValue: 4, remainingValue: 2})
	if _, err := service.RecordUserRespect(context.Background(), 0, 1, time.Now()); err == nil {
		t.Fatalf("expected invalid actor id error")
	}
	if _, err := service.RecordUserRespect(context.Background(), 1, 1, time.Now()); err == nil {
		t.Fatalf("expected self respect error")
	}
	result, err := service.RecordUserRespect(context.Background(), 1, 2, time.Now())
	if err != nil {
		t.Fatalf("expected respect success, got %v", err)
	}
	if result.RespectsReceived != 4 || result.Remaining != 2 {
		t.Fatalf("unexpected respect result %+v", result)
	}
}

// repositoryStub defines deterministic repository behavior for profile service tests.
type repositoryStub struct {
	// recordRespectValue stores deterministic respects received counter.
	recordRespectValue int
	// remainingValue stores deterministic remaining respects value.
	remainingValue int
}

// Create returns deterministic user payload.
func (stub repositoryStub) Create(context.Context, string) (domain.User, error) {
	return domain.User{}, nil
}

// FindByID returns deterministic user payload.
func (stub repositoryStub) FindByID(context.Context, int) (domain.User, error) {
	return domain.User{ID: 1}, nil
}

// DeleteByID returns deterministic delete result.
func (stub repositoryStub) DeleteByID(context.Context, int) error { return nil }

// UpdateProfile returns deterministic profile payload.
func (stub repositoryStub) UpdateProfile(context.Context, int, domain.ProfilePatch) (domain.User, error) {
	return domain.User{ID: 1, Gender: "M"}, nil
}

// LoadSettings returns deterministic settings payload.
func (stub repositoryStub) LoadSettings(context.Context, int) (domain.Settings, error) {
	return domain.Settings{UserID: 1, VolumeSystem: 100}, nil
}

// SaveSettings returns deterministic settings payload.
func (stub repositoryStub) SaveSettings(context.Context, int, domain.SettingsPatch) (domain.Settings, error) {
	return domain.Settings{UserID: 1, VolumeSystem: 10}, nil
}

// RecordRespect returns deterministic respect result.
func (stub repositoryStub) RecordRespect(context.Context, int, int, domain.RespectTargetType, time.Time) (int, error) {
	if stub.recordRespectValue == 0 {
		return 0, errors.New("record respect error")
	}
	return stub.recordRespectValue, nil
}

// RemainingRespects returns deterministic remaining respects value.
func (stub repositoryStub) RemainingRespects(context.Context, int, domain.RespectTargetType, time.Time) (int, error) {
	return stub.remainingValue, nil
}

// RecordLogin returns deterministic login event output.
func (stub repositoryStub) RecordLogin(context.Context, int, string, time.Time) (bool, error) {
	return false, nil
}

// LoadWardrobe returns deterministic wardrobe list.
func (stub repositoryStub) LoadWardrobe(context.Context, int) ([]domain.WardrobeSlot, error) {
	return []domain.WardrobeSlot{}, nil
}

// SaveWardrobeSlot returns deterministic save result.
func (stub repositoryStub) SaveWardrobeSlot(context.Context, int, domain.WardrobeSlot) error {
	return nil
}

// ListIgnoredUsernames returns deterministic ignore list.
func (stub repositoryStub) ListIgnoredUsernames(context.Context, int) ([]string, error) {
	return []string{}, nil
}

// IgnoreUserByUsername returns deterministic ignored user identifier.
func (stub repositoryStub) IgnoreUserByUsername(context.Context, int, string) (int, error) {
	return 2, nil
}

// IgnoreUserByID returns deterministic ignore result.
func (stub repositoryStub) IgnoreUserByID(context.Context, int, int) error {
	return nil
}

// UnignoreUserByUsername returns deterministic unignore identifier.
func (stub repositoryStub) UnignoreUserByUsername(context.Context, int, string) (int, error) {
	return 2, nil
}

// LoadProfile returns deterministic profile payload.
func (stub repositoryStub) LoadProfile(context.Context, int, bool) (domain.Profile, error) {
	return domain.Profile{UserID: 1}, nil
}

// ListRespects returns deterministic respect list payload.
func (stub repositoryStub) ListRespects(context.Context, int, int, int) ([]domain.RespectRecord, error) {
	return []domain.RespectRecord{}, nil
}

// IsUsernameAvailable returns deterministic username availability.
func (stub repositoryStub) IsUsernameAvailable(context.Context, string, int) (bool, error) {
	return true, nil
}

// ChangeUsername returns deterministic changed user payload.
func (stub repositoryStub) ChangeUsername(_ context.Context, _ int, username string, _ bool) (domain.User, error) {
	return domain.User{ID: 1, Username: username}, nil
}
