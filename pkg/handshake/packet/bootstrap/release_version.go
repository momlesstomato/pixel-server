package bootstrap

import "github.com/momlesstomato/pixel-server/core/codec"

// ReleaseVersionPacketID identifies handshake.release_version packet.
const ReleaseVersionPacketID uint16 = 4000

// ReleaseVersionPacket carries client release metadata.
type ReleaseVersionPacket struct {
	// ReleaseVersion stores client release identifier.
	ReleaseVersion string
	// ClientType stores runtime client type value.
	ClientType string
	// Platform stores platform numeric code.
	Platform int32
	// DeviceCategory stores device category numeric code.
	DeviceCategory int32
}

// PacketID returns protocol packet id.
func (packet ReleaseVersionPacket) PacketID() uint16 { return ReleaseVersionPacketID }

// Decode parses packet body into struct fields.
func (packet *ReleaseVersionPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	releaseVersion, err := reader.ReadString()
	if err != nil {
		return err
	}
	clientType, err := reader.ReadString()
	if err != nil {
		return err
	}
	platform, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	deviceCategory, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	packet.ReleaseVersion = releaseVersion
	packet.ClientType = clientType
	packet.Platform = platform
	packet.DeviceCategory = deviceCategory
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ReleaseVersionPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.ReleaseVersion); err != nil {
		return nil, err
	}
	if err := writer.WriteString(packet.ClientType); err != nil {
		return nil, err
	}
	writer.WriteInt32(packet.Platform)
	writer.WriteInt32(packet.DeviceCategory)
	return writer.Bytes(), nil
}
