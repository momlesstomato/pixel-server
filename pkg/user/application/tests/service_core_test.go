package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkuser "github.com/momlesstomato/pixel-sdk/events/user"
	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// TestNewServiceRejectsNilRepository verifies constructor precondition validation.
func TestNewServiceRejectsNilRepository(t *testing.T) {
	if _, err := userapplication.NewService(nil); err == nil {
		t.Fatalf("expected nil repository validation failure")
	}
}

// TestServiceCreateAndFindByID verifies create and find flow behavior.
func TestServiceCreateAndFindByID(t *testing.T) {
	stub := coreRepositoryStub{saved: domain.User{ID: 3, Username: "alpha"}}
	service, _ := userapplication.NewService(stub)
	created, err := service.Create(context.Background(), " alpha ")
	if err != nil || created.ID != 3 || created.Username != "alpha" {
		t.Fatalf("unexpected create result %+v err=%v", created, err)
	}
	if _, err := service.FindByID(context.Background(), 0); err == nil {
		t.Fatalf("expected find failure for invalid id")
	}
	loaded, err := service.FindByID(context.Background(), 3)
	if err != nil || loaded.ID != 3 {
		t.Fatalf("unexpected find result %+v err=%v", loaded, err)
	}
}

// TestServiceDeleteLoginAndNameFlows verifies delete, login, and name operation behavior.
func TestServiceDeleteLoginAndNameFlows(t *testing.T) {
	service, _ := userapplication.NewService(coreRepositoryStub{saved: domain.User{ID: 1, Username: "alpha"}, firstLogin: true})
	if err := service.DeleteByID(context.Background(), 0); err == nil {
		t.Fatalf("expected delete failure for invalid id")
	}
	firstLogin, err := service.RecordLogin(context.Background(), 1, "pixel-server", time.Now().UTC())
	if err != nil || !firstLogin {
		t.Fatalf("unexpected login result first=%v err=%v", firstLogin, err)
	}
	result, err := service.CheckName(context.Background(), " ", 1)
	if err != nil || result.ResultCode != domain.NameResultInvalid {
		t.Fatalf("unexpected name check result %+v err=%v", result, err)
	}
	result, err = service.ForceChangeName(context.Background(), 1, "bravo")
	if err != nil || result.ResultCode != domain.NameResultAvailable || result.Name != "bravo" {
		t.Fatalf("unexpected force change result %+v err=%v", result, err)
	}
}

// TestServicePluginCancellableEvents verifies cancellable user events block writes.
func TestServicePluginCancellableEvents(t *testing.T) {
	service, _ := userapplication.NewService(coreRepositoryStub{saved: domain.User{ID: 1, Username: "alpha", Motto: "old", Figure: "old-figure"}})
	service.SetEventFirer(func(event sdk.Event) {
		switch value := event.(type) {
		case *sdkuser.MottoChanged:
			value.Cancel()
		case *sdkuser.Respected:
			value.Cancel()
		}
	})
	if _, err := service.UpdateMotto(context.Background(), "conn-1", 1, "new"); err == nil {
		t.Fatalf("expected motto change cancellation")
	}
	if _, err := service.RecordUserRespectWithConn(context.Background(), "conn-1", 1, 2, time.Now().UTC()); err == nil {
		t.Fatalf("expected respect cancellation")
	}
}

// TestServicePropagatesRepositoryErrors verifies repository error propagation.
func TestServicePropagatesRepositoryErrors(t *testing.T) {
	service, _ := userapplication.NewService(coreRepositoryStub{findErr: errors.New("boom"), deleteErr: errors.New("boom"), loginErr: errors.New("boom")})
	if _, err := service.FindByID(context.Background(), 1); err == nil {
		t.Fatalf("expected find failure")
	}
	if err := service.DeleteByID(context.Background(), 1); err == nil {
		t.Fatalf("expected delete failure")
	}
	if _, err := service.RecordLogin(context.Background(), 1, "pixel-server", time.Now().UTC()); err == nil {
		t.Fatalf("expected login failure")
	}
}
