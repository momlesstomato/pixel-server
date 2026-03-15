package tests

import (
	"context"
	"errors"
	"testing"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/authflow"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
)

// userFinderStub defines deterministic user display name resolution.
type userFinderStub struct {
	// username stores the resolved display name.
	username string
	// err stores the resolution failure output.
	err error
}

// FindByID returns configured username or error.
func (stub userFinderStub) FindByID(_ context.Context, _ int) (string, error) {
	if stub.err != nil {
		return "", stub.err
	}
	return stub.username, nil
}

// newUsersUseCase creates an AuthenticateUseCase wired with provided user finder.
func newUsersUseCase(finder authflow.UserFinder) (*authflow.AuthenticateUseCase, *transportStub, *sessionStub) {
	sessions := &sessionStub{byConn: map[string]coreconnection.Session{}, byUser: map[int]string{}}
	transport := &transportStub{}
	useCase, _ := authflow.NewAuthenticateUseCase(validatorStub{userID: 5}, sessions, transport)
	useCase.SetUserFinder(finder)
	return useCase, transport, sessions
}

// TestUserFinderResolvedUsernameUsedInIdentityPacket verifies real username is sent when UserFinder is set.
func TestUserFinderResolvedUsernameUsedInIdentityPacket(t *testing.T) {
	useCase, transport, _ := newUsersUseCase(userFinderStub{username: "habbo_player"})
	_, err := useCase.Authenticate(context.Background(), authflow.AuthenticateRequest{ConnID: "conn-1", Ticket: "ticket"})
	if err != nil {
		t.Fatalf("expected auth success, got %v", err)
	}
	if len(transport.sent) != 2 || transport.sent[1] != packetauth.IdentityAccountsPacketID {
		t.Fatalf("expected identity accounts packet, got %v", transport.sent)
	}
}

// TestUserFinderNotFoundDisconnectsWithReason verifies disconnect on user not found.
func TestUserFinderNotFoundDisconnectsWithReason(t *testing.T) {
	useCase, transport, _ := newUsersUseCase(userFinderStub{err: errors.New("user not found")})
	_, err := useCase.Authenticate(context.Background(), authflow.AuthenticateRequest{ConnID: "conn-2", Ticket: "ticket"})
	if err == nil {
		t.Fatalf("expected error when user not found")
	}
	if len(transport.closed) == 0 || transport.closed[0] != "conn-2" {
		t.Fatalf("expected connection closed on user-not-found, got %v", transport.closed)
	}
	if len(transport.sent) == 0 || transport.sent[0] != packetauth.DisconnectReasonPacketID {
		t.Fatalf("expected disconnect_reason packet before close, got %v", transport.sent)
	}
}

// TestUserFinderNilFallsBackToPlayerID verifies Player#ID fallback when UserFinder is nil.
func TestUserFinderNilFallsBackToPlayerID(t *testing.T) {
	sessions := &sessionStub{byConn: map[string]coreconnection.Session{}, byUser: map[int]string{}}
	transport := &transportStub{}
	useCase, _ := authflow.NewAuthenticateUseCase(validatorStub{userID: 9}, sessions, transport)
	_, err := useCase.Authenticate(context.Background(), authflow.AuthenticateRequest{ConnID: "conn-3", Ticket: "ticket"})
	if err != nil {
		t.Fatalf("expected fallback auth success, got %v", err)
	}
	if len(transport.sent) != 2 || transport.sent[1] != packetauth.IdentityAccountsPacketID {
		t.Fatalf("expected identity accounts packet with fallback name, got %v", transport.sent)
	}
}

// TestUserFinderErrorDoesNotRegisterSession verifies no session is stored when user lookup fails.
func TestUserFinderErrorDoesNotRegisterSession(t *testing.T) {
	useCase, _, sessions := newUsersUseCase(userFinderStub{err: errors.New("database unavailable")})
	_, err := useCase.Authenticate(context.Background(), authflow.AuthenticateRequest{ConnID: "conn-4", Ticket: "ticket"})
	if err == nil {
		t.Fatalf("expected error when user finder fails")
	}
	if _, found := sessions.FindByUserID(5); found {
		t.Fatalf("expected no session registered when user lookup fails before session registration")
	}
}
