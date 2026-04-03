package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// OpenFlatConnectionPacket decodes room entry request (c2s 189).
type OpenFlatConnectionPacket struct {
	// RoomID stores the requested room identifier.
	RoomID int32
	// Password stores the optional room password.
	Password string
}

// PacketID returns the protocol packet identifier.
func (p OpenFlatConnectionPacket) PacketID() uint16 { return OpenFlatConnectionPacketID }

// Decode parses packet body payload.
func (p *OpenFlatConnectionPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	roomID, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.RoomID = roomID
	pwd, err := r.ReadString()
	if err != nil {
		p.Password = ""
		return nil
	}
	p.Password = pwd
	return nil
}

// Encode serializes packet body payload.
func (p OpenFlatConnectionPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.RoomID)
	if err := w.WriteString(p.Password); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// OpenConnectionComposer acknowledges room connection (s2c 3566).
type OpenConnectionComposer struct{}

// PacketID returns the protocol packet identifier.
func (p OpenConnectionComposer) PacketID() uint16 { return OpenConnectionComposerID }

// Encode serializes packet body.
func (p OpenConnectionComposer) Encode() ([]byte, error) { return []byte{}, nil }

// RoomReadyComposer sends room model and ID (s2c 768).
type RoomReadyComposer struct {
	// ModelSlug stores the room model identifier string.
	ModelSlug string
	// RoomID stores the room identifier.
	RoomID int32
}

// PacketID returns the protocol packet identifier.
func (p RoomReadyComposer) PacketID() uint16 { return RoomReadyComposerID }

// Encode serializes the room ready response.
func (p RoomReadyComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := w.WriteString(p.ModelSlug); err != nil {
		return nil, err
	}
	w.WriteInt32(p.RoomID)
	return w.Bytes(), nil
}
