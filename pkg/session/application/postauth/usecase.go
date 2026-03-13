package postauth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
	packetsession "github.com/momlesstomato/pixel-server/pkg/handshake/packet/session"
	packetsavailability "github.com/momlesstomato/pixel-server/pkg/session/packet/availability"
	packetsnavigation "github.com/momlesstomato/pixel-server/pkg/session/packet/navigation"
	statusdomain "github.com/momlesstomato/pixel-server/pkg/status/domain"
)

// ErrHotelClosed defines hotel-closed authentication continuation behavior.
var ErrHotelClosed = errors.New("hotel is closed")

// Transport defines packet transport behavior.
type Transport interface {
	// Send writes one encoded packet payload to one connection.
	Send(string, uint16, []byte) error
	// Close closes one connection with close code and reason payload.
	Close(string, int, string) error
}

// StatusReader defines hotel status read behavior.
type StatusReader interface {
	// Current returns one active hotel status snapshot.
	Current(context.Context) (statusdomain.HotelStatus, error)
}

// LoginRecorder defines successful login stamp behavior.
type LoginRecorder interface {
	// RecordLogin persists one successful login event and returns first-login-of-day marker.
	RecordLogin(context.Context, int, string, time.Time) (bool, error)
}

// UseCase defines post-authentication packet burst behavior.
type UseCase struct {
	// transport sends packets to active connection.
	transport Transport
	// status reads current hotel status.
	status StatusReader
	// logins records successful login stamps.
	logins LoginRecorder
	// holder stores holder identifier for stamped login records.
	holder string
	// now provides deterministic timestamp source for tests.
	now func() time.Time
}

// NewUseCase creates one post-authentication burst use case.
func NewUseCase(transport Transport, status StatusReader, logins LoginRecorder, holder string) (*UseCase, error) {
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	if status == nil {
		return nil, fmt.Errorf("status reader is required")
	}
	if logins == nil {
		return nil, fmt.Errorf("login recorder is required")
	}
	resolvedHolder := strings.TrimSpace(holder)
	if resolvedHolder == "" {
		resolvedHolder = "pixel-server"
	}
	return &UseCase{transport: transport, status: status, logins: logins, holder: resolvedHolder, now: time.Now}, nil
}

// Run sends availability status, optional first-login-of-day, and immediate ping packet.
func (useCase *UseCase) Run(ctx context.Context, connID string, userID int) error {
	status, err := useCase.status.Current(ctx)
	if err != nil {
		return err
	}
	if !status.IsOpen() {
		disconnect := packetauth.DisconnectReasonPacket{Reason: packetauth.DisconnectReasonHotelClosed}
		if err := sendPacket(useCase.transport, connID, disconnect.PacketID(), disconnect); err != nil {
			return err
		}
		if err := useCase.transport.Close(connID, websocket.ClosePolicyViolation, "hotel closed"); err != nil {
			return err
		}
		return ErrHotelClosed
	}
	availability := packetsavailability.StatusPacket{IsOpen: status.IsOpen(), OnShutdown: status.OnShutdown(), IsAuthentic: true}
	if err := sendPacket(useCase.transport, connID, availability.PacketID(), availability); err != nil {
		return err
	}
	firstLoginOfDay, err := useCase.logins.RecordLogin(ctx, userID, useCase.holder, useCase.now().UTC())
	if err != nil {
		return err
	}
	if firstLoginOfDay {
		firstLogin := packetsnavigation.FirstLoginOfDayPacket{IsFirstLogin: true}
		if err := sendPacket(useCase.transport, connID, firstLogin.PacketID(), firstLogin); err != nil {
			return err
		}
	}
	ping := packetsession.ClientPingPacket{}
	return sendPacket(useCase.transport, connID, ping.PacketID(), ping)
}

// sendPacket encodes and writes one packet payload.
func sendPacket(transport Transport, connID string, packetID uint16, packet interface{ Encode() ([]byte, error) }) error {
	body, err := packet.Encode()
	if err != nil {
		return err
	}
	return transport.Send(connID, packetID, body)
}
