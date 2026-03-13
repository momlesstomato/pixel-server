package name

import "github.com/momlesstomato/pixel-server/core/codec"

// UserNameResultPacket defines user.check_name_result and user.change_name_result payload.
type UserNameResultPacket struct {
	// ResultCode stores result code payload.
	ResultCode int32
	// Name stores requested name payload.
	Name string
	// Suggestions stores fallback name suggestions.
	Suggestions []string
}

// PacketID returns protocol packet identifier.
func (packet UserNameResultPacket) PacketID() uint16 { return UserCheckNameResultPacketID }

// Encode serializes packet body payload.
func (packet UserNameResultPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.ResultCode)
	if err := writer.WriteString(packet.Name); err != nil {
		return nil, err
	}
	writer.WriteInt32(int32(len(packet.Suggestions)))
	for _, suggestion := range packet.Suggestions {
		if err := writer.WriteString(suggestion); err != nil {
			return nil, err
		}
	}
	return writer.Bytes(), nil
}
