package bootstrap

import "testing"

// TestReleaseVersionEncodeDecode verifies release_version packet round-trip behavior.
func TestReleaseVersionEncodeDecode(t *testing.T) {
	source := ReleaseVersionPacket{ReleaseVersion: "NITRO-1-6-6", ClientType: "HTML5", Platform: 1, DeviceCategory: 2}
	encoded, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ReleaseVersionPacket{}
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.ReleaseVersion != source.ReleaseVersion || decoded.ClientType != source.ClientType {
		t.Fatalf("unexpected decode result: %+v", decoded)
	}
}
