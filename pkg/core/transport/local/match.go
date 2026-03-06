package local

import "strings"

// matchTopic matches a NATS-style wildcard pattern against a concrete topic.
func matchTopic(pattern string, topic string) bool {
	patternTokens := strings.Split(pattern, ".")
	topicTokens := strings.Split(topic, ".")
	for idx, token := range patternTokens {
		if token == ">" {
			return true
		}
		if idx >= len(topicTokens) {
			return false
		}
		if token != "*" && token != topicTokens[idx] {
			return false
		}
	}
	return len(patternTokens) == len(topicTokens)
}
