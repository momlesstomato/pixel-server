package domain

import "errors"

// ErrFriendListFull is returned when the requesting user's friend limit is reached.
var ErrFriendListFull = errors.New("friend list is full")

// ErrTargetFriendListFull is returned when the target user's friend limit is reached.
var ErrTargetFriendListFull = errors.New("target friend list is full")

// ErrTargetNotAccepting is returned when the target is not accepting friend requests.
var ErrTargetNotAccepting = errors.New("target not accepting requests")

// ErrTargetNotFound is returned when the target user does not exist.
var ErrTargetNotFound = errors.New("target user not found")

// ErrAlreadyFriends is returned when the two users are already friends.
var ErrAlreadyFriends = errors.New("already friends")

// ErrNotFriends is returned when the two users are not friends.
var ErrNotFriends = errors.New("not friends")

// ErrRequestNotFound is returned when a friend request row does not exist.
var ErrRequestNotFound = errors.New("friend request not found")

// ErrSelfRequest is returned when a user attempts to send a request to themselves.
var ErrSelfRequest = errors.New("cannot request self")

// ErrSenderMuted is returned when the sender is currently flood-muted.
var ErrSenderMuted = errors.New("sender is muted")

// ErrFriendNotInRoom is returned when the target friend is not in any room.
var ErrFriendNotInRoom = errors.New("friend not in a room")

// ErrFollowBlocked is returned when the friend has blocked follow-to-room.
var ErrFollowBlocked = errors.New("friend blocked following")

// ErrInvalidRelationship is returned when the relationship type is not registered.
var ErrInvalidRelationship = errors.New("invalid relationship type")

// ErrDuplicateRequest is returned when a request from this user already exists.
var ErrDuplicateRequest = errors.New("friend request already sent")
