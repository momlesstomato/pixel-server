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

// footprintSize returns the oriented width and length for one furniture footprint.
func footprintSize(dir, width, length int) (int, int) {
	if width < 1 {
		width = 1
	}
	if length < 1 {
		length = 1
	}
	switch normalizeFurnitureDir(dir) {
	case 2, 6:
		return length, width
	default:
		return width, length
	}
}

// footprintTiles returns the rectangular tile footprint covered by one placed furniture item.
func footprintTiles(x, y, dir, width, length int) [][2]int {
	width, length = footprintSize(dir, width, length)
	tiles := make([][2]int, 0, width*length)
	for dy := 0; dy < length; dy++ {
		for dx := 0; dx < width; dx++ {
			tiles = append(tiles, [2]int{x + dx, y + dy})
		}
	}
	return tiles
}

// laySlotAnchorTile returns the target tile for one lay slot within a multi-tile bed.
func laySlotAnchorTile(x, y, dir, width, length, dx, dy int) [2]int {
	if width < 1 {
		width = 1
	}
	if length < 1 {
		length = 1
	}
	switch normalizeFurnitureDir(dir) {
	case 2:
		return [2]int{x + length - 1, y + dy}
	case 4:
		return [2]int{x + dx, y + length - 1}
	case 6:
		return [2]int{x, y + dy}
	default:
		return [2]int{x + dx, y}
	}
}

// seatEntriesFromFootprint builds seat cache entries for every covered tile of one item.
func seatEntriesFromFootprint(itemID, x, y, dir int, height float64, width, length int, canSit, canLay bool) []seatEntry {
	if !canSit && !canLay {
		return nil
	}
	footprintWidth, footprintLength := footprintSize(dir, width, length)
	entries := make([]seatEntry, 0, footprintWidth*footprintLength)
	for dy := 0; dy < footprintLength; dy++ {
		for dx := 0; dx < footprintWidth; dx++ {
			tile := [2]int{x + dx, y + dy}
			anchor := tile
			entryCanLay := false
			if canLay {
				anchor = laySlotAnchorTile(x, y, dir, width, length, dx, dy)
				entryCanLay = tile == anchor
			}
			entries = append(entries, seatEntry{
				itemID:  itemID,
				x:       tile[0],
				y:       tile[1],
				anchorX: anchor[0],
				anchorY: anchor[1],
				height:  height,
				dir:     dir,
				canSit:  canSit,
				canLay:  entryCanLay,
			})
		}
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
