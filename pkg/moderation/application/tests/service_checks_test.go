package tests

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsHotelBannedTrue verifies hotel ban lookup returns true.
func TestIsHotelBannedTrue(t *testing.T) {
	svc, _ := createService(t)
	seedAction(t, svc, domain.ScopeHotel, domain.TypeBan, 42)
	banned, err := svc.IsHotelBanned(context.Background(), 42)
	require.NoError(t, err)
	assert.True(t, banned)
}

// TestIsHotelBannedFalse verifies no active ban returns false.
func TestIsHotelBannedFalse(t *testing.T) {
	svc, _ := createService(t)
	banned, err := svc.IsHotelBanned(context.Background(), 99)
	require.NoError(t, err)
	assert.False(t, banned)
}

// TestIsHotelMutedTrue verifies hotel mute lookup returns true.
func TestIsHotelMutedTrue(t *testing.T) {
	svc, _ := createService(t)
	seedAction(t, svc, domain.ScopeHotel, domain.TypeMute, 42)
	muted, err := svc.IsHotelMuted(context.Background(), 42)
	require.NoError(t, err)
	assert.True(t, muted)
}

// TestIsHotelMutedFalse verifies no active mute returns false.
func TestIsHotelMutedFalse(t *testing.T) {
	svc, _ := createService(t)
	muted, err := svc.IsHotelMuted(context.Background(), 99)
	require.NoError(t, err)
	assert.False(t, muted)
}

// TestIsIPBannedTrue verifies IP ban lookup returns true.
func TestIsIPBannedTrue(t *testing.T) {
	svc, _ := createService(t)
	a := &domain.Action{TargetUserID: 1, Scope: domain.ScopeHotel, ActionType: domain.TypeBan, IPAddress: "10.0.0.1"}
	require.NoError(t, svc.Create(context.Background(), a))
	banned, err := svc.IsIPBanned(context.Background(), "10.0.0.1")
	require.NoError(t, err)
	assert.True(t, banned)
}

// TestIsIPBannedFalse verifies no IP ban returns false.
func TestIsIPBannedFalse(t *testing.T) {
	svc, _ := createService(t)
	banned, err := svc.IsIPBanned(context.Background(), "10.0.0.1")
	require.NoError(t, err)
	assert.False(t, banned)
}

// TestFindByIDNotFound verifies missing action returns error.
func TestFindByIDNotFound(t *testing.T) {
	svc, _ := createService(t)
	_, err := svc.FindByID(context.Background(), 999)
	assert.ErrorIs(t, err, domain.ErrActionNotFound)
}

// TestFindByIDSuccess verifies existing action is returned.
func TestFindByIDSuccess(t *testing.T) {
	svc, _ := createService(t)
	a := seedAction(t, svc, domain.ScopeHotel, domain.TypeWarn, 5)
	found, err := svc.FindByID(context.Background(), a.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TypeWarn, found.ActionType)
}

// TestSetEventFirerNoop verifies event firer setter works.
func TestSetEventFirerNoop(t *testing.T) {
	svc, _ := createService(t)
	var fired bool
	svc.SetEventFirer(func(_ sdk.Event) { fired = true })
	assert.False(t, fired)
}
