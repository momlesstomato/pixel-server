package realtime

import (
	"context"

	packetmsginit "github.com/momlesstomato/pixel-server/pkg/messenger/packet/msginit"
	packetrequest "github.com/momlesstomato/pixel-server/pkg/messenger/packet/request"
	packetmessage "github.com/momlesstomato/pixel-server/pkg/messenger/packet/message"
	packetsocial "github.com/momlesstomato/pixel-server/pkg/messenger/packet/social"
	"go.uber.org/zap"
)

// dispatch routes one messenger packet to its handler.
func (runtime *Runtime) dispatch(ctx context.Context, connID string, userID int, packetID uint16, body []byte) (bool, error) {
	switch packetID {
	case packetmsginit.MessengerInitPacketID:
		return true, runtime.handleInit(ctx, connID, userID)
	case packetmsginit.MessengerGetFriendsPacketID:
		return true, runtime.handleGetFriends(ctx, connID, userID)
	case packetmsginit.MessengerGetRequestsPacketID:
		return true, runtime.handleGetRequests(ctx, connID, userID)
	case packetrequest.MessengerSendRequestPacketID:
		return true, runtime.handleSendRequest(ctx, connID, userID, body)
	case packetrequest.MessengerAcceptFriendPacketID:
		return true, runtime.handleAcceptFriend(ctx, connID, userID, body)
	case packetrequest.MessengerDeclineFriendPacketID:
		return true, runtime.handleDeclineFriend(ctx, connID, userID, body)
	case packetrequest.MessengerRemoveFriendPacketID:
		return true, runtime.handleRemoveFriend(ctx, connID, userID, body)
	case packetmessage.MessengerSendMsgPacketID:
		return true, runtime.handleSendMsg(ctx, connID, userID, body)
	case packetsocial.MessengerSearchPacketID:
		return true, runtime.handleSearch(ctx, connID, body)
	case packetsocial.MessengerSetRelationshipPacketID:
		return true, runtime.handleSetRelationship(ctx, connID, userID, body)
	case packetsocial.MessengerGetRelationshipsPacketID:
		return true, runtime.handleGetRelationships(ctx, connID, body)
	case packetsocial.MessengerFollowFriendPacketID:
		return true, runtime.handleFollowFriend(ctx, connID, userID, body)
	case packetsocial.MessengerSendInvitePacketID:
		return true, runtime.handleSendInvite(ctx, connID, userID, body)
	default:
		return false, nil
	}
}

// logError logs one packet handling failure and returns the error.
func (runtime *Runtime) logError(connID string, packetID uint16, err error) error {
	if err != nil {
		runtime.logger.Warn("messenger packet handling failed",
			zap.String("conn_id", connID),
			zap.Uint16("packet_id", packetID),
			zap.Error(err))
	}
	return err
}
