package msginit

import "github.com/momlesstomato/pixel-server/core/codec"

// RequestEntry defines one friend request wire format.
type RequestEntry struct {
	// ID stores the request identifier.
	ID int32
	// FromUsername stores the requesting user display name.
	FromUsername string
	// FromFigure stores the requesting user figure string.
	FromFigure string
}

// MessengerRequestsComposer defines server messenger.requests payload.
type MessengerRequestsComposer struct {
	// Requests stores all pending request entries.
	Requests []RequestEntry
}

// PacketID returns protocol packet identifier.
func (p MessengerRequestsComposer) PacketID() uint16 { return MessengerRequestsComposerID }

// Encode serializes packet body payload.
func (p MessengerRequestsComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	total := int32(len(p.Requests))
		w.WriteInt32(total)
		w.WriteInt32(total)
	for _, req := range p.Requests {
		w.WriteInt32(req.ID)
		if err := w.WriteString(req.FromUsername); err != nil {
			return nil, err
		}
		if err := w.WriteString(req.FromFigure); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}
