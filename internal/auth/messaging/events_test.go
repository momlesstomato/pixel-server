package messaging

import "testing"

// TestEventTicketValidatedConstant validates event naming contract.
func TestEventTicketValidatedConstant(t *testing.T) {
	if EventTicketValidated != "auth.ticket.validated" {
		t.Fatalf("unexpected event name: %s", EventTicketValidated)
	}
}

// TestHandshakeEventConstants validates handshake event naming contracts.
func TestHandshakeEventConstants(t *testing.T) {
	if EventReleaseVersionReceived == "" || EventDiffieInitialized == "" || EventDiffieCompleted == "" || EventMachineIDReceived == "" {
		t.Fatalf("expected non-empty handshake event names")
	}
}
