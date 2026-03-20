package domain

import "time"

// MessageLog defines one persisted message record for auditing and security.
type MessageLog struct {
	// ID stores the auto-incremented record identifier.
	ID int
	// FromUserID stores the sender user identifier.
	FromUserID int
	// ToUserID stores the recipient user identifier.
	ToUserID int
	// Message stores the message content.
	Message string
	// SentAt stores the original send timestamp.
	SentAt time.Time
}
