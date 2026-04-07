package realtime

type seatTarget struct {
	x int
	y int
}

func bestSeatEntryForTile(entries []seatEntry, x, y int) (seatEntry, bool) {
	bestHeight := 0.0
	found := false
	best := seatEntry{}
	for _, entry := range entries {
		if entry.x != x || entry.y != y {
			continue
		}
		if !found || entry.height >= bestHeight {
			best = entry
			bestHeight = entry.height
			found = true
		}
	}
	return best, found
}

func seatTargetsForItem(entries []seatEntry, itemID int) []seatTarget {
	seen := make(map[[2]int]struct{})
	targets := make([]seatTarget, 0)
	for _, entry := range entries {
		if entry.itemID != itemID || (!entry.canSit && !entry.canLay) {
			continue
		}
		key := [2]int{entry.anchorX, entry.anchorY}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		targets = append(targets, seatTarget{x: entry.anchorX, y: entry.anchorY})
	}
	return targets
}

func seatTargetDistance(x, y int, target seatTarget) int {
	dx := target.x - x
	if dx < 0 {
		dx = -dx
	}
	dy := target.y - y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func isBetterSeatFallback(current seatTarget, currentDistance int, candidate seatTarget, candidateDistance int) bool {
	if candidateDistance != currentDistance {
		return candidateDistance < currentDistance
	}
	if candidate.y != current.y {
		return candidate.y < current.y
	}
	return candidate.x < current.x
}

func (runtime *Runtime) resolveSeatTarget(roomID int, entries []seatEntry, clickedX, clickedY int) (targetX, targetY int, ok bool) {
	entry, found := bestSeatEntryForTile(entries, clickedX, clickedY)
	if !found {
		return 0, 0, false
	}
	preferred := seatTarget{x: entry.anchorX, y: entry.anchorY}
	targets := seatTargetsForItem(entries, entry.itemID)
	fallback := seatTarget{}
	fallbackDistance := 0
	fallbackFound := false
	preferredFound := false
	for _, target := range targets {
		if target == preferred {
			preferredFound = true
			if !runtime.isTileOccupied(roomID, target.x, target.y) {
				return target.x, target.y, true
			}
			continue
		}
		if runtime.isTileOccupied(roomID, target.x, target.y) {
			continue
		}
		distance := seatTargetDistance(clickedX, clickedY, target)
		if !fallbackFound || isBetterSeatFallback(fallback, fallbackDistance, target, distance) {
			fallback = target
			fallbackDistance = distance
			fallbackFound = true
		}
	}
	if fallbackFound {
		return fallback.x, fallback.y, true
	}
	if preferredFound {
		return preferred.x, preferred.y, true
	}
	return 0, 0, false
}
