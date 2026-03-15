package tests

import (
	"context"
	"errors"
	"testing"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/authflow"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
)

// userFinderStub defines deterministic user existence check behavior.
type userFinderStub struct {
	// err stores the error returned by FindByID.
	err error
}

// FindByID returns the configured error or nil.
func (stub userFinderStub) FindByID(_ context.Context, _ int) error {
	return stub.err
}

// TestAuthenticateUseCaseRejectsUnknownUser verifies disconnect when user not found.
func TestAuthenticateUseCaseRejectsUnknownUser(t *testing.T) {
	sessions := &sessionStub{byConn: map[string]coreconnection.Session{}, byUser: map[int]string{}}
	transport := &transportStub{}
	useCase, err := authflow.NewAuthenticateUseCase(validatorStub{userID: 5}, sessions, transport)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	useCase.SetUserFinder(userFinderStub{err: errors.New("user not found")})
	_, authErr := useCase.Authenticate(context.Background(), authflow.AuthenticateRequest{ConnID: "conn-1", Ticket: "ticket"})
	if authErr == nil {
		t.Fatalf("expected auth to fail for unknown user")
	}
	if len(transport.closed) != 1 || transport.closed[0] != "conn-1" {
		t.Fatalf("expected connection closed for unknown user, got %v", transport.closed)
	}
	if len(transport.sent) == 0 || transport.sent[0] != packetauth.DisconnectReasonPacketID {
		t.Fatalf("expected disconnect_reason packet before close, got %v", transport.sent)
	}
	if _, found := sessions.FindByUserID(5); found {
		t.Fatalf("expected no session registered for unknown user")
	}
}

// TestAuthenticateUseCaseSucceedsWithKnownUser verifies normal auth flow when user exists.
func TestAuthenticateUseCaseSucceedsWithKnownUser(t *testing.T) {
	sessions := &sessionStub{byConn: map[string]coreconnection.Session{}, byUser: map[int]string{}}
	transport := &transportStub{}
	useCase, err := authflow.NewAuthenticateUseCase(validatorStub{userID: 5}, sessions, transport)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	useCase.SetUserFinder(userFinderStub{err: nil})
	result, authErr := useCase.Authenticate(context.Background(), authflow.AuthenticateRequest{ConnID: "conn-1", Ticket: "ticket"})
	if authErr != nil {
		t.Fatalf("expected auth success for known user, got %v", authErr)
	}
	if result.UserID != 5 {
		t.Fatalf("expected user id 5, got %d", result.UserID)
	}
	if _, found := sessions.FindByUserID(5); !found {
		t.Fatalf("expected session registered for known user")
	}
}

// TestAuthenticateUseCaseSkipsUserCheckWhenFinderAbsent verifies backward-compatible behavior.
func TestAuthenticateUseCaseSkipsUserCheckWhenFinderAbsent(t *testing.T) {
	sessions := &sessionStub{byConn: map[string]coreconnection.Session{}, byUser: map[int]string{}}
	transport := &transportStub{}
	useCase, err := authflow.NewAuthenticateUseCase(validatorStub{userID: 99}, sessions, transport)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	result, authErr := useCase.Authenticate(context.Background(), authflow.AuthenticateRequest{ConnID: "conn-2", Ticket: "ticket"})
	if authErr != nil {
		t.Fatalf("expected auth success when no user finder set, got %v", authErr)
	}
	if result.UserID != 99 {
		t.Fatalf("expected user id 99, got %d", result.UserID)
	}
}
