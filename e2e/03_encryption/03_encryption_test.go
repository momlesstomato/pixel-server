package encryption

import (
	"context"
	"crypto/rc4"
	"crypto/sha1"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	gws "github.com/gorilla/websocket"
	"github.com/momlesstomato/pixel-server/core/codec"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/e2e/testkit"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	packetcrypto "github.com/momlesstomato/pixel-server/pkg/handshake/packet/crypto"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	redislib "github.com/redis/go-redis/v9"
)

// validatorMap defines deterministic ticket validation behavior.
type validatorMap struct {
	// values maps ticket values to user identifiers.
	values map[string]int
}

// Validate resolves ticket values using map lookup.
func (validator validatorMap) Validate(_ context.Context, ticket string) (int, error) {
	userID, found := validator.values[ticket]
	if !found {
		return 0, errors.New("invalid ticket")
	}
	return userID, nil
}

// Test03EncryptionHandshakeAuthenticatesWithEncryptedPackets verifies encrypted authentication flow.
func Test03EncryptionHandshakeAuthenticatesWithEncryptedPackets(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup success, got %v", err)
	}
	defer redisServer.Close()
	client := redislib.NewClient(&redislib.Options{Addr: redisServer.Addr()})
	defer client.Close()
	bus, _ := handshakerealtime.NewRedisCloseSignalBus(client, "handshake:test")
	registry, _ := coreconnection.NewRedisSessionRegistry(client)
	handler, _ := handshakerealtime.NewHandler(validatorMap{values: map[string]int{"ticket-enc": 11}}, registry, packetsecurity.NewMachineIDPolicy(strings.NewReader(strings.Repeat("d", 32))), bus, nil, 2*time.Second)
	connection, cleanup := testkit.StartWebSocket(t, handler.Handle)
	defer cleanup()
	testkit.SendPacket(t, connection, packetcrypto.ClientInitDiffiePacket{})
	initFrame := testkit.ReadFrameByID(t, connection, packetcrypto.ServerInitDiffiePacketID)
	initPacket := packetcrypto.ServerInitDiffiePacket{}
	if err := initPacket.Decode(initFrame.Body); err != nil {
		t.Fatalf("expected init_diffie decode success, got %v", err)
	}
	prime, _ := new(big.Int).SetString(initPacket.EncryptedPrime, 10)
	generator, _ := new(big.Int).SetString(initPacket.EncryptedGenerator, 10)
	clientPrivate := big.NewInt(7)
	clientPublic := new(big.Int).Exp(generator, clientPrivate, prime)
	testkit.SendPacket(t, connection, packetcrypto.ClientCompleteDiffiePacket{EncryptedPublicKey: clientPublic.String()})
	completeFrame := testkit.ReadFrameByID(t, connection, packetcrypto.ServerCompleteDiffiePacketID)
	completePacket := packetcrypto.ServerCompleteDiffiePacket{}
	_ = completePacket.Decode(completeFrame.Body)
	serverPublic, _ := new(big.Int).SetString(completePacket.EncryptedPublicKey, 10)
	shared := new(big.Int).Exp(serverPublic, clientPrivate, prime).Bytes()
	if len(shared) == 0 {
		shared = []byte{0}
	}
	digest := sha1.Sum(shared)
	toServer, _ := rc4.NewCipher(digest[:])
	fromServer, _ := rc4.NewCipher(digest[:])
	sso := packetsecurity.SSOTicketPacket{Ticket: "ticket-enc"}
	writeEncryptedPacket(t, connection, toServer, sso.PacketID(), mustEncodePacket(t, sso))
	first := readEncryptedFrame(t, connection, fromServer)
	second := readEncryptedFrame(t, connection, fromServer)
	if first.PacketID != 2491 || second.PacketID != 3523 {
		t.Fatalf("expected encrypted auth packets 2491/3523, got %d/%d", first.PacketID, second.PacketID)
	}
}

// mustEncodePacket serializes one packet body for encrypted writes.
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
