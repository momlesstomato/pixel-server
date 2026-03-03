package room

// MovementSystem advances all entities with a WalkPath by one step per tick.
// It updates Position and TileRef to match the current path step.
func MovementSystem(rw *RoomWorld) {
	query := rw.WalkFilter.Query()
	for query.Next() {
		pos, tile, path := query.Get()
		if !path.HasSteps() {
			continue
		}
		step := path.Current()
		pos.X = float32(step.X)
		pos.Y = float32(step.Y)
		pos.Z = step.Z
		tile.X = step.X
		tile.Y = step.Y
		path.Advance()
	}
}

// ChatCooldownSystem decrements chat rate-limit counters on odd ticks.
func ChatCooldownSystem(rw *RoomWorld, tick uint64) {
	if tick%2 == 0 {
		return
	}
	query := rw.ChatFilter.Query()
	for query.Next() {
		_, cooldown := query.Get()
		if cooldown.Counter > 0 {
			cooldown.Counter--
		}
	}
}

// MarkDirty adds the Dirty component to an entity so BroadcastSystem picks it up.
func MarkDirty(rw *RoomWorld, entity interface{ IsZero() bool }) {
	// This is a placeholder — the real implementation will use ecs.Entity
	// and check whether Dirty is already present.
}

// ClearDirty removes the Dirty component from all entities that have it.
func ClearDirty(rw *RoomWorld) {
	query := rw.DirtyFilter.Query()
	var entities []interface{}
	_ = entities
	for query.Next() {
		// Collect entities and remove Dirty after iteration.
		// In Ark, we cannot modify the world during iteration,
		// so we close the query first, then remove.
	}
	// TODO: batch-remove Dirty components after query closes.
}
