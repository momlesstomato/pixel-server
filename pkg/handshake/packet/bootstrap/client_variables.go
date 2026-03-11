package bootstrap

import "github.com/momlesstomato/pixel-server/core/codec"

// ClientVariablesPacketID identifies handshake.client_variables packet.
const ClientVariablesPacketID uint16 = 1053

// ClientVariablesPacket carries client bootstrap URL metadata.
type ClientVariablesPacket struct {
	// ClientID stores client numeric identifier.
	ClientID int32
	// ClientURL stores runtime client URL.
	ClientURL string
	// ExternalVariablesURL stores external variables URL.
	ExternalVariablesURL string
}

// PacketID returns protocol packet id.
func (packet ClientVariablesPacket) PacketID() uint16 { return ClientVariablesPacketID }

// Decode parses packet body into struct fields.
func (packet *ClientVariablesPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	clientID, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	clientURL, err := reader.ReadString()
	if err != nil {
		return err
	}
	externalVariablesURL, err := reader.ReadString()
	if err != nil {
		return err
	}
	packet.ClientID = clientID
	packet.ClientURL = clientURL
	packet.ExternalVariablesURL = externalVariablesURL
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ClientVariablesPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.ClientID)
	if err := writer.WriteString(packet.ClientURL); err != nil {
		return nil, err
	}
	if err := writer.WriteString(packet.ExternalVariablesURL); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}
