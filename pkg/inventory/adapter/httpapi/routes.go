package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
)

// RegisterRoutes registers inventory API routes on an HTTP module.
func RegisterRoutes(module *corehttp.Module, service Service) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if service == nil {
		return fmt.Errorf("inventory service is required")
	}
	registerCurrencyRoutes(module, service)
	registerBadgeRoutes(module, service)
	registerEffectRoutes(module, service)
	return nil
}

// registerCurrencyRoutes registers currency balance routes.
func registerCurrencyRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/inventory/:userId/credits", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("userId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		credits, findErr := service.GetCredits(ctx.UserContext(), userID)
		if findErr != nil {
			return mapInventoryError(findErr)
		}
		return ctx.JSON(fiber.Map{"credits": credits})
	})
	module.RegisterGET("/api/v1/inventory/:userId/currencies", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("userId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		currencies, findErr := service.ListCurrencies(ctx.UserContext(), userID)
		if findErr != nil {
			return mapInventoryError(findErr)
		}
		return ctx.JSON(currencies)
	})
}

// registerBadgeRoutes registers badge management routes.
func registerBadgeRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/inventory/:userId/badges", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("userId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		badges, findErr := service.ListBadges(ctx.UserContext(), userID)
		if findErr != nil {
			return mapInventoryError(findErr)
		}
		return ctx.JSON(badges)
	})
	module.RegisterPOST("/api/v1/inventory/:userId/badges", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("userId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		var payload badgeRequest
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		badge, awardErr := service.AwardBadge(ctx.UserContext(), userID, payload.BadgeCode)
		if awardErr != nil {
			return mapInventoryError(awardErr)
		}
		return ctx.Status(http.StatusCreated).JSON(badge)
	})
}

// registerEffectRoutes registers effect listing route.
func registerEffectRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/inventory/:userId/effects", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("userId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		effects, findErr := service.ListEffects(ctx.UserContext(), userID)
		if findErr != nil {
			return mapInventoryError(findErr)
		}
		return ctx.JSON(effects)
	})
}

// parsePositiveID validates and parses a positive integer identifier.
func parsePositiveID(value string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("id must be a positive integer")
	}
	return id, nil
}

// mapInventoryError maps domain errors to HTTP responses.
func mapInventoryError(err error) error {
	switch err {
	case domain.ErrBadgeNotFound, domain.ErrEffectNotFound, domain.ErrCurrencyTypeUnknown:
		return fiber.NewError(http.StatusNotFound, err.Error())
	case domain.ErrBadgeAlreadyOwned, domain.ErrBadgeSlotInvalid:
		return fiber.NewError(http.StatusConflict, err.Error())
	case domain.ErrInsufficientCurrency, domain.ErrInventoryFull:
		return fiber.NewError(http.StatusForbidden, err.Error())
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "must") || strings.Contains(msg, "required") {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	return fiber.NewError(http.StatusInternalServerError, err.Error())
}

// badgeRequest defines badge award payload.
type badgeRequest struct {
	// BadgeCode stores badge identifier code.
	BadgeCode string `json:"badge_code"`
}
