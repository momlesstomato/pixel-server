package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// RegisterRoutes registers room HTTP API routes on an HTTP module.
func RegisterRoutes(module *corehttp.Module, chatLogs ChatLogService) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if chatLogs == nil {
		return fmt.Errorf("chat log service is required")
	}
	module.RegisterGET("/api/v1/rooms/:roomId/chat-logs", func(ctx *fiber.Ctx) error {
		roomID, err := parsePositiveID(ctx.Params("roomId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		from, to, parseErr := parseDateRange(ctx.Query("from"), ctx.Query("to"))
		if parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, parseErr.Error())
		}
		entries, findErr := chatLogs.ListByRoom(ctx.UserContext(), roomID, from, to)
		if findErr != nil {
			return fiber.NewError(http.StatusInternalServerError, "failed to retrieve chat logs")
		}
		return ctx.JSON(entries)
	})
	return nil
}

// parsePositiveID validates and converts a string to a positive integer.
func parsePositiveID(value string) (int, error) {
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid positive id: %s", value)
	}
	return id, nil
}

// parseDateRange parses optional from/to date strings into time range.
func parseDateRange(fromStr string, toStr string) (time.Time, time.Time, error) {
	now := time.Now()
	from := now.Truncate(24 * time.Hour)
	to := now
	if fromStr != "" {
		parsed, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid from date: %s", fromStr)
		}
		from = parsed
	}
	if toStr != "" {
		parsed, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid to date: %s", toStr)
		}
		to = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	return from, to, nil
}
