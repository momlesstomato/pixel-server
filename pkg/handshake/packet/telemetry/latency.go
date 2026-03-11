package telemetry

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/core/codec"
)

// ClientLatencyTestPacketID identifies client.latency_test packet.
const ClientLatencyTestPacketID uint16 = 295

// ClientLatencyResponsePacketID identifies client.latency_response packet.
const ClientLatencyResponsePacketID uint16 = 10

// ClientLatencyTestPacket carries a latency measurement request id.
type ClientLatencyTestPacket struct {
	// RequestID stores latency request identifier.
	RequestID int32
}

// ClientLatencyResponsePacket carries a latency response request id.
type ClientLatencyResponsePacket struct {
	// RequestID stores latency response identifier.
	RequestID int32
}

// PacketID returns protocol packet id.
func (packet ClientLatencyTestPacket) PacketID() uint16 { return ClientLatencyTestPacketID }

// PacketID returns protocol packet id.
func (packet ClientLatencyResponsePacket) PacketID() uint16 { return ClientLatencyResponsePacketID }

// Decode parses packet body into struct fields.
func (packet *ClientLatencyTestPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	requestID, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	if reader.Remaining() != 0 {
		return fmt.Errorf("client.latency_test body has %d trailing bytes", reader.Remaining())
	}
	packet.RequestID = requestID
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ClientLatencyTestPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.RequestID)
	return writer.Bytes(), nil
}

// Decode parses packet body into struct fields.
func (packet *ClientLatencyResponsePacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	requestID, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	if reader.Remaining() != 0 {
		return fmt.Errorf("client.latency_response body has %d trailing bytes", reader.Remaining())
	}
	packet.RequestID = requestID
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ClientLatencyResponsePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.RequestID)
	return writer.Bytes(), nil
}
