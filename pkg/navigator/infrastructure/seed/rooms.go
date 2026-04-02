package seed

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	navmodel "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/model"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// demoRoomSpec defines one demo room seeding specification.
type demoRoomSpec struct {
	// OwnerUsername stores the owning test user name.
	OwnerUsername string
	// Name stores the room display name.
	Name string
	// Description stores the room description.
	Description string
	// State stores the room access state.
	State string
	// MaxUsers stores the room capacity.
	MaxUsers int
	// Tags stores comma-separated tags.
	Tags string
}

// demoRooms defines demo rooms assigned to test/demo users.
var demoRooms = []demoRoomSpec{
	{OwnerUsername: "test_admin", Name: "Welcome Lounge", Description: "Official welcome room", State: "open", MaxUsers: 50, Tags: "welcome,official"},
	{OwnerUsername: "test_admin", Name: "Staff Office", Description: "Staff-only room", State: "password", MaxUsers: 10, Tags: "staff"},
	{OwnerUsername: "test_default", Name: "Chill Room", Description: "Relax and chat", State: "open", MaxUsers: 25, Tags: "chill,social"},
	{OwnerUsername: "demo_vip", Name: "VIP Lounge", Description: "Exclusive VIP hangout", State: "locked", MaxUsers: 20, Tags: "vip,exclusive"},
	{OwnerUsername: "demo_default", Name: "Trade Hub", Description: "Trading room for everyone", State: "open", MaxUsers: 30, Tags: "trade,marketplace"},
}

// Step02DemoRooms returns seed step for demo rooms.
func Step02DemoRooms() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260326_S02_nav_demo_rooms",
		Migrate: func(database *gorm.DB) error {
			return migrateDemoRooms(database)
		},
		Rollback: func(database *gorm.DB) error {
			names := make([]string, 0, len(demoRooms))
			for _, r := range demoRooms {
				names = append(names, r.Name)
			}
			return database.Where("name IN ?", names).Delete(&navmodel.Room{}).Error
		},
	}
}

// migrateDemoRooms resolves owner IDs and creates missing demo rooms.
func migrateDemoRooms(database *gorm.DB) error {
	for _, spec := range demoRooms {
		var owner usermodel.Record
		q := database.Where("username = ?", spec.OwnerUsername).Limit(1).Find(&owner)
		if q.Error != nil {
			return q.Error
		}
		if q.RowsAffected == 0 {
			continue
		}
		existing := navmodel.Room{}
		rq := database.Where("name = ? AND owner_id = ?", spec.Name, owner.ID).Limit(1).Find(&existing)
		if rq.Error != nil {
			return rq.Error
		}
		if rq.RowsAffected > 0 {
			continue
		}
		room := navmodel.Room{
			OwnerID: owner.ID, OwnerName: owner.Username, Name: spec.Name,
			Description: spec.Description, State: spec.State, MaxUsers: spec.MaxUsers,
			Tags: spec.Tags,
		}
		if err := database.Create(&room).Error; err != nil {
			return err
		}
	}
	return nil
}
