package seeds

import gormigrate "github.com/go-gormigrate/gormigrate/v2"

// Registry returns ordered essential seed steps.
func Registry() []*gormigrate.Migration {
	return []*gormigrate.Migration{}
}
