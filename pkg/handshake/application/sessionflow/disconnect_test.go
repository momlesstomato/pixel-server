package sessionflow

import (
	"testing"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
)

// disconnectRegistryStub defines in-memory session storage behavior.
type disconnectRegistryStub struct {
	// sessions stores one connection-keyed session map.
	sessions map[string]coreconnection.Session
}

// Register stores one session.
func (stub *disconnectRegistryStub) Register(session coreconnection.Session) error {
	stub.sessions[session.ConnID] = session
	return nil
}

// FindByConnID resolves one session by connection identifier.
func (stub *disconnectRegistryStub) FindByConnID(connID string) (coreconnection.Session, bool) {
	value, found := stub.sessions[connID]
	return value, found
}

// Remove deletes one session by connection identifier.
func (stub *disconnectRegistryStub) Remove(connID string) {
	delete(stub.sessions, connID)
}

// disconnectTransportStub defines close capture behavior.
type disconnectTransportStub struct {
	// closed stores closed connection identifiers.
	closed []string
}

// Send is a no-op for disconnect tests.
func (stub *disconnectTransportStub) Send(_ string, _ uint16, _ []byte) error {
	return nil
}

// Close captures closed connection identifiers.
func (stub *disconnectTransportStub) Close(connID string, _ int, _ string) error {
	stub.closed = append(stub.closed, connID)
	return nil
}

// TestNewDisconnectUseCaseRejectsMissingDependencies verifies constructor checks.
func TestNewDisconnectUseCaseRejectsMissingDependencies(t *testing.T) {
	if _, err := NewDisconnectUseCase(nil, &disconnectTransportStub{}); err == nil {
		t.Fatalf("expected session registry precondition error")
	}
	if _, err := NewDisconnectUseCase(&disconnectRegistryStub{sessions: map[string]coreconnection.Session{}}, nil); err == nil {
		t.Fatalf("expected transport precondition error")
	}
}

// TestDisconnectUseCaseDisconnectsAndRemovesSession verifies disconnect workflow.
func TestDisconnectUseCaseDisconnectsAndRemovesSession(t *testing.T) {
	registry := &disconnectRegistryStub{sessions: map[string]coreconnection.Session{"conn-1": {ConnID: "conn-1", State: coreconnection.StateAuthenticated}}}
	transport := &disconnectTransportStub{}
	useCase, err := NewDisconnectUseCase(registry, transport)
	if err != nil {
		t.Fatalf("expected constructor success, got %v", err)
	}
	if err := useCase.Disconnect("conn-1"); err != nil {
		t.Fatalf("expected disconnect success, got %v", err)
	}
	if len(transport.closed) != 1 || transport.closed[0] != "conn-1" {
		t.Fatalf("expected connection close call, got %v", transport.closed)
	}
	if _, found := registry.FindByConnID("conn-1"); found {
		t.Fatalf("expected session removed after disconnect")
	}
}

// TestDisconnectUseCaseCleanupHandlesAbruptClose verifies cleanup behavior.
func TestDisconnectUseCaseCleanupHandlesAbruptClose(t *testing.T) {
	registry := &disconnectRegistryStub{sessions: map[string]coreconnection.Session{"conn-1": {ConnID: "conn-1"}}}
	useCase, _ := NewDisconnectUseCase(registry, &disconnectTransportStub{})
	useCase.Cleanup("conn-1")
	if _, found := registry.FindByConnID("conn-1"); found {
		t.Fatalf("expected session removal on cleanup")
	}
	useCase.Cleanup("")
}
