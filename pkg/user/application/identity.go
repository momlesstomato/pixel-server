package application

import (
	"context"
	"fmt"
	"strings"

	sdkuser "github.com/momlesstomato/pixel-sdk/events/user"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// UpdateMotto validates and applies one motto change.
func (service *Service) UpdateMotto(ctx context.Context, connID string, userID int, motto string) (domain.User, error) {
	trimmed := strings.TrimSpace(motto)
	if len(trimmed) > 127 {
		return domain.User{}, fmt.Errorf("motto must be <= 127 characters")
	}
	current, err := service.FindByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}
	if service.fire != nil {
		event := &sdkuser.MottoChanged{ConnID: connID, UserID: userID, OldMotto: current.Motto, NewMotto: trimmed}
		service.fire(event)
		if event.Cancelled() {
			return domain.User{}, fmt.Errorf("motto change cancelled by plugin")
		}
	}
	return service.repository.UpdateProfile(ctx, userID, domain.ProfilePatch{Motto: &trimmed})
}

// UpdateFigure validates and applies one figure and gender change.
func (service *Service) UpdateFigure(ctx context.Context, connID string, userID int, gender string, figure string) (domain.User, error) {
	cleanGender := strings.ToUpper(strings.TrimSpace(gender))
	cleanFigure := strings.TrimSpace(figure)
	if cleanGender != "M" && cleanGender != "F" {
		return domain.User{}, fmt.Errorf("gender must be M or F")
	}
	if cleanFigure == "" || len(cleanFigure) > 255 {
		return domain.User{}, fmt.Errorf("figure must be between 1 and 255 characters")
	}
	current, err := service.FindByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}
	if service.fire != nil {
		event := &sdkuser.FigureChanged{
			ConnID: connID, UserID: userID, OldFigure: current.Figure, NewFigure: cleanFigure, Gender: cleanGender,
		}
		service.fire(event)
		if event.Cancelled() {
			return domain.User{}, fmt.Errorf("figure change cancelled by plugin")
		}
	}
	return service.repository.UpdateProfile(ctx, userID, domain.ProfilePatch{Gender: &cleanGender, Figure: &cleanFigure})
}

// SetHomeRoom validates and applies one home-room change.
func (service *Service) SetHomeRoom(ctx context.Context, userID int, homeRoomID int) (domain.User, error) {
	if homeRoomID < -1 {
		return domain.User{}, fmt.Errorf("home room id must be >= -1")
	}
	return service.repository.UpdateProfile(ctx, userID, domain.ProfilePatch{HomeRoomID: &homeRoomID})
}
