package model

// RoomExtension defines additional room columns added by the room realm migration.
type RoomExtension struct {
	// ModelSlug stores the room model template identifier.
	ModelSlug string `gorm:"column:model_slug;size:50;not null;default:model_a"`
	// CustomHeightmap stores user-created heightmap override.
	CustomHeightmap string `gorm:"column:custom_heightmap;type:text;default:null"`
	// WallHeight stores custom wall height (-1 = auto).
	WallHeight int `gorm:"column:wall_height;not null;default:-1"`
	// FloorThickness stores floor rendering thickness.
	FloorThickness int `gorm:"column:floor_thickness;not null;default:0"`
	// WallThickness stores wall rendering thickness.
	WallThickness int `gorm:"column:wall_thickness;not null;default:0"`
	// PasswordHash stores bcrypt password hash for password rooms.
	PasswordHash string `gorm:"column:password_hash;size:255;not null;default:''"`
	// AllowPets stores whether pet placement is allowed.
	AllowPets bool `gorm:"column:allow_pets;not null;default:true"`
	// AllowTrading stores whether trading is enabled.
	AllowTrading bool `gorm:"column:allow_trading;not null;default:false"`
}
