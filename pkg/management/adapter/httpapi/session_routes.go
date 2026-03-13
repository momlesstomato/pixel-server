package httpapi

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// RegisterSessionRoutes registers session management API routes.
func RegisterSessionRoutes(module *corehttp.Module, sessions SessionLister, closer SessionCloser) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if sessions == nil {
		return fmt.Errorf("session lister is required")
	}
	module.RegisterGET("/api/v1/sessions", func(ctx *fiber.Ctx) error {
		all, err := sessions.ListAll()
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
		instance := ctx.Query("instance")
		if instance != "" {
			var filtered []sessionResponse
			for _, s := range all {
				if s.InstanceID == instance {
					filtered = append(filtered, mapSession(s))
				}
			}
			return ctx.JSON(fiber.Map{"sessions": filtered, "count": len(filtered)})
		}
		mapped := make([]sessionResponse, 0, len(all))
		for _, s := range all {
			mapped = append(mapped, mapSession(s))
		}
		return ctx.JSON(fiber.Map{"sessions": mapped, "count": len(mapped)})
	})
	module.RegisterGET("/api/v1/sessions/:connID", func(ctx *fiber.Ctx) error {
		session, found := sessions.FindByConnID(ctx.Params("connID"))
		if !found {
			return fiber.NewError(http.StatusNotFound, "session not found")
		}
		return ctx.JSON(mapSession(session))
	})
	module.RegisterDELETE("/api/v1/sessions/:connID", func(ctx *fiber.Ctx) error {
		connID := ctx.Params("connID")
		_, found := sessions.FindByConnID(connID)
		if !found {
			return fiber.NewError(http.StatusNotFound, "session not found")
		}
		if closer != nil {
			_ = closer.Close(ctx.UserContext(), connID, 1008, "management disconnect")
		}
		sessions.Remove(connID)
		return ctx.JSON(fiber.Map{"disconnected": connID})
	})
	return nil
}
