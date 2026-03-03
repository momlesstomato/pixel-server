//go:build !linux && !darwin

package plugin

import "errors"

// LoadFromFile is not supported on this platform.
// The Go plugin package only works on Linux and macOS.
func LoadFromFile(_ string) (Plugin, error) {
	return nil, errors.New("plugin loading is not supported on this platform (Linux and macOS only)")
}
