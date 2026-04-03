package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkroomchat "github.com/momlesstomato/pixel-sdk/events/room/chat"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"go.uber.org/zap"
)

const (
	chatRange         = 14
	floodLimit        = 3
	floodWindow       = 3 * time.Second
	floodMuteDuration = 10 * time.Second
)

// floodEntry tracks message rate for one entity.
type floodEntry struct {
	// count stores messages sent in the current window.
	count int
	// resetAt stores when the current window expires.
	resetAt time.Time
	// mutedUntil stores when the mute expires.
	mutedUntil time.Time
}

// ChatService manages room chat with flood control and event dispatch.
type ChatService struct {
	// fire stores optional plugin event dispatch.
	fire func(sdk.Event)
	// logger stores structured logging behavior.
	logger *zap.Logger
	// mu protects flood map access.
	mu sync.Mutex
	// flood stores per-entity flood tracking entries.
	flood map[int]*floodEntry
}

// NewChatService creates one room chat service.
func NewChatService(logger *zap.Logger) (*ChatService, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ChatService{logger: logger, flood: make(map[int]*floodEntry)}, nil
}

// SetEventFirer configures optional plugin event dispatch behavior.
func (s *ChatService) SetEventFirer(fire func(sdk.Event)) {
	s.fire = fire
}

// isMuted reports whether an entity is currently flood-muted.
func (s *ChatService) isMuted(virtualID int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.flood[virtualID]
	return ok && time.Now().Before(entry.mutedUntil)
}

// recordFlood increments flood counter and applies mute when limit is reached.
func (s *ChatService) recordFlood(virtualID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	entry, ok := s.flood[virtualID]
	if !ok || now.After(entry.resetAt) {
		s.flood[virtualID] = &floodEntry{count: 1, resetAt: now.Add(floodWindow)}
		return
	}
	entry.count++
	if entry.count >= floodLimit {
		entry.mutedUntil = now.Add(floodMuteDuration)
	}
}

// Talk delivers a proximity chat message to nearby room entities.
func (s *ChatService) Talk(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity, roomID int, msg string, bubble int) ([]domain.RoomEntity, error) {
	if s.isMuted(entity.VirtualID) {
		return nil, domain.ErrFloodControl
	}
	if s.fire != nil {
		ev := &sdkroomchat.ChatSending{RoomID: roomID, UserID: entity.UserID, VirtualID: entity.VirtualID, Message: msg, ChatType: "talk"}
		s.fire(ev)
		if ev.Cancelled() {
			return nil, domain.ErrAccessDenied
		}
	}
	s.recordFlood(entity.VirtualID)
	recipients := proximityFilter(inst.Entities(), entity, chatRange)
	if s.fire != nil {
		s.fire(&sdkroomchat.ChatSent{RoomID: roomID, UserID: entity.UserID, VirtualID: entity.VirtualID, Message: msg, ChatType: "talk"})
	}
	return recipients, nil
}

// Shout delivers a room-wide shout message to all room entities.
func (s *ChatService) Shout(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity, roomID int, msg string, bubble int) ([]domain.RoomEntity, error) {
	if s.isMuted(entity.VirtualID) {
		return nil, domain.ErrFloodControl
	}
	if s.fire != nil {
		ev := &sdkroomchat.ChatSending{RoomID: roomID, UserID: entity.UserID, VirtualID: entity.VirtualID, Message: msg, ChatType: "shout"}
		s.fire(ev)
		if ev.Cancelled() {
			return nil, domain.ErrAccessDenied
		}
	}
	s.recordFlood(entity.VirtualID)
	recipients := inst.Entities()
	if s.fire != nil {
		s.fire(&sdkroomchat.ChatSent{RoomID: roomID, UserID: entity.UserID, VirtualID: entity.VirtualID, Message: msg, ChatType: "shout"})
	}
	return recipients, nil
}

// Whisper delivers a private message to the sender and one target entity.
func (s *ChatService) Whisper(_ context.Context, entity *domain.RoomEntity, roomID int, target *domain.RoomEntity, msg string, bubble int) ([]domain.RoomEntity, error) {
	if s.fire != nil {
		ev := &sdkroomchat.ChatSending{RoomID: roomID, UserID: entity.UserID, VirtualID: entity.VirtualID, Message: msg, ChatType: "whisper"}
		s.fire(ev)
		if ev.Cancelled() {
			return nil, domain.ErrAccessDenied
		}
	}
	recipients := []domain.RoomEntity{*entity, *target}
	if s.fire != nil {
		s.fire(&sdkroomchat.ChatSent{RoomID: roomID, UserID: entity.UserID, VirtualID: entity.VirtualID, Message: msg, ChatType: "whisper"})
	}
	return recipients, nil
}

// proximityFilter returns entities within Manhattan distance of origin.
func proximityFilter(entities []domain.RoomEntity, origin *domain.RoomEntity, maxRange int) []domain.RoomEntity {
	result := make([]domain.RoomEntity, 0, len(entities))
	for i := range entities {
		e := entities[i]
		dx := e.Position.X - origin.Position.X
		dy := e.Position.Y - origin.Position.Y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx+dy <= maxRange {
			result = append(result, e)
		}
	}
	return result
}

// _ suppresses unused import warning.
var _ = fmt.Sprintf
