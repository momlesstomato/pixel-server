package store

import (
	"context"
	"strings"
	"time"

	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
)

// RecordLogin stores one successful login event and returns first-login-of-day status.
func (repository *Repository) RecordLogin(ctx context.Context, userID int, holder string, loggedAt time.Time) (bool, error) {
	dayStart := utcDayStart(loggedAt)
	dayEnd := dayStart.Add(24 * time.Hour)
	var count int64
	query := repository.database.WithContext(ctx).Model(&usermodel.LoginEvent{}).Where("user_id = ? AND logged_at >= ? AND logged_at < ?", userID, dayStart, dayEnd)
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	event := usermodel.LoginEvent{UserID: userID, Holder: strings.TrimSpace(holder), LoggedAt: loggedAt.UTC()}
	if err := repository.database.WithContext(ctx).Create(&event).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}
