package realtime

import (
	"context"
	"time"

	"github.com/gofiber/contrib/websocket"
	sdk "github.com/momlesstomato/pixel-sdk"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/authflow"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/sessionflow"
	packetcrypto "github.com/momlesstomato/pixel-server/pkg/handshake/packet/crypto"
	packetdisconnect "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	packetsession "github.com/momlesstomato/pixel-server/pkg/handshake/packet/session"
)

const sessionLeaseRefreshInterval = 60 * time.Second
const protocolErrorUnknownPacket int32 = 1
const protocolErrorMalformedPacket int32 = 2
const protocolErrorWrongState int32 = 3
const protocolErrorLimitPerMinute = 10

// abortConnection closes one websocket connection after startup failure.
// A close frame is sent before closing so the client receives a clean disconnect signal.
func (handler *Handler) abortConnection(connection *websocket.Conn) {
	_ = connection.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, ""), time.Now().Add(time.Second))
	_ = connection.Close()
}

// disposeConnection applies one shared connection cleanup lifecycle.
func (handler *Handler) disposeConnection(cancel context.CancelFunc, disposables []coreconnection.Disposable, disconnect *sessionflow.DisconnectUseCase, heartbeatStop func(), connID string, connection *websocket.Conn, userRuntime UserRuntime) {
	heartbeatStop()
	if userRuntime != nil {
		userRuntime.Dispose(connID)
	}
	for _, disposable := range disposables {
		if disposable != nil {
			_ = disposable.Dispose()
		}
	}
	disconnect.Cleanup(connID)
	if handler.fire != nil {
		handler.fire(&sdk.ConnectionClosed{ConnID: connID})
	}
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
	go handler.refreshSessionLease(heartbeatCtx, connID, cancel)
	return stop
}

// refreshSessionLease keeps Redis-backed session keys alive while heartbeat is active.
func (handler *Handler) refreshSessionLease(ctx context.Context, connID string, cancel context.CancelFunc) {
	ticker := time.NewTicker(sessionLeaseRefreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if handler.sessions.Touch(connID) != nil {
				cancel()
				return
			}
		}
	}
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

// handleMachineID normalizes and echoes machine identifier payload.
func (handler *Handler) handleMachineID(connID string, body []byte, transport *Transport, machineID *string) {
	packet := packetauth.ClientMachineIDPacket{}
	if packet.Decode(body) != nil {
		return
	}
	normalized, err := handler.policy.Normalize(packet.MachineID)
	if err != nil {
		return
	}
	*machineID = normalized
	response := packetauth.ServerMachineIDPacket{MachineID: normalized}
	encoded, encodeErr := response.Encode()
	if encodeErr == nil {
		_ = transport.Send(connID, response.PacketID(), encoded)
	}
}

// handleSSO authenticates one SSO packet payload.
func (handler *Handler) handleSSO(ctx context.Context, connID string, body []byte, machineID string, useCase *authflow.AuthenticateUseCase) (int, bool) {
	packet := packetauth.SSOTicketPacket{}
	if packet.Decode(body) != nil {
		return 0, false
	}
	result, err := useCase.Authenticate(ctx, authflow.AuthenticateRequest{ConnID: connID, Ticket: packet.Ticket, MachineID: machineID})
	return result.UserID, err == nil
}

// handleDisconnect handles one client disconnect packet and closes connection.
func (handler *Handler) handleDisconnect(connID string, body []byte, useCase *sessionflow.DisconnectUseCase) bool {
	packet := packetsession.ClientDisconnectPacket{}
	if packet.Decode(body) != nil {
		return false
	}
	_ = useCase.Disconnect(connID)
	return true
}

// startRuntimeWatchers starts timeout and distributed-close watcher goroutines.
func (handler *Handler) startRuntimeWatchers(ctx context.Context, useCases *runtimeUseCases, connID string, authSignal <-chan struct{}, signals <-chan CloseSignal, transport *Transport, cancel context.CancelFunc) {
	go func() {
		if useCases.timeout.Wait(ctx, connID, authSignal) != nil {
			cancel()
		}
	}()
	go func() {
		select {
		case signal, open := <-signals:
			if open {
				if signal.ProtocolReason != 0 {
					pkt := packetdisconnect.DisconnectReasonPacket{Reason: signal.ProtocolReason}
					if body, err := pkt.Encode(); err == nil {
						_ = transport.writeFrame(pkt.PacketID(), body)
					}
				}
				_ = transport.closeLocal(signal.Code, signal.Reason)
			}
			cancel()
		case <-ctx.Done():
		}
	}()
}
