package protocol

import (
	"errors"

	"pixelsv/pkg/codec"
)

// ErrUnknownHeader indicates a packet header has no registered decoder.
var ErrUnknownHeader = errors.New("unknown packet header")

// Decoder decodes one packet payload body into a typed packet.
type Decoder func(reader *codec.Reader) (Packet, error)

// Definition describes one packet contract in the decoder registry.
type Definition struct {
	// ID is the uint16 packet header identifier.
	ID uint16
	// Name is the packet canonical name.
	Name string
	// Realm is the packet realm identifier.
	Realm string
	// Summary describes packet behavior.
	Summary string
	// Decode decodes packet payload bytes.
	Decode Decoder
}

// Packet is the common interface for generated protocol packet structs.
type Packet interface {
	// HeaderID returns packet header identifier.
	HeaderID() uint16
	// PacketName returns canonical packet name.
	PacketName() string
	// Realm returns packet realm identifier.
	Realm() string
	// Encode writes packet payload to codec writer.
	Encode(writer *codec.Writer) error
}

var c2sRegistry = map[uint16]Definition{}

// LookupC2S returns packet metadata and decoder for one client packet header.
func LookupC2S(header uint16) (Definition, bool) {
	def, ok := c2sRegistry[header]
	return def, ok
}

// DecodeC2S decodes a client packet payload using the header registry.
func DecodeC2S(header uint16, payload []byte) (Packet, error) {
	def, ok := LookupC2S(header)
	if !ok {
		return nil, ErrUnknownHeader
	}
	return def.Decode(codec.NewReader(payload))
}

func registerC2S(definition Definition) {
	c2sRegistry[definition.ID] = definition
}
