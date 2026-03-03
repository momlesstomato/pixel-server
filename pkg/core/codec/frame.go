package codec

import (
	"encoding/binary"
	"fmt"
)

// Packet is the interface all generated packet structs implement.
type Packet interface {
	// HeaderID returns the uint16 message identifier for this packet.
	HeaderID() uint16
	// Encode writes the packet's fields to a Writer.
	Encode(w *Writer) error
}

// ParseFrame reads a single framed message from raw bytes and returns the
// header ID and the payload slice. It does not copy the payload.
//
// Wire format:
//
//	[4 bytes: uint32 length] [2 bytes: uint16 headerID] [N bytes: payload]
//
// The length field equals 2 + len(payload).
func ParseFrame(data []byte) (headerID uint16, payload []byte, rest []byte, err error) {
	if len(data) < 4 {
		return 0, nil, data, fmt.Errorf("codec: frame too short for length prefix: %d bytes", len(data))
	}
	frameLen := int(binary.BigEndian.Uint32(data[0:4]))
	if frameLen < 2 {
		return 0, nil, data, fmt.Errorf("codec: frame length %d too small (minimum 2)", frameLen)
	}
	totalLen := 4 + frameLen
	if len(data) < totalLen {
		return 0, nil, data, fmt.Errorf("codec: frame incomplete: need %d bytes, have %d", totalLen, len(data))
	}
	headerID = binary.BigEndian.Uint16(data[4:6])
	payload = data[6:totalLen]
	rest = data[totalLen:]
	return headerID, payload, rest, nil
}

// ParseFrames splits a concatenated buffer into individual frames.
// Multiple packets can be concatenated inside a single WebSocket message.
func ParseFrames(data []byte) (frames []RawFrame, err error) {
	for len(data) > 0 {
		headerID, payload, rest, err := ParseFrame(data)
		if err != nil {
			return frames, err
		}
		frames = append(frames, RawFrame{HeaderID: headerID, Payload: payload})
		data = rest
	}
	return frames, nil
}

// RawFrame holds a parsed but undecoded frame.
type RawFrame struct {
	HeaderID uint16
	Payload  []byte
}
