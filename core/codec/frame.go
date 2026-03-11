package codec

import (
	"encoding/binary"
	"fmt"
)

// Frame represents a decoded protocol frame.
type Frame struct {
	// Length stores packet-id plus body size.
	Length uint32
	// PacketID stores the protocol packet identifier.
	PacketID uint16
	// Body stores packet payload bytes.
	Body []byte
}

// EncodeFrame serializes packet ID and body into protocol wire format.
func EncodeFrame(packetID uint16, body []byte) []byte {
	payloadLength := uint32(len(body) + 2)
	buffer := make([]byte, 6+len(body))
	binary.BigEndian.PutUint32(buffer[0:4], payloadLength)
	binary.BigEndian.PutUint16(buffer[4:6], packetID)
	copy(buffer[6:], body)
	return buffer
}

// DecodeFrame parses one frame and returns consumed byte count.
func DecodeFrame(data []byte) (Frame, int, error) {
	if len(data) < 6 {
		return Frame{}, 0, fmt.Errorf("incomplete frame header")
	}
	length := binary.BigEndian.Uint32(data[0:4])
	if length < 2 {
		return Frame{}, 0, fmt.Errorf("invalid frame length %d", length)
	}
	totalLength := int(4 + length)
	if len(data) < totalLength {
		return Frame{}, 0, fmt.Errorf("incomplete frame payload")
	}
	packetID := binary.BigEndian.Uint16(data[4:6])
	body := make([]byte, int(length)-2)
	copy(body, data[6:totalLength])
	return Frame{Length: length, PacketID: packetID, Body: body}, totalLength, nil
}

// DecodeFrames parses concatenated frames from a single transport payload.
func DecodeFrames(data []byte) ([]Frame, error) {
	frames := make([]Frame, 0, 2)
	offset := 0
	for offset < len(data) {
		frame, consumed, err := DecodeFrame(data[offset:])
		if err != nil {
			return nil, err
		}
		offset += consumed
		frames = append(frames, frame)
	}
	return frames, nil
}
