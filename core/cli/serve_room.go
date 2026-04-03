package cli

import (
	"context"

	"github.com/momlesstomato/pixel-server/core/initializer"
	roomapplication "github.com/momlesstomato/pixel-server/pkg/room/application"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	roomstore "github.com/momlesstomato/pixel-server/pkg/room/infrastructure/store"
)

// buildRoomServices constructs room-realm application services.
func buildRoomServices(runtime *initializer.Runtime, broadcaster engine.EntityBroadcaster) (*roomapplication.Service, error) {
	modelRepo, err := roomstore.NewModelStore(runtime.PostgreSQL)
	if err != nil {
		return nil, err
	}
	banRepo, err := roomstore.NewBanStore(runtime.PostgreSQL)
	if err != nil {
		return nil, err
	}
	rightsRepo, err := roomstore.NewRightsStore(runtime.PostgreSQL)
	if err != nil {
		return nil, err
	}
	mgr := engine.NewManager(context.Background(), runtime.Logger, broadcaster)
	return roomapplication.NewService(modelRepo, banRepo, rightsRepo, mgr, runtime.Logger)
}

// noopEntityBroadcaster is a default no-op broadcaster used before transport wiring.
func noopEntityBroadcaster(_ int, _ []domain.RoomEntity, _ []byte) {}
