package store

import (
	"context"
	"errors"
	"strings"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"gorm.io/gorm"
)

// roomRecord represents the full rooms row including room-realm extension columns.
type roomRecord struct {
	// ID stores the stable room identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// OwnerID stores the room creator identifier.
	OwnerID uint `gorm:"column:owner_id"`
	// OwnerName stores the room creator display name.
	OwnerName string `gorm:"column:owner_name"`
	// Name stores the room display name.
	Name string `gorm:"column:name"`
	// Description stores the room description.
	Description string `gorm:"column:description"`
	// State stores the access state string (open/locked/password).
	State string `gorm:"column:state"`
	// CategoryID stores the navigator category reference.
	CategoryID uint `gorm:"column:category_id"`
	// MaxUsers stores the room capacity.
	MaxUsers int `gorm:"column:max_users"`
	// Score stores the room star rating.
	Score int `gorm:"column:score"`
	// Tags stores comma-separated room tags.
	Tags string `gorm:"column:tags"`
	// TradeMode stores the trade policy code.
	TradeMode int `gorm:"column:trade_mode"`
	// ModelSlug stores the room model template identifier.
	ModelSlug string `gorm:"column:model_slug"`
	// PasswordHash stores bcrypt hash for password-protected rooms.
	PasswordHash string `gorm:"column:password_hash"`
	// WallHeight stores custom wall height (-1 = auto).
	WallHeight int `gorm:"column:wall_height"`
	// FloorThickness stores floor rendering thickness.
	FloorThickness int `gorm:"column:floor_thickness"`
	// WallThickness stores wall rendering thickness.
	WallThickness int `gorm:"column:wall_thickness"`
	// AllowPets stores whether pet placement is allowed.
	AllowPets bool `gorm:"column:allow_pets"`
	// AllowTrading stores whether trading is enabled.
	AllowTrading bool `gorm:"column:allow_trading"`
}

// TableName returns the PostgreSQL table name for roomRecord.
func (roomRecord) TableName() string { return "rooms" }

// RoomStore implements domain.RoomRepository using PostgreSQL.
type RoomStore struct {
	// db stores the database connection.
	db *gorm.DB
}

// NewRoomStore creates one room store instance.
func NewRoomStore(db *gorm.DB) (*RoomStore, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}
	return &RoomStore{db: db}, nil
}

// FindByID resolves one full room record by stable identifier.
func (s *RoomStore) FindByID(ctx context.Context, roomID int) (domain.Room, error) {
	var rec roomRecord
	if err := s.db.WithContext(ctx).First(&rec, uint(roomID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Room{}, domain.ErrRoomNotFound
		}
		return domain.Room{}, err
	}
	return recordToDomain(rec), nil
}

// SaveSettings persists updated room settings for one room.
func (s *RoomStore) SaveSettings(ctx context.Context, room domain.Room) error {
	updates := map[string]interface{}{
		"name": room.Name, "description": room.Description,
		"state": string(room.State), "max_users": room.MaxUsers,
		"password_hash": room.Password, "wall_height": room.WallHeight,
		"floor_thickness": room.FloorThickness, "wall_thickness": room.WallThickness,
		"allow_pets": room.AllowPets, "allow_trading": room.AllowTrading,
		"trade_mode": room.TradeMode,
	}
	return s.db.WithContext(ctx).Table("rooms").Where("id = ?", room.ID).Updates(updates).Error
}

// recordToDomain converts a database record to the domain Room aggregate.
func recordToDomain(rec roomRecord) domain.Room {
	var tags []string
	if rec.Tags != "" {
		tags = strings.Split(rec.Tags, ",")
	}
	return domain.Room{
		ID: int(rec.ID), OwnerID: int(rec.OwnerID), OwnerName: rec.OwnerName,
		Name: rec.Name, Description: rec.Description,
		State:          domain.AccessState(rec.State),
		ModelSlug:      rec.ModelSlug,
		CategoryID:     int(rec.CategoryID),
		MaxUsers:       rec.MaxUsers,
		Password:       rec.PasswordHash,
		Score:          rec.Score,
		Tags:           tags,
		TradeMode:      rec.TradeMode,
		WallHeight:     rec.WallHeight,
		FloorThickness: rec.FloorThickness,
		WallThickness:  rec.WallThickness,
		AllowPets:      rec.AllowPets,
		AllowTrading:   rec.AllowTrading,
	}
}
