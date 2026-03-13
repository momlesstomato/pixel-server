package notification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	"github.com/momlesstomato/pixel-server/core/codec"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
	packeterror "github.com/momlesstomato/pixel-server/pkg/session/packet/error"
	packetnotification "github.com/momlesstomato/pixel-server/pkg/session/packet/notification"
)

// Service defines session notification publish behavior.
type Service struct {
	// broadcaster publishes packet frames to distributed channels.
	broadcaster broadcast.Broadcaster
	// now returns deterministic timestamps for tests.
	now func() time.Time
}

// NewService creates one session notification service.
func NewService(broadcaster broadcast.Broadcaster) (*Service, error) {
	if broadcaster == nil {
		return nil, fmt.Errorf("broadcaster is required")
	}
	return &Service{broadcaster: broadcaster, now: time.Now}, nil
}

// SendConnectionError sends one protocol connection error packet to one user channel.
func (service *Service) SendConnectionError(ctx context.Context, userID int, messageID uint16, errorCode int32) error {
	packet := packeterror.ConnectionErrorPacket{
		MessageID: int32(messageID), ErrorCode: errorCode, Timestamp: service.now().UTC().Format(time.RFC3339Nano),
	}
	return service.publishPacket(ctx, UserChannel(userID), packet.PacketID(), packet)
}

// SendGenericError sends one generic error packet to one user channel.
func (service *Service) SendGenericError(ctx context.Context, userID int, errorCode int32) error {
	packet := packetnotification.GenericErrorPacket{ErrorCode: errorCode}
	return service.publishPacket(ctx, UserChannel(userID), packet.PacketID(), packet)
}

// SendGenericAlert sends one generic alert packet to one user channel.
func (service *Service) SendGenericAlert(ctx context.Context, userID int, message string) error {
	packet := packetnotification.GenericAlertPacket{Message: strings.TrimSpace(message)}
	return service.publishPacket(ctx, UserChannel(userID), packet.PacketID(), packet)
}

// SendModerationCaution sends one moderation caution packet to one user channel.
func (service *Service) SendModerationCaution(ctx context.Context, userID int, message string, detail string) error {
	packet := packetnotification.ModerationCautionPacket{Message: strings.TrimSpace(message), Detail: strings.TrimSpace(detail)}
	return service.publishPacket(ctx, UserChannel(userID), packet.PacketID(), packet)
}

// SendJustBannedDisconnect sends one just-banned disconnect reason to one user channel.
func (service *Service) SendJustBannedDisconnect(ctx context.Context, userID int) error {
	packet := packetauth.DisconnectReasonPacket{Reason: packetauth.DisconnectReasonJustBanned}
	return service.publishPacket(ctx, UserChannel(userID), packet.PacketID(), packet)
}

// SendStillBannedDisconnect sends one still-banned disconnect reason to one user channel.
func (service *Service) SendStillBannedDisconnect(ctx context.Context, userID int) error {
	packet := packetauth.DisconnectReasonPacket{Reason: packetauth.DisconnectReasonStillBanned}
	return service.publishPacket(ctx, UserChannel(userID), packet.PacketID(), packet)
}

// publishPacket encodes one packet frame and publishes it to one channel.
func (service *Service) publishPacket(ctx context.Context, channel string, packetID uint16, packet interface{ Encode() ([]byte, error) }) error {
	if userID, err := parseUserChannel(channel); err != nil || userID <= 0 {
		return fmt.Errorf("valid user channel is required")
	}
	body, err := packet.Encode()
	if err != nil {
		return err
	}
	return service.broadcaster.Publish(ctx, channel, codec.EncodeFrame(packetID, body))
}

// parseUserChannel parses one user channel and returns user identifier.
func parseUserChannel(channel string) (int, error) {
	var userID int
	_, err := fmt.Sscanf(channel, "broadcast:user:%d", &userID)
	return userID, err
}
