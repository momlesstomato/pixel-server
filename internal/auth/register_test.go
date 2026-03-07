package auth

import (
	"bytes"
	"context"
	"encoding/binary"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	authmessaging "pixelsv/internal/auth/messaging"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
	coretransport "pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/local"
	"pixelsv/pkg/plugin/eventbus"
	"pixelsv/pkg/protocol"
)

// TestRegisterHTTPRoutes validates auth route registration.
func TestRegisterHTTPRoutes(t *testing.T) {
	app := fiber.New()
	bus := local.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := Register(ctx, app, bus, eventbus.New(), nil, "secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	request := httptest.NewRequest("POST", "/api/v1/auth/tickets", bytes.NewReader([]byte(`{"user_id":1}`)))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", "secret")
	response, _ := app.Test(request)
	if response.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected created status, got %d", response.StatusCode)
	}
}

// TestRegisterTransportFlow validates register subscriber flow for sso ticket packets.
func TestRegisterTransportFlow(t *testing.T) {
	bus := local.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runtime, err := Register(ctx, nil, bus, eventbus.New(), nil, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	ticket, _, err := runtime.Service.CreateTicket(7, 60)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	authenticated := make(chan []byte, 1)
	_, _ = bus.Subscribe(ctx, sessionmessaging.TopicAuthenticated, func(_ context.Context, message coretransport.Message) error {
		authenticated <- message.Payload
		return nil
	})
	packet := &protocol.SecuritySsoTicketPacket{Ticket: ticket}
	release := &protocol.HandshakeReleaseVersionPacket{ReleaseVersion: "NITRO-1-6-6", ClientType: "HTML5", Platform: 2, DeviceCategory: 1}
	if err := bus.Publish(ctx, authmessaging.PacketIngressTopic("s1"), encodeBody(t, release)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := bus.Publish(ctx, authmessaging.PacketIngressTopic("s1"), encodeBody(t, packet)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	select {
	case payload := <-authenticated:
		reader := codec.NewReader(payload)
		sessionID, _ := reader.ReadString()
		userID, _ := reader.ReadInt32()
		if sessionID != "s1" || userID != 7 {
			t.Fatalf("unexpected auth payload")
		}
	case <-time.After(time.Second):
		t.Fatalf("expected auth event")
	}
}

// encodeBody encodes one packet into transport body format.
func encodeBody(t *testing.T, packet protocol.Packet) []byte {
	t.Helper()
	writer := codec.NewWriter(64)
	if err := packet.Encode(writer); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	payload := writer.Bytes()
	body := make([]byte, 2+len(payload))
	binary.BigEndian.PutUint16(body[:2], packet.HeaderID())
	copy(body[2:], payload)
	return body
}
