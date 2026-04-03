package tests

import (
	"context"
	"testing"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestManager_LoadGet verifies room load and retrieval.
func TestManager_LoadGet(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	defer mgr.StopAll()
	layout := testLayout()
	inst := mgr.Load(42, layout)
	require.NotNil(t, inst)
	assert.Equal(t, 42, inst.RoomID)
	got, ok := mgr.Get(42)
	require.True(t, ok)
	assert.Equal(t, inst, got)
}

// TestManager_LoadIdempotent verifies loading same room returns same instance.
func TestManager_LoadIdempotent(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	defer mgr.StopAll()
	layout := testLayout()
	inst1 := mgr.Load(1, layout)
	inst2 := mgr.Load(1, layout)
	assert.Equal(t, inst1, inst2)
}

// TestManager_Unload verifies room removal.
func TestManager_Unload(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	mgr.Load(1, testLayout())
	assert.Equal(t, 1, mgr.Count())
	mgr.Unload(1)
	assert.Equal(t, 0, mgr.Count())
	_, ok := mgr.Get(1)
	assert.False(t, ok)
}

// TestManager_Count verifies loaded room counting.
func TestManager_Count(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	defer mgr.StopAll()
	mgr.Load(1, testLayout())
	mgr.Load(2, testLayout())
	assert.Equal(t, 2, mgr.Count())
}

// TestManager_StopAll verifies all rooms shutdown.
func TestManager_StopAll(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	inst := mgr.Load(1, testLayout())
	mgr.StopAll()
	select {
	case <-inst.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("stop all did not complete in time")
	}
	assert.Equal(t, 0, mgr.Count())
}

// TestManager_GetMissing verifies missing room returns false.
func TestManager_GetMissing(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	_, ok := mgr.Get(999)
	assert.False(t, ok)
}

// TestManager_Cleanup verifies stopped rooms are removed.
func TestManager_Cleanup(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	inst := mgr.Load(1, testLayout())
	inst.Stop()
	<-inst.Done()
	time.Sleep(50 * time.Millisecond)
	removed := mgr.Cleanup()
	assert.Equal(t, 1, removed)
	assert.Equal(t, 0, mgr.Count())
}
