package store

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/economy/domain"
	economymodel "github.com/momlesstomato/pixel-server/pkg/economy/infrastructure/model"
)

// GetPriceHistory resolves price history data for one sprite.
func (store *Store) GetPriceHistory(ctx context.Context, spriteID int) ([]domain.PriceHistory, error) {
	var rows []economymodel.PriceHistory
	err := store.database.WithContext(ctx).
		Where("sprite_id = ?", spriteID).Order("day_offset ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]domain.PriceHistory, len(rows))
	for i, row := range rows {
		result[i] = domain.PriceHistory{
			ID: int(row.ID), SpriteID: row.SpriteID,
			DayOffset: row.DayOffset, AvgPrice: row.AvgPrice,
			SoldCount: row.SoldCount, RecordedAt: row.RecordedAt,
		}
	}
	return result, nil
}

// RecordPriceHistory persists one aggregated price data point.
func (store *Store) RecordPriceHistory(ctx context.Context, ph domain.PriceHistory) error {
	row := economymodel.PriceHistory{
		SpriteID: ph.SpriteID, DayOffset: ph.DayOffset,
		AvgPrice: ph.AvgPrice, SoldCount: ph.SoldCount,
		RecordedAt: ph.RecordedAt,
	}
	return store.database.WithContext(ctx).Create(&row).Error
}
