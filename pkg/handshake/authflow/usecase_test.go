package authflow

import (
	"context"
	"errors"
	"testing"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
)

// validatorStub defines deterministic ticket validation behavior.
type validatorStub struct {
	// userID stores successful validation output.
	userID int
	// err stores validation failure output.
	err error
}

// Validate resolves configured validator behavior.
func (stub validatorStub) Validate(_ context.Context, _ string) (int, error) {
	if stub.err != nil {
		return 0, stub.err
	}
	return stub.userID, nil
}

// transportStub defines packet and close capture behavior.
type transportStub struct {
	// sent stores packet identifiers emitted by flow.
	sent []uint16
	// closed stores closed connection identifiers.
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

// sessionStub defines in-memory session registry behavior.
type sessionStub struct {
	// byConn stores session records by connection id.
	byConn map[string]coreconnection.Session
	// byUser stores user index to connection id.
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

// TestNewAuthenticateUseCaseRejectsMissingDependencies verifies constructor checks.
func TestNewAuthenticateUseCaseRejectsMissingDependencies(t *testing.T) {
	if _, err := NewAuthenticateUseCase(nil, &sessionStub{}, &transportStub{}); err == nil {
		t.Fatalf("expected validator precondition error")
	}
	if _, err := NewAuthenticateUseCase(validatorStub{userID: 1}, nil, &transportStub{}); err == nil {
		t.Fatalf("expected session precondition error")
	}
	if _, err := NewAuthenticateUseCase(validatorStub{userID: 1}, &sessionStub{}, nil); err == nil {
		t.Fatalf("expected transport precondition error")
	}
}

// TestAuthenticateUseCaseAuthenticatesAndSendsPackets verifies success flow.
func TestAuthenticateUseCaseAuthenticatesAndSendsPackets(t *testing.T) {
	sessions := &sessionStub{byConn: map[string]coreconnection.Session{}, byUser: map[int]string{}}
	transport := &transportStub{}
	useCase, _ := NewAuthenticateUseCase(validatorStub{userID: 9}, sessions, transport)
	result, err := useCase.Authenticate(context.Background(), AuthenticateRequest{ConnID: "conn-new", Ticket: "ticket", MachineID: "machine"})
	if err != nil || result.UserID != 9 || result.KickedConnID != "" {
		t.Fatalf("unexpected authenticate result: %#v err=%v", result, err)
	}
	if len(transport.sent) != 2 || transport.sent[0] != packetauth.AuthenticationOKPacketID || transport.sent[1] != packetauth.IdentityAccountsPacketID {
		t.Fatalf("expected auth packet sequence, got %v", transport.sent)
	}
	session, found := sessions.FindByUserID(9)
	if !found || session.ConnID != "conn-new" || session.State != coreconnection.StateAuthenticated {
		t.Fatalf("expected authenticated session, got %#v found=%v", session, found)
	}
}

// TestAuthenticateUseCaseKicksDuplicateSession verifies duplicate login kick.
func TestAuthenticateUseCaseKicksDuplicateSession(t *testing.T) {
	sessions := &sessionStub{byConn: map[string]coreconnection.Session{"conn-old": {ConnID: "conn-old", UserID: 9}}, byUser: map[int]string{9: "conn-old"}}
	transport := &transportStub{}
	useCase, _ := NewAuthenticateUseCase(validatorStub{userID: 9}, sessions, transport)
	result, err := useCase.Authenticate(context.Background(), AuthenticateRequest{ConnID: "conn-new", Ticket: "ticket"})
	if err != nil || result.KickedConnID != "conn-old" {
		t.Fatalf("expected duplicate kick result, got %#v err=%v", result, err)
	}
	if len(transport.closed) != 1 || transport.closed[0] != "conn-old" {
		t.Fatalf("expected old connection close, got %v", transport.closed)
	}
}

// TestAuthenticateUseCaseRejectsInvalidInputs verifies error handling behavior.
func TestAuthenticateUseCaseRejectsInvalidInputs(t *testing.T) {
	sessions := &sessionStub{byConn: map[string]coreconnection.Session{}, byUser: map[int]string{}}
	transport := &transportStub{}
	useCase, _ := NewAuthenticateUseCase(validatorStub{err: errors.New("invalid")}, sessions, transport)
	if _, err := useCase.Authenticate(context.Background(), AuthenticateRequest{ConnID: "", Ticket: "ticket"}); err == nil {
		t.Fatalf("expected connection id error")
	}
	if _, err := useCase.Authenticate(context.Background(), AuthenticateRequest{ConnID: "conn-new", Ticket: " "}); err == nil {
		t.Fatalf("expected ticket validation error")
	}
	if _, err := useCase.Authenticate(context.Background(), AuthenticateRequest{ConnID: "conn-new", Ticket: "ticket"}); err == nil {
		t.Fatalf("expected validator error")
	}
	if len(transport.closed) == 0 {
		t.Fatalf("expected unauthorized close calls")
	}
}
