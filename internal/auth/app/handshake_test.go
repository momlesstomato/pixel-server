package app

import (
	"testing"
	"time"

	"pixelsv/internal/auth/adapters/memory"
	"pixelsv/pkg/protocol"
)

// TestRecordReleaseVersionValidation validates release-version requirements.
func TestRecordReleaseVersionValidation(t *testing.T) {
	service := NewService(memory.NewTicketStore(), nil)
	err := service.RecordReleaseVersion("s1", &protocol.HandshakeReleaseVersionPacket{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if err := service.RecordReleaseVersion("s1", &protocol.HandshakeReleaseVersionPacket{ReleaseVersion: "NITRO-1-6-6"}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestDiffieFlow validates init and complete diffie handshake behavior.
func TestDiffieFlow(t *testing.T) {
	service := NewService(memory.NewTicketStore(), nil)
	initResponse, err := service.InitDiffie("s1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if initResponse.SignedPrime == "" || initResponse.SignedGenerator == "" {
		t.Fatalf("expected non-empty init response")
	}
	complete, err := service.CompleteDiffie("s1", "7")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if complete.PublicKey == "" {
		t.Fatalf("expected non-empty complete response")
	}
}

// TestMachineIDNormalization validates invalid machine-id replacement behavior.
func TestMachineIDNormalization(t *testing.T) {
	service := NewService(memory.NewTicketStore(), nil)
	normalized, changed, err := service.UpdateMachineID("s1", "~bad", "fp", "cap")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !changed {
		t.Fatalf("expected normalized machine id")
	}
	if len(normalized) != 64 {
		t.Fatalf("expected 64-char id, got %d", len(normalized))
	}
}

// TestExpireUnauthenticatedSessions validates timeout-based state eviction.
func TestExpireUnauthenticatedSessions(t *testing.T) {
	service := NewService(memory.NewTicketStore(), nil)
	service.timeout = 10 * time.Millisecond
	if _, err := service.touchSession("s1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	expired := service.ExpireUnauthenticatedSessions(time.Now())
	if len(expired) != 1 || expired[0] != "s1" {
		t.Fatalf("unexpected expired sessions: %v", expired)
	}
}
