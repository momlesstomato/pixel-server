package realtime

import furnituredomain "github.com/momlesstomato/pixel-server/pkg/furniture/domain"

// blockEntry caches one blocked floor tile for a placed item.
type blockEntry struct {
	// itemID stores the placed item identifier for cache invalidation.
	itemID int
	// x stores the blocked tile horizontal coordinate.
	x int
	// y stores the blocked tile vertical coordinate.
	y int
}

func shouldBlockFloorItem(def furnituredomain.Definition) bool {
	return def.ItemType == furnituredomain.ItemTypeFloor && !def.IsWalkable && !def.CanSit && !def.CanLay
}

func blockEntriesFromFootprint(itemID, x, y, dir, width, length int) []blockEntry {
	tiles := footprintTiles(x, y, dir, width, length)
	entries := make([]blockEntry, 0, len(tiles))
	for _, tile := range tiles {
		entries = append(entries, blockEntry{itemID: itemID, x: tile[0], y: tile[1]})
	}
	return entries
}

// TileBlockCheckerFor reports whether one room tile is blocked by a cached floor-item footprint.
func (runtime *Runtime) TileBlockCheckerFor(roomID, x, y int) bool {
	runtime.seatMu.RLock()
	defer runtime.seatMu.RUnlock()
	for _, entry := range runtime.blockCache[roomID] {
		if entry.x == x && entry.y == y {
			return true
		}
	}
	return false
}

func (runtime *Runtime) replaceBlockEntries(roomID, itemID int, entries []blockEntry) {
	runtime.seatMu.Lock()
	defer runtime.seatMu.Unlock()
	filtered := runtime.blockCache[roomID][:0]
	for _, entry := range runtime.blockCache[roomID] {
		if entry.itemID != itemID {
			filtered = append(filtered, entry)
		}
	}
	runtime.blockCache[roomID] = append(filtered, entries...)
}

func (runtime *Runtime) removeBlockEntries(roomID, itemID int) {
	runtime.seatMu.Lock()
	defer runtime.seatMu.Unlock()
	filtered := runtime.blockCache[roomID][:0]
	for _, entry := range runtime.blockCache[roomID] {
		if entry.itemID != itemID {
			filtered = append(filtered, entry)
		}
	}
	runtime.blockCache[roomID] = filtered
}

func (runtime *Runtime) clearRoomPlacementEntries(roomID int) {
	runtime.seatMu.Lock()
	defer runtime.seatMu.Unlock()
	delete(runtime.seatCache, roomID)
	delete(runtime.blockCache, roomID)
}

func (runtime *Runtime) syncFloorItemEntries(roomID int, item furnituredomain.Item, def furnituredomain.Definition) {
	if def.CanSit || def.CanLay {
		runtime.replaceSeatEntries(roomID, item.ID, seatEntriesFromFootprint(item.ID, item.X, item.Y, item.Dir, runtime.effectiveStackHeight(item, def), def.Width, def.Length, def.CanSit, def.CanLay))
	} else {
		runtime.removeSeatEntries(roomID, item.ID)
	}
	if shouldBlockFloorItem(def) {
		runtime.replaceBlockEntries(roomID, item.ID, blockEntriesFromFootprint(item.ID, item.X, item.Y, item.Dir, def.Width, def.Length))
		return
	}
	runtime.removeBlockEntries(roomID, item.ID)
}
