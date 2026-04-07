package store

import (
	"context"
	"errors"

	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
	submodel "github.com/momlesstomato/pixel-server/pkg/subscription/infrastructure/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const clubGiftWithSprite = `
	SELECT cg.*, id.sprite_id AS effective_sprite_id
	FROM subscription_club_gifts cg
	LEFT JOIN item_definitions id ON id.id = cg.item_definition_id
`

type resolvedClubGift struct {
	submodel.ClubGift
	EffectiveSpriteID int `gorm:"column:effective_sprite_id"`
}

// FindPaydayConfig resolves the active HC payday configuration.
func (store *Store) FindPaydayConfig(ctx context.Context) (domain.PaydayConfig, error) {
	var row submodel.PaydayConfig
	err := store.database.WithContext(ctx).First(&row, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.PaydayConfig{}, domain.ErrPaydayConfigNotFound
	}
	if err != nil {
		return domain.PaydayConfig{}, err
	}
	return mapPaydayConfig(row), nil
}

// SavePaydayConfig upserts the active HC payday configuration.
func (store *Store) SavePaydayConfig(ctx context.Context, cfg domain.PaydayConfig) (domain.PaydayConfig, error) {
	row := submodel.PaydayConfig{
		ID:                  1,
		IntervalDays:        cfg.IntervalDays,
		KickbackPercentage:  cfg.KickbackPercentage,
		FlatCredits:         cfg.FlatCredits,
		MinimumCreditsSpent: cfg.MinimumCreditsSpent,
		StreakBonusCredits:  cfg.StreakBonusCredits,
	}
	if err := store.database.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error; err != nil {
		return domain.PaydayConfig{}, err
	}
	return mapPaydayConfig(row), nil
}

// FindBenefitsState resolves per-user subscription benefits progress.
func (store *Store) FindBenefitsState(ctx context.Context, userID int) (domain.BenefitsState, error) {
	var row submodel.BenefitsState
	err := store.database.WithContext(ctx).First(&row, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.BenefitsState{}, domain.ErrBenefitsStateNotFound
	}
	if err != nil {
		return domain.BenefitsState{}, err
	}
	return mapBenefitsState(row), nil
}

// SaveBenefitsState upserts per-user subscription benefits progress.
func (store *Store) SaveBenefitsState(ctx context.Context, state domain.BenefitsState) (domain.BenefitsState, error) {
	row := submodel.BenefitsState{
		UserID:               uint(state.UserID),
		FirstSubscriptionAt:  state.FirstSubscriptionAt,
		NextPaydayAt:         state.NextPaydayAt,
		CycleCreditsSpent:    state.CycleCreditsSpent,
		RewardStreak:         state.RewardStreak,
		TotalCreditsRewarded: state.TotalCreditsRewarded,
		TotalCreditsMissed:   state.TotalCreditsMissed,
		ClubGiftsClaimed:     state.ClubGiftsClaimed,
	}
	if err := store.database.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error; err != nil {
		return domain.BenefitsState{}, err
	}
	return store.FindBenefitsState(ctx, state.UserID)
}

// ListClubGifts resolves all enabled club gift options.
func (store *Store) ListClubGifts(ctx context.Context) ([]domain.ClubGift, error) {
	var rows []resolvedClubGift
	if err := store.database.WithContext(ctx).Raw(clubGiftWithSprite + "WHERE cg.enabled = true ORDER BY cg.order_num ASC, cg.id ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.ClubGift, len(rows))
	for i, row := range rows {
		result[i] = mapClubGift(row)
	}
	return result, nil
}

// FindClubGiftByName resolves one club gift by case-insensitive name.
func (store *Store) FindClubGiftByName(ctx context.Context, name string) (domain.ClubGift, error) {
	var row resolvedClubGift
	res := store.database.WithContext(ctx).Raw(clubGiftWithSprite+"WHERE LOWER(cg.name) = LOWER(?) LIMIT 1", name).Scan(&row)
	if res.Error != nil {
		return domain.ClubGift{}, res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ClubGift{}, domain.ErrClubGiftNotFound
	}
	return mapClubGift(row), nil
}

func mapPaydayConfig(row submodel.PaydayConfig) domain.PaydayConfig {
	return domain.PaydayConfig{
		IntervalDays:        row.IntervalDays,
		KickbackPercentage:  row.KickbackPercentage,
		FlatCredits:         row.FlatCredits,
		MinimumCreditsSpent: row.MinimumCreditsSpent,
		StreakBonusCredits:  row.StreakBonusCredits,
		UpdatedAt:           row.UpdatedAt,
	}
}

func mapBenefitsState(row submodel.BenefitsState) domain.BenefitsState {
	return domain.BenefitsState{
		UserID:               int(row.UserID),
		FirstSubscriptionAt:  row.FirstSubscriptionAt,
		NextPaydayAt:         row.NextPaydayAt,
		CycleCreditsSpent:    row.CycleCreditsSpent,
		RewardStreak:         row.RewardStreak,
		TotalCreditsRewarded: row.TotalCreditsRewarded,
		TotalCreditsMissed:   row.TotalCreditsMissed,
		ClubGiftsClaimed:     row.ClubGiftsClaimed,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}

func mapClubGift(row resolvedClubGift) domain.ClubGift {
	return domain.ClubGift{
		ID:               int(row.ID),
		Name:             row.Name,
		ItemDefinitionID: int(row.ItemDefinitionID),
		SpriteID:         row.EffectiveSpriteID,
		ExtraData:        row.ExtraData,
		DaysRequired:     row.DaysRequired,
		VIPOnly:          row.VIPOnly,
		Enabled:          row.Enabled,
		OrderNum:         row.OrderNum,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}
