package security

import "github.com/momlesstomato/pixel-server/core/codec"

// ClientMachineIDPacketID identifies security.machine_id C2S packet.
const ClientMachineIDPacketID uint16 = 2490

// ServerMachineIDPacketID identifies security.machine_id S2C packet.
const ServerMachineIDPacketID uint16 = 1488

// ClientMachineIDPacket carries machine fingerprint metadata.
type ClientMachineIDPacket struct {
	// MachineID stores machine identifier hash.
	MachineID string
	// Fingerprint stores browser/device fingerprint hash.
	Fingerprint string
	// Capabilities stores capabilities string.
	Capabilities string
}

// ServerMachineIDPacket carries machine identifier echo payload.
type ServerMachineIDPacket struct {
	// MachineID stores machine identifier hash.
	MachineID string
}

// PacketID returns protocol packet id.
func (packet ClientMachineIDPacket) PacketID() uint16 { return ClientMachineIDPacketID }

// PacketID returns protocol packet id.
func (packet ServerMachineIDPacket) PacketID() uint16 { return ServerMachineIDPacketID }

// Decode parses C2S packet body into struct fields.
func (packet *ClientMachineIDPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	machineID, err := reader.ReadString()
	if err != nil {
		return err
	}
	fingerprint, err := reader.ReadString()
	if err != nil {
		return err
	}
	capabilities, err := reader.ReadString()
	if err != nil {
		return err
	}
	packet.MachineID = machineID
	packet.Fingerprint = fingerprint
	packet.Capabilities = capabilities
	return nil
}

// Encode serializes C2S packet fields into protocol body bytes.
func (packet ClientMachineIDPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.MachineID); err != nil {
		return nil, err
	}
	if err := writer.WriteString(packet.Fingerprint); err != nil {
		return nil, err
	}
	if err := writer.WriteString(packet.Capabilities); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// Decode parses S2C packet body into struct fields.
func (packet *ServerMachineIDPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	machineID, err := reader.ReadString()
	if err != nil {
		return err
	}
	packet.MachineID = machineID
	return nil
}

// Encode serializes S2C packet fields into protocol body bytes.
func (packet ServerMachineIDPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.MachineID); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}
