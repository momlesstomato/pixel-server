package tests

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	coreplugin "github.com/momlesstomato/pixel-server/core/plugin"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/authflow"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
)

// validatorStub defines deterministic ticket validation behavior.
type validatorStub struct {
	userID int
}

// Validate returns configured user identifier.
func (stub validatorStub) Validate(_ context.Context, _ string) (int, error) {
	return stub.userID, nil
}

// transportStub defines packet and close capture behavior.
type transportStub struct {
	sent   []uint16
	closed []string
}

// Send captures packet identifier.
func (stub *transportStub) Send(_ string, packetID uint16, _ []byte) error {
	stub.sent = append(stub.sent, packetID)
	return nil
}

// Close captures connection identifier.
func (stub *transportStub) Close(connID string, _ int, _ string) error {
	stub.closed = append(stub.closed, connID)
	return nil
}

// CloseWithProtocolReason captures the disconnect reason and closed connection identifier.
func (stub *transportStub) CloseWithProtocolReason(connID string, protocolReason int32, _ int, _ string) error {
	if protocolReason != 0 {
		stub.sent = append(stub.sent, packetauth.DisconnectReasonPacketID)
	}
	stub.closed = append(stub.closed, connID)
	return nil
}

// sessionStub defines in-memory session registry behavior.
type sessionStub struct {
	byConn map[string]coreconnection.Session
	byUser map[int]string
}

// Register stores one session record.
func (stub *sessionStub) Register(session coreconnection.Session) error {
	stub.byConn[session.ConnID] = session
	stub.byUser[session.UserID] = session.ConnID
	return nil
}

// FindByUserID resolves one session by user id.
func (stub *sessionStub) FindByUserID(userID int) (coreconnection.Session, bool) {
	connID, found := stub.byUser[userID]
	if !found {
		return coreconnection.Session{}, false
	}
	session, exists := stub.byConn[connID]
	return session, exists
}

// Remove deletes one session record by connection id.
func (stub *sessionStub) Remove(connID string) {
	if session, found := stub.byConn[connID]; found {
		delete(stub.byConn, connID)
		delete(stub.byUser, session.UserID)
	}
}

// TestCancelAuthValidatingPreventsAuthentication verifies plugin cancellation blocks auth.
func TestCancelAuthValidatingPreventsAuthentication(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(nil)
	dispatcher.Subscribe("test", func(e *sdk.AuthValidating) { e.Cancel() })
	sessions := &sessionStub{byConn: map[string]coreconnection.Session{}, byUser: map[int]string{}}
	transport := &transportStub{}
	useCase, err := authflow.NewAuthenticateUseCase(validatorStub{userID: 42}, sessions, transport)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	useCase.SetEventFirer(dispatcher.Fire)
	_, authErr := useCase.Authenticate(context.Background(), authflow.AuthenticateRequest{ConnID: "conn-1", Ticket: "ticket", MachineID: "machine"})
	if authErr == nil {
		t.Fatalf("expected auth to fail when cancelled by plugin")
	}
	if _, found := sessions.FindByUserID(42); found {
		t.Fatalf("expected no session registered when auth cancelled")
	}
	if len(transport.closed) != 1 || transport.closed[0] != "conn-1" {
		t.Fatalf("expected connection closed after cancellation, got %v", transport.closed)
	}
}

// TestCancelDuplicateKickAllowsBothSessions verifies cancelling DuplicateKick keeps old session.
func TestCancelDuplicateKickAllowsBothSessions(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(nil)
	dispatcher.Subscribe("test", func(e *sdk.DuplicateKick) { e.Cancel() })
	sessions := &sessionStub{
		byConn: map[string]coreconnection.Session{"old": {ConnID: "old", UserID: 42}},
		byUser: map[int]string{42: "old"},
	}
	transport := &transportStub{}
	useCase, err := authflow.NewAuthenticateUseCase(validatorStub{userID: 42}, sessions, transport)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	useCase.SetEventFirer(dispatcher.Fire)
	result, authErr := useCase.Authenticate(context.Background(), authflow.AuthenticateRequest{ConnID: "new", Ticket: "ticket"})
	if authErr != nil {
		t.Fatalf("expected auth to succeed, got %v", authErr)
	}
	if result.KickedConnID != "" {
		t.Fatalf("expected no kick when DuplicateKick cancelled, got %q", result.KickedConnID)
	}
}

// TestAuthCompletedEventFires verifies AuthCompleted fires after successful auth.
func TestAuthCompletedEventFires(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(nil)
	var completed bool
	dispatcher.Subscribe("test", func(_ *sdk.AuthCompleted) { completed = true })
	sessions := &sessionStub{byConn: map[string]coreconnection.Session{}, byUser: map[int]string{}}
	useCase, _ := authflow.NewAuthenticateUseCase(validatorStub{userID: 1}, sessions, &transportStub{})
	useCase.SetEventFirer(dispatcher.Fire)
	_, err := useCase.Authenticate(context.Background(), authflow.AuthenticateRequest{ConnID: "conn-1", Ticket: "ticket"})
	if err != nil {
		t.Fatalf("expected auth success, got %v", err)
	}
	if !completed {
		t.Fatalf("expected AuthCompleted event to fire")
	}
}
