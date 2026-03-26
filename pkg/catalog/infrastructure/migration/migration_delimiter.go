package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Step07PageDelimiter returns the migration that converts the images and texts
// columns on catalog_pages from comma-delimited to pipe-delimited storage,
// aligning with the Habbo ecosystem convention used by the protocol client.
// Only rows that contain commas but no pipes are updated, preserving rows that
// were already inserted with the correct delimiter.
func Step07PageDelimiter() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260324_07c_catalog_pages_pipe_delimiter",
		Migrate: func(database *gorm.DB) error {
			if err := database.Exec(`
				UPDATE catalog_pages
				SET images = replace(images, ',', '|')
				WHERE images <> '' AND images NOT LIKE '%|%'
			`).Error; err != nil {
				return err
			}
			return database.Exec(`
				UPDATE catalog_pages
				SET texts = replace(texts, ',', '|')
				WHERE texts <> '' AND texts NOT LIKE '%|%'
			`).Error
		},
		Rollback: func(database *gorm.DB) error {
			if err := database.Exec(`
				UPDATE catalog_pages
				SET images = replace(images, '|', ',')
				WHERE images <> '' AND images NOT LIKE '%,%'
			`).Error; err != nil {
				return err
			}
			return database.Exec(`
				UPDATE catalog_pages
				SET texts = replace(texts, '|', ',')
				WHERE texts <> '' AND texts NOT LIKE '%,%'
			`).Error
		},
	}
}
