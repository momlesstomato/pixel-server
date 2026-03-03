package memory

import (
	"context"
	"sync"

	"pixel-server/pkg/user"
)

// BadgeRepo is an in-memory fake that implements user.BadgeRepository.
type BadgeRepo struct {
	mu     sync.RWMutex
	badges map[int32][]*user.Badge
}

// NewBadgeRepo returns a ready-to-use in-memory BadgeRepo.
func NewBadgeRepo() *BadgeRepo {
	return &BadgeRepo{badges: make(map[int32][]*user.Badge)}
}

// AddBadge is a test helper to seed badges.
func (r *BadgeRepo) AddBadge(b *user.Badge) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *b
	r.badges[b.UserID] = append(r.badges[b.UserID], &cp)
}

func (r *BadgeRepo) GetByUser(_ context.Context, userID int32) ([]*user.Badge, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	badges := r.badges[userID]
	out := make([]*user.Badge, len(badges))
	for i, b := range badges {
		cp := *b
		out[i] = &cp
	}
	return out, nil
}

func (r *BadgeRepo) GetEquipped(_ context.Context, userID int32) ([]*user.Badge, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*user.Badge
	for _, b := range r.badges[userID] {
		if b.Slot > 0 {
			cp := *b
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *BadgeRepo) Equip(_ context.Context, userID int32, code string, slot int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, b := range r.badges[userID] {
		if b.Slot == slot {
			b.Slot = 0
		}
	}
	for _, b := range r.badges[userID] {
		if b.Code == code {
			b.Slot = slot
			return nil
		}
	}
	return user.ErrNotFound
}

func (r *BadgeRepo) Unequip(_ context.Context, userID int32, slot int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, b := range r.badges[userID] {
		if b.Slot == slot {
			b.Slot = 0
			return nil
		}
	}
	return nil
}

// compile-time interface check.
var _ user.BadgeRepository = (*BadgeRepo)(nil)
