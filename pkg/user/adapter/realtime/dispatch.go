package realtime

import (
	"context"

	packetignore "github.com/momlesstomato/pixel-server/pkg/user/packet/ignore"
	packetname "github.com/momlesstomato/pixel-server/pkg/user/packet/name"
	packetprofile "github.com/momlesstomato/pixel-server/pkg/user/packet/profile"
	packetprofileview "github.com/momlesstomato/pixel-server/pkg/user/packet/profileview"
	packetwardrobe "github.com/momlesstomato/pixel-server/pkg/user/packet/wardrobe"
	"go.uber.org/zap"
)

// Handle dispatches one authenticated user packet payload.
func (runtime *Runtime) Handle(ctx context.Context, connID string, packetID uint16, body []byte) (bool, error) {
	userID, ok := runtime.userID(connID)
	if !ok {
		return false, nil
	}
	switch packetID {
	case packetprofile.UserGetInfoPacketID:
		return true, runtime.handleGetInfo(ctx, connID, userID)
	case packetprofile.UserUpdateMottoPacketID:
		return true, runtime.handleUpdateMotto(ctx, connID, userID, body)
	case packetprofile.UserUpdateFigurePacketID:
		return true, runtime.handleUpdateFigure(ctx, connID, userID, body)
	case packetprofile.UserSetHomeRoomPacketID:
		return true, runtime.handleSetHomeRoom(ctx, connID, userID, body)
	case packetprofile.UserRespectPacketID:
		return true, runtime.handleRespect(ctx, connID, userID, body)
	case packetprofile.UserSettingsVolumePacketID:
		return true, runtime.handleSettingsVolume(connID, userID, body)
	case runtime.packetIDs.SettingsRoomInvites:
		return true, runtime.handleSettingsRoomInvites(connID, userID, body)
	case runtime.packetIDs.SettingsOldChat:
		return true, runtime.handleSettingsOldChat(connID, userID, body)
	case packetwardrobe.UserGetWardrobePacketID:
		return true, runtime.handleGetWardrobe(ctx, connID, userID, body)
	case packetwardrobe.UserSaveWardrobeOutfitPacketID:
		return true, runtime.handleSaveWardrobe(ctx, connID, userID, body)
	case packetignore.UserGetIgnoredPacketID:
		return true, runtime.handleGetIgnored(ctx, connID, userID)
	case packetignore.UserIgnorePacketID:
		return true, runtime.handleIgnore(ctx, connID, userID, body)
	case runtime.packetIDs.Unignore:
		return true, runtime.handleUnignore(ctx, connID, userID, body)
	case runtime.packetIDs.IgnoreByID:
		return true, runtime.handleIgnoreByID(ctx, connID, userID, body)
	case packetprofileview.UserGetProfilePacketID:
		return true, runtime.handleGetProfile(ctx, connID, body)
	case packetname.UserCheckNamePacketID:
		return true, runtime.handleCheckName(ctx, connID, userID, body)
	case packetname.UserChangeNamePacketID:
		return true, runtime.handleChangeName(ctx, connID, userID, body, false)
	case runtime.packetIDs.ApproveName:
		return true, runtime.handleChangeName(ctx, connID, userID, body, true)
	default:
		return false, nil
	}
}

// userID resolves authenticated user identifier for one connection.
func (runtime *Runtime) userID(connID string) (int, bool) {
	session, found := runtime.sessions.FindByConnID(connID)
	if !found || session.UserID <= 0 {
		return 0, false
	}
	return session.UserID, true
}

// sendPacket encodes and sends one packet payload.
func (runtime *Runtime) sendPacket(connID string, packet interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) error {
	body, err := packet.Encode()
	if err != nil {
		return err
	}
	return runtime.transport.Send(connID, packet.PacketID(), body)
}

// logError logs one packet handling failure.
func (runtime *Runtime) logError(connID string, packetID uint16, err error) error {
	if err != nil {
		runtime.logger.Warn("user packet handling failed", zap.String("conn_id", connID), zap.Uint16("packet_id", packetID), zap.Error(err))
	}
	return err
}
