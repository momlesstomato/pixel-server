package authflow

import (
	"fmt"
	"strings"
)

const maxTicketLength = 128

// normalizeTicket validates and normalizes SSO ticket input.
func normalizeTicket(ticket string) (string, error) {
	trimmed := strings.TrimSpace(ticket)
	if trimmed == "" {
		return "", fmt.Errorf("ticket is required")
	}
	if len(trimmed) > maxTicketLength {
		return "", fmt.Errorf("ticket exceeds max length of %d", maxTicketLength)
	}
	for _, value := range trimmed {
		if value < 33 || value > 126 {
			return "", fmt.Errorf("ticket contains non-printable characters")
		}
	}
	return trimmed, nil
}
