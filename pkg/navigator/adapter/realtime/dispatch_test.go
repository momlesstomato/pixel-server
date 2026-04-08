package realtime

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	navapp "github.com/momlesstomato/pixel-server/pkg/navigator/application"
	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
	"github.com/momlesstomato/pixel-server/pkg/navigator/packet"
)

type runtimeRepoStub struct {
	createdRoom domain.Room
	lastFilter  domain.RoomFilter
}

func (stub *runtimeRepoStub) ListCategories(context.Context) ([]domain.Category, error) {
	return nil, nil
}
func (stub *runtimeRepoStub) FindCategoryByID(context.Context, int) (domain.Category, error) {
	return domain.Category{}, nil
}
func (stub *runtimeRepoStub) CreateCategory(_ context.Context, cat domain.Category) (domain.Category, error) {
	return cat, nil
}
func (stub *runtimeRepoStub) DeleteCategory(context.Context, int) error { return nil }
func (stub *runtimeRepoStub) ListRooms(_ context.Context, filter domain.RoomFilter) ([]domain.Room, int, error) {
	stub.lastFilter = filter
	return nil, 0, nil
}
func (stub *runtimeRepoStub) FindRoomByID(context.Context, int) (domain.Room, error) {
	return domain.Room{}, nil
}
func (stub *runtimeRepoStub) CreateRoom(_ context.Context, room domain.Room) (domain.Room, error) {
	stub.createdRoom = room
	room.ID = 42
	return room, nil
}
func (stub *runtimeRepoStub) UpdateRoom(context.Context, int, domain.RoomPatch) (domain.Room, error) {
	return domain.Room{}, nil
}
func (stub *runtimeRepoStub) DeleteRoom(context.Context, int) error { return nil }
func (stub *runtimeRepoStub) ListSavedSearches(context.Context, int) ([]domain.SavedSearch, error) {
	return nil, nil
}
func (stub *runtimeRepoStub) CreateSavedSearch(_ context.Context, search domain.SavedSearch) (domain.SavedSearch, error) {
	return search, nil
}
func (stub *runtimeRepoStub) DeleteSavedSearch(context.Context, int) error { return nil }
func (stub *runtimeRepoStub) ListFavourites(context.Context, int) ([]domain.Favourite, error) {
	return nil, nil
}
func (stub *runtimeRepoStub) AddFavourite(context.Context, int, int) error      { return nil }
func (stub *runtimeRepoStub) RemoveFavourite(context.Context, int, int) error   { return nil }
func (stub *runtimeRepoStub) CountFavourites(context.Context, int) (int, error) { return 0, nil }

type runtimeSessionStub struct{}

func (runtimeSessionStub) Register(coreconnection.Session) error { return nil }
func (runtimeSessionStub) FindByConnID(connID string) (coreconnection.Session, bool) {
	if connID == "conn1" {
		return coreconnection.Session{ConnID: connID, UserID: 7}, true
	}
	return coreconnection.Session{}, false
}
func (runtimeSessionStub) FindByUserID(int) (coreconnection.Session, bool) {
	return coreconnection.Session{}, false
}
func (runtimeSessionStub) Touch(string) error                         { return nil }
func (runtimeSessionStub) Remove(string)                              {}
func (runtimeSessionStub) ListAll() ([]coreconnection.Session, error) { return nil, nil }

type runtimePacketRecord struct {
	packetID uint16
	body     []byte
}

type runtimeTransportStub struct{ packets []runtimePacketRecord }

func (stub *runtimeTransportStub) Send(_ string, packetID uint16, body []byte) error {
	copyBody := append([]byte(nil), body...)
	stub.packets = append(stub.packets, runtimePacketRecord{packetID: packetID, body: copyBody})
	return nil
}

// TestHandleCreateRoomResolvesOwnerName verifies navigator room creation populates owner metadata from the configured resolver.
func TestHandleCreateRoomResolvesOwnerName(t *testing.T) {
	repo := &runtimeRepoStub{}
	service, _ := navapp.NewService(repo)
	transport := &runtimeTransportStub{}
	runtime, _ := NewRuntime(service, runtimeSessionStub{}, transport, nil)
	runtime.SetUsernameResolver(func(context.Context, int) (string, error) { return "alice", nil })
	w := codec.NewWriter()
	_ = w.WriteString("My Room")
	_ = w.WriteString("Desc")
	_ = w.WriteString("model_a")
	w.WriteInt32(1)
	w.WriteInt32(25)
	w.WriteInt32(0)
	w.WriteInt32(0)
	handled, err := runtime.Handle(context.Background(), "conn1", packet.CreateRoomPacketID, w.Bytes())
	if err != nil || !handled {
		t.Fatalf("expected handled create room, got handled=%v err=%v", handled, err)
	}
	if repo.createdRoom.OwnerName != "alice" {
		t.Fatalf("expected owner name alice, got %q", repo.createdRoom.OwnerName)
	}
	if len(transport.packets) != 1 || transport.packets[0].packetID != packet.RoomCreatedPacketID {
		t.Fatalf("expected room created response, got %+v", transport.packets)
	}
}

// TestHandleSearchMyWorldUsesOwnerFilter verifies myworld searches constrain results to the authenticated user.
func TestHandleSearchMyWorldUsesOwnerFilter(t *testing.T) {
	repo := &runtimeRepoStub{}
	service, _ := navapp.NewService(repo)
	runtime, _ := NewRuntime(service, runtimeSessionStub{}, &runtimeTransportStub{}, nil)
	w := codec.NewWriter()
	_ = w.WriteString("myworld_view")
	_ = w.WriteString("")
	handled, err := runtime.Handle(context.Background(), "conn1", packet.SearchRoomsPacketID, w.Bytes())
	if err != nil || !handled {
		t.Fatalf("expected handled search, got handled=%v err=%v", handled, err)
	}
	if repo.lastFilter.OwnerID == nil || *repo.lastFilter.OwnerID != 7 {
		t.Fatalf("expected owner filter 7, got %+v", repo.lastFilter.OwnerID)
	}
}

// TestHandleMyRoomsUsesLegacyPrivateRoomsPacket verifies Nitro my-rooms requests use the legacy private-room result payload.
func TestHandleMyRoomsUsesLegacyPrivateRoomsPacket(t *testing.T) {
	repo := &runtimeRepoStub{}
	service, _ := navapp.NewService(repo)
	transport := &runtimeTransportStub{}
	runtime, _ := NewRuntime(service, runtimeSessionStub{}, transport, nil)
	handled, err := runtime.Handle(context.Background(), "conn1", packet.MyRoomsSearchPacketID, nil)
	if err != nil || !handled {
		t.Fatalf("expected handled my rooms search, got handled=%v err=%v", handled, err)
	}
	if repo.lastFilter.OwnerID == nil || *repo.lastFilter.OwnerID != 7 {
		t.Fatalf("expected owner filter 7, got %+v", repo.lastFilter.OwnerID)
	}
	if len(transport.packets) != 1 || transport.packets[0].packetID != packet.GuestRoomSearchResultPacketID {
		t.Fatalf("expected guest room search result response, got %+v", transport.packets)
	}
	r := codec.NewReader(transport.packets[0].body)
	searchType, readErr := r.ReadInt32()
	if readErr != nil {
		t.Fatalf("read search type: %v", readErr)
	}
	searchParam, readErr := r.ReadString()
	if readErr != nil {
		t.Fatalf("read search param: %v", readErr)
	}
	roomCount, readErr := r.ReadInt32()
	if readErr != nil {
		t.Fatalf("read room count: %v", readErr)
	}
	if searchType != 2 || searchParam != "" || roomCount != 0 {
		t.Fatalf("unexpected body values type=%d param=%q rooms=%d", searchType, searchParam, roomCount)
	}
}
