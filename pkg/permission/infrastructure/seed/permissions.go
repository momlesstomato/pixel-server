package seed

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Step02DefaultPermissions returns seed step for essential permission grants.
func Step02DefaultPermissions() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260314_02_default_group_permissions",
		Migrate: func(database *gorm.DB) error {
			return ensurePermissions(database)
		},
		Rollback: func(database *gorm.DB) error {
			return database.Where("permission IN ?", []string{
				"*", "perk.*", "perk.safe_chat", "perk.helpers", "perk.citizen",
				"moderation.kick", "moderation.mute", "moderation.alert",
				"messenger.friends.extended", "messenger.flood.bypass",
			}).Delete(&permissionmodel.Grant{}).Error
		},
	}
}

// ensurePermissions creates essential grants for default groups.
func ensurePermissions(database *gorm.DB) error {
	nameToID := map[string]uint{}
	var groups []permissionmodel.Group
	if err := database.Where("name IN ?", []string{"default", "vip", "moderator", "admin"}).Find(&groups).Error; err != nil {
		return err
	}
	for _, group := range groups {
		nameToID[group.Name] = group.ID
	}
	grants := []permissionmodel.Grant{
		{GroupID: nameToID["default"], Permission: "perk.safe_chat"},
		{GroupID: nameToID["default"], Permission: "perk.helpers"},
		{GroupID: nameToID["default"], Permission: "perk.citizen"},
		{GroupID: nameToID["vip"], Permission: "perk.*"},
		{GroupID: nameToID["vip"], Permission: "messenger.friends.extended"},
		{GroupID: nameToID["moderator"], Permission: "perk.*"},
		{GroupID: nameToID["moderator"], Permission: "moderation.kick"},
		{GroupID: nameToID["moderator"], Permission: "moderation.mute"},
		{GroupID: nameToID["moderator"], Permission: "moderation.alert"},
		{GroupID: nameToID["moderator"], Permission: "messenger.flood.bypass"},
		{GroupID: nameToID["admin"], Permission: "*"},
	}
	filtered := make([]permissionmodel.Grant, 0, len(grants))
	for _, grant := range grants {
		if grant.GroupID > 0 {
			filtered = append(filtered, grant)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return database.Clauses(clause.OnConflict{DoNothing: true}).Create(&filtered).Error
}
