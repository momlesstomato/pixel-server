package application

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
)

// PresetService implements moderation preset business logic.
type PresetService struct {
	// repo stores the preset persistence layer.
	repo domain.PresetRepository
}

// NewPresetService creates a preset service.
func NewPresetService(repo domain.PresetRepository) (*PresetService, error) {
	if repo == nil {
		return nil, domain.ErrMissingTarget
	}
	return &PresetService{repo: repo}, nil
}

// Create stores a new moderation preset.
func (s *PresetService) Create(ctx context.Context, preset *domain.Preset) error {
	if preset.Name == "" || preset.Category == "" {
		return domain.ErrMissingTarget
	}
	preset.Active = true
	return s.repo.Create(ctx, preset)
}

// FindByID retrieves one preset.
func (s *PresetService) FindByID(ctx context.Context, id int64) (*domain.Preset, error) {
	return s.repo.FindByID(ctx, id)
}

// ListActive returns all active presets.
func (s *PresetService) ListActive(ctx context.Context) ([]domain.Preset, error) {
	return s.repo.ListActive(ctx)
}

// Update persists changes to a preset.
func (s *PresetService) Update(ctx context.Context, preset *domain.Preset) error {
	return s.repo.Update(ctx, preset)
}

// Delete removes a preset.
func (s *PresetService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
