package domain

import "time"

// RelationshipType defines a friendship relationship label value.
type RelationshipType int

const (
	// RelationshipNone defines no relationship label.
	RelationshipNone RelationshipType = 0
	// RelationshipHeart defines heart relationship label.
	RelationshipHeart RelationshipType = 1
	// RelationshipSmile defines smile relationship label.
	RelationshipSmile RelationshipType = 2
	// RelationshipBobba defines bobba relationship label.
	RelationshipBobba RelationshipType = 3
)

// KnownRelationships maps all registered relationship types to their labels.
// Plugins extend this registry at startup via RegisterRelationship.
var KnownRelationships = map[RelationshipType]string{
	RelationshipNone:  "none",
	RelationshipHeart: "heart",
	RelationshipSmile: "smile",
	RelationshipBobba: "bobba",
}

// RegisterRelationship adds a new relationship type to the valid registry.
func RegisterRelationship(t RelationshipType, label string) {
	KnownRelationships[t] = label
}

// IsValidRelationship reports whether a relationship type is registered.
func IsValidRelationship(t RelationshipType) bool {
	_, ok := KnownRelationships[t]
	return ok
}

// Friendship defines one bidirectional friendship row payload.
type Friendship struct {
	// UserOneID stores the perspective owner user identifier.
	UserOneID int
	// UserTwoID stores the friend user identifier.
	UserTwoID int
	// Relationship stores the relationship type from UserOneID's perspective.
	Relationship RelationshipType
	// CreatedAt stores the friendship creation timestamp.
	CreatedAt time.Time
}

// FriendRequest defines one pending friend request payload.
type FriendRequest struct {
	// ID stores the request row identifier.
	ID int
	// FromUserID stores the requesting user identifier.
	FromUserID int
	// ToUserID stores the target user identifier.
	ToUserID int
	// CreatedAt stores the request creation timestamp.
	CreatedAt time.Time
}

// OfflineMessage defines one offline message delivery payload.
type OfflineMessage struct {
	// ID stores the message row identifier.
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

// SearchResult defines one user search result payload.
type SearchResult struct {
	// ID stores the user identifier.
	ID int
	// Username stores the display username.
	Username string
	// Figure stores the avatar appearance string.
	Figure string
	// Gender stores the gender value.
	Gender string
	// Motto stores the player motto.
	Motto string
	// Online stores whether the user is currently connected.
	Online bool
	// IsFriend stores whether the result is already a friend of the searcher.
	IsFriend bool
}

// RelationshipCount defines one grouped relationship count row.
type RelationshipCount struct {
	// Type stores the relationship type.
	Type RelationshipType
	// Count stores the number of friends with this relationship.
	Count int
	// SampleUserIDs stores up to three sample friend identifiers.
	SampleUserIDs []int
}
