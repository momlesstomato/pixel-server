package domain

import "errors"

// ErrRoomNotFound indicates the requested room does not exist.
var ErrRoomNotFound = errors.New("room not found")

// ErrRoomModelNotFound indicates the requested room model does not exist.
var ErrRoomModelNotFound = errors.New("room model not found")

// ErrRoomFull indicates the room has reached maximum capacity.
var ErrRoomFull = errors.New("room is full")

// ErrRoomBanned indicates the user is banned from the room.
var ErrRoomBanned = errors.New("user is banned from this room")

// ErrAccessDenied indicates the user does not have access to the room.
var ErrAccessDenied = errors.New("access denied")

// ErrInvalidPassword indicates the supplied room password is incorrect.
var ErrInvalidPassword = errors.New("invalid room password")

// ErrInvalidHeightmap indicates the heightmap string is malformed.
var ErrInvalidHeightmap = errors.New("invalid heightmap format")

// ErrEntityNotFound indicates the entity is not present in the room.
var ErrEntityNotFound = errors.New("entity not found")

// ErrPathBlocked indicates no walkable path exists to the target.
var ErrPathBlocked = errors.New("path is blocked")

// ErrAlreadyInRoom indicates the user is already present in the room.
var ErrAlreadyInRoom = errors.New("user is already in the room")

// ErrFloodControl indicates the entity is temporarily muted by flood protection.
var ErrFloodControl = errors.New("flood control active")
