package transport

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"pixelsv/internal/auth/adapters/memory"
	"pixelsv/internal/auth/app"
	authmessaging "pixelsv/internal/auth/messaging"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
	coretransport "pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/local"
	"pixelsv/pkg/protocol"
)

// TestSubscriberSSOTicketFlow validates authenticated publish and auth-ok output.
func TestSubscriberSSOTicketFlow(t *testing.T) {
	store := memory.NewTicketStore()
	service := app.NewService(store, nil)
	ticket, _, err := service.CreateTicket(33, 60)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	bus := local.New()
	subscriber := NewSubscriber(bus, service, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := subscriber.Start(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	authenticated := make(chan []byte, 1)
	output := make(chan []byte, 1)
	_, _ = bus.Subscribe(ctx, sessionmessaging.TopicAuthenticated, func(_ context.Context, message coretransport.Message) error {
		authenticated <- message.Payload
		return nil
	})
	_, _ = bus.Subscribe(ctx, sessionmessaging.OutputTopic("s1"), func(_ context.Context, message coretransport.Message) error {
		output <- message.Payload
		return nil
	})
	release := &protocol.HandshakeReleaseVersionPacket{ReleaseVersion: "NITRO-1-6-6", ClientType: "HTML5", Platform: 2, DeviceCategory: 1}
	if err := bus.Publish(ctx, authmessaging.PacketIngressTopic("s1"), encodeBody(t, release)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	packet := &protocol.SecuritySsoTicketPacket{Ticket: ticket}
	if err := bus.Publish(ctx, authmessaging.PacketIngressTopic("s1"), encodeBody(t, packet)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	select {
	case value := <-authenticated:
		reader := codec.NewReader(value)
		sessionID, _ := reader.ReadString()
		userID, _ := reader.ReadInt32()
		if sessionID != "s1" || userID != 33 {
			t.Fatalf("unexpected auth payload")
		}
	case <-time.After(time.Second):
		t.Fatalf("expected authenticated event")
	}
	select {
	case value := <-output:
		frames, err := codec.SplitFrames(value)
		if err != nil || len(frames) != 2 || frames[0].Header != 2491 || frames[1].Header != 3523 {
			t.Fatalf("unexpected output payload")
		}
	case <-time.After(time.Second):
		t.Fatalf("expected session output")
	}
}

// TestSubscriberInvalidTicket validates invalid tickets do not publish auth events.
func TestSubscriberInvalidTicket(t *testing.T) {
	service := app.NewService(memory.NewTicketStore(), nil)
	bus := local.New()
	subscriber := NewSubscriber(bus, service, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := subscriber.Start(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	authenticated := make(chan []byte, 1)
	disconnect := make(chan []byte, 1)
	_, _ = bus.Subscribe(ctx, sessionmessaging.TopicAuthenticated, func(_ context.Context, message coretransport.Message) error {
		authenticated <- message.Payload
		return nil
	})
	_, _ = bus.Subscribe(ctx, sessionmessaging.DisconnectTopic("s1"), func(_ context.Context, message coretransport.Message) error {
		disconnect <- message.Payload
		return nil
	})
	release := &protocol.HandshakeReleaseVersionPacket{ReleaseVersion: "NITRO-1-6-6", ClientType: "HTML5", Platform: 2, DeviceCategory: 1}
	if err := bus.Publish(ctx, authmessaging.PacketIngressTopic("s1"), encodeBody(t, release)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	packet := &protocol.SecuritySsoTicketPacket{Ticket: "missing"}
	if err := bus.Publish(ctx, authmessaging.PacketIngressTopic("s1"), encodeBody(t, packet)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	select {
	case <-authenticated:
		t.Fatalf("did not expect authenticated event")
	case <-time.After(100 * time.Millisecond):
	}
	select {
	case <-disconnect:
	case <-time.After(time.Second):
		t.Fatalf("expected disconnect control event")
	}
	expired := service.ExpireUnauthenticatedSessions(time.Now().Add(24 * time.Hour))
	if len(expired) != 0 {
		t.Fatalf("expected session cleanup on reject, got expired=%v", expired)
	}
}

// encodeBody encodes one packet into subscriber message body format.
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
