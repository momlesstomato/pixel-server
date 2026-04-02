package realtime

import (
	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
)

// parseBadgeSlots reads badge slot assignments from an update_badges body.
func parseBadgeSlots(body []byte) []domain.BadgeSlot {
	reader := codec.NewReader(body)
	var slots []domain.BadgeSlot
	for i := 1; i <= domain.MaxBadgeSlots; i++ {
		slotID, err := reader.ReadInt32()
		if err != nil {
			break
		}
		code, err := reader.ReadString()
		if err != nil {
			break
		}
		if code != "" {
			slots = append(slots, domain.BadgeSlot{SlotID: int(slotID), BadgeCode: code})
		}
	}
	return slots
}

// parseEffectID reads the effect identifier from an effect_activate body.
func parseEffectID(body []byte) int {
	reader := codec.NewReader(body)
	id, err := reader.ReadInt32()
	if err != nil {
		return 0
	}
	return int(id)
}
