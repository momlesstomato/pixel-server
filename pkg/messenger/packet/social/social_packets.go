package social

import "github.com/momlesstomato/pixel-server/core/codec"

// MessengerSearchPacket defines client messenger.search payload.
type MessengerSearchPacket struct {
	// Query stores the search query string.
	Query string
}

// PacketID returns protocol packet identifier.
func (p MessengerSearchPacket) PacketID() uint16 { return MessengerSearchPacketID }

// Encode serializes packet body payload.
func (p MessengerSearchPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (p *MessengerSearchPacket) Decode(payload []byte) error {
	val, err := codec.NewReader(payload).ReadString()
	if err == nil {
		p.Query = val
	}
	return err
}

// MessengerSetRelationshipPacket defines client messenger.set_relationship payload.
type MessengerSetRelationshipPacket struct {
	// UserID stores the target friend user identifier.
	UserID int32
	// RelType stores the new relationship type value.
	RelType int32
}

// PacketID returns protocol packet identifier.
func (p MessengerSetRelationshipPacket) PacketID() uint16 { return MessengerSetRelationshipPacketID }

// Encode serializes packet body payload.
func (p MessengerSetRelationshipPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (p *MessengerSetRelationshipPacket) Decode(payload []byte) error {
	r := codec.NewReader(payload)
	userID, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.UserID = userID
	relType, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.RelType = relType
	return nil
}

// MessengerGetRelationshipsPacket defines client messenger.get_relationships payload.
type MessengerGetRelationshipsPacket struct {
	// UserID stores the target user identifier to view relationships for.
	UserID int32
}

// PacketID returns protocol packet identifier.
func (p MessengerGetRelationshipsPacket) PacketID() uint16 { return MessengerGetRelationshipsPacketID }

// Encode serializes packet body payload.
func (p MessengerGetRelationshipsPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (p *MessengerGetRelationshipsPacket) Decode(payload []byte) error {
	val, err := codec.NewReader(payload).ReadInt32()
	if err == nil {
		p.UserID = val
	}
	return err
}

// MessengerFollowFriendPacket defines client messenger.follow_friend payload.
type MessengerFollowFriendPacket struct {
	// FriendID stores the friend user identifier to follow.
	FriendID int32
}

// PacketID returns protocol packet identifier.
func (p MessengerFollowFriendPacket) PacketID() uint16 { return MessengerFollowFriendPacketID }

// Encode serializes packet body payload.
func (p MessengerFollowFriendPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (p *MessengerFollowFriendPacket) Decode(payload []byte) error {
	val, err := codec.NewReader(payload).ReadInt32()
	if err == nil {
		p.FriendID = val
	}
	return err
}

// MessengerSendInvitePacket defines client messenger.send_invite payload.
type MessengerSendInvitePacket struct {
	// UserIDs stores the recipient user identifiers.
	UserIDs []int32
	// Message stores the invite message.
	Message string
}

// PacketID returns protocol packet identifier.
func (p MessengerSendInvitePacket) PacketID() uint16 { return MessengerSendInvitePacketID }

// Encode serializes packet body payload.
func (p MessengerSendInvitePacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (p *MessengerSendInvitePacket) Decode(payload []byte) error {
	r := codec.NewReader(payload)
	count, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.UserIDs = make([]int32, 0, count)
	for i := int32(0); i < count; i++ {
		id, err := r.ReadInt32()
		if err != nil {
			return err
		}
		p.UserIDs = append(p.UserIDs, id)
	}
	msg, err := r.ReadString()
	if err != nil {
		return err
	}
	p.Message = msg
	return nil
}
