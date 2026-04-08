package tests

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkmoderation "github.com/momlesstomato/pixel-sdk/events/moderation"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateFiresKickEvents verifies kick actions emit before and after SDK events.
func TestCreateFiresKickEvents(t *testing.T) {
	svc, _ := createService(t)
	events := make([]sdk.Event, 0, 2)
	svc.SetEventFirer(func(event sdk.Event) {
		events = append(events, event)
	})
	action := &domain.Action{TargetUserID: 8, IssuerID: 4, Scope: domain.ScopeHotel, ActionType: domain.TypeKick}
	require.NoError(t, svc.Create(context.Background(), action))
	require.Len(t, events, 2)
	_, isBefore := events[0].(*sdkmoderation.UserKicking)
	_, isAfter := events[1].(*sdkmoderation.UserKicked)
	assert.True(t, isBefore)
	assert.True(t, isAfter)
}

// TestCreateHonorsCancelledKickEvent verifies cancelled kick events prevent persistence.
func TestCreateHonorsCancelledKickEvent(t *testing.T) {
	svc, repo := createService(t)
	svc.SetEventFirer(func(event sdk.Event) {
		if kick, ok := event.(*sdkmoderation.UserKicking); ok {
			kick.Cancel()
		}
	})
	action := &domain.Action{TargetUserID: 8, IssuerID: 4, Scope: domain.ScopeHotel, ActionType: domain.TypeKick}
	err := svc.Create(context.Background(), action)
	assert.Error(t, err)
	assert.Empty(t, repo.actions)
}

// TestCreateAllowsRoomAlertRegistry verifies room alert warn records can be stored without a target user.
func TestCreateAllowsRoomAlertRegistry(t *testing.T) {
	svc, _ := createService(t)
	action := &domain.Action{Scope: domain.ScopeRoom, ActionType: domain.TypeWarn, RoomID: 9, IssuerID: 4, Reason: "room alert"}
	require.NoError(t, svc.Create(context.Background(), action))
	stored, err := svc.List(context.Background(), domain.ListFilter{RoomID: 9, Limit: 10})
	require.NoError(t, err)
	require.Len(t, stored, 1)
	assert.Equal(t, 0, stored[0].TargetUserID)
	assert.Equal(t, 9, stored[0].RoomID)
}
