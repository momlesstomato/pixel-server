package hotelstatus

import (
	"context"
	"fmt"
	"time"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	corestatus "github.com/momlesstomato/pixel-server/core/status"
	statusdomain "github.com/momlesstomato/pixel-server/pkg/status/domain"
)

// Store defines hotel status persistence and compare-and-swap behavior.
type Store interface {
	// Load retrieves persisted hotel status and reports whether a record exists.
	Load(context.Context) (statusdomain.HotelStatus, bool, error)
	// Save persists one hotel status snapshot.
	Save(context.Context, statusdomain.HotelStatus) error
	// CompareAndSwap updates stored status when expected snapshot matches current value.
	CompareAndSwap(context.Context, statusdomain.HotelStatus, statusdomain.HotelStatus) (bool, error)
}

// Service defines hotel status lifecycle behavior.
type Service struct {
	// store persists current hotel status snapshot.
	store Store
	// broadcaster publishes hotel lifecycle packets to all instances.
	broadcaster broadcast.Broadcaster
	// config stores hotel scheduling and broadcast settings.
	config corestatus.Config
	// now returns current time for deterministic tests.
	now func() time.Time
}

// NewService creates one hotel status service.
func NewService(store Store, broadcaster broadcast.Broadcaster, config corestatus.Config) (*Service, error) {
	if store == nil {
		return nil, fmt.Errorf("status store is required")
	}
	if broadcaster == nil {
		return nil, fmt.Errorf("broadcaster is required")
	}
	return &Service{store: store, broadcaster: broadcaster, config: config, now: time.Now}, nil
}

// Current returns active hotel status and initializes from schedule when missing.
func (service *Service) Current(ctx context.Context) (statusdomain.HotelStatus, error) {
	status, found, err := service.store.Load(ctx)
	if err != nil {
		return statusdomain.HotelStatus{}, err
	}
	if found {
		return status, nil
	}
	initialized := service.statusFromSchedule(service.now().UTC())
	if err := service.store.Save(ctx, initialized); err != nil {
		return statusdomain.HotelStatus{}, err
	}
	return initialized, nil
}

// ScheduleClose transitions hotel state into closing mode and publishes countdown packets.
func (service *Service) ScheduleClose(ctx context.Context, minutesUntilClose int32, durationMinutes int32, throwUsers bool) (statusdomain.HotelStatus, error) {
	if minutesUntilClose < 0 {
		minutesUntilClose = 0
	}
	duration := durationMinutes
	if duration <= 0 {
		duration = int32(service.config.DefaultMaintenanceDurationMinutes)
	}
	current, err := service.Current(ctx)
	if err != nil {
		return statusdomain.HotelStatus{}, err
	}
	closeAt := service.now().UTC().Add(time.Duration(minutesUntilClose) * time.Minute)
	reopenAt := closeAt.Add(time.Duration(duration) * time.Minute)
	next := statusdomain.HotelStatus{State: statusdomain.StateClosing, CloseAt: &closeAt, ReopenAt: &reopenAt, UserThrownOutAtClose: throwUsers}
	if err := service.swap(ctx, current, next); err != nil {
		return statusdomain.HotelStatus{}, err
	}
	service.publishClosingPackets(ctx, next, minutesUntilClose, duration)
	return next, nil
}

// Reopen transitions hotel state into open mode and persists it.
func (service *Service) Reopen(ctx context.Context) (statusdomain.HotelStatus, error) {
	current, err := service.Current(ctx)
	if err != nil {
		return statusdomain.HotelStatus{}, err
	}
	next := statusdomain.HotelStatus{State: statusdomain.StateOpen}
	if err := service.swap(ctx, current, next); err != nil {
		return statusdomain.HotelStatus{}, err
	}
	return next, nil
}

// swap applies one optimistic transition using compare-and-swap semantics.
func (service *Service) swap(ctx context.Context, expected statusdomain.HotelStatus, next statusdomain.HotelStatus) error {
	swapped, err := service.store.CompareAndSwap(ctx, expected, next)
	if err != nil {
		return err
	}
	if !swapped {
		return fmt.Errorf("hotel status transition conflict")
	}
	return nil
}

// statusFromSchedule computes current hotel status from configured open and close schedule.
func (service *Service) statusFromSchedule(now time.Time) statusdomain.HotelStatus {
	openAt := time.Date(now.Year(), now.Month(), now.Day(), service.config.OpenHour, service.config.OpenMinute, 0, 0, time.UTC)
	closeAt := time.Date(now.Year(), now.Month(), now.Day(), service.config.CloseHour, service.config.CloseMinute, 0, 0, time.UTC)
	if now.Before(openAt) {
		return statusdomain.HotelStatus{State: statusdomain.StateClosed, ReopenAt: &openAt}
	}
	if now.Before(closeAt) {
		return statusdomain.HotelStatus{State: statusdomain.StateOpen}
	}
	nextOpen := openAt.Add(24 * time.Hour)
	return statusdomain.HotelStatus{State: statusdomain.StateClosed, ReopenAt: &nextOpen}
}
