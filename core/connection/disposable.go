package connection

// Disposable defines a release lifecycle for closeable resources.
type Disposable interface {
	// Dispose releases all held resources.
	Dispose() error
}

// DisposeFunc adapts a function to implement Disposable.
type DisposeFunc func() error

// Dispose executes the wrapped release function.
func (value DisposeFunc) Dispose() error {
	return value()
}
