package packet

import "github.com/momlesstomato/pixel-server/core/codec"

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
