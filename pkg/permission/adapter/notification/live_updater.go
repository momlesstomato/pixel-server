package notification

import (
	"context"
	"fmt"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	"github.com/momlesstomato/pixel-server/core/codec"
	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
	userpacket "github.com/momlesstomato/pixel-server/pkg/user/packet/profile"
)

// LiveUpdater publishes updated user access packets after group changes.
type LiveUpdater struct {
	// broadcaster stores distributed publish behavior.
	broadcaster broadcast.Broadcaster
}

// NewLiveUpdater creates one permission live updater.
func NewLiveUpdater(broadcaster broadcast.Broadcaster) (*LiveUpdater, error) {
	if broadcaster == nil {
		return nil, fmt.Errorf("broadcaster is required")
	}
	return &LiveUpdater{broadcaster: broadcaster}, nil
}

// PushAccessUpdate publishes updated user.permissions and user.perks packets.
func (updater *LiveUpdater) PushAccessUpdate(ctx context.Context, access permissiondomain.Access, perks []permissiondomain.PerkGrant) error {
	perkEntries := make([]userpacket.PerkEntry, 0, len(perks))
	for _, perk := range perks {
		perkEntries = append(perkEntries, userpacket.PerkEntry{Code: perk.Code, ErrorMessage: perk.ErrorMessage, IsAllowed: perk.IsAllowed})
	}
	permissions := userpacket.UserPermissionsPacket{
		ClubLevel: int32(access.PrimaryGroup.ClubLevel), SecurityLevel: int32(access.PrimaryGroup.SecurityLevel), IsAmbassador: access.PrimaryGroup.IsAmbassador,
	}
	perksPacket := userpacket.UserPerksPacket{Entries: perkEntries}
	if err := updater.publish(ctx, access.UserID, permissions.PacketID(), permissions); err != nil {
		return err
	}
	return updater.publish(ctx, access.UserID, perksPacket.PacketID(), perksPacket)
}

// publish encodes and publishes one packet frame to the user notification channel.
func (updater *LiveUpdater) publish(ctx context.Context, userID int, packetID uint16, packet interface{ Encode() ([]byte, error) }) error {
	body, err := packet.Encode()
	if err != nil {
		return err
	}
	return updater.broadcaster.Publish(ctx, sessionnotification.UserChannel(userID), codec.EncodeFrame(packetID, body))
}
