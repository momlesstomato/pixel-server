package httpapi

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// registerAssignmentRoutes registers user group-assignment routes.
func registerAssignmentRoutes(module *corehttp.Module, service Service) {
	module.RegisterPATCH("/api/v1/users/:id/group", func(ctx *fiber.Ctx) error {
		userID, err := parseIDParam(ctx.Params("id"), "user id")
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		payload := struct {
			GroupID int `json:"group_id"`
		}{}
		if bodyErr := ctx.BodyParser(&payload); bodyErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		access, assignErr := service.ReplaceUserGroups(ctx.UserContext(), userID, []int{payload.GroupID})
		if assignErr != nil {
			return mapPermissionError(assignErr)
		}
		return ctx.JSON(access)
	})
	module.RegisterPATCH("/api/v1/users/:id/groups", func(ctx *fiber.Ctx) error {
		userID, err := parseIDParam(ctx.Params("id"), "user id")
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		payload := struct {
			GroupIDs []int `json:"group_ids"`
		}{}
		if bodyErr := ctx.BodyParser(&payload); bodyErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		access, assignErr := service.ReplaceUserGroups(ctx.UserContext(), userID, payload.GroupIDs)
		if assignErr != nil {
			return mapPermissionError(assignErr)
		}
		return ctx.JSON(access)
	})
}
