package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	corehttp "github.com/momlesstomato/pixel-server/core/http"
	userhttpapi "github.com/momlesstomato/pixel-server/pkg/user/adapter/httpapi"
	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// TestRegisterRoutesAndHandlers verifies user routes behavior.
func TestRegisterRoutesAndHandlers(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	service := serviceStub{}
	if err := userhttpapi.RegisterRoutes(module, service); err != nil {
		t.Fatalf("expected register routes success, got %v", err)
	}
	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/users/7", nil)
	getResponse, err := module.App().Test(getRequest)
	if err != nil || getResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected get user success, got status=%d err=%v", getResponse.StatusCode, err)
	}
	settingsRequest := httptest.NewRequest(http.MethodGet, "/api/v1/users/7/settings", nil)
	settingsResponse, err := module.App().Test(settingsRequest)
	if err != nil || settingsResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected get settings success, got status=%d err=%v", settingsResponse.StatusCode, err)
	}
}

// TestRespectRouteMapsConflict verifies respect conflict mapping behavior.
func TestRespectRouteMapsConflict(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	if err := userhttpapi.RegisterRoutes(module, serviceStub{respectErr: domain.ErrRespectLimitReached}); err != nil {
		t.Fatalf("expected register routes success, got %v", err)
	}
	payload, _ := json.Marshal(map[string]int{"actor_user_id": 1})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/users/2/respect", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected request success, got %v", err)
	}
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", response.StatusCode)
	}
}

// serviceStub defines deterministic user service behavior.
type serviceStub struct {
	// respectErr stores deterministic respect operation error.
	respectErr error
}

// FindByID returns deterministic user payload.
func (stub serviceStub) FindByID(context.Context, int) (domain.User, error) {
	return domain.User{ID: 7, Username: "alpha"}, nil
}

// UpdateProfile returns deterministic user payload.
func (stub serviceStub) UpdateProfile(context.Context, int, domain.ProfilePatch) (domain.User, error) {
	return domain.User{ID: 7, Username: "alpha"}, nil
}

// LoadSettings returns deterministic settings payload.
func (stub serviceStub) LoadSettings(context.Context, int) (domain.Settings, error) {
	return domain.Settings{UserID: 7, VolumeSystem: 100, VolumeFurni: 100, VolumeTrax: 100, RoomInvites: true, CameraFollow: true}, nil
}

// SaveSettings returns deterministic settings payload.
func (stub serviceStub) SaveSettings(context.Context, int, domain.SettingsPatch) (domain.Settings, error) {
	return domain.Settings{UserID: 7, VolumeSystem: 10}, nil
}

// RecordUserRespect returns deterministic respect result.
func (stub serviceStub) RecordUserRespect(context.Context, int, int, time.Time) (userapplication.RespectResult, error) {
	if stub.respectErr != nil {
		return userapplication.RespectResult{}, stub.respectErr
	}
	return userapplication.RespectResult{RespectsReceived: 2, Remaining: 1}, nil
}
