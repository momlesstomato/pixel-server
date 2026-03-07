package messaging

import coretransport "pixelsv/pkg/core/transport"

const (
	// RealmSessionConnection is the packet realm identifier.
	RealmSessionConnection = "session-connection"
)

// PacketIngressTopic builds session-connection ingress topic for one session.
func PacketIngressTopic(sessionID string) string {
	return coretransport.PacketC2STopic(RealmSessionConnection, sessionID)
}

// PacketIngressWildcardTopic builds wildcard ingress topic for session-connection packets.
func PacketIngressWildcardTopic() string {
	return PacketIngressTopic("*")
}

// ParsePacketIngressTopic extracts session id for one session-connection ingress topic.
func ParsePacketIngressTopic(topic string) (string, bool) {
	realm, sessionID, ok := coretransport.ParsePacketC2STopic(topic)
	if !ok || realm != RealmSessionConnection {
		return "", false
	}
	return sessionID, true
}
