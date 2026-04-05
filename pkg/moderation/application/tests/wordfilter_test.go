package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/moderation/application"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWordFilterRepo struct {
	entries []domain.WordFilterEntry
	nextID  int64
}

func (m *mockWordFilterRepo) Create(_ context.Context, e *domain.WordFilterEntry) error {
	m.nextID++
	e.ID = m.nextID
	m.entries = append(m.entries, *e)
	return nil
}

func (m *mockWordFilterRepo) FindByID(_ context.Context, id int64) (*domain.WordFilterEntry, error) {
	for i := range m.entries {
		if m.entries[i].ID == id {
			return &m.entries[i], nil
		}
	}
	return nil, domain.ErrWordFilterNotFound
}

func (m *mockWordFilterRepo) ListActive(_ context.Context, _ string, _ int) ([]domain.WordFilterEntry, error) {
	var out []domain.WordFilterEntry
	for _, e := range m.entries {
		if e.Active {
			out = append(out, e)
		}
	}
	return out, nil
}

func (m *mockWordFilterRepo) Update(_ context.Context, e *domain.WordFilterEntry) error {
	for i := range m.entries {
		if m.entries[i].ID == e.ID {
			m.entries[i] = *e
			return nil
		}
	}
	return domain.ErrWordFilterNotFound
}

func (m *mockWordFilterRepo) Delete(_ context.Context, id int64) error {
	for i := range m.entries {
		if m.entries[i].ID == id {
			m.entries = append(m.entries[:i], m.entries[i+1:]...)
			return nil
		}
	}
	return nil
}

// TestNewWordFilterServiceNilRepo verifies nil repo is rejected.
func TestNewWordFilterServiceNilRepo(t *testing.T) {
	_, err := application.NewWordFilterService(nil)
	assert.ErrorIs(t, err, domain.ErrMissingTarget)
}

// TestWordFilterCreateEmptyPattern verifies empty pattern is rejected.
func TestWordFilterCreateEmptyPattern(t *testing.T) {
	svc, _ := application.NewWordFilterService(&mockWordFilterRepo{})
	err := svc.Create(context.Background(), &domain.WordFilterEntry{})
	assert.ErrorIs(t, err, domain.ErrMissingTarget)
}

// TestWordFilterCreateSuccess verifies successful filter creation.
func TestWordFilterCreateSuccess(t *testing.T) {
	svc, _ := application.NewWordFilterService(&mockWordFilterRepo{})
	entry := &domain.WordFilterEntry{Pattern: "badword", Replacement: "***"}
	require.NoError(t, svc.Create(context.Background(), entry))
	assert.True(t, entry.Active)
	assert.Greater(t, entry.ID, int64(0))
}

// TestWordFilterFilterMessagePlain verifies plain text replacement.
func TestWordFilterFilterMessagePlain(t *testing.T) {
	repo := &mockWordFilterRepo{}
	svc, _ := application.NewWordFilterService(repo)
	_ = svc.Create(context.Background(), &domain.WordFilterEntry{Pattern: "bad", Replacement: "***"})
	result, modified := svc.FilterMessage(context.Background(), 0, "this is bad stuff")
	assert.True(t, modified)
	assert.Contains(t, result, "***")
}

// TestWordFilterFilterMessageRegex verifies regex pattern replacement.
func TestWordFilterFilterMessageRegex(t *testing.T) {
	repo := &mockWordFilterRepo{}
	svc, _ := application.NewWordFilterService(repo)
	_ = svc.Create(context.Background(), &domain.WordFilterEntry{Pattern: "b[a4]d", Replacement: "***", IsRegex: true})
	result, modified := svc.FilterMessage(context.Background(), 0, "this is b4d stuff")
	assert.True(t, modified)
	assert.Contains(t, result, "***")
}

// TestWordFilterFilterMessageNoMatch verifies unmodified message on no match.
func TestWordFilterFilterMessageNoMatch(t *testing.T) {
	repo := &mockWordFilterRepo{}
	svc, _ := application.NewWordFilterService(repo)
	_ = svc.Create(context.Background(), &domain.WordFilterEntry{Pattern: "badword", Replacement: "***"})
	result, modified := svc.FilterMessage(context.Background(), 0, "clean message")
	assert.False(t, modified)
	assert.Equal(t, "clean message", result)
}

// TestWordFilterDeleteSuccess verifies delete works.
func TestWordFilterDeleteSuccess(t *testing.T) {
	repo := &mockWordFilterRepo{}
	svc, _ := application.NewWordFilterService(repo)
	entry := &domain.WordFilterEntry{Pattern: "test", Replacement: "*"}
	_ = svc.Create(context.Background(), entry)
	assert.NoError(t, svc.Delete(context.Background(), entry.ID))
}
