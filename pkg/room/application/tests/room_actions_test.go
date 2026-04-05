package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/room/application"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestSoftDelete_NoRepository verifies ErrRoomNotFound when no repo is set.
func TestSoftDelete_NoRepository(t *testing.T) {
	svc := newService(t)
	err := svc.SoftDelete(context.Background(), 1)
	assert.ErrorIs(t, err, domain.ErrRoomNotFound)
}

// TestSoftDelete_Success verifies soft delete delegates to repository.
func TestSoftDelete_Success(t *testing.T) {
	svc := newService(t)
	repo := &roomRepoStub{rooms: map[int]domain.Room{1: {ID: 1}}}
	svc.SetRoomRepository(repo)
	assert.NoError(t, svc.SoftDelete(context.Background(), 1))
}

// TestListBans_ReturnsEmpty verifies empty ban list from stub.
func TestListBans_ReturnsEmpty(t *testing.T) {
	svc := newService(t)
	bans, err := svc.ListBans(context.Background(), 1)
	assert.NoError(t, err)
	assert.Empty(t, bans)
}

// TestFindBan_NoBan verifies nil ban for non-banned user.
func TestFindBan_NoBan(t *testing.T) {
	svc := newService(t)
	ban, err := svc.FindBan(context.Background(), 1, 99)
	assert.NoError(t, err)
	assert.Nil(t, ban)
}

// TestFindBan_Active verifies existing ban is returned.
func TestFindBan_Active(t *testing.T) {
	bans := &banRepoStub{banned: map[[2]int]bool{{1, 10}: true}}
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	svc, err := application.NewService(newModelRepo(), bans, &rightsRepoStub{}, mgr, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	ban, findErr := svc.FindBan(context.Background(), 1, 10)
	assert.NoError(t, findErr)
	assert.NotNil(t, ban)
}

// TestRemoveBan_Success verifies ban removal delegates to repository.
func TestRemoveBan_Success(t *testing.T) {
	svc := newService(t)
	assert.NoError(t, svc.RemoveBan(context.Background(), 1))
}
