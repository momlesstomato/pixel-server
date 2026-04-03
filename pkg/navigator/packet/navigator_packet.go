package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// NavigatorMetaDataPacket defines navigator.metadata (s2c 3052) payload.
type NavigatorMetaDataPacket struct {
	// TopLevelContexts stores navigator tab entries.
	TopLevelContexts []string
}

// PacketID returns the wire protocol packet identifier.
func (p NavigatorMetaDataPacket) PacketID() uint16 { return NavigatorMetaDataPacketID }

// Encode serializes navigator metadata into packet body.
func (p NavigatorMetaDataPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(len(p.TopLevelContexts)))
	for _, ctx := range p.TopLevelContexts {
		if err := w.WriteString(ctx); err != nil {
			return nil, err
		}
		w.WriteInt32(0)
	}
	return w.Bytes(), nil
}

// NavigatorCollapsedPacket defines navigator.collapsed (s2c 1543) payload.
type NavigatorCollapsedPacket struct {
	// Categories stores collapsed category identifiers.
	Categories []string
}

// PacketID returns the wire protocol packet identifier.
func (p NavigatorCollapsedPacket) PacketID() uint16 { return NavigatorCollapsedPacketID }

// Encode serializes collapsed categories into packet body.
func (p NavigatorCollapsedPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(len(p.Categories)))
	for _, cat := range p.Categories {
		if err := w.WriteString(cat); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// NavigatorSettingsPacket defines navigator.settings (s2c 518) payload.
type NavigatorSettingsPacket struct {
	// X stores the window X position.
	X int32
	// Y stores the window Y position.
	Y int32
	// Width stores the window width.
	Width int32
	// Height stores the window height.
	Height int32
	// LeftPanelHidden stores flag for left panel visibility.
	LeftPanelHidden bool
	// ResultMode stores the display mode for results.
	ResultMode int32
}

// PacketID returns the wire protocol packet identifier.
func (p NavigatorSettingsPacket) PacketID() uint16 { return NavigatorSettingsPacketID }

// Encode serializes navigator settings into packet body.
func (p NavigatorSettingsPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.X)
	w.WriteInt32(p.Y)
	w.WriteInt32(p.Width)
	w.WriteInt32(p.Height)
	w.WriteBool(p.LeftPanelHidden)
	w.WriteInt32(p.ResultMode)
	return w.Bytes(), nil
}

// NavigatorSavedSearchesPacket defines navigator.saved_searches (s2c 3984) payload.
type NavigatorSavedSearchesPacket struct {
	// Searches stores per-user saved search entries.
	Searches []SavedSearchEntry
}

// SavedSearchEntry defines one saved search entry for encoding.
type SavedSearchEntry struct {
	// ID stores saved search identifier.
	ID int32
	// SearchCode stores the search tab key.
	SearchCode string
	// Filter stores the user filter string.
	Filter string
}

// PacketID returns the wire protocol packet identifier.
func (p NavigatorSavedSearchesPacket) PacketID() uint16 { return NavigatorSavedSearchesPacketID }

// Encode serializes saved searches into packet body.
func (p NavigatorSavedSearchesPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(len(p.Searches)))
	for _, s := range p.Searches {
		w.WriteInt32(s.ID)
		if err := w.WriteString(s.SearchCode); err != nil {
			return nil, err
		}
		if err := w.WriteString(s.Filter); err != nil {
			return nil, err
		}
		if err := w.WriteString(""); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// NavigatorEventCategoriesPacket defines navigator.event_categories (s2c 3244) payload.
type NavigatorEventCategoriesPacket struct{}

// PacketID returns the wire protocol packet identifier.
func (p NavigatorEventCategoriesPacket) PacketID() uint16 {
	return NavigatorEventCategoriesPacketID
}

// Encode serializes an empty event categories list into packet body.
func (p NavigatorEventCategoriesPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(0)
	return w.Bytes(), nil
}
