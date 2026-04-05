package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/moderation/application"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPresetRepo struct {
	presets map[int64]*domain.Preset
	nextID  int64
}

func newMockPresetRepo() *mockPresetRepo {
	return &mockPresetRepo{presets: make(map[int64]*domain.Preset), nextID: 1}
}

func (m *mockPresetRepo) Create(_ context.Context, p *domain.Preset) error {
	p.ID = m.nextID
	m.nextID++
	cp := *p
	m.presets[p.ID] = &cp
	return nil
}

func (m *mockPresetRepo) FindByID(_ context.Context, id int64) (*domain.Preset, error) {
	p, ok := m.presets[id]
	if !ok {
		return nil, domain.ErrPresetNotFound
	}
	return p, nil
}

func (m *mockPresetRepo) ListActive(_ context.Context) ([]domain.Preset, error) {
	var out []domain.Preset
	for _, p := range m.presets {
		if p.Active {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (m *mockPresetRepo) Update(_ context.Context, p *domain.Preset) error {
	_, ok := m.presets[p.ID]
	if !ok {
		return domain.ErrPresetNotFound
	}
	cp := *p
	m.presets[p.ID] = &cp
	return nil
}

func (m *mockPresetRepo) Delete(_ context.Context, id int64) error {
	delete(m.presets, id)
	return nil
}

// TestNewPresetServiceNilRepo verifies nil repo is rejected.
func TestNewPresetServiceNilRepo(t *testing.T) {
	_, err := application.NewPresetService(nil)
	assert.ErrorIs(t, err, domain.ErrMissingTarget)
}

// TestPresetCreateMissingFields verifies empty name/category is rejected.
func TestPresetCreateMissingFields(t *testing.T) {
	svc, _ := application.NewPresetService(newMockPresetRepo())
	err := svc.Create(context.Background(), &domain.Preset{})
	assert.ErrorIs(t, err, domain.ErrMissingTarget)
}

// TestPresetCreateSuccess verifies successful preset creation.
func TestPresetCreateSuccess(t *testing.T) {
	svc, _ := application.NewPresetService(newMockPresetRepo())
	p := &domain.Preset{Name: "test", Category: "general", ActionType: "warn"}
	require.NoError(t, svc.Create(context.Background(), p))
	assert.True(t, p.Active)
	assert.Greater(t, p.ID, int64(0))
}

// TestPresetListActiveFilters verifies only active presets are listed.
func TestPresetListActiveFilters(t *testing.T) {
	repo := newMockPresetRepo()
	svc, _ := application.NewPresetService(repo)
	_ = svc.Create(context.Background(), &domain.Preset{Name: "a", Category: "c"})
	active, err := svc.ListActive(context.Background())
	require.NoError(t, err)
	assert.Len(t, active, 1)
}

// TestPresetDeleteSuccess verifies deletion works.
func TestPresetDeleteSuccess(t *testing.T) {
	repo := newMockPresetRepo()
	svc, _ := application.NewPresetService(repo)
	p := &domain.Preset{Name: "a", Category: "c"}
	_ = svc.Create(context.Background(), p)
	assert.NoError(t, svc.Delete(context.Background(), p.ID))
}
