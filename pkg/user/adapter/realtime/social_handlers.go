package realtime

import (
	"context"
	"strings"

	userdomain "github.com/momlesstomato/pixel-server/pkg/user/domain"
	packetignore "github.com/momlesstomato/pixel-server/pkg/user/packet/ignore"
	packetprofileview "github.com/momlesstomato/pixel-server/pkg/user/packet/profileview"
	packetwardrobe "github.com/momlesstomato/pixel-server/pkg/user/packet/wardrobe"
)

// handleGetWardrobe processes one user.get_wardrobe packet.
func (runtime *Runtime) handleGetWardrobe(ctx context.Context, connID string, userID int, body []byte) error {
	packet := packetwardrobe.UserGetWardrobePacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	slots, err := runtime.service.LoadWardrobe(ctx, userID)
	if err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	entries := make([]packetwardrobe.SlotEntry, 0, len(slots))
	for _, slot := range slots {
		entries = append(entries, packetwardrobe.SlotEntry{SlotID: int32(slot.SlotID), Figure: slot.Figure, Gender: slot.Gender})
	}
	return runtime.sendPacket(connID, packetwardrobe.UserWardrobePagePacket{PageID: packet.PageID, Slots: entries})
}

// handleSaveWardrobe processes one user.save_wardrobe_outfit packet.
func (runtime *Runtime) handleSaveWardrobe(ctx context.Context, connID string, userID int, body []byte) error {
	packet := packetwardrobe.UserSaveWardrobeOutfitPacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	slot := userdomain.WardrobeSlot{SlotID: int(packet.SlotID), Figure: packet.Figure, Gender: packet.Gender}
	if err := runtime.service.SaveWardrobeSlot(ctx, userID, slot); err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	return runtime.handleGetWardrobe(ctx, connID, userID, mustEncodeWardrobe(packetwardrobe.UserGetWardrobePacket{PageID: 1}))
}

// handleGetIgnored processes one user.get_ignored packet.
func (runtime *Runtime) handleGetIgnored(ctx context.Context, connID string, userID int) error {
	names, err := runtime.service.ListIgnoredUsernames(ctx, userID)
	if err != nil {
		return runtime.logError(connID, packetignore.UserGetIgnoredPacketID, err)
	}
	return runtime.sendPacket(connID, packetignore.UserIgnoredUsersPacket{Usernames: names})
}

// handleIgnore processes one user.ignore packet.
func (runtime *Runtime) handleIgnore(ctx context.Context, connID string, userID int, body []byte) error {
	packet := packetignore.UserIgnorePacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	result := int32(0)
	_, err := runtime.service.IgnoreUserByUsername(ctx, connID, userID, packet.Username)
	if err != nil {
		result = 1
	}
	if sendErr := runtime.sendPacket(connID, packetignore.UserIgnoreResultPacket{Result: result, Name: strings.TrimSpace(packet.Username)}); sendErr != nil {
		return sendErr
	}
	return runtime.handleGetIgnored(ctx, connID, userID)
}

// handleUnignore processes one user.unignore packet.
func (runtime *Runtime) handleUnignore(ctx context.Context, connID string, userID int, body []byte) error {
	packet := packetignore.UserIgnorePacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, runtime.packetIDs.Unignore, err)
	}
	result := int32(0)
	if _, err := runtime.service.UnignoreUserByUsername(ctx, connID, userID, packet.Username); err != nil {
		result = 1
	}
	if sendErr := runtime.sendPacket(connID, packetignore.UserIgnoreResultPacket{Result: result, Name: strings.TrimSpace(packet.Username)}); sendErr != nil {
		return sendErr
	}
	return runtime.handleGetIgnored(ctx, connID, userID)
}

// handleIgnoreByID processes one user.ignore_id packet.
func (runtime *Runtime) handleIgnoreByID(ctx context.Context, connID string, userID int, body []byte) error {
	packet := packetignore.UserIgnoreByIDPacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, runtime.packetIDs.IgnoreByID, err)
	}
	_ = runtime.service.IgnoreUserByID(ctx, connID, userID, int(packet.UserID))
	return runtime.handleGetIgnored(ctx, connID, userID)
}

// handleGetProfile processes one user.get_profile packet.
func (runtime *Runtime) handleGetProfile(ctx context.Context, connID string, body []byte) error {
	packet := packetprofileview.UserGetProfilePacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	profile, err := runtime.service.LoadProfile(ctx, int(packet.UserID), packet.OpenProfileWindow)
	if err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	response := packetprofileview.UserProfilePacket{
		UserID: int32(profile.UserID), Username: profile.Username, Figure: profile.Figure,
		Motto: profile.Motto, Registration: "", AchievementPoints: 0, FriendsCount: 0,
		IsMyFriend: false, RequestSent: false, IsOnline: profile.IsOnline, SecondsSinceLastVisit: 0,
		OpenProfileWindow: profile.OpenProfileWindow,
	}
	return runtime.sendPacket(connID, response)
}

// mustEncodeWardrobe serializes wardrobe request payload.
func mustEncodeWardrobe(packet packetwardrobe.UserGetWardrobePacket) []byte {
	body, _ := packet.Encode()
	return body
}
