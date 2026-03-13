package postauth

import (
	"context"
	"time"

	userdomain "github.com/momlesstomato/pixel-server/pkg/user/domain"
	userpacket "github.com/momlesstomato/pixel-server/pkg/user/packet/profile"
)

// userBurstSnapshot defines data required to compose post-auth user packets.
type userBurstSnapshot struct {
	// user stores resolved identity payload.
	user userdomain.User
	// settings stores resolved settings payload.
	settings userdomain.Settings
	// respectsRemaining stores remaining user-to-user respects.
	respectsRemaining int
	// respectsPetRemaining stores remaining user-to-pet respects.
	respectsPetRemaining int
	// lastAccessDate stores formatted last access date payload.
	lastAccessDate string
}

// sendUserBurst sends user profile, permissions, perks, noobness, settings, and home room packets.
func (useCase *UseCase) sendUserBurst(ctx context.Context, connID string, userID int) error {
	snapshot, err := useCase.loadUserBurstSnapshot(ctx, userID)
	if err != nil {
		return err
	}
	if err := useCase.sendUserIdentityPackets(connID, snapshot); err != nil {
		return err
	}
	if err := useCase.sendUserAccessPackets(connID, snapshot); err != nil {
		return err
	}
	return useCase.sendUserSettingsPackets(connID, snapshot)
}

// loadUserBurstSnapshot loads one post-auth user snapshot payload.
func (useCase *UseCase) loadUserBurstSnapshot(ctx context.Context, userID int) (userBurstSnapshot, error) {
	user, err := useCase.profiles.FindByID(ctx, userID)
	if err != nil {
		return userBurstSnapshot{}, err
	}
	settings, err := useCase.profiles.LoadSettings(ctx, userID)
	if err != nil {
		return userBurstSnapshot{}, err
	}
	remaining, err := useCase.profiles.RemainingRespects(ctx, userID, userdomain.RespectTargetUser, useCase.now().UTC())
	if err != nil {
		return userBurstSnapshot{}, err
	}
	petRemaining, err := useCase.profiles.RemainingRespects(ctx, userID, userdomain.RespectTargetPet, useCase.now().UTC())
	if err != nil {
		return userBurstSnapshot{}, err
	}
	return userBurstSnapshot{user: user, settings: settings, respectsRemaining: remaining, respectsPetRemaining: petRemaining, lastAccessDate: useCase.now().UTC().Format(time.RFC3339)}, nil
}

// sendUserIdentityPackets sends user.info packet.
func (useCase *UseCase) sendUserIdentityPackets(connID string, snapshot userBurstSnapshot) error {
	user := snapshot.user
	packet := userpacket.UserInfoPacket{
		UserID: int32(user.ID), Username: user.Username, Figure: user.Figure,
		Gender: user.Gender, Motto: user.Motto, RealName: user.RealName,
		DirectMail: false, RespectsReceived: int32(user.RespectsReceived),
		RespectsRemaining: int32(snapshot.respectsRemaining), RespectsPetRemaining: int32(snapshot.respectsPetRemaining),
		StreamPublishingAllowed: false, LastAccessDate: snapshot.lastAccessDate,
		CanChangeName: user.CanChangeName, SafetyLocked: user.SafetyLocked,
	}
	return sendPacket(useCase.transport, connID, packet.PacketID(), packet)
}

// sendUserAccessPackets sends permissions, perks, and noobness packets.
func (useCase *UseCase) sendUserAccessPackets(connID string, snapshot userBurstSnapshot) error {
	if err := sendPacket(useCase.transport, connID, userpacket.UserPermissionsPacketID, userpacket.UserPermissionsPacket{}); err != nil {
		return err
	}
	perks := userpacket.UserPerksPacket{Entries: []userpacket.PerkEntry{{Code: "USE_GUIDE", ErrorMessage: "", IsAllowed: true}, {Code: "CAMERA", ErrorMessage: "", IsAllowed: true}}}
	if err := sendPacket(useCase.transport, connID, userpacket.UserPerksPacketID, perks); err != nil {
		return err
	}
	noobness := userpacket.UserNoobnessLevelPacket{NoobnessLevel: int32(snapshot.user.NoobnessLevel)}
	return sendPacket(useCase.transport, connID, noobness.PacketID(), noobness)
}

// sendUserSettingsPackets sends user.settings and user.home_room packets.
func (useCase *UseCase) sendUserSettingsPackets(connID string, snapshot userBurstSnapshot) error {
	settings := snapshot.settings
	settingsPacket := userpacket.UserSettingsPacket{
		VolumeSystem: int32(settings.VolumeSystem), VolumeFurni: int32(settings.VolumeFurni),
		VolumeTrax: int32(settings.VolumeTrax), OldChat: settings.OldChat,
		RoomInvites: settings.RoomInvites, CameraFollow: settings.CameraFollow,
		Flags: int32(settings.Flags), ChatType: int32(settings.ChatType),
	}
	if err := sendPacket(useCase.transport, connID, settingsPacket.PacketID(), settingsPacket); err != nil {
		return err
	}
	homeRoom := int32(snapshot.user.HomeRoomID)
	homeRoomPacket := userpacket.UserHomeRoomPacket{HomeRoomID: homeRoom, RoomIDToEnter: homeRoom}
	return sendPacket(useCase.transport, connID, homeRoomPacket.PacketID(), homeRoomPacket)
}
