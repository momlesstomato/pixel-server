package hotelstatus

import (
	"context"
	"time"

	"github.com/momlesstomato/pixel-server/core/codec"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
	packethotel "github.com/momlesstomato/pixel-server/pkg/session/packet/hotel"
	statusdomain "github.com/momlesstomato/pixel-server/pkg/status/domain"
)

// StartCountdownTicker starts one countdown loop that advances closing/closed transitions.
func (service *Service) StartCountdownTicker(ctx context.Context) {
	seconds := service.config.CountdownTickSeconds
	if seconds <= 0 {
		seconds = 60
	}
	ticker := time.NewTicker(time.Duration(seconds) * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, _ = service.Tick(ctx)
			}
		}
	}()
}

// Tick advances lifecycle countdown and emits hotel close/open packets.
func (service *Service) Tick(ctx context.Context) (statusdomain.HotelStatus, error) {
	current, err := service.Current(ctx)
	if err != nil {
		return statusdomain.HotelStatus{}, err
	}
	now := service.now().UTC()
	if current.State == statusdomain.StateClosing && current.CloseAt != nil {
		remaining := int32(current.CloseAt.Sub(now).Minutes())
		if remaining > 0 {
			service.publishWillClose(ctx, remaining)
			return current, nil
		}
		next := statusdomain.HotelStatus{State: statusdomain.StateClosed, ReopenAt: current.ReopenAt, UserThrownOutAtClose: current.UserThrownOutAtClose}
		if err := service.swap(ctx, current, next); err != nil {
			return statusdomain.HotelStatus{}, err
		}
		service.publishClosedPacket(ctx, next)
		service.publishDisconnectIfThrownOut(ctx, next)
		return next, nil
	}
	if current.State == statusdomain.StateClosed && current.ReopenAt != nil && !now.Before(current.ReopenAt.UTC()) {
		next := statusdomain.HotelStatus{State: statusdomain.StateOpen}
		if err := service.swap(ctx, current, next); err != nil {
			return statusdomain.HotelStatus{}, err
		}
		return next, nil
	}
	return current, nil
}

// publishClosingPackets publishes maintenance and reopen schedule packets.
func (service *Service) publishClosingPackets(ctx context.Context, status statusdomain.HotelStatus, minutesUntilClose int32, duration int32) {
	service.publishWillClose(ctx, minutesUntilClose)
	if status.ReopenAt != nil {
		packet := packethotel.ClosesAndOpensAtPacket{OpenHour: int32(status.ReopenAt.UTC().Hour()), OpenMinute: int32(status.ReopenAt.UTC().Minute()), UserThrownOutAtClose: status.UserThrownOutAtClose}
		service.publishPacket(ctx, packet.PacketID(), packet)
	}
	packet := packethotel.MaintenancePacket{IsInMaintenance: true, MinutesUntilChange: minutesUntilClose, Duration: duration}
	service.publishPacket(ctx, packet.PacketID(), packet)
}

// publishWillClose publishes one closing countdown packet.
func (service *Service) publishWillClose(ctx context.Context, minutes int32) {
	packet := packethotel.WillClosePacket{Minutes: minutes}
	service.publishPacket(ctx, packet.PacketID(), packet)
}

// publishClosedPacket publishes one closed-and-opens packet.
func (service *Service) publishClosedPacket(ctx context.Context, status statusdomain.HotelStatus) {
	if status.ReopenAt == nil {
		return
	}
	packet := packethotel.ClosedAndOpensPacket{OpenHour: int32(status.ReopenAt.UTC().Hour()), OpenMinute: int32(status.ReopenAt.UTC().Minute())}
	service.publishPacket(ctx, packet.PacketID(), packet)
}

// publishDisconnectIfThrownOut publishes hotel-closed disconnect reason when configured.
func (service *Service) publishDisconnectIfThrownOut(ctx context.Context, status statusdomain.HotelStatus) {
	if !status.UserThrownOutAtClose {
		return
	}
	packet := packetauth.DisconnectReasonPacket{Reason: packetauth.DisconnectReasonHotelClosed}
	service.publishPacket(ctx, packet.PacketID(), packet)
}

// publishPacket publishes one encoded protocol packet to hotel broadcast channel.
func (service *Service) publishPacket(ctx context.Context, packetID uint16, packet interface{ Encode() ([]byte, error) }) {
	body, err := packet.Encode()
	if err != nil {
		return
	}
	_ = service.broadcaster.Publish(ctx, service.config.BroadcastChannel, codec.EncodeFrame(packetID, body))
}
