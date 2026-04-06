package realtime

// normalizeFurnitureDir returns an even furniture direction in the 0-7 range.
func normalizeFurnitureDir(dir int) int {
	dir %= 8
	if dir < 0 {
		dir += 8
	}
	if dir%2 != 0 {
		dir--
	}
	return dir
}

// footprintTiles returns the rectangular tile footprint covered by one placed furniture item.
func footprintTiles(x, y, dir, width, length int) [][2]int {
	if width < 1 {
		width = 1
	}
	if length < 1 {
		length = 1
	}
	switch normalizeFurnitureDir(dir) {
	case 2, 6:
		width, length = length, width
	}
	tiles := make([][2]int, 0, width*length)
	for dy := 0; dy < length; dy++ {
		for dx := 0; dx < width; dx++ {
			tiles = append(tiles, [2]int{x + dx, y + dy})
		}
	}
	return tiles
}

// seatEntriesFromFootprint builds seat cache entries for every covered tile of one item.
func seatEntriesFromFootprint(itemID, x, y, dir int, height float64, width, length int, canSit, canLay bool) []seatEntry {
	if !canSit && !canLay {
		return nil
	}
	tiles := footprintTiles(x, y, dir, width, length)
	entries := make([]seatEntry, 0, len(tiles))
	for _, tile := range tiles {
		entries = append(entries, seatEntry{
			itemID:  itemID,
			x:       tile[0],
			y:       tile[1],
			height:  height,
			dir:     dir,
			canSit:  canSit,
			canLay:  canLay,
		})
	}
	return entries
}

// uniqueSeatTiles returns the unique tile coordinates covered by one seat entry slice.
func uniqueSeatTiles(entries []seatEntry) [][2]int {
	seen := make(map[[2]int]struct{}, len(entries))
	tiles := make([][2]int, 0, len(entries))
	for _, entry := range entries {
		tile := [2]int{entry.x, entry.y}
		if _, ok := seen[tile]; ok {
			continue
		}
		seen[tile] = struct{}{}
		tiles = append(tiles, tile)
	}
	return tiles
}

// sameSeatTiles reports whether two seat entry slices cover the same tile set.
func sameSeatTiles(left, right []seatEntry) bool {
	leftTiles := uniqueSeatTiles(left)
	rightTiles := uniqueSeatTiles(right)
	if len(leftTiles) != len(rightTiles) {
		return false
	}
	rightSet := make(map[[2]int]struct{}, len(rightTiles))
	for _, tile := range rightTiles {
		rightSet[tile] = struct{}{}
	}
	for _, tile := range leftTiles {
		if _, ok := rightSet[tile]; !ok {
			return false
		}
	}
	return true
}