package application

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkmoderation "github.com/momlesstomato/pixel-sdk/events/moderation"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
)

// Service implements moderation business logic.
type Service struct {
	// repo stores the action persistence layer.
	repo domain.ActionRepository
	// fire dispatches SDK events when configured.
	fire func(sdk.Event)
	// ambassadorNotifier sends alerts to ambassadors when configured.
	ambassadorNotifier AmbassadorNotifier
}

// AmbassadorNotifier defines ambassador alert dispatch behavior.
type AmbassadorNotifier interface {
	// NotifyAmbassadors broadcasts a moderation alert to all ambassadors.
	NotifyAmbassadors(ctx context.Context, message string) error
}

// NewService creates a moderation service.
func NewService(repo domain.ActionRepository) (*Service, error) {
	if repo == nil {
		return nil, domain.ErrMissingTarget
	}
	return &Service{repo: repo}, nil
}

// SetEventFirer configures the SDK event dispatcher.
func (s *Service) SetEventFirer(fn func(sdk.Event)) {
	s.fire = fn
}

// SetAmbassadorNotifier configures the ambassador alert dispatcher.
func (s *Service) SetAmbassadorNotifier(notifier AmbassadorNotifier) {
	s.ambassadorNotifier = notifier
}

// Create records a new moderation action.
func (s *Service) Create(ctx context.Context, action *domain.Action) error {
	if action.TargetUserID <= 0 {
		return domain.ErrMissingTarget
	}
	if action.Scope != domain.ScopeRoom && action.Scope != domain.ScopeHotel {
		return domain.ErrInvalidScope
	}
	if err := s.fireCreateBefore(action); err != nil {
		return err
	}
	action.Active = true
	if action.DurationMinutes > 0 {
		exp := time.Now().Add(time.Duration(action.DurationMinutes) * time.Minute)
		action.ExpiresAt = &exp
	}
	if err := s.repo.Create(ctx, action); err != nil {
		return err
	}
	s.fireCreateAfter(action)
	return nil
}

// Deactivate marks one action as inactive.
func (s *Service) Deactivate(ctx context.Context, id int64, deactivatedBy int) error {
	action, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !action.Active {
		return domain.ErrAlreadyInactive
	}
	return s.repo.Deactivate(ctx, id, deactivatedBy)
}

// Delete hard-deletes a room-scoped action.
func (s *Service) Delete(ctx context.Context, id int64) error {
	action, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if action.Scope == domain.ScopeHotel {
		return domain.ErrCannotDeleteHotelAction
	}
	return s.repo.Delete(ctx, id)
}

// FindByID retrieves one action.
func (s *Service) FindByID(ctx context.Context, id int64) (*domain.Action, error) {
	return s.repo.FindByID(ctx, id)
}

// List returns actions matching the filter.
func (s *Service) List(ctx context.Context, filter domain.ListFilter) ([]domain.Action, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}
	return s.repo.List(ctx, filter)
}

// IsHotelBanned checks if a user has an active hotel ban.
func (s *Service) IsHotelBanned(ctx context.Context, userID int) (bool, error) {
	return s.repo.HasActiveBan(ctx, userID, domain.ScopeHotel)
}

// IsHotelMuted checks if a user has an active hotel mute.
func (s *Service) IsHotelMuted(ctx context.Context, userID int) (bool, error) {
	return s.repo.HasActiveMute(ctx, userID, domain.ScopeHotel)
}

// IsIPBanned checks if an IP address has an active ban.
func (s *Service) IsIPBanned(ctx context.Context, ip string) (bool, error) {
	return s.repo.HasActiveIPBan(ctx, ip)
}

// IsTradeLocked checks if a user has an active trade lock.
func (s *Service) IsTradeLocked(ctx context.Context, userID int) (bool, error) {
	return s.repo.HasActiveTradeLock(ctx, userID)
}

// Escalate determines the next sanction level based on user history.
func (s *Service) Escalate(ctx context.Context, userID int) (*domain.Action, error) {
	filter := domain.ListFilter{TargetUserID: userID, Limit: 100}
	active := true
	filter.Active = &active
	history, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	count := len(history)
	switch {
	case count == 0:
		return &domain.Action{ActionType: domain.TypeWarn, Scope: domain.ScopeHotel}, nil
	case count <= 2:
		return &domain.Action{ActionType: domain.TypeMute, Scope: domain.ScopeHotel, DurationMinutes: 120}, nil
	case count <= 5:
		return &domain.Action{ActionType: domain.TypeBan, Scope: domain.ScopeHotel, DurationMinutes: 1440}, nil
	default:
		return &domain.Action{ActionType: domain.TypeBan, Scope: domain.ScopeHotel, DurationMinutes: 0}, nil
	}
}

// AlertAmbassadors sends a moderation activity alert to ambassadors.
func (s *Service) AlertAmbassadors(ctx context.Context, message string) {
	if s.ambassadorNotifier != nil {
		_ = s.ambassadorNotifier.NotifyAmbassadors(ctx, message)
	}
}

// fireSafe dispatches one event when the firer is configured.
func (s *Service) fireSafe(event sdk.Event) {
	if s.fire != nil {
		s.fire(event)
	}
}

// fireCreateBefore dispatches the cancellable event for one moderation action.
func (s *Service) fireCreateBefore(action *domain.Action) error {
	scope := string(action.Scope)
	switch action.ActionType {
	case domain.TypeKick:
		event := &sdkmoderation.UserKicking{TargetID: action.TargetUserID, IssuerID: action.IssuerID, RoomID: action.RoomID, Scope: scope}
		s.fireSafe(event)
		if event.Cancelled() {
			return fmt.Errorf("moderation kick cancelled by plugin")
		}
	case domain.TypeMute:
		event := &sdkmoderation.UserMuting{TargetID: action.TargetUserID, IssuerID: action.IssuerID, Scope: scope, DurationMinutes: action.DurationMinutes, Reason: action.Reason}
		s.fireSafe(event)
		if event.Cancelled() {
			return fmt.Errorf("moderation mute cancelled by plugin")
		}
	case domain.TypeBan:
		event := &sdkmoderation.UserBanning{TargetID: action.TargetUserID, IssuerID: action.IssuerID, Scope: scope, Reason: action.Reason, DurationMinutes: action.DurationMinutes}
		s.fireSafe(event)
		if event.Cancelled() {
			return fmt.Errorf("moderation ban cancelled by plugin")
		}
	case domain.TypeWarn:
		event := &sdkmoderation.UserWarning{TargetID: action.TargetUserID, IssuerID: action.IssuerID, Message: action.Reason}
		s.fireSafe(event)
		if event.Cancelled() {
			return fmt.Errorf("moderation warning cancelled by plugin")
		}
	}
	return nil
}

// fireCreateAfter dispatches the non-cancellable event for one moderation action.
func (s *Service) fireCreateAfter(action *domain.Action) {
	scope := string(action.Scope)
	switch action.ActionType {
	case domain.TypeKick:
		s.fireSafe(&sdkmoderation.UserKicked{TargetID: action.TargetUserID, IssuerID: action.IssuerID, RoomID: action.RoomID, Scope: scope})
	case domain.TypeMute:
		s.fireSafe(&sdkmoderation.UserMuted{TargetID: action.TargetUserID, IssuerID: action.IssuerID, Scope: scope, DurationMinutes: action.DurationMinutes, Reason: action.Reason})
	case domain.TypeBan:
		s.fireSafe(&sdkmoderation.UserBanned{TargetID: action.TargetUserID, IssuerID: action.IssuerID, Scope: scope, Reason: action.Reason, DurationMinutes: action.DurationMinutes})
	case domain.TypeWarn:
		s.fireSafe(&sdkmoderation.UserWarned{TargetID: action.TargetUserID, IssuerID: action.IssuerID, Message: action.Reason})
	}
}
