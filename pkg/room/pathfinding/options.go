package pathfinding

// Options configures pathfinding behavior.
type Options struct {
	// AllowDiagonal enables 8-directional movement.
	AllowDiagonal bool
	// MaxStepHeight stores the maximum height change per step.
	MaxStepHeight float64
	// MaxIterations limits search iterations to prevent runaway.
	MaxIterations int
	// HeightCostEnabled enables height-aware movement costs.
	HeightCostEnabled bool
	// AscentMultiplier scales cost for upward movement.
	AscentMultiplier float64
	// DescentMultiplier scales cost for downward movement.
	DescentMultiplier float64
}

// DefaultOptions returns standard pathfinding configuration.
func DefaultOptions() Options {
	return Options{
		AllowDiagonal:     true,
		MaxStepHeight:     1.5,
		MaxIterations:     10000,
		HeightCostEnabled: false,
		AscentMultiplier:  2.0,
		DescentMultiplier: 0.5,
	}
}
