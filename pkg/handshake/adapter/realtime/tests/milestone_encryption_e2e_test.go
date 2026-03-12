package tests

import (
	"crypto/rc4"
	"crypto/sha1"
	"math/big"
	"testing"
	"time"

	gws "github.com/gorilla/websocket"
	"github.com/momlesstomato/pixel-server/core/codec"
	packetcrypto "github.com/momlesstomato/pixel-server/pkg/handshake/packet/crypto"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
)

// TestMilestone7EncryptedAuthenticationFlow verifies diffie handshake and encrypted auth packet exchange.
func TestMilestone7EncryptedAuthenticationFlow(t *testing.T) {
	handler, cleanup := createHandler(t, map[string]int{"ticket-enc": 11}, 500*time.Millisecond, 100*time.Millisecond, 300*time.Millisecond)
	defer cleanup()
	connection, closeConnection := startWebSocket(t, handler.Handle)
	defer closeConnection()
	sendPacket(t, connection, packetcrypto.ClientInitDiffiePacket{})
	initFrame := readFrameByID(t, connection, packetcrypto.ServerInitDiffiePacketID)
	initPacket := packetcrypto.ServerInitDiffiePacket{}
	if err := initPacket.Decode(initFrame.Body); err != nil {
		t.Fatalf("expected init_diffie decode success, got %v", err)
	}
	prime, primeOK := new(big.Int).SetString(initPacket.EncryptedPrime, 10)
	generator, generatorOK := new(big.Int).SetString(initPacket.EncryptedGenerator, 10)
	if !primeOK || !generatorOK {
		t.Fatalf("expected decimal prime and generator")
	}
	clientPrivate := big.NewInt(7)
	clientPublic := new(big.Int).Exp(generator, clientPrivate, prime)
	sendPacket(t, connection, packetcrypto.ClientCompleteDiffiePacket{EncryptedPublicKey: clientPublic.String()})
	completeFrame := readFrameByID(t, connection, packetcrypto.ServerCompleteDiffiePacketID)
	completePacket := packetcrypto.ServerCompleteDiffiePacket{}
	if err := completePacket.Decode(completeFrame.Body); err != nil {
		t.Fatalf("expected complete_diffie decode success, got %v", err)
	}
	serverPublic, publicOK := new(big.Int).SetString(completePacket.EncryptedPublicKey, 10)
	if !publicOK {
		t.Fatalf("expected decimal server public key")
	}
	shared := new(big.Int).Exp(serverPublic, clientPrivate, prime).Bytes()
	if len(shared) == 0 {
		shared = []byte{0}
	}
	digest := sha1.Sum(shared)
	toServer, err := rc4.NewCipher(digest[:])
	if err != nil {
		t.Fatalf("expected rc4 client->server creation success, got %v", err)
	}
	fromServer, err := rc4.NewCipher(digest[:])
	if err != nil {
		t.Fatalf("expected rc4 server->client creation success, got %v", err)
	}
	sso := packetsecurity.SSOTicketPacket{Ticket: "ticket-enc"}
	writeEncryptedPacket(t, connection, toServer, sso.PacketID(), mustEncodePacket(t, sso))
	first := readEncryptedFrame(t, connection, fromServer)
	second := readEncryptedFrame(t, connection, fromServer)
	if first.PacketID != 2491 || second.PacketID != 3523 {
		t.Fatalf("expected encrypted auth packets 2491/3523, got %d/%d", first.PacketID, second.PacketID)
	}
}

// mustEncodePacket serializes one packet body for test writes.
func mustEncodePacket(t *testing.T, packet interface{ Encode() ([]byte, error) }) []byte {
	t.Helper()
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected packet encode success, got %v", err)
	}
	return body
}

// writeEncryptedPacket writes one encrypted protocol frame into websocket.
func writeEncryptedPacket(t *testing.T, connection *gws.Conn, cipher *rc4.Cipher, packetID uint16, body []byte) {
	t.Helper()
	frame := codec.EncodeFrame(packetID, body)
	encrypted := make([]byte, len(frame))
	cipher.XORKeyStream(encrypted, frame)
	if err := connection.WriteMessage(gws.BinaryMessage, encrypted); err != nil {
		t.Fatalf("expected encrypted websocket write success, got %v", err)
	}
}

// readEncryptedFrame reads one encrypted websocket frame and decrypts it.
func readEncryptedFrame(t *testing.T, connection *gws.Conn, cipher *rc4.Cipher) codec.Frame {
	t.Helper()
	connection.SetReadDeadline(time.Now().Add(time.Second))
	_, payload, err := connection.ReadMessage()
	if err != nil {
		t.Fatalf("expected encrypted websocket read success, got %v", err)
	}
	decrypted := make([]byte, len(payload))
	cipher.XORKeyStream(decrypted, payload)
	frame, _, err := codec.DecodeFrame(decrypted)
	if err != nil {
		t.Fatalf("expected decrypted frame decode success, got %v", err)
	}
	return frame
}
