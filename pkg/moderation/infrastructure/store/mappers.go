package store

import (
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/infrastructure/model"
	"gorm.io/gorm"
)

// toModel converts a domain action to a GORM model.
func toModel(a *domain.Action) model.ModerationAction {
	return model.ModerationAction{
		Scope:           string(a.Scope),
		ActionType:      string(a.ActionType),
		TargetUserID:    a.TargetUserID,
		IssuerID:        a.IssuerID,
		RoomID:          a.RoomID,
		Reason:          a.Reason,
		DurationMinutes: a.DurationMinutes,
		ExpiresAt:       a.ExpiresAt,
		Active:          a.Active,
		IPAddress:       a.IPAddress,
		MachineID:       a.MachineID,
	}
}

// toDomain converts a GORM model to a domain action.
func toDomain(m *model.ModerationAction) domain.Action {
	return domain.Action{
		ID:              m.ID,
		Scope:           domain.ActionScope(m.Scope),
		ActionType:      domain.ActionType(m.ActionType),
		TargetUserID:    m.TargetUserID,
		IssuerID:        m.IssuerID,
		RoomID:          m.RoomID,
		Reason:          m.Reason,
		DurationMinutes: m.DurationMinutes,
		ExpiresAt:       m.ExpiresAt,
		Active:          m.Active,
		DeactivatedBy:   m.DeactivatedBy,
		DeactivatedAt:   m.DeactivatedAt,
		IPAddress:       m.IPAddress,
		MachineID:       m.MachineID,
		CreatedAt:       m.CreatedAt,
	}
}

// applyFilter adds where clauses for a list filter.
func applyFilter(q *gorm.DB, f domain.ListFilter) *gorm.DB {
	if f.Scope != "" {
		q = q.Where("scope = ?", string(f.Scope))
	}
	if f.ActionType != "" {
		q = q.Where("action_type = ?", string(f.ActionType))
	}
	if f.IssuerID > 0 {
		q = q.Where("issuer_id = ?", f.IssuerID)
	}
	if f.TargetUserID > 0 {
		q = q.Where("target_user_id = ?", f.TargetUserID)
	}
	if f.RoomID > 0 {
		q = q.Where("room_id = ?", f.RoomID)
	}
	if f.Active != nil {
		q = q.Where("active = ?", *f.Active)
	}
	return q
}
