package httpapi

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	sdk "github.com/momlesstomato/pixel-sdk"
	sdkmgmt "github.com/momlesstomato/pixel-sdk/events/management"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// RegisterHotelRoutes registers hotel status management API routes.
func RegisterHotelRoutes(module *corehttp.Module, hotel HotelManager, fire func(sdk.Event)) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if hotel == nil {
		return fmt.Errorf("hotel manager is required")
	}
	module.RegisterGET("/api/v1/hotel/status", func(ctx *fiber.Ctx) error {
		status, err := hotel.Current(ctx.UserContext())
		if err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
		return ctx.JSON(mapHotelStatus(status))
	})
	module.RegisterPOST("/api/v1/hotel/close", func(ctx *fiber.Ctx) error {
		var payload closeRequest
		if err := ctx.BodyParser(&payload); err != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		if fire != nil {
			event := &sdkmgmt.HotelClosing{MinutesUntilClose: payload.MinutesUntilClose, DurationMinutes: payload.DurationMinutes, ThrowUsers: payload.ThrowUsers}
			fire(event)
			if event.Cancelled() {
				return fiber.NewError(http.StatusConflict, "hotel close cancelled by plugin")
			}
		}
		status, err := hotel.ScheduleClose(ctx.UserContext(), payload.MinutesUntilClose, payload.DurationMinutes, payload.ThrowUsers)
		if err != nil {
			return fiber.NewError(http.StatusConflict, err.Error())
		}
		return ctx.JSON(mapHotelStatus(status))
	})
	module.RegisterPOST("/api/v1/hotel/reopen", func(ctx *fiber.Ctx) error {
		if fire != nil {
			event := &sdkmgmt.HotelReopening{}
			fire(event)
			if event.Cancelled() {
				return fiber.NewError(http.StatusConflict, "hotel reopen cancelled by plugin")
			}
		}
		status, err := hotel.Reopen(ctx.UserContext())
		if err != nil {
			return fiber.NewError(http.StatusConflict, err.Error())
		}
		return ctx.JSON(mapHotelStatus(status))
	})
	return nil
}

// closeRequest defines hotel close scheduling payload.
type closeRequest struct {
	// MinutesUntilClose defines countdown minutes before closing.
	MinutesUntilClose int32 `json:"minutes_until_close"`
	// DurationMinutes defines maintenance window duration in minutes.
	DurationMinutes int32 `json:"duration_minutes"`
	// ThrowUsers defines whether connected users are disconnected at close.
	ThrowUsers bool `json:"throw_users"`
}

// hotelStatusResponse defines hotel status API response.
type hotelStatusResponse struct {
	// State defines current hotel lifecycle state.
	State string `json:"state"`
	// CloseAt defines scheduled close timestamp.
	CloseAt *string `json:"close_at,omitempty"`
	// ReopenAt defines scheduled reopen timestamp.
	ReopenAt *string `json:"reopen_at,omitempty"`
	// ThrowUsers defines whether users are removed at close.
	ThrowUsers bool `json:"throw_users"`
}
