package bootstrap

import "testing"

// TestClientVariablesEncodeDecode verifies client_variables packet round-trip behavior.
func TestClientVariablesEncodeDecode(t *testing.T) {
	source := ClientVariablesPacket{ClientID: 7, ClientURL: "client", ExternalVariablesURL: "external"}
	encoded, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ClientVariablesPacket{}
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.ClientID != source.ClientID || decoded.ClientURL != source.ClientURL {
		t.Fatalf("unexpected decode result: %+v", decoded)
	}
}
