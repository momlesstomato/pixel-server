package httpapi

import (
	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// registerSettingsRoutes registers user settings routes.
func registerSettingsRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/users/:id/settings", func(ctx *fiber.Ctx) error {
		userID, err := parseUserID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		settings, loadErr := service.LoadSettings(ctx.UserContext(), userID)
		if loadErr != nil {
			return mapUserError(loadErr)
		}
		return ctx.JSON(settings)
	})
	module.RegisterPATCH("/api/v1/users/:id/settings", func(ctx *fiber.Ctx) error {
		userID, err := parseUserID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		var payload settingsPatchRequest
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}
		patch := domain.SettingsPatch{
			VolumeSystem: payload.VolumeSystem, VolumeFurni: payload.VolumeFurni, VolumeTrax: payload.VolumeTrax,
			OldChat: payload.OldChat, RoomInvites: payload.RoomInvites, CameraFollow: payload.CameraFollow,
			Flags: payload.Flags, ChatType: payload.ChatType,
		}
		settings, saveErr := service.SaveSettings(ctx.UserContext(), userID, patch)
		if saveErr != nil {
			return mapUserError(saveErr)
		}
		return ctx.JSON(settings)
	})
}

// settingsPatchRequest defines settings patch payload.
type settingsPatchRequest struct {
	// VolumeSystem stores optional global system volume percentage.
	VolumeSystem *int `json:"volume_system"`
	// VolumeFurni stores optional furniture volume percentage.
	VolumeFurni *int `json:"volume_furni"`
	// VolumeTrax stores optional trax volume percentage.
	VolumeTrax *int `json:"volume_trax"`
	// OldChat stores optional classic chat style preference.
	OldChat *bool `json:"old_chat"`
	// RoomInvites stores optional room invite preference.
	RoomInvites *bool `json:"room_invites"`
	// CameraFollow stores optional camera follow preference.
	CameraFollow *bool `json:"camera_follow"`
	// Flags stores optional settings bitmask field.
	Flags *int `json:"flags"`
	// ChatType stores optional chat rendering type.
	ChatType *int `json:"chat_type"`
}

// userResponse defines user profile API response.
type userResponse struct {
	// ID stores stable user identifier.
	ID int `json:"id"`
	// Username stores account username.
	Username string `json:"username"`
	// Figure stores avatar figure string.
	Figure string `json:"figure"`
	// Gender stores avatar gender marker.
	Gender string `json:"gender"`
	// Motto stores profile motto.
	Motto string `json:"motto"`
	// RealName stores profile real name value.
	RealName string `json:"real_name"`
	// RespectsReceived stores total received respects.
	RespectsReceived int `json:"respects_received"`
	// HomeRoomID stores configured home room identifier.
	HomeRoomID int `json:"home_room_id"`
	// CanChangeName stores account rename capability.
	CanChangeName bool `json:"can_change_name"`
	// NoobnessLevel stores account age tier marker.
	NoobnessLevel int `json:"noobness_level"`
	// SafetyLocked stores account safety lock marker.
	SafetyLocked bool `json:"safety_locked"`
	// GroupID stores permission group identifier.
	GroupID int `json:"group_id"`
}

// userResponseFromDomain maps domain user payload to API response payload.
func userResponseFromDomain(user domain.User) userResponse {
	return userResponse{
		ID: user.ID, Username: user.Username, Figure: user.Figure, Gender: user.Gender,
		Motto: user.Motto, RealName: user.RealName, RespectsReceived: user.RespectsReceived,
		HomeRoomID: user.HomeRoomID, CanChangeName: user.CanChangeName,
		NoobnessLevel: user.NoobnessLevel, SafetyLocked: user.SafetyLocked, GroupID: user.GroupID,
	}
}
