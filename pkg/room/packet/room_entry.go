package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// OpenFlatConnectionPacket decodes room entry request (c2s 2312).
type OpenFlatConnectionPacket struct {
	// RoomID stores the requested room identifier.
	RoomID int32
	// Password stores the optional room password.
	Password string
}

// PacketID returns the protocol packet identifier.
func (p OpenFlatConnectionPacket) PacketID() uint16 { return OpenFlatConnectionPacketID }

// Decode parses packet body payload.
func (p *OpenFlatConnectionPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	roomID, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.RoomID = roomID
	pwd, err := r.ReadString()
	if err != nil {
		p.Password = ""
		return nil
	}
	p.Password = pwd
	return nil
}

// Encode serializes packet body payload.
func (p OpenFlatConnectionPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.RoomID)
	if err := w.WriteString(p.Password); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// OpenConnectionComposer acknowledges room connection (s2c 758).
type OpenConnectionComposer struct{}

// PacketID returns the protocol packet identifier.
func (p OpenConnectionComposer) PacketID() uint16 { return OpenConnectionComposerID }

// Encode serializes packet body.
func (p OpenConnectionComposer) Encode() ([]byte, error) { return []byte{}, nil }

// RoomReadyComposer sends room model and ID (s2c 768).
type RoomReadyComposer struct {
	// ModelSlug stores the room model identifier string.
	ModelSlug string
	// RoomID stores the room identifier.
	RoomID int32
}

// PacketID returns the protocol packet identifier.
func (p RoomReadyComposer) PacketID() uint16 { return RoomReadyComposerID }

// Encode serializes the room ready response.
func (p RoomReadyComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := w.WriteString(p.ModelSlug); err != nil {
		return nil, err
	}
	w.WriteInt32(p.RoomID)
	return w.Bytes(), nil
}

// DoorbellComposer notifies room owner of a visitor at the door (s2c 2068).
type DoorbellComposer struct {
	// Username stores the visitor display name.
	Username string
}

// PacketID returns the protocol packet identifier.
func (p DoorbellComposer) PacketID() uint16 { return DoorbellComposerID }

// Encode serializes the doorbell notification.
func (p DoorbellComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := w.WriteString(p.Username); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// FlatAccessibleComposer informs a visitor of entry approval result (s2c 735).
type FlatAccessibleComposer struct {
	// Username stores the approved visitor display name.
	Username string
	// Accessible reports whether entry was approved.
	Accessible bool
}

// PacketID returns the protocol packet identifier.
func (p FlatAccessibleComposer) PacketID() uint16 { return FlatAccessibleComposerID }

// Encode serializes the flat access result.
func (p FlatAccessibleComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := w.WriteString(p.Username); err != nil {
		return nil, err
	}
	w.WriteBool(p.Accessible)
	return w.Bytes(), nil
}

// FloodControlComposer informs a client that room entry is temporarily rate limited (s2c 566).
type FloodControlComposer struct {
	// Seconds stores the remaining cooldown duration in seconds.
	Seconds int32
}

// PacketID returns the protocol packet identifier.
func (p FloodControlComposer) PacketID() uint16 { return FloodControlComposerID }

// Encode serializes the flood-control payload.
func (p FloodControlComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.Seconds)
	return w.Bytes(), nil
}

// LetUserInPacket decodes room owner doorbell approval (c2s 1781).
type LetUserInPacket struct {
	// Username stores the visitor display name to approve or deny.
	Username string
	// Let reports whether the visitor is approved.
	Let bool
}

// PacketID returns the protocol packet identifier.
func (p LetUserInPacket) PacketID() uint16 { return LetUserInPacketID }

// Decode parses packet body payload.
func (p *LetUserInPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	name, err := r.ReadString()
	if err != nil {
		return err
	}
	p.Username = name
	let, err := r.ReadBool()
	if err != nil {
		return err
	}
	p.Let = let
	return nil
}

// DesktopViewComposer navigates the client back to hotel view (s2c 122).
type DesktopViewComposer struct{}

// PacketID returns the protocol packet identifier.
func (p DesktopViewComposer) PacketID() uint16 { return DesktopViewComposerID }

// Encode serializes the hotel view redirect payload.
func (p DesktopViewComposer) Encode() ([]byte, error) { return []byte{}, nil }

// RoomForwardComposer redirects the client to another room (s2c 160).
type RoomForwardComposer struct {
	// RoomID stores the target room identifier.
	RoomID int32
}

// PacketID returns the protocol packet identifier.
func (p RoomForwardComposer) PacketID() uint16 { return RoomForwardComposerID }

// Encode serializes the room forward payload.
func (p RoomForwardComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.RoomID)
	return w.Bytes(), nil
}

// DeleteRoomPacket decodes client room deletion request (c2s 532).
type DeleteRoomPacket struct {
	// RoomID stores the room to delete.
	RoomID int32
}

// PacketID returns the protocol packet identifier.
func (p DeleteRoomPacket) PacketID() uint16 { return DeleteRoomPacketID }

// Decode parses packet body payload.
func (p *DeleteRoomPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	id, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.RoomID = id
	return nil
}

// GiveRoomScorePacket decodes client room vote request (c2s 3616).
type GiveRoomScorePacket struct {
	// Score stores the vote value (typically 1 or -1).
	Score int32
}

// PacketID returns the protocol packet identifier.
func (p GiveRoomScorePacket) PacketID() uint16 { return GiveRoomScorePacketID }

// Decode parses packet body payload.
func (p *GiveRoomScorePacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	score, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.Score = score
	return nil
}

// RoomScoreComposer sends room score and voting eligibility (s2c 3271).
type RoomScoreComposer struct {
	// Score stores the current room score.
	Score int32
	// CanVote reports whether the recipient may vote.
	CanVote bool
}

// PacketID returns the protocol packet identifier.
func (p RoomScoreComposer) PacketID() uint16 { return RoomScoreComposerID }

// Encode serializes room score response.
func (p RoomScoreComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.Score)
	w.WriteBool(p.CanVote)
	return w.Bytes(), nil
}

// YouAreControllerComposer informs a user of their room rights level (s2c 780).
type YouAreControllerComposer struct {
	// Level stores the rights level (0 = none, 1 = rights holder).
	Level int32
}

// PacketID returns the protocol packet identifier.
func (p YouAreControllerComposer) PacketID() uint16 { return YouAreControllerComposerID }

// Encode serializes the rights level response.
func (p YouAreControllerComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.Level)
	return w.Bytes(), nil
}

// YouAreNotControllerComposer clears the recipient room rights state (s2c 2392).
type YouAreNotControllerComposer struct{}

// PacketID returns the protocol packet identifier.
func (p YouAreNotControllerComposer) PacketID() uint16 { return YouAreNotControllerComposerID }

// Encode serializes the empty rights-clear response.
func (p YouAreNotControllerComposer) Encode() ([]byte, error) { return []byte{}, nil }

// YouAreOwnerComposer marks the recipient as the room owner (s2c 339).
type YouAreOwnerComposer struct{}

// PacketID returns the protocol packet identifier.
func (p YouAreOwnerComposer) PacketID() uint16 { return YouAreOwnerComposerID }

// Encode serializes the empty owner response.
func (p YouAreOwnerComposer) Encode() ([]byte, error) { return []byte{}, nil }

// RightsEntry defines one rights holder entry for the rights list composer.
type RightsEntry struct {
	// UserID stores the rights holder identifier.
	UserID int32
	// Username stores the rights holder display name.
	Username string
}

// RoomRightsListComposer sends all room rights holders to the owner (s2c 1284).
type RoomRightsListComposer struct {
	// RoomID stores the room identifier.
	RoomID int32
	// Entries stores the rights holder list.
	Entries []RightsEntry
}

// PacketID returns the protocol packet identifier.
func (p RoomRightsListComposer) PacketID() uint16 { return RoomRightsListComposerID }

// Encode serializes the rights list response.
func (p RoomRightsListComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.RoomID)
	w.WriteInt32(int32(len(p.Entries)))
	for _, e := range p.Entries {
		w.WriteInt32(e.UserID)
		if err := w.WriteString(e.Username); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// RoomRightsAddedComposer sends one incremental rights addition (s2c 2088).
type RoomRightsAddedComposer struct {
	// RoomID stores the room identifier.
	RoomID int32
	// Entry stores the rights holder that was added.
	Entry RightsEntry
}

// PacketID returns the protocol packet identifier.
func (p RoomRightsAddedComposer) PacketID() uint16 { return RoomRightsAddedComposerID }

// Encode serializes the incremental rights-add response.
func (p RoomRightsAddedComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.RoomID)
	w.WriteInt32(p.Entry.UserID)
	if err := w.WriteString(p.Entry.Username); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// RoomRightsRemovedComposer sends one incremental rights removal (s2c 1327).
type RoomRightsRemovedComposer struct {
	// RoomID stores the room identifier.
	RoomID int32
	// UserID stores the rights holder that was removed.
	UserID int32
}

// PacketID returns the protocol packet identifier.
func (p RoomRightsRemovedComposer) PacketID() uint16 { return RoomRightsRemovedComposerID }

// Encode serializes the incremental rights-remove response.
func (p RoomRightsRemovedComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.RoomID)
	w.WriteInt32(p.UserID)
	return w.Bytes(), nil
}
