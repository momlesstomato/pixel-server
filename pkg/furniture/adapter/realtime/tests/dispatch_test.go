package tests

import (
	"context"
	"errors"
	"testing"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/furniture/adapter/realtime"
	furnitureapplication "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	furnituredomain "github.com/momlesstomato/pixel-server/pkg/furniture/domain"
	furnipacket "github.com/momlesstomato/pixel-server/pkg/furniture/packet"
)

// repoStub provides deterministic furniture repository behavior.
type repoStub struct {
	items  []furnituredomain.Item
	def    furnituredomain.Definition
	defErr error
}

// FindDefinitionByID returns the stub definition or an error.
func (r repoStub) FindDefinitionByID(_ context.Context, _ int) (furnituredomain.Definition, error) {
	return r.def, r.defErr
}

// FindDefinitionByName returns an empty definition.
func (r repoStub) FindDefinitionByName(_ context.Context, _ string) (furnituredomain.Definition, error) {
	return furnituredomain.Definition{}, nil
}

// ListDefinitions returns nil.
func (r repoStub) ListDefinitions(_ context.Context) ([]furnituredomain.Definition, error) {
	return nil, nil
}

// CreateDefinition returns the definition unchanged.
func (r repoStub) CreateDefinition(_ context.Context, d furnituredomain.Definition) (furnituredomain.Definition, error) {
	return d, nil
}

// UpdateDefinition returns an empty definition.
func (r repoStub) UpdateDefinition(_ context.Context, _ int, _ furnituredomain.DefinitionPatch) (furnituredomain.Definition, error) {
	return furnituredomain.Definition{}, nil
}

// DeleteDefinition returns nil.
func (r repoStub) DeleteDefinition(_ context.Context, _ int) error { return nil }

// FindItemByID returns an empty item.
func (r repoStub) FindItemByID(_ context.Context, _ int) (furnituredomain.Item, error) {
	return furnituredomain.Item{}, nil
}

// ListItemsByUserID returns the stub items.
func (r repoStub) ListItemsByUserID(_ context.Context, _ int) ([]furnituredomain.Item, error) {
	return r.items, nil
}

// CreateItem returns the item unchanged.
func (r repoStub) CreateItem(_ context.Context, i furnituredomain.Item) (furnituredomain.Item, error) {
	return i, nil
}

// DeleteItem returns nil.
func (r repoStub) DeleteItem(_ context.Context, _ int) error { return nil }

// TransferItem returns nil.
func (r repoStub) TransferItem(_ context.Context, _ int, _ int) error { return nil }

// CountItemsByUserID returns zero.
func (r repoStub) CountItemsByUserID(_ context.Context, _ int) (int, error) { return 0, nil }

// PlaceItem returns nil (no-op placement for tests).
func (r repoStub) PlaceItem(_ context.Context, _ int, _ int, _ int, _ int, _ float64, _ int) error {
	return nil
}

// ListItemsByRoomID returns deterministic items placed in a room.
func (r repoStub) ListItemsByRoomID(_ context.Context, _ int) ([]furnituredomain.Item, error) {
	return r.items, nil
}

// sessionStub provides deterministic authenticated session lookup.
type sessionStub struct{}

// Register returns nil.
func (sessionStub) Register(coreconnection.Session) error { return nil }

// FindByConnID recognises only "conn1".
func (sessionStub) FindByConnID(id string) (coreconnection.Session, bool) {
	if id == "conn1" {
		return coreconnection.Session{UserID: 1}, true
	}
	return coreconnection.Session{}, false
}

// FindByUserID always returns not found.
func (sessionStub) FindByUserID(int) (coreconnection.Session, bool) {
	return coreconnection.Session{}, false
}

// Touch returns nil.
func (sessionStub) Touch(string) error { return nil }

// Remove is a no-op.
func (sessionStub) Remove(string) {}

// ListAll returns nil.
func (sessionStub) ListAll() ([]coreconnection.Session, error) { return nil, nil }

// transportStub captures sent packet identifiers.
type transportStub struct{ sent []uint16 }

// Send records the packet identifier.
func (t *transportStub) Send(_ string, packetID uint16, _ []byte) error {
	t.sent = append(t.sent, packetID)
	return nil
}

// buildRuntime creates a furniture realtime runtime backed by the given stub.
func buildRuntime(repo furnituredomain.Repository) (*realtime.Runtime, *transportStub) {
	svc, _ := furnitureapplication.NewService(repo)
	transport := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	return rt, transport
}

// TestHandleGetFurnitureSendsFurniList verifies 3150 triggers a 994 furni list response.
func TestHandleGetFurnitureSendsFurniList(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 0, DefinitionID: 5}
	def := furnituredomain.Definition{ID: 5, ItemType: furnituredomain.ItemTypeFloor, SpriteID: 100, AllowRecycle: true, AllowTrade: true}
	rt, transport := buildRuntime(repoStub{items: []furnituredomain.Item{item}, def: def})
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.GetFurniturePacketID, nil)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != furnipacket.FurniListPacketID {
		t.Fatalf("expected furni_list packet %d, got %v", furnipacket.FurniListPacketID, transport.sent)
	}
}

// TestHandleGetFurnitureExcludesPlacedItems verifies items placed in rooms are not sent.
func TestHandleGetFurnitureExcludesPlacedItems(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 99, DefinitionID: 5}
	def := furnituredomain.Definition{ID: 5, ItemType: furnituredomain.ItemTypeFloor, SpriteID: 100}
	rt, transport := buildRuntime(repoStub{items: []furnituredomain.Item{item}, def: def})
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.GetFurniturePacketID, nil)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != furnipacket.FurniListPacketID {
		t.Fatalf("expected exactly one packet 994, got %v", transport.sent)
	}
}

// TestHandleGetFurnitureEmptyInventory verifies an empty inventory sends 994 with no items.
func TestHandleGetFurnitureEmptyInventory(t *testing.T) {
	rt, transport := buildRuntime(repoStub{})
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.GetFurniturePacketID, nil)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != furnipacket.FurniListPacketID {
		t.Fatalf("expected furni_list packet %d, got %v", furnipacket.FurniListPacketID, transport.sent)
	}
}

// TestHandleGetFurnitureSkipsItemsWithMissingDefinition verifies definition lookup failures skip the item silently.
func TestHandleGetFurnitureSkipsItemsWithMissingDefinition(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 0, DefinitionID: 99}
	rt, transport := buildRuntime(repoStub{items: []furnituredomain.Item{item}, defErr: errors.New("not found")})
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.GetFurniturePacketID, nil)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != furnipacket.FurniListPacketID {
		t.Fatalf("expected furni_list packet %d, got %v", furnipacket.FurniListPacketID, transport.sent)
	}
}

// TestHandleUnknownPacketNotHandled verifies unknown packet IDs return handled=false.
func TestHandleUnknownPacketNotHandled(t *testing.T) {
	rt, transport := buildRuntime(repoStub{})
	handled, err := rt.Handle(context.Background(), "conn1", 9999, nil)
	if err != nil || handled {
		t.Fatalf("expected not handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 0 {
		t.Fatalf("expected no packets sent, got %v", transport.sent)
	}
}
