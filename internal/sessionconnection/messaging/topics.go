package messaging

import "strings"

const (
	// TopicAuthenticated broadcasts successful session authentication events.
	TopicAuthenticated = "session.authenticated"
	// TopicConnected broadcasts newly connected session events.
	TopicConnected = "session.connected"
	// TopicDisconnected broadcasts session disconnect events.
	TopicDisconnected = "session.disconnected"
	// TopicOutputPrefix defines session output topic prefix.
	TopicOutputPrefix = "session.output"
	// TopicDisconnectPrefix defines session disconnect control topic prefix.
	TopicDisconnectPrefix = "session.disconnect"
)

// OutputTopic builds a session output topic for one session id.
func OutputTopic(sessionID string) string {
	return join(TopicOutputPrefix, sessionID)
}

// OutputWildcardTopic builds the wildcard session output subscription topic.
func OutputWildcardTopic() string {
	return TopicOutputPrefix + ".>"
}

// DisconnectTopic builds a session disconnect control topic for one session id.
func DisconnectTopic(sessionID string) string {
	return join(TopicDisconnectPrefix, sessionID)
}

// DisconnectWildcardTopic builds a wildcard disconnect control topic.
func DisconnectWildcardTopic() string {
	return TopicDisconnectPrefix + ".>"
}

// ParseOutputTopic extracts session id from one session output topic.
func ParseOutputTopic(topic string) (string, bool) {
	tokens := strings.Split(topic, ".")
	if len(tokens) != 3 {
		return "", false
	}
	if tokens[0] != "session" || tokens[1] != "output" || tokens[2] == "" {
		return "", false
	}
	return tokens[2], true
}

// ParseDisconnectTopic extracts session id from one disconnect control topic.
func ParseDisconnectTopic(topic string) (string, bool) {
	tokens := strings.Split(topic, ".")
	if len(tokens) != 3 {
		return "", false
	}
	if tokens[0] != "session" || tokens[1] != "disconnect" || tokens[2] == "" {
		return "", false
	}
	return tokens[2], true
}

func join(a string, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	return a + "." + b
}
