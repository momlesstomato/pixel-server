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
	if r.Remaining() > 0 {
		_, _ = r.ReadString()
	}
	if r.Remaining() > 0 {
		_, _ = r.ReadString()
	}
	if r.Remaining() >= 4 {
		_, _ = r.ReadInt32()
	}
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
	// CfhTopic stores the optional moderation topic or note.
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
	duration, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.Duration = duration
	tailOffset := len(body) - r.Remaining()
	topic, banType, ok := decodeLegacyBanTail(body[tailOffset:])
	if ok {
		p.CfhTopic = topic
		p.BanType = banType
		return nil
	}
	banType, ok = decodeNitroBanTail(body[tailOffset:])
	if ok {
		p.BanType = banType
		return nil
	}
	p.BanType = 0
	return nil
}

func decodeLegacyBanTail(body []byte) (string, int32, bool) {
	r := codec.NewReader(body)
	topic, err := r.ReadString()
	if err != nil {
		return "", 0, false
	}
	ipBan := false
	machineBan := false
	if r.Remaining() > 0 {
		value, boolErr := r.ReadBool()
		if boolErr != nil {
			return "", 0, false
		}
		ipBan = value
	}
	if r.Remaining() > 0 {
		value, boolErr := r.ReadBool()
		if boolErr != nil {
			return "", 0, false
		}
		machineBan = value
	}
	if r.Remaining() == 4 {
		if _, err := r.ReadInt32(); err != nil {
			return "", 0, false
		}
	}
	if r.Remaining() != 0 {
		return "", 0, false
	}
	switch {
	case machineBan:
		return topic, 2, true
	case ipBan:
		return topic, 1, true
	default:
		return topic, 0, true
	}
}

func decodeNitroBanTail(body []byte) (int32, bool) {
	r := codec.NewReader(body)
	banType, err := r.ReadInt32()
	if err != nil {
		return 0, false
	}
	if r.Remaining() > 0 {
		if _, err := r.ReadBool(); err != nil {
			return 0, false
		}
	}
	if r.Remaining() == 4 {
		if _, err := r.ReadInt32(); err != nil {
			return 0, false
		}
	}
	if r.Remaining() != 0 {
		return 0, false
	}
	return banType, true
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

// ModAlertUserPacket decodes a moderator direct alert/message request.
type ModAlertUserPacket struct {
	// UserID stores the target user identifier.
	UserID int32
	// Message stores the moderator message sent to the user.
	Message string
}

// Decode reads fields from the packet body.
func (p *ModAlertUserPacket) Decode(body []byte) error {
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
	if r.Remaining() >= 4 {
		_, _ = r.ReadInt32()
	}
	return nil
}

// ModRoomAlertPacket decodes a moderator current-room alert request.
type ModRoomAlertPacket struct {
	// ActionType stores the room moderation action code.
	ActionType int32
	// Message stores the alert content.
	Message string
	// Detail stores the optional moderator detail string.
	Detail string
}

// Decode reads fields from the packet body.
func (p *ModRoomAlertPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	actionType, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.ActionType = actionType
	msg, err := r.ReadString()
	if err != nil {
		return err
	}
	p.Message = msg
	if r.Remaining() > 0 {
		detail, detailErr := r.ReadString()
		if detailErr == nil {
			p.Detail = detail
		}
	}
	return nil
}
