package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/navigator/application"
	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
	"github.com/momlesstomato/pixel-server/pkg/navigator/packet"
	"go.uber.org/zap"
)

// Handle dispatches one authenticated navigator packet payload.
func (runtime *Runtime) Handle(ctx context.Context, connID string, packetID uint16, body []byte) (bool, error) {
	userID, ok := runtime.userID(connID)
	if !ok {
		return false, nil
	}
	switch packetID {
	case packet.InitNavigatorPacketID:
		return true, runtime.handleInit(ctx, connID, userID)
	case packet.SearchRoomsPacketID:
		return true, runtime.handleSearch(ctx, connID, userID, body)
	case packet.MyRoomsSearchPacketID:
		return true, runtime.handleMyRooms(ctx, connID, userID)
	case packet.GetGuestRoomPacketID:
		return true, runtime.handleGetGuestRoom(ctx, connID, body)
	case packet.GetFlatCategoriesPacketID:
		return true, runtime.handleGetFlatCategories(ctx, connID)
	case packet.CanCreateRoomPacketID:
		return true, runtime.handleCanCreateRoom(connID)
	case packet.CreateRoomPacketID:
		return true, runtime.handleCreateRoom(ctx, connID, userID, body)
	case packet.AddFavouritePacketID:
		return true, runtime.handleAddFavourite(ctx, connID, userID, body)
	case packet.RemoveFavouritePacketID:
		return true, runtime.handleRemoveFavourite(ctx, connID, userID, body)
	case packet.SaveSearchPacketID:
		return true, runtime.handleSaveSearch(ctx, connID, userID, body)
	case packet.DeleteSearchPacketID:
		return true, runtime.handleDeleteSearch(ctx, connID, userID, body)
	case packet.GetUserEventCatsPacketID:
		return true, runtime.handleGetEventCats(connID)
	case packet.PurchaseRoomAdPacketID:
		return true, runtime.handlePurchaseRoomAd(ctx, connID, userID, body)
	case packet.StaffPickPacketID:
		return true, runtime.handleStaffPick(ctx, connID, userID, body)
	case packet.SaveSettingsPacketID:
		return true, nil
	default:
		return false, nil
	}
}

// handleMyRooms responds with the caller's owned rooms using the legacy private-room result packet.
func (runtime *Runtime) handleMyRooms(ctx context.Context, connID string, userID int) error {
	rooms, err := runtime.listRooms(ctx, domain.RoomFilter{OwnerID: &userID, Limit: 50})
	if err != nil {
		runtime.logger.Error("list my rooms failed", zap.Int("user_id", userID), zap.Error(err))
		return nil
	}
	return runtime.sendPacket(connID, packet.GuestRoomSearchResultPacket{
		SearchType:  2,
		SearchParam: "",
		Rooms:       rooms,
	})
}

// handleInit sends navigator initialization data.
func (runtime *Runtime) handleInit(ctx context.Context, connID string, userID int) error {
	tabs := []string{"official_view", "hotel_view", "roomads_view", "myworld_view"}
	if err := runtime.sendPacket(connID, packet.NavigatorMetaDataPacket{TopLevelContexts: tabs}); err != nil {
		return err
	}
	if err := runtime.sendPacket(connID, packet.NavigatorCollapsedPacket{}); err != nil {
		return err
	}
	if err := runtime.sendPacket(connID, packet.NavigatorSettingsPacket{Width: 425, Height: 535}); err != nil {
		return err
	}
	searches, err := runtime.service.ListSavedSearches(ctx, userID)
	if err != nil {
		runtime.logger.Error("init saved searches failed", zap.Int("user_id", userID), zap.Error(err))
		return runtime.sendPacket(connID, packet.NavigatorSavedSearchesPacket{})
	}
	entries := make([]packet.SavedSearchEntry, len(searches))
	for i, s := range searches {
		entries[i] = packet.SavedSearchEntry{ID: int32(s.ID), SearchCode: s.SearchCode, Filter: s.Filter}
	}
	return runtime.sendPacket(connID, packet.NavigatorSavedSearchesPacket{Searches: entries})
}

// handleSearch responds with room search results.
func (runtime *Runtime) handleSearch(ctx context.Context, connID string, userID int, body []byte) error {
	searchCode, filter := parseSearchParams(body)
	roomFilter := domain.RoomFilter{SearchQuery: filter, Limit: 50}
	switch searchCode {
	case "myworld_view":
		roomFilter.OwnerID = &userID
	case "roomads_view":
		roomFilter.PromotedOnly = true
	case "official_view":
		roomFilter.StaffPickOnly = true
	}
	rooms, err := runtime.listRooms(ctx, roomFilter)
	if err != nil {
		runtime.logger.Error("search rooms failed", zap.String("code", searchCode), zap.Error(err))
		return nil
	}
	block := packet.SearchResultBlock{SearchCode: searchCode, Text: searchCode, Rooms: rooms}
	return runtime.sendPacket(connID, packet.NavigatorSearchResultsPacket{
		SearchCode: searchCode, Filter: filter, Results: []packet.SearchResultBlock{block},
	})
}

// listRooms fetches rooms and overlays live occupant counts when available.
func (runtime *Runtime) listRooms(ctx context.Context, filter domain.RoomFilter) ([]domain.Room, error) {
	rooms, _, err := runtime.service.ListRooms(ctx, filter)
	if err != nil {
		return nil, err
	}
	if runtime.liveRoomCount != nil {
		for i := range rooms {
			rooms[i].CurrentUsers = runtime.liveRoomCount(rooms[i].ID)
		}
	}
	return rooms, nil
}

// handleGetEventCats responds with an empty event categories list.
func (runtime *Runtime) handleGetEventCats(connID string) error {
	return runtime.sendPacket(connID, packet.NavigatorEventCategoriesPacket{})
}

// handleGetFlatCategories responds with room category list.
func (runtime *Runtime) handleGetFlatCategories(ctx context.Context, connID string) error {
	cats, err := runtime.service.ListCategories(ctx)
	if err != nil {
		runtime.logger.Error("get flat categories failed", zap.Error(err))
		return nil
	}
	return runtime.sendPacket(connID, packet.FlatCategoriesPacket{Categories: cats})
}

// handleCanCreateRoom responds with room creation eligibility.
func (runtime *Runtime) handleCanCreateRoom(connID string) error {
	return runtime.sendPacket(connID, packet.CanCreateRoomResponsePacket{ResultCode: 0, MaxRooms: application.MaxRoomsPerm})
}

// handleSaveSearch persists one saved search entry.
func (runtime *Runtime) handleSaveSearch(ctx context.Context, connID string, userID int, body []byte) error {
	searchCode, filter := parseSearchParams(body)
	_, err := runtime.service.CreateSavedSearch(ctx, domain.SavedSearch{UserID: userID, SearchCode: searchCode, Filter: filter})
	if err != nil {
		runtime.logger.Error("save search failed", zap.Int("user_id", userID), zap.Error(err))
		return nil
	}
	return runtime.sendSavedSearches(ctx, connID, userID)
}

// handleDeleteSearch removes one saved search entry.
func (runtime *Runtime) handleDeleteSearch(ctx context.Context, connID string, userID int, body []byte) error {
	id := parseRoomID(body)
	if id <= 0 {
		return nil
	}
	if err := runtime.service.DeleteSavedSearch(ctx, id); err != nil {
		runtime.logger.Error("delete search failed", zap.Int("user_id", userID), zap.Error(err))
	}
	return runtime.sendSavedSearches(ctx, connID, userID)
}

// sendSavedSearches sends updated saved searches list to one connection.
func (runtime *Runtime) sendSavedSearches(ctx context.Context, connID string, userID int) error {
	searches, err := runtime.service.ListSavedSearches(ctx, userID)
	if err != nil {
		return nil
	}
	entries := make([]packet.SavedSearchEntry, len(searches))
	for i, s := range searches {
		entries[i] = packet.SavedSearchEntry{ID: int32(s.ID), SearchCode: s.SearchCode, Filter: s.Filter}
	}
	return runtime.sendPacket(connID, packet.NavigatorSavedSearchesPacket{Searches: entries})
}

// parseSearchParams reads search code and filter from packet body.
func parseSearchParams(body []byte) (string, string) {
	reader := codec.NewReader(body)
	code, _ := reader.ReadString()
	filter, _ := reader.ReadString()
	return code, filter
}

// parseRoomID reads one int32 room or search ID from packet body.
func parseRoomID(body []byte) int {
	reader := codec.NewReader(body)
	id, err := reader.ReadInt32()
	if err != nil {
		return 0
	}
	return int(id)
}
