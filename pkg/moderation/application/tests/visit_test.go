package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/moderation/application"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockVisitRepo struct {
	records []domain.VisitRecord
	nextID  int64
}

func (m *mockVisitRepo) Record(_ context.Context, r *domain.VisitRecord) error {
	m.nextID++
	r.ID = m.nextID
	m.records = append(m.records, *r)
	return nil
}

func (m *mockVisitRepo) ListByUser(_ context.Context, uid int, limit int) ([]domain.VisitRecord, error) {
	var out []domain.VisitRecord
	for _, r := range m.records {
		if r.UserID == uid {
			out = append(out, r)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (m *mockVisitRepo) ListByRoom(_ context.Context, rid int, limit int) ([]domain.VisitRecord, error) {
	var out []domain.VisitRecord
	for _, r := range m.records {
		if r.RoomID == rid {
			out = append(out, r)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

// TestNewVisitServiceNilRepo verifies nil repo is rejected.
func TestNewVisitServiceNilRepo(t *testing.T) {
	_, err := application.NewVisitService(nil)
	assert.ErrorIs(t, err, domain.ErrMissingTarget)
}

// TestVisitRecordSuccess verifies successful visit recording.
func TestVisitRecordSuccess(t *testing.T) {
	svc, _ := application.NewVisitService(&mockVisitRepo{})
	require.NoError(t, svc.RecordVisit(context.Background(), 1, 100))
}

// TestVisitListByUser verifies user-scoped visit listing.
func TestVisitListByUser(t *testing.T) {
	repo := &mockVisitRepo{}
	svc, _ := application.NewVisitService(repo)
	_ = svc.RecordVisit(context.Background(), 1, 100)
	_ = svc.RecordVisit(context.Background(), 2, 200)
	_ = svc.RecordVisit(context.Background(), 1, 300)
	visits, err := svc.ListByUser(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Len(t, visits, 2)
}

// TestVisitListByRoom verifies room-scoped visit listing.
func TestVisitListByRoom(t *testing.T) {
	repo := &mockVisitRepo{}
	svc, _ := application.NewVisitService(repo)
	_ = svc.RecordVisit(context.Background(), 1, 100)
	_ = svc.RecordVisit(context.Background(), 2, 100)
	visits, err := svc.ListByRoom(context.Background(), 100, 10)
	require.NoError(t, err)
	assert.Len(t, visits, 2)
}

// TestVisitListLimitClamp verifies invalid limit is clamped.
func TestVisitListLimitClamp(t *testing.T) {
	svc, _ := application.NewVisitService(&mockVisitRepo{})
	visits, err := svc.ListByUser(context.Background(), 1, 0)
	require.NoError(t, err)
	assert.Empty(t, visits)
}
