package msginit

// MessengerInitPacket defines client messenger.init payload.
type MessengerInitPacket struct{}

// PacketID returns protocol packet identifier.
func (p MessengerInitPacket) PacketID() uint16 { return MessengerInitPacketID }

// Encode serializes packet body payload.
func (p MessengerInitPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (p *MessengerInitPacket) Decode(_ []byte) error { return nil }

// MessengerGetFriendsPacket defines client messenger.get_friends payload.
type MessengerGetFriendsPacket struct{}

// PacketID returns protocol packet identifier.
func (p MessengerGetFriendsPacket) PacketID() uint16 { return MessengerGetFriendsPacketID }

// Encode serializes packet body payload.
func (p MessengerGetFriendsPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (p *MessengerGetFriendsPacket) Decode(_ []byte) error { return nil }

// MessengerGetRequestsPacket defines client messenger.get_requests payload.
type MessengerGetRequestsPacket struct{}

// PacketID returns protocol packet identifier.
func (p MessengerGetRequestsPacket) PacketID() uint16 { return MessengerGetRequestsPacketID }

// Encode serializes packet body payload.
func (p MessengerGetRequestsPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (p *MessengerGetRequestsPacket) Decode(_ []byte) error { return nil }
