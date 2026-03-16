package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

// addFriendRequest defines the POST /friends request body.
type addFriendRequest struct {
	// FriendID stores the friend user identifier to add.
	FriendID int `json:"friend_id"`
}

// relationshipPatchRequest defines the PATCH /relationship request body.
type relationshipPatchRequest struct {
	// Type stores the new relationship type value.
	Type int `json:"type"`
}

// parsePositiveID parses a positive integer from a path parameter string.
func parsePositiveID(value string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("id must be a positive integer")
	}
	return id, nil
}

// mapMessengerError maps domain errors to HTTP fiber errors.
func mapMessengerError(err error) error {
	if err == nil {
		return nil
	}
	switch err {
	case domain.ErrNotFriends, domain.ErrRequestNotFound:
		return fiber.NewError(http.StatusNotFound, err.Error())
	case domain.ErrAlreadyFriends, domain.ErrDuplicateRequest, domain.ErrSelfRequest:
		return fiber.NewError(http.StatusConflict, err.Error())
	case domain.ErrInvalidRelationship, domain.ErrTargetNotFound:
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	return fiber.NewError(http.StatusInternalServerError, err.Error())
}

// mapFriendships converts domain friendships to JSON response maps.
func mapFriendships(friends []domain.Friendship) []fiber.Map {
	result := make([]fiber.Map, len(friends))
	for i, f := range friends {
		result[i] = fiber.Map{
			"user_one_id":  f.UserOneID,
			"user_two_id":  f.UserTwoID,
			"relationship": int(f.Relationship),
			"created_at":   f.CreatedAt,
		}
	}
	return result
}

// mapRequests converts domain friend requests to JSON response maps.
func mapRequests(requests []domain.FriendRequest) []fiber.Map {
	result := make([]fiber.Map, len(requests))
	for i, r := range requests {
		result[i] = fiber.Map{
			"id":           r.ID,
			"from_user_id": r.FromUserID,
			"to_user_id":   r.ToUserID,
			"created_at":   r.CreatedAt,
		}
	}
	return result
}
