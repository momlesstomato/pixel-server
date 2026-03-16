package msginit

import (
	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

// FriendEntry defines one friend record wire format.
type FriendEntry struct {
	// ID stores the friend user identifier.
	ID int32
	// Username stores the display name.
	Username string
	// Gender stores 0 for male, 1 for female.
	Gender int32
	// Online stores whether the friend is currently connected.
	Online bool
	// Figure stores the avatar appearance string.
	Figure string
	// Motto stores the player motto.
	Motto string
	// Relationship stores the relationship type from viewer perspective.
	Relationship int16
}

// MessengerFriendsComposer defines server messenger.friends fragment payload.
type MessengerFriendsComposer struct {
	// TotalFragments stores the total number of fragments.
	TotalFragments int32
	// FragmentNumber stores this fragment index (0-based).
	FragmentNumber int32
	// Friends stores friend records in this fragment.
	Friends []FriendEntry
}

// PacketID returns protocol packet identifier.
func (p MessengerFriendsComposer) PacketID() uint16 { return MessengerFriendsComposerID }

// Encode serializes packet body payload.
func (p MessengerFriendsComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.TotalFragments)
	w.WriteInt32(p.FragmentNumber)
	w.WriteInt32(int32(len(p.Friends)))
	for _, f := range p.Friends {
		if err := encodeFriendEntry(w, f); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// encodeFriendEntry writes one friend record to a codec writer.
func encodeFriendEntry(w *codec.Writer, f FriendEntry) error {
	w.WriteInt32(f.ID)
	if err := w.WriteString(f.Username); err != nil {
		return err
	}
	w.WriteInt32(f.Gender)
	w.WriteBool(f.Online)
	w.WriteBool(false)
	figure := ""
	if f.Online {
		figure = f.Figure
	}
	if err := w.WriteString(figure); err != nil {
		return err
	}
	w.WriteInt32(0)
	if err := w.WriteString(f.Motto); err != nil {
		return err
	}
	if err := w.WriteString(""); err != nil {
		return err
	}
	if err := w.WriteString(""); err != nil {
		return err
	}
	w.WriteBool(true)
	w.WriteBool(false)
	w.WriteBool(false)
	w.WriteUint16(uint16(f.Relationship))
	return nil
}

// MapRelationship converts domain relationship type to wire int16.
func MapRelationship(rel domain.RelationshipType) int16 { return int16(rel) }
