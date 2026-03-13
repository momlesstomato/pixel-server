package postauth

import (
	"context"
	"errors"
	"testing"
	"time"

	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
	statusdomain "github.com/momlesstomato/pixel-server/pkg/status/domain"
)

// TestNewUseCaseRejectsMissingDependencies verifies constructor validation behavior.
func TestNewUseCaseRejectsMissingDependencies(t *testing.T) {
	if _, err := NewUseCase(nil, statusStub{}, loginStub{}, "pixel-server"); err == nil {
		t.Fatalf("expected transport validation failure")
	}
	if _, err := NewUseCase(&transportStub{}, nil, loginStub{}, "pixel-server"); err == nil {
		t.Fatalf("expected status validation failure")
	}
	if _, err := NewUseCase(&transportStub{}, statusStub{}, nil, "pixel-server"); err == nil {
		t.Fatalf("expected login recorder validation failure")
	}
}

// TestUseCaseRunSendsAvailabilityFirstLoginAndPing verifies post-auth packet sequence.
func TestUseCaseRunSendsAvailabilityFirstLoginAndPing(t *testing.T) {
	transport := &transportStub{}
	useCase, _ := NewUseCase(transport, statusStub{status: statusdomain.HotelStatus{State: statusdomain.StateOpen}}, loginStub{first: true}, "pixel-server")
	useCase.now = func() time.Time { return time.Date(2026, time.March, 12, 12, 0, 0, 0, time.UTC) }
	if err := useCase.Run(context.Background(), "conn-1", 7); err != nil {
		t.Fatalf("expected run success, got %v", err)
	}
	if len(transport.sent) != 3 || transport.sent[0] != 2033 || transport.sent[1] != 793 || transport.sent[2] != 3928 {
		t.Fatalf("unexpected packet sequence %v", transport.sent)
	}
}

// TestUseCaseRunSkipsFirstLoginPacketWhenNotFirst verifies optional first-login packet behavior.
func TestUseCaseRunSkipsFirstLoginPacketWhenNotFirst(t *testing.T) {
	transport := &transportStub{}
	useCase, _ := NewUseCase(transport, statusStub{status: statusdomain.HotelStatus{State: statusdomain.StateOpen}}, loginStub{first: false}, "pixel-server")
	if err := useCase.Run(context.Background(), "conn-1", 7); err != nil {
		t.Fatalf("expected run success, got %v", err)
	}
	if len(transport.sent) != 2 || transport.sent[0] != 2033 || transport.sent[1] != 3928 {
		t.Fatalf("unexpected packet sequence %v", transport.sent)
	}
}

// TestUseCaseRunSendsHotelClosedDisconnect verifies closed hotel disconnect behavior.
func TestUseCaseRunSendsHotelClosedDisconnect(t *testing.T) {
	transport := &transportStub{}
	useCase, _ := NewUseCase(transport, statusStub{status: statusdomain.HotelStatus{State: statusdomain.StateClosed}}, loginStub{first: true}, "pixel-server")
	if err := useCase.Run(context.Background(), "conn-1", 7); !errors.Is(err, ErrHotelClosed) {
		t.Fatalf("expected closed-hotel error, got %v", err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != packetauth.DisconnectReasonPacketID {
		t.Fatalf("unexpected packet sequence %v", transport.sent)
	}
	if len(transport.closed) != 1 || transport.closed[0] != "conn-1" {
		t.Fatalf("expected connection close call, got %v", transport.closed)
	}
}

// transportStub captures sent packet identifiers.
type transportStub struct {
	// sent stores sent packet identifiers.
	sent []uint16
	// closed stores closed connection identifiers.
	closed []string
}

// Send records sent packet identifiers.
func (stub *transportStub) Send(_ string, packetID uint16, _ []byte) error {
	stub.sent = append(stub.sent, packetID)
	return nil
}

// Close records closed connection identifiers.
func (stub *transportStub) Close(connID string, _ int, _ string) error {
	stub.closed = append(stub.closed, connID)
	return nil
}

// statusStub provides deterministic status responses.
type statusStub struct{ status statusdomain.HotelStatus }

// Current returns deterministic hotel status.
func (stub statusStub) Current(context.Context) (statusdomain.HotelStatus, error) {
	return stub.status, nil
}

// loginStub provides deterministic login stamp responses.
type loginStub struct{ first bool }

// RecordLogin returns deterministic first-login marker.
func (stub loginStub) RecordLogin(context.Context, int, string, time.Time) (bool, error) {
	return stub.first, nil
}
