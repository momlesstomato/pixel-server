package navigation

import (
	"context"
	"errors"
	"testing"

	packetsnavigation "github.com/momlesstomato/pixel-server/pkg/session/packet/navigation"
)

// TestNewDesktopViewUseCaseRejectsMissingDependencies verifies constructor validation behavior.
func TestNewDesktopViewUseCaseRejectsMissingDependencies(t *testing.T) {
	if _, err := NewDesktopViewUseCase(nil, nil); err == nil {
		t.Fatalf("expected transport validation failure")
	}
}

// TestDesktopViewUseCaseRunSendsDesktopViewWhenInRoom verifies response send behavior.
func TestDesktopViewUseCaseRunSendsDesktopViewWhenInRoom(t *testing.T) {
	transport := &transportStub{}
	useCase, _ := NewDesktopViewUseCase(transport, checkerStub{inRoom: true})
	if err := useCase.Run(context.Background(), "conn-1", 7); err != nil {
		t.Fatalf("expected run success, got %v", err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != packetsnavigation.DesktopViewResponsePacketID {
		t.Fatalf("unexpected sent packets %v", transport.sent)
	}
}

// TestDesktopViewUseCaseRunSkipsWhenNotInRoom verifies no-op behavior outside rooms.
func TestDesktopViewUseCaseRunSkipsWhenNotInRoom(t *testing.T) {
	transport := &transportStub{}
	useCase, _ := NewDesktopViewUseCase(transport, checkerStub{inRoom: false})
	if err := useCase.Run(context.Background(), "conn-1", 7); err != nil {
		t.Fatalf("expected run success, got %v", err)
	}
	if len(transport.sent) != 0 {
		t.Fatalf("expected no packet send, got %v", transport.sent)
	}
}

// TestDesktopViewUseCaseRunHandlesValidationAndCheckerErrors verifies error behavior.
func TestDesktopViewUseCaseRunHandlesValidationAndCheckerErrors(t *testing.T) {
	useCase, _ := NewDesktopViewUseCase(&transportStub{}, checkerStub{err: errors.New("boom")})
	if err := useCase.Run(context.Background(), "conn-1", 0); err == nil {
		t.Fatalf("expected user id validation failure")
	}
	if err := useCase.Run(context.Background(), "conn-1", 7); err == nil {
		t.Fatalf("expected checker error propagation")
	}
}

// transportStub captures sent packet identifiers.
type transportStub struct{ sent []uint16 }

// Send records sent packet identifiers.
func (stub *transportStub) Send(_ string, packetID uint16, _ []byte) error {
	stub.sent = append(stub.sent, packetID)
	return nil
}

// checkerStub provides deterministic room presence behavior.
type checkerStub struct {
	// inRoom stores deterministic room presence marker.
	inRoom bool
	// err stores deterministic room check failure.
	err error
}

// IsInRoom returns deterministic room presence output.
func (stub checkerStub) IsInRoom(context.Context, int) (bool, error) {
	return stub.inRoom, stub.err
}
