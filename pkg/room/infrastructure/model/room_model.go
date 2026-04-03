package model

// RoomModel stores one predefined room template in PostgreSQL.
type RoomModel struct {
	// ID stores the stable model identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// Slug stores the model string identifier.
	Slug string `gorm:"size:50;uniqueIndex;not null"`
	// Heightmap stores the raw heightmap string.
	Heightmap string `gorm:"type:text;not null"`
	// DoorX stores the door tile horizontal coordinate.
	DoorX int `gorm:"not null;default:0"`
	// DoorY stores the door tile vertical coordinate.
	DoorY int `gorm:"not null;default:0"`
	// DoorZ stores the door tile height value.
	DoorZ float64 `gorm:"not null;default:0"`
	// DoorDir stores the door facing direction (0-7).
	DoorDir int `gorm:"not null;default:2"`
	// WallHeight stores custom wall height (-1 = auto).
	WallHeight int `gorm:"not null;default:-1"`
}

// TableName returns the PostgreSQL table name for RoomModel.
func (RoomModel) TableName() string { return "room_models" }
