package messaging

import coretransport "pixelsv/pkg/core/transport"

const (
	// RealmHandshakeSecurity is the auth handshake packet realm identifier.
	RealmHandshakeSecurity = "handshake-security"
)

// PacketIngressTopic builds auth handshake ingress topic for one session.
func PacketIngressTopic(sessionID string) string {
	return coretransport.PacketC2STopic(RealmHandshakeSecurity, sessionID)
}

// PacketIngressWildcardTopic builds auth handshake ingress wildcard topic.
func PacketIngressWildcardTopic() string {
	return PacketIngressTopic("*")
}

// ParsePacketIngressTopic extracts session id for auth handshake packet topics.
func ParsePacketIngressTopic(topic string) (string, bool) {
	realm, sessionID, ok := coretransport.ParsePacketC2STopic(topic)
	if !ok || realm != RealmHandshakeSecurity {
		return "", false
	}
	return sessionID, true
}
