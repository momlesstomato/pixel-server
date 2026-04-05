package store

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/infrastructure/model"
	"gorm.io/gorm"
)

// PresetStore implements domain.PresetRepository using GORM.
type PresetStore struct {
	// db stores the GORM database connection.
	db *gorm.DB
}

// NewPresetStore creates a preset store backed by the given database.
func NewPresetStore(db *gorm.DB) (*PresetStore, error) {
	if db == nil {
		return nil, domain.ErrMissingTarget
	}
	return &PresetStore{db: db}, nil
}

// Create persists a new moderation preset.
func (s *PresetStore) Create(ctx context.Context, preset *domain.Preset) error {
	m := presetToModel(preset)
	if err := s.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	preset.ID = m.ID
	preset.CreatedAt = m.CreatedAt
	return nil
}

// FindByID retrieves one preset by identifier.
func (s *PresetStore) FindByID(ctx context.Context, id int64) (*domain.Preset, error) {
	var m model.ModerationPreset
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, domain.ErrPresetNotFound
	}
	p := presetToDomain(&m)
	return &p, nil
}

// ListActive returns all active moderation presets.
func (s *PresetStore) ListActive(ctx context.Context) ([]domain.Preset, error) {
	var rows []model.ModerationPreset
	if err := s.db.WithContext(ctx).Where("active = true").Order("category, name").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Preset, len(rows))
	for i := range rows {
		out[i] = presetToDomain(&rows[i])
	}
	return out, nil
}

// Update persists changes to a moderation preset.
func (s *PresetStore) Update(ctx context.Context, preset *domain.Preset) error {
	return s.db.WithContext(ctx).Model(&model.ModerationPreset{}).Where("id = ?", preset.ID).Updates(map[string]interface{}{
		"category": preset.Category, "name": preset.Name, "action_type": string(preset.ActionType),
		"default_duration": preset.DefaultDuration, "default_reason": preset.DefaultReason, "active": preset.Active,
	}).Error
}

// Delete hard-deletes one preset.
func (s *PresetStore) Delete(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Delete(&model.ModerationPreset{}, id).Error
}

func presetToModel(p *domain.Preset) model.ModerationPreset {
	return model.ModerationPreset{
		Category: p.Category, Name: p.Name, ActionType: string(p.ActionType),
		DefaultDuration: p.DefaultDuration, DefaultReason: p.DefaultReason, Active: p.Active,
	}
}

func presetToDomain(m *model.ModerationPreset) domain.Preset {
	return domain.Preset{
		ID: m.ID, Category: m.Category, Name: m.Name, ActionType: domain.ActionType(m.ActionType),
		DefaultDuration: m.DefaultDuration, DefaultReason: m.DefaultReason, Active: m.Active, CreatedAt: m.CreatedAt,
	}
}
