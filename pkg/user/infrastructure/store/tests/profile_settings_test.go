package tests

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	userstore "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestRepositoryUpdateProfileAndSettings verifies profile and settings persistence behavior.
func TestRepositoryUpdateProfileAndSettings(t *testing.T) {
	repository := openRepository(t)
	created, err := repository.Create(context.Background(), "alpha")
	if err != nil {
		t.Fatalf("expected user create success, got %v", err)
	}
	figure := "hd-190-5"
	gender := "f"
	motto := "hello"
	homeRoom := 123
	updated, err := repository.UpdateProfile(context.Background(), created.ID, domain.ProfilePatch{
		Figure: &figure, Gender: &gender, Motto: &motto, HomeRoomID: &homeRoom,
	})
	if err != nil {
		t.Fatalf("expected profile update success, got %v", err)
	}
	if updated.Figure != figure || updated.Gender != "F" || updated.Motto != motto || updated.HomeRoomID != homeRoom {
		t.Fatalf("unexpected updated profile payload %+v", updated)
	}
	settings, err := repository.LoadSettings(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected settings load success, got %v", err)
	}
	if settings.VolumeSystem != 100 || !settings.RoomInvites || !settings.CameraFollow {
		t.Fatalf("unexpected default settings payload %+v", settings)
	}
	volume := 15
	oldChat := true
	saved, err := repository.SaveSettings(context.Background(), created.ID, domain.SettingsPatch{VolumeSystem: &volume, OldChat: &oldChat})
	if err != nil {
		t.Fatalf("expected settings save success, got %v", err)
	}
	if saved.VolumeSystem != 15 || !saved.OldChat {
		t.Fatalf("unexpected saved settings payload %+v", saved)
	}
}

// TestRepositoryUpdateProfileNotFound verifies not-found update behavior.
func TestRepositoryUpdateProfileNotFound(t *testing.T) {
	repository := openRepository(t)
	motto := "x"
	if _, err := repository.UpdateProfile(context.Background(), 99, domain.ProfilePatch{Motto: &motto}); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected user not found, got %v", err)
	}
}

// openRepository creates sqlite repository with migrated user schemas.
func openRepository(t *testing.T) *userstore.Repository {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&usermodel.Record{}, &usermodel.Settings{}, &usermodel.Respect{}, &usermodel.LoginEvent{}, &usermodel.WardrobeSlot{}, &usermodel.Ignore{}); err != nil {
		t.Fatalf("expected sqlite migration success, got %v", err)
	}
	repository, err := userstore.NewRepository(database)
	if err != nil {
		t.Fatalf("expected repository creation success, got %v", err)
	}
	return repository
}
