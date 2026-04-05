package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/moderation/packet"
	"go.uber.org/zap"
)

// sendModeratorInit sends the moderator tool initialization packet.
func (rt *Runtime) sendModeratorInit(ctx context.Context, connID string) error {
	if rt.presets == nil {
		return nil
	}
	presets, err := rt.presets.ListActive(ctx)
	if err != nil {
		return nil
	}
	categories := make(map[string][]string)
	for _, p := range presets {
		categories[p.Category] = append(categories[p.Category], p.Name)
	}
	pktCategories := make([]packet.PresetCategory, 0, len(categories))
	for name, entries := range categories {
		pktCategories = append(pktCategories, packet.PresetCategory{Name: name, Entries: entries})
	}
	pkt := packet.ModeratorInitPacket{
		Presets: pktCategories, TicketPermission: true, ChatlogPermission: true,
	}
	body, err := pkt.Encode()
	if err != nil {
		return err
	}
	return rt.transport.Send(connID, packet.ModeratorInitPacketID, body)
}

// SendModToolInit is the public entry point for sending mod tool init on login.
func (rt *Runtime) SendModToolInit(ctx context.Context, connID string, userID int) {
	if err := rt.sendModeratorInit(ctx, connID); err != nil {
		rt.logger.Warn("send mod tool init failed", zap.Int("user_id", userID), zap.Error(err))
	}
}
