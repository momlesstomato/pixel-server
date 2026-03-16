package social

import "github.com/momlesstomato/pixel-server/core/codec"

// RelationshipEntry defines one relationship count wire entry.
type RelationshipEntry struct {
	// Type stores the relationship type value.
	Type int32
	// Count stores the number of friends with this relationship.
	Count int32
	// SampleUserIDs stores up to three sample user identifiers.
	SampleUserIDs []int32
}

// MessengerRelationshipsComposer defines server messenger.relationships payload.
type MessengerRelationshipsComposer struct {
	// UserID stores the profile user identifier.
	UserID int32
	// Entries stores relationship count entries.
	Entries []RelationshipEntry
}

// PacketID returns protocol packet identifier.
func (p MessengerRelationshipsComposer) PacketID() uint16 {
	return MessengerRelationshipsComposerID
}

// Encode serializes packet body payload.
func (p MessengerRelationshipsComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.UserID)
	w.WriteInt32(int32(len(p.Entries)))
	for _, e := range p.Entries {
		w.WriteInt32(e.Type)
		w.WriteInt32(e.Count)
		w.WriteInt32(int32(len(e.SampleUserIDs)))
		for _, id := range e.SampleUserIDs {
			w.WriteInt32(id)
		}
	}
	return w.Bytes(), nil
}

// MessengerFollowFailedComposer defines server messenger.follow_failed payload.
type MessengerFollowFailedComposer struct {
	// ErrorCode stores 0=not friend, 1=offline, 2=not in room, 3=blocked.
	ErrorCode int32
}

// PacketID returns protocol packet identifier.
func (p MessengerFollowFailedComposer) PacketID() uint16 {
	return MessengerFollowFailedComposerID
}

// Encode serializes packet body payload.
func (p MessengerFollowFailedComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.ErrorCode)
	return w.Bytes(), nil
}

// MessengerRoomInviteComposer defines server messenger.room_invite payload.
type MessengerRoomInviteComposer struct {
	// SenderID stores the inviting user identifier.
	SenderID int32
	// Message stores the invite message.
	Message string
}

// PacketID returns protocol packet identifier.
func (p MessengerRoomInviteComposer) PacketID() uint16 {
	return MessengerRoomInviteComposerID
}

// Encode serializes packet body payload.
func (p MessengerRoomInviteComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.SenderID)
	return w.Bytes(), w.WriteString(p.Message)
}
