package ignore

import "github.com/momlesstomato/pixel-server/core/codec"

// UserIgnoreResultPacket defines user.ignore_result packet payload.
type UserIgnoreResultPacket struct {
	// Result stores operation result code.
	Result int32
	// Name stores target username payload.
	Name string
}

// PacketID returns protocol packet identifier.
func (packet UserIgnoreResultPacket) PacketID() uint16 { return UserIgnoreResultPacketID }

// Encode serializes packet body payload.
func (packet UserIgnoreResultPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.Result)
	if err := writer.WriteString(packet.Name); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}
