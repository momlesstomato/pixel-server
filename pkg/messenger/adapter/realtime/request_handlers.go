package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/core/codec"
	packetmsginit "github.com/momlesstomato/pixel-server/pkg/messenger/packet/msginit"
	packetrequest "github.com/momlesstomato/pixel-server/pkg/messenger/packet/request"
)

const maxAcceptBatch = 50
const maxRemoveBatch = 100

// handleSendRequest handles messenger.send_request.
func (runtime *Runtime) handleSendRequest(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packetrequest.MessengerSendRequestPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	req, autoAccepted, err := runtime.service.SendRequest(ctx, connID, userID, pkt.Username)
	if err != nil {
		runtime.logger.Sugar().Warnw("send request failed", "conn", connID, "err", err)
		return nil
	}
	if autoAccepted {
		go runtime.notifyFriendAddedBoth(ctx, userID, req.ToUserID)
		return nil
	}
	go runtime.notifyNewRequest(ctx, userID, req.ToUserID)
	return nil
}

// handleAcceptFriend handles messenger.accept_friend.
func (runtime *Runtime) handleAcceptFriend(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packetrequest.MessengerAcceptFriendPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	seen := map[int32]struct{}{}
	cap := maxAcceptBatch
	for _, fromUserID := range pkt.RequestIDs {
		if _, exists := seen[fromUserID]; exists {
			continue
		}
		seen[fromUserID] = struct{}{}
		if cap <= 0 {
			break
		}
		cap--
		if err := runtime.service.AcceptRequest(ctx, userID, int(fromUserID)); err != nil {
			runtime.logger.Sugar().Warnw("accept request failed", "conn", connID, "from_user_id", fromUserID, "err", err)
			continue
		}
		go runtime.notifyFriendAddedBoth(ctx, userID, int(fromUserID))
	}
	return nil
}

// handleDeclineFriend handles messenger.decline_friend.
func (runtime *Runtime) handleDeclineFriend(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packetrequest.MessengerDeclineFriendPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	if pkt.DeclineAll {
		return runtime.service.DeclineAllRequests(ctx, userID)
	}
	seen := map[int32]struct{}{}
	for _, reqID := range pkt.RequestIDs {
		if _, exists := seen[reqID]; exists {
			continue
		}
		seen[reqID] = struct{}{}
		if err := runtime.service.DeclineRequest(ctx, userID, int(reqID)); err != nil {
			runtime.logger.Sugar().Warnw("decline request failed", "conn", connID, "req_id", reqID, "err", err)
		}
	}
	return nil
}

// notifyFriendAddedBoth sends action=1 (added) update to both newly-connected friends.
func (runtime *Runtime) notifyFriendAddedBoth(ctx context.Context, userOneID, userTwoID int) {
	profiles, err := runtime.service.GetUserProfiles(ctx, []int{userOneID, userTwoID})
	if err != nil {
		return
	}
	var p1, p2 struct{ Username, Figure, Motto string }
	for _, p := range profiles {
		if p.ID == userOneID {
			p1.Username, p1.Figure, p1.Motto = p.Username, p.Figure, p.Motto
		}
		if p.ID == userTwoID {
			p2.Username, p2.Figure, p2.Motto = p.Username, p.Figure, p.Motto
		}
	}
	_, o1 := runtime.sessions.FindByUserID(userOneID)
	_, o2 := runtime.sessions.FindByUserID(userTwoID)
	runtime.publishFriendUpdate(ctx, userOneID, []packetrequest.FriendUpdateEntry{
		{Action: 1, FriendID: int32(userTwoID), Username: p2.Username, Online: o2, Figure: p2.Figure, Motto: p2.Motto},
	})
	runtime.publishFriendUpdate(ctx, userTwoID, []packetrequest.FriendUpdateEntry{
		{Action: 1, FriendID: int32(userOneID), Username: p1.Username, Online: o1, Figure: p1.Figure, Motto: p1.Motto},
	})
}

// notifyNewRequest sends a new friend request notification to one target user.
// Delivery goes through the broadcaster so the user's own pumpNotifications
// goroutine forwards it via their connection-local transport.
func (runtime *Runtime) notifyNewRequest(ctx context.Context, fromUserID, toUserID int) {
	profiles, err := runtime.service.GetUserProfiles(ctx, []int{fromUserID})
	if err != nil || len(profiles) == 0 {
		return
	}
	p := profiles[0]
	composer := packetrequest.MessengerNewRequestComposer{
		RequestID: int32(fromUserID), FromUsername: p.Username, FromFigure: p.Figure,
	}
	if session, online := runtime.sessions.FindByUserID(toUserID); online {
		if err = runtime.sendPacket(session.ConnID, composer); err == nil {
			runtime.pushPendingRequestsSnapshot(ctx, toUserID)
			return
		}
	}
	body, err := composer.Encode()
	if err != nil {
		return
	}
	frame := codec.EncodeFrame(composer.PacketID(), body)
	if err = runtime.service.NotifyFriendUpdate(ctx, toUserID, frame); err != nil {
		return
	}
	runtime.pushPendingRequestsSnapshot(ctx, toUserID)
}

// pushPendingRequestsSnapshot sends a fresh pending-requests payload to one user.
func (runtime *Runtime) pushPendingRequestsSnapshot(ctx context.Context, toUserID int) {
	requests, err := runtime.service.ListPendingRequests(ctx, toUserID)
	if err != nil {
		return
	}
	fromIDs := make([]int, 0, len(requests))
	for _, req := range requests {
		fromIDs = append(fromIDs, req.FromUserID)
	}
	profiles, _ := runtime.service.GetUserProfiles(ctx, fromIDs)
	type prof struct{ Username, Figure string }
	pm := make(map[int]prof, len(profiles))
	for _, p := range profiles {
		pm[p.ID] = prof{p.Username, p.Figure}
	}
	entries := make([]packetmsginit.RequestEntry, 0, len(requests))
	for _, req := range requests {
		p := pm[req.FromUserID]
		entries = append(entries, packetmsginit.RequestEntry{
			ID: int32(req.FromUserID), FromUsername: p.Username, FromFigure: p.Figure,
		})
	}
	composer := packetmsginit.MessengerRequestsComposer{Requests: entries}
	if session, online := runtime.sessions.FindByUserID(toUserID); online {
		if err = runtime.sendPacket(session.ConnID, composer); err == nil {
			return
		}
	}
	body, err := composer.Encode()
	if err != nil {
		return
	}
	_ = runtime.service.NotifyFriendUpdate(ctx, toUserID, codec.EncodeFrame(composer.PacketID(), body))
}

// handleRemoveFriend handles messenger.remove_friend.
func (runtime *Runtime) handleRemoveFriend(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packetrequest.MessengerRemoveFriendPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	cap := maxRemoveBatch
	for _, friendID := range pkt.UserIDs {
		if cap <= 0 {
			break
		}
		cap--
		if err := runtime.service.RemoveFriendship(ctx, userID, int(friendID)); err != nil {
			runtime.logger.Sugar().Warnw("remove friend failed", "conn", connID, "friend_id", friendID, "err", err)
			continue
		}
		fid := int(friendID)
		go runtime.publishFriendUpdate(ctx, userID, []packetrequest.FriendUpdateEntry{{Action: -1, FriendID: int32(fid)}})
		go runtime.publishFriendUpdate(ctx, fid, []packetrequest.FriendUpdateEntry{{Action: -1, FriendID: int32(userID)}})
	}
	return nil
}
