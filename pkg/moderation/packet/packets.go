package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// ModKickUserPacket decodes a moderator kick request.
type ModKickUserPacket struct {
	// UserID stores the target user identifier.
	UserID int32
	// Message stores the kick reason displayed to the user.
	Message string
}

// Decode reads fields from the packet body.
func (p *ModKickUserPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	uid, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.UserID = uid
	msg, err := r.ReadString()
	if err != nil {
		return err
	}
	p.Message = msg
	return nil
}

// ModMuteUserPacket decodes a moderator mute request.
type ModMuteUserPacket struct {
	// UserID stores the target user identifier.
	UserID int32
	// Message stores the mute reason.
	Message string
	// Minutes stores the mute duration in minutes.
	Minutes int32
}

// Decode reads fields from the packet body.
func (p *ModMuteUserPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	uid, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.UserID = uid
	msg, err := r.ReadString()
	if err != nil {
		return err
	}
	p.Message = msg
	mins, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.Minutes = mins
	return nil
}

// ModBanUserPacket decodes a moderator ban request.
type ModBanUserPacket struct {
	// UserID stores the target user identifier.
	UserID int32
	// Message stores the ban reason.
	Message string
	// BanType stores the ban kind (0=account, 1=ip, 2=machine).
	BanType int32
	// CfhTopic stores the optional call-for-help topic.
	CfhTopic string
	// Duration stores the ban duration in hours.
	Duration int32
}

// Decode reads fields from the packet body.
func (p *ModBanUserPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	uid, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.UserID = uid
	msg, err := r.ReadString()
	if err != nil {
		return err
	}
	p.Message = msg
	bt, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.BanType = bt
	topic, err := r.ReadString()
	if err != nil {
		return err
	}
	p.CfhTopic = topic
	dur, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.Duration = dur
	return nil
}

// ModWarnUserPacket decodes a moderator warn/caution request.
type ModWarnUserPacket struct {
	// UserID stores the target user identifier.
	UserID int32
	// Message stores the warning message sent to the user.
	Message string
}

// Decode reads fields from the packet body.
func (p *ModWarnUserPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	uid, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.UserID = uid
	msg, err := r.ReadString()
	if err != nil {
		return err
	}
	p.Message = msg
	return nil
}
