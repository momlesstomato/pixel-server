package store

import (
	"context"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/infrastructure/model"
	"gorm.io/gorm"
)

// ActionStore implements domain.ActionRepository using GORM.
type ActionStore struct {
	// db stores the GORM database connection.
	db *gorm.DB
}

// NewActionStore creates an action store backed by the given database.
func NewActionStore(db *gorm.DB) (*ActionStore, error) {
	if db == nil {
		return nil, domain.ErrMissingTarget
	}
	return &ActionStore{db: db}, nil
}

// Create persists a new moderation action.
func (s *ActionStore) Create(ctx context.Context, action *domain.Action) error {
	m := toModel(action)
	if err := s.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	action.ID = m.ID
	action.CreatedAt = m.CreatedAt
	return nil
}

// FindByID retrieves one action by identifier.
func (s *ActionStore) FindByID(ctx context.Context, id int64) (*domain.Action, error) {
	var m model.ModerationAction
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, domain.ErrActionNotFound
	}
	a := toDomain(&m)
	return &a, nil
}

// List returns actions matching the given filter.
func (s *ActionStore) List(ctx context.Context, f domain.ListFilter) ([]domain.Action, error) {
	q := s.db.WithContext(ctx).Model(&model.ModerationAction{})
	q = applyFilter(q, f)
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}
	q = q.Order("created_at DESC")
	var rows []model.ModerationAction
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Action, len(rows))
	for i := range rows {
		out[i] = toDomain(&rows[i])
	}
	return out, nil
}

// Deactivate marks one action as inactive.
func (s *ActionStore) Deactivate(ctx context.Context, id int64, deactivatedBy int) error {
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&model.ModerationAction{}).
		Where("id = ? AND active = true", id).
		Updates(map[string]interface{}{
			"active":         false,
			"deactivated_by": deactivatedBy,
			"deactivated_at": now,
		})
	if result.RowsAffected == 0 {
		return domain.ErrAlreadyInactive
	}
	return result.Error
}

// Delete hard-deletes one action row.
func (s *ActionStore) Delete(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Delete(&model.ModerationAction{}, id).Error
}

// HasActiveBan checks for an active ban on a user by scope.
func (s *ActionStore) HasActiveBan(ctx context.Context, userID int, scope domain.ActionScope) (bool, error) {
	return s.hasActive(ctx, userID, scope, domain.TypeBan)
}

// HasActiveMute checks for an active mute on a user by scope.
func (s *ActionStore) HasActiveMute(ctx context.Context, userID int, scope domain.ActionScope) (bool, error) {
	return s.hasActive(ctx, userID, scope, domain.TypeMute)
}

// HasActiveTradeLock checks for an active trade lock on a user.
func (s *ActionStore) HasActiveTradeLock(ctx context.Context, userID int) (bool, error) {
	return s.hasActive(ctx, userID, domain.ScopeHotel, domain.TypeTradeLock)
}

// HasActiveIPBan checks for an active ban on an IP address.
func (s *ActionStore) HasActiveIPBan(ctx context.Context, ip string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.ModerationAction{}).
		Where("ip_address = ? AND active = true AND action_type = ? AND (expires_at IS NULL OR expires_at > ?)",
			ip, string(domain.TypeBan), time.Now()).
		Count(&count).Error
	return count > 0, err
}

func (s *ActionStore) hasActive(ctx context.Context, userID int, scope domain.ActionScope, actionType domain.ActionType) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.ModerationAction{}).
		Where("target_user_id = ? AND scope = ? AND action_type = ? AND active = true AND (expires_at IS NULL OR expires_at > ?)",
			userID, string(scope), string(actionType), time.Now()).
		Count(&count).Error
	return count > 0, err
}
