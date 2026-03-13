package realtime

import (
	"context"
	"time"

	userdomain "github.com/momlesstomato/pixel-server/pkg/user/domain"
	packetprofile "github.com/momlesstomato/pixel-server/pkg/user/packet/profile"
)

// handleGetInfo sends one user.info packet.
func (runtime *Runtime) handleGetInfo(ctx context.Context, connID string, userID int) error {
	user, err := runtime.service.FindByID(ctx, userID)
	if err != nil {
		return runtime.logError(connID, packetprofile.UserGetInfoPacketID, err)
	}
	remaining, err := runtime.service.RemainingRespects(ctx, userID, userdomain.RespectTargetUser, time.Now().UTC())
	if err != nil {
		return runtime.logError(connID, packetprofile.UserGetInfoPacketID, err)
	}
	petRemaining, err := runtime.service.RemainingRespects(ctx, userID, userdomain.RespectTargetPet, time.Now().UTC())
	if err != nil {
		return runtime.logError(connID, packetprofile.UserGetInfoPacketID, err)
	}
	packet := packetprofile.UserInfoPacket{
		UserID: int32(user.ID), Username: user.Username, Figure: user.Figure, Gender: user.Gender,
		Motto: user.Motto, RealName: user.RealName, DirectMail: false, RespectsReceived: int32(user.RespectsReceived),
		RespectsRemaining: int32(remaining), RespectsPetRemaining: int32(petRemaining), StreamPublishingAllowed: false,
		LastAccessDate: time.Now().UTC().Format(time.RFC3339), CanChangeName: user.CanChangeName, SafetyLocked: user.SafetyLocked,
	}
	return runtime.sendPacket(connID, packet)
}

// handleUpdateMotto processes one user.update_motto packet.
func (runtime *Runtime) handleUpdateMotto(ctx context.Context, connID string, userID int, body []byte) error {
	packet := packetprofile.UserUpdateMottoPacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	_, err := runtime.service.UpdateMotto(ctx, connID, userID, packet.Motto)
	return runtime.logError(connID, packet.PacketID(), err)
}

// handleUpdateFigure processes one user.update_figure packet.
func (runtime *Runtime) handleUpdateFigure(ctx context.Context, connID string, userID int, body []byte) error {
	packet := packetprofile.UserUpdateFigurePacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	updated, err := runtime.service.UpdateFigure(ctx, connID, userID, packet.Gender, packet.Figure)
	if err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	return runtime.sendPacket(connID, packetprofile.UserFigurePacket{Figure: updated.Figure, Gender: updated.Gender})
}

// handleSetHomeRoom processes one user.set_home_room packet.
func (runtime *Runtime) handleSetHomeRoom(ctx context.Context, connID string, userID int, body []byte) error {
	packet := packetprofile.UserSetHomeRoomPacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	updated, err := runtime.service.SetHomeRoom(ctx, userID, int(packet.RoomID))
	if err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	value := int32(updated.HomeRoomID)
	return runtime.sendPacket(connID, packetprofile.UserHomeRoomPacket{HomeRoomID: value, RoomIDToEnter: value})
}

// handleRespect processes one user.respect packet.
func (runtime *Runtime) handleRespect(ctx context.Context, connID string, actorUserID int, body []byte) error {
	packet := packetprofile.UserRespectPacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	result, err := runtime.service.RecordUserRespectWithConn(ctx, connID, actorUserID, int(packet.UserID), time.Now().UTC())
	if err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	response := packetprofile.UserRespectReceivedPacket{UserID: packet.UserID, RespectsReceived: int32(result.RespectsReceived)}
	return runtime.sendPacket(connID, response)
}
