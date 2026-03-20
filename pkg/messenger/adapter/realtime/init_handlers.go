package realtime

import (
	"context"
	"time"

	packetmessage "github.com/momlesstomato/pixel-server/pkg/messenger/packet/message"
	packetmsginit "github.com/momlesstomato/pixel-server/pkg/messenger/packet/msginit"
)

// handleInit handles messenger.init - sends init, friends fragments, requests, delivers offline, subscribes.
func (runtime *Runtime) handleInit(ctx context.Context, connID string, userID int) error {
	resolved := runtime.service.ResolvedFriendLimit(ctx, userID)
	if resolved == 0 {
		resolved = runtime.service.Config().MaxFriendsVIP
	}
	initComposer := packetmsginit.MessengerInitComposer{
		UserFriendLimit: int32(resolved),
		NormalLimit:     int32(runtime.service.Config().MaxFriends),
		ExtendedLimit:   int32(runtime.service.Config().MaxFriendsVIP),
	}
	if err := runtime.sendPacket(connID, initComposer); err != nil {
		return err
	}
	offline, err := runtime.service.DeliverOfflineMessages(ctx, userID)
	if err != nil {
		return err
	}
	senders := make(map[int]bool, len(offline))
	for _, msg := range offline {
		senders[msg.FromUserID] = true
	}
	if err := runtime.sendFriendFragments(ctx, connID, userID, senders); err != nil {
		return err
	}
	for _, msg := range offline {
		seconds := int32(time.Since(msg.SentAt).Seconds())
		if seconds < 0 {
			seconds = 0
		}
		composer := packetmessage.MessengerNewMessageComposer{
			SenderID: int32(msg.FromUserID), Message: msg.Message, SecondsSinceSent: seconds,
		}
		if err := runtime.sendPacket(connID, composer); err != nil {
			runtime.logger.Sugar().Warnw("offline msg delivery failed", "conn", connID, "err", err)
		}
	}
	go runtime.notifyFriendsStatus(userID, true)
	return nil
}

// handleGetFriends handles messenger.get_friends - resends friend list fragments.
func (runtime *Runtime) handleGetFriends(ctx context.Context, connID string, userID int) error {
	return runtime.sendFriendFragments(ctx, connID, userID, nil)
}

// handleGetRequests handles messenger.get_requests - resends pending requests.
func (runtime *Runtime) handleGetRequests(ctx context.Context, connID string, userID int) error {
	return runtime.sendPendingRequests(ctx, connID, userID)
}

// sendFriendFragments sends all friend list fragment packets for one user.
// persistedSenders marks which friend user IDs have offline messages waiting; nil means none.
func (runtime *Runtime) sendFriendFragments(ctx context.Context, connID string, userID int, persistedSenders map[int]bool) error {
	friends, err := runtime.service.ListFriends(ctx, userID)
	if err != nil {
		return err
	}
	friendIDs := make([]int, 0, len(friends))
	for _, f := range friends {
		friendIDs = append(friendIDs, f.UserTwoID)
	}
	profiles, _ := runtime.service.GetUserProfiles(ctx, friendIDs)
	type prof struct{ Username, Figure, Motto string }
	pm := make(map[int]prof, len(profiles))
	for _, p := range profiles {
		pm[p.ID] = prof{p.Username, p.Figure, p.Motto}
	}
	size := runtime.service.Config().FragmentSize
	entries := make([]packetmsginit.FriendEntry, 0, len(friends))
	for _, f := range friends {
		_, online := runtime.sessions.FindByUserID(f.UserTwoID)
		p := pm[f.UserTwoID]
		entries = append(entries, packetmsginit.FriendEntry{
			ID: int32(f.UserTwoID), Username: p.Username, Figure: p.Figure,
			Motto: p.Motto, Online: online,
			Relationship:     packetmsginit.MapRelationship(f.Relationship),
			PersistedMessage: persistedSenders[f.UserTwoID],
		})
	}
	total := (len(entries) + size - 1) / size
	if total == 0 {
		total = 1
	}
	for i := 0; i < total; i++ {
		start := i * size
		end := start + size
		if end > len(entries) {
			end = len(entries)
		}
		composer := packetmsginit.MessengerFriendsComposer{
			TotalFragments: int32(total),
			FragmentNumber: int32(i),
			Friends:        entries[start:end],
		}
		if err := runtime.sendPacket(connID, composer); err != nil {
			return err
		}
	}
	return nil
}

// sendPendingRequests sends the pending friend requests packet for one user.
func (runtime *Runtime) sendPendingRequests(ctx context.Context, connID string, userID int) error {
	requests, err := runtime.service.ListPendingRequests(ctx, userID)
	if err != nil {
		return err
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
	return runtime.sendPacket(connID, packetmsginit.MessengerRequestsComposer{Requests: entries})
}

