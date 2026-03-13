package hotel

import "github.com/momlesstomato/pixel-server/core/codec"

// MaintenancePacketID identifies hotel.maintenance packet.
const MaintenancePacketID uint16 = 1350

// MaintenancePacket carries maintenance status details.
type MaintenancePacket struct {
	// IsInMaintenance stores whether hotel is in maintenance mode.
	IsInMaintenance bool
	// MinutesUntilChange stores minutes until maintenance starts or ends.
	MinutesUntilChange int32
	// Duration stores expected maintenance duration in minutes.
	Duration int32
}

// PacketID returns protocol packet identifier.
func (packet MaintenancePacket) PacketID() uint16 { return MaintenancePacketID }

// Decode parses packet body into struct fields.
func (packet *MaintenancePacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	isInMaintenance, err := reader.ReadBool()
	if err != nil {
		return err
	}
	minutesUntilChange, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	duration, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	packet.IsInMaintenance = isInMaintenance
	packet.MinutesUntilChange = minutesUntilChange
	packet.Duration = duration
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet MaintenancePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteBool(packet.IsInMaintenance)
	writer.WriteInt32(packet.MinutesUntilChange)
	writer.WriteInt32(packet.Duration)
	return writer.Bytes(), nil
}
