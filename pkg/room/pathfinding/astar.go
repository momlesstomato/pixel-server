package pathfinding

import (
	"math"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
)

// cardinalCost stores the movement cost for cardinal directions.
const cardinalCost = 10.0

// diagonalCost stores the movement cost for diagonal directions.
const diagonalCost = 14.0

// directions stores relative offsets for 8-directional movement.
var directions = [][2]int{
	{0, -1}, {1, 0}, {0, 1}, {-1, 0},
	{1, -1}, {1, 1}, {-1, 1}, {-1, -1},
}

// FindPath computes the shortest walkable path between two positions.
func FindPath(grid Grid, startX, startY, endX, endY int, opts Options) []domain.Tile {
	if !grid.IsWalkable(endX, endY) {
		return nil
	}
	if startX == endX && startY == endY {
		return nil
	}
	dirCount := 4
	if opts.AllowDiagonal {
		dirCount = 8
	}
	open := &nodeHeap{}
	closed := make([]bool, grid.Width()*grid.Height())
	nodes := make(map[int]*node)
	start := &node{x: startX, y: startY, g: 0, f: 0, parentX: -1, parentY: -1}
	start.f = heuristic(startX, startY, endX, endY)
	pushNode(open, start)
	nodes[key(startX, startY, grid.Width())] = start
	iterations := 0
	for open.Len() > 0 {
		iterations++
		if iterations > opts.MaxIterations {
			return nil
		}
		current := popNode(open)
		if current.x == endX && current.y == endY {
			return tracePath(current, nodes, grid)
		}
		closed[key(current.x, current.y, grid.Width())] = true
		for i := 0; i < dirCount; i++ {
			nx, ny := current.x+directions[i][0], current.y+directions[i][1]
			if !grid.IsWalkable(nx, ny) {
				continue
			}
			if closed[key(nx, ny, grid.Width())] {
				continue
			}
			if i >= 4 && !isDiagonalOpen(grid, current.x, current.y, directions[i]) {
				continue
			}
			heightDelta := math.Abs(grid.HeightAt(nx, ny) - grid.HeightAt(current.x, current.y))
			if heightDelta > opts.MaxStepHeight {
				continue
			}
			moveCost := cardinalCost
			if i >= 4 {
				moveCost = diagonalCost
			}
			if opts.HeightCostEnabled {
				moveCost += heightCost(grid, current.x, current.y, nx, ny, opts)
			}
			ng := current.g + moveCost
			k := key(nx, ny, grid.Width())
			existing, found := nodes[k]
			if found && ng >= existing.g {
				continue
			}
			if !found {
				existing = &node{x: nx, y: ny}
				nodes[k] = existing
			}
			existing.g = ng
			existing.f = ng + heuristic(nx, ny, endX, endY)
			existing.parentX = current.x
			existing.parentY = current.y
			if !found {
				pushNode(open, existing)
			}
		}
	}
	return nil
}

// heuristic computes Manhattan distance scaled by cardinal cost.
func heuristic(x1, y1, x2, y2 int) float64 {
	dx := math.Abs(float64(x1 - x2))
	dy := math.Abs(float64(y1 - y2))
	return (dx + dy) * cardinalCost
}

// isDiagonalOpen checks that both adjacent cardinals are passable.
func isDiagonalOpen(grid Grid, x, y int, dir [2]int) bool {
	return grid.IsWalkable(x+dir[0], y) && grid.IsWalkable(x, y+dir[1])
}

// heightCost computes extra cost based on elevation change.
func heightCost(grid Grid, cx, cy, nx, ny int, opts Options) float64 {
	delta := grid.HeightAt(nx, ny) - grid.HeightAt(cx, cy)
	if delta > 0 {
		return delta * opts.AscentMultiplier * cardinalCost
	}
	if delta < 0 {
		return math.Abs(delta) * opts.DescentMultiplier * cardinalCost
	}
	return 0
}

// key computes a unique flat index for a grid coordinate.
func key(x, y, width int) int {
	return y*width + x
}

// tracePath reconstructs the path from end to start via parent links.
func tracePath(end *node, nodes map[int]*node, grid Grid) []domain.Tile {
	path := []domain.Tile{}
	current := end
	for current.parentX >= 0 {
		path = append(path, domain.Tile{
			X: current.x, Y: current.y, Z: grid.HeightAt(current.x, current.y),
			State: domain.TileOpen,
		})
		current = nodes[key(current.parentX, current.parentY, grid.Width())]
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}
