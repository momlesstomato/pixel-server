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
	wardrobeRequest := httptest.NewRequest(http.MethodGet, "/api/v1/users/7/wardrobe", nil)
	wardrobeResponse, err := module.App().Test(wardrobeRequest)
	if err != nil || wardrobeResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected get wardrobe success, got status=%d err=%v", wardrobeResponse.StatusCode, err)
	}
	respectsRequest := httptest.NewRequest(http.MethodGet, "/api/v1/users/7/respects?limit=20&offset=0", nil)
	respectsResponse, err := module.App().Test(respectsRequest)
	if err != nil || respectsResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected get respects success, got status=%d err=%v", respectsResponse.StatusCode, err)
	}
	namePayload, _ := json.Marshal(map[string]string{"name": "beta"})
	nameChangeRequest := httptest.NewRequest(http.MethodPost, "/api/v1/users/7/name-change", bytes.NewReader(namePayload))
	nameChangeRequest.Header.Set("Content-Type", "application/json")
	nameChangeResponse, err := module.App().Test(nameChangeRequest)
	if err != nil || nameChangeResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected post name-change success, got status=%d err=%v", nameChangeResponse.StatusCode, err)
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

// LoadWardrobe returns deterministic wardrobe slots payload.
func (stub serviceStub) LoadWardrobe(context.Context, int) ([]domain.WardrobeSlot, error) {
	return []domain.WardrobeSlot{{SlotID: 1, Figure: "hr-1", Gender: "M"}}, nil
}

// ListRespects returns deterministic respect history payload.
func (stub serviceStub) ListRespects(context.Context, int, int, int) ([]domain.RespectRecord, error) {
	return []domain.RespectRecord{{ID: 1, ActorUserID: 2, TargetID: 7, TargetType: domain.RespectTargetUser}}, nil
}

// ForceChangeName returns deterministic name change result payload.
func (stub serviceStub) ForceChangeName(context.Context, int, string) (domain.NameResult, error) {
	return domain.NameResult{ResultCode: domain.NameResultAvailable, Name: "beta"}, nil
}
