package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// TestNewServiceRejectsNilRepository verifies constructor precondition validation.
func TestNewServiceRejectsNilRepository(t *testing.T) {
	if _, err := NewService(nil); err == nil {
		t.Fatalf("expected nil repository validation failure")
	}
}

// TestServiceCreateValidatesUsername verifies username validation behavior.
func TestServiceCreateValidatesUsername(t *testing.T) {
	service, _ := NewService(repositoryStub{})
	if _, err := service.Create(context.Background(), " "); err == nil {
		t.Fatalf("expected create failure for empty username")
	}
	if _, err := service.Create(context.Background(), longUsername()); err == nil {
		t.Fatalf("expected create failure for oversized username")
	}
}

// TestServiceCreateAndFindByID verifies create and find flow behavior.
func TestServiceCreateAndFindByID(t *testing.T) {
	stub := repositoryStub{saved: domain.User{ID: 3, Username: "alpha"}}
	service, _ := NewService(stub)
	created, err := service.Create(context.Background(), " alpha ")
	if err != nil {
		t.Fatalf("expected create success, got %v", err)
	}
	if created.ID != 3 || created.Username != "alpha" {
		t.Fatalf("unexpected created user payload")
	}
	if _, err := service.FindByID(context.Background(), 0); err == nil {
		t.Fatalf("expected find failure for invalid id")
	}
	loaded, err := service.FindByID(context.Background(), 3)
	if err != nil || loaded.ID != 3 {
		t.Fatalf("expected find success, got %v and %+v", err, loaded)
	}
}

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
func (stub repositoryStub) DeleteByID(_ context.Context, _ int) error {
	return stub.deleteErr
}

// RecordLogin returns deterministic login stamp output.
func (stub repositoryStub) RecordLogin(_ context.Context, _ int, _ string, _ time.Time) (bool, error) {
	if stub.loginErr != nil {
		return false, stub.loginErr
	}
	return stub.firstLogin, nil
}

// longUsername returns a username bigger than allowed max length.
func longUsername() string {
	value := ""
	for index := 0; index < 65; index++ {
		value += "a"
	}
	return value
}

// TestServiceFindByIDReturnsRepositoryError verifies repository error propagation.
func TestServiceFindByIDReturnsRepositoryError(t *testing.T) {
	service, _ := NewService(repositoryStub{findErr: errors.New("boom")})
	if _, err := service.FindByID(context.Background(), 1); err == nil {
		t.Fatalf("expected find failure")
	}
}

// TestServiceDeleteByIDValidatesAndPropagatesErrors verifies delete validation and error propagation.
func TestServiceDeleteByIDValidatesAndPropagatesErrors(t *testing.T) {
	service, _ := NewService(repositoryStub{})
	if err := service.DeleteByID(context.Background(), 0); err == nil {
		t.Fatalf("expected delete failure for invalid id")
	}
	service, _ = NewService(repositoryStub{deleteErr: errors.New("boom")})
	if err := service.DeleteByID(context.Background(), 1); err == nil {
		t.Fatalf("expected delete failure propagation")
	}
}

// TestServiceRecordLoginValidatesAndReturnsResult verifies login stamp behavior.
func TestServiceRecordLoginValidatesAndReturnsResult(t *testing.T) {
	service, _ := NewService(repositoryStub{firstLogin: true})
	if _, err := service.RecordLogin(context.Background(), 0, "pixel-server", time.Now()); err == nil {
		t.Fatalf("expected login stamp failure for invalid user id")
	}
	if _, err := service.RecordLogin(context.Background(), 1, " ", time.Now()); err == nil {
		t.Fatalf("expected login stamp failure for empty holder")
	}
	if _, err := service.RecordLogin(context.Background(), 1, "pixel-server", time.Time{}); err == nil {
		t.Fatalf("expected login stamp failure for zero timestamp")
	}
	firstLogin, err := service.RecordLogin(context.Background(), 1, "pixel-server", time.Now())
	if err != nil || !firstLogin {
		t.Fatalf("expected first-login stamp success, got %v and %v", err, firstLogin)
	}
	service, _ = NewService(repositoryStub{loginErr: errors.New("boom")})
	if _, err := service.RecordLogin(context.Background(), 1, "pixel-server", time.Now()); err == nil {
		t.Fatalf("expected login stamp failure propagation")
	}
}
