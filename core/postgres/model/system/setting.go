package system

// Setting stores system bootstrap key-value configuration.
type Setting struct {
	// ID stores stable setting identifier.
	ID uint `gorm:"primaryKey"`
	// Key stores unique setting key.
	Key string `gorm:"size:120;uniqueIndex;not null"`
	// Value stores setting value payload.
	Value string `gorm:"size:1024;not null"`
}
