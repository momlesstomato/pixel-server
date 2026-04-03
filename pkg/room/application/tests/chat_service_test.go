package tests

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkroomchat "github.com/momlesstomato/pixel-sdk/events/room/chat"
	"github.com/momlesstomato/pixel-server/pkg/room/application"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newChatService(t *testing.T) *application.ChatService {
	t.Helper()
	svc, err := application.NewChatService(zap.NewNop())
	require.NoError(t, err)
	return svc
}

func newInstanceWithTwoEntities(t *testing.T, mgr *engine.Manager) (*engine.Instance, *domain.RoomEntity, *domain.RoomEntity) {
	t.Helper()
	inst := mgr.Load(2, domain.Layout{Slug: "two", DoorX: 0, DoorY: 0, Grid: must3x3Grid()})
	sender := domain.NewPlayerEntity(0, 1, "c1", "Alice", "", "", "M", domain.Tile{X: 1, Y: 1, State: domain.TileOpen})
	target := domain.NewPlayerEntity(1, 2, "c2", "Bob", "", "", "M", domain.Tile{X: 1, Y: 2, State: domain.TileOpen})
	r1 := make(chan error, 1)
	r2 := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: &sender, Reply: r1})
	require.NoError(t, <-r1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: &target, Reply: r2})
	require.NoError(t, <-r2)
	return inst, &sender, &target
}

// TestNewChatService_NilLogger creates service with nil logger successfully.
func TestNewChatService_NilLogger(t *testing.T) {
	svc, err := application.NewChatService(nil)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

// TestChatService_Talk_ReturnsRecipients verifies talk returns nearby entities.
func TestChatService_Talk_ReturnsRecipients(t *testing.T) {
	svc := newChatService(t)
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	defer mgr.StopAll()
	inst, sender, _ := newInstanceWithTwoEntities(t, mgr)
	recipients, err := svc.Talk(context.Background(), inst, sender, 2, "hello", 0)
	require.NoError(t, err)
	assert.NotEmpty(t, recipients)
}

// TestChatService_Talk_FiresEvents verifies ChatSending and ChatSent are fired.
func TestChatService_Talk_FiresEvents(t *testing.T) {
	svc := newChatService(t)
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	defer mgr.StopAll()
	inst, sender, _ := newInstanceWithTwoEntities(t, mgr)
	var fired []string
	svc.SetEventFirer(func(e sdk.Event) {
		switch e.(type) {
		case *sdkroomchat.ChatSending:
			fired = append(fired, "sending")
		case *sdkroomchat.ChatSent:
			fired = append(fired, "sent")
		}
	})
	_, err := svc.Talk(context.Background(), inst, sender, 2, "hi", 0)
	require.NoError(t, err)
	assert.Contains(t, fired, "sending")
	assert.Contains(t, fired, "sent")
}

// TestChatService_Talk_Cancelled verifies cancelled ChatSending aborts chat.
func TestChatService_Talk_Cancelled(t *testing.T) {
	svc := newChatService(t)
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	defer mgr.StopAll()
	inst, sender, _ := newInstanceWithTwoEntities(t, mgr)
	svc.SetEventFirer(func(e sdk.Event) {
		if ev, ok := e.(*sdkroomchat.ChatSending); ok {
			ev.Cancel()
		}
	})
	_, err := svc.Talk(context.Background(), inst, sender, 2, "blocked", 0)
	assert.Equal(t, domain.ErrAccessDenied, err)
}

// TestChatService_Talk_FloodControl verifies mute activates after flood limit.
func TestChatService_Talk_FloodControl(t *testing.T) {
	svc := newChatService(t)
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	defer mgr.StopAll()
	inst, sender, _ := newInstanceWithTwoEntities(t, mgr)
	for i := 0; i < 3; i++ {
		_, _ = svc.Talk(context.Background(), inst, sender, 2, "msg", 0)
	}
	_, err := svc.Talk(context.Background(), inst, sender, 2, "flood", 0)
	assert.Equal(t, domain.ErrFloodControl, err)
}

// TestChatService_Shout_AllEntities verifies shout returns all room entities.
func TestChatService_Shout_AllEntities(t *testing.T) {
	svc := newChatService(t)
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	defer mgr.StopAll()
	inst, sender, _ := newInstanceWithTwoEntities(t, mgr)
	recipients, err := svc.Shout(context.Background(), inst, sender, 2, "hey", 0)
	require.NoError(t, err)
	assert.Len(t, recipients, 2)
}

// TestChatService_Whisper_TwoRecipients verifies whisper returns sender and target.
func TestChatService_Whisper_TwoRecipients(t *testing.T) {
	svc := newChatService(t)
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	defer mgr.StopAll()
	_, sender, target := newInstanceWithTwoEntities(t, mgr)
	recipients, err := svc.Whisper(context.Background(), sender, 2, target, "secret", 0)
	require.NoError(t, err)
	assert.Len(t, recipients, 2)
}

// TestChatService_Whisper_Cancelled verifies cancelled ChatSending aborts whisper.
func TestChatService_Whisper_Cancelled(t *testing.T) {
	svc := newChatService(t)
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	defer mgr.StopAll()
	_, sender, target := newInstanceWithTwoEntities(t, mgr)
	svc.SetEventFirer(func(e sdk.Event) {
		if ev, ok := e.(*sdkroomchat.ChatSending); ok {
			ev.Cancel()
		}
	})
	_, err := svc.Whisper(context.Background(), sender, 2, target, "blocked", 0)
	assert.Equal(t, domain.ErrAccessDenied, err)
}
