package social

import "github.com/momlesstomato/pixel-server/core/codec"

// SearchResultEntry defines one user search result wire payload.
type SearchResultEntry struct {
	// ID stores the user identifier.
	ID int32
	// Username stores the display username.
	Username string
	// Motto stores the player motto.
	Motto string
	// Online stores whether the user is currently online.
	Online bool
	// Figure stores the avatar figure string.
	Figure string
}

// MessengerSearchResultComposer defines server messenger.search_result payload.
type MessengerSearchResultComposer struct {
	// Friends stores matching results that are already friends.
	Friends []SearchResultEntry
	// Others stores matching results that are not friends.
	Others []SearchResultEntry
}

// PacketID returns protocol packet identifier.
func (p MessengerSearchResultComposer) PacketID() uint16 {
	return MessengerSearchResultComposerID
}

// Encode serializes packet body payload.
func (p MessengerSearchResultComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := encodeSearchEntries(w, p.Friends); err != nil {
		return nil, err
	}
	if err := encodeSearchEntries(w, p.Others); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// encodeSearchEntries writes a list of search result entries to a codec writer.
func encodeSearchEntries(w *codec.Writer, entries []SearchResultEntry) error {
	w.WriteInt32(int32(len(entries)))
	for _, e := range entries {
		w.WriteInt32(e.ID)
		if err := w.WriteString(e.Username); err != nil {
			return err
		}
		if err := w.WriteString(e.Motto); err != nil {
			return err
		}
		w.WriteBool(e.Online)
		w.WriteBool(false)
		if err := w.WriteString(""); err != nil {
			return err
		}
		w.WriteInt32(0)
		if err := w.WriteString(e.Figure); err != nil {
			return err
		}
		if err := w.WriteString(""); err != nil {
			return err
		}
	}
	return nil
}
