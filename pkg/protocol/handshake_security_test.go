package protocol

import (
	"testing"

	"pixelsv/pkg/codec"
)

// TestHandshakeSecurityRoundTrip validates generated handshake-security packet codecs.
func TestHandshakeSecurityRoundTrip(t *testing.T) {
	timestamp := int32(99)
	tests := []struct {
		name   string
		header uint16
		source Packet
		check  func(Packet) bool
	}{
		{name: "client latency measure", header: HeaderHandshakeClientLatencyMeasurePacket, source: &HandshakeClientLatencyMeasurePacket{}, check: func(packet Packet) bool { _, ok := packet.(*HandshakeClientLatencyMeasurePacket); return ok }},
		{name: "complete diffie", header: HeaderHandshakeCompleteDiffiePacket, source: &HandshakeCompleteDiffiePacket{EncryptedPublicKey: "abc"}, check: func(packet Packet) bool {
			value, ok := packet.(*HandshakeCompleteDiffiePacket)
			return ok && value.EncryptedPublicKey == "abc"
		}},
		{name: "client variables", header: HeaderHandshakeClientVariablesPacket, source: &HandshakeClientVariablesPacket{ClientId: 10, ClientUrl: "u", ExternalVariablesUrl: "v"}, check: func(packet Packet) bool {
			value, ok := packet.(*HandshakeClientVariablesPacket)
			return ok && value.ClientId == 10 && value.ClientUrl == "u" && value.ExternalVariablesUrl == "v"
		}},
		{name: "sso with timestamp", header: HeaderSecuritySsoTicketPacket, source: &SecuritySsoTicketPacket{Ticket: "t1", Timestamp: &timestamp}, check: func(packet Packet) bool {
			value, ok := packet.(*SecuritySsoTicketPacket)
			return ok && value.Ticket == "t1" && value.Timestamp != nil && *value.Timestamp == 99
		}},
		{name: "machine id", header: HeaderSecurityMachineIdPacket, source: &SecurityMachineIdPacket{MachineId: "m", Fingerprint: "f", Capabilities: "c"}, check: func(packet Packet) bool {
			value, ok := packet.(*SecurityMachineIdPacket)
			return ok && value.MachineId == "m" && value.Fingerprint == "f" && value.Capabilities == "c"
		}},
		{name: "init diffie", header: HeaderHandshakeInitDiffiePacket, source: &HandshakeInitDiffiePacket{}, check: func(packet Packet) bool { _, ok := packet.(*HandshakeInitDiffiePacket); return ok }},
		{name: "release version", header: HeaderHandshakeReleaseVersionPacket, source: &HandshakeReleaseVersionPacket{ReleaseVersion: "NITRO-1-6-6", ClientType: "HTML5", Platform: 2, DeviceCategory: 1}, check: func(packet Packet) bool {
			value, ok := packet.(*HandshakeReleaseVersionPacket)
			return ok && value.ReleaseVersion == "NITRO-1-6-6" && value.ClientType == "HTML5" && value.Platform == 2 && value.DeviceCategory == 1
		}},
		{name: "client policy", header: HeaderHandshakeClientPolicyPacket, source: &HandshakeClientPolicyPacket{}, check: func(packet Packet) bool { _, ok := packet.(*HandshakeClientPolicyPacket); return ok }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := codec.NewWriter(128)
			if err := tt.source.Encode(writer); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			packet, err := DecodeC2S(tt.header, writer.Bytes())
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !tt.check(packet) {
				t.Fatalf("unexpected packet: %#v", packet)
			}
		})
	}
}

// TestSecuritySsoTicketOptionalTimestamp validates optional field decode behavior.
func TestSecuritySsoTicketOptionalTimestamp(t *testing.T) {
	writer := codec.NewWriter(32)
	packet := SecuritySsoTicketPacket{Ticket: "required-only"}
	if err := packet.Encode(writer); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	decoded, err := DecodeC2S(HeaderSecuritySsoTicketPacket, writer.Bytes())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	value, ok := decoded.(*SecuritySsoTicketPacket)
	if !ok {
		t.Fatalf("unexpected packet type: %T", decoded)
	}
	if value.Ticket != "required-only" || value.Timestamp != nil {
		t.Fatalf("unexpected packet value: %#v", value)
	}
}
