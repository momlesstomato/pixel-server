package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	roommodel "github.com/momlesstomato/pixel-server/pkg/room/infrastructure/model"
	"gorm.io/gorm"
)

// Step05ChatLogs returns the migration that creates the room_chat_logs table.
func Step05ChatLogs() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260402_05_room_chat_logs",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&roommodel.ChatLog{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&roommodel.ChatLog{})
		},
	}
}

// Step06RoomVotes returns the migration that creates the room_votes table.
func Step06RoomVotes() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260402_06_room_votes",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&roommodel.RoomVote{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&roommodel.RoomVote{})
		},
	}
}

// Step07RoomSoftDelete returns the migration that adds deleted_at to the rooms table.
func Step07RoomSoftDelete() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260402_07_room_soft_delete",
		Migrate: func(database *gorm.DB) error {
			type rooms struct {
				DeletedAt *string `gorm:"column:deleted_at;type:timestamptz;default:null"`
			}
			return database.Table("rooms").AutoMigrate(&rooms{})
		},
		Rollback: func(database *gorm.DB) error {
			if database.Migrator().HasColumn("rooms", "deleted_at") {
				return database.Migrator().DropColumn("rooms", "deleted_at")
			}
			return nil
		},
	}
}

// Step08RoomForward returns the migration that adds forward_room_id to the rooms table.
func Step08RoomForward() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260404_08_room_forward",
		Migrate: func(database *gorm.DB) error {
			type rooms struct {
				ForwardRoomID *int `gorm:"column:forward_room_id;default:null"`
			}
			return database.Table("rooms").AutoMigrate(&rooms{})
		},
		Rollback: func(database *gorm.DB) error {
			if database.Migrator().HasColumn("rooms", "forward_room_id") {
				return database.Migrator().DropColumn("rooms", "forward_room_id")
			}
			return nil
		},
	}
}
