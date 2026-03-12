package authflow

import (
	"context"
	"fmt"
	"time"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
)

// AuthenticateUseCase defines authentication workflow behavior.
type AuthenticateUseCase struct {
	// validator validates and consumes SSO tickets.
	validator TicketValidator
	// sessions stores authenticated session lifecycle state.
	sessions SessionRegistry
	// transport sends packets and closes connections.
	transport Transport
	// now provides deterministic time source for session timestamps.
	now func() time.Time
}

// NewAuthenticateUseCase creates authentication workflow behavior.
func NewAuthenticateUseCase(validator TicketValidator, sessions SessionRegistry, transport Transport) (*AuthenticateUseCase, error) {
	if validator == nil {
		return nil, fmt.Errorf("ticket validator is required")
	}
	if sessions == nil {
		return nil, fmt.Errorf("session registry is required")
	}
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	return &AuthenticateUseCase{validator: validator, sessions: sessions, transport: transport, now: time.Now}, nil
}

// Authenticate validates ticket, handles duplicate sessions, and emits auth packets.
func (useCase *AuthenticateUseCase) Authenticate(ctx context.Context, request AuthenticateRequest) (AuthenticateResult, error) {
	if request.ConnID == "" {
		return AuthenticateResult{}, fmt.Errorf("connection id is required")
	}
	ticket, err := normalizeTicket(request.Ticket)
	if err != nil {
		useCase.closeWithReason(request.ConnID, packetauth.DisconnectReasonInvalidLoginTicket, UnauthorizedCloseCode, "unauthorized")
		return AuthenticateResult{}, err
	}
	userID, err := useCase.validator.Validate(ctx, ticket)
	if err != nil {
		useCase.closeWithReason(request.ConnID, packetauth.DisconnectReasonInvalidLoginTicket, UnauthorizedCloseCode, "unauthorized")
		return AuthenticateResult{}, err
	}
	kickedConnID := ""
	existing, found := useCase.sessions.FindByUserID(userID)
	if found && existing.ConnID != "" && existing.ConnID != request.ConnID {
		kickedConnID = existing.ConnID
		useCase.closeWithReason(existing.ConnID, packetauth.DisconnectReasonConcurrentLogin, DuplicateLoginCloseCode, "duplicate login")
		useCase.sessions.Remove(existing.ConnID)
	}
	session := coreconnection.Session{
		ConnID: request.ConnID, UserID: userID, MachineID: request.MachineID, State: coreconnection.StateAuthenticated, CreatedAt: useCase.now(),
	}
	if err := useCase.sessions.Register(session); err != nil {
		return AuthenticateResult{}, err
	}
	if err := useCase.sendAuthenticationOK(request.ConnID); err != nil {
		return AuthenticateResult{}, err
	}
	if err := useCase.sendIdentityAccounts(request.ConnID, userID); err != nil {
		return AuthenticateResult{}, err
	}
	return AuthenticateResult{UserID: userID, KickedConnID: kickedConnID}, nil
}

// sendAuthenticationOK sends the authentication success packet.
func (useCase *AuthenticateUseCase) sendAuthenticationOK(connID string) error {
	packet := packetauth.AuthenticationOKPacket{}
	body, err := packet.Encode()
	if err != nil {
		return err
	}
	return useCase.transport.Send(connID, packet.PacketID(), body)
}

// sendIdentityAccounts sends identity account list packet.
func (useCase *AuthenticateUseCase) sendIdentityAccounts(connID string, userID int) error {
	packet := packetauth.IdentityAccountsPacket{Accounts: []packetauth.IdentityAccount{{ID: int32(userID), Name: fmt.Sprintf("Player#%d", userID)}}}
	body, err := packet.Encode()
	if err != nil {
		return err
	}
	return useCase.transport.Send(connID, packet.PacketID(), body)
}

// closeWithReason sends one disconnect reason packet and closes the connection.
func (useCase *AuthenticateUseCase) closeWithReason(connID string, reason int32, closeCode int, closeReason string) {
	packet := packetauth.DisconnectReasonPacket{Reason: reason}
	body, err := packet.Encode()
	if err == nil {
		_ = useCase.transport.Send(connID, packet.PacketID(), body)
	}
	_ = useCase.transport.Close(connID, closeCode, closeReason)
}
