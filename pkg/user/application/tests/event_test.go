package tests

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkuser "github.com/momlesstomato/pixel-sdk/events/user"
	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// TestProfileUpdatingEventCancelsUpdate verifies ProfileUpdating cancellation aborts profile update.
func TestProfileUpdatingEventCancelsUpdate(t *testing.T) {
	service, _ := userapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if value, ok := event.(*sdkuser.ProfileUpdating); ok {
			value.Cancel()
		}
	})
	if _, err := service.UpdateProfile(context.Background(), 1, domain.ProfilePatch{}); err == nil {
		t.Fatalf("expected profile update to be cancelled")
	}
}

// TestProfileUpdatingEventAllowsUpdate verifies ProfileUpdating passes through without cancellation.
func TestProfileUpdatingEventAllowsUpdate(t *testing.T) {
	var fired bool
	service, _ := userapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if _, ok := event.(*sdkuser.ProfileUpdating); ok {
			fired = true
		}
	})
	user, err := service.UpdateProfile(context.Background(), 1, domain.ProfilePatch{})
	if err != nil || user.ID != 1 {
		t.Fatalf("unexpected update result %+v err=%v", user, err)
	}
	if !fired {
		t.Fatalf("expected ProfileUpdating event to fire")
	}
}

// TestSettingsUpdatingEventCancelsUpdate verifies SettingsUpdating cancellation aborts settings save.
func TestSettingsUpdatingEventCancelsUpdate(t *testing.T) {
	service, _ := userapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if value, ok := event.(*sdkuser.SettingsUpdating); ok {
			value.Cancel()
		}
	})
	if _, err := service.SaveSettings(context.Background(), 1, domain.SettingsPatch{}); err == nil {
		t.Fatalf("expected settings save to be cancelled")
	}
}

// TestSettingsUpdatingEventAllowsUpdate verifies SettingsUpdating passes through without cancellation.
func TestSettingsUpdatingEventAllowsUpdate(t *testing.T) {
	var fired bool
	service, _ := userapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if _, ok := event.(*sdkuser.SettingsUpdating); ok {
			fired = true
		}
	})
	settings, err := service.SaveSettings(context.Background(), 1, domain.SettingsPatch{})
	if err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}
	if settings.UserID != 1 {
		t.Fatalf("unexpected settings result %+v", settings)
	}
	if !fired {
		t.Fatalf("expected SettingsUpdating event to fire")
	}
}
