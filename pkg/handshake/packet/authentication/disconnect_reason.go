package authentication

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/core/codec"
)

// DisconnectReasonPacketID identifies handshake.disconnect_reason packet.
const DisconnectReasonPacketID uint16 = 4000

// DisconnectReasonConcurrentLogin identifies concurrent login disconnect reason.
const DisconnectReasonConcurrentLogin int32 = 2

// DisconnectReasonLogout identifies user-initiated logout disconnect reason.
const DisconnectReasonLogout int32 = 0

// DisconnectReasonJustBanned identifies live-session ban disconnect reason.
const DisconnectReasonJustBanned int32 = 1

// DisconnectReasonStillBanned identifies login rejected due active ban reason.
const DisconnectReasonStillBanned int32 = 10

// DisconnectReasonHotelClosed identifies hotel closed disconnect reason.
const DisconnectReasonHotelClosed int32 = 12

// DisconnectReasonHotelClosing identifies server-initiated shutdown disconnect reason.
const DisconnectReasonHotelClosing int32 = 19

// DisconnectReasonInvalidLoginTicket identifies invalid login ticket disconnect reason.
const DisconnectReasonInvalidLoginTicket int32 = 22

// DisconnectReasonPongTimeout identifies heartbeat pong timeout disconnect reason.
const DisconnectReasonPongTimeout int32 = 113

// DisconnectReasonIdleNotAuthenticated identifies auth-timeout disconnect reason.
const DisconnectReasonIdleNotAuthenticated int32 = 114

// DisconnectReasonPacket carries one structured disconnect reason code.
type DisconnectReasonPacket struct {
	// Reason stores disconnect reason code consumed by Nitro.
	Reason int32
}

// PacketID returns protocol packet id.
func (packet DisconnectReasonPacket) PacketID() uint16 { return DisconnectReasonPacketID }

// Decode parses packet body into struct fields.
func (packet *DisconnectReasonPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	reason, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	if reader.Remaining() != 0 {
		return fmt.Errorf("disconnect_reason body has %d trailing bytes", reader.Remaining())
	}
	packet.Reason = reason
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet DisconnectReasonPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.Reason)
	return writer.Bytes(), nil
}
