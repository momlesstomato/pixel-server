package application

import (
	"context"
	"time"
)

// PurgeOldOfflineMessages deletes offline messages older than the configured TTL.
func (service *Service) PurgeOldOfflineMessages(ctx context.Context) error {
	cutoff := time.Now().UTC().AddDate(0, 0, -service.config.OfflineMsgTTLDays)
	return service.repository.DeleteOfflineMessagesOlderThan(ctx, cutoff.Unix())
}

// PurgeOldMessageLogs deletes message log entries older than the configured TTL.
func (service *Service) PurgeOldMessageLogs(ctx context.Context) error {
	cutoff := time.Now().UTC().AddDate(0, 0, -service.config.MessageLogTTLDays)
	return service.repository.DeleteMessageLogOlderThan(ctx, cutoff.Unix())
}

// StartPurgeTicker launches a background goroutine that periodically purges expired records.
// The goroutine exits when ctx is cancelled.
func (service *Service) StartPurgeTicker(ctx context.Context) {
	interval := time.Duration(service.config.PurgeIntervalSeconds) * time.Second
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = service.PurgeOldOfflineMessages(ctx)
				_ = service.PurgeOldMessageLogs(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}
