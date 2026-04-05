package packet

import (
	"strings"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
)

// FloorHeightMapComposer sends raw heightmap (s2c 1819).
type FloorHeightMapComposer struct {
	// Scale reports whether the heightmap uses full-scale rendering.
	Scale bool
	// WallHeight stores the fixed wall height override (-1 for auto).
	WallHeight int32
	// Heightmap stores the raw heightmap string with CR row separators.
	Heightmap string
}

// PacketID returns the protocol packet identifier.
func (p FloorHeightMapComposer) PacketID() uint16 { return FloorHeightMapComposerID }

// Encode serializes the floor heightmap response.
func (p FloorHeightMapComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteBool(p.Scale)
	w.WriteInt32(p.WallHeight)
	if err := w.WriteString(p.Heightmap); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// HeightMapComposer sends stacking height array (s2c 1232).
type HeightMapComposer struct {
	// Width stores the grid column count.
	Width int32
	// TotalTiles stores the total number of tiles.
	TotalTiles int32
	// Heights stores the stacking height short array.
	Heights []int16
}

// PacketID returns the protocol packet identifier.
func (p HeightMapComposer) PacketID() uint16 { return HeightMapComposerID }

// Encode serializes the stacking heightmap response.
func (p HeightMapComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.Width)
	w.WriteInt32(p.TotalTiles)
	for _, h := range p.Heights {
		w.WriteUint16(uint16(h))
	}
	return w.Bytes(), nil
}

// RoomEntryInfoComposer sends room ownership and ID (s2c 3675).
type RoomEntryInfoComposer struct {
	// RoomID stores the room identifier.
	RoomID int32
	// IsOwner reports whether the recipient owns the room.
	IsOwner bool
}

// PacketID returns the protocol packet identifier.
func (p RoomEntryInfoComposer) PacketID() uint16 { return RoomEntryInfoComposerID }

// Encode serializes the room entry info.
func (p RoomEntryInfoComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.RoomID)
	w.WriteBool(p.IsOwner)
	return w.Bytes(), nil
}

// RoomVisualizationComposer sends wall/floor settings (s2c 3003).
type RoomVisualizationComposer struct {
	// WallsHidden reports whether walls are hidden.
	WallsHidden bool
	// WallThickness stores wall rendering thickness.
	WallThickness int32
	// FloorThickness stores floor rendering thickness.
	FloorThickness int32
}

// PacketID returns the protocol packet identifier.
func (p RoomVisualizationComposer) PacketID() uint16 { return RoomVisualizationComposerID }

// Encode serializes room visualization settings.
func (p RoomVisualizationComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteBool(p.WallsHidden)
	w.WriteInt32(p.WallThickness)
	w.WriteInt32(p.FloorThickness)
	return w.Bytes(), nil
}

// FurnitureAliasesComposer sends empty furniture alias map (s2c 2159).
type FurnitureAliasesComposer struct{}

// PacketID returns the protocol packet identifier.
func (p FurnitureAliasesComposer) PacketID() uint16 { return FurnitureAliasesComposerID }

// Encode serializes empty furniture aliases.
func (p FurnitureAliasesComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(0)
	return w.Bytes(), nil
}

// CantConnectComposer notifies room entry failure (s2c 200).
type CantConnectComposer struct {
	// ErrorCode stores the failure reason code.
	ErrorCode int32
}

// PacketID returns the protocol packet identifier.
func (p CantConnectComposer) PacketID() uint16 { return CantConnectComposerID }

// Encode serializes the error code.
func (p CantConnectComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.ErrorCode)
	return w.Bytes(), nil
}

// GetRoomSettingsPacket decodes client room settings request (c2s 3700).
type GetRoomSettingsPacket struct {
	// RoomID stores the room identifier.
	RoomID int32
}

// PacketID returns the protocol packet identifier.
func (p GetRoomSettingsPacket) PacketID() uint16 { return GetRoomSettingsPacketID }

// Decode parses packet body payload.
func (p *GetRoomSettingsPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	id, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.RoomID = id
	return nil
}

// RoomSettingsSavedComposer confirms room settings were saved (s2c 539).
type RoomSettingsSavedComposer struct {
	// RoomID stores the updated room identifier.
	RoomID int32
}

// PacketID returns the protocol packet identifier.
func (p RoomSettingsSavedComposer) PacketID() uint16 { return RoomSettingsSavedComposerID }

// Encode serializes the save confirmation.
func (p RoomSettingsSavedComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.RoomID)
	return w.Bytes(), nil
}

// RoomSettingsComposer sends full room settings to the owner (s2c 3075).
type RoomSettingsComposer struct {
	// Room stores the room aggregate to serialize.
	Room domain.Room
}

// PacketID returns the protocol packet identifier.
func (p RoomSettingsComposer) PacketID() uint16 { return RoomSettingsComposerID }

// Encode serializes the room settings payload.
func (p RoomSettingsComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(p.Room.ID))
	if err := w.WriteString(p.Room.Name); err != nil {
		return nil, err
	}
	if err := w.WriteString(p.Room.Description); err != nil {
		return nil, err
	}
	w.WriteInt32(accessStateToInt(p.Room.State))
	if err := w.WriteString(p.Room.Password); err != nil {
		return nil, err
	}
	w.WriteInt32(int32(p.Room.MaxUsers))
	w.WriteInt32(int32(p.Room.CategoryID))
	if err := w.WriteString(strings.Join(p.Room.Tags, ",")); err != nil {
		return nil, err
	}
	w.WriteInt32(int32(p.Room.TradeMode))
	w.WriteBool(p.Room.AllowPets)
	w.WriteBool(p.Room.AllowTrading)
	w.WriteInt32(int32(p.Room.WallThickness))
	w.WriteInt32(int32(p.Room.FloorThickness))
	w.WriteInt32(int32(p.Room.WallHeight))
	return w.Bytes(), nil
}

// SaveRoomSettingsPacket decodes client room settings save (c2s 1090).
type SaveRoomSettingsPacket struct {
	// RoomID stores the target room identifier.
	RoomID int32
	// Name stores the new room display name.
	Name string
	// Description stores the new room description.
	Description string
	// State stores the new access state integer (0=open, 1=locked, 2=password).
	State int32
	// Password stores the new plain-text password (empty when not changing).
	Password string
	// MaxUsers stores the new capacity limit.
	MaxUsers int32
	// AllowPets stores the new pet permission flag.
	AllowPets bool
	// AllowTrading stores the new trading permission flag.
	AllowTrading bool
	// TradeMode stores the new trade policy code.
	TradeMode int32
	// WallThickness stores new wall thickness value.
	WallThickness int32
	// FloorThickness stores new floor thickness value.
	FloorThickness int32
	// WallHeight stores new wall height value.
	WallHeight int32
}

// PacketID returns the protocol packet identifier.
func (p SaveRoomSettingsPacket) PacketID() uint16 { return SaveRoomSettingsPacketID }

// Decode parses packet body payload.
func (p *SaveRoomSettingsPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	id, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.RoomID = id
	if p.Name, err = r.ReadString(); err != nil {
		return err
	}
	if p.Description, err = r.ReadString(); err != nil {
		return err
	}
	if p.State, err = r.ReadInt32(); err != nil {
		return err
	}
	if p.Password, err = r.ReadString(); err != nil {
		p.Password = ""
	}
	if p.MaxUsers, err = r.ReadInt32(); err != nil {
		p.MaxUsers = 25
	}
	p.AllowPets, _ = r.ReadBool()
	p.AllowTrading, _ = r.ReadBool()
	p.TradeMode, _ = r.ReadInt32()
	p.WallThickness, _ = r.ReadInt32()
	p.FloorThickness, _ = r.ReadInt32()
	p.WallHeight, _ = r.ReadInt32()
	return nil
}

// accessStateToInt converts AccessState to protocol integer representation.
func accessStateToInt(s domain.AccessState) int32 {
	switch s {
	case domain.AccessLocked:
		return 1
	case domain.AccessPassword:
		return 2
	case domain.AccessInvisible:
		return 3
	default:
		return 0
	}
}

// IntToAccessState converts protocol integer to AccessState.
func IntToAccessState(n int32) domain.AccessState {
	switch n {
	case 1:
		return domain.AccessLocked
	case 2:
		return domain.AccessPassword
	case 3:
		return domain.AccessInvisible
	default:
		return domain.AccessOpen
	}
}

// GetBannedUsersPacket decodes client ban list request (c2s 2652).
type GetBannedUsersPacket struct {
	// RoomID stores the target room identifier.
	RoomID int32
}

// PacketID returns the protocol packet identifier.
func (p GetBannedUsersPacket) PacketID() uint16 { return GetBannedUsersPacketID }

// Decode parses packet body payload.
func (p *GetBannedUsersPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	id, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.RoomID = id
	return nil
}

// BannedUserEntry defines one entry in the banned users list.
type BannedUserEntry struct {
	// UserID stores the banned user identifier.
	UserID int32
	// Username stores the banned user display name.
	Username string
}

// BannedUsersComposer sends the ban list for a room (s2c 1869).
type BannedUsersComposer struct {
	// RoomID stores the room identifier.
	RoomID int32
	// Bans stores the list of banned users.
	Bans []BannedUserEntry
}

// PacketID returns the protocol packet identifier.
func (p BannedUsersComposer) PacketID() uint16 { return BannedUsersComposerID }

// Encode serializes the banned users list response.
func (p BannedUsersComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.RoomID)
	w.WriteInt32(int32(len(p.Bans)))
	for _, b := range p.Bans {
		w.WriteInt32(b.UserID)
		if err := w.WriteString(b.Username); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// UnbanUserPacket decodes client unban request (c2s 3842).
type UnbanUserPacket struct {
	// UserID stores the user to unban.
	UserID int32
	// RoomID stores the target room identifier.
	RoomID int32
}

// PacketID returns the protocol packet identifier.
func (p UnbanUserPacket) PacketID() uint16 { return UnbanUserPacketID }

// Decode parses packet body payload.
func (p *UnbanUserPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	uid, err := r.ReadInt32()
	if err != nil {
		return err
	}
	rid, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.UserID, p.RoomID = uid, rid
	return nil
}
