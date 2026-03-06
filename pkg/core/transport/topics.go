package transport

import "strings"

const (
	// TopicHandshakeC2S is the gateway to auth handshake input topic prefix.
	TopicHandshakeC2S = "handshake.c2s"
	// TopicSessionAuthenticated broadcasts successful session authentication.
	TopicSessionAuthenticated = "session.authenticated"
	// TopicSessionDisconnected broadcasts session disconnect events.
	TopicSessionDisconnected = "session.disconnected"
	// TopicRoomInput is the gateway to game room input topic prefix.
	TopicRoomInput = "room.input"
	// TopicSessionOutput is the game to gateway session output topic prefix.
	TopicSessionOutput = "session.output"
	// TopicSocialNotification is the social fan-out topic prefix.
	TopicSocialNotification = "social.notification"
	// TopicNavigatorRoomUpdated is the room update notification topic prefix.
	TopicNavigatorRoomUpdated = "navigator.room_updated"
	// TopicCatalogPurchaseCompleted is the purchase completion topic.
	TopicCatalogPurchaseCompleted = "catalog.purchase.completed"
	// TopicModerationBanIssued is the moderation ban notification topic prefix.
	TopicModerationBanIssued = "moderation.ban.issued"
)

// HandshakeC2STopic builds a handshake ingress topic for one session.
func HandshakeC2STopic(sessionID string) string {
	return joinTopic(TopicHandshakeC2S, sessionID)
}

// RoomInputTopic builds a room ingress topic for one room id.
func RoomInputTopic(roomID string) string {
	return joinTopic(TopicRoomInput, roomID)
}

// SessionOutputTopic builds a session egress topic for one session.
func SessionOutputTopic(sessionID string) string {
	return joinTopic(TopicSessionOutput, sessionID)
}

// SocialNotificationTopic builds a social notification topic for one user.
func SocialNotificationTopic(userID string) string {
	return joinTopic(TopicSocialNotification, userID)
}

// NavigatorRoomUpdatedTopic builds a navigator room update topic.
func NavigatorRoomUpdatedTopic(roomID string) string {
	return joinTopic(TopicNavigatorRoomUpdated, roomID)
}

// ModerationBanIssuedTopic builds a ban issued topic for one user.
func ModerationBanIssuedTopic(userID string) string {
	return joinTopic(TopicModerationBanIssued, userID)
}

// joinTopic concatenates non-empty topic parts using dot separators.
func joinTopic(parts ...string) string {
	filtered := parts[:0]
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, ".")
}
