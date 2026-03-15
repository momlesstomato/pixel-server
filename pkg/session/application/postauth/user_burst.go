package postauth

import (
	"context"
	"time"

	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
	userdomain "github.com/momlesstomato/pixel-server/pkg/user/domain"
	userignorepacket "github.com/momlesstomato/pixel-server/pkg/user/packet/ignore"
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
	// ignoredUsernames stores ignored usernames payload.
	ignoredUsernames []string
	// access stores resolved user access payload.
	access permissiondomain.Access
	// perks stores resolved user perk grants.
	perks []permissiondomain.PerkGrant
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
	if err := useCase.sendUserSettingsPackets(connID, snapshot); err != nil {
		return err
	}
	ignored := userignorepacket.UserIgnoredUsersPacket{Usernames: snapshot.ignoredUsernames}
	return sendPacket(useCase.transport, connID, ignored.PacketID(), ignored)
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
	ignoredUsernames, err := useCase.profiles.ListIgnoredUsernames(ctx, userID)
	if err != nil {
		return userBurstSnapshot{}, err
	}
	access, err := useCase.access.ResolveAccess(ctx, userID)
	if err != nil {
		return userBurstSnapshot{}, err
	}
	return userBurstSnapshot{
		user: user, settings: settings, respectsRemaining: remaining, respectsPetRemaining: petRemaining,
		lastAccessDate: useCase.now().UTC().Format(time.RFC3339), ignoredUsernames: ignoredUsernames,
		access: access, perks: useCase.access.ResolvePerks(access),
	}, nil
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
	permissions := userpacket.UserPermissionsPacket{
		ClubLevel:     int32(snapshot.access.PrimaryGroup.ClubLevel),
		SecurityLevel: int32(snapshot.access.PrimaryGroup.SecurityLevel),
		IsAmbassador:  snapshot.access.PrimaryGroup.IsAmbassador,
	}
	if err := sendPacket(useCase.transport, connID, userpacket.UserPermissionsPacketID, permissions); err != nil {
		return err
	}
	entries := make([]userpacket.PerkEntry, 0, len(snapshot.perks))
	for _, perk := range snapshot.perks {
		entries = append(entries, userpacket.PerkEntry{Code: perk.Code, ErrorMessage: perk.ErrorMessage, IsAllowed: perk.IsAllowed})
	}
	perks := userpacket.UserPerksPacket{Entries: entries}
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
