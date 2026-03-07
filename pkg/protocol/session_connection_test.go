package protocol

import (
	"testing"

	"pixelsv/pkg/codec"
)

// TestSessionConnectionRoundTrip validates generated session-connection packet codecs.
func TestSessionConnectionRoundTrip(t *testing.T) {
	tests := []struct {
		header uint16
		packet Packet
	}{
		{header: HeaderClientLatencyTestPacket, packet: &ClientLatencyTestPacket{RequestId: 22}},
		{header: HeaderClientPongPacket, packet: &ClientPongPacket{}},
		{header: HeaderClientDisconnectPacket, packet: &ClientDisconnectPacket{}},
		{header: HeaderSessionDesktopViewPacket, packet: &SessionDesktopViewPacket{}},
		{header: HeaderSessionPeerUsersClassificationPacket, packet: &SessionPeerUsersClassificationPacket{}},
		{header: HeaderSessionClientToolbarTogglePacket, packet: &SessionClientToolbarTogglePacket{}},
		{header: HeaderSessionRenderRoomPacket, packet: &SessionRenderRoomPacket{}},
		{header: HeaderSessionTrackingPerformanceLogPacket, packet: &SessionTrackingPerformanceLogPacket{}},
		{header: HeaderSessionEventTrackerPacket, packet: &SessionEventTrackerPacket{}},
		{header: HeaderSessionTrackingLagWarningReportPacket, packet: &SessionTrackingLagWarningReportPacket{}},
	}
	for _, tt := range tests {
		writer := codec.NewWriter(64)
		if err := tt.packet.Encode(writer); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		decoded, err := DecodeC2S(tt.header, writer.Bytes())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if decoded.HeaderID() != tt.header || decoded.Realm() != "session-connection" {
			t.Fatalf("unexpected decode result: %T", decoded)
		}
	}
}
