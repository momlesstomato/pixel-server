package packet

import (
	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
)

// NavigatorSearchResultsPacket defines navigator.search_results (s2c 2690) payload.
type NavigatorSearchResultsPacket struct {
	// SearchCode stores the search category key.
	SearchCode string
	// Filter stores the applied search filter.
	Filter string
	// Results stores search result blocks.
	Results []SearchResultBlock
}

// SearchResultBlock defines one navigator search result block.
type SearchResultBlock struct {
	// SearchCode stores the block search code.
	SearchCode string
	// Text stores the block display text.
	Text string
	// ActionAllowed stores the block action type.
	ActionAllowed int32
	// ForceClosed stores collapse state flag.
	ForceClosed bool
	// ViewMode stores the display layout mode.
	ViewMode int32
	// Rooms stores rooms in this result block.
	Rooms []domain.Room
}

// PacketID returns the wire protocol packet identifier.
func (p NavigatorSearchResultsPacket) PacketID() uint16 {
	return NavigatorSearchResultsPacketID
}

// Encode serializes navigator search results into packet body.
func (p NavigatorSearchResultsPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := w.WriteString(p.SearchCode); err != nil {
		return nil, err
	}
	if err := w.WriteString(p.Filter); err != nil {
		return nil, err
	}
	w.WriteInt32(int32(len(p.Results)))
	for _, block := range p.Results {
		if err := encodeSearchBlock(w, block); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// encodeSearchBlock writes one search result block to the writer.
func encodeSearchBlock(w *codec.Writer, block SearchResultBlock) error {
	if err := w.WriteString(block.SearchCode); err != nil {
		return err
	}
	if err := w.WriteString(block.Text); err != nil {
		return err
	}
	w.WriteInt32(block.ActionAllowed)
	w.WriteBool(block.ForceClosed)
	w.WriteInt32(block.ViewMode)
	w.WriteInt32(int32(len(block.Rooms)))
	for _, room := range block.Rooms {
		if err := EncodeRoomData(w, room); err != nil {
			return err
		}
	}
	return nil
}

// GuestRoomDataPacket defines navigator.guest_room_data (s2c 687) payload.
type GuestRoomDataPacket struct {
	// Room stores the room data to encode.
	Room domain.Room
	// Forward stores whether to enter the room.
	Forward bool
}

// PacketID returns the wire protocol packet identifier.
func (p GuestRoomDataPacket) PacketID() uint16 { return GuestRoomDataPacketID }

// Encode serializes guest room data into packet body.
func (p GuestRoomDataPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteBool(false)
	if err := EncodeRoomData(w, p.Room); err != nil {
		return nil, err
	}
	w.WriteBool(p.Forward)
	w.WriteBool(false)
	w.WriteBool(false)
	w.WriteBool(false)
	w.WriteInt32(0)
	w.WriteInt32(1)
	w.WriteInt32(1)
	w.WriteBool(false)
	w.WriteInt32(0)
	w.WriteInt32(0)
	w.WriteInt32(1)
	w.WriteInt32(14)
	w.WriteInt32(1)
	return w.Bytes(), nil
}

// EncodeRoomData writes room fields to a codec writer.
func EncodeRoomData(w *codec.Writer, room domain.Room) error {
	w.WriteInt32(int32(room.ID))
	if err := w.WriteString(room.Name); err != nil {
		return err
	}
	w.WriteInt32(int32(room.OwnerID))
	if err := w.WriteString(room.OwnerName); err != nil {
		return err
	}
	w.WriteInt32(stateToInt(room.State))
	w.WriteInt32(int32(room.CurrentUsers))
	w.WriteInt32(int32(room.MaxUsers))
	if err := w.WriteString(room.Description); err != nil {
		return err
	}
	w.WriteInt32(int32(room.TradeMode))
	w.WriteInt32(int32(room.Score))
	w.WriteInt32(0)
	w.WriteInt32(int32(room.CategoryID))
	w.WriteInt32(int32(len(room.Tags)))
	for _, tag := range room.Tags {
		if err := w.WriteString(tag); err != nil {
			return err
		}
	}
	w.WriteInt32(8)
	return nil
}

// stateToInt maps room state string to protocol integer.
func stateToInt(state string) int32 {
	switch state {
	case "locked":
		return 1
	case "password":
		return 2
	case "invisible":
		return 3
	default:
		return 0
	}
}
