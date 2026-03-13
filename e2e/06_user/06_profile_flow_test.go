package user

import (
	"context"
	"errors"
	"testing"
	"time"

	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	userstore "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Test06UserProfileSettingsAndRespectFlow verifies user milestone 1 and 2 service flows.
func Test06UserProfileSettingsAndRespectFlow(t *testing.T) {
	database := openProfileDatabase(t)
	repository, _ := userstore.NewRepository(database)
	service, _ := userapplication.NewService(repository)
	actor, _ := service.Create(context.Background(), "actor")
	target, _ := service.Create(context.Background(), "target")
	motto := "new motto"
	figure := "hd-180-1"
	gender := "F"
	home := 77
	updated, err := service.UpdateProfile(context.Background(), target.ID, domain.ProfilePatch{Motto: &motto, Figure: &figure, Gender: &gender, HomeRoomID: &home})
	if err != nil {
		t.Fatalf("expected profile update success, got %v", err)
	}
	if updated.Motto != motto || updated.Figure != figure || updated.Gender != "F" || updated.HomeRoomID != home {
		t.Fatalf("unexpected updated profile payload %+v", updated)
	}
	settings, err := service.LoadSettings(context.Background(), target.ID)
	if err != nil || settings.VolumeSystem != 100 {
		t.Fatalf("expected default settings load success, got %+v err=%v", settings, err)
	}
	volume := 25
	oldChat := true
	saved, err := service.SaveSettings(context.Background(), target.ID, domain.SettingsPatch{VolumeSystem: &volume, OldChat: &oldChat})
	if err != nil || saved.VolumeSystem != 25 || !saved.OldChat {
		t.Fatalf("unexpected saved settings payload %+v err=%v", saved, err)
	}
	now := time.Date(2026, time.March, 13, 10, 0, 0, 0, time.UTC)
	for idx := 0; idx < 3; idx++ {
		if _, err := service.RecordUserRespect(context.Background(), actor.ID, target.ID, now); err != nil {
			t.Fatalf("expected respect success, got %v", err)
		}
	}
	if _, err := service.RecordUserRespect(context.Background(), actor.ID, target.ID, now); !errors.Is(err, domain.ErrRespectLimitReached) {
		t.Fatalf("expected respect limit error, got %v", err)
	}
}

// openProfileDatabase creates sqlite database with user profile schemas.
func openProfileDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&usermodel.Record{}, &usermodel.LoginEvent{}, &usermodel.Settings{}, &usermodel.Respect{}); err != nil {
		t.Fatalf("expected sqlite migration success, got %v", err)
	}
	return database
}
