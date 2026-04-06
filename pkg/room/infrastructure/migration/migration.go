package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	roommodel "github.com/momlesstomato/pixel-server/pkg/room/infrastructure/model"
	"gorm.io/gorm"
)

// Step01RoomModels returns the migration that creates the room_models table.
func Step01RoomModels() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260401_01_room_models",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&roommodel.RoomModel{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&roommodel.RoomModel{})
		},
	}
}

// Step02RoomExtension returns the migration that adds room realm columns to the rooms table.
func Step02RoomExtension() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260401_02_room_extension",
		Migrate: func(database *gorm.DB) error {
			type rooms struct {
				ModelSlug       string  `gorm:"column:model_slug;size:50;not null;default:model_a"`
				CustomHeightmap *string `gorm:"column:custom_heightmap;type:text;default:null"`
				WallHeight      int     `gorm:"column:wall_height;not null;default:-1"`
				FloorThickness  int     `gorm:"column:floor_thickness;not null;default:0"`
				WallThickness   int     `gorm:"column:wall_thickness;not null;default:0"`
				PasswordHash    string  `gorm:"column:password_hash;size:255;not null;default:''"`
				AllowPets       bool    `gorm:"column:allow_pets;not null;default:true"`
				AllowTrading    bool    `gorm:"column:allow_trading;not null;default:false"`
			}
			return database.Table("rooms").AutoMigrate(&rooms{})
		},
		Rollback: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			columns := []string{"model_slug", "custom_heightmap", "wall_height", "floor_thickness", "wall_thickness", "password_hash", "allow_pets", "allow_trading"}
			for _, col := range columns {
				if database.Migrator().HasColumn("rooms", col) {
					if err := database.Migrator().DropColumn("rooms", col); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
}

// Step03RoomBans returns the migration that creates the room_bans table.
func Step03RoomBans() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260401_03_room_bans",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&roommodel.RoomBan{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&roommodel.RoomBan{})
		},
	}
}

// Step04RoomRights returns the migration that creates the room_rights table.
func Step04RoomRights() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260401_04_room_rights",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&roommodel.RoomRight{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&roommodel.RoomRight{})
		},
	}
}
