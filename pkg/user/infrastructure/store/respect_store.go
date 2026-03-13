package store

import (
	"context"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// RemainingRespects returns remaining daily respects for one actor and target type.
func (repository *Repository) RemainingRespects(ctx context.Context, actorUserID int, targetType domain.RespectTargetType, at time.Time) (int, error) {
	dayStart := utcDayStart(at)
	var count int64
	err := repository.database.WithContext(ctx).Model(&usermodel.Respect{}).Where("actor_user_id = ? AND respected_at = ? AND target_type = ?", actorUserID, dayStart, int16(targetType)).Count(&count).Error
	if err != nil {
		return 0, err
	}
	remaining := domain.DefaultDailyRespects - int(count)
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}

// RecordRespect persists one respect event and returns updated respects received counter.
func (repository *Repository) RecordRespect(ctx context.Context, actorUserID int, targetID int, targetType domain.RespectTargetType, at time.Time) (int, error) {
	dayStart := utcDayStart(at)
	updated := 0
	err := repository.database.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if _, err := repository.loadRecord(ctx, actorUserID); err != nil {
			return err
		}
		if targetType == domain.RespectTargetUser {
			if _, err := repository.loadRecord(ctx, targetID); err != nil {
				return err
			}
		}
		var count int64
		query := transaction.Model(&usermodel.Respect{}).Where("actor_user_id = ? AND respected_at = ? AND target_type = ?", actorUserID, dayStart, int16(targetType))
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if int(count) >= domain.DefaultDailyRespects {
			return domain.ErrRespectLimitReached
		}
		respect := usermodel.Respect{ActorUserID: uint(actorUserID), TargetID: uint(targetID), TargetType: int16(targetType), RespectedAt: dayStart}
		if err := transaction.Create(&respect).Error; err != nil {
			return err
		}
		if targetType != domain.RespectTargetUser {
			updated = 0
			return nil
		}
		if err := transaction.Model(&usermodel.Record{}).Where("id = ?", targetID).Update("respects_received", gorm.Expr("respects_received + 1")).Error; err != nil {
			return err
		}
		var record usermodel.Record
		if err := transaction.Select("respects_received").First(&record, targetID).Error; err != nil {
			return err
		}
		updated = record.RespectsReceived
		return nil
	})
	if err != nil {
		return 0, err
	}
	return updated, nil
}
