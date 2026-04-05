package store

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/infrastructure/model"
	"gorm.io/gorm"
)

// WordFilterStore implements domain.WordFilterRepository using GORM.
type WordFilterStore struct {
	// db stores the GORM database connection.
	db *gorm.DB
}

// NewWordFilterStore creates a word filter store backed by the given database.
func NewWordFilterStore(db *gorm.DB) (*WordFilterStore, error) {
	if db == nil {
		return nil, domain.ErrMissingTarget
	}
	return &WordFilterStore{db: db}, nil
}

// Create persists a new word filter entry.
func (s *WordFilterStore) Create(ctx context.Context, entry *domain.WordFilterEntry) error {
	m := filterToModel(entry)
	if err := s.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	entry.ID = m.ID
	entry.CreatedAt = m.CreatedAt
	return nil
}

// FindByID retrieves one word filter entry by identifier.
func (s *WordFilterStore) FindByID(ctx context.Context, id int64) (*domain.WordFilterEntry, error) {
	var m model.ModerationWordFilter
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, domain.ErrWordFilterNotFound
	}
	e := filterToDomain(&m)
	return &e, nil
}

// ListActive returns active word filters for the given scope and room.
func (s *WordFilterStore) ListActive(ctx context.Context, scope string, roomID int) ([]domain.WordFilterEntry, error) {
	q := s.db.WithContext(ctx).Model(&model.ModerationWordFilter{}).Where("active = true")
	if scope != "" {
		q = q.Where("scope = ?", scope)
	}
	if roomID > 0 {
		q = q.Where("(scope = 'global' OR room_id = ?)", roomID)
	}
	var rows []model.ModerationWordFilter
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.WordFilterEntry, len(rows))
	for i := range rows {
		out[i] = filterToDomain(&rows[i])
	}
	return out, nil
}

// Update persists changes to a word filter entry.
func (s *WordFilterStore) Update(ctx context.Context, entry *domain.WordFilterEntry) error {
	return s.db.WithContext(ctx).Model(&model.ModerationWordFilter{}).Where("id = ?", entry.ID).Updates(map[string]interface{}{
		"pattern": entry.Pattern, "replacement": entry.Replacement,
		"is_regex": entry.IsRegex, "scope": entry.Scope, "room_id": entry.RoomID, "active": entry.Active,
	}).Error
}

// Delete hard-deletes one word filter entry.
func (s *WordFilterStore) Delete(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Delete(&model.ModerationWordFilter{}, id).Error
}

func filterToModel(e *domain.WordFilterEntry) model.ModerationWordFilter {
	return model.ModerationWordFilter{
		Pattern: e.Pattern, Replacement: e.Replacement, IsRegex: e.IsRegex,
		Scope: e.Scope, RoomID: e.RoomID, Active: e.Active,
	}
}

func filterToDomain(m *model.ModerationWordFilter) domain.WordFilterEntry {
	return domain.WordFilterEntry{
		ID: m.ID, Pattern: m.Pattern, Replacement: m.Replacement, IsRegex: m.IsRegex,
		Scope: m.Scope, RoomID: m.RoomID, Active: m.Active, CreatedAt: m.CreatedAt,
	}
}
