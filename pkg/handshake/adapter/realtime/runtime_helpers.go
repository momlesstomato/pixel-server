package realtime

import (
	"context"

	"github.com/gofiber/contrib/websocket"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/sessionflow"
	packetcrypto "github.com/momlesstomato/pixel-server/pkg/handshake/packet/crypto"
)

// abortConnection closes one websocket connection after startup failure.
func (handler *Handler) abortConnection(connection *websocket.Conn) {
	_ = connection.Close()
}

// disposeConnection applies one shared connection cleanup lifecycle.
func (handler *Handler) disposeConnection(cancel context.CancelFunc, disposable coreconnection.Disposable, disconnect *sessionflow.DisconnectUseCase, heartbeatStop func(), connID string, connection *websocket.Conn) {
	heartbeatStop()
	_ = disposable.Dispose()
	disconnect.Cleanup(connID)
	cancel()
	_ = connection.Close()
}

// startHeartbeat starts one heartbeat loop and returns one stop function.
func (handler *Handler) startHeartbeat(ctx context.Context, connID string, useCase *sessionflow.HeartbeatUseCase, pongSignal <-chan struct{}, cancel context.CancelFunc) func() {
	heartbeatCtx, stop := context.WithCancel(ctx)
	go func() {
		if useCase.Run(heartbeatCtx, connID, pongSignal) != nil {
			cancel()
		}
	}()
	return stop
}

// handleInitDiffie handles one init_diffie packet and sends server parameters.
func (handler *Handler) handleInitDiffie(connID string, body []byte, transport *Transport, useCases *runtimeUseCases) {
	request := packetcrypto.ClientInitDiffiePacket{}
	if request.Decode(body) != nil {
		return
	}
	response, err := useCases.crypto.Begin()
	if err != nil {
		return
	}
	packet := packetcrypto.ServerInitDiffiePacket{EncryptedPrime: response.EncryptedPrime, EncryptedGenerator: response.EncryptedGenerator}
	if encoded, encodeErr := packet.Encode(); encodeErr == nil {
		_ = transport.Send(connID, packet.PacketID(), encoded)
	}
}

// handleCompleteDiffie handles one complete_diffie packet and enables stream encryption.
func (handler *Handler) handleCompleteDiffie(connID string, body []byte, transport *Transport, useCases *runtimeUseCases) {
	request := packetcrypto.ClientCompleteDiffiePacket{}
	if request.Decode(body) != nil {
		return
	}
	response, cipher, err := useCases.crypto.Complete(request.EncryptedPublicKey)
	if err != nil {
		return
	}
	packet := packetcrypto.ServerCompleteDiffiePacket{EncryptedPublicKey: response.EncryptedPublicKey, ServerClientEncryption: response.ServerClientEncryption}
	if encoded, encodeErr := packet.Encode(); encodeErr == nil {
		_ = transport.Send(connID, packet.PacketID(), encoded)
		transport.SetCipher(cipher)
	}
}
