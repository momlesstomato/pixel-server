package application

import (
	"context"
	"fmt"
	"strings"

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
