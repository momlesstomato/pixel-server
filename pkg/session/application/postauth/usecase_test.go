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
	if _, err := NewUseCase(nil, statusStub{}, loginStub{}, profileStub{}, accessStub{}, "pixel-server"); err == nil {
		t.Fatalf("expected transport validation failure")
	}
	if _, err := NewUseCase(&transportStub{}, nil, loginStub{}, profileStub{}, accessStub{}, "pixel-server"); err == nil {
		t.Fatalf("expected status validation failure")
	}
	if _, err := NewUseCase(&transportStub{}, statusStub{}, nil, profileStub{}, accessStub{}, "pixel-server"); err == nil {
		t.Fatalf("expected login recorder validation failure")
	}
	if _, err := NewUseCase(&transportStub{}, statusStub{}, loginStub{}, nil, accessStub{}, "pixel-server"); err == nil {
		t.Fatalf("expected profile reader validation failure")
	}
	if _, err := NewUseCase(&transportStub{}, statusStub{}, loginStub{}, profileStub{}, nil, "pixel-server"); err == nil {
		t.Fatalf("expected access reader validation failure")
	}
}

// TestUseCaseRunSendsAvailabilityFirstLoginAndPing verifies post-auth packet sequence.
func TestUseCaseRunSendsAvailabilityFirstLoginAndPing(t *testing.T) {
	transport := &transportStub{}
	useCase, _ := NewUseCase(transport, statusStub{status: statusdomain.HotelStatus{State: statusdomain.StateOpen}}, loginStub{first: true}, profileStub{}, accessStub{}, "pixel-server")
	useCase.now = func() time.Time { return time.Date(2026, time.March, 12, 12, 0, 0, 0, time.UTC) }
	if err := useCase.Run(context.Background(), "conn-1", 7); err != nil {
		t.Fatalf("expected run success, got %v", err)
	}
	expected := []uint16{2033, 2725, 411, 2586, 3738, 513, 2875, 126, 793, 3928}
	if len(transport.sent) != len(expected) || !equalIDs(transport.sent, expected) {
		t.Fatalf("unexpected packet sequence %v", transport.sent)
	}
}

// TestUseCaseRunSkipsFirstLoginPacketWhenNotFirst verifies optional first-login packet behavior.
func TestUseCaseRunSkipsFirstLoginPacketWhenNotFirst(t *testing.T) {
	transport := &transportStub{}
	useCase, _ := NewUseCase(transport, statusStub{status: statusdomain.HotelStatus{State: statusdomain.StateOpen}}, loginStub{first: false}, profileStub{}, accessStub{}, "pixel-server")
	if err := useCase.Run(context.Background(), "conn-1", 7); err != nil {
		t.Fatalf("expected run success, got %v", err)
	}
	expected := []uint16{2033, 2725, 411, 2586, 3738, 513, 2875, 126, 3928}
	if len(transport.sent) != len(expected) || !equalIDs(transport.sent, expected) {
		t.Fatalf("unexpected packet sequence %v", transport.sent)
	}
}

// TestUseCaseRunSendsHotelClosedDisconnect verifies closed hotel disconnect behavior.
func TestUseCaseRunSendsHotelClosedDisconnect(t *testing.T) {
	transport := &transportStub{}
	useCase, _ := NewUseCase(transport, statusStub{status: statusdomain.HotelStatus{State: statusdomain.StateClosed}}, loginStub{first: true}, profileStub{}, accessStub{}, "pixel-server")
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
