package application

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
)

// validateDefinitionCreate enforces application constraints for new definitions.
func validateDefinitionCreate(def domain.Definition) error {
	if def.ItemName == "" {
		return fmt.Errorf("item name is required")
	}
	if def.SpriteID <= 0 {
		return fmt.Errorf("sprite id must be positive")
	}
	return nil
}