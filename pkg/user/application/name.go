package application

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	sdkuser "github.com/momlesstomato/pixel-sdk/events/user"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]{3,24}$`)

// CheckName validates one username and checks availability.
func (service *Service) CheckName(ctx context.Context, name string, excludeUserID int) (domain.NameResult, error) {
	clean, err := validateName(name)
	if err != nil {
		return domain.NameResult{ResultCode: domain.NameResultInvalid, Name: strings.TrimSpace(name), Suggestions: suggestions(name)}, nil
	}
	available, err := service.repository.IsUsernameAvailable(ctx, clean, excludeUserID)
	if err != nil {
		return domain.NameResult{}, err
	}
	if !available {
		return domain.NameResult{ResultCode: domain.NameResultTaken, Name: clean, Suggestions: suggestions(clean)}, nil
	}
	return domain.NameResult{ResultCode: domain.NameResultAvailable, Name: clean}, nil
}

// ChangeName validates and applies one user rename operation.
func (service *Service) ChangeName(ctx context.Context, connID string, userID int, name string, force bool) (domain.NameResult, error) {
	result, err := service.CheckName(ctx, name, userID)
	if err != nil {
		return domain.NameResult{}, err
	}
	if result.ResultCode != domain.NameResultAvailable {
		return result, nil
	}
	current, err := service.FindByID(ctx, userID)
	if err != nil {
		return domain.NameResult{}, err
	}
	if service.fire != nil {
		event := &sdkuser.NameChanged{ConnID: connID, UserID: userID, OldName: current.Username, NewName: result.Name}
		service.fire(event)
		if event.Cancelled() {
			return domain.NameResult{ResultCode: domain.NameResultNotAllowed, Name: result.Name}, nil
		}
	}
	if _, err := service.repository.ChangeUsername(ctx, userID, result.Name, force); err != nil {
		if err == domain.ErrNameChangeNotAllowed {
			return domain.NameResult{ResultCode: domain.NameResultNotAllowed, Name: result.Name}, nil
		}
		if err == domain.ErrNameAlreadyTaken {
			return domain.NameResult{ResultCode: domain.NameResultTaken, Name: result.Name, Suggestions: suggestions(result.Name)}, nil
		}
		return domain.NameResult{}, err
	}
	return domain.NameResult{ResultCode: domain.NameResultAvailable, Name: result.Name}, nil
}

// ForceChangeName applies one administrative name change operation.
func (service *Service) ForceChangeName(ctx context.Context, userID int, name string) (domain.NameResult, error) {
	return service.ChangeName(ctx, "", userID, name, true)
}

// validateName validates one username and returns a normalized value.
func validateName(name string) (string, error) {
	clean := strings.TrimSpace(name)
	if !usernamePattern.MatchString(clean) {
		return "", domain.ErrInvalidName
	}
	return clean, nil
}

// suggestions builds deterministic fallback username suggestions.
func suggestions(base string) []string {
	clean := strings.TrimSpace(base)
	if clean == "" {
		clean = "user"
	}
	clean = strings.ToLower(clean)
	return []string{fmt.Sprintf("%s1", clean), fmt.Sprintf("%s_%d", clean, 2)}
}
