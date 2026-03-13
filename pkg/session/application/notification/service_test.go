package notification

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	"github.com/momlesstomato/pixel-server/core/codec"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
	packeterror "github.com/momlesstomato/pixel-server/pkg/session/packet/error"
	packetnotification "github.com/momlesstomato/pixel-server/pkg/session/packet/notification"
)

// TestNewServiceRejectsMissingDependencies verifies constructor validation behavior.
func TestNewServiceRejectsMissingDependencies(t *testing.T) {
	if _, err := NewService(nil); err == nil {
		t.Fatalf("expected broadcaster validation failure")
	}
}

// TestServiceSendsTargetedNotificationPackets verifies targeted user-channel publish behavior.
func TestServiceSendsTargetedNotificationPackets(t *testing.T) {
	broadcaster := broadcast.NewLocalBroadcaster()
	ctx := context.Background()
	stream, disposable, err := broadcaster.Subscribe(ctx, UserChannel(7))
	if err != nil {
		t.Fatalf("expected subscribe success, got %v", err)
	}
	defer disposable.Dispose()
	service, _ := NewService(broadcaster)
	if err := service.SendGenericAlert(ctx, 7, "hello"); err != nil {
		t.Fatalf("expected generic alert publish success, got %v", err)
	}
	frame := readFrame(t, stream)
	if frame.PacketID != packetnotification.GenericAlertPacketID {
		t.Fatalf("unexpected packet id %d", frame.PacketID)
	}
	alert := packetnotification.GenericAlertPacket{}
	if err := alert.Decode(frame.Body); err != nil || alert.Message != "hello" {
		t.Fatalf("unexpected alert payload %#v err=%v", alert, err)
	}
	if err := service.SendGenericError(ctx, 7, -400); err != nil {
		t.Fatalf("expected generic error publish success, got %v", err)
	}
	frame = readFrame(t, stream)
	if frame.PacketID != packetnotification.GenericErrorPacketID {
		t.Fatalf("unexpected packet id %d", frame.PacketID)
	}
}

// TestServiceSendsConnectionAndDisconnectReasons verifies disconnect and connection-error publish behavior.
func TestServiceSendsConnectionAndDisconnectReasons(t *testing.T) {
	broadcaster := broadcast.NewLocalBroadcaster()
	ctx := context.Background()
	stream, disposable, _ := broadcaster.Subscribe(ctx, UserChannel(8))
	defer disposable.Dispose()
	service, _ := NewService(broadcaster)
	if err := service.SendConnectionError(ctx, 8, 9000, 2); err != nil {
		t.Fatalf("expected connection error publish success, got %v", err)
	}
	frame := readFrame(t, stream)
	if frame.PacketID != packeterror.ConnectionErrorPacketID {
		t.Fatalf("unexpected packet id %d", frame.PacketID)
	}
	if err := service.SendJustBannedDisconnect(ctx, 8); err != nil {
		t.Fatalf("expected just banned disconnect publish success, got %v", err)
	}
	frame = readFrame(t, stream)
	reason := packetauth.DisconnectReasonPacket{}
	if frame.PacketID != packetauth.DisconnectReasonPacketID || reason.Decode(frame.Body) != nil || reason.Reason != packetauth.DisconnectReasonJustBanned {
		t.Fatalf("unexpected disconnect payload %#v", reason)
	}
	if err := service.SendStillBannedDisconnect(ctx, 8); err != nil {
		t.Fatalf("expected still banned disconnect publish success, got %v", err)
	}
	frame = readFrame(t, stream)
	if frame.PacketID != packetauth.DisconnectReasonPacketID || reason.Decode(frame.Body) != nil || reason.Reason != packetauth.DisconnectReasonStillBanned {
		t.Fatalf("unexpected disconnect payload %#v", reason)
	}
}

// readFrame decodes one packet frame from one message stream.
func readFrame(t *testing.T, stream <-chan []byte) codec.Frame {
	t.Helper()
	payload := <-stream
	frame, _, err := codec.DecodeFrame(payload)
	if err != nil {
		t.Fatalf("expected frame decode success, got %v", err)
	}
	return frame
}
