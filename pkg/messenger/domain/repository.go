package domain

import "context"

// Repository defines persistent messenger storage behavior.
type Repository interface {
	// ListFriendships returns all friendship rows for one user.
	ListFriendships(ctx context.Context, userID int) ([]Friendship, error)
	// AreFriends reports whether two users share a friendship row.
	AreFriends(ctx context.Context, userOneID, userTwoID int) (bool, error)
	// CountFriends returns the number of friends for one user.
	CountFriends(ctx context.Context, userID int) (int, error)
	// AddFriendship persists one canonical row for one friendship pair.
	AddFriendship(ctx context.Context, userOneID, userTwoID int) error
	// RemoveFriendship deletes one canonical row for one friendship pair.
	RemoveFriendship(ctx context.Context, userOneID, userTwoID int) error
	// SetRelationship updates the relationship type for one directional row.
	SetRelationship(ctx context.Context, userID, friendID int, rel RelationshipType) error
	// GetRelationship returns the relationship type for one directional row.
	GetRelationship(ctx context.Context, userID, friendID int) (RelationshipType, error)
	// GetRelationshipCounts returns grouped relationship counts for one user profile.
	GetRelationshipCounts(ctx context.Context, userID int) ([]RelationshipCount, error)
	// CreateRequest persists one friend request row.
	CreateRequest(ctx context.Context, fromUserID, toUserID int) (FriendRequest, error)
	// FindRequest returns one friend request by identifier.
	FindRequest(ctx context.Context, requestID int) (FriendRequest, error)
	// FindRequestByUsers returns a request row between two users if it exists.
	FindRequestByUsers(ctx context.Context, fromUserID, toUserID int) (FriendRequest, bool, error)
	// ListRequests returns all pending requests addressed to one user.
	ListRequests(ctx context.Context, toUserID int) ([]FriendRequest, error)
	// DeleteRequest removes one request row by identifier.
	DeleteRequest(ctx context.Context, requestID int) error
	// DeleteAllRequests removes all pending requests addressed to one user.
	DeleteAllRequests(ctx context.Context, toUserID int) error
	// SaveOfflineMessage persists one offline message row.
	SaveOfflineMessage(ctx context.Context, fromUserID, toUserID int, message string) error
	// GetAndDeleteOfflineMessages returns and atomically removes offline messages for one user.
	GetAndDeleteOfflineMessages(ctx context.Context, userID int) ([]OfflineMessage, error)
	// DeleteOfflineMessagesOlderThan removes messages whose sent_at precedes the cutoff epoch.
	DeleteOfflineMessagesOlderThan(ctx context.Context, cutoffUnix int64) error
	// SearchUsers returns users whose username contains the query string.
	SearchUsers(ctx context.Context, query string, limit int) ([]SearchResult, error)
	// FindUserIDByUsername returns the identifier for one username if it exists.
	FindUserIDByUsername(ctx context.Context, username string) (int, bool, error)
	// FindUsersByIDs returns profile records for a set of user identifiers.
	FindUsersByIDs(ctx context.Context, ids []int) ([]SearchResult, error)
}
