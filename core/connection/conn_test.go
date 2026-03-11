package connection

import (
	"context"
	"testing"
)

// TestMemoryConnectionReadWriteFlow verifies in-memory transport behavior.
func TestMemoryConnectionReadWriteFlow(t *testing.T) {
	connection := NewMemoryConnection("conn-1", 1)
	if err := connection.PushInbound([]byte{1, 2, 3}); err != nil {
		t.Fatalf("expected inbound push success, got %v", err)
	}
	inbound, err := connection.Read(context.Background())
	if err != nil || len(inbound) != 3 {
		t.Fatalf("expected inbound read success, got %v, %v", inbound, err)
	}
	if err := connection.Write(context.Background(), []byte{4}); err != nil {
		t.Fatalf("expected outbound write success, got %v", err)
	}
	outbound, err := connection.ReadOutbound(context.Background())
	if err != nil || len(outbound) != 1 || outbound[0] != 4 {
		t.Fatalf("expected outbound read success, got %v, %v", outbound, err)
	}
	if err := connection.Dispose(); err != nil {
		t.Fatalf("expected dispose success, got %v", err)
	}
}

// TestMemoryConnectionRejectsOperationsAfterDispose verifies close-state validation.
func TestMemoryConnectionRejectsOperationsAfterDispose(t *testing.T) {
	connection := NewMemoryConnection("conn-1", 1)
	_ = connection.Dispose()
	if err := connection.Write(context.Background(), []byte{1}); err == nil {
		t.Fatalf("expected write failure after dispose")
	}
	if err := connection.PushInbound([]byte{1}); err == nil {
		t.Fatalf("expected inbound failure after dispose")
	}
}
