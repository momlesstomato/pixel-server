package domain

// EntityType defines the type of room entity.
type EntityType int

const (
	// EntityPlayer represents a human user entity.
	EntityPlayer EntityType = iota
	// EntityBot represents a server-controlled entity.
	EntityBot
	// EntityPet represents a player-owned pet entity.
	EntityPet
)

// RoomEntity represents a positioned entity inside a room instance.
type RoomEntity struct {
	// VirtualID stores the room-scoped entity identifier.
	VirtualID int
	// Type stores the entity classification.
	Type EntityType
	// UserID stores the backing user identifier (players only).
	UserID int
	// ConnID stores the connection identifier (players only).
	ConnID string
	// Username stores the entity display name.
	Username string
	// Look stores the entity figure string.
	Look string
	// Motto stores the entity motto text.
	Motto string
	// Gender stores the entity gender code.
	Gender string
	// Position stores the current tile position.
	Position Tile
	// GoalPosition stores the walk destination tile.
	GoalPosition *Tile
	// Path stores the remaining walk path steps.
	Path []Tile
	// StepFrom stores the tile the entity stepped from in the current movement tick.
	// It is used as the broadcast position for smooth client-side walk animation.
	StepFrom *Tile
	// BodyRotation stores facing direction (0-7).
	BodyRotation int
	// HeadRotation stores head facing direction (0-7).
	HeadRotation int
	// Statuses stores active status key/value pairs.
	Statuses map[string]string
	// IsWalking reports whether the entity is currently walking.
	IsWalking bool
	// IsIdle reports whether the entity is idle.
	IsIdle bool
	// IdleTimer stores ticks since last user activity.
	IdleTimer int
	// CanWalk reports whether the entity is allowed to move.
	CanWalk bool
	// DanceID stores the current dance animation identifier.
	DanceID int
	// CarryItem stores the held hand-item identifier.
	CarryItem int
	// CarryTimer stores remaining carry ticks before drop.
	CarryTimer int
	// IsSitting reports whether the entity is currently sitting.
	IsSitting bool
	// IsSittingAuto reports whether the current sit was applied automatically by furniture (no Z adjustment).
	IsSittingAuto bool
	// UpdateNeeded marks the entity for broadcast this tick.
	UpdateNeeded bool
}

// NewPlayerEntity creates one player entity with initial state.
func NewPlayerEntity(virtualID int, userID int, connID string, username string, look string, motto string, gender string, position Tile) RoomEntity {
	return RoomEntity{
		VirtualID: virtualID, Type: EntityPlayer,
		UserID: userID, ConnID: connID,
		Username: username, Look: look, Motto: motto, Gender: gender,
		Position: position, BodyRotation: 2, HeadRotation: 2,
		Statuses: make(map[string]string), CanWalk: true,
	}
}
