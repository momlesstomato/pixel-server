package notification

import "fmt"

// AllChannel returns the shared broadcast channel for all active connections.
func AllChannel() string {
	return "broadcast:all"
}

// UserChannel returns the targeted broadcast channel for one user identifier.
func UserChannel(userID int) string {
	return fmt.Sprintf("broadcast:user:%d", userID)
}
