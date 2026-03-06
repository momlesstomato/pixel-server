package transport

import "strings"

const (
	// TopicPacketC2S is the gateway inbound packet topic prefix.
	TopicPacketC2S = "packet.c2s"
)

// PacketC2STopic builds a packet ingress topic for one realm and session.
func PacketC2STopic(realm string, sessionID string) string {
	return joinTopic(TopicPacketC2S, realm, sessionID)
}

// ParsePacketC2STopic extracts realm and session id from packet c2s topics.
func ParsePacketC2STopic(topic string) (string, string, bool) {
	tokens := strings.Split(topic, ".")
	if len(tokens) != 4 {
		return "", "", false
	}
	if tokens[0] != "packet" || tokens[1] != "c2s" {
		return "", "", false
	}
	if tokens[2] == "" || tokens[3] == "" {
		return "", "", false
	}
	return tokens[2], tokens[3], true
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
