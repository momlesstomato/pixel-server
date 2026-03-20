package store

import (
	"context"
	"time"

	messengermodel "github.com/momlesstomato/pixel-server/pkg/messenger/infrastructure/model"
)

// LogMessage persists one message log row for auditing and security.
func (store *Store) LogMessage(ctx context.Context, fromUserID, toUserID int, message string) error {
	row := messengermodel.MessageLog{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Message:    message,
		SentAt:     time.Now().UTC(),
	}
	return store.database.WithContext(ctx).Create(&row).Error
}

// DeleteMessageLogOlderThan removes log rows whose sent_at precedes the cutoff epoch.
func (store *Store) DeleteMessageLogOlderThan(ctx context.Context, cutoffUnix int64) error {
	cutoff := time.Unix(cutoffUnix, 0).UTC()
	return store.database.WithContext(ctx).Where("sent_at < ?", cutoff).Delete(&messengermodel.MessageLog{}).Error
}
