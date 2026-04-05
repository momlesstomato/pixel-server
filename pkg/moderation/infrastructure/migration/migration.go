package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"github.com/momlesstomato/pixel-server/pkg/moderation/infrastructure/model"
	"gorm.io/gorm"
)

// Step01ModerationActions creates the moderation_actions table.
func Step01ModerationActions() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260404_01_moderation_actions",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&model.ModerationAction{}); err != nil {
				return err
			}
			indexes := []string{
				"CREATE INDEX IF NOT EXISTS idx_mod_actions_target ON moderation_actions (target_user_id, scope, active)",
				"CREATE INDEX IF NOT EXISTS idx_mod_actions_room ON moderation_actions (room_id, active) WHERE scope = 'room'",
				"CREATE INDEX IF NOT EXISTS idx_mod_actions_expires ON moderation_actions (expires_at) WHERE active = true",
				"CREATE INDEX IF NOT EXISTS idx_mod_actions_ip ON moderation_actions (ip_address) WHERE ip_address IS NOT NULL AND ip_address != ''",
			}
			for _, idx := range indexes {
				if err := tx.Exec(idx).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("moderation_actions")
		},
	}
}
