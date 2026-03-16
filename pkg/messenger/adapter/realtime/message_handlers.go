package realtime

import (
	"context"
	"errors"

	"github.com/momlesstomato/pixel-server/core/codec"
	messengerdomain "github.com/momlesstomato/pixel-server/pkg/messenger/domain"
	packetmessage "github.com/momlesstomato/pixel-server/pkg/messenger/packet/message"
	packetsocial "github.com/momlesstomato/pixel-server/pkg/messenger/packet/social"
)

// handleSendMsg handles messenger.send_msg.
func (runtime *Runtime) handleSendMsg(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packetmessage.MessengerSendMsgPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	if err := runtime.service.SendMessage(ctx, connID, userID, int(pkt.UserID), pkt.Message); err != nil {
		runtime.logger.Sugar().Warnw("send message failed", "conn", connID, "to", pkt.UserID, "err", err)
		_ = runtime.sendPacket(connID, packetmessage.MessengerMessageErrorComposer{ErrorCode: mapMessageErrorCode(err), UserID: pkt.UserID})
		return nil
	}
	_, online := runtime.sessions.FindByUserID(int(pkt.UserID))
	if !online {
		return nil
	}
	composer := packetmessage.MessengerNewMessageComposer{
		SenderID: int32(userID),
		Message:  pkt.Message,
	}
	body, err := composer.Encode()
	if err != nil {
		return nil
	}
	frame := codec.EncodeFrame(composer.PacketID(), body)
	_ = runtime.service.NotifyFriendUpdate(ctx, int(pkt.UserID), frame)
	return nil
}

// mapMessageErrorCode converts application message delivery errors to protocol codes.
func mapMessageErrorCode(err error) int32 {
	if errors.Is(err, messengerdomain.ErrNotFriends) {
		return 0
	}
	if errors.Is(err, messengerdomain.ErrSenderMuted) {
		return 1
	}
	return 2
}

// handleSendInvite handles messenger.send_invite.
func (runtime *Runtime) handleSendInvite(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packetsocial.MessengerSendInvitePacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	toUserIDs := make([]int, 0, len(pkt.UserIDs))
	for _, id := range pkt.UserIDs {
		toUserIDs = append(toUserIDs, int(id))
	}
	if err := runtime.service.SendRoomInvite(ctx, connID, userID, toUserIDs, pkt.Message); err != nil {
		return nil
	}
	for _, toUserID := range toUserIDs {
		_, online := runtime.sessions.FindByUserID(toUserID)
		if !online {
			continue
		}
		composer := packetsocial.MessengerRoomInviteComposer{SenderID: int32(userID), Message: pkt.Message}
		body, encodeErr := composer.Encode()
		if encodeErr != nil {
			continue
		}
		frame := codec.EncodeFrame(composer.PacketID(), body)
		if err := runtime.service.NotifyFriendUpdate(ctx, toUserID, frame); err != nil {
			runtime.logger.Sugar().Warnw("send invite failed", "to", toUserID, "err", err)
		}
	}
	return nil
}

// handleFollowFriend handles messenger.follow_friend.
func (runtime *Runtime) handleFollowFriend(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packetsocial.MessengerFollowFriendPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	if err := runtime.service.FollowFriend(ctx, userID, int(pkt.FriendID)); err != nil {
		code := int32(2)
		if errors.Is(err, messengerdomain.ErrNotFriends) {
			code = 0
		} else if errors.Is(err, messengerdomain.ErrFollowBlocked) {
			code = 3
		}
		return runtime.sendPacket(connID, packetsocial.MessengerFollowFailedComposer{ErrorCode: code})
	}
	return nil
}
