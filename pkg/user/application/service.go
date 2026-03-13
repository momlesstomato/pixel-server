package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// Service defines user application use-cases.
type Service struct {
	// repository stores user persistence contract implementation.
	repository domain.Repository
}

// NewService creates one user service.
func NewService(repository domain.Repository) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf("user repository is required")
	}
	return &Service{repository: repository}, nil
}

// Create creates one user using username validation.
func (service *Service) Create(ctx context.Context, username string) (domain.User, error) {
	trimmed := strings.TrimSpace(username)
	if trimmed == "" {
		return domain.User{}, fmt.Errorf("username is required")
	}
	if len(trimmed) > 64 {
		return domain.User{}, fmt.Errorf("username must be <= 64 characters")
	}
	return service.repository.Create(ctx, trimmed)
}

// FindByID resolves one user by identifier.
func (service *Service) FindByID(ctx context.Context, id int) (domain.User, error) {
	if id <= 0 {
		return domain.User{}, fmt.Errorf("user id must be positive")
	}
	return service.repository.FindByID(ctx, id)
}

// DeleteByID soft-deletes one user by identifier.
func (service *Service) DeleteByID(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("user id must be positive")
	}
	return service.repository.DeleteByID(ctx, id)
}

// RecordLogin stamps one successful login event and reports first-login-of-day state.
func (service *Service) RecordLogin(ctx context.Context, userID int, holder string, loggedAt time.Time) (bool, error) {
	if userID <= 0 {
		return false, fmt.Errorf("user id must be positive")
	}
	if strings.TrimSpace(holder) == "" {
		return false, fmt.Errorf("holder is required")
	}
	if loggedAt.IsZero() {
		return false, fmt.Errorf("logged at timestamp is required")
	}
	return service.repository.RecordLogin(ctx, userID, strings.TrimSpace(holder), loggedAt.UTC())
}
