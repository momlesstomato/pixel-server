package httpapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
)

// RegisterRoutes registers moderation HTTP routes on an HTTP module.
func RegisterRoutes(module *corehttp.Module, svc ModerationService) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if svc == nil {
		return fmt.Errorf("moderation service is required")
	}
	module.RegisterGET("/api/v1/moderation/actions", func(ctx *fiber.Ctx) error {
		filter := domain.ListFilter{Limit: 50}
		if v := ctx.Query("scope"); v != "" {
			filter.Scope = domain.ActionScope(v)
		}
		if v := ctx.Query("action_type"); v != "" {
			filter.ActionType = domain.ActionType(v)
		}
		if v := ctx.Query("target_user_id"); v != "" {
			if id, err := strconv.Atoi(v); err == nil && id > 0 {
				filter.TargetUserID = id
			}
		}
		if v := ctx.Query("active"); v != "" {
			b := v == "true"
			filter.Active = &b
		}
		if v := ctx.Query("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
				filter.Limit = n
			}
		}
		if v := ctx.Query("page"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				filter.Offset = (n - 1) * filter.Limit
			}
		}
		actions, err := svc.List(ctx.UserContext(), filter)
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to list actions")
		}
		return ctx.JSON(actions)
	})
	module.RegisterGET("/api/v1/moderation/actions/:id", func(ctx *fiber.Ctx) error {
		id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return fiber.NewError(http.StatusBadRequest, "invalid action id")
		}
		action, err := svc.FindByID(ctx.UserContext(), id)
		if err != nil {
			return fiber.NewError(http.StatusNotFound, "action not found")
		}
		return ctx.JSON(action)
	})
	module.RegisterGET("/api/v1/moderation/users/:userId/actions", func(ctx *fiber.Ctx) error {
		userID, err := strconv.Atoi(ctx.Params("userId"))
		if err != nil || userID <= 0 {
			return fiber.NewError(http.StatusBadRequest, "invalid user id")
		}
		filter := domain.ListFilter{TargetUserID: userID, Limit: 50}
		actions, err := svc.List(ctx.UserContext(), filter)
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to list actions")
		}
		return ctx.JSON(actions)
	})
	module.RegisterPATCH("/api/v1/moderation/actions/:id/deactivate", func(ctx *fiber.Ctx) error {
		id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return fiber.NewError(http.StatusBadRequest, "invalid action id")
		}
		if err := svc.Deactivate(ctx.UserContext(), id, 0); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		return ctx.JSON(map[string]string{"status": "deactivated"})
	})
	return nil
}
