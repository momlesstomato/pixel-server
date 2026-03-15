package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// registerWardrobeRoutes registers wardrobe, respect-history, and ignore routes.
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
	module.RegisterGET("/api/v1/users/:id/ignores", func(ctx *fiber.Ctx) error {
		userID, err := parseUserID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		entries, listErr := service.ListIgnoredUsers(ctx.UserContext(), userID)
		if listErr != nil {
			return mapUserError(listErr)
		}
		return ctx.JSON(fiber.Map{"entries": entries})
	})
	module.RegisterPOST("/api/v1/users/:id/ignores", func(ctx *fiber.Ctx) error {
		userID, err := parseUserID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		var payload ignoreRequest
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		if payload.TargetUserID <= 0 {
			return fiber.NewError(http.StatusBadRequest, "target_user_id must be a positive integer")
		}
		if ignoreErr := service.AdminIgnoreUser(ctx.UserContext(), userID, payload.TargetUserID); ignoreErr != nil {
			return mapUserError(ignoreErr)
		}
		return ctx.SendStatus(http.StatusNoContent)
	})
	module.RegisterDELETE("/api/v1/users/:id/ignores/:targetId", func(ctx *fiber.Ctx) error {
		userID, err := parseUserID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		targetID, targetErr := parseUserID(ctx.Params("targetId"))
		if targetErr != nil {
			return fiber.NewError(http.StatusBadRequest, targetErr.Error())
		}
		if unignoreErr := service.AdminUnignoreUser(ctx.UserContext(), userID, targetID); unignoreErr != nil {
			return mapUserError(unignoreErr)
		}
		return ctx.SendStatus(http.StatusNoContent)
	})
}

// ignoreRequest defines admin ignore payload.
type ignoreRequest struct {
	// TargetUserID stores target user identifier.
	TargetUserID int `json:"target_user_id"`
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
