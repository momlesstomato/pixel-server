package transport

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	sessionapp "pixelsv/internal/sessionconnection/app"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
	"pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/local"
	"pixelsv/pkg/plugin"
	"pixelsv/pkg/plugin/eventbus"
	"pixelsv/pkg/protocol"
)

// TestSubscriberAuthenticatedFlow validates availability output and concurrent-login disconnect.
func TestSubscriberAuthenticatedFlow(t *testing.T) {
	bus := local.New()
	service := sessionapp.NewService(nil, time.Second)
	subscriber := NewSubscriber(bus, service, nil, Config{PingInterval: time.Hour, PongTimeout: 2 * time.Hour, AvailabilityOpen: true, AvailabilityOnShutdown: false, AvailabilityAuthentic: true})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := subscriber.Start(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	disconnects := make(chan transport.Message, 1)
	outputs := make(chan transport.Message, 4)
	_, _ = bus.Subscribe(ctx, sessionmessaging.DisconnectTopic("s1"), func(_ context.Context, message transport.Message) error { disconnects <- message; return nil })
	_, _ = bus.Subscribe(ctx, sessionmessaging.OutputTopic("s1"), func(_ context.Context, message transport.Message) error { outputs <- message; return nil })
	_, _ = bus.Subscribe(ctx, sessionmessaging.OutputTopic("s2"), func(_ context.Context, message transport.Message) error { outputs <- message; return nil })
	_ = bus.Publish(ctx, sessionmessaging.TopicConnected, []byte("s1"))
	_ = bus.Publish(ctx, sessionmessaging.TopicConnected, []byte("s2"))
	_ = bus.Publish(ctx, sessionmessaging.TopicAuthenticated, sessionmessaging.EncodeAuthenticatedEvent("s1", 7))
	_ = bus.Publish(ctx, sessionmessaging.TopicAuthenticated, sessionmessaging.EncodeAuthenticatedEvent("s2", 7))
	waitHeaders(t, outputs, map[uint16]int{2033: 2, 4000: 1})
	select {
	case message := <-disconnects:
		if binary.BigEndian.Uint32(message.Payload[:4]) != uint32(sessionmessaging.DisconnectReasonConcurrentLogin) {
			t.Fatalf("expected reason 2")
		}
	case <-time.After(time.Second):
		t.Fatalf("expected disconnect event")
	}
}

// TestSubscriberPacketFlow validates core packet handlers.
func TestSubscriberPacketFlow(t *testing.T) {
	bus := local.New()
	events := eventbus.New()
	service := sessionapp.NewService(events, time.Second)
	subscriber := NewSubscriber(bus, service, nil, Config{PingInterval: 25 * time.Millisecond, PongTimeout: time.Second, AvailabilityOpen: true, AvailabilityOnShutdown: false, AvailabilityAuthentic: true})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := subscriber.Start(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	outputs := make(chan transport.Message, 8)
	disconnects := make(chan transport.Message, 1)
	packetEvents := make(chan plugin.Event, 4)
	_, _ = bus.Subscribe(ctx, sessionmessaging.OutputTopic("s1"), func(_ context.Context, message transport.Message) error { outputs <- message; return nil })
	_, _ = bus.Subscribe(ctx, sessionmessaging.DisconnectTopic("s1"), func(_ context.Context, message transport.Message) error { disconnects <- message; return nil })
	events.On(sessionmessaging.EventPacketReceived, func(event *plugin.Event) error { packetEvents <- *event; return nil })
	_ = bus.Publish(ctx, sessionmessaging.TopicConnected, []byte("s1"))
	_ = bus.Publish(ctx, sessionmessaging.TopicAuthenticated, sessionmessaging.EncodeAuthenticatedEvent("s1", 9))
	waitHeader(t, outputs, 2033)
	_ = bus.Publish(ctx, sessionmessaging.PacketIngressTopic("s1"), encodeBody(t, &protocol.ClientLatencyTestPacket{RequestId: 44}))
	waitHeader(t, outputs, 10)
	_ = bus.Publish(ctx, sessionmessaging.PacketIngressTopic("s1"), encodeBody(t, &protocol.SessionDesktopViewPacket{}))
	waitHeader(t, outputs, 122)
	waitPacketEvent(t, packetEvents, protocol.HeaderSessionDesktopViewPacket)
	_ = bus.Publish(ctx, sessionmessaging.PacketIngressTopic("s1"), encodeBody(t, &protocol.ClientDisconnectPacket{}))
	waitHeader(t, outputs, 4000)
	select {
	case <-disconnects:
	case <-time.After(time.Second):
		t.Fatalf("expected disconnect control event")
	}
	waitHeader(t, outputs, 3928)
}

// TestSubscriberCleanupAfterContextCancel validates disconnect cleanup after startup context cancel.
func TestSubscriberCleanupAfterContextCancel(t *testing.T) {
	bus := local.New()
	service := sessionapp.NewService(nil, time.Second)
	subscriber := NewSubscriber(bus, service, nil, Config{PingInterval: time.Hour, PongTimeout: 2 * time.Hour, AvailabilityOpen: true, AvailabilityOnShutdown: false, AvailabilityAuthentic: true})
	ctx, cancel := context.WithCancel(context.Background())
	if err := subscriber.Start(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_ = bus.Publish(ctx, sessionmessaging.TopicConnected, []byte("s1"))
	_ = bus.Publish(ctx, sessionmessaging.TopicAuthenticated, sessionmessaging.EncodeAuthenticatedEvent("s1", 9))
	cancel()
	_ = bus.Publish(context.Background(), sessionmessaging.TopicDisconnected, []byte("s1"))
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if len(service.ActiveAuthenticatedSessions()) == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected session cleanup after disconnect publish")
}

func waitPacketEvent(t *testing.T, ch <-chan plugin.Event, header uint16) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		select {
		case event := <-ch:
			payload, ok := event.Data.(sessionmessaging.PacketReceivedEventData)
			if ok && payload.Header == header {
				return
			}
		case <-time.After(10 * time.Millisecond):
		}
	}
	t.Fatalf("expected packet event")
}

func encodeBody(t *testing.T, packet protocol.Packet) []byte {
	t.Helper()
	writer := codec.NewWriter(64)
	if err := packet.Encode(writer); err != nil {
		t.Fatalf("encode packet failed: %v", err)
	}
	body := make([]byte, 2+len(writer.Bytes()))
	binary.BigEndian.PutUint16(body[:2], packet.HeaderID())
	copy(body[2:], writer.Bytes())
	return body
}

func waitHeader(t *testing.T, ch <-chan transport.Message, expected uint16) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		select {
		case message := <-ch:
			frames, err := codec.SplitFrames(message.Payload)
			if err != nil || len(frames) != 1 {
				t.Fatalf("unexpected frame payload")
			}
			if frames[0].Header == expected {
				return
			}
		case <-time.After(10 * time.Millisecond):
		}
	}
	t.Fatalf("expected output header %d", expected)
}

func waitHeaders(t *testing.T, ch <-chan transport.Message, expected map[uint16]int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	remaining := 0
	for _, count := range expected {
		remaining += count
	}
	for remaining > 0 && time.Now().Before(deadline) {
		select {
		case message := <-ch:
			frames, err := codec.SplitFrames(message.Payload)
			if err != nil || len(frames) != 1 {
				t.Fatalf("unexpected frame payload")
			}
			header := frames[0].Header
			if expected[header] > 0 {
				expected[header]--
				remaining--
			}
		case <-time.After(10 * time.Millisecond):
		}
	}
	if remaining == 0 {
		return
	}
	t.Fatalf("expected headers not received: %+v", expected)
}
