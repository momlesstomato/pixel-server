package request

import "github.com/momlesstomato/pixel-server/core/codec"

// MessengerSendRequestPacket defines client messenger.send_request payload.
type MessengerSendRequestPacket struct {
	// Username stores the target username.
	Username string
}

// PacketID returns protocol packet identifier.
func (p MessengerSendRequestPacket) PacketID() uint16 { return MessengerSendRequestPacketID }

// Encode serializes packet body payload.
func (p MessengerSendRequestPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (p *MessengerSendRequestPacket) Decode(payload []byte) error {
	val, err := codec.NewReader(payload).ReadString()
	if err == nil {
		p.Username = val
	}
	return err
}

// MessengerAcceptFriendPacket defines client messenger.accept_friend payload.
type MessengerAcceptFriendPacket struct {
	// RequestIDs stores the request identifiers to accept.
	RequestIDs []int32
}

// PacketID returns protocol packet identifier.
func (p MessengerAcceptFriendPacket) PacketID() uint16 { return MessengerAcceptFriendPacketID }

// Encode serializes packet body payload.
func (p MessengerAcceptFriendPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (p *MessengerAcceptFriendPacket) Decode(payload []byte) error {
	r := codec.NewReader(payload)
	count, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.RequestIDs = make([]int32, 0, count)
	for i := int32(0); i < count; i++ {
		id, err := r.ReadInt32()
		if err != nil {
			return err
		}
		p.RequestIDs = append(p.RequestIDs, id)
	}
	return nil
}

// MessengerDeclineFriendPacket defines client messenger.decline_friend payload.
type MessengerDeclineFriendPacket struct {
	// DeclineAll stores whether all requests should be declined.
	DeclineAll bool
	// RequestIDs stores specific request identifiers when DeclineAll is false.
	RequestIDs []int32
}

// PacketID returns protocol packet identifier.
func (p MessengerDeclineFriendPacket) PacketID() uint16 { return MessengerDeclineFriendPacketID }

// Encode serializes packet body payload.
func (p MessengerDeclineFriendPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (p *MessengerDeclineFriendPacket) Decode(payload []byte) error {
	r := codec.NewReader(payload)
	declineAll, err := r.ReadBool()
	if err != nil {
		return err
	}
	p.DeclineAll = declineAll
	count, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.RequestIDs = make([]int32, 0, count)
	for i := int32(0); i < count; i++ {
		id, err := r.ReadInt32()
		if err != nil {
			return err
		}
		p.RequestIDs = append(p.RequestIDs, id)
	}
	return nil
}

// MessengerRemoveFriendPacket defines client messenger.remove_friend payload.
type MessengerRemoveFriendPacket struct {
	// UserIDs stores the friend user identifiers to remove.
	UserIDs []int32
}

// PacketID returns protocol packet identifier.
func (p MessengerRemoveFriendPacket) PacketID() uint16 { return MessengerRemoveFriendPacketID }

// Encode serializes packet body payload.
func (p MessengerRemoveFriendPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (p *MessengerRemoveFriendPacket) Decode(payload []byte) error {
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
	return nil
}
