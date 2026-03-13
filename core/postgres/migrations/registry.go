package migrations

import gormigrate "github.com/go-gormigrate/gormigrate/v2"

// Registry returns ordered schema migration steps.
func Registry() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		Step01Users(),
		Step02UserLoginEvents(),
	}
}
