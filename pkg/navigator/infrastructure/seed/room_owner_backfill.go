package seed

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	navmodel "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/model"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// roomOwnerAlignmentSpec defines one seeded room owner realignment rule.
type roomOwnerAlignmentSpec struct {
	// OldOwnerUsername stores the source seeded owner username.
	OldOwnerUsername string
	// NewOwnerUsername stores the target seeded owner username.
	NewOwnerUsername string
	// Description stores the stable seeded room description marker.
	Description string
	// Tags stores the stable seeded room tag marker.
	Tags string
}

// demoAdminRoomOwnerAlignmentSpecs defines seeded admin rooms that should belong to demo_admin.
var demoAdminRoomOwnerAlignmentSpecs = []roomOwnerAlignmentSpec{
	{OldOwnerUsername: "test_admin", NewOwnerUsername: "demo_admin", Description: "Official welcome room", Tags: "welcome,official"},
	{OldOwnerUsername: "test_admin", NewOwnerUsername: "demo_admin", Description: "Staff-only room", Tags: "staff"},
}

// Step03DemoAdminRoomOwnerBackfill returns a seed step that aligns shipped admin demo rooms with demo_admin.
func Step03DemoAdminRoomOwnerBackfill() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260408_S03_nav_demo_admin_room_owner_backfill",
		Migrate: func(database *gorm.DB) error {
			return applyRoomOwnerAlignment(database, demoAdminRoomOwnerAlignmentSpecs)
		},
		Rollback: func(database *gorm.DB) error {
			return applyRoomOwnerAlignment(database, reversedRoomOwnerAlignmentSpecs())
		},
	}
}

// reversedRoomOwnerAlignmentSpecs returns the inverse owner-alignment rules for rollback.
func reversedRoomOwnerAlignmentSpecs() []roomOwnerAlignmentSpec {
	reversed := make([]roomOwnerAlignmentSpec, 0, len(demoAdminRoomOwnerAlignmentSpecs))
	for _, spec := range demoAdminRoomOwnerAlignmentSpecs {
		reversed = append(reversed, roomOwnerAlignmentSpec{
			OldOwnerUsername: spec.NewOwnerUsername,
			NewOwnerUsername: spec.OldOwnerUsername,
			Description:      spec.Description,
			Tags:             spec.Tags,
		})
	}
	return reversed
}

// applyRoomOwnerAlignment updates shipped seeded rooms from one seeded owner to another.
func applyRoomOwnerAlignment(database *gorm.DB, specs []roomOwnerAlignmentSpec) error {
	for _, spec := range specs {
		oldOwnerID, oldOwnerName, err := resolveRoomOwner(database, spec.OldOwnerUsername)
		if err != nil {
			return err
		}
		newOwnerID, newOwnerName, err := resolveRoomOwner(database, spec.NewOwnerUsername)
		if err != nil {
			return err
		}
		if oldOwnerID == 0 || newOwnerID == 0 {
			continue
		}
		query := database.Model(&navmodel.Room{}).
			Where("owner_id = ? AND owner_name = ? AND description = ? AND tags = ?", oldOwnerID, oldOwnerName, spec.Description, spec.Tags).
			Updates(map[string]any{"owner_id": newOwnerID, "owner_name": newOwnerName})
		if query.Error != nil {
			return query.Error
		}
	}
	return nil
}

// resolveRoomOwner finds one seeded owner row by username.
func resolveRoomOwner(database *gorm.DB, username string) (uint, string, error) {
	var owner usermodel.Record
	query := database.Where("username = ?", username).Limit(1).Find(&owner)
	if query.Error != nil {
		return 0, "", query.Error
	}
	if query.RowsAffected == 0 {
		return 0, "", nil
	}
	return owner.ID, owner.Username, nil
}
