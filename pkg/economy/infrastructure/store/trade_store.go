package store

import (
	"context"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/economy/domain"
	economymodel "github.com/momlesstomato/pixel-server/pkg/economy/infrastructure/model"
	"gorm.io/gorm"
)

// CreateTradeLog persists one trade audit row.
func (store *Store) CreateTradeLog(ctx context.Context, log domain.TradeLog) (domain.TradeLog, error) {
	var created domain.TradeLog
	err := store.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := economymodel.TradeLog{
			UserOneID: uint(log.UserOneID), UserTwoID: uint(log.UserTwoID),
			TradedAt: time.Now().UTC(),
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		created = domain.TradeLog{
			ID: int(row.ID), UserOneID: log.UserOneID,
			UserTwoID: log.UserTwoID, TradedAt: row.TradedAt,
		}
		return nil
	})
	return created, err
}
