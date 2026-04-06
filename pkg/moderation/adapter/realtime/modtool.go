package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/packet"
	"go.uber.org/zap"
)

// sendModeratorInit sends the moderator tool initialization packet.
func (rt *Runtime) sendModeratorInit(ctx context.Context, connID string, userID int) error {
	if rt.presets == nil {
		return nil
	}
	if rt.permissions == nil {
		return nil
	}
	ok, err := rt.permissions.HasPermission(ctx, userID, domain.PermTool)
	if err != nil || !ok {
		return nil
	}
	presets, err := rt.presets.ListActive(ctx)
	if err != nil {
		return nil
	}
	templates := make([]string, 0, len(presets))
	for _, p := range presets {
		templates = append(templates, p.Name)
	}
	hasCfh, _ := rt.permissions.HasPermission(ctx, userID, domain.PermTool)
	hasChatlogs, _ := rt.permissions.HasPermission(ctx, userID, domain.PermHistory)
	hasAlert, _ := rt.permissions.HasPermission(ctx, userID, domain.PermWarn)
	hasKick, _ := rt.permissions.HasPermission(ctx, userID, domain.PermKick)
	hasBan, _ := rt.permissions.HasPermission(ctx, userID, domain.PermBan)
	pkt := packet.ModeratorInitPacket{
		MessageTemplates:    templates,
		CfhPermission:       hasCfh,
		ChatlogsPermission:  hasChatlogs,
		AlertPermission:     hasAlert,
		KickPermission:      hasKick,
		BanPermission:       hasBan,
		RoomAlertPermission: hasAlert,
		RoomKickPermission:  hasKick,
	}
	body, err := pkt.Encode()
	if err != nil {
		return err
	}
	return rt.transport.Send(connID, packet.ModeratorInitPacketID, body)
}

// SendModToolInit is the public entry point for sending mod tool init on login.
func (rt *Runtime) SendModToolInit(ctx context.Context, connID string, userID int) {
	if err := rt.sendModeratorInit(ctx, connID, userID); err != nil {
		rt.logger.Warn("send mod tool init failed", zap.Int("user_id", userID), zap.Error(err))
	}
}

// OnPostAuth sends the moderator tool initialization after authentication.
func (rt *Runtime) OnPostAuth(ctx context.Context, connID string, userID int) {
	rt.SendModToolInit(ctx, connID, userID)
}
