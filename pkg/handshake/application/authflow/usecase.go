package authflow

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/momlesstomato/pixel-sdk"
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
	// users resolves real user display names for identity packets.
	users UserFinder
	// now provides deterministic time source for session timestamps.
	now func() time.Time
	// fire dispatches optional plugin lifecycle events.
	fire func(sdk.Event)
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

// SetEventFirer sets optional plugin event dispatch behavior.
func (useCase *AuthenticateUseCase) SetEventFirer(fire func(sdk.Event)) {
	useCase.fire = fire
}

// SetUserFinder wires real user identity resolution for identity account packets.
func (useCase *AuthenticateUseCase) SetUserFinder(users UserFinder) {
	useCase.users = users
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
	username, err := useCase.resolveUsername(ctx, request.ConnID, userID)
	if err != nil {
		return AuthenticateResult{}, err
	}
	if useCase.fire != nil {
		validating := &sdk.AuthValidating{ConnID: request.ConnID, UserID: userID, Ticket: ticket}
		useCase.fire(validating)
		if validating.Cancelled() {
			useCase.closeWithReason(request.ConnID, packetauth.DisconnectReasonInvalidLoginTicket, UnauthorizedCloseCode, "cancelled by plugin")
			return AuthenticateResult{}, fmt.Errorf("authentication cancelled by plugin")
		}
	}
	kickedConnID := ""
	existing, found := useCase.sessions.FindByUserID(userID)
	if found && existing.ConnID != "" && existing.ConnID != request.ConnID {
		kicked := true
		if useCase.fire != nil {
			dupKick := &sdk.DuplicateKick{OldConnID: existing.ConnID, NewConnID: request.ConnID, UserID: userID}
			useCase.fire(dupKick)
			kicked = !dupKick.Cancelled()
		}
		if kicked {
			kickedConnID = existing.ConnID
			useCase.closeWithReason(existing.ConnID, packetauth.DisconnectReasonConcurrentLogin, DuplicateLoginCloseCode, "duplicate login")
			useCase.sessions.Remove(existing.ConnID)
		}
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
	if err := useCase.sendIdentityAccounts(request.ConnID, userID, username); err != nil {
		return AuthenticateResult{}, err
	}
	if useCase.fire != nil {
		useCase.fire(&sdk.AuthCompleted{ConnID: request.ConnID, UserID: userID})
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

// sendIdentityAccounts sends identity account list packet with provided display name.
func (useCase *AuthenticateUseCase) sendIdentityAccounts(connID string, userID int, name string) error {
	packet := packetauth.IdentityAccountsPacket{Accounts: []packetauth.IdentityAccount{{ID: int32(userID), Name: name}}}
	body, err := packet.Encode()
	if err != nil {
		return err
	}
	return useCase.transport.Send(connID, packet.PacketID(), body)
}

// resolveUsername returns the real username when UserFinder is set, or disconnects on failure.
func (useCase *AuthenticateUseCase) resolveUsername(ctx context.Context, connID string, userID int) (string, error) {
	if useCase.users == nil {
		return fmt.Sprintf("Player#%d", userID), nil
	}
	username, err := useCase.users.FindByID(ctx, userID)
	if err != nil {
		useCase.closeWithReason(connID, packetauth.DisconnectReasonInvalidLoginTicket, UnauthorizedCloseCode, "user not found")
		return "", fmt.Errorf("resolve username for user %d: %w", userID, err)
	}
	return username, nil
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
