package codec

import "encoding/binary"

// Frame is one protocol packet decoded from a websocket frame payload.
type Frame struct {
	// Length is payload byte size excluding the outer uint32 length prefix.
	Length uint32
	// Header is the uint16 message identifier.
	Header uint16
	// Payload is the packet payload body excluding the header.
	Payload []byte
	// Body is the packet bytes including header and payload.
	Body []byte
}

// EncodeFrame builds one wire frame from header and payload.
func EncodeFrame(header uint16, payload []byte) []byte {
	bodyLength := 2 + len(payload)
	out := make([]byte, 4+bodyLength)
	binary.BigEndian.PutUint32(out[:4], uint32(bodyLength))
	binary.BigEndian.PutUint16(out[4:6], header)
	copy(out[6:], payload)
	return out
}

// SplitFrames parses one websocket binary payload that can contain multiple packets.
func SplitFrames(raw []byte) ([]Frame, error) {
	frames := make([]Frame, 0, 1)
	offset := 0
	for offset < len(raw) {
		if len(raw)-offset < 4 {
			return nil, ErrInvalidFrame
		}
		bodyLength := int(binary.BigEndian.Uint32(raw[offset : offset+4]))
		offset += 4
		if bodyLength < 2 || len(raw)-offset < bodyLength {
			return nil, ErrInvalidFrame
		}
		body := raw[offset : offset+bodyLength]
		header := binary.BigEndian.Uint16(body[:2])
		frames = append(frames, Frame{
			Length:  uint32(bodyLength),
			Header:  header,
			Payload: body[2:],
			Body:    body,
		})
		offset += bodyLength
	}
	return frames, nil
}
