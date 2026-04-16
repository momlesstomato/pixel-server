package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
)

const maxStackHelperHeight = 40.0

// EffectiveStackHeight resolves the client-visible height for one item definition pair.
func EffectiveStackHeight(item domain.Item, def domain.Definition) float64 {
	if !IsStackHelperDefinition(def) {
		return def.StackHeight
	}
	value, err := strconv.ParseFloat(item.ExtraData, 64)
	if err != nil || value < 0 {
		return def.StackHeight
	}
	return value
}

// IsStackHelperDefinition reports whether the item definition behaves like a stack helper.
func IsStackHelperDefinition(def domain.Definition) bool {
	if def.InteractionType == domain.InteractionStackHelper {
		return true
	}
	name := strings.ToLower(def.ItemName + " " + def.PublicName)
	return strings.Contains(name, "stackmagic") || strings.Contains(name, "stack helper")
}

// SetStackHeight applies one bounded override height to a placed stack-helper item.
func (service *Service) SetStackHeight(ctx context.Context, itemID, roomID int, height float64) (domain.Item, domain.Definition, error) {
	item, def, err := service.loadPlacedItemDefinition(ctx, itemID, roomID)
	if err != nil {
		return domain.Item{}, domain.Definition{}, err
	}
	if !IsStackHelperDefinition(def) {
		return domain.Item{}, domain.Definition{}, domain.ErrInvalidInteraction
	}
	if height < 0 {
		height = 0
	}
	if height > maxStackHelperHeight {
		height = maxStackHelperHeight
	}
	item.ExtraData = fmt.Sprintf("%.2f", height)
	if err := service.repository.UpdateItemData(ctx, item.ID, item.ExtraData); err != nil {
		return domain.Item{}, domain.Definition{}, err
	}
	return item, def, nil
}
