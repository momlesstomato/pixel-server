package realtime

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/core/broadcast"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/authflow"
	sessionnavigation "github.com/momlesstomato/pixel-server/pkg/session/application/navigation"
	sessionpostauth "github.com/momlesstomato/pixel-server/pkg/session/application/postauth"
)

// ConfigurePostAuth wires post-authentication packet burst dependencies.
func (handler *Handler) ConfigurePostAuth(status sessionpostauth.StatusReader, logins sessionpostauth.LoginRecorder, profiles sessionpostauth.ProfileReader, access sessionpostauth.AccessReader, holder string) {
	handler.postAuthFactory = func(transport *Transport) (*sessionpostauth.UseCase, error) {
		return sessionpostauth.NewUseCase(transport, status, logins, profiles, access, holder)
	}
}

// ConfigureBroadcaster wires distributed broadcast channels for session notifications.
func (handler *Handler) ConfigureBroadcaster(broadcaster broadcast.Broadcaster) {
	handler.broadcaster = broadcaster
}

// ConfigureDesktopView wires desktop-view navigation behavior.
func (handler *Handler) ConfigureDesktopView(checker sessionnavigation.RoomChecker) {
	handler.desktopFactory = func(transport *Transport) (*sessionnavigation.DesktopViewUseCase, error) {
		return sessionnavigation.NewDesktopViewUseCase(transport, checker)
	}
}

// ConfigurePluginEvents wires plugin event dispatch behavior.
func (handler *Handler) ConfigurePluginEvents(fire func(sdk.Event)) {
	handler.fire = fire
}

// ConfigureUserFinder wires real user identity resolution for identity account packets.
func (handler *Handler) ConfigureUserFinder(finder authflow.UserFinder) {
	handler.userFinder = finder
}

// ConfigureUserRuntime wires authenticated user packet runtime behavior.
func (handler *Handler) ConfigureUserRuntime(factory func(*Transport) (UserRuntime, error)) {
	handler.userRuntimeFactory = factory
}

// GenerateConnectionID creates one connection identifier string.
func GenerateConnectionID(source io.Reader) (string, error) {
	reader := source
	if reader == nil {
		reader = rand.Reader
	}
	buffer := make([]byte, 16)
	if _, err := io.ReadFull(reader, buffer); err != nil {
		return "", fmt.Errorf("generate connection id: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}
