package store

import (
	"context"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
	messengermodel "github.com/momlesstomato/pixel-server/pkg/messenger/infrastructure/model"
)

// SaveOfflineMessage persists one offline message row.
func (store *Store) SaveOfflineMessage(ctx context.Context, fromUserID, toUserID int, message string) error {
	row := messengermodel.OfflineMessage{FromUserID: fromUserID, ToUserID: toUserID, Message: message, SentAt: time.Now().UTC()}
	return store.database.WithContext(ctx).Create(&row).Error
}

// GetAndDeleteOfflineMessages returns and atomically removes offline messages for one user.
func (store *Store) GetAndDeleteOfflineMessages(ctx context.Context, userID int) ([]domain.OfflineMessage, error) {
	var rows []messengermodel.OfflineMessage
	if err := store.database.WithContext(ctx).Raw(
		"DELETE FROM offline_messages WHERE to_user_id = ? RETURNING id, from_user_id, to_user_id, message, sent_at",
		userID,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.OfflineMessage, len(rows))
	for i, row := range rows {
		result[i] = domain.OfflineMessage{
			ID: row.ID, FromUserID: row.FromUserID, ToUserID: row.ToUserID,
			Message: row.Message, SentAt: row.SentAt,
		}
	}
	return result, nil
}

// DeleteOfflineMessagesOlderThan removes messages whose sent_at precedes the cutoff epoch.
func (store *Store) DeleteOfflineMessagesOlderThan(ctx context.Context, cutoffUnix int64) error {
	cutoff := time.Unix(cutoffUnix, 0).UTC()
	return store.database.WithContext(ctx).Where("sent_at < ?", cutoff).Delete(&messengermodel.OfflineMessage{}).Error
}
