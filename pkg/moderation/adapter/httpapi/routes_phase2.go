package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
)

// RegisterPhase2Routes registers Phase 2 HTTP routes on an HTTP module.
func RegisterPhase2Routes(module *corehttp.Module, tickets TicketService, filters WordFilterService, presets PresetService, visits VisitService) error {
	if tickets != nil {
		registerTicketRoutes(module, tickets)
	}
	if filters != nil {
		registerFilterRoutes(module, filters)
	}
	if presets != nil {
		registerPresetRoutes(module, presets)
	}
	if visits != nil {
		registerVisitRoutes(module, visits)
	}
	return nil
}

func registerTicketRoutes(module *corehttp.Module, svc TicketService) {
	module.RegisterGET("/api/v1/moderation/tickets", func(ctx *fiber.Ctx) error {
		status := domain.TicketStatus(ctx.Query("status"))
		tickets, err := svc.List(ctx.UserContext(), status, 50)
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to list tickets")
		}
		return ctx.JSON(tickets)
	})
	module.RegisterGET("/api/v1/moderation/tickets/:id", func(ctx *fiber.Ctx) error {
		id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return fiber.NewError(http.StatusBadRequest, "invalid ticket id")
		}
		ticket, err := svc.FindByID(ctx.UserContext(), id)
		if err != nil {
			return fiber.NewError(http.StatusNotFound, "ticket not found")
		}
		return ctx.JSON(ticket)
	})
	module.RegisterPATCH("/api/v1/moderation/tickets/:id/close", func(ctx *fiber.Ctx) error {
		id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return fiber.NewError(http.StatusBadRequest, "invalid ticket id")
		}
		if err := svc.Close(ctx.UserContext(), id, domain.TicketClosed); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		return ctx.JSON(map[string]string{"status": "closed"})
	})
}

func registerFilterRoutes(module *corehttp.Module, svc WordFilterService) {
	module.RegisterGET("/api/v1/moderation/wordfilters", func(ctx *fiber.Ctx) error {
		filters, err := svc.ListActive(ctx.UserContext(), ctx.Query("scope"), 0)
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to list filters")
		}
		return ctx.JSON(filters)
	})
	module.RegisterDELETE("/api/v1/moderation/wordfilters/:id", func(ctx *fiber.Ctx) error {
		id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return fiber.NewError(http.StatusBadRequest, "invalid filter id")
		}
		if err := svc.Delete(ctx.UserContext(), id); err != nil {
			return fiber.NewError(http.StatusNotFound, "filter not found")
		}
		return ctx.JSON(map[string]string{"status": "deleted"})
	})
}

func registerPresetRoutes(module *corehttp.Module, svc PresetService) {
	module.RegisterGET("/api/v1/moderation/presets", func(ctx *fiber.Ctx) error {
		presets, err := svc.ListActive(ctx.UserContext())
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to list presets")
		}
		return ctx.JSON(presets)
	})
	module.RegisterDELETE("/api/v1/moderation/presets/:id", func(ctx *fiber.Ctx) error {
		id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return fiber.NewError(http.StatusBadRequest, "invalid preset id")
		}
		if err := svc.Delete(ctx.UserContext(), id); err != nil {
			return fiber.NewError(http.StatusNotFound, "preset not found")
		}
		return ctx.JSON(map[string]string{"status": "deleted"})
	})
}

func registerVisitRoutes(module *corehttp.Module, svc VisitService) {
	module.RegisterGET("/api/v1/moderation/visits/users/:userId", func(ctx *fiber.Ctx) error {
		userID, err := strconv.Atoi(ctx.Params("userId"))
		if err != nil || userID <= 0 {
			return fiber.NewError(http.StatusBadRequest, "invalid user id")
		}
		visits, err := svc.ListByUser(ctx.UserContext(), userID, 50)
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to list visits")
		}
		return ctx.JSON(visits)
	})
	module.RegisterGET("/api/v1/moderation/visits/rooms/:roomId", func(ctx *fiber.Ctx) error {
		roomID, err := strconv.Atoi(ctx.Params("roomId"))
		if err != nil || roomID <= 0 {
			return fiber.NewError(http.StatusBadRequest, "invalid room id")
		}
		visits, err := svc.ListByRoom(ctx.UserContext(), roomID, 50)
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to list visits")
		}
		return ctx.JSON(visits)
	})
}
