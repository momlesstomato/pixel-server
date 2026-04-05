package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"github.com/momlesstomato/pixel-server/pkg/moderation/infrastructure/model"
	"gorm.io/gorm"
)

// Step02Phase2Tables creates Phase 2 moderation tables.
func Step02Phase2Tables() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260404_02_moderation_phase2",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&model.ModerationTicket{}); err != nil {
				return err
			}
			if err := tx.AutoMigrate(&model.ModerationWordFilter{}); err != nil {
				return err
			}
			if err := tx.AutoMigrate(&model.ModerationPreset{}); err != nil {
				return err
			}
			if err := tx.AutoMigrate(&model.ModerationRoomVisit{}); err != nil {
				return err
			}
			indexes := []string{
				"CREATE INDEX IF NOT EXISTS idx_mod_tickets_status ON moderation_tickets (status, created_at DESC)",
				"CREATE INDEX IF NOT EXISTS idx_mod_tickets_reporter ON moderation_tickets (reporter_id)",
				"CREATE INDEX IF NOT EXISTS idx_mod_wordfilters_scope ON moderation_word_filters (scope, active)",
				"CREATE INDEX IF NOT EXISTS idx_mod_visits_user ON moderation_room_visits (user_id, visited_at DESC)",
				"CREATE INDEX IF NOT EXISTS idx_mod_visits_room ON moderation_room_visits (room_id, visited_at DESC)",
			}
			for _, idx := range indexes {
				if err := tx.Exec(idx).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			for _, table := range []string{"moderation_room_visits", "moderation_presets", "moderation_word_filters", "moderation_tickets"} {
				if err := tx.Migrator().DropTable(table); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
