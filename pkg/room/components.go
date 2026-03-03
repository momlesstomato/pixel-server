package room

// Component types for room entities.
// All components are plain structs with no methods — logic lives in systems.

// Position in tile-space. Z is stack height (0.0 = floor, 1.0 = one furniture height, etc.).
type Position struct {
	X, Y float32
	Z    float32
}

// TileRef is the grid-snapped tile index for collision and pathfinding lookups.
type TileRef struct {
	X, Y int16
}

// PathStep is a single waypoint on a walk path.
type PathStep struct {
	X, Y int16
	Z    float32
}

// WalkPath holds the ordered walk steps assigned to this entity.
type WalkPath struct {
	Steps  []PathStep
	Cursor int
}

// HasSteps reports whether there are remaining steps to traverse.
func (w *WalkPath) HasSteps() bool {
	return w.Cursor < len(w.Steps)
}

// Current returns the current step, or the zero value if exhausted.
func (w *WalkPath) Current() PathStep {
	if w.Cursor >= len(w.Steps) {
		return PathStep{}
	}
	return w.Steps[w.Cursor]
}

// Advance moves the cursor to the next step.
func (w *WalkPath) Advance() {
	if w.Cursor < len(w.Steps) {
		w.Cursor++
	}
}

// Kind constants for EntityKind.
const (
	KindAvatar uint8 = 1
	KindBot    uint8 = 2
	KindPet    uint8 = 3
	KindItem   uint8 = 4
)

// EntityKind distinguishes the simulation role of an entity.
type EntityKind struct {
	Kind uint8
}

// AvatarID links an ECS entity to the database user record and room-scoped unit index.
type AvatarID struct {
	UserID   int64
	RoomUnit int32 // room-scoped unit index sent to clients
}

// Status encodes posture and visual effects as compact bit fields.
type Status struct {
	Posture uint8  // sit=1 stand=2 lay=3 wave=4 …
	Effects uint32 // bitmask of active effect IDs
}

// Posture constants.
const (
	PostureSit   uint8 = 1
	PostureStand uint8 = 2
	PostureLay   uint8 = 3
	PostureWave  uint8 = 4
)

// ChatCooldown tracks the rate-limiter counter (decremented every odd tick).
type ChatCooldown struct {
	Counter int32
}

// BotAI is present only on bot entities.
type BotAI struct {
	Behaviour uint8
	ChatLines []string
	ChatIndex int
}

// PetAI is present only on pet entities.
type PetAI struct {
	HappyLevel int32
	Energy     int32
}

// ItemInteraction is present only on interactive floor/wall items.
type ItemInteraction struct {
	FurniID    int64
	ExtraData  string
	CycleCount int
}

// Dirty flags an entity as having state that must be broadcast this tick.
type Dirty struct{}
