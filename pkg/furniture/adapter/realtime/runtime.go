package realtime

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	"github.com/momlesstomato/pixel-server/core/codec"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	furnitureapp "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	furnituredomain "github.com/momlesstomato/pixel-server/pkg/furniture/domain"
	furnipacket "github.com/momlesstomato/pixel-server/pkg/furniture/packet"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
	"go.uber.org/zap"
)

// Transport defines packet write behavior required by furniture realtime runtime.
type Transport interface {
	// Send writes one encoded packet payload to one connection identifier.
	Send(string, uint16, []byte) error
}

// UsernameResolver resolves a display name for one authenticated user identifier.
type UsernameResolver func(ctx context.Context, userID int) (string, error)

// RoomAccessChecker reports whether a user may manage furniture in a room.
type RoomAccessChecker func(ctx context.Context, roomID, userID int) bool

// RoomOccupancyChecker reports whether a tile is currently occupied by a player entity.
type RoomOccupancyChecker func(roomID, x, y int) bool

// RoomEntityTileResolver resolves the active room tile for one connection.
type RoomEntityTileResolver func(connID string) (roomID, x, y int, ok bool)

// RoomTileSnapshot stores the layout-facing tile data required by roller safety checks.
type RoomTileSnapshot struct {
	// Z stores the static base tile height.
	Z float64
	// Walkable reports whether the layout allows movement onto the tile.
	Walkable bool
}

// RoomTileResolver resolves one room layout tile snapshot.
type RoomTileResolver func(roomID, x, y int) (RoomTileSnapshot, bool)

// RoomEntitySnapshot stores the room-facing entity data required by furniture interactions.
type RoomEntitySnapshot struct {
	// ConnID stores the owning connection identifier.
	ConnID string
	// VirtualID stores the room-scoped entity identifier.
	VirtualID int
	// UserID stores the backing user identifier.
	UserID int
	// X stores the current tile horizontal coordinate.
	X int
	// Y stores the current tile vertical coordinate.
	Y int
	// Z stores the current tile height.
	Z float64
	// Dir stores the current body rotation.
	Dir int
	// IsWalking reports whether the entity is currently following a walk path.
	IsWalking bool
}

// RoomEntitySnapshotter resolves the current entity list for one room.
type RoomEntitySnapshotter func(roomID int) []RoomEntitySnapshot

// RoomEntityWalker requests one entity walk toward a destination tile.
type RoomEntityWalker func(ctx context.Context, connID string, x, y int) error

// RoomEntityWarper requests one direct entity relocation in a room.
type RoomEntityWarper func(ctx context.Context, roomID, virtualID, x, y int, z float64, dir int, silent bool, animate bool) error

// TeleporterForwarder queues one cross-room teleporter arrival and forwards the client.
type TeleporterForwarder func(ctx context.Context, connID string, roomID, spawnX, spawnY int, spawnZ float64, spawnDir, exitX, exitY int) error

// seatEntry caches the seat properties of one placed sittable furniture item.
type seatEntry struct {
	// itemID stores the placed item identifier for cache invalidation.
	itemID int
	// x stores the tile horizontal coordinate.
	x int
	// y stores the tile vertical coordinate.
	y int
	// anchorX stores the canonical tile horizontal coordinate for seat or lay targeting.
	anchorX int
	// anchorY stores the canonical tile vertical coordinate for seat or lay targeting.
	anchorY int
	// height stores the furniture stack height used as the sit value.
	height float64
	// dir stores the furniture rotation direction (0-7).
	dir int
	// canSit reports whether avatars can sit on this item.
	canSit bool
	// canLay reports whether avatars can lay on this item.
	canLay bool
}

// RoomEntityRotator updates seated entities at a tile to face a new furniture direction.
type RoomEntityRotator func(roomID, x, y, dir int)

// RoomEntityEvictor clears sit/lay state for entities at a tile and leaves them standing in place.
type RoomEntityEvictor func(roomID, x, y int)

// Runtime defines furniture realm websocket packet behavior.
type Runtime struct {
	// service stores furniture application behavior.
	service *furnitureapp.Service
	// sessions stores authenticated connection lookup behavior.
	sessions coreconnection.SessionRegistry
	// transport stores packet write behavior.
	transport Transport
	// broadcaster publishes owner-targeted inventory updates to active user sessions.
	broadcaster broadcast.Broadcaster
	// logger stores runtime logging behavior.
	logger *zap.Logger
	// roomFinder resolves the room identifier for a given connection.
	roomFinder func(connID string) (int, bool)
	// roomBroadcaster sends an encoded payload to all players in a room.
	roomBroadcaster func(roomID int, packetID uint16, body []byte)
	// usernameResolver resolves display names for item owner identifiers.
	usernameResolver UsernameResolver
	// roomAccessChecker validates whether a user may manage furniture in a room.
	roomAccessChecker RoomAccessChecker
	// roomOccupancyChecker validates whether a target tile is already occupied by a player.
	roomOccupancyChecker RoomOccupancyChecker
	// roomEntityTileResolver resolves the current tile for one active room connection.
	roomEntityTileResolver RoomEntityTileResolver
	// roomTileResolver resolves one layout tile snapshot for roller movement validation.
	roomTileResolver RoomTileResolver
	// entityRotator rotates seated entities on a tile to match new furniture direction.
	entityRotator RoomEntityRotator
	// entityEvictor clears seated entities from a tile when furniture is moved or removed.
	entityEvictor RoomEntityEvictor
	// roomEntitySnapshotter resolves room entity snapshots for roller processing.
	roomEntitySnapshotter RoomEntitySnapshotter
	// roomEntityWalker requests avatar walks used by teleporter interactions.
	roomEntityWalker RoomEntityWalker
	// roomEntityWarper requests direct avatar relocation used by rollers and teleporters.
	roomEntityWarper RoomEntityWarper
	// teleporterForwarder forwards avatars into destination rooms through teleporters.
	teleporterForwarder TeleporterForwarder
	// diceRandomizer resolves one zero-based dice result.
	diceRandomizer func(int) int
	// diceRollDelay stores the rolling animation duration.
	diceRollDelay time.Duration
	// diceMu protects diceCancels.
	diceMu sync.Mutex
	// diceCancels stores cancellation callbacks for in-flight dice rolls.
	diceCancels map[int]context.CancelFunc
	// seatCache maps room identifier to its list of sittable item entries.
	seatCache map[int][]seatEntry
	// blockCache maps room identifier to blocked floor-item footprint tiles.
	blockCache map[int][]blockEntry
	// seatMu protects seatCache from concurrent access.
	seatMu sync.RWMutex
}

// NewRuntime creates one furniture realtime runtime instance.
func NewRuntime(service *furnitureapp.Service, sessions coreconnection.SessionRegistry, transport Transport, logger *zap.Logger) (*Runtime, error) {
	if service == nil {
		return nil, fmt.Errorf("furniture service is required")
	}
	if sessions == nil {
		return nil, fmt.Errorf("session registry is required")
	}
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Runtime{
		service: service, sessions: sessions, transport: transport,
		logger: logger, seatCache: make(map[int][]seatEntry), blockCache: make(map[int][]blockEntry),
		diceRandomizer: rand.Intn, diceRollDelay: 2500 * time.Millisecond,
		diceCancels: make(map[int]context.CancelFunc),
	}, nil
}

// SetBroadcaster configures the broadcaster used for per-user inventory updates.
func (runtime *Runtime) SetBroadcaster(value broadcast.Broadcaster) {
	runtime.broadcaster = value
}

// userID resolves authenticated user identifier for one connection.
func (runtime *Runtime) userID(connID string) (int, bool) {
	session, found := runtime.sessions.FindByConnID(connID)
	if !found {
		return 0, false
	}
	return session.UserID, true
}

// Dispose releases per-connection resources.
func (runtime *Runtime) Dispose(_ string) {}

// SetRoomFinder configures the function used to resolve room membership for a connection.
func (runtime *Runtime) SetRoomFinder(fn func(connID string) (int, bool)) {
	runtime.roomFinder = fn
}

// SetRoomBroadcaster configures the function used to broadcast packets to a room.
func (runtime *Runtime) SetRoomBroadcaster(fn func(roomID int, packetID uint16, body []byte)) {
	runtime.roomBroadcaster = fn
}

// SetUsernameResolver configures the display name lookup function for item owners.
func (runtime *Runtime) SetUsernameResolver(fn UsernameResolver) {
	runtime.usernameResolver = fn
}

// SetRoomAccessChecker configures the room-level furniture permission check.
func (runtime *Runtime) SetRoomAccessChecker(fn RoomAccessChecker) {
	runtime.roomAccessChecker = fn
}

// SetRoomOccupancyChecker configures the room-level tile occupancy check.
func (runtime *Runtime) SetRoomOccupancyChecker(fn RoomOccupancyChecker) {
	runtime.roomOccupancyChecker = fn
}

// SetRoomEntityTileResolver configures the callback that resolves one player's current room tile.
func (runtime *Runtime) SetRoomEntityTileResolver(fn RoomEntityTileResolver) {
	runtime.roomEntityTileResolver = fn
}

// SetRoomTileResolver configures the callback that resolves one room layout tile snapshot.
func (runtime *Runtime) SetRoomTileResolver(fn RoomTileResolver) {
	runtime.roomTileResolver = fn
}

// SetRoomEntityRotator configures the callback that rotates seated entities when furniture rotates.
func (runtime *Runtime) SetRoomEntityRotator(fn RoomEntityRotator) {
	runtime.entityRotator = fn
}

// SetRoomEntityEvictor configures the callback that clears seated entities when furniture is moved or removed.
func (runtime *Runtime) SetRoomEntityEvictor(fn RoomEntityEvictor) {
	runtime.entityEvictor = fn
}

// SetRoomEntitySnapshotter configures the callback that resolves room entity snapshots.
func (runtime *Runtime) SetRoomEntitySnapshotter(fn RoomEntitySnapshotter) {
	runtime.roomEntitySnapshotter = fn
}

// SetRoomEntityWalker configures the callback that walks avatars to a tile.
func (runtime *Runtime) SetRoomEntityWalker(fn RoomEntityWalker) {
	runtime.roomEntityWalker = fn
}

// SetRoomEntityWarper configures the callback that relocates avatars in-place.
func (runtime *Runtime) SetRoomEntityWarper(fn RoomEntityWarper) {
	runtime.roomEntityWarper = fn
}

// SetTeleporterForwarder configures the callback that forwards avatars into destination rooms.
func (runtime *Runtime) SetTeleporterForwarder(fn TeleporterForwarder) {
	runtime.teleporterForwarder = fn
}

// SetDiceRandomizer configures the zero-based dice result generator used by delayed rolls.
func (runtime *Runtime) SetDiceRandomizer(fn func(int) int) {
	runtime.diceRandomizer = fn
}

// SetDiceRollDelay configures the rolling animation duration.
func (runtime *Runtime) SetDiceRollDelay(delay time.Duration) {
	runtime.diceRollDelay = delay
}

// sendPacket encodes and transmits one outgoing packet.
func (runtime *Runtime) sendPacket(connID string, pkt interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) error {
	body, err := pkt.Encode()
	if err != nil {
		return err
	}
	return runtime.transport.Send(connID, pkt.PacketID(), body)
}

// sendUserPacket publishes one packet to the owner's broadcast channel when available.
func (runtime *Runtime) sendUserPacket(ctx context.Context, fallbackConnID string, userID int, pkt interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) error {
	if runtime.broadcaster == nil || userID <= 0 {
		return runtime.sendPacket(fallbackConnID, pkt)
	}
	body, err := pkt.Encode()
	if err != nil {
		return err
	}
	return runtime.broadcaster.Publish(ctx, sessionnotification.UserChannel(userID), codec.EncodeFrame(pkt.PacketID(), body))
}

// canModifyRoom reports whether a user may manage furniture in the target room.
func (runtime *Runtime) canModifyRoom(ctx context.Context, roomID, userID int) bool {
	if runtime.roomAccessChecker == nil {
		return true
	}
	return runtime.roomAccessChecker(ctx, roomID, userID)
}

// isTileOccupied reports whether the target room tile currently contains a player entity.
func (runtime *Runtime) isTileOccupied(roomID, x, y int) bool {
	if runtime.roomOccupancyChecker == nil {
		return false
	}
	return runtime.roomOccupancyChecker(roomID, x, y)
}

// canManageItem reports whether the actor may manage one furniture item in the target room.
func (runtime *Runtime) canManageItem(ctx context.Context, roomID, userID, itemID int) bool {
	return runtime.canModifyRoom(ctx, roomID, userID)
}

// seatEntriesFor returns the cached seat entries for one placed item, if present.
func (runtime *Runtime) seatEntriesFor(roomID, itemID int) []seatEntry {
	runtime.seatMu.RLock()
	defer runtime.seatMu.RUnlock()
	entries := make([]seatEntry, 0)
	for _, e := range runtime.seatCache[roomID] {
		if e.itemID == itemID {
			entries = append(entries, e)
		}
	}
	return entries
}

// TileSeatCheckerFor returns seat properties for the topmost sittable item at a tile.
func (runtime *Runtime) TileSeatCheckerFor(roomID, x, y int) (height float64, dir int, canSit, canLay bool) {
	runtime.seatMu.RLock()
	defer runtime.seatMu.RUnlock()
	bestHeight := 0.0
	found := false
	for _, e := range runtime.seatCache[roomID] {
		if e.x == x && e.y == y && (e.canSit || e.canLay) {
			if !found || e.height >= bestHeight {
				height = e.height
				dir = e.dir
				canSit = e.canSit
				canLay = e.canLay
				bestHeight = e.height
				found = true
			}
		}
	}
	if !found {
		return 0, 0, false, false
	}
	return height, dir, canSit, canLay
}

// ResolveSeatTargetFor returns the resolved target tile for a seat or lay-capable furniture tile.
func (runtime *Runtime) ResolveSeatTargetFor(roomID, x, y int) (targetX, targetY int, ok bool) {
	runtime.seatMu.RLock()
	entries := append([]seatEntry(nil), runtime.seatCache[roomID]...)
	runtime.seatMu.RUnlock()
	return runtime.resolveSeatTarget(roomID, entries, x, y)
}

// replaceSeatEntries replaces all cached seat entries for one item.
func (runtime *Runtime) replaceSeatEntries(roomID, itemID int, entries []seatEntry) {
	runtime.seatMu.Lock()
	defer runtime.seatMu.Unlock()
	filtered := runtime.seatCache[roomID][:0]
	for _, entry := range runtime.seatCache[roomID] {
		if entry.itemID != itemID {
			filtered = append(filtered, entry)
		}
	}
	runtime.seatCache[roomID] = append(filtered, entries...)
}

// removeSeatEntries removes all seat cache entries for one item.
func (runtime *Runtime) removeSeatEntries(roomID, itemID int) {
	runtime.seatMu.Lock()
	defer runtime.seatMu.Unlock()
	filtered := runtime.seatCache[roomID][:0]
	for _, entry := range runtime.seatCache[roomID] {
		if entry.itemID != itemID {
			filtered = append(filtered, entry)
		}
	}
	runtime.seatCache[roomID] = filtered
}

// SendRoomFloorItems loads placed items for one room and sends the floor list to one connection.
func (runtime *Runtime) SendRoomFloorItems(ctx context.Context, connID string, roomID int) error {
	items, err := runtime.service.ListRoomItems(ctx, roomID)
	if err != nil {
		return err
	}
	runtime.clearRoomPlacementEntries(roomID)
	owners := make(map[int]string)
	floorItems := make([]furnipacket.FurnitureFloorItem, 0, len(items))
	for _, item := range items {
		def, defErr := runtime.service.FindDefinitionByID(ctx, item.DefinitionID)
		if defErr != nil {
			continue
		}
		floorItems = append(floorItems, furnipacket.FurnitureFloorItem{
			ItemID: item.ID, SpriteID: def.SpriteID,
			X: item.X, Y: item.Y, Dir: item.Dir, Z: item.Z,
			StackHeight: runtime.effectiveStackHeight(item, def),
			ExtraData:   item.ExtraData, UserID: item.UserID,
		})
		if _, seen := owners[item.UserID]; !seen {
			name := ""
			if runtime.usernameResolver != nil {
				if n, resolveErr := runtime.usernameResolver(ctx, item.UserID); resolveErr == nil {
					name = n
				}
			}
			owners[item.UserID] = name
		}
		runtime.syncFloorItemEntries(roomID, item, def)
	}
	return runtime.sendPacket(connID, furnipacket.FurnitureFloorComposer{
		Items: floorItems, Owners: owners,
	})
}

// SendRoomWallItems loads placed wall items for one room and sends the wall list to one connection.
func (runtime *Runtime) SendRoomWallItems(ctx context.Context, connID string, roomID int) error {
	items, err := runtime.service.ListRoomItems(ctx, roomID)
	if err != nil {
		return err
	}
	owners := make(map[int]string)
	wallItems := make([]furnipacket.FurnitureWallItem, 0, len(items))
	for _, item := range items {
		def, defErr := runtime.service.FindDefinitionByID(ctx, item.DefinitionID)
		if defErr != nil || def.ItemType != furnituredomain.ItemTypeWall {
			continue
		}
		wallItems = append(wallItems, furnipacket.FurnitureWallItem{ItemID: item.ID, SpriteID: def.SpriteID, WallPosition: item.WallPosition, ExtraData: item.ExtraData, UserID: item.UserID})
		if _, seen := owners[item.UserID]; !seen {
			name := ""
			if runtime.usernameResolver != nil {
				if resolved, resolveErr := runtime.usernameResolver(ctx, item.UserID); resolveErr == nil {
					name = resolved
				}
			}
			owners[item.UserID] = name
		}
	}
	return runtime.sendPacket(connID, furnipacket.FurnitureWallComposer{Items: wallItems, Owners: owners})
}
