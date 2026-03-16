package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
	packetrequest "github.com/momlesstomato/pixel-server/pkg/messenger/packet/request"
	packetsocial "github.com/momlesstomato/pixel-server/pkg/messenger/packet/social"
)

// handleSearch handles messenger.search.
func (runtime *Runtime) handleSearch(ctx context.Context, connID string, body []byte) error {
	var pkt packetsocial.MessengerSearchPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	results, err := runtime.service.SearchUsers(ctx, pkt.Query, 50)
	if err != nil {
		return nil
	}
	userID, _ := runtime.userID(connID)
	friends := make([]packetsocial.SearchResultEntry, 0)
	others := make([]packetsocial.SearchResultEntry, 0)
	for _, r := range results {
		entry := packetsocial.SearchResultEntry{
			ID: int32(r.ID), Username: r.Username, Motto: r.Motto, Online: r.Online, Figure: r.Figure,
		}
		if userID > 0 {
			isFriend, _ := runtime.service.AreFriends(ctx, userID, r.ID)
			r.IsFriend = isFriend
		}
		if r.IsFriend {
			friends = append(friends, entry)
		} else {
			others = append(others, entry)
		}
	}
	return runtime.sendPacket(connID, packetsocial.MessengerSearchResultComposer{Friends: friends, Others: others})
}

// handleSetRelationship handles messenger.set_relationship.
func (runtime *Runtime) handleSetRelationship(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packetsocial.MessengerSetRelationshipPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	rel := domain.RelationshipType(pkt.RelType)
	if err := runtime.service.SetRelationship(ctx, userID, int(pkt.UserID), rel); err != nil {
		runtime.logger.Sugar().Warnw("set relationship failed", "conn", connID, "err", err)
		return nil
	}
	profiles, err := runtime.service.GetUserProfiles(ctx, []int{userID})
	if err != nil || len(profiles) == 0 {
		return nil
	}
	p := profiles[0]
	_, online := runtime.sessions.FindByUserID(userID)
	entry := packetrequest.FriendUpdateEntry{
		Action: 0, FriendID: int32(userID), Username: p.Username,
		Online: online, Figure: p.Figure, Motto: p.Motto,
		Relationship: int16(rel),
	}
	go runtime.publishFriendUpdate(ctx, int(pkt.UserID), []packetrequest.FriendUpdateEntry{entry})
	return nil
}

// handleGetRelationships handles messenger.get_relationships.
func (runtime *Runtime) handleGetRelationships(ctx context.Context, connID string, body []byte) error {
	var pkt packetsocial.MessengerGetRelationshipsPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	counts, err := runtime.service.GetRelationshipCounts(ctx, int(pkt.UserID))
	if err != nil {
		return nil
	}
	entries := make([]packetsocial.RelationshipEntry, 0, len(counts))
	for _, c := range counts {
		sampleIDs := make([]int32, 0, len(c.SampleUserIDs))
		for _, id := range c.SampleUserIDs {
			sampleIDs = append(sampleIDs, int32(id))
		}
		entries = append(entries, packetsocial.RelationshipEntry{
			Type: int32(c.Type), Count: int32(c.Count), SampleUserIDs: sampleIDs,
		})
	}
	return runtime.sendPacket(connID, packetsocial.MessengerRelationshipsComposer{
		UserID: pkt.UserID, Entries: entries,
	})
}
