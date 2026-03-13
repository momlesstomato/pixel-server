package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// registerWardrobeRoutes registers wardrobe and respect-history routes.
func registerWardrobeRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/users/:id/wardrobe", func(ctx *fiber.Ctx) error {
		userID, err := parseUserID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		slots, loadErr := service.LoadWardrobe(ctx.UserContext(), userID)
		if loadErr != nil {
			return mapUserError(loadErr)
		}
		return ctx.JSON(fiber.Map{"slots": slots})
	})
	module.RegisterGET("/api/v1/users/:id/respects", func(ctx *fiber.Ctx) error {
		userID, err := parseUserID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		limit := parseQueryInt(ctx, "limit", 50)
		offset := parseQueryInt(ctx, "offset", 0)
		records, listErr := service.ListRespects(ctx.UserContext(), userID, limit, offset)
		if listErr != nil {
			return mapUserError(listErr)
		}
		return ctx.JSON(fiber.Map{"records": records, "limit": limit, "offset": offset})
	})
}

// parseQueryInt parses one query integer with default fallback.
func parseQueryInt(ctx *fiber.Ctx, key string, fallback int) int {
	value := ctx.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
