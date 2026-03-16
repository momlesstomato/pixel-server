package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	messengermodel "github.com/momlesstomato/pixel-server/pkg/messenger/infrastructure/model"
	"gorm.io/gorm"
)

// Step01MessengerFriendships returns the migration that creates the messenger_friendships table.
func Step01MessengerFriendships() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260315_01_messenger_friendships",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&messengermodel.Friendship{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&messengermodel.Friendship{})
		},
	}
}

// Step02FriendRequests returns the migration that creates the friend_requests table.
func Step02FriendRequests() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260315_02_friend_requests",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&messengermodel.Request{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&messengermodel.Request{})
		},
	}
}

// Step03OfflineMessages returns the migration that creates the offline_messages table.
func Step03OfflineMessages() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260315_03_offline_messages",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&messengermodel.OfflineMessage{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&messengermodel.OfflineMessage{})
		},
	}
}

// Step04NormalizeFriendships returns the migration that canonicalizes friendship rows to one tuple per pair.
func Step04NormalizeFriendships() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260315_04_normalize_messenger_friendships",
		Migrate: func(database *gorm.DB) error {
			if err := database.Exec(`
				CREATE TABLE IF NOT EXISTS messenger_friendships_v2 (
					user_one_id BIGINT NOT NULL,
					user_two_id BIGINT NOT NULL,
					relationship SMALLINT NOT NULL DEFAULT 0,
					created_at TIMESTAMPTZ,
					PRIMARY KEY (user_one_id, user_two_id)
				)
			`).Error; err != nil {
				return err
			}
			if err := database.Exec(`
				INSERT INTO messenger_friendships_v2 (user_one_id, user_two_id, relationship, created_at)
				SELECT CASE WHEN user_one_id <= user_two_id THEN user_one_id ELSE user_two_id END,
				       CASE WHEN user_one_id <= user_two_id THEN user_two_id ELSE user_one_id END,
				       MAX(relationship),
				       MIN(created_at)
				FROM messenger_friendships
				GROUP BY
				       CASE WHEN user_one_id <= user_two_id THEN user_one_id ELSE user_two_id END,
				       CASE WHEN user_one_id <= user_two_id THEN user_two_id ELSE user_one_id END
				ON CONFLICT (user_one_id, user_two_id) DO NOTHING
			`).Error; err != nil {
				return err
			}
			if err := database.Migrator().DropTable("messenger_friendships"); err != nil {
				return err
			}
			return database.Exec(`ALTER TABLE messenger_friendships_v2 RENAME TO messenger_friendships`).Error
		},
		Rollback: func(database *gorm.DB) error {
			if err := database.Exec(`
				CREATE TABLE IF NOT EXISTS messenger_friendships_v1 (
					user_one_id BIGINT NOT NULL,
					user_two_id BIGINT NOT NULL,
					relationship SMALLINT NOT NULL DEFAULT 0,
					created_at TIMESTAMPTZ,
					PRIMARY KEY (user_one_id, user_two_id)
				)
			`).Error; err != nil {
				return err
			}
			if err := database.Exec(`
				INSERT INTO messenger_friendships_v1 (user_one_id, user_two_id, relationship, created_at)
				SELECT user_one_id, user_two_id, relationship, created_at FROM messenger_friendships
				UNION ALL
				SELECT user_two_id, user_one_id, relationship, created_at FROM messenger_friendships
			`).Error; err != nil {
				return err
			}
			if err := database.Migrator().DropTable("messenger_friendships"); err != nil {
				return err
			}
			return database.Exec(`ALTER TABLE messenger_friendships_v1 RENAME TO messenger_friendships`).Error
		},
	}
}
