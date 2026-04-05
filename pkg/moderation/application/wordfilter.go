package application

import (
	"context"
	"regexp"
	"strings"

	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
)

// WordFilterService implements word filter business logic.
type WordFilterService struct {
	// repo stores the word filter persistence layer.
	repo domain.WordFilterRepository
}

// NewWordFilterService creates a word filter service.
func NewWordFilterService(repo domain.WordFilterRepository) (*WordFilterService, error) {
	if repo == nil {
		return nil, domain.ErrMissingTarget
	}
	return &WordFilterService{repo: repo}, nil
}

// Create stores a new word filter rule.
func (s *WordFilterService) Create(ctx context.Context, entry *domain.WordFilterEntry) error {
	if entry.Pattern == "" {
		return domain.ErrMissingTarget
	}
	entry.Active = true
	return s.repo.Create(ctx, entry)
}

// FindByID retrieves one word filter entry.
func (s *WordFilterService) FindByID(ctx context.Context, id int64) (*domain.WordFilterEntry, error) {
	return s.repo.FindByID(ctx, id)
}

// ListActive returns active filters for the given scope.
func (s *WordFilterService) ListActive(ctx context.Context, scope string, roomID int) ([]domain.WordFilterEntry, error) {
	return s.repo.ListActive(ctx, scope, roomID)
}

// Update persists changes to a word filter entry.
func (s *WordFilterService) Update(ctx context.Context, entry *domain.WordFilterEntry) error {
	return s.repo.Update(ctx, entry)
}

// Delete removes a word filter entry.
func (s *WordFilterService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// FilterMessage applies all active filters to a message and returns the result.
func (s *WordFilterService) FilterMessage(ctx context.Context, roomID int, message string) (string, bool) {
	filters, err := s.repo.ListActive(ctx, "", roomID)
	if err != nil || len(filters) == 0 {
		return message, false
	}
	modified := false
	result := message
	for _, f := range filters {
		if f.IsRegex {
			re, err := regexp.Compile("(?i)" + f.Pattern)
			if err != nil {
				continue
			}
			replaced := re.ReplaceAllString(result, f.Replacement)
			if replaced != result {
				result = replaced
				modified = true
			}
			continue
		}
		lower := strings.ToLower(result)
		if strings.Contains(lower, strings.ToLower(f.Pattern)) {
			result = strings.NewReplacer(f.Pattern, f.Replacement).Replace(result)
			modified = true
		}
	}
	return result, modified
}
