package pathfinding

import "container/heap"

// node represents one A* search state in the priority queue.
type node struct {
	// x stores the horizontal coordinate.
	x int
	// y stores the vertical coordinate.
	y int
	// g stores the cost from start to this node.
	g float64
	// f stores the estimated total cost (g + h).
	f float64
	// parentX stores the parent node horizontal coordinate.
	parentX int
	// parentY stores the parent node vertical coordinate.
	parentY int
	// index stores the heap position for update operations.
	index int
}

// nodeHeap implements a min-heap of A* nodes ordered by f-cost.
type nodeHeap []*node

// Len returns the heap size.
func (h nodeHeap) Len() int { return len(h) }

// Less reports whether node i has lower f-cost than node j.
func (h nodeHeap) Less(i, j int) bool { return h[i].f < h[j].f }

// Swap exchanges two heap elements.
func (h nodeHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

// Push appends one node to the heap.
func (h *nodeHeap) Push(x any) {
	n := x.(*node)
	n.index = len(*h)
	*h = append(*h, n)
}

// Pop removes the minimum f-cost node from the heap.
func (h *nodeHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[:n-1]
	return item
}

// pushNode adds a node and maintains heap order.
func pushNode(h *nodeHeap, n *node) {
	heap.Push(h, n)
}

// popNode removes the best node and maintains heap order.
func popNode(h *nodeHeap) *node {
	return heap.Pop(h).(*node)
}
