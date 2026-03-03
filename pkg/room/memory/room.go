package memory

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"

	"pixel-server/pkg/room"
)

// RoomRepo is an in-memory fake that implements room.Repository.
type RoomRepo struct {
	mu     sync.RWMutex
	rooms  map[int32]*room.Room
	nextID atomic.Int32
}

// NewRoomRepo returns a ready-to-use in-memory RoomRepo.
func NewRoomRepo() *RoomRepo {
	return &RoomRepo{rooms: make(map[int32]*room.Room)}
}

func (r *RoomRepo) GetByID(_ context.Context, id int32) (*room.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rm, ok := r.rooms[id]
	if !ok {
		return nil, room.ErrNotFound
	}
	cp := *rm
	return &cp, nil
}

func (r *RoomRepo) GetByOwner(_ context.Context, ownerID int32) ([]*room.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*room.Room
	for _, rm := range r.rooms {
		if rm.OwnerID == ownerID {
			cp := *rm
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *RoomRepo) Create(_ context.Context, rm *room.Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rm.ID == 0 {
		rm.ID = r.nextID.Add(1)
	}
	cp := *rm
	r.rooms[cp.ID] = &cp
	return nil
}

func (r *RoomRepo) Update(_ context.Context, rm *room.Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.rooms[rm.ID]; !ok {
		return room.ErrNotFound
	}
	cp := *rm
	r.rooms[cp.ID] = &cp
	return nil
}

func (r *RoomRepo) Delete(_ context.Context, id int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.rooms[id]; !ok {
		return room.ErrNotFound
	}
	delete(r.rooms, id)
	return nil
}

func (r *RoomRepo) GetModel(_ context.Context, _ string) (*room.Model, error) {
	return nil, room.ErrNotFound
}

func (r *RoomRepo) Search(_ context.Context, query string, limit int) ([]*room.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*room.Room
	for _, rm := range r.rooms {
		if strings.Contains(strings.ToLower(rm.Name), strings.ToLower(query)) {
			cp := *rm
			out = append(out, &cp)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *RoomRepo) GetPopular(_ context.Context, limit int) ([]*room.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*room.Room
	for _, rm := range r.rooms {
		cp := *rm
		out = append(out, &cp)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

// compile-time interface check.
var _ room.Repository = (*RoomRepo)(nil)
