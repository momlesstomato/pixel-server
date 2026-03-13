package tests

import (
	"testing"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	coreplugin "github.com/momlesstomato/pixel-server/core/plugin"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/sessionflow"
	sdk "github.com/momlesstomato/pixel-sdk"
)

// disconnectRegistryStub defines in-memory session storage behavior.
type disconnectRegistryStub struct {
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

// TestCancelSessionDisconnectingPreventsDisconnect verifies plugin cancellation blocks disconnect.
func TestCancelSessionDisconnectingPreventsDisconnect(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(nil)
	dispatcher.Subscribe("test", func(e *sdk.SessionDisconnecting) { e.Cancel() })
	registry := &disconnectRegistryStub{sessions: map[string]coreconnection.Session{
		"conn-1": {ConnID: "conn-1", UserID: 10, State: coreconnection.StateAuthenticated},
	}}
	transport := &disconnectTransportStub{}
	useCase, err := sessionflow.NewDisconnectUseCase(registry, transport)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	useCase.SetEventFirer(dispatcher.Fire)
	if err := useCase.Disconnect("conn-1"); err != nil {
		t.Fatalf("expected no error on cancelled disconnect, got %v", err)
	}
	if len(transport.closed) != 0 {
		t.Fatalf("expected no connection close when disconnect cancelled, got %v", transport.closed)
	}
	if _, found := registry.FindByConnID("conn-1"); !found {
		t.Fatalf("expected session preserved when disconnect cancelled")
	}
}

// TestSessionDisconnectingEventFiresOnGracefulDisconnect verifies event fires.
func TestSessionDisconnectingEventFiresOnGracefulDisconnect(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(nil)
	var fired bool
	dispatcher.Subscribe("test", func(_ *sdk.SessionDisconnecting) { fired = true })
	registry := &disconnectRegistryStub{sessions: map[string]coreconnection.Session{
		"conn-1": {ConnID: "conn-1", UserID: 5, State: coreconnection.StateAuthenticated},
	}}
	useCase, _ := sessionflow.NewDisconnectUseCase(registry, &disconnectTransportStub{})
	useCase.SetEventFirer(dispatcher.Fire)
	_ = useCase.Disconnect("conn-1")
	if !fired {
		t.Fatalf("expected SessionDisconnecting event to fire")
	}
}

// TestDisconnectWithoutEventFirerStillWorks verifies nil firer does not cause panic.
func TestDisconnectWithoutEventFirerStillWorks(t *testing.T) {
	registry := &disconnectRegistryStub{sessions: map[string]coreconnection.Session{
		"conn-1": {ConnID: "conn-1", UserID: 5},
	}}
	transport := &disconnectTransportStub{}
	useCase, _ := sessionflow.NewDisconnectUseCase(registry, transport)
	if err := useCase.Disconnect("conn-1"); err != nil {
		t.Fatalf("expected disconnect to succeed without event firer, got %v", err)
	}
	if len(transport.closed) != 1 {
		t.Fatalf("expected connection close call, got %v", transport.closed)
	}
}
