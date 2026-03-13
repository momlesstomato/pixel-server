package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// RegisterRoutes registers user profile API routes on an HTTP module.
func RegisterRoutes(module *corehttp.Module, service Service) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if service == nil {
		return fmt.Errorf("user service is required")
	}
	registerIdentityRoutes(module, service)
	registerSettingsRoutes(module, service)
	registerRespectRoutes(module, service)
	return nil
}

// registerIdentityRoutes registers identity routes.
func registerIdentityRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/users/:id", func(ctx *fiber.Ctx) error {
		userID, err := parseUserID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		user, findErr := service.FindByID(ctx.UserContext(), userID)
		if findErr != nil {
			return mapUserError(findErr)
		}
		return ctx.JSON(userResponseFromDomain(user))
	})
	module.RegisterPATCH("/api/v1/users/:id", func(ctx *fiber.Ctx) error {
		userID, err := parseUserID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		var payload userPatchRequest
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		patch := domain.ProfilePatch{Figure: payload.Figure, Gender: payload.Gender, Motto: payload.Motto, HomeRoomID: payload.HomeRoomID}
		updated, updateErr := service.UpdateProfile(ctx.UserContext(), userID, patch)
		if updateErr != nil {
			return mapUserError(updateErr)
		}
		return ctx.JSON(userResponseFromDomain(updated))
	})
}

// registerRespectRoutes registers respect routes.
func registerRespectRoutes(module *corehttp.Module, service Service) {
	module.RegisterPOST("/api/v1/users/:id/respect", func(ctx *fiber.Ctx) error {
		targetID, err := parseUserID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		var payload respectRequest
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		result, respectErr := service.RecordUserRespect(ctx.UserContext(), payload.ActorUserID, targetID, time.Now().UTC())
		if respectErr != nil {
			return mapUserError(respectErr)
		}
		packet := userapplication.RespectResult{RespectsReceived: result.RespectsReceived, Remaining: result.Remaining}
		return ctx.JSON(fiber.Map{"respects_received": packet.RespectsReceived, "remaining": packet.Remaining})
	})
}

// parseUserID validates and parses one user identifier string.
func parseUserID(value string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("user id must be a positive integer")
	}
	return id, nil
}

// mapUserError maps domain/application errors to HTTP responses.
func mapUserError(err error) error {
	if err == nil {
		return nil
	}
	if err == domain.ErrUserNotFound {
		return fiber.NewError(http.StatusNotFound, err.Error())
	}
	if err == domain.ErrRespectLimitReached {
		return fiber.NewError(http.StatusConflict, err.Error())
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "must") || strings.Contains(message, "invalid") || strings.Contains(message, "required") {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	return fiber.NewError(http.StatusInternalServerError, err.Error())
}

// userPatchRequest defines user identity patch payload.
type userPatchRequest struct {
	// Figure stores optional figure value.
	Figure *string `json:"figure"`
	// Gender stores optional gender value.
	Gender *string `json:"gender"`
	// Motto stores optional motto value.
	Motto *string `json:"motto"`
	// HomeRoomID stores optional home room identifier.
	HomeRoomID *int `json:"home_room_id"`
}

// respectRequest defines user respect payload.
type respectRequest struct {
	// ActorUserID stores actor user identifier.
	ActorUserID int `json:"actor_user_id"`
}
