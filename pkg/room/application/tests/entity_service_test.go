package tests

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkentity "github.com/momlesstomato/pixel-sdk/events/room/entity"
	"github.com/momlesstomato/pixel-server/pkg/room/application"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newEntityService(t *testing.T) (*application.EntityService, *engine.Manager) {
	t.Helper()
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	svc, err := application.NewEntityService(mgr, zap.NewNop())
	require.NoError(t, err)
	return svc, mgr
}

func must3x3Grid() [][]domain.Tile {
	grid := make([][]domain.Tile, 3)
	for i := range grid {
		grid[i] = make([]domain.Tile, 3)
		for j := range grid[i] {
			grid[i][j] = domain.Tile{X: j, Y: i, State: domain.TileOpen}
		}
	}
	return grid
}

func loadedInstance(t *testing.T, mgr *engine.Manager) (*engine.Instance, *domain.RoomEntity) {
	t.Helper()
	layout := domain.Layout{Slug: "test", DoorX: 1, DoorY: 1, Grid: must3x3Grid()}
	inst := mgr.Load(1, layout)
	entity := domain.NewPlayerEntity(0, 1, "conn1", "Alice", "", "", "M",
		domain.Tile{X: 1, Y: 1, State: domain.TileOpen})
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: &entity, Reply: reply})
	require.NoError(t, <-reply)
	return inst, &entity
}

// TestNewEntityService_NilManager verifies nil manager is rejected.
func TestNewEntityService_NilManager(t *testing.T) {
	_, err := application.NewEntityService(nil, zap.NewNop())
	assert.Error(t, err)
}

// TestNewEntityService_NilLogger creates service with nil logger successfully.
func TestNewEntityService_NilLogger(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	svc, err := application.NewEntityService(mgr, nil)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

// TestEntityService_Walk_Success verifies walk is dispatched to engine.
func TestEntityService_Walk_Success(t *testing.T) {
	svc, mgr := newEntityService(t)
	defer mgr.StopAll()
	inst, entity := loadedInstance(t, mgr)
	err := svc.Walk(context.Background(), inst, entity, 2, 2)
	assert.NoError(t, err)
}

// TestEntityService_Walk_FiresEvents verifies EntityMoving and EntityMoved are fired.
func TestEntityService_Walk_FiresEvents(t *testing.T) {
	svc, mgr := newEntityService(t)
	defer mgr.StopAll()
	inst, entity := loadedInstance(t, mgr)
	var fired []string
	svc.SetEventFirer(func(e sdk.Event) {
		switch e.(type) {
		case *sdkentity.EntityMoving:
			fired = append(fired, "moving")
		case *sdkentity.EntityMoved:
			fired = append(fired, "moved")
		}
	})
	require.NoError(t, svc.Walk(context.Background(), inst, entity, 2, 2))
	assert.Contains(t, fired, "moving")
	assert.Contains(t, fired, "moved")
}

// TestEntityService_Walk_Cancelled verifies cancelled EntityMoving aborts walk.
func TestEntityService_Walk_Cancelled(t *testing.T) {
	svc, mgr := newEntityService(t)
	defer mgr.StopAll()
	inst, entity := loadedInstance(t, mgr)
	svc.SetEventFirer(func(e sdk.Event) {
		if ev, ok := e.(*sdkentity.EntityMoving); ok {
			ev.Cancel()
		}
	})
	err := svc.Walk(context.Background(), inst, entity, 2, 2)
	assert.Equal(t, domain.ErrAccessDenied, err)
}

// TestEntityService_Dance_Success verifies dance message is processed.
func TestEntityService_Dance_Success(t *testing.T) {
	svc, mgr := newEntityService(t)
	defer mgr.StopAll()
	inst, entity := loadedInstance(t, mgr)
	err := svc.Dance(context.Background(), inst, entity, 2)
	assert.NoError(t, err)
}

// TestEntityService_Action_Success verifies action message is processed.
func TestEntityService_Action_Success(t *testing.T) {
	svc, mgr := newEntityService(t)
	defer mgr.StopAll()
	inst, entity := loadedInstance(t, mgr)
	err := svc.Action(context.Background(), inst, entity, 1)
	assert.NoError(t, err)
}

// TestEntityService_Sign_Success verifies sign message is processed.
func TestEntityService_Sign_Success(t *testing.T) {
	svc, mgr := newEntityService(t)
	defer mgr.StopAll()
	inst, entity := loadedInstance(t, mgr)
	err := svc.Sign(context.Background(), inst, entity, 5)
	assert.NoError(t, err)
}

// TestEntityService_StartTyping_Success verifies typing start is processed.
func TestEntityService_StartTyping_Success(t *testing.T) {
	svc, mgr := newEntityService(t)
	defer mgr.StopAll()
	inst, entity := loadedInstance(t, mgr)
	err := svc.StartTyping(context.Background(), inst, entity)
	assert.NoError(t, err)
}

// TestEntityService_StopTyping_Success verifies typing stop is processed.
func TestEntityService_StopTyping_Success(t *testing.T) {
	svc, mgr := newEntityService(t)
	defer mgr.StopAll()
	inst, entity := loadedInstance(t, mgr)
	_ = svc.StartTyping(context.Background(), inst, entity)
	err := svc.StopTyping(context.Background(), inst, entity)
	assert.NoError(t, err)
}

// TestEntityService_LookTo_Success verifies look-to message is processed.
func TestEntityService_LookTo_Success(t *testing.T) {
	svc, mgr := newEntityService(t)
	defer mgr.StopAll()
	inst, entity := loadedInstance(t, mgr)
	err := svc.LookTo(context.Background(), inst, entity, 2, 2)
	assert.NoError(t, err)
}
