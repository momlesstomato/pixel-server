package protocol

import (
	"errors"
	"testing"

	"pixelsv/pkg/codec"
)

// TestDecodeC2SUnknownHeader validates error on missing packet decoder.
func TestDecodeC2SUnknownHeader(t *testing.T) {
	_, err := DecodeC2S(65535, nil)
	if !errors.Is(err, ErrUnknownHeader) {
		t.Fatalf("expected unknown header error, got %v", err)
	}
}

// TestDecodeC2SReleaseVersion validates handshake packet decode behavior.
func TestDecodeC2SReleaseVersion(t *testing.T) {
	writer := codec.NewWriter(64)
	source := HandshakeReleaseVersionPacket{
		ReleaseVersion: "NITRO-1-6-6",
		ClientType:     "HTML5",
		Platform:       2,
		DeviceCategory: 1,
	}
	if err := source.Encode(writer); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	packet, err := DecodeC2S(HeaderHandshakeReleaseVersionPacket, writer.Bytes())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	release, ok := packet.(*HandshakeReleaseVersionPacket)
	if !ok {
		t.Fatalf("unexpected packet type: %T", packet)
	}
	if release.ReleaseVersion != "NITRO-1-6-6" || release.ClientType != "HTML5" || release.Platform != 2 || release.DeviceCategory != 1 {
		t.Fatalf("unexpected packet value: %+v", release)
	}
}
