package tests

import (
	"context"
	"errors"
	"testing"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/authflow"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
)

// banCheckerStub defines deterministic hotel ban check behavior.
type banCheckerStub struct {
	// banned stores the ban status result.
	banned bool
	// err stores the optional check failure.
	err error
}

// IsHotelBanned returns configured ban status or error.
func (stub banCheckerStub) IsHotelBanned(_ context.Context, _ int) (bool, error) {
	if stub.err != nil {
		return false, stub.err
	}
	return stub.banned, nil
}

// newBanUseCase creates an AuthenticateUseCase wired with ban checker.
func newBanUseCase(
	checker authflow.BanChecker,
) (*authflow.AuthenticateUseCase, *transportStub, *sessionStub) {
	sessions := &sessionStub{
		byConn: map[string]coreconnection.Session{},
		byUser: map[int]string{},
	}
	transport := &transportStub{}
	uc, _ := authflow.NewAuthenticateUseCase(
		validatorStub{userID: 7}, sessions, transport,
	)
	uc.SetBanChecker(checker)
	return uc, transport, sessions
}

// TestBanCheckerAllowsUnbannedUser verifies success for unbanned user.
func TestBanCheckerAllowsUnbannedUser(t *testing.T) {
	uc, transport, sessions := newBanUseCase(
		banCheckerStub{banned: false},
	)
	req := authflow.AuthenticateRequest{
		ConnID: "conn-1", Ticket: "ticket",
	}
	_, err := uc.Authenticate(context.Background(), req)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(transport.closed) != 0 {
		t.Fatalf("unexpected close: %v", transport.closed)
	}
	if _, ok := sessions.byConn["conn-1"]; !ok {
		t.Fatal("expected session registered")
	}
}

// TestBanCheckerRejectsBannedUser verifies disconnect for banned user.
func TestBanCheckerRejectsBannedUser(t *testing.T) {
	uc, transport, sessions := newBanUseCase(
		banCheckerStub{banned: true},
	)
	req := authflow.AuthenticateRequest{
		ConnID: "conn-2", Ticket: "ticket",
	}
	_, err := uc.Authenticate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for banned user")
	}
	if len(transport.closed) == 0 {
		t.Fatal("expected connection closed")
	}
	if transport.closed[0] != "conn-2" {
		t.Fatalf("wrong conn closed: %s", transport.closed[0])
	}
	if transport.sent[0] != packetauth.DisconnectReasonPacketID {
		t.Fatal("expected disconnect_reason packet")
	}
	if _, ok := sessions.byConn["conn-2"]; ok {
		t.Fatal("no session expected for banned user")
	}
}

// TestBanCheckerErrorRejects verifies disconnect on ban check error.
func TestBanCheckerErrorRejects(t *testing.T) {
	uc, transport, _ := newBanUseCase(
		banCheckerStub{err: errors.New("redis timeout")},
	)
	req := authflow.AuthenticateRequest{
		ConnID: "conn-3", Ticket: "ticket",
	}
	_, err := uc.Authenticate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error on ban check failure")
	}
	if len(transport.closed) == 0 {
		t.Fatal("expected close on ban check error")
	}
}

// TestBanCheckerNilSkipsCheck verifies auth proceeds without checker.
func TestBanCheckerNilSkipsCheck(t *testing.T) {
	sessions := &sessionStub{
		byConn: map[string]coreconnection.Session{},
		byUser: map[int]string{},
	}
	transport := &transportStub{}
	uc, _ := authflow.NewAuthenticateUseCase(
		validatorStub{userID: 3}, sessions, transport,
	)
	req := authflow.AuthenticateRequest{
		ConnID: "conn-4", Ticket: "ticket",
	}
	_, err := uc.Authenticate(context.Background(), req)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if _, ok := sessions.byConn["conn-4"]; !ok {
		t.Fatal("expected session when checker is nil")
	}
}
