package pathfinding

import (
	"container/heap"
	"math"
)

// Cost constants for movement.
const (
	CostFlat     float32 = 1.0
	CostDiagonal float32 = 1.414
	CostClimb    float32 = 1.5 // per unit of positive dZ
	CostDescend  float32 = 0.8 // per unit of negative dZ
	MaxStepUp    float32 = 1.1 // maximum positive dZ per step
	MaxStepDown  float32 = 2.0 // maximum negative dZ per step
)

// Options control pathfinding behaviour per request.
type Options struct {
	AllowDiagonal       bool
	Flying              bool
	WalkthroughEntities bool
}

// FindPath returns an ordered slice of PathStep from start to goal,
// or nil if no path exists. The start tile is excluded from the result.
func FindPath(l *Layout, from, to Tile, opts Options) []PathStep {
	if !l.InBounds(int(from.X), int(from.Y)) || !l.InBounds(int(to.X), int(to.Y)) {
		return nil
	}
	dest := l.At(int(to.X), int(to.Y))
	if !opts.Flying && dest.State == TileBlocked {
		return nil
	}

	w, h := l.Width, l.Height
	size := w * h
	gScore := make([]float32, size)
	for i := range gScore {
		gScore[i] = math.MaxFloat32
	}
	parent := make([]int, size)
	for i := range parent {
		parent[i] = -1
	}
	closed := make([]bool, size)

	idx := func(x, y int) int { return y*w + x }
	startIdx := idx(int(from.X), int(from.Y))
	goalIdx := idx(int(to.X), int(to.Y))

	gScore[startIdx] = 0

	oh := &openHeap{}
	heap.Init(oh)
	heap.Push(oh, &astarNode{
		x: int(from.X), y: int(from.Y),
		g: 0, f: heuristic3d(&from, dest),
	})

	dirs8 := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}, {1, -1}, {1, 1}, {-1, 1}, {-1, -1}}
	dirs4 := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

	dirs := dirs4
	if opts.AllowDiagonal {
		dirs = dirs8
	}

	for oh.Len() > 0 {
		cur := heap.Pop(oh).(*astarNode)
		ci := idx(cur.x, cur.y)
		if ci == goalIdx {
			return reconstruct(l, parent, goalIdx, w)
		}
		if closed[ci] {
			continue
		}
		closed[ci] = true

		curTile := l.At(cur.x, cur.y)
		for _, d := range dirs {
			nx, ny := cur.x+d[0], cur.y+d[1]
			if !l.InBounds(nx, ny) {
				continue
			}
			ni := idx(nx, ny)
			if closed[ni] {
				continue
			}
			nTile := l.At(nx, ny)
			if !passable(nTile, opts) {
				continue
			}
			diagonal := d[0] != 0 && d[1] != 0
			cost := moveCost(curTile, nTile, diagonal, opts)
			if cost < 0 {
				continue // step too high
			}
			ng := gScore[ci] + cost
			if ng < gScore[ni] {
				gScore[ni] = ng
				parent[ni] = ci
				h := heuristic3d(nTile, dest)
				heap.Push(oh, &astarNode{x: nx, y: ny, g: ng, f: ng + h})
			}
		}
	}
	return nil
}

func passable(t *Tile, opts Options) bool {
	if opts.Flying {
		return true // flying entities traverse all tiles
	}
	return t.State != TileBlocked
}

func moveCost(from, to *Tile, diagonal bool, opts Options) float32 {
	if opts.Flying {
		if diagonal {
			return CostDiagonal
		}
		return CostFlat
	}
	dz := to.Z - from.Z
	if dz > MaxStepUp {
		return -1
	}
	if dz < -MaxStepDown {
		return -1
	}
	var base float32
	if diagonal {
		base = CostDiagonal
	} else {
		base = CostFlat
	}
	if dz > 0 {
		return base + dz*CostClimb
	}
	if dz < 0 {
		return base + abs32(dz)*CostDescend
	}
	return base
}

func heuristic3d(a, b *Tile) float32 {
	dx := abs32(float32(a.X - b.X))
	dy := abs32(float32(a.Y - b.Y))
	diag := min32(dx, dy)
	straight := (dx + dy) - 2*diag
	return CostFlat*straight + CostDiagonal*diag
}

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func reconstruct(l *Layout, parent []int, goalIdx, w int) []PathStep {
	var path []PathStep
	for ci := goalIdx; ci != -1; ci = parent[ci] {
		x := ci % w
		y := ci / w
		t := l.At(x, y)
		path = append(path, PathStep{X: int16(x), Y: int16(y), Z: t.Z})
	}
	// reverse
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	// remove start tile (caller already knows it)
	if len(path) > 1 {
		path = path[1:]
	}
	return path
}

// --- min-heap for A* open set ---

type astarNode struct {
	x, y int
	g, f float32
}

type openHeap []*astarNode

func (h openHeap) Len() int           { return len(h) }
func (h openHeap) Less(i, j int) bool { return h[i].f < h[j].f }
func (h openHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *openHeap) Push(x interface{}) { *h = append(*h, x.(*astarNode)) }

func (h *openHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return item
}
